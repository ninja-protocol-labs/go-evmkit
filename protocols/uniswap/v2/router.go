package v2

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IRouter mirrors IUniswapV2Router02 (which embeds IUniswapV2Router01). No tx.Builder-taking
// Create* constructors or eth_call-simulating Call* methods here -- see the concrete Router's
// own CreateAddLiquidity/CallAddLiquidity etc.
type IRouter interface {
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	WETH(ctx context.Context, contract *types.Address, block string) (*types.Address, error)

	Quote(ctx context.Context, contract *types.Address, amountA, reserveA, reserveB *big.Int, block string) (*big.Int, error)
	GetAmountOut(ctx context.Context, contract *types.Address, amountIn, reserveIn, reserveOut *big.Int, block string) (*big.Int, error)
	GetAmountIn(ctx context.Context, contract *types.Address, amountOut, reserveIn, reserveOut *big.Int, block string) (*big.Int, error)
	GetAmountsOut(ctx context.Context, contract *types.Address, amountIn *big.Int, path []*types.Address, block string) ([]*big.Int, error)
	GetAmountsIn(ctx context.Context, contract *types.Address, amountOut *big.Int, path []*types.Address, block string) ([]*big.Int, error)

	AddLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	AddLiquidityETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidityETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidityWithPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidityETHWithPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactTokensForTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapTokensForExactTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactETHForTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapTokensForExactETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactTokensForETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapETHForExactTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SwapExactTokensForETHSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IRouter = (*Router)(nil)

// Router calls any UniswapV2Router02 contract over a Client; it holds no contract address, so
// one instance serves every deployment.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling routers via c.
func NewRouter(c rpc.Client) *Router {
	return &Router{
		c: c,
	}
}

// Factory returns contract's UniswapV2Factory.
func (r *Router) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := r.call(ctx, nil, contract, EncodeRouterFactory(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}
	addr, err := DecodeRouterFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}
	return addr, nil
}

// WETH returns contract's wrapped native token.
func (r *Router) WETH(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := r.call(ctx, nil, contract, EncodeWETH(), nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: weth: %w", err)
	}
	addr, err := DecodeWETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: weth: %w", err)
	}
	return addr, nil
}

// Quote returns the amount of the other asset equivalent to amountA, given reserveA/reserveB --
// a direct ratio with no 0.3% fee applied (unlike GetAmountOut).
func (r *Router) Quote(ctx context.Context, contract *types.Address, amountA, reserveA, reserveB *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeQuote(amountA, reserveA, reserveB)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}
	data, err := r.call(ctx, nil, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}
	n, err := DecodeQuote(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: quote: %w", err)
	}
	return n, nil
}

// GetAmountOut returns the output amount of a single swap of amountIn, given reserveIn/reserveOut.
func (r *Router) GetAmountOut(ctx context.Context, contract *types.Address, amountIn, reserveIn, reserveOut *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeGetAmountOut(amountIn, reserveIn, reserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}
	data, err := r.call(ctx, nil, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}
	n, err := DecodeGetAmountOut(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountOut: %w", err)
	}
	return n, nil
}

// GetAmountIn returns the input amount a single swap needs to produce amountOut, given
// reserveIn/reserveOut.
func (r *Router) GetAmountIn(ctx context.Context, contract *types.Address, amountOut, reserveIn, reserveOut *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeGetAmountIn(amountOut, reserveIn, reserveOut)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}
	data, err := r.call(ctx, nil, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}
	n, err := DecodeGetAmountIn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountIn: %w", err)
	}
	return n, nil
}

// GetAmountsOut returns the output amount of amountIn swapped along path, one hop at a time --
// amounts[0] is amountIn itself, amounts[len(amounts)-1] is the final output.
func (r *Router) GetAmountsOut(ctx context.Context, contract *types.Address, amountIn *big.Int, path []*types.Address, block string) ([]*big.Int, error) {
	calldata, err := EncodeGetAmountsOut(amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}
	data, err := r.call(ctx, nil, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}
	amounts, err := DecodeGetAmountsOut(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsOut: %w", err)
	}
	return amounts, nil
}

// GetAmountsIn returns the input amount path needs to produce amountOut at its far end, one hop
// at a time -- amounts[0] is the required input, amounts[len(amounts)-1] is amountOut itself.
func (r *Router) GetAmountsIn(ctx context.Context, contract *types.Address, amountOut *big.Int, path []*types.Address, block string) ([]*big.Int, error) {
	calldata, err := EncodeGetAmountsIn(amountOut, path)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}
	data, err := r.call(ctx, nil, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}
	amounts, err := DecodeGetAmountsIn(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getAmountsIn: %w", err)
	}
	return amounts, nil
}

