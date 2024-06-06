// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.12;

import "./IPaymaster.sys.sol";

abstract contract TokenPaymaster is IPaymaster {
    function addToken(address token, address oracle) virtual external;

    function getToken(address token) virtual external view returns (address);
}
