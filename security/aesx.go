package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	rfc2898Salt int = 16
)

func EncryptString(text string, secret string) (string, error) {
	// If there is no concern about interactions between multiple uses
	// of the same key (or a prefix of that key) with the password-
	// based encryption and authentication techniques supported for a
	// given password, then the salt may be generated at random and
	// need not be checked for a particular format by the party
	// receiving the salt. It should be at least eight octets (64
	// bits) long - let's at least double it
	salt := make([]byte, rfc2898Salt)
	n, err := rand.Read(salt)
	if err != nil {
		return "", err
	} else if n != len(salt) {
		return "", errors.New("incorrect salt length returned by rand")
	}

	// the drafted v2.1 specification allows use of all five FIPS Approved Hash Functions
	// SHA-1, SHA-224, SHA-256, SHA-384 and SHA-512 for HMAC. To choose, you can pass
	// the `New` functions from the different SHA packages to pbkdf2.Key.
	// Key derives a key from the password, salt and iteration count,
	// returning a []byte of length keylen that can be used as cryptographic key.
	// The key is derived based on the method described as PBKDF2 with the HMAC variant
	// using the supplied hash function.
	// To use a HMAC-SHA-256 based PBKDF2 key derivation function, you can get a derived key
	// for e.g. AES-256 (which needs a 32-byte key):
	// Source: https://pkg.go.dev/golang.org/x/crypto/pbkdf2#pkg-overview
	pbkdf2Key := pbkdf2.Key([]byte(secret), salt, 4096, 32, sha256.New)

	// Seal the plain text. Note: nonce is prepended before encryption
	ciphertext, err := Encrypt([]byte(text), pbkdf2Key)
	if err != nil {
		return "", err
	}
	// Salt value must be known when regenerating the cipher key to in order to decrypt
	ciphertextSalt := make([]byte, 0)
	ciphertextSalt = append(ciphertextSalt, salt...)
	ciphertextSalt = append(ciphertextSalt, ciphertext...)
	// Return a base64 encoded encryption
	return base64.StdEncoding.EncodeToString(ciphertextSalt), nil
}

func DecryptString(ciphertextSalt_b64, secret string) (string, error) {
	// Decode the base64 encoded encrhypted payload
	// DecodeString returns the bytes represented by the base64 string
	ciphertextSalt, err := base64.StdEncoding.DecodeString(ciphertextSalt_b64)
	if err != nil {
		return "", err
	}

	// Our Encryption algorithm has prepended the salt value
	// needed to re-generate the key
	salt := ciphertextSalt[:rfc2898Salt]
	// Get nonce and ciphertext payload
	ciphertextNonce := ciphertextSalt[rfc2898Salt:]

	// The key is derived based on the method described as PBKDF2 with the HMAC variant
	// using the supplied hash function.
	// See Encryption method comments for more details
	// Source: https://pkg.go.dev/golang.org/x/crypto/pbkdf2#pkg-overview
	pbrdf2Key := pbkdf2.Key([]byte(secret), salt, 4096, 32, sha256.New)

	// Decrypt the contents of ciphertextNonce - nonce + ciphertext
	plaintext, err := Decrypt(ciphertextNonce, pbrdf2Key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func Encrypt(plaintext, key []byte) ([]byte, error) {
	// NewCipher creates and returns a new cipher.Block.
	// The key argument should be the AES key, either 16, 24, or 32 bytes
	// to select AES-128, AES-192, or AES-256.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Galois/Counter Mode (GCM) is a mode of operation for symmetric-key
	// cryptographic block ciphers which is widely adopted for its performance.
	// GCM throughput rates for state-of-the-art, high-speed communication channels
	// can be achieved with inexpensive hardware resources.
	// The operation is an authenticated encryption algorithm designed to provide
	// both data authenticity (integrity) and confidentiality.
	// GCM is defined for block ciphers with a block size of 128 bits.
	// Galois Message Authentication Code (GMAC) is an authentication-only variant
	// of the GCM which can form an incremental message authentication code.
	// Both GCM and GMAC can accept initialization vectors of arbitrary length
	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	// To reuse plaintext's storage for the encrypted output, use plaintext[:0]
	// as dst. Otherwise, the remaining capacity of dst must not overlap plaintext.
	nonce := make([]byte, cipher.NonceSize())
	n, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	} else if n != len(nonce) {
		return nil, errors.New("incorrect nonce length returned")
	}
	// Do not want to save the nonce somewhere else in this case.
	// Add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := cipher.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(ciphertextNonce, pbrdf2Key []byte) ([]byte, error) {

	// NewCipher creates and returns a new cipher.Block.
	//The key argument should be the AES key, either 16, 24, or 32
	// bytes to select AES-128, AES-192, or AES-256.
	block, err := aes.NewCipher(pbrdf2Key)
	if err != nil {
		return nil, err
	}

	cipher, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	//Get the nonce size from AEAD cipher
	nonceSize := cipher.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := ciphertextNonce[:nonceSize], ciphertextNonce[nonceSize:]

	//Decrypt the data
	plaintext, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
