package tx

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type SetCodeBuilder struct {
	authorizationList    []core.Authorization
	chainid              *big.Int
	accessList           core.AccessList
	nonce                *uint64
	nonceBlockTag        string
	gasLimit             *uint64
	gasFeeCap            *big.Int
	gasTipCap            *big.Int
	gasLimitBufferPct    uint64
	stateOverride        *rpc.StateOverride
	baseFeeMultiplierPct uint64
}

// NewSetCodeBuilder returns a Builder producing a SetCodeTxConfig with the given authorization list.
func NewSetCodeBuilder(authorizationList []core.Authorization) *SetCodeBuilder {
	return &SetCodeBuilder{
		authorizationList: authorizationList,
	}
}

// AccessList carries list onto every SetCodeTxConfig produced by Build (not part of the Builder interface).
func (b *SetCodeBuilder) AccessList(list core.AccessList) {
	b.accessList = list
}

// ChainID carries id onto every SetCodeTxConfig produced by Build.
func (b *SetCodeBuilder) ChainID(id *big.Int) {
	b.chainid = id
}

// Nonce carries nonce onto every SetCodeTxConfig produced by Build.
func (b *SetCodeBuilder) Nonce(nonce uint64) {
	b.nonce = &nonce
}

// NonceBlockTag carries tag onto every SetCodeTxConfig produced by Build.
func (b *SetCodeBuilder) NonceBlockTag(tag string) {
	b.nonceBlockTag = tag
}

// Limit carries limit onto every SetCodeTxConfig produced by Build.
func (b *SetCodeBuilder) Limit(limit uint64) {
	b.gasLimit = &limit
}

// Price carries feeCap and tipCap onto every SetCodeTxConfig produced by Build as GasFeeCap and GasTipCap.
func (b *SetCodeBuilder) Price(feeCap, tipCap *big.Int) {
	b.gasFeeCap = feeCap
	b.gasTipCap = tipCap
}

// LimitBuffer carries percent onto every SetCodeTxConfig produced by Build as its gas limit buffer.
func (b *SetCodeBuilder) LimitBuffer(percent uint64) {
	b.gasLimitBufferPct = percent
}

// StateOverride carries override onto every SetCodeTxConfig produced by Build as the eth_estimateGas state override.
func (b *SetCodeBuilder) StateOverride(override *rpc.StateOverride) {
	b.stateOverride = override
}

// PriceBuffer carries percent onto every SetCodeTxConfig produced by Build as its base fee multiplier.
func (b *SetCodeBuilder) PriceBuffer(percent uint64) {
	b.baseFeeMultiplierPct = percent
}

func (b *SetCodeBuilder) Build(from, to *types.Address, value *big.Int, data []byte) Packer {
	c := NewSetCodeTxConfig(from, to, value, data, b.authorizationList).
		WithChainID(b.chainid).
		WithAccessList(b.accessList).
		WithGasLimitBuffer(b.gasLimitBufferPct).
		WithStateOverride(b.stateOverride).
		WithBaseFeeMultiplier(b.baseFeeMultiplierPct)

	if b.nonce != nil {
		c = c.WithNonce(*b.nonce)
	}

	if b.nonceBlockTag != "" {
		c = c.WithNonceBlockTag(b.nonceBlockTag)
	}

	if b.gasLimit != nil {
		c = c.WithGasLimit(*b.gasLimit)
	}

	if b.gasFeeCap != nil {
		c = c.WithGasFeeCap(b.gasFeeCap)
	}

	if b.gasTipCap != nil {
		c = c.WithGasTipCap(b.gasTipCap)
	}

	return c
}

// SetCodeTxConfig builds a type-4 (EIP-7702) transaction; fields not set via With* are resolved by Pack.
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
	stateOverride        *rpc.StateOverride
}

// NewSetCodeTxConfig returns a SetCodeTxConfig from from to to; to and a non-empty authorizationList are required.
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

