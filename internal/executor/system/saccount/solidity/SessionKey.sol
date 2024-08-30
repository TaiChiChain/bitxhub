// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.12;

struct SessionKey {
    address addr;
    uint256 spendingLimit;
    uint256 spentAmount;
    uint64 validUntil;
    uint64 validAfter;
}