// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IStateEnhance {
    function enhance(address from, address to, bytes calldata originalInput) external;
}
