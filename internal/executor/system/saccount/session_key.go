package saccount

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

var _ interfaces.IValidator = (*SessionKey)(nil)

// SessionKey is temporary key to control the smart account
// SessionKey has spending limit and valid time range
type SessionKey struct {
	// used to check signature
	Addr ethcommon.Address

	// max limit for spending
	SpendingLimit *big.Int

	SpentAmount *big.Int

	// valid time range
	ValidUntil uint64

	ValidAfter uint64
}

// Validate implements interfaces.Validator.
func (sessionKey *SessionKey) Validate(userOp *interfaces.UserOperation, userOpHash [32]byte) (*interfaces.Validation, error) {
	validationData := &interfaces.Validation{
		SigValidation: interfaces.SigValidationFailed,
	}

	addr, err := recoveryAddrFromSignature(userOpHash, userOp.Signature)
	if err != nil {
		return validationData, fmt.Errorf("session key recovery addr from signature failed: %s", err)
	}
	// cached addr
	validationData.RecoveryAddr = addr

	// check if the address is the same
	if addr != sessionKey.Addr {
		return validationData, fmt.Errorf("userOp signature is not from session key, session key addr: %s, recovery addr: %s", sessionKey.Addr, addr.String())
	}

	validationData.SigValidation = interfaces.SigValidationSucceeded
	validationData.ValidAfter = sessionKey.ValidAfter
	validationData.ValidUntil = sessionKey.ValidUntil
	validationData.RemainingLimit = big.NewInt(0)
	if sessionKey.SpentAmount.Cmp(sessionKey.SpendingLimit) < 0 {
		validationData.RemainingLimit = new(big.Int).Sub(sessionKey.SpendingLimit, sessionKey.SpentAmount)
	}

	return validationData, nil
}

// PostUserOp implements interfaces.Validator.
func (sessionKey *SessionKey) PostUserOp(validation *interfaces.Validation, actualGasCost *big.Int, totalValue *big.Int) error {
	// judge if used session key
	if validation != nil && validation.RemainingLimit != nil {
		spentAmout := new(big.Int).Add(sessionKey.SpentAmount, actualGasCost)
		spentAmout.Add(spentAmout, totalValue)
		if spentAmout.Cmp(sessionKey.SpendingLimit) > 0 {
			return errors.New("spent amount exceeds session spending limit")
		}
		sessionKey.SpentAmount = spentAmout
	}

	return nil
}
