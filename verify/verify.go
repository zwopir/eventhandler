package verify

import (
	"golang.org/x/crypto/openpgp"
	// "github.com/prometheus/common/log"
	"io"
	"errors"
	"bytes"
)


type Signer struct {
	entity *openpgp.Entity
}

type Verifier struct {
	keyring openpgp.KeyRing
}

func NewSigner(privateKeyringBuffer io.Reader) (*Signer, error) {
	// entityList implements the openpgp.KeyRing interface
	entityList, err := readKeyring(privateKeyringBuffer)
	if err != nil {
		return nil, err
	}
	return &Signer{
		entity: entityList[0],
	}, nil
}

func NewVerifier(publicKeyringBuffer io.Reader) (*Verifier, error) {
	keyring, err := readKeyring(publicKeyringBuffer)
	if err != nil {
		return nil, err
	}
	return &Verifier{
		keyring: keyring,
	}, nil
}

func readKeyring(keyringBuffer io.Reader) (openpgp.EntityList, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(keyringBuffer)
	if err != nil {
		return nil, err
	}
	if len(keyring) < 1 {
		return nil, errors.New("no keys found in keyring")
	}
	return keyring, nil
}

func (s *Signer) Sign(message io.Reader) ([]byte, error) {
	signature := new(bytes.Buffer)
	err := openpgp.ArmoredDetachSign(signature, s.entity, message, nil)
	if err != nil {
		return nil, err
	}
	return signature.Bytes(), nil
}


func (v *Verifier) Verify(message, signature []byte) error {
	messageBuffer := bytes.NewBuffer(message)
	signatureBuffer := bytes.NewBuffer(signature)
	_, err := openpgp.CheckArmoredDetachedSignature(v.keyring, messageBuffer, signatureBuffer)
	return err
}







