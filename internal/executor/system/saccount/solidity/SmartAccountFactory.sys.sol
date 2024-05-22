// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.12;

import "./SmartAccount.sys.sol";
import "./Ownable.sys.sol";

/**
 * A sample factory contract for SmartAccount
 * A UserOperations "initCode" holds the address of the factory, and a method call (to createAccount, in this sample factory).
 * The factory's createAccount returns the target account address even if it is already installed.
 * This way, the entryPoint.getSenderAddress() can be called either before or after the account is created.
 */
abstract contract SmartAccountFactory is Ownable {
    constructor(IEntryPoint _entryPoint) {
    }

    /**
     * create an account, and return its address.
     * returns the address even if the account is already deployed.
     * Note that during UserOperation execution, this method is called only if the account is not deployed.
     * This method returns an existing account address so that entryPoint.getSenderAddress() would work even after account creation
     */
    function createAccount(address owner,uint256 salt) public returns (SmartAccount ret) {
    }

    /**
     * calculate the counterfactual address of this account as it would be returned by createAccount()
     */
    function getAddress(address owner,uint256 salt) public view returns (address) {
    }

    // set another account factory to get address
    function setAccountFactory(address _factory) external {
    }

    function getAccountFactory() external view returns (address) {
    }
}
