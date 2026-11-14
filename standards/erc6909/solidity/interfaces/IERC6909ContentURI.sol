// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v5.5.0) (interfaces/IERC6909.sol)

pragma solidity >=0.6.2;

import {IERC6909} from "./IERC6909.sol";

/**
 * @dev Optional extension of {IERC6909} that adds content URI functions.
 */
interface IERC6909ContentURI is IERC6909 {
    /**
     * @dev Returns URI for the contract.
     */
    function contractURI() external view returns (string memory);

    /**
     * @dev Returns the URI for the token of type `id`.
     */
    function tokenURI(uint256 id) external view returns (string memory);
}
