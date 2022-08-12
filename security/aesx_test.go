package security

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetPrefix("Cipher Suite >")
	log.Println("Running tests...")
	exitVal := m.Run()
	log.Println("All tests are run")

	os.Exit(exitVal)
}

func TestEncryptDecrypt(t *testing.T) {
	const secret string = "secret"
	var plaintext_canned string = "plain text"
	ciphertext, err := EncryptString(plaintext_canned, secret)
	if err != nil {
		t.Failed()
		return
	}
	log.Printf("Encrypted text: %s", ciphertext)
	plaintext, err := DecryptString(ciphertext, secret)
	if err != nil {
		log.Println(err)
		t.Failed()
		return
	}
	if len(plaintext) != len(plaintext_canned) {
		log.Println("Failed: lenths are not equal")
		t.Failed()
		return
	}
	if plaintext != plaintext_canned {
		log.Printf("Failed: texts are not equal %s != %s", plaintext, plaintext_canned)
		t.Failed()
		return
	}
	log.Printf("Decrypted text: %s", plaintext)
}

func TestEncrypt(t *testing.T) {
	const secret string = "secret"
	var plaintext string = "plain text"
	ciphertext, err := EncryptString(plaintext, secret)
	if err != nil {
		t.Failed()
		return
	}
	log.Println(ciphertext)
}

func TestDecrypt(t *testing.T) {
	const secret string = "secret"
	var ciphertext string = "otGPpRUxQPe7Paxi8yqf/NyHg6haa5CfynTkOdWkhHWxwoP9bbTASgYOjDYChPlQMaKau+xJ"
	plaintext, err := DecryptString(ciphertext, secret)
	if err != nil {
		t.Failed()
		return
	}
	log.Println(plaintext)
}
