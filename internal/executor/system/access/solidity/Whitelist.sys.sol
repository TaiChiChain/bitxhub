// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.0;

    enum UserRole {Basic, Super}

    struct AuthInfo {
        address addr;
        address[] providers;
        UserRole role;
    }

    struct ProviderInfo {
        address addr;
    }

interface Whitelist {
    event Submit(address indexed submitter, address[] addresses);

    event Remove(address indexed submitter, address[] addresses);

    event UpdateProviders(bool isAdd, ProviderInfo[] providers);

    function submit(address[] calldata addresses) external;

    function remove(address[] calldata addresses) external;

    function queryAuthInfo(address addr) external view returns (AuthInfo memory authInfo);

    function queryProviderInfo(address addr) external view returns (ProviderInfo memory providerInfo);
}