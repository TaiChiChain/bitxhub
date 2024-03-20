package finance

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

type Incentive struct {
	repo                *repo.GenesisConfig
	logger              logrus.FieldLogger
	miningAddr          ethcommon.Address
	userAcquisitionAddr ethcommon.Address

	totalAmount *big.Int
	// todo: change the remaining in memory to the receiver address in the future, in case restart vanish the data
	remaining    *big.Int
	blockNumHalf uint64
	rules        []MiningRules

	axc *token.Manager
}

func NewIncentive(config *repo.GenesisConfig, nvm common.VirtualMachine) (*Incentive, error) {
	logger := loggers.Logger(loggers.Finance)

	systemContract := nvm.GetContractInstance(types.NewAddressByStr(common.AXCContractAddr))
	manager, ok := systemContract.(*token.Manager)
	if !ok {
		return nil, ErrAXCContractType
	}

	return &Incentive{
		repo:   config,
		logger: logger,
		axc:    manager,
	}, nil
}

func (in *Incentive) SetMiningRewards(receiver ethcommon.Address, ledger ledger.StateLedger, blockHeight uint64) error {
	msgFrom := in.miningAddr
	logs := make([]common.Log, 0)
	in.axc.SetContext(&common.VMContext{
		CurrentLogs:   &logs,
		StateLedger:   ledger,
		CurrentHeight: blockHeight,
		CurrentUser:   &msgFrom,
	})
	value := in.calculateMiningRewards(blockHeight)
	return in.axc.Transfer(receiver, value)
}

func (in *Incentive) calculateMiningRewards(currentBlock uint64) *big.Int {
	return big.NewInt(0)
}

func (in *Incentive) Unlock(_ ethcommon.Address, _ *big.Int) {
	// todo: implement it
}

func (in *Incentive) SetUserAcquisitionReward(ledger ledger.StateLedger, blockHeight uint64, block *types.Block) error {
	if blockHeight > in.repo.Incentive.Referral.BlockToNone {
		return nil
	}
	msgFrom := in.userAcquisitionAddr
	logs := make([]common.Log, 0)
	in.axc.SetContext(&common.VMContext{
		CurrentLogs:   &logs,
		StateLedger:   ledger,
		CurrentHeight: blockHeight,
		CurrentUser:   &msgFrom,
	})
	return in.setUserAcquisitionReward(block)
}

func (in *Incentive) setUserAcquisitionReward(block *types.Block) error {
	return nil
}
