// Package tx builds ready-to-sign transactions, resolving whatever fields
// the caller did not supply from the chain through an rpc.Client.
//
// Building a transaction goes through three stages:
//
//   - Builder configures a transaction without committing to a call target
//     yet (nonce, gas limit/price and their buffers, ...), and without the
//     caller needing to know which concrete transaction type is in use.
//     NewLegacyBuilder, NewDynamicFeeBuilder and NewSetCodeBuilder each
//     return one. Build(from, to, value, data) then fixes the call target
//     and returns a Packer.
//   - Packer resolves every field the Builder didn't set explicitly
//     against an rpc.Client, in a single batch, and returns the
//     signed-ready transaction. *LegacyTxConfig, *DynamicFeeTxConfig and
//     *SetCodeTxConfig each implement it, so a Packer can also be built
//     directly (NewLegacyTxConfig, NewDynamicFeeTxConfig,
//     NewSetCodeTxConfig) when the caller already knows which type it
//     wants and has from/to/value/data in hand.
//   - core.Transaction is the result: sign it, RLP-encode it and broadcast
//     it (or call Submit to do all three, plus Pack, in one step).
//
// Contract packages (e.g. contracts/erc20) accept a Builder so their own
// callers can pick a transaction type without the contract package
// hardcoding one.
package tx

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type LegacyBuilder struct {
	chainid           *big.Int
	nonce             *uint64
	nonceBlockTag     string
	gasLimit          *uint64
	gasPrice          *big.Int
	gasLimitBufferPct uint64
	gasPriceBufferPct uint64
}

// NewLegacyBuilder returns a Builder producing a LegacyTxConfig.
func NewLegacyBuilder() *LegacyBuilder {
	return &LegacyBuilder{}
}

// ChainID carries id onto every LegacyTxConfig produced by Build.
func (b *LegacyBuilder) ChainID(id *big.Int) {
	b.chainid = id
}

// Nonce carries nonce onto every LegacyTxConfig produced by Build.
func (b *LegacyBuilder) Nonce(nonce uint64) {
	b.nonce = &nonce
}

// NonceBlockTag carries tag onto every LegacyTxConfig produced by Build.
func (b *LegacyBuilder) NonceBlockTag(tag string) {
	b.nonceBlockTag = tag
}

// Limit carries limit onto every LegacyTxConfig produced by Build.
func (b *LegacyBuilder) Limit(limit uint64) {
	b.gasLimit = &limit
}

// Price carries feeCap onto every LegacyTxConfig produced by Build as
// GasPrice. tipCap has no equivalent on a Legacy transaction and is ignored.
func (b *LegacyBuilder) Price(feeCap, _ *big.Int) {
	b.gasPrice = feeCap
}

// LimitBuffer carries percent onto every LegacyTxConfig produced by Build
// as its gas limit buffer (see LegacyTxConfig.WithGasLimitBuffer).
func (b *LegacyBuilder) LimitBuffer(percent uint64) {
	b.gasLimitBufferPct = percent
}

// PriceBuffer carries percent onto every LegacyTxConfig produced by Build
// as its gas price buffer (see LegacyTxConfig.WithGasPriceBuffer).
func (b *LegacyBuilder) PriceBuffer(percent uint64) {
	b.gasPriceBufferPct = percent
}

func (b *LegacyBuilder) Build(from, to *types.Address, value *big.Int, data []byte) Packer {
	c := NewLegacyTxConfig(from, to, value, data).
		WithChainID(b.chainid).
		WithGasLimitBuffer(b.gasLimitBufferPct).
		WithGasPriceBuffer(b.gasPriceBufferPct)

	if b.nonce != nil {
		c = c.WithNonce(*b.nonce)
	}

	if b.nonceBlockTag != "" {
		c = c.WithNonceBlockTag(b.nonceBlockTag)
	}

	if b.gasLimit != nil {
		c = c.WithGasLimit(*b.gasLimit)
	}

	if b.gasPrice != nil {
		c = c.WithGasPrice(b.gasPrice)
	}

	return c
}

// LegacyTxConfig builds a type-0 transaction. From, To, Value and Data are
// fixed at construction; ChainID, Nonce, GasPrice and GasLimit are
// resolved by Pack unless set explicitly via the With* methods.
type LegacyTxConfig struct {
	from  *types.Address
	to    *types.Address
	value *big.Int
	data  []byte

	chainID  *big.Int
	nonce    *uint64
	gasPrice *big.Int
	gasLimit *uint64

	nonceBlockTag     string
	gasPriceBufferPct uint64
	gasLimitBufferPct uint64
}

// NewLegacyTxConfig returns a LegacyTxConfig sending value and data from
// from to to. to may be nil for a contract creation.
func NewLegacyTxConfig(from, to *types.Address, value *big.Int, data []byte) *LegacyTxConfig {
	return &LegacyTxConfig{
		from:  from,
		to:    to,
		value: value,
		data:  data,
	}
}

