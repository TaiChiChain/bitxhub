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
    string OperatorAddress;
    NodeMetaData MetaData;
    Status Status;
}

struct ConsensusVotingPower {
    uint64 NodeID;
    int64  ConsensusVotingPower;
}

interface NodeManager {
    error incorrectStatus(uint8 status);
    error pendingInactiveSetIsFull();

    event Registered(uint64 indexed nodeID);
    event JoinedCandidateSet(uint64 indexed nodeID);
    event LeavedCandidateSet(uint64 indexed nodeID);
    event JoinedPendingInactiveSet(uint64 indexed nodeID);
    event UpdateMetaData(uint64 indexed nodeID, NodeMetaData metaData);
    event UpdateOperator(uint64 indexed nodeID, string newOperatorAddress);

    function joinCandidateSet(uint64 nodeID) external;

    function leaveValidatorOrCandidateSet(uint64 nodeID) external;

    function updateMetaData(uint64 nodeID, NodeMetaData memory metaData) external;

    function updateOperator(uint64 nodeID, string memory newOperatorAddress) external;

    function GetNodeInfo(uint64 nodeID) external view returns (NodeInfo memory info);

    function GetTotalNodeCount() external view returns (uint64);

    function GetNodeInfos(uint64[] memory nodeIDs) external view returns (NodeInfo[] memory info);

    function GetActiveValidatorSet() external view returns (NodeInfo[] memory info, ConsensusVotingPower[] memory votingPowers);

    function GetDataSyncerSet() external view returns (NodeInfo[] memory infos);

    function GetCandidateSet() external view returns (NodeInfo[] memory infos);

    function GetPendingInactiveSet() external view returns (NodeInfo[] memory infos);

    function GetExitedSet() external view returns (NodeInfo[] memory infos);
}