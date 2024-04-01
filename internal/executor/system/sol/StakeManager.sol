// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

interface IStakeManager {
    event Stake(uint64 indexed poolID, address indexed owner, uint256 amount);

    event Unlock(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount);

    event Withdraw(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount);

    function stake(uint64 poolID, address owner, uint256 amount) external payable;

    function unlock(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) external;

    function withdraw(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) external;
}