package token

import (
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

type TokenManager struct {
	gov *governance.Governance

	logger      logrus.FieldLogger
	account     ledger.IAccount
	stateLedger ledger.StateLedger
}

func NewTokenManager(cfg *common.SystemContractConfig) *TokenManager {
	return &TokenManager{
		logger: cfg.Logger,
	}
}

func (tm *TokenManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	tm.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.TokenManagerContractAddr))
	tm.stateLedger = stateLedger
}

func (tm *TokenManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	return nil, nil
}

func (tm *TokenManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := tm.getArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}

	return common.CalculateDynamicGas(*callArgs.Data), nil
}

func (tm *TokenManager) getArgs(msg *vm.Message) (*vm.Message, error) {
	return msg, nil
}
