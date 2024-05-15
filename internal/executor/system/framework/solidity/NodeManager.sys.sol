// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

    enum Status {
        StatusSyncing,
        StatusCandidate,
        StatusActive,
        StatusPendingInactive,
        StatusExited
    }

    struct NodeMetaData {
        string name;
        string desc;
        string imageURL;
        string websiteURL;
    }

    struct NodeInfo {
        uint64 ID;
        string ConsensusPubKey;
        string P2PPubKey;
        string P2PID;
        address Operator;
        NodeMetaData MetaData;
        Status Status;
    }

    struct ConsensusVotingPower {
        uint64 NodeID;
        int64 ConsensusVotingPower;
    }

interface NodeManager {
    error IncorrectStatus(uint8 status);
    error PendingInactiveSetIsFull();

    event Register(uint64 indexed nodeID, NodeInfo info);
    event JoinedCandidateSet(uint64 indexed nodeID);
    event LeavedCandidateSet(uint64 indexed nodeID);
    event JoinedPendingInactiveSet(uint64 indexed nodeID);
    event UpdateMetaData(uint64 indexed nodeID, NodeMetaData metaData);
    event UpdateOperator(uint64 indexed nodeID, address newOperator);

    function joinCandidateSet(uint64 nodeID, uint64 commissionRate) external;

    function exit(uint64 nodeID) external;

    function updateMetaData(uint64 nodeID, NodeMetaData memory metaData) external;

    function updateOperator(uint64 nodeID, address newOperator) external;

    function getInfo(uint64 nodeID) external view returns (NodeInfo memory info);

    function getTotalCount() external view returns (uint64);

    function getInfos(uint64[] memory nodeIDs) external view returns (NodeInfo[] memory info);

    function getActiveValidatorSet() external view returns (NodeInfo[] memory info, ConsensusVotingPower[] memory votingPowers);

    function getDataSyncerSet() external view returns (NodeInfo[] memory infos);

    function getCandidateSet() external view returns (NodeInfo[] memory infos);

    function getPendingInactiveSet() external view returns (NodeInfo[] memory infos);

    function getExitedSet() external view returns (NodeInfo[] memory infos);
}