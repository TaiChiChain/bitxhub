// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

    struct LiquidStakingTokenRate {
        uint256 StakeAmount;
        uint256 LiquidStakingTokenAmount;
    }

    struct PoolInfo {
        uint64 ID;
        bool IsActive;
        uint256 ActiveStake;
        uint256 TotalLiquidStakingToken;
        uint256 PendingActiveStake;
        uint256 PendingInactiveStake;
        uint256 PendingInactiveLiquidStakingTokenAmount;
        uint64 CommissionRate;
        uint64 NextEpochCommissionRate;
        uint256 LastEpochReward;
        uint256 LastEpochCommission;
        uint256 CumulativeReward;
        uint256 CumulativeCommission;
        uint256 OperatorLiquidStakingTokenID;
        uint64 LastRateEpoch;
    }

interface StakingManager {
    event AddStake(uint64 indexed poolID, address indexed owner, uint256 amount, uint256 liquidStakingTokenID);

    event Unlock(uint256 liquidStakingTokenID, uint256 amount, uint64 unlockTimestamp);

    event Withdraw(uint256 liquidStakingTokenID, address indexed recipient, uint256 amount);

    // owner will be lst owner
    function addStake(uint64 poolID, address owner, uint256 amount) external payable;

    // called by lst owner
    function unlock(uint256 liquidStakingTokenID, uint256 amount) external;

    // called by lst owner
    function withdraw(uint256 liquidStakingTokenID, address recipient, uint256 amount) external;

    // called by lst owner
    function batchUnlock(uint256[] memory liquidStakingTokenIDs, uint256[] memory amounts) external;

    // called by lst owner
    function batchWithdraw(uint256[] memory liquidStakingTokenIDs, address recipient, uint256[] memory amounts) external;

    // called by operator, newCommissionRate range: [0,10000]
    function updatePoolCommissionRate(uint64 poolID, uint64 newCommissionRate) external;

    function getPoolInfo(uint64 poolID) external view returns (PoolInfo memory poolInfo);

    function getPoolInfos(uint64[] memory poolIDs) external view returns (PoolInfo[] memory poolInfos);

    function getPoolHistoryLiquidStakingTokenRate(uint64 poolID, uint64 epoch) external view returns (LiquidStakingTokenRate memory poolHistoryLiquidStakingTokenRate);

    function getPoolHistoryLiquidStakingTokenRates(uint64[] memory poolIDs, uint64 epoch) external view returns (LiquidStakingTokenRate[] memory poolHistoryLiquidStakingTokenRate);
}