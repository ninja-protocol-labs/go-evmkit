package v2

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IUniswapV2Factory, for encoding calls and decoding results.
var (
	feeToFn          = abi.NewFunction("feeTo", nil, abi.NewTypes(abi.Address))
	feeToSetterFn    = abi.NewFunction("feeToSetter", nil, abi.NewTypes(abi.Address))
	getPairFn        = abi.NewFunction("getPair", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Address))
	allPairsFn       = abi.NewFunction("allPairs", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Address))
	allPairsLengthFn = abi.NewFunction("allPairsLength", nil, abi.NewTypes(abi.Uint256))
	createPairFn     = abi.NewFunction("createPair", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Address))
	setFeeToFn       = abi.NewFunction("setFeeTo", abi.NewTypes(abi.Address), nil)
	setFeeToSetterFn = abi.NewFunction("setFeeToSetter", abi.NewTypes(abi.Address), nil)
)

// Events from IUniswapV2Factory, for decoding logs via Event.Decode.
var pairCreatedEvent = abi.NewEvent("PairCreated", []abi.EventParam{
	{Type: abi.Address, Indexed: true},
	{Type: abi.Address, Indexed: true},
	{Type: abi.Address},
	{Type: abi.Uint256},
}, false)

// UniswapV2Factory predates Solidity custom errors; every revert is a plain require(cond, reason)
// string, decoded via abi.DecodeRevert rather than a 4-byte selector match.
const (
	reasonIdenticalAddresses = "UniswapV2: IDENTICAL_ADDRESSES"
	reasonZeroAddress        = "UniswapV2: ZERO_ADDRESS"
	reasonPairExists         = "UniswapV2: PAIR_EXISTS"
	reasonForbidden          = "UniswapV2: FORBIDDEN"
)

// ErrIdenticalAddresses, etc. mirror UniswapV2Factory's revert reasons, for errors.Is.
var (
	ErrIdenticalAddresses = errors.New("uniswapv2: identical addresses")
	ErrZeroAddress        = errors.New("uniswapv2: zero address")
	ErrPairExists         = errors.New("uniswapv2: pair exists")
	ErrForbidden          = errors.New("uniswapv2: forbidden")
	ErrUnknown            = errors.New("uniswapv2: unknown revert")
)

// EncodeFeeTo returns the calldata for feeTo().
func EncodeFeeTo() []byte {
	return feeToFn.SelectorBytes()
}

// DecodeFeeTo decodes the return data of feeTo.
func DecodeFeeTo(data []byte) (*types.Address, error) {
	addr, err := feeToFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeFeeTo: %w", err)
	}
	return addr, nil
}

// EncodeFeeToSetter returns the calldata for feeToSetter().
func EncodeFeeToSetter() []byte {
	return feeToSetterFn.SelectorBytes()
}

// DecodeFeeToSetter decodes the return data of feeToSetter.
func DecodeFeeToSetter(data []byte) (*types.Address, error) {
	addr, err := feeToSetterFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeFeeToSetter: %w", err)
	}
	return addr, nil
}

// EncodeGetPair returns the calldata for getPair(tokenA, tokenB).
func EncodeGetPair(tokenA, tokenB *types.Address) ([]byte, error) {
	data, err := getPairFn.EncodeCall(tokenA, tokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeGetPair: %w", err)
	}
	return data, nil
}

// DecodeGetPair decodes the return data of getPair; a zero address means no pair exists yet.
func DecodeGetPair(data []byte) (*types.Address, error) {
	addr, err := getPairFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetPair: %w", err)
	}
	return addr, nil
}

// EncodeAllPairs returns the calldata for allPairs(index).
func EncodeAllPairs(index *big.Int) ([]byte, error) {
	data, err := allPairsFn.EncodeCall(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeAllPairs: %w", err)
	}
	return data, nil
}

// DecodeAllPairs decodes the return data of allPairs.
func DecodeAllPairs(data []byte) (*types.Address, error) {
	addr, err := allPairsFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeAllPairs: %w", err)
	}
	return addr, nil
}

// EncodeAllPairsLength returns the calldata for allPairsLength().
func EncodeAllPairsLength() []byte {
	return allPairsLengthFn.SelectorBytes()
}

// DecodeAllPairsLength decodes the return data of allPairsLength.
func DecodeAllPairsLength(data []byte) (*big.Int, error) {
	n, err := allPairsLengthFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeAllPairsLength: %w", err)
	}
	return n, nil
}

// EncodeCreatePair returns the calldata for createPair(tokenA, tokenB).
func EncodeCreatePair(tokenA, tokenB *types.Address) ([]byte, error) {
	data, err := createPairFn.EncodeCall(tokenA, tokenB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeCreatePair: %w", err)
	}
	return data, nil
}

// DecodeCreatePair decodes the return data of createPair -- the new pair's address. Only
// meaningful from an eth_call simulation (e.g. CallCreatePair); a real tx returns this via
// the PairCreated event instead, since a mined transaction's return value isn't retrievable.
func DecodeCreatePair(data []byte) (*types.Address, error) {
	addr, err := createPairFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeCreatePair: %w", err)
	}
	return addr, nil
}

// EncodeSetFeeTo returns the calldata for setFeeTo(feeTo).
func EncodeSetFeeTo(feeTo *types.Address) ([]byte, error) {
	data, err := setFeeToFn.EncodeCall(feeTo)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSetFeeTo: %w", err)
	}
	return data, nil
}

// EncodeSetFeeToSetter returns the calldata for setFeeToSetter(feeToSetter).
func EncodeSetFeeToSetter(feeToSetter *types.Address) ([]byte, error) {
	data, err := setFeeToSetterFn.EncodeCall(feeToSetter)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSetFeeToSetter: %w", err)
	}
	return data, nil
}
