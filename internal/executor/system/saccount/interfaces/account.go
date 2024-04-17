package interfaces

import (
	"math/big"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

const (
	SigValidationSucceeded = 0
	SigValidationFailed    = 1
)

var (
	MaxUint160, _ = new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffff", 16)
	MaxUint48, _  = new(big.Int).SetString("ffffffffffff", 16)

	defaultRemainingLimit = big.NewInt(3000000000000000000)
)

type Validation struct {
	// signature validation result
	SigValidation uint

	// valid time range
	ValidUntil uint64

	ValidAfter uint64

	// remaining limit to spend
	RemainingLimit *big.Int
}

type IAccount interface {
	common.SystemContract

	// validateUserOp validate user's signature and nonce
	// the entryPoint will make the call to the recipient only if this validation call returns successfully.
	// signature failure should be reported by returning SIG_VALIDATION_FAILED (1).
	// This allows making a "simulation call" without a valid signature
	// Other failures (e.g. nonce mismatch, or invalid signature format) should still revert to signal failure.
	ValidateUserOp(userOp UserOperation, userOpHash []byte, missingAccountFunds *big.Int) (validationData *big.Int, err error)
}

// extract validation
// keep compatibility with eth abstract account
//
//	struct ValidationData {
//		address aggregator;
//		uint48 validAfter;
//		uint48 validUntil;
//	}
func ParseValidationData(validationData *big.Int) *Validation {
	if validationData == nil {
		return nil
	}

	sigValidation := new(big.Int).And(validationData, MaxUint160)
	validUntil := new(big.Int).And(new(big.Int).Rsh(validationData, 160), MaxUint48)
	if validUntil.Sign() == 0 {
		validUntil = new(big.Int).SetBytes(MaxUint48.Bytes())
	}

	validAfter := new(big.Int).And(new(big.Int).Rsh(validationData, 160+48), MaxUint48)
	return &Validation{
		SigValidation:  uint(sigValidation.Uint64()),
		ValidUntil:     validUntil.Uint64(),
		ValidAfter:     validAfter.Uint64(),
		RemainingLimit: defaultRemainingLimit,
	}
}

// keep compatibility with eth abstract account
func PackValidationData(validation *Validation) *big.Int {
	if validation == nil {
		return nil
	}

	return new(big.Int).Or(
		new(big.Int).Or(
			big.NewInt(int64(validation.SigValidation)),
			new(big.Int).Lsh(big.NewInt(int64(validation.ValidUntil)), 160),
		),
		new(big.Int).Lsh(big.NewInt(int64(validation.ValidAfter)), 160+48),
	)
}

func IntersectTimeRange(validationData *Validation, paymasterValidationData *big.Int) *Validation {
	if validationData == nil || paymasterValidationData == nil {
		return nil
	}

	accountValidationData := validationData
	pmValidationData := ParseValidationData(paymasterValidationData)
	sigValidationResult := accountValidationData.SigValidation
	if sigValidationResult == SigValidationSucceeded {
		sigValidationResult = pmValidationData.SigValidation
	}
	validAfter := accountValidationData.ValidAfter
	validUntil := accountValidationData.ValidUntil
	pmValidAfter := pmValidationData.ValidAfter
	pmValidUntil := pmValidationData.ValidUntil

	if validAfter < pmValidAfter {
		validAfter = pmValidAfter
	}
	if validUntil > pmValidUntil {
		validUntil = pmValidUntil
	}
	return &Validation{
		SigValidation:  sigValidationResult,
		ValidAfter:     validAfter,
		ValidUntil:     validUntil,
		RemainingLimit: accountValidationData.RemainingLimit,
	}
}
