package saccount

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
)

const (
	AlgoSecp256R1 uint8 = 0
)

var (
	ErrInvalidPasskeySignatureLen   = errors.New("invalid passkey signature length")
	ErrInvalidPasskeyAlgo           = errors.New("invalid passkey algo")
	ErrPasskeyVerificationFailed    = errors.New("passkey verification failed")
	ErrPasskeyUserOpHashCheckFailed = errors.New("passkey userOpHash check failed")
)

type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

var _ interfaces.IValidator = (*Passkey)(nil)

type Passkey struct {
	PubKeyX *big.Int
	PubKeyY *big.Int
	Algo    uint8
}

func (passkey *Passkey) Validate(userOp *interfaces.UserOperation, userOpHash [32]byte) (*interfaces.Validation, error) {
	validation := &interfaces.Validation{
		SigValidation: interfaces.SigValidationFailed,
	}
	if len(userOp.Signature) != 64 {
		return validation, ErrInvalidPasskeySignatureLen
	}

	switch passkey.Algo {
	case AlgoSecp256R1:
		// parse clientData and check userOpHash
		clientData := &ClientData{}
		err := json.Unmarshal(userOp.ClientData, clientData)
		if err != nil {
			return validation, err
		}

		challenge, err := base64.RawURLEncoding.DecodeString(clientData.Challenge)
		if err != nil {
			return validation, err
		}
		if !bytes.Equal(challenge, userOpHash[:]) {
			return validation, ErrPasskeyUserOpHashCheckFailed
		}

		pk := &ecdsa.PublicKey{Curve: elliptic.P256(), X: passkey.PubKeyX, Y: passkey.PubKeyY}
		clientDataHash := sha256.Sum256(userOp.ClientData)
		h := sha256.Sum256(append(userOp.AuthData, clientDataHash[:]...))

		r := new(big.Int).SetBytes(userOp.Signature[:32])
		s := new(big.Int).SetBytes(userOp.Signature[32:])
		if !ecdsa.Verify(pk, h[:], r, s) {
			return validation, ErrPasskeyVerificationFailed
		}
		validation.SigValidation = interfaces.SigValidationSucceeded
		return validation, nil
	}

	return validation, ErrInvalidPasskeyAlgo
}

func (passkey *Passkey) PostUserOp(validation *interfaces.Validation, actualGasCost, totalValue *big.Int) error {
	return nil
}
