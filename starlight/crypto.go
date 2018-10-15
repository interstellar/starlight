package starlight

import (
	"crypto/rand"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

// sealBox encrypts plaintext using the given password and salt.
// The password is a user-supplied value, it can be any length
// and need not have any particular properties (other than
// having enough entropy).
func sealBox(plaintext, password []byte) (box []byte) {
	b := make([]byte, 24+24) // scrypt-salt || secretbox-nonce
	randRead(b)
	var nonce [24]byte
	copy(nonce[:], b[24:])
	key := derivePasswordKey(password, b[:24])
	return secretbox.Seal(b, plaintext, &nonce, key)
}

// openBox decrypts box using password and returns the contents.
// The password and salt must match the values previously used to seal box;
// if it doesn't match, openBox returns nil.
func openBox(box, password []byte) (plaintext []byte) {
	// Layout of box: salt || nonce || ciphertext.
	// That is 24 bytes, 24 bytes, and the rest.
	key := derivePasswordKey(password, box[:24])
	var nonce [24]byte
	copy(nonce[:], box[24:])
	plaintext, ok := secretbox.Open(nil, box[48:], &nonce, key)
	if !ok {
		return nil
	}
	return plaintext
}

// derivePasswordKey derives a 32-byte key from password
// using scrypt as a KDF.
func derivePasswordKey(password, salt []byte) (key *[32]byte) {
	// Difficulty parameters for scrypt.
	// See https://blog.filippo.io/the-scrypt-parameters/.
	// (There's no point in using higher difficulty than the
	// password digest, since that will be stored right along
	// with any data encrypted with the key we derive here.)
	const N, r, p = 1 << 15, 8, 1

	key = new([32]byte)
	b, err := scrypt.Key(password, salt, N, r, p, len(key))
	if err != nil {
		panic(err) // means a bug choosing values for N, r, or p
	}
	copy(key[:], b)
	return key
}

// randRead fills p with random bytes.
func randRead(p []byte) {
	_, err := rand.Read(p)
	if err != nil {
		panic(err) // don't try to operate with a bad RNG
	}
}
