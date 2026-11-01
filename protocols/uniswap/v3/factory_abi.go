package v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Default fee tiers baked into UniswapV3Factory's constructor; the owner can enableFeeAmount
// additional tiers at runtime, so this isn't an exhaustive list.
const (
	FeeLow    = 500   // 0.05%, tickSpacing 10
	FeeMedium = 3000  // 0.3%,  tickSpacing 60
	FeeHigh   = 10000 // 1%,    tickSpacing 200
)

// Function signatures from IUniswapV3Factory, for encoding calls and decoding results.
var (
	ownerFn                = abi.NewFunction("owner", nil, abi.NewTypes(abi.Address))
	feeAmountTickSpacingFn = abi.NewFunction("feeAmountTickSpacing", abi.NewTypes(abi.Uint24), abi.NewTypes(abi.Int24))
	getPoolFn              = abi.NewFunction("getPool", abi.NewTypes(abi.Address, abi.Address, abi.Uint24), abi.NewTypes(abi.Address))
	createPoolFn           = abi.NewFunction("createPool", abi.NewTypes(abi.Address, abi.Address, abi.Uint24), abi.NewTypes(abi.Address))
	setOwnerFn             = abi.NewFunction("setOwner", abi.NewTypes(abi.Address), nil)
	enableFeeAmountFn      = abi.NewFunction("enableFeeAmount", abi.NewTypes(abi.Uint24, abi.Int24), nil)
)

// Events from IUniswapV3Factory, for decoding logs via Event.Decode.
var (
	ownerChangedEvent = abi.NewEvent("OwnerChanged", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
	}, false)

	poolCreatedEvent = abi.NewEvent("PoolCreated", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint24, Indexed: true},
		{Type: abi.Int24},
		{Type: abi.Address},
	}, false)

	feeAmountEnabledEvent = abi.NewEvent("FeeAmountEnabled", []abi.EventParam{
		{Type: abi.Uint24, Indexed: true},
		{Type: abi.Int24, Indexed: true},
	}, false)
)

// ErrReverted mirrors any UniswapV3Factory revert -- its requires carry no reason string, so
// there's nothing to distinguish one failed precondition from another.
var ErrReverted = errors.New("uniswapv3: reverted")

// EncodeOwner returns the calldata for owner().
func EncodeOwner() []byte {
	return ownerFn.SelectorBytes()
}

// DecodeOwner decodes the return data of owner.
func DecodeOwner(data []byte) (*types.Address, error) {
	addr, err := ownerFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeOwner: %w", err)
	}
	return addr, nil
}

// EncodeFeeAmountTickSpacing returns the calldata for feeAmountTickSpacing(fee).
func EncodeFeeAmountTickSpacing(fee *big.Int) ([]byte, error) {
	data, err := feeAmountTickSpacingFn.EncodeCall(fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeFeeAmountTickSpacing: %w", err)
	}
	return data, nil
}

// DecodeFeeAmountTickSpacing decodes the return data of feeAmountTickSpacing;
// zero means fee isn't an enabled fee tier.
func DecodeFeeAmountTickSpacing(data []byte) (*big.Int, error) {
	n, err := feeAmountTickSpacingFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeFeeAmountTickSpacing: %w", err)
	}
	return n, nil
}

// EncodeGetPool returns the calldata for getPool(tokenA, tokenB, fee).
func EncodeGetPool(tokenA, tokenB *types.Address, fee *big.Int) ([]byte, error) {
	data, err := getPoolFn.EncodeCall(tokenA, tokenB, fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeGetPool: %w", err)
	}
	return data, nil
}

// DecodeGetPool decodes the return data of getPool; a zero address means no pool exists yet.
func DecodeGetPool(data []byte) (*types.Address, error) {
	addr, err := getPoolFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeGetPool: %w", err)
	}
	return addr, nil
}

// EncodeCreatePool returns the calldata for createPool(tokenA, tokenB, fee).
func EncodeCreatePool(tokenA, tokenB *types.Address, fee *big.Int) ([]byte, error) {
	data, err := createPoolFn.EncodeCall(tokenA, tokenB, fee)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeCreatePool: %w", err)
	}
	return data, nil
}

// DecodeCreatePool decodes the return data of createPool -- the new pool's address. Only
// meaningful from an eth_call simulation; a real tx returns this via the PoolCreated event instead.
func DecodeCreatePool(data []byte) (*types.Address, error) {
	addr, err := createPoolFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decodeCreatePool: %w", err)
	}
	return addr, nil
}

// EncodeSetOwner returns the calldata for setOwner(_owner).
func EncodeSetOwner(owner *types.Address) ([]byte, error) {
	data, err := setOwnerFn.EncodeCall(owner)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeSetOwner: %w", err)
	}
	return data, nil
}

// EncodeEnableFeeAmount returns the calldata for enableFeeAmount(fee, tickSpacing).
func EncodeEnableFeeAmount(fee, tickSpacing *big.Int) ([]byte, error) {
	data, err := enableFeeAmountFn.EncodeCall(fee, tickSpacing)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: encodeEnableFeeAmount: %w", err)
	}
	return data, nil
}
