// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.12;

interface Ownable {
    /**
     * @dev The caller account is not authorized to perform an operation.
     */
    error OwnableUnauthorizedAccount(address account);

    /**
     * @dev The owner is not a valid owner account. (eg. `address(0)`)
     */
    error OwnableInvalidOwner(address owner);

    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);


    /**
     * @dev Returns the address of the current owner.
     */
    function owner() external view returns (address);


    /**
     * @dev Transfers ownership of the contract to a new account (`newOwner`).
     *
     * Emits an {OwnershipTransferred} event.
     *
     * Requirements:
     *
     * - `newOwner` cannot be the zero address.
     */
    function transferOwnership(address newOwner) external;
}