// CallAddLiquidity simulates addLiquidity(...) via eth_call, returning the (amountA, amountB,
// liquidity) it would add -- the actual amounts can differ from amountADesired/amountBDesired
// depending on the pair's current reserve ratio.
func (r *Router) CallAddLiquidity(ctx context.Context, contract, sender, tokenA, tokenB *types.Address, amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, block string) (*AddLiquidityResult, error) {
	calldata, err := EncodeAddLiquidity(tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}
	result, err := DecodeAddLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidity: %w", err)
	}
	return result, nil
}

// CreateAddLiquidity builds an unsigned addLiquidity(...) transaction via b, sent by sender.
func (r *Router) CreateAddLiquidity(ctx context.Context, b tx.Builder, contract, sender, tokenA, tokenB *types.Address, amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeAddLiquidity(tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createAddLiquidity: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createAddLiquidity: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// AddLiquidity broadcasts an already-signed addLiquidity transaction.
func (r *Router) AddLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: addLiquidity: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: addLiquidity: %w", err)
	}
	return hash, nil
}

// CallAddLiquidityETH simulates addLiquidityETH(...) via eth_call, returning the (amountToken,
// amountETH, liquidity) it would add.
func (r *Router) CallAddLiquidityETH(ctx context.Context, contract, sender, token *types.Address, amountTokenDesired, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline, value *big.Int, block string) (*AddLiquidityETHResult, error) {
	calldata, err := EncodeAddLiquidityETH(token, amountTokenDesired, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, value, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}
	result, err := DecodeAddLiquidityETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callAddLiquidityETH: %w", err)
	}
	return result, nil
}

