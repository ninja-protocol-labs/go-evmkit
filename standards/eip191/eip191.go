// Package eip191 implements the EIP-191 personal-sign message digest.
package eip191

import (
	"strconv"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// eip191Prefix is the personal-sign prefix defined by EIP-191.
const eip191Prefix = "\x19Ethereum Signed Message:\n"

// Hash returns the Keccak-256 hash of msg with the EIP-191 personal
// message prefix ("\x19Ethereum Signed Message:\n" + len(msg)) prepended,
// matching the digest wallets sign for eth_sign/personal_sign.
func Hash(msg []byte) *types.Hash {
	prefixed := make([]byte, 0, len(eip191Prefix)+20+len(msg))
	prefixed = append(prefixed, eip191Prefix...)
	prefixed = strconv.AppendInt(prefixed, int64(len(msg)), 10)
	prefixed = append(prefixed, msg...)
	return core.Keccak256(prefixed)
}
