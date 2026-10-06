package tx

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// defaultBaseFeeMultiplierPct is applied to a fetched base fee when
// computing GasFeeCap, so the cap stays valid as the base fee rises across
// several blocks before inclusion.
const defaultBaseFeeMultiplierPct = 200

// DynamicFeeTxConfig builds a type-2 (EIP-1559) transaction. From, To,
// Value and Data are fixed at construction; AccessList, ChainID, Nonce,
// GasTipCap, GasFeeCap and GasLimit are resolved by Pack unless set
// explicitly via the With* methods.
type DynamicFeeTxConfig struct {
	from       *types.Address
	to         *types.Address
	value      *big.Int
	data       []byte
	accessList core.AccessList

	chainID   *big.Int
	nonce     *uint64
	gasTipCap *big.Int
	gasFeeCap *big.Int
	gasLimit  *uint64

	nonceBlockTag        string
	baseFeeMultiplierPct uint64
	gasLimitBufferPct    uint64
}

// NewDynamicFeeTxConfig returns a DynamicFeeTxConfig sending value and data
// from from to to. to may be nil for a contract creation.
func NewDynamicFeeTxConfig(from, to *types.Address, value *big.Int, data []byte) *DynamicFeeTxConfig {
	return &DynamicFeeTxConfig{
		from:  from,
		to:    to,
		value: value,
		data:  data,
	}
}

// WithAccessList sets AccessList. The default is nil (empty).
func (c *DynamicFeeTxConfig) WithAccessList(accessList core.AccessList) *DynamicFeeTxConfig {
	c.accessList = accessList
	return c
}

// WithChainID sets ChainID explicitly, skipping eth_chainId in Pack.
func (c *DynamicFeeTxConfig) WithChainID(chainID *big.Int) *DynamicFeeTxConfig {
	c.chainID = chainID
	return c
}

// WithNonce sets Nonce explicitly, skipping eth_getTransactionCount in Pack.
func (c *DynamicFeeTxConfig) WithNonce(nonce uint64) *DynamicFeeTxConfig {
	c.nonce = &nonce
	return c
}

// WithNonceBlockTag sets the block parameter Pack passes to
// eth_getTransactionCount when Nonce is not set explicitly (default
// rpc.BlockTagPending).
func (c *DynamicFeeTxConfig) WithNonceBlockTag(tag string) *DynamicFeeTxConfig {
	c.nonceBlockTag = tag
	return c
}

// WithGasTipCap sets GasTipCap (maxPriorityFeePerGas) explicitly, skipping
// eth_maxPriorityFeePerGas in Pack.
func (c *DynamicFeeTxConfig) WithGasTipCap(gasTipCap *big.Int) *DynamicFeeTxConfig {
	c.gasTipCap = gasTipCap
	return c
}

// WithGasFeeCap sets GasFeeCap (maxFeePerGas) explicitly, skipping the base
// fee lookup in Pack.
func (c *DynamicFeeTxConfig) WithGasFeeCap(gasFeeCap *big.Int) *DynamicFeeTxConfig {
	c.gasFeeCap = gasFeeCap
	return c
}

// WithBaseFeeMultiplier scales the latest block's base fee by percent/100
// when computing a fetched GasFeeCap, as GasFeeCap = baseFee*percent/100 +
// GasTipCap. It has no effect on a GasFeeCap set via WithGasFeeCap. The
// default, 0, applies 200 (2x).
func (c *DynamicFeeTxConfig) WithBaseFeeMultiplier(percent uint64) *DynamicFeeTxConfig {
	c.baseFeeMultiplierPct = percent
	return c
}

// WithGasLimit sets GasLimit explicitly, skipping eth_estimateGas in Pack.
func (c *DynamicFeeTxConfig) WithGasLimit(gasLimit uint64) *DynamicFeeTxConfig {
	c.gasLimit = &gasLimit
	return c
}

// WithGasLimitBuffer scales a GasLimit fetched via eth_estimateGas by
// percent/100 (e.g. 120 for 1.2x). It has no effect on a GasLimit set via
// WithGasLimit. The default, 0, applies no buffer.
func (c *DynamicFeeTxConfig) WithGasLimitBuffer(percent uint64) *DynamicFeeTxConfig {
	c.gasLimitBufferPct = percent
	return c
}

// Pack resolves every field not already set via a With* method against
// client, in a single batch, and returns the unsigned transaction.
func (c *DynamicFeeTxConfig) Pack(ctx context.Context, client rpc.Client) (core.Transaction, error) {
	nonceBlockTag := c.nonceBlockTag
	if nonceBlockTag == "" {
		nonceBlockTag = rpc.BlockTagPending
	}

	needsBaseFee := c.gasFeeCap == nil

	var (
		chainIDHex  string
		nonceHex    string
		tipHex      string
		block       map[string]any
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
	if c.gasTipCap == nil {
		elems.With(rpc.ETHMaxPriorityFeePerGas("gasTipCap", &tipHex))
	}
	if needsBaseFee {
		elems.With(rpc.ETHGetBlockByNumber("block", rpc.BlockTagLatest, false, &block))
	}
	if c.gasLimit == nil {
		callParams := map[string]any{
			"from":  c.from.String(),
			"value": quantityHex(c.value),
			"data":  "0x" + hex.EncodeToString(c.data),
		}
		if c.to != nil {
			callParams["to"] = c.to.String()
		}
		if len(c.accessList) > 0 {
			callParams["accessList"] = accessListToRPC(c.accessList)
		}
		elems.With(rpc.ETHEstimateGas("gasLimit", callParams, rpc.BlockTagLatest, &gasLimitHex))
	}

	if elems.Len() > 0 {
		if err := client.Batch(ctx, elems); err != nil {
			return nil, fmt.Errorf("tx: pack dynamic fee tx: %w", err)
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

	gasTipCap := c.gasTipCap
	if gasTipCap == nil {
		gasTipCap, err = parseQuantity(tipHex)
		if err != nil {
			return nil, fmt.Errorf("tx: gasTipCap: %w", err)
		}
	}

	gasFeeCap := c.gasFeeCap
	if gasFeeCap == nil {
		baseFeeHex, ok := block["baseFeePerGas"].(string)
		if !ok {
			return nil, fmt.Errorf("tx: gasFeeCap: chain does not report baseFeePerGas")
		}
		baseFee, err := parseQuantity(baseFeeHex)
		if err != nil {
			return nil, fmt.Errorf("tx: gasFeeCap: %w", err)
		}

		mult := c.baseFeeMultiplierPct
		if mult == 0 {
			mult = defaultBaseFeeMultiplierPct
		}
		gasFeeCap = new(big.Int).Add(bufferBigInt(baseFee, mult), gasTipCap)
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

	return &core.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      *nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		GasLimit:   *gasLimit,
		To:         c.to,
		Value:      c.value,
		Data:       c.data,
		AccessList: c.accessList,
	}, nil
}

// accessListToRPC converts al to the JSON-RPC accessList shape:
// [{"address": "0x...", "storageKeys": ["0x...", ...]}, ...].
func accessListToRPC(al core.AccessList) []map[string]any {
	out := make([]map[string]any, len(al))
	for i, tuple := range al {
		keys := make([]string, len(tuple.StorageKeys))
		for j, k := range tuple.StorageKeys {
			keys[j] = k.String()
		}
		out[i] = map[string]any{
			"address":     tuple.Address.String(),
			"storageKeys": keys,
		}
	}
	return out
}