// CreateAddLiquidityETH builds an unsigned addLiquidityETH(...) transaction via b, sent by
// sender with value wei of native currency attached.
func (r *Router) CreateAddLiquidityETH(ctx context.Context, b tx.Builder, contract, sender, token *types.Address, amountTokenDesired, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeAddLiquidityETH(token, amountTokenDesired, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createAddLiquidityETH: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createAddLiquidityETH: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// AddLiquidityETH broadcasts an already-signed addLiquidityETH transaction.
func (r *Router) AddLiquidityETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: addLiquidityETH: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidity simulates removeLiquidity(...) via eth_call, returning the (amountA,
// amountB) it would withdraw.
func (r *Router) CallRemoveLiquidity(ctx context.Context, contract, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, block string) (*RemoveLiquidityResult, error) {
	calldata, err := EncodeRemoveLiquidity(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}
	result, err := DecodeRemoveLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidity: %w", err)
	}
	return result, nil
}

// CreateRemoveLiquidity builds an unsigned removeLiquidity(...) transaction via b, sent by
// sender.
func (r *Router) CreateRemoveLiquidity(ctx context.Context, b tx.Builder, contract, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidity(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidity: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidity: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidity broadcasts an already-signed removeLiquidity transaction.
func (r *Router) RemoveLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidity: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidity: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidityETH simulates removeLiquidityETH(...) via eth_call, returning the
// (amountToken, amountETH) it would withdraw.
func (r *Router) CallRemoveLiquidityETH(ctx context.Context, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, block string) (*RemoveLiquidityETHResult, error) {
	calldata, err := EncodeRemoveLiquidityETH(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}
	result, err := DecodeRemoveLiquidityETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETH: %w", err)
	}
	return result, nil
}

// CreateRemoveLiquidityETH builds an unsigned removeLiquidityETH(...) transaction via b, sent
// by sender.
func (r *Router) CreateRemoveLiquidityETH(ctx context.Context, b tx.Builder, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidityETH(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETH: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETH: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidityETH broadcasts an already-signed removeLiquidityETH transaction.
func (r *Router) RemoveLiquidityETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETH: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidityWithPermit simulates removeLiquidityWithPermit(...) via eth_call, returning
// the (amountA, amountB) it would withdraw.
func (r *Router) CallRemoveLiquidityWithPermit(ctx context.Context, contract, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature, block string) (*RemoveLiquidityResult, error) {
	calldata, err := EncodeRemoveLiquidityWithPermit(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}
	result, err := DecodeRemoveLiquidityWithPermit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityWithPermit: %w", err)
	}
	return result, nil
}

// CreateRemoveLiquidityWithPermit builds an unsigned removeLiquidityWithPermit(...) transaction
// via b, sent by sender.
func (r *Router) CreateRemoveLiquidityWithPermit(ctx context.Context, b tx.Builder, contract, sender, tokenA, tokenB *types.Address, liquidity, amountAMin, amountBMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidityWithPermit(tokenA, tokenB, liquidity, amountAMin, amountBMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityWithPermit: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityWithPermit: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidityWithPermit broadcasts an already-signed removeLiquidityWithPermit transaction.
func (r *Router) RemoveLiquidityWithPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityWithPermit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityWithPermit: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidityETHWithPermit simulates removeLiquidityETHWithPermit(...) via eth_call,
// returning the (amountToken, amountETH) it would withdraw.
func (r *Router) CallRemoveLiquidityETHWithPermit(ctx context.Context, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature, block string) (*RemoveLiquidityETHResult, error) {
	calldata, err := EncodeRemoveLiquidityETHWithPermit(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}
	result, err := DecodeRemoveLiquidityETHWithPermit(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermit: %w", err)
	}
	return result, nil
}

// CreateRemoveLiquidityETHWithPermit builds an unsigned removeLiquidityETHWithPermit(...)
// transaction via b, sent by sender.
func (r *Router) CreateRemoveLiquidityETHWithPermit(ctx context.Context, b tx.Builder, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidityETHWithPermit(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHWithPermit: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHWithPermit: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidityETHWithPermit broadcasts an already-signed removeLiquidityETHWithPermit
// transaction.
func (r *Router) RemoveLiquidityETHWithPermit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermit: %w", err)
	}
	return hash, nil
}

// CallSwapExactTokensForTokens simulates swapExactTokensForTokens(...) via eth_call, returning
// the amounts path-aligned with path -- equivalent to GetAmountsOut, but as an actual simulated
// call instead of a pure computation.
func (r *Router) CallSwapExactTokensForTokens(ctx context.Context, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapExactTokensForTokens(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}
	amounts, err := DecodeSwapExactTokensForTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForTokens: %w", err)
	}
	return amounts, nil
}

// CreateSwapExactTokensForTokens builds an unsigned swapExactTokensForTokens(...) transaction
// via b, sent by sender.
func (r *Router) CreateSwapExactTokensForTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactTokensForTokens(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactTokensForTokens broadcasts an already-signed swapExactTokensForTokens transaction.
func (r *Router) SwapExactTokensForTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForTokens: %w", err)
	}
	return hash, nil
}

// CallSwapTokensForExactTokens simulates swapTokensForExactTokens(...) via eth_call, returning
// the amounts path-aligned with path.
func (r *Router) CallSwapTokensForExactTokens(ctx context.Context, contract, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapTokensForExactTokens(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}
	amounts, err := DecodeSwapTokensForExactTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactTokens: %w", err)
	}
	return amounts, nil
}

// CreateSwapTokensForExactTokens builds an unsigned swapTokensForExactTokens(...) transaction
// via b, sent by sender.
func (r *Router) CreateSwapTokensForExactTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapTokensForExactTokens(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapTokensForExactTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapTokensForExactTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapTokensForExactTokens broadcasts an already-signed swapTokensForExactTokens transaction.
func (r *Router) SwapTokensForExactTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapTokensForExactTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapTokensForExactTokens: %w", err)
	}
	return hash, nil
}

// CallSwapExactETHForTokens simulates swapExactETHForTokens(...) via eth_call with value wei of
// native currency, returning the amounts path-aligned with path.
func (r *Router) CallSwapExactETHForTokens(ctx context.Context, contract, sender *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapExactETHForTokens(amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, value, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}
	amounts, err := DecodeSwapExactETHForTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactETHForTokens: %w", err)
	}
	return amounts, nil
}

// CreateSwapExactETHForTokens builds an unsigned swapExactETHForTokens(...) transaction via b,
// sent by sender with value wei of native currency attached.
func (r *Router) CreateSwapExactETHForTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactETHForTokens(amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactETHForTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactETHForTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactETHForTokens broadcasts an already-signed swapExactETHForTokens transaction.
func (r *Router) SwapExactETHForTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactETHForTokens: %w", err)
	}
	return hash, nil
}

