package finance

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axc"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

type Incentive struct {
	repo          *repo.GenesisConfig
	logger        logrus.FieldLogger
	communityAddr string

	totalAmount *big.Int
	// todo: change the remaining in memory to the receiver address in the future, in case restart vanish the data
	remaining    *big.Int
	blockNumHalf uint64
	rules        []MiningRules

	axc *axc.Manager
}

func NewIncentive(config *repo.GenesisConfig, nvm common.VirtualMachine) (*Incentive, error) {
	logger := loggers.Logger(loggers.Finance)

	systemContract := nvm.GetContractInstance(types.NewAddressByStr(common.AXCContractAddr))
	manager, ok := systemContract.(*axc.Manager)
	if !ok {
		return nil, ErrAXCContractType
	}

	var communityAddr string
	for _, entity := range config.Incentive.Distributions {
		if entity.Name == "Community" {
			communityAddr = entity.Addr
		}
	}
	if communityAddr == "" {
		return nil, ErrNoCommunityFound
	}
	totalAmount, _ := new(big.Int).SetString(config.Incentive.Mining.TotalAmount, 10)
	years := config.Incentive.Mining.BlockNumToNone / config.Incentive.Mining.BlockNumToHalf
	rules := make([]MiningRules, years)
	accumulativeRatio := float64(0)
	startRatio := 0.5
	for i := uint64(0); i < years; i++ {
		rules[i] = MiningRules{
			startBlock: i * config.Incentive.Mining.BlockNumToHalf,
			endBlock:   (i + 1) * config.Incentive.Mining.BlockNumToHalf,
			percentage: startRatio,
		}
		accumulativeRatio += startRatio
		if i < years-2 {
			startRatio /= 2
		} else {
			startRatio = 1 - accumulativeRatio
		}
	}

	return &Incentive{
		repo:          config,
		logger:        logger,
		communityAddr: communityAddr,

		totalAmount:  totalAmount,
		remaining:    totalAmount,
		blockNumHalf: config.Incentive.Mining.BlockNumToHalf,
		rules:        rules,
		axc:          manager,
	}, nil
}

func (in *Incentive) CalculateMiningRewards(receiver ethcommon.Address, ledger ledger.StateLedger, blockHeight uint64) error {
	msgFrom := types.NewAddressByStr(in.communityAddr).ETHAddress()
	logs := make([]common.Log, 0)
	in.axc.SetContext(&common.VMContext{
		CurrentLogs:   &logs,
		StateLedger:   ledger,
		CurrentHeight: blockHeight,
		CurrentUser:   &msgFrom,
	})
	value := in.calculateMiningRewards(blockHeight)
	if value.Int64() != 0 && in.remaining.Uint64() == 0 {
		return ErrMiningRewardExceeds
	}
	if value.Int64() != 0 && in.remaining.Cmp(value) == -1 {
		value = in.remaining
	}
	in.remaining = new(big.Int).Sub(in.remaining, value)
	return in.axc.Transfer(receiver, value)
}

func (in *Incentive) calculateMiningRewards(currentBlock uint64) *big.Int {
	index := currentBlock / in.blockNumHalf
	if index >= uint64(len(in.rules)) {
		return big.NewInt(0)
	}
	percentage := big.NewInt(int64(10000 * in.rules[index].percentage))
	rewardsTotal := new(big.Int).Div(new(big.Int).Mul(in.totalAmount, percentage), big.NewInt(10000))
	return new(big.Int).Div(rewardsTotal, new(big.Int).SetUint64(in.blockNumHalf))
}

func (in *Incentive) Unlock(_ ethcommon.Address, _ *big.Int) {
	// todo: implement it
}
