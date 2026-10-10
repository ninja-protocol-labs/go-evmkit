// Package tx builds ready-to-sign transactions, resolving whatever fields
// the caller did not supply from the chain through an rpc.Client.
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

type LegacyBuilder struct{}

// NewLegacyBuilder returns a Builder producing a LegacyTxConfig.
func NewLegacyBuilder() Builder {
	return &LegacyBuilder{}
}

func (_ *LegacyBuilder) Build(from, to *types.Address, value *big.Int, data []byte) Packer {
	return NewLegacyTxConfig(from, to, value, data)
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

// Gas returns GasLimit, estimating and caching it via eth_estimateGas if
// not already set (via WithGasLimit or an earlier Gas call) — so a later
// Pack skips estimating it again.
func (c *LegacyTxConfig) Gas(ctx context.Context, client rpc.Client) (uint64, error) {
	if c.gasLimit != nil {
		return *c.gasLimit, nil
	}

	callParams := map[string]any{
		"from":  c.from.String(),
		"value": quantityHex(c.value),
		"data":  encoding.Hex.EncodePrefixed(c.data),
	}
	if c.to != nil {
		callParams["to"] = c.to.String()
	}

	limit, err := estimateGasLimit(ctx, client, callParams, c.gasLimitBufferPct)
	if err != nil {
		return 0, err
	}
	c.gasLimit = &limit
	return limit, nil
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
