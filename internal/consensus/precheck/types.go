package precheck

import (
	"errors"
	"fmt"

	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/ethereum/go-ethereum/core"
)

var precheckErrPrefix = errors.New("verify tx err")

var (
	errTxSign                       = errors.New("tx signature verify failed")
	errTo                           = errors.New("tx from and to address is same")
	errGasPriceTooLow               = errors.New("gas price too low")
	errFeeCapVeryHigh               = core.ErrFeeCapVeryHigh
	errTipVeryHigh                  = core.ErrTipVeryHigh
	errTipAboveFeeCap               = core.ErrTipAboveFeeCap
	errFeeCapTooLow                 = core.ErrFeeCapTooLow
	errMaxInitCodeSizeExceeded      = core.ErrMaxInitCodeSizeExceeded
	errInsufficientFunds            = core.ErrInsufficientFunds
	errIntrinsicGas                 = core.ErrIntrinsicGas
	errInsufficientFundsForTransfer = core.ErrInsufficientFundsForTransfer
)

var errorTypes = map[error]string{
	errTxSign:                       "failed_signature",
	errTo:                           "failed_to_addr",
	errGasPriceTooLow:               "gas_price_too_low",
	errFeeCapVeryHigh:               "fee_cap_very_high",
	errTipVeryHigh:                  "tip_very_high",
	errTipAboveFeeCap:               "err_tip_above_fee_cap",
	errFeeCapTooLow:                 "fee_cap_too_low",
	errMaxInitCodeSizeExceeded:      "max_init_code_size_exceeded",
	errInsufficientFunds:            "err_insufficient_funds",
	errIntrinsicGas:                 "err_intrinsic_gas",
	errInsufficientFundsForTransfer: "insufficient_funds_for_transfer",
}

func convertErrorType(err error) string {
	if res, ok := errorTypes[err]; ok {
		return res
	}
	return err.Error()
}

func wrapError(err error) error {
	return fmt.Errorf("%w:%w", precheckErrPrefix, err)
}

func RespLocalTx(ch chan *consensustypes.TxResp, err error) {
	resp := &consensustypes.TxResp{
		Status: true,
	}

	if err != nil {
		resp.Status = false
		resp.ErrorMsg = err.Error()
	}

	ch <- resp
}
