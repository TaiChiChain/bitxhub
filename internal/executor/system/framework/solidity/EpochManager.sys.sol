// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.7.0 <0.9.0;

    struct ConsensusParams {
        string ProposerElectionType;
        uint64 CheckpointPeriod;
        uint64 HighWatermarkCheckpointPeriod;
        uint64 MaxValidatorNum;
        uint64 BlockMaxTxNum;
        bool EnableTimedGenEmptyBlock;
        int64 NotActiveWeight;
        uint64 AbnormalNodeExcludeView;
        uint64 AgainProposeIntervalBlockInValidatorsNumPercentage;
        uint64 ContinuousNullRequestToleranceNumber;
        uint64 ReBroadcastToleranceNumber;
    }

    struct FinanceParams {
        uint64 GasLimit;
        uint64 MinGasPrice;
    }

    struct MiscParams {
        uint64 TxMaxSize;
    }

    struct EpochInfo {
        uint64 Version;
        uint64 Epoch;
        uint64 EpochPeriod;
        uint64 StartBlock;
        ConsensusParams ConsensusParams;
        FinanceParams FinanceParams;
        MiscParams MiscParams;
    }

interface EpochManager {
    function currentEpoch() external view returns (EpochInfo memory epochInfo);

    function nextEpoch() external view returns (EpochInfo memory epochInfo);

    function historyEpoch(uint64 epochID)
    external
    view
    returns (EpochInfo memory epochInfo);
}
