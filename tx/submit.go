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
	// ChainID sets ChainID explicitly on every Packer Build produces, skipping eth_chainId in Pack.
	ChainID(id *big.Int)
	// Nonce sets Nonce explicitly on every Packer Build produces, skipping eth_getTransactionCount in Pack.
	Nonce(nonce uint64)
	// NonceBlockTag sets the block parameter Pack passes to eth_getTransactionCount when Nonce isn't set explicitly.
	NonceBlockTag(tag string)
	// Limit sets GasLimit explicitly on every Packer Build produces, skipping eth_estimateGas in Pack.
	Limit(limit uint64)
	// Price sets feeCap (Legacy GasPrice or 1559 GasFeeCap) and tipCap (ignored for Legacy), skipping Pack's fee lookups.
	Price(feeCap, tipCap *big.Int)
	// LimitBuffer scales a GasLimit fetched via eth_estimateGas by
	// percent/100 on every Packer Build produces. It has no effect on an
	// explicit Limit.
	LimitBuffer(percent uint64)
	// PriceBuffer scales a fetched gas price by percent/100 on every Packer
	// Build produces: GasPrice (Legacy), or the base fee multiplier used to
	// compute GasFeeCap (DynamicFee/SetCode). It has no effect on an
	// explicit Price.
	PriceBuffer(percent uint64)
	// StateOverride sets the state override Pack passes to eth_estimateGas; it has no effect on an explicit Limit.
	StateOverride(override *rpc.StateOverride)

	// Build returns a Packer for a call to "to" with data.
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

// Broadcast sends raw via eth_sendRawTransaction and returns keccak256(raw); a failed send still succeeds if the node already knows that hash.
func Broadcast(ctx context.Context, c rpc.Client, raw []byte) (*types.Hash, error) {
	hash := core.Keccak256(raw)

	var result string
	sendErr := c.Call(ctx, rpc.ETHSendRawTransaction("send", encoding.Hex.EncodePrefixed(raw), &result))
	if sendErr == nil {
		return hash, nil
	}

	var known map[string]any
	if err := c.Call(ctx, rpc.ETHGetTransactionByHash("known", hash.String(), &known)); err == nil && known != nil {
		return hash, nil
	}
	return nil, fmt.Errorf("tx: broadcast: %w", sendErr)
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
