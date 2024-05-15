// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package epoch_manager

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
	_ = packer.RevertError{}
)

// ConsensusParams is an auto generated low-level Go binding around an user-defined struct.
type ConsensusParams struct {
	ProposerElectionType                               string
	CheckpointPeriod                                   uint64
	HighWatermarkCheckpointPeriod                      uint64
	MaxValidatorNum                                    uint64
	MinValidatorNum                                    uint64
	BlockMaxTxNum                                      uint64
	EnableTimedGenEmptyBlock                           bool
	NotActiveWeight                                    int64
	AbnormalNodeExcludeView                            uint64
	AgainProposeIntervalBlockInValidatorsNumPercentage uint64
	ContinuousNullRequestToleranceNumber               uint64
	ReBroadcastToleranceNumber                         uint64
}

// EpochInfo is an auto generated low-level Go binding around an user-defined struct.
type EpochInfo struct {
	Epoch           uint64
	EpochPeriod     uint64
	StartBlock      uint64
	ConsensusParams ConsensusParams
	FinanceParams   FinanceParams
	MiscParams      MiscParams
	StakeParams     StakeParams
}

// FinanceParams is an auto generated low-level Go binding around an user-defined struct.
type FinanceParams struct {
	GasLimit               uint64
	MinGasPrice            *big.Int
	StartGasPriceAvailable bool
	StartGasPrice          *big.Int
	MaxGasPrice            *big.Int
	GasChangeRateValue     uint64
	GasChangeRateDecimals  uint64
}

// MiscParams is an auto generated low-level Go binding around an user-defined struct.
type MiscParams struct {
	TxMaxSize uint64
}

// StakeParams is an auto generated low-level Go binding around an user-defined struct.
type StakeParams struct {
	StakeEnable                      bool
	MaxAddStakeRatio                 uint64
	MaxUnlockStakeRatio              uint64
	UnlockPeriod                     uint64
	MaxPendingInactiveValidatorRatio uint64
	MinDelegateStake                 *big.Int
	MinValidatorStake                *big.Int
	MaxValidatorStake                *big.Int
}

type EpochManager interface {

	// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
	//
	// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint256,bool,uint256,uint256,uint64,uint64),(uint64),(bool,uint64,uint64,uint64,uint64,uint256,uint256,uint256)) epochInfo)
	CurrentEpoch() (EpochInfo, error)

	// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
	//
	// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint256,bool,uint256,uint256,uint64,uint64),(uint64),(bool,uint64,uint64,uint64,uint64,uint256,uint256,uint256)) epochInfo)
	HistoryEpoch(epochID uint64) (EpochInfo, error)

	// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
	//
	// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint256,bool,uint256,uint256,uint64,uint64),(uint64),(bool,uint64,uint64,uint64,uint64,uint256,uint256,uint256)) epochInfo)
	NextEpoch() (EpochInfo, error)
}
