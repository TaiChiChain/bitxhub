// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

interface AXC {
    /**
     * @dev Returns the value of tokens in existence.
     */
    function totalSupply() external view returns (uint256);

    /**
    * @notice This method can only be called through the governance contract
    * @param amount How much should be minted for token contract account
    * @dev Returns the success status of the mint operation.
    */
    function mint(uint256 amount) external returns (bool);

    /**
    * @notice This method can only be called through the governance contract
    * @param amount How much should be burned
    */
    function burn(uint256 amount) external;
}