// WithChainID sets ChainID explicitly, skipping eth_chainId in Pack.
func (c *LegacyTxConfig) WithChainID(chainID *big.Int) *LegacyTxConfig {
	c.chainID = chainID
	return c
}

// WithNonce sets Nonce explicitly, skipping eth_getTransactionCount in Pack.
func (c *LegacyTxConfig) WithNonce(nonce uint64) *LegacyTxConfig {
	c.nonce = &nonce
	return c
}

// WithNonceBlockTag sets the block parameter Pack passes to
// eth_getTransactionCount when Nonce is not set explicitly (default
// rpc.BlockTagPending).
func (c *LegacyTxConfig) WithNonceBlockTag(tag string) *LegacyTxConfig {
	c.nonceBlockTag = tag
	return c
}

// WithGasPrice sets GasPrice explicitly, skipping eth_gasPrice in Pack.
func (c *LegacyTxConfig) WithGasPrice(gasPrice *big.Int) *LegacyTxConfig {
	c.gasPrice = gasPrice
	return c
}

// WithGasPriceBuffer scales a GasPrice fetched via eth_gasPrice by
// percent/100 (e.g. 150 for 1.5x). It has no effect on a GasPrice set via
// WithGasPrice. The default, 0, applies no buffer.
func (c *LegacyTxConfig) WithGasPriceBuffer(percent uint64) *LegacyTxConfig {
	c.gasPriceBufferPct = percent
	return c
}

// WithGasLimit sets GasLimit explicitly, skipping eth_estimateGas in Pack.
func (c *LegacyTxConfig) WithGasLimit(gasLimit uint64) *LegacyTxConfig {
	c.gasLimit = &gasLimit
	return c
}

// WithGasLimitBuffer scales a GasLimit fetched via eth_estimateGas by
// percent/100 (e.g. 120 for 1.2x). It has no effect on a GasLimit set via
// WithGasLimit. The default, 0, applies no buffer.
func (c *LegacyTxConfig) WithGasLimitBuffer(percent uint64) *LegacyTxConfig {
	c.gasLimitBufferPct = percent
	return c
}

// Pack resolves every field not already set via a With* method against
// client, in a single batch, and returns the unsigned transaction.
func (c *LegacyTxConfig) Pack(ctx context.Context, client rpc.Client) (core.Transaction, error) {
	nonceBlockTag := c.nonceBlockTag
	if nonceBlockTag == "" {
		nonceBlockTag = rpc.BlockTagPending
	}

	var (
		chainIDHex  string
		nonceHex    string
		gasPriceHex string
		gasLimitHex string
		err         error
	)
	elems := rpc.Elements{}
	if c.chainID == nil {
		elems.With(rpc.ETHChainID("chainId", &chainIDHex))
	}
	if c.nonce == nil {
		elems.With(rpc.ETHGetTransactionCount("nonce", c.from.String(), nonceBlockTag, &nonceHex))
	}
	if c.gasPrice == nil {
		elems.With(rpc.ETHGasPrice("gasPrice", &gasPriceHex))
	}
	if c.gasLimit == nil {
		callParams := map[string]any{
			"from":  c.from.String(),
			"value": quantityHex(c.value),
			"data":  encoding.Hex.EncodePrefixed(c.data),
		}
		if c.to != nil {
			callParams["to"] = c.to.String()
		}
		elems.With(rpc.ETHEstimateGas("gasLimit", callParams, rpc.BlockTagLatest, &gasLimitHex))
	}

	if elems.Len() > 0 {
		if err := client.Batch(ctx, elems); err != nil {
			return nil, fmt.Errorf("tx: pack legacy tx: %w", err)
		}
	}

	chainID := c.chainID
	if chainID == nil {
		chainID, err = parseQuantity(chainIDHex)
		if err != nil {
			return nil, fmt.Errorf("tx: chainId: %w", err)
		}
	}

	nonce := c.nonce
	if nonce == nil {
		n, err := parseQuantity(nonceHex)
		if err != nil {
			return nil, fmt.Errorf("tx: nonce: %w", err)
		}
		v := n.Uint64()
		nonce = &v
	}

	gasPrice := c.gasPrice
	if gasPrice == nil {
		gasPrice, err = parseQuantity(gasPriceHex)
		if err != nil {
			return nil, fmt.Errorf("tx: gasPrice: %w", err)
		}
		gasPrice = bufferBigInt(gasPrice, c.gasPriceBufferPct)
	}

	gasLimit := c.gasLimit
	if gasLimit == nil {
		n, err := parseQuantity(gasLimitHex)
		if err != nil {
			return nil, fmt.Errorf("tx: gasLimit: %w", err)
		}
		v := bufferUint64(n.Uint64(), c.gasLimitBufferPct)
		gasLimit = &v
	}

	return &core.LegacyTx{
		ChainID:  chainID,
		Nonce:    *nonce,
		GasPrice: gasPrice,
		GasLimit: *gasLimit,
		To:       c.to,
		Value:    c.value,
		Data:     c.data,
	}, nil
}