// CallSwapTokensForExactETH simulates swapTokensForExactETH(...) via eth_call, returning the
// amounts path-aligned with path.
func (r *Router) CallSwapTokensForExactETH(ctx context.Context, contract, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapTokensForExactETH(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}
	amounts, err := DecodeSwapTokensForExactETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapTokensForExactETH: %w", err)
	}
	return amounts, nil
}

// CreateSwapTokensForExactETH builds an unsigned swapTokensForExactETH(...) transaction via b,
// sent by sender.
func (r *Router) CreateSwapTokensForExactETH(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountOut, amountInMax *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapTokensForExactETH(amountOut, amountInMax, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapTokensForExactETH: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapTokensForExactETH: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapTokensForExactETH broadcasts an already-signed swapTokensForExactETH transaction.
func (r *Router) SwapTokensForExactETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapTokensForExactETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapTokensForExactETH: %w", err)
	}
	return hash, nil
}

// CallSwapExactTokensForETH simulates swapExactTokensForETH(...) via eth_call, returning the
// amounts path-aligned with path.
func (r *Router) CallSwapExactTokensForETH(ctx context.Context, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapExactTokensForETH(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}
	amounts, err := DecodeSwapExactTokensForETH(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapExactTokensForETH: %w", err)
	}
	return amounts, nil
}

// CreateSwapExactTokensForETH builds an unsigned swapExactTokensForETH(...) transaction via b,
// sent by sender.
func (r *Router) CreateSwapExactTokensForETH(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactTokensForETH(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForETH: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForETH: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactTokensForETH broadcasts an already-signed swapExactTokensForETH transaction.
func (r *Router) SwapExactTokensForETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForETH: %w", err)
	}
	return hash, nil
}

// CallSwapETHForExactTokens simulates swapETHForExactTokens(...) via eth_call with value wei of
// native currency (the max input), returning the amounts path-aligned with path.
func (r *Router) CallSwapETHForExactTokens(ctx context.Context, contract, sender *types.Address, amountOut *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int, block string) ([]*big.Int, error) {
	calldata, err := EncodeSwapETHForExactTokens(amountOut, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, value, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}
	amounts, err := DecodeSwapETHForExactTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callSwapETHForExactTokens: %w", err)
	}
	return amounts, nil
}

// CreateSwapETHForExactTokens builds an unsigned swapETHForExactTokens(...) transaction via b,
// sent by sender with value wei of native currency attached.
func (r *Router) CreateSwapETHForExactTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountOut *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapETHForExactTokens(amountOut, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapETHForExactTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapETHForExactTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapETHForExactTokens broadcasts an already-signed swapETHForExactTokens transaction.
func (r *Router) SwapETHForExactTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapETHForExactTokens: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidityETHSupportingFeeOnTransferTokens simulates
// removeLiquidityETHSupportingFeeOnTransferTokens(...) via eth_call, returning the amountETH it
// would withdraw.
func (r *Router) CallRemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	amountETH, err := DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	return amountETH, nil
}

// CreateRemoveLiquidityETHSupportingFeeOnTransferTokens builds an unsigned
// removeLiquidityETHSupportingFeeOnTransferTokens(...) transaction via b, sent by sender.
func (r *Router) CreateRemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, b tx.Builder, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidityETHSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidityETHSupportingFeeOnTransferTokens broadcasts an already-signed
// removeLiquidityETHSupportingFeeOnTransferTokens transaction.
func (r *Router) RemoveLiquidityETHSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHSupportingFeeOnTransferTokens: %w", err)
	}
	return hash, nil
}

// CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens simulates
// removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(...) via eth_call, returning the
// amountETH it would withdraw.
func (r *Router) CallRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature, block string) (*big.Int, error) {
	calldata, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	data, err := r.call(ctx, sender, contract, calldata, nil, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	amountETH, err := DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	return amountETH, nil
}

// CreateRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens builds an unsigned
// removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(...) transaction via b, sent by
// sender.
func (r *Router) CreateRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, b tx.Builder, contract, sender, token *types.Address, liquidity, amountTokenMin, amountETHMin *big.Int, to *types.Address, deadline *big.Int, approveMax bool, sig *types.Signature) (core.Transaction, error) {
	calldata, err := EncodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(token, liquidity, amountTokenMin, amountETHMin, to, deadline, approveMax, sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens broadcasts an already-signed
// removeLiquidityETHWithPermitSupportingFeeOnTransferTokens transaction.
func (r *Router) RemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: removeLiquidityETHWithPermitSupportingFeeOnTransferTokens: %w", err)
	}
	return hash, nil
}

