// Command starlightd is a web-UI Starlight wallet.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/kr/secureheader"
	"golang.org/x/crypto/acme/autocert"

	i10rnet "github.com/interstellar/starlight/net"
	"github.com/interstellar/starlight/starlight"
	"github.com/interstellar/starlight/starlight/walletrpc"
)

func main() {
	var (
		listen = flag.String("listen", "localhost:7000", "listen `address` (if no LISTEN_FDS)")
		dir    = flag.String("data", "./starlight-data", "data directory")
		debug  = flag.Bool("debug", false, "print verbose debugging output")
		name   = flag.String("name", "", "name for the agent, used in log output")
	)
	flag.Parse()

	err := os.MkdirAll(*dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	db, err := bolt.Open(filepath.Join(*dir, "db"), 0600, nil)
	if err != nil {
		log.Fatalf("error opening database: %s", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, err := starlight.StartAgent(ctx, db)
	if err != nil {
		log.Fatalf("error starting agent: %s", err)
	}
	g.SetDebug(*debug, *name)

	handler := walletrpc.Handler(g)
	if !i10rnet.IsLoopback(*listen) {
		handler = secureheader.Handler(handler)
	}

	serveLn, redirLn, err := systemdListenersOrListen(*listen)
	if err != nil {
		log.Fatalf("listen: %s", err)
	}
	serveLn = &keepAliveListener{serveLn}

	cert, key, err := findCertKey(*dir)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	if redirLn != nil {
		go func() {
			s := &http.Server{
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Connection", "close")
					url := "https://" + req.Host + req.URL.String()
					http.Redirect(w, req, url, http.StatusMovedPermanently)
				}),
			}
			err := s.Serve(redirLn)
			panic(err)
		}()
	}

	// Timeout settings based on Filippo's late-2016 blog post
	// https://blog.filippo.io/exposing-go-on-the-internet/.
	srv := &http.Server{
		Addr:        *listen,
		ReadTimeout: 5 * time.Second,

		// must be higher than the event handler timeout (10s)
		WriteTimeout: 15 * time.Second,

		IdleTimeout: 120 * time.Second,
		Handler:     handler,
	}
	if i10rnet.IsLoopback(*listen) {
		srv.Serve(serveLn)
	}

	if cert != "" {
		err = srv.ServeTLS(serveLn, cert, key)
	} else {
		tlsConfig := (&autocert.Manager{
			Cache:      autocert.DirCache(filepath.Join(*dir, "autocert")),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autoHostWhitelist(db),
		}).TLSConfig()

		// Security settings from Filippo's late-2016 blog post
		// https://blog.filippo.io/exposing-go-on-the-internet/.
		tlsConfig.PreferServerCipherSuites = true
		tlsConfig.CurvePreferences = []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		}
		tlsConfig.MinVersion = tls.VersionTLS12
		tlsConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		}

		srv.TLSConfig = tlsConfig
		err = srv.Serve(tls.NewListener(serveLn, tlsConfig))
	}
	if err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}

// autoHostWhitelist provides a TOFU-like mechanism as an
// autocert host policy. It whitelists the first-requested
// name and rejects all subsequent names.
func autoHostWhitelist(db *bolt.DB) autocert.HostPolicy {
	return func(ctx context.Context, host string) error {
		return db.Update(func(tx *bolt.Tx) error {
			bu, err := tx.CreateBucketIfNotExists([]byte("daemon"))
			if err != nil {
				return err
			}
			storedHost := string(bu.Get([]byte("acmehost")))
			if host == storedHost {
				return nil
			}
			if storedHost != "" {
				return errors.New("mismatch")
			}
			return bu.Put([]byte("acmehost"), []byte(host))
		})
	}
}

func findCertKey(dir string) (cert, key string, err error) {
	cert = filepath.Join(dir, "localhost.pem")
	key = filepath.Join(dir, "localhost-key.pem")
	_, err = os.Stat(cert)
	if err != nil {
		return "", "", err
	}
	fi, err := os.Stat(key)
	if err != nil {
		return "", "", err
	}
	if fi.Mode()&077 != 0 {
		return "", "", fmt.Errorf(key, "must be accessible only to current user")
	}
	return cert, key, nil
}

// systemdListenersOrListen returns listeners
// for the inherited systemd fds:
// serve for serving responses (on port 443)
// and redir for redirecting to https (on port 80).
// If there are no inherited fds, it listens on addr
// and returns serve; in that case, redir is nil.
func systemdListenersOrListen(addr string) (serve, redir net.Listener, err error) {
	serve, redir, err = systemdListeners()
	if err != nil {
		return nil, nil, err
	}
	if serve != nil {
		return serve, redir, nil
	}
	serve, err = net.Listen("tcp", addr)
	return serve, nil, err
}

func systemdListeners() (serve, redir net.Listener, err error) {
	// Env vars LISTEN_FDS and LISTEN_PID are how systemd
	// tells us the number of inherited fds we have and that
	// they're meant for this process (as opposed to an
	// ancestor process that neglected to mark them as
	// close-on-exec). See also
	// https://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return nil, nil, nil
	}
	n, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil {
		return nil, nil, nil
	}
	if n != 2 {
		return nil, nil, fmt.Errorf("got %d inherited fds, need 2 (port 80 and 443, in that order)", n)
	}
	os.Unsetenv("LISTEN_PID")
	os.Unsetenv("LISTEN_FDS")
	const fd = 3 // systemd always uses fd 3
	syscall.CloseOnExec(fd)
	syscall.CloseOnExec(fd + 1)
	fredir := os.NewFile(fd, fmt.Sprintf("fd%d", fd))     // port 80
	fserve := os.NewFile(fd+1, fmt.Sprintf("fd%d", fd+1)) // port 443
	redir, err = net.FileListener(fredir)
	if err != nil {
		return nil, nil, err
	}
	serve, err = net.FileListener(fserve)
	return serve, redir, err
}

type keepAliveListener struct {
	net.Listener
}

func (ln *keepAliveListener) Accept() (net.Conn, error) {
	conn, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)

	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(3 * time.Minute)

	return tcpConn, nil
}
