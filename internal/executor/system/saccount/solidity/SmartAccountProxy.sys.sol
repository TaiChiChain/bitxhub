// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.12;

import "./PassKey.sol";
import "./SessionKey.sol";

contract SmartAccountProxy {
    function getOwner(address) public view returns (address) {
    }

    function getStatus(address) public view returns (uint64) {
    }

    function getGuardian(address) public view returns (address) {
    }

    function getPasskeys(address) public view returns (PassKey[] memory) {
    }

    function getSessions(address) public view returns (SessionKey[] memory) {
    }
}