package v2

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IUniswapV2Pair's own surface -- name/symbol/decimals/totalSupply/
// balanceOf/allowance/approve/transfer/transferFrom/DOMAIN_SEPARATOR/nonces/permit come from
// erc20/erc2612 instead, since Pair's ABI for those is identical.
var (
	minimumLiquidityFn     = abi.NewFunction("MINIMUM_LIQUIDITY", nil, abi.NewTypes(abi.Uint256))
	factoryFn              = abi.NewFunction("factory", nil, abi.NewTypes(abi.Address))
	token0Fn               = abi.NewFunction("token0", nil, abi.NewTypes(abi.Address))
	token1Fn               = abi.NewFunction("token1", nil, abi.NewTypes(abi.Address))
	getReservesFn          = abi.NewFunction("getReserves", nil, abi.NewTypes(abi.Uint112, abi.Uint112, abi.Uint32))
	price0CumulativeLastFn = abi.NewFunction("price0CumulativeLast", nil, abi.NewTypes(abi.Uint256))
	price1CumulativeLastFn = abi.NewFunction("price1CumulativeLast", nil, abi.NewTypes(abi.Uint256))
	kLastFn                = abi.NewFunction("kLast", nil, abi.NewTypes(abi.Uint256))
	mintFn                 = abi.NewFunction("mint", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	burnFn                 = abi.NewFunction("burn", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256, abi.Uint256))
	swapFn                 = abi.NewFunction("swap", abi.NewTypes(abi.Uint256, abi.Uint256, abi.Address, abi.Bytes), nil)
	skimFn                 = abi.NewFunction("skim", abi.NewTypes(abi.Address), nil)
	syncFn                 = abi.NewFunction("sync", nil, nil)
	initializeFn           = abi.NewFunction("initialize", abi.NewTypes(abi.Address, abi.Address), nil)
	permitTypehashFn       = abi.NewFunction("PERMIT_TYPEHASH", nil, abi.NewTypes(abi.Bytes32))
)

// Events from IUniswapV2Pair's own surface -- Transfer/Approval come from erc20 instead.
var (
	mintEvent = abi.NewEvent("Mint", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)
	burnEvent = abi.NewEvent("Burn", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Address, Indexed: true},
	}, false)
	swapEvent = abi.NewEvent("Swap", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
		{Type: abi.Address, Indexed: true},
	}, false)
	syncEvent = abi.NewEvent("Sync", []abi.EventParam{
		{Type: abi.Uint112},
		{Type: abi.Uint112},
	}, false)
)

// UniswapV2Pair predates Solidity custom errors too; every revert is a plain require(cond,
// reason) string, decoded via abi.DecodeRevert. reasonForbidden/ErrForbidden are already
// declared in factory_abi.go and reused here for initialize's own forbidden check.
const (
	reasonLocked                      = "UniswapV2: LOCKED"
	reasonTransferFailed              = "UniswapV2: TRANSFER_FAILED"
	reasonOverflow                    = "UniswapV2: OVERFLOW"
	reasonInsufficientLiquidityMinted = "UniswapV2: INSUFFICIENT_LIQUIDITY_MINTED"
	reasonInsufficientLiquidityBurned = "UniswapV2: INSUFFICIENT_LIQUIDITY_BURNED"
	reasonInsufficientOutputAmount    = "UniswapV2: INSUFFICIENT_OUTPUT_AMOUNT"
	reasonInsufficientLiquidity       = "UniswapV2: INSUFFICIENT_LIQUIDITY"
	reasonInvalidTo                   = "UniswapV2: INVALID_TO"
	reasonInsufficientInputAmount     = "UniswapV2: INSUFFICIENT_INPUT_AMOUNT"
	reasonK                           = "UniswapV2: K"
	reasonAddOverflow                 = "ds-math-add-overflow"
	reasonSubUnderflow                = "ds-math-sub-underflow"
	reasonMulOverflow                 = "ds-math-mul-overflow"
)