// CreateSwapExactTokensForTokensSupportingFeeOnTransferTokens builds an unsigned
// swapExactTokensForTokensSupportingFeeOnTransferTokens(...) transaction via b, sent by sender.
func (r *Router) CreateSwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactTokensForTokensSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForTokensSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactTokensForTokensSupportingFeeOnTransferTokens broadcasts an already-signed
// swapExactTokensForTokensSupportingFeeOnTransferTokens transaction.
func (r *Router) SwapExactTokensForTokensSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	return hash, nil
}

// CreateSwapExactETHForTokensSupportingFeeOnTransferTokens builds an unsigned
// swapExactETHForTokensSupportingFeeOnTransferTokens(...) transaction via b, sent by sender
// with value wei of native currency attached.
func (r *Router) CreateSwapExactETHForTokensSupportingFeeOnTransferTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactETHForTokensSupportingFeeOnTransferTokens(amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactETHForTokensSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactETHForTokensSupportingFeeOnTransferTokens broadcasts an already-signed
// swapExactETHForTokensSupportingFeeOnTransferTokens transaction.
func (r *Router) SwapExactETHForTokensSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}
	return hash, nil
}

// CreateSwapExactTokensForETHSupportingFeeOnTransferTokens builds an unsigned
// swapExactTokensForETHSupportingFeeOnTransferTokens(...) transaction via b, sent by sender.
func (r *Router) CreateSwapExactTokensForETHSupportingFeeOnTransferTokens(ctx context.Context, b tx.Builder, contract, sender *types.Address, amountIn, amountOutMin *big.Int, path []*types.Address, to *types.Address, deadline *big.Int) (core.Transaction, error) {
	calldata, err := EncodeSwapExactTokensForETHSupportingFeeOnTransferTokens(amountIn, amountOutMin, path, to, deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwapExactTokensForETHSupportingFeeOnTransferTokens: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// SwapExactTokensForETHSupportingFeeOnTransferTokens broadcasts an already-signed
// swapExactTokensForETHSupportingFeeOnTransferTokens transaction.
func (r *Router) SwapExactTokensForETHSupportingFeeOnTransferTokens(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swapExactTokensForETHSupportingFeeOnTransferTokens: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, and msg.value value, if non-nil),
// decoding a reverted require(cond, reason) to its sentinel.
func (r *Router) call(ctx context.Context, from, contract *types.Address, calldata []byte, value *big.Int, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}
	if value != nil {
		params["value"] = "0x" + value.Text(16)
	}

	var raw string
	if err := r.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, r.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known UniswapV2Router/UniswapV2Library revert reason in err to its
// sentinel. Some reasons are shared in meaning with an existing Factory/Pair sentinel -- see the
// comment above ErrExpired in router_abi.go.
func (_ *Router) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return err
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return err
	}

	switch reason {
	case reasonExpired:
		return fmt.Errorf("%w: %w", ErrExpired, err)
	case reasonInsufficientAAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientAAmount, err)
	case reasonInsufficientBAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientBAmount, err)
	case reasonRouterInsufficientOutput, reasonLibraryInsufficientOutputOld, reasonLibraryInsufficientOutputNew:
		return fmt.Errorf("%w: %w", ErrInsufficientOutputAmount, err)
	case reasonExcessiveInputAmount:
		return fmt.Errorf("%w: %w", ErrExcessiveInputAmount, err)
	case reasonRouterInvalidPath, reasonLibraryInvalidPathOld, reasonLibraryInvalidPathNew:
		return fmt.Errorf("%w: %w", ErrInvalidPath, err)
	case reasonLibraryIdenticalAddressesOld, reasonLibraryIdenticalAddressesNew:
		return fmt.Errorf("%w: %w", ErrIdenticalAddresses, err)
	case reasonLibraryZeroAddressOld, reasonLibraryZeroAddressNew:
		return fmt.Errorf("%w: %w", ErrZeroAddress, err)
	case reasonLibraryInsufficientAmountOld, reasonLibraryInsufficientAmountNew:
		return fmt.Errorf("%w: %w", ErrInsufficientAmount, err)
	case reasonLibraryInsufficientLiquidityOld, reasonLibraryInsufficientLiquidityNew:
		return fmt.Errorf("%w: %w", ErrInsufficientLiquidity, err)
	case reasonLibraryInsufficientInputOld, reasonLibraryInsufficientInputNew:
		return fmt.Errorf("%w: %w", ErrInsufficientInputAmount, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
