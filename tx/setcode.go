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

// perEmptyAccountCost is EIP-7702's PER_EMPTY_ACCOUNT_COST: the upfront
// intrinsic gas charged per authorization tuple, regardless of whether the
// authority account turns out to be empty (a partial refund applies after
// execution, but the upfront charge — what eth_estimateGas must clear — is
// always this). Most nodes' eth_estimateGas does not yet account for it,
// so Pack adds it on top of the estimate itself.
const perEmptyAccountCost = 25000

// SetCodeTxConfig builds a type-4 (EIP-7702) transaction. From, To, Value,
// Data and AuthorizationList are fixed at construction (To and
// AuthorizationList are required, per EIP-7702); AccessList, ChainID,
// Nonce, GasTipCap, GasFeeCap and GasLimit are resolved by Pack unless set
// explicitly via the With* methods.
type SetCodeTxConfig struct {
	from          *types.Address
	to            *types.Address
	value         *big.Int
	data          []byte
	accessList    core.AccessList
	authorization []core.Authorization

	chainID   *big.Int
	nonce     *uint64
	gasTipCap *big.Int
	gasFeeCap *big.Int
	gasLimit  *uint64

	nonceBlockTag        string
	baseFeeMultiplierPct uint64
	gasLimitBufferPct    uint64
}

// NewSetCodeTxConfig returns a SetCodeTxConfig sending value and data from
// from to to, delegating per authorizationList. to and authorizationList
// must not be nil/empty: EIP-7702 has no contract creation and requires at
// least one authorization.
func NewSetCodeTxConfig(from, to *types.Address, value *big.Int, data []byte, authorizationList []core.Authorization) *SetCodeTxConfig {
	return &SetCodeTxConfig{
		from:          from,
		to:            to,
		value:         value,
		data:          data,
		authorization: authorizationList,
	}
}

// WithAccessList sets AccessList. The default is nil (empty).
func (c *SetCodeTxConfig) WithAccessList(accessList core.AccessList) *SetCodeTxConfig {
	c.accessList = accessList
	return c
}

// WithChainID sets ChainID explicitly, skipping eth_chainId in Pack.
func (c *SetCodeTxConfig) WithChainID(chainID *big.Int) *SetCodeTxConfig {
	c.chainID = chainID
	return c
}

// WithNonce sets Nonce explicitly, skipping eth_getTransactionCount in Pack.
func (c *SetCodeTxConfig) WithNonce(nonce uint64) *SetCodeTxConfig {
	c.nonce = &nonce
	return c
}

// WithNonceBlockTag sets the block parameter Pack passes to
// eth_getTransactionCount when Nonce is not set explicitly (default
// rpc.BlockTagPending).
func (c *SetCodeTxConfig) WithNonceBlockTag(tag string) *SetCodeTxConfig {
	c.nonceBlockTag = tag
	return c
}

// WithGasTipCap sets GasTipCap (maxPriorityFeePerGas) explicitly, skipping
// eth_maxPriorityFeePerGas in Pack.
func (c *SetCodeTxConfig) WithGasTipCap(gasTipCap *big.Int) *SetCodeTxConfig {
	c.gasTipCap = gasTipCap
	return c
}

// WithGasFeeCap sets GasFeeCap (maxFeePerGas) explicitly, skipping the base
// fee lookup in Pack.
func (c *SetCodeTxConfig) WithGasFeeCap(gasFeeCap *big.Int) *SetCodeTxConfig {
	c.gasFeeCap = gasFeeCap
	return c
}

// WithBaseFeeMultiplier scales the latest block's base fee by percent/100
// when computing a fetched GasFeeCap, as GasFeeCap = baseFee*percent/100 +
// GasTipCap. It has no effect on a GasFeeCap set via WithGasFeeCap. The
// default, 0, applies 200 (2x).
func (c *SetCodeTxConfig) WithBaseFeeMultiplier(percent uint64) *SetCodeTxConfig {
	c.baseFeeMultiplierPct = percent
	return c
}

// WithGasLimit sets GasLimit explicitly, skipping eth_estimateGas in Pack.
func (c *SetCodeTxConfig) WithGasLimit(gasLimit uint64) *SetCodeTxConfig {
	c.gasLimit = &gasLimit
	return c
}

// WithGasLimitBuffer scales a GasLimit fetched via eth_estimateGas by
// percent/100 (e.g. 120 for 1.2x). It has no effect on a GasLimit set via
// WithGasLimit. The default, 0, applies no buffer.
func (c *SetCodeTxConfig) WithGasLimitBuffer(percent uint64) *SetCodeTxConfig {
	c.gasLimitBufferPct = percent
	return c
}

// Pack resolves every field not already set via a With* method against
// client, in a single batch, and returns the unsigned transaction.
func (c *SetCodeTxConfig) Pack(ctx context.Context, client rpc.Client) (core.Transaction, error) {
	if c.to == nil {
		return nil, fmt.Errorf("tx: pack set code tx: to is required")
	}
	if len(c.authorization) == 0 {
		return nil, fmt.Errorf("tx: pack set code tx: authorizationList is required")
	}

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
			"to":    c.to.String(),
			"value": quantityHex(c.value),
			"data":  "0x" + hex.EncodeToString(c.data),
		}
		if len(c.accessList) > 0 {
			callParams["accessList"] = accessListToRPC(c.accessList)
		}
		elems.With(rpc.ETHEstimateGas("gasLimit", callParams, rpc.BlockTagLatest, &gasLimitHex))
	}

	if elems.Len() > 0 {
		if err := client.Batch(ctx, elems); err != nil {
			return nil, fmt.Errorf("tx: pack set code tx: %w", err)
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
		v := bufferUint64(n.Uint64(), c.gasLimitBufferPct) + perEmptyAccountCost*uint64(len(c.authorization))
		gasLimit = &v
	}

	return &core.SetCodeTx{
		ChainID:    chainID,
		Nonce:      *nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		GasLimit:   *gasLimit,
		To:         c.to,
		Value:      c.value,
		Data:       c.data,
		AccessList: c.accessList,
		AuthList:   c.authorization,
	}, nil
}