// ErrLocked, etc. mirror UniswapV2Pair's revert reasons, for errors.Is.
var (
	ErrLocked                      = errors.New("uniswapv2: locked")
	ErrTransferFailed              = errors.New("uniswapv2: transfer failed")
	ErrOverflow                    = errors.New("uniswapv2: overflow")
	ErrInsufficientLiquidityMinted = errors.New("uniswapv2: insufficient liquidity minted")
	ErrInsufficientLiquidityBurned = errors.New("uniswapv2: insufficient liquidity burned")
	ErrInsufficientOutputAmount    = errors.New("uniswapv2: insufficient output amount")
	ErrInsufficientLiquidity       = errors.New("uniswapv2: insufficient liquidity")
	ErrInvalidTo                   = errors.New("uniswapv2: invalid to")
	ErrInsufficientInputAmount     = errors.New("uniswapv2: insufficient input amount")
	ErrK                           = errors.New("uniswapv2: k")
	ErrAddOverflow                 = errors.New("uniswapv2: add overflow")
	ErrSubUnderflow                = errors.New("uniswapv2: sub underflow")
	ErrMulOverflow                 = errors.New("uniswapv2: mul overflow")
)

// EncodeMinimumLiquidity returns the calldata for MINIMUM_LIQUIDITY().
func EncodeMinimumLiquidity() []byte {
	return minimumLiquidityFn.SelectorBytes()
}

// DecodeMinimumLiquidity decodes the return data of MINIMUM_LIQUIDITY.
func DecodeMinimumLiquidity(data []byte) (*big.Int, error) {
	n, err := minimumLiquidityFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeMinimumLiquidity: %w", err)
	}
	return n, nil
}

// EncodeFactory returns the calldata for factory().
func EncodeFactory() []byte {
	return factoryFn.SelectorBytes()
}

// DecodeFactory decodes the return data of factory.
func DecodeFactory(data []byte) (*types.Address, error) {
	addr, err := factoryFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeFactory: %w", err)
	}
	return addr, nil
}

// EncodeToken0 returns the calldata for token0().
func EncodeToken0() []byte {
	return token0Fn.SelectorBytes()
}

// DecodeToken0 decodes the return data of token0.
func DecodeToken0(data []byte) (*types.Address, error) {
	addr, err := token0Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeToken0: %w", err)
	}
	return addr, nil
}

// EncodeToken1 returns the calldata for token1().
func EncodeToken1() []byte {
	return token1Fn.SelectorBytes()
}

// DecodeToken1 decodes the return data of token1.
func DecodeToken1(data []byte) (*types.Address, error) {
	addr, err := token1Fn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeToken1: %w", err)
	}
	return addr, nil
}

// EncodeGetReserves returns the calldata for getReserves().
func EncodeGetReserves() []byte {
	return getReservesFn.SelectorBytes()
}

// DecodeGetReserves decodes the return data of getReserves.
func DecodeGetReserves(data []byte) (*Reserves, error) {
	vals, err := getReservesFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeGetReserves: %w", err)
	}
	reserve0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeGetReserves: unexpected decoded type for reserve0: %T", vals[0])
	}
	reserve1, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeGetReserves: unexpected decoded type for reserve1: %T", vals[1])
	}
	blockTimestampLast, ok := vals[2].(uint32)
	if !ok {
		return nil, fmt.Errorf("uniswapv2: decodeGetReserves: unexpected decoded type for blockTimestampLast: %T", vals[2])
	}
	return NewReserves(reserve0, reserve1, blockTimestampLast), nil
}

// EncodePrice0CumulativeLast returns the calldata for price0CumulativeLast().
func EncodePrice0CumulativeLast() []byte {
	return price0CumulativeLastFn.SelectorBytes()
}

// DecodePrice0CumulativeLast decodes the return data of price0CumulativeLast.
func DecodePrice0CumulativeLast(data []byte) (*big.Int, error) {
	n, err := price0CumulativeLastFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodePrice0CumulativeLast: %w", err)
	}
	return n, nil
}

