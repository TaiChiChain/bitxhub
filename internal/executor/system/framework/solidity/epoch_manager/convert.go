package epoch_manager

import (
	"github.com/axiomesh/axiom-kit/types"
)

func (e *EpochInfo) Convert() *types.EpochInfo {
	return &types.EpochInfo{
		Epoch:       e.Epoch,
		EpochPeriod: e.EpochPeriod,
		StartBlock:  e.StartBlock,
		ConsensusParams: types.ConsensusParams{
			ProposerElectionType:          e.ConsensusParams.ProposerElectionType,
			CheckpointPeriod:              e.ConsensusParams.CheckpointPeriod,
			HighWatermarkCheckpointPeriod: e.ConsensusParams.HighWatermarkCheckpointPeriod,
			MaxValidatorNum:               e.ConsensusParams.MaxValidatorNum,
			MinValidatorNum:               e.ConsensusParams.MinValidatorNum,
			BlockMaxTxNum:                 e.ConsensusParams.BlockMaxTxNum,
			EnableTimedGenEmptyBlock:      e.ConsensusParams.EnableTimedGenEmptyBlock,
			NotActiveWeight:               e.ConsensusParams.NotActiveWeight,
			AbnormalNodeExcludeView:       e.ConsensusParams.AbnormalNodeExcludeView,
			AgainProposeIntervalBlockInValidatorsNumPercentage: e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage,
			ContinuousNullRequestToleranceNumber:               e.ConsensusParams.ContinuousNullRequestToleranceNumber,
			ReBroadcastToleranceNumber:                         e.ConsensusParams.ReBroadcastToleranceNumber,
		},
		FinanceParams: types.FinanceParams{
			GasLimit:               e.FinanceParams.GasLimit,
			StartGasPriceAvailable: e.FinanceParams.StartGasPriceAvailable,
			StartGasPrice:          types.CoinNumberByBigInt(e.FinanceParams.StartGasPrice),
			MaxGasPrice:            types.CoinNumberByBigInt(e.FinanceParams.MaxGasPrice),
			MinGasPrice:            types.CoinNumberByBigInt(e.FinanceParams.MinGasPrice),
			GasChangeRateValue:     e.FinanceParams.GasChangeRateValue,
			GasChangeRateDecimals:  e.FinanceParams.GasChangeRateDecimals,
		},
		StakeParams: types.StakeParams{
			StakeEnable:                      e.StakeParams.StakeEnable,
			MaxAddStakeRatio:                 e.StakeParams.MaxAddStakeRatio,
			MaxUnlockStakeRatio:              e.StakeParams.MaxUnlockStakeRatio,
			UnlockPeriodEpochNumber:          e.StakeParams.UnlockPeriodEpochNumber,
			MaxPendingInactiveValidatorRatio: e.StakeParams.MaxPendingInactiveValidatorRatio,
			MinDelegateStake:                 types.CoinNumberByBigInt(e.StakeParams.MinDelegateStake),
			MinValidatorStake:                types.CoinNumberByBigInt(e.StakeParams.MaxValidatorStake),
			MaxValidatorStake:                types.CoinNumberByBigInt(e.StakeParams.MaxValidatorStake),
		},
		MiscParams: types.MiscParams{
			TxMaxSize: e.MiscParams.TxMaxSize,
		},
	}
}
