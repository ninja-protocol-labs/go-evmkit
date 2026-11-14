// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v5.5.0) (interfaces/IERC6909.sol)

pragma solidity >=0.6.2;

import {IERC6909} from "./IERC6909.sol";

/**
 * @dev Optional extension of {IERC6909} that adds metadata functions.
 */
interface IERC6909Metadata is IERC6909 {
    /**
     * @dev Returns the name of the token of type `id`.
     */
    function name(uint256 id) external view returns (string memory);

    /**
     * @dev Returns the ticker symbol of the token of type `id`.
     */
    function symbol(uint256 id) external view returns (string memory);

    /**
     * @dev Returns the number of decimals for the token of type `id`.
     */
    function decimals(uint256 id) external view returns (uint8);
}
