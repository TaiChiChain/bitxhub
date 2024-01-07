// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.0;

interface WhiteList {
    function submit(bytes calldata addresses) external;

    function remove(bytes calldata addresses) external;

    function queryAuthInfo(bytes calldata user) external view returns (bytes memory authInfo);

    function queryWhiteListProvider(bytes calldata whiteListProviderAddr) external view returns (bytes memory whiteListProvider);
}