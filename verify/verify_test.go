package verify

import (
	"testing"
	"os"
	"bytes"
)

func TestNewSigner(t *testing.T) {
	t.Log("allet jut")
}


var verifyTestTable = []struct{
	message []byte
	failingMessage []byte
	privateKeyPath string
	publicKeyPath string
}{
	{
		[]byte(`a test message`),
		[]byte(`a modified test message`),
		"testdata/private.key",
		"testdata/public.key",
	},
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
