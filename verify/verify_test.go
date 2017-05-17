package verify

import (
	"bytes"
	"os"
	"testing"
)

var verifyTestTable = []struct {
	message              []byte
	failingMessage       []byte
	privateKeyPath       string
	publicKeyPath        string
	failingPublicKeyPath string
}{
	{
		[]byte(`a test message`),
		[]byte(`a modified test message`),
		"testdata/private.key",
		"testdata/public.key",
		"testdata/non_matching_public.key",
	},
}

func TestNewSigner(t *testing.T) {
	for _, tt := range verifyTestTable {
		privkeyBuffer, err := os.Open(tt.privateKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewSigner(privkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNewVerifier(t *testing.T) {
	for _, tt := range verifyTestTable {
		pubkeyBuffer, err := os.Open(tt.publicKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewVerifier(pubkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSigner_Sign(t *testing.T) {
	for _, tt := range verifyTestTable {
		keyringBuffer, err := os.Open(tt.privateKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		signer, err := NewSigner(keyringBuffer)
		if err != nil {
			t.Fatal(err)
		}

		message := bytes.NewBuffer(tt.message)
		signature, err := signer.Sign(message)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("successfully signed %q to %q", tt.message, string(signature))
	}
}

func TestVerifier_Verify(t *testing.T) {
	for _, tt := range verifyTestTable {
		privkeyBuffer, err := os.Open(tt.privateKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		pubkeyBuffer, err := os.Open(tt.publicKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		signer, err := NewSigner(privkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
		verifier, err := NewVerifier(pubkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
		message := bytes.NewBuffer(tt.message)
		signature, err := signer.Sign(message)
		if err != nil {
			t.Fatal(err)
		}
		err = verifier.Verify(tt.message, signature)
		if err != nil {
			t.Fatal(err)
		}
		err = verifier.Verify(tt.failingMessage, signature)
		if err == nil {
			t.Fatal("signature check of a modified message passes")
		} else {
			t.Logf("signature check of a modified message correctly fails with: %s", err)
		}
	}
}

func TestVerifier_Verify2(t *testing.T) {
	for _, tt := range verifyTestTable {
		privkeyBuffer, err := os.Open(tt.privateKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		pubkeyBuffer, err := os.Open(tt.failingPublicKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		signer, err := NewSigner(privkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
		verifier, err := NewVerifier(pubkeyBuffer)
		if err != nil {
			t.Fatal(err)
		}
		message := bytes.NewBuffer(tt.message)
		signature, err := signer.Sign(message)
		err = verifier.Verify(tt.failingMessage, signature)
		if err == nil {
			t.Fatal("signature check with a non-matching public key passes")
		} else {
			t.Logf("signature check with a non-matching public key correctly fails with: %s", err)
		}
	}
}
