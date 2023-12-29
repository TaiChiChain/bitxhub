// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.7.0 <0.9.0;

    struct ConsensusParams {
        string ValidatorElectionType;
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
        bool StartGasPriceAvailable;
        uint64 StartGasPrice;
        uint64 MaxGasPrice;
        uint64 MinGasPrice;
        uint64 GasChangeRateValue;
        uint64 GasChangeRateDecimals;
    }

    struct MiscParams {
        uint64 TxMaxSize;
    }

    struct NodeInfo {
        uint64 ID;
        string AccountAddress;
        string P2PNodeID;
        int64 ConsensusVotingPower;
    }

    struct EpochInfo {
        uint64 Version;
        uint64 Epoch;
        uint64 EpochPeriod;
        uint64 StartBlock;
        string[] P2PBootstrapNodeAddresses;
        ConsensusParams ConsensusParams;
        FinanceParams FinanceParams;
        MiscParams MiscParams;
        NodeInfo[] ValidatorSet;
        NodeInfo[] CandidateSet;
        NodeInfo[] DataSyncerSet;
    }

interface EpochManager {
    function currentEpoch() external view returns (EpochInfo memory epochInfo);

    function nextEpoch() external view returns (EpochInfo memory epochInfo);

    function historyEpoch(uint64 epochID)
    external
    view
    returns (EpochInfo memory epochInfo);
}