// WithNonceBlockTag sets the eth_getTransactionCount block tag used when Nonce is unset (default pending).
func (c *SetCodeTxConfig) WithNonceBlockTag(tag string) *SetCodeTxConfig {
	c.nonceBlockTag = tag
	return c
}

// WithGasTipCap sets GasTipCap explicitly, skipping eth_maxPriorityFeePerGas in Pack.
func (c *SetCodeTxConfig) WithGasTipCap(gasTipCap *big.Int) *SetCodeTxConfig {
	c.gasTipCap = gasTipCap
	return c
}

// WithGasFeeCap sets GasFeeCap explicitly, skipping the base fee lookup in Pack.
func (c *SetCodeTxConfig) WithGasFeeCap(gasFeeCap *big.Int) *SetCodeTxConfig {
	c.gasFeeCap = gasFeeCap
	return c
}

// WithBaseFeeMultiplier sets GasFeeCap = baseFee*percent/100 + GasTipCap when fetched (default 200).
func (c *SetCodeTxConfig) WithBaseFeeMultiplier(percent uint64) *SetCodeTxConfig {
	c.baseFeeMultiplierPct = percent
	return c
}

// WithGasLimit sets GasLimit explicitly, skipping eth_estimateGas in Pack.
func (c *SetCodeTxConfig) WithGasLimit(gasLimit uint64) *SetCodeTxConfig {
	c.gasLimit = &gasLimit
	return c
}

// WithGasLimitBuffer scales an estimated GasLimit by percent/100 (0 = no buffer).
func (c *SetCodeTxConfig) WithGasLimitBuffer(percent uint64) *SetCodeTxConfig {
	c.gasLimitBufferPct = percent
	return c
}

// WithStateOverride passes override to eth_estimateGas when GasLimit isn't set explicitly.
func (c *SetCodeTxConfig) WithStateOverride(override *rpc.StateOverride) *SetCodeTxConfig {
	c.stateOverride = override
	return c
}

// Pack resolves unset fields in one batch, estimating gas with the signed authorizationList, and returns the unsigned tx.
func (c *SetCodeTxConfig) Pack(ctx context.Context, client rpc.Client) (core.Transaction, error) {
	var authorizationList []map[string]any
	if c.gasLimit == nil {
		if err := c.stateOverride.Validate(); err != nil {
			return nil, fmt.Errorf("tx: gasLimit: %w", err)
		}
		authorizationList = make([]map[string]any, len(c.authorization))
		for i, auth := range c.authorization {
			if auth.Address == nil {
				return nil, fmt.Errorf("tx: gasLimit: authorization %d: address is required", i)
			}
			if auth.Signature.IsZero() {
				return nil, fmt.Errorf("tx: gasLimit: authorization %d: %w", i, core.ErrUnsigned)
			}
			yParity, err := rlp.RecoveryID(auth.Signature)
			if err != nil {
				return nil, fmt.Errorf("tx: gasLimit: authorization %d: %w", i, err)
			}
			authorizationList[i] = map[string]any{
				"chainId": quantityHex(auth.ChainID),
				"address": auth.Address.String(),
				"nonce":   quantityHex(new(big.Int).SetUint64(auth.Nonce)),
				"yParity": quantityHex(new(big.Int).SetUint64(uint64(yParity))),
				"r":       quantityHex(auth.Signature.R()),
				"s":       quantityHex(auth.Signature.S()),
			}
		}
	}

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
			"data":  encoding.Hex.EncodePrefixed(c.data),
		}
		if len(c.accessList) > 0 {
			callParams["accessList"] = accessListToRPC(c.accessList)
		}
		callParams["authorizationList"] = authorizationList
		elems.With(rpc.ETHEstimateGas("gasLimit", callParams, rpc.BlockTagLatest, c.stateOverride, &gasLimitHex))
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
		v := bufferUint64(n.Uint64(), c.gasLimitBufferPct)
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
