package saccount

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

var _ interfaces.IStakeManager = (*StakeManager)(nil)

type StakeManager struct {
	stateLedger ledger.StateLedger
}

func NewStakeManager() *StakeManager {
	return &StakeManager{}
}

func (sm *StakeManager) Init(stateLedger ledger.StateLedger) {
	sm.stateLedger = stateLedger
}

func (sm *StakeManager) GetDepositInfo(account ethcommon.Address) *interfaces.DepositInfo {
	depositInfo := &interfaces.DepositInfo{
		Staked:       false,
		Stake:        big.NewInt(0),
		WithdrawTime: big.NewInt(0),
	}
	depositInfo.Deposit = sm.stateLedger.GetBalance(types.NewAddress(account.Bytes()))
	return depositInfo
}

// nolint
func (sm *StakeManager) getStakeInfo(addr ethcommon.Address) (info *interfaces.StakeInfo) {
	info = &interfaces.StakeInfo{
		Stake:           big.NewInt(0),
		UnstakeDelaySec: big.NewInt(0),
	}
	return info
}

func (sm *StakeManager) BalanceOf(account ethcommon.Address) *big.Int {
	return sm.stateLedger.GetBalance(types.NewAddress(account.Bytes()))
}

func (sm *StakeManager) DepositTo(account ethcommon.Address) {
	// empty implementation
}

func (sm *StakeManager) AddStake(unstakeDelaySec uint32) {
	// empty implementation
}

func (sm *StakeManager) UnlockStake() {
	// empty implementation
}

func (sm *StakeManager) WithdrawStake(withdrawAddress ethcommon.Address) {
	// empty implementation
}

func (sm *StakeManager) WithdrawTo(withdrawAddress ethcommon.Address, withdrawAmount *big.Int) {
	// empty implementation
}
