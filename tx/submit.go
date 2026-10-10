package tx

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// Builder builds a Packer for a call to "to" with data, letting callers
// (e.g. contract packages) pick a transaction type without hardcoding one.
type Builder interface {
	Build(from, to *types.Address, value *big.Int, data []byte) Packer
}

var (
	_ Builder = (*LegacyBuilder)(nil)
	_ Builder = (*DynamicFeeBuilder)(nil)
	_ Builder = (*SetCodeBuilder)(nil)
)

// Packer is implemented by *LegacyTxConfig, *DynamicFeeTxConfig and
// *SetCodeTxConfig.
type Packer interface {
	Gas(context.Context, rpc.Client) (uint64, error)
	Pack(context.Context, rpc.Client) (core.Transaction, error)
}

var (
	_ Packer = (*LegacyTxConfig)(nil)
	_ Packer = (*DynamicFeeTxConfig)(nil)
	_ Packer = (*SetCodeTxConfig)(nil)
)

// Submit packs cfg, signs the result with signer, and broadcasts it,
// returning its hash. Callers who need to stop partway through (sign now,
// broadcast later; inspect the raw bytes first) should call Pack, Sign,
// EncodeRLP and Broadcast individually instead.
func Submit(ctx context.Context, c rpc.Client, p Packer, s core.Signer) (*types.Hash, error) {
	txn, err := p.Pack(ctx, c)
	if err != nil {
		return nil, err
	}

	if err = txn.Sign(s); err != nil {
		return nil, fmt.Errorf("tx: sign: %w", err)
	}

	raw, err := txn.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("tx: encode: %w", err)
	}

	return Broadcast(ctx, c, raw)
}

// Broadcast sends a signed, RLP-encoded transaction via
// eth_sendRawTransaction and returns its hash, keccak256(raw).
func Broadcast(ctx context.Context, c rpc.Client, raw []byte) (*types.Hash, error) {
	var result string
	if err := c.Call(ctx, rpc.ETHSendRawTransaction("send", encoding.Hex.EncodePrefixed(raw), &result)); err != nil {
		return nil, fmt.Errorf("tx: broadcast: %w", err)
	}
	return core.Keccak256(raw), nil
}

// parseQuantity parses a JSON-RPC quantity (a "0x"-prefixed hex integer,
// e.g. "0x1a") into a big.Int.
func parseQuantity(hex string) (*big.Int, error) {
	trimmed := strings.TrimPrefix(hex, "0x")
	if trimmed == "" {
		trimmed = "0"
	}

	n, ok := new(big.Int).SetString(trimmed, 16)
	if !ok {
		return nil, fmt.Errorf("tx: invalid quantity %q", hex)
	}
	return n, nil
}

// quantityHex formats n as a JSON-RPC quantity. A nil n is treated as zero.
func quantityHex(n *big.Int) string {
	if n == nil || n.Sign() == 0 {
		return "0x0"
	}
	return "0x" + n.Text(16)
}

// bufferBigInt returns n*percent/100, or n unchanged when percent is 0 or 100.
func bufferBigInt(n *big.Int, percent uint64) *big.Int {
	if percent == 0 || percent == 100 {
		return n
	}

	buffered := new(big.Int).Mul(n, new(big.Int).SetUint64(percent))
	return buffered.Div(buffered, big.NewInt(100))
}

// bufferUint64 returns n*percent/100, or n unchanged when percent is 0 or 100.
func bufferUint64(n uint64, percent uint64) uint64 {
	if percent == 0 || percent == 100 {
		return n
	}
	return n * percent / 100
}

// estimateGasLimit calls eth_estimateGas directly (not batched with any
// other element) and applies bufferPct, for a Config's Gas method.
func estimateGasLimit(ctx context.Context, client rpc.Client, callParams map[string]any, bufferPct uint64) (uint64, error) {
	var gasLimitHex string
	if err := client.Call(ctx, rpc.ETHEstimateGas("gasLimit", callParams, rpc.BlockTagLatest, &gasLimitHex)); err != nil {
		return 0, fmt.Errorf("tx: gas: %w", err)
	}

	n, err := parseQuantity(gasLimitHex)
	if err != nil {
		return 0, fmt.Errorf("tx: gas: %w", err)
	}
	return bufferUint64(n.Uint64(), bufferPct), nil
}
