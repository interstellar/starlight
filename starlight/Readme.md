Starlight is a demo for payment channels on Stellar.

Please note that this software is still in development and will
likely contain multiple security and functional bugs and issues.
This code is for testing only on the Stellar testnet. It is not
compatible with the Stellar mainnet. Attempting to modify it for
use on the Stellar mainnet may result in permanent XLM loss. Any
use of this software is strictly at your own risk, and we will not
be liable for any losses or damages.

# Dev Setup

You only have to do this part once.

This assumes you have a standard i10r environment ready to go.

Set up mkcert:
	$ brew install mkcert
	$ brew install nss
	$ mkcert -install

Generate a localhost cert where starlightd can find it:
	$ cd $I10R/starlight/wallet
	$ mkdir starlight
	$ cd starlight/ && mkcert localhost


# Compile and Run

This is the main development workflow.

Build the Go part & run it:
	$ go install i10r.io/cmd/starlightd
	$ cd $I10R/starlight/wallet
	$ starlightd

Check its health:
	$ open https://localhost:7000

Build the React part & run it:
	$ cd $I10R/starlight/wallet
	$ npm install
	$ npm start

App should be running at:
	$ open https://localhost:5000


# Run multiple starlight instances

To simulate payments between two users, you'll need to run
two starlight instances simultaneously. To do this, you'll
need to set up a separate starlight data directory and start
a starlightd server and React app on separate ports.

You can have as many as you want at a time. This directory
holds the complete state for starlightd, so if you make two
different directories, they won't interfere.

Create a new data directory, generate localhost cert:
	$ cd $I10R/starlight/wallet
	$ mkdir starlight-1
	$ cd starlight-1/ && mkcert localhost

Start starlightd:
	$ cd $I10R/starlight/wallet
	$ starlightd -listen localhost:7001 -data ./starlight-1

Check its health:
	$ open https://localhost:7001

Start the React app
	$ cd $I10R/starlight/wallet
	$ STARLIGHTD_URL=https://localhost:7001 PORT=5001 npm start

# Deploying your own custom client

Starlight serves the client by grabbing a deployed version from an S3 bucket.
The go server loads html that, with JS, fetches the webpack compiled
`index.html` and replaces its contents with the S3 bucket `index.html`.

If you'd like to make changes to the client, you can do so by deploying your
own version of the react app to your own S3 bucket.

1. Update [`sync-frontend.sh`](https://github.com/interstellar/i10r/blob/main/starlight/sync-frontend.sh) to point to your S3 bucket.
1. Update the [webpack configuration](https://github.com/interstellar/i10r/blob/main/starlight/wallet/webpack/webpack.app.js)
	 `publicPath` to point to your S3 bucket URL.
1. Run `./sync-frontend.sh`. This script uses the aws cli, so make
	 sure your system is set up with proper credentials. We use [`aws-vault`](https://github.com/99designs/aws-vault) to
	 manage our credentials.
1. Update the [api index endpoint](https://github.com/interstellar/i10r/blob/main/starlight/walletrpc/handler.go#L64) to point to
	 the `index.html` that has been uploaded to your S3 bucket.

# Running tests

The Starlight project has unit tests and integration tests for the
starlight server and the wallet frontend.

To run the Starlight server unit tests:
	$ cd $I10R/starlight
	$ go test ./...

To run the Starlight wallet unit tests:
	$ cd $I10R/starlight/wallet
	$ ./bin/tests

To run the Starlight integration tests:
	$ cd $I10R/starlight/starlighttest
	$ go test
