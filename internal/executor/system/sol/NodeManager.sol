// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

interface INodeManager {
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

    struct NodeInfo  {
        uint64 ID;
        string NodePubKey;
        string OperatorAddress;
        NodeMetaData MetaData;
        Status Status;
    }

    struct ConsensusVotingPower {
        uint256 NodeID;
        uint256 ConsensusVotingPower;
    }

    function joinCandidate(uint256 nodeID) external;
    function leaveValidatorSet(uint256 nodeID) external;
    function updateMetaData(uint256 nodeID, NodeMetaData memory metaData) external;
    function updateOperator(uint256 nodeID, string memory newOperatorAddress) external;

    function GetNodeInfo(uint256 nodeID) external view returns (NodeInfo memory info);
    function GetTotalNodeCount() external view returns (uint256);
    function GetNodeInfos(uint256[] memory nodeIDs) external view returns (NodeInfo[] memory info);
    function GetActiveValidatorSet() external view returns (NodeInfo[] memory info, ConsensusVotingPower[] memory votingPowers);
    function GetDataSyncerSet() external view returns (NodeInfo[] memory infos);
    function GetCandidateSet() external view returns (NodeInfo[] memory infos);
    function GetPendingInactiveSet() external view returns (NodeInfo[] memory infos);
    function GetExitedSet() external view returns (NodeInfo[] memory infos);
}