// EncodePrice1CumulativeLast returns the calldata for price1CumulativeLast().
func EncodePrice1CumulativeLast() []byte {
	return price1CumulativeLastFn.SelectorBytes()
}

// DecodePrice1CumulativeLast decodes the return data of price1CumulativeLast.
func DecodePrice1CumulativeLast(data []byte) (*big.Int, error) {
	n, err := price1CumulativeLastFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodePrice1CumulativeLast: %w", err)
	}
	return n, nil
}

// EncodeKLast returns the calldata for kLast().
func EncodeKLast() []byte {
	return kLastFn.SelectorBytes()
}

// DecodeKLast decodes the return data of kLast.
func DecodeKLast(data []byte) (*big.Int, error) {
	n, err := kLastFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeKLast: %w", err)
	}
	return n, nil
}

// EncodeMint returns the calldata for mint(to).
func EncodeMint(to *types.Address) ([]byte, error) {
	data, err := mintFn.EncodeCall(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeMint: %w", err)
	}
	return data, nil
}

// DecodeMint decodes the return data of mint -- the liquidity minted. Only meaningful from an
// eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeMint(data []byte) (*big.Int, error) {
	n, err := mintFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodeMint: %w", err)
	}
	return n, nil
}

// EncodeBurn returns the calldata for burn(to).
func EncodeBurn(to *types.Address) ([]byte, error) {
	data, err := burnFn.EncodeCall(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeBurn: %w", err)
	}
	return data, nil
}

// DecodeBurn decodes the return data of burn -- (amount0, amount1) withdrawn. Only meaningful
// from an eth_call simulation; a mined transaction's return value isn't retrievable.
func DecodeBurn(data []byte) (amount0, amount1 *big.Int, err error) {
	vals, err := burnFn.DecodeReturn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: decodeBurn: %w", err)
	}
	amount0, ok := vals[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv2: decodeBurn: unexpected decoded type for amount0: %T", vals[0])
	}
	amount1, ok = vals[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("uniswapv2: decodeBurn: unexpected decoded type for amount1: %T", vals[1])
	}
	return amount0, amount1, nil
}

// EncodeSwap returns the calldata for swap(amount0Out, amount1Out, to, data). data is passed to
// IUniswapV2Callee(to).uniswapV2Call for a flash swap; pass nil/empty for an ordinary swap.
func EncodeSwap(amount0Out, amount1Out *big.Int, to *types.Address, data []byte) ([]byte, error) {
	calldata, err := swapFn.EncodeCall(amount0Out, amount1Out, to, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSwap: %w", err)
	}
	return calldata, nil
}

// EncodeSkim returns the calldata for skim(to) -- withdraws any excess balance over reserves.
func EncodeSkim(to *types.Address) ([]byte, error) {
	data, err := skimFn.EncodeCall(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeSkim: %w", err)
	}
	return data, nil
}

// EncodeSync returns the calldata for sync() -- forces reserves to match the pair's actual balances.
func EncodeSync() []byte {
	return syncFn.SelectorBytes()
}

// EncodeInitialize returns the calldata for initialize(token0, token1). Only the factory that
// deployed this pair can call it successfully (ErrForbidden otherwise), and only once.
func EncodeInitialize(token0, token1 *types.Address) ([]byte, error) {
	data, err := initializeFn.EncodeCall(token0, token1)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: encodeInitialize: %w", err)
	}
	return data, nil
}

// EncodePermitTypehash returns the calldata for PERMIT_TYPEHASH().
func EncodePermitTypehash() []byte {
	return permitTypehashFn.SelectorBytes()
}

// DecodePermitTypehash decodes the return data of PERMIT_TYPEHASH.
func DecodePermitTypehash(data []byte) ([]byte, error) {
	b, err := permitTypehashFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: decodePermitTypehash: %w", err)
	}
	return b, nil
}
