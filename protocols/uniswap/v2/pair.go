package v2

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc2612"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IPair mirrors IUniswapV2Pair, embedding erc20/erc2612 for the ERC20+Permit surface they share byte-for-byte.
type IPair interface {
	erc20.IERC20Metadata
	erc2612.IERC2612

	MinimumLiquidity(ctx context.Context, opts *bind.CallOptions, req *MinimumLiquidityRequest) (*big.Int, error)
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	Token0(ctx context.Context, opts *bind.CallOptions, req *Token0Request) (*types.Address, error)
	Token1(ctx context.Context, opts *bind.CallOptions, req *Token1Request) (*types.Address, error)
	GetReserves(ctx context.Context, opts *bind.CallOptions, req *GetReservesRequest) (*Reserves, error)
	Price0CumulativeLast(ctx context.Context, opts *bind.CallOptions, req *Price0CumulativeLastRequest) (*big.Int, error)
	Price1CumulativeLast(ctx context.Context, opts *bind.CallOptions, req *Price1CumulativeLastRequest) (*big.Int, error)
	KLast(ctx context.Context, opts *bind.CallOptions, req *KLastRequest) (*big.Int, error)
	PermitTypehash(ctx context.Context, opts *bind.CallOptions, req *PermitTypehashRequest) ([]byte, error)
	CallMint(ctx context.Context, opts *bind.CallOptions, req *MintRequest) (*big.Int, error)
	CallBurn(ctx context.Context, opts *bind.CallOptions, req *BurnRequest) (amount0, amount1 *big.Int, err error)

	Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error)
	Mint(ctx context.Context, opts *bind.TransactOptions, req *MintRequest) (core.Transaction, *types.Hash, error)
	Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error)
	Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error)
	Skim(ctx context.Context, opts *bind.TransactOptions, req *SkimRequest) (core.Transaction, *types.Hash, error)
	Sync(ctx context.Context, opts *bind.TransactOptions, req *SyncRequest) (core.Transaction, *types.Hash, error)
}

var _ IPair = (*Pair)(nil)

// Pair calls any UniswapV2Pair contract over a Client; it holds no contract address, so one instance serves every pair.
type Pair struct {
	*erc20.ERC20
	*erc2612.ERC2612

	c rpc.Client
}

// NewPair returns a Pair calling pairs via cli.
func NewPair(cli rpc.Client) *Pair {
	return &Pair{
		ERC20:   erc20.New(cli),
		ERC2612: erc2612.New(cli),
		c:       cli,
	}
}

// MinimumLiquidity returns req.Contract's minimum liquidity burned to the zero address on first mint.
func (pr *Pair) MinimumLiquidity(ctx context.Context, opts *bind.CallOptions, req *MinimumLiquidityRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: minimumLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeMinimumLiquidity())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: minimumLiquidity: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeMinimumLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: minimumLiquidity: %w", err)
	}

	return out, nil
}

// Factory returns the UniswapV2Factory that deployed req.Contract.
func (pr *Pair) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}

	return out, nil
}

// Token0 returns req.Contract's token0 (the lower-valued of its two tokens).
func (pr *Pair) Token0(ctx context.Context, opts *bind.CallOptions, req *Token0Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: token0: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeToken0())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token0: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeToken0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token0: %w", err)
	}

	return out, nil
}

// Token1 returns req.Contract's token1 (the higher-valued of its two tokens).
func (pr *Pair) Token1(ctx context.Context, opts *bind.CallOptions, req *Token1Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: token1: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeToken1())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token1: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeToken1(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token1: %w", err)
	}

	return out, nil
}

// GetReserves returns req.Contract's current (reserve0, reserve1, blockTimestampLast).
func (pr *Pair) GetReserves(ctx context.Context, opts *bind.CallOptions, req *GetReservesRequest) (*Reserves, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: getReserves: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeGetReserves())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getReserves: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeGetReserves(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getReserves: %w", err)
	}

	return out, nil
}

// Price0CumulativeLast returns req.Contract's cumulative price of token0, as of the last update.
func (pr *Pair) Price0CumulativeLast(ctx context.Context, opts *bind.CallOptions, req *Price0CumulativeLastRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: price0CumulativeLast: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodePrice0CumulativeLast())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price0CumulativeLast: %w", pr.wrapRevertError(err))
	}

	out, err := DecodePrice0CumulativeLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price0CumulativeLast: %w", err)
	}

	return out, nil
}

// Price1CumulativeLast returns req.Contract's cumulative price of token1, as of the last update.
func (pr *Pair) Price1CumulativeLast(ctx context.Context, opts *bind.CallOptions, req *Price1CumulativeLastRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: price1CumulativeLast: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodePrice1CumulativeLast())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price1CumulativeLast: %w", pr.wrapRevertError(err))
	}

	out, err := DecodePrice1CumulativeLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price1CumulativeLast: %w", err)
	}

	return out, nil
}

// KLast returns req.Contract's reserve0*reserve1 as of the most recent liquidity event.
func (pr *Pair) KLast(ctx context.Context, opts *bind.CallOptions, req *KLastRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: kLast: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodeKLast())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: kLast: %w", pr.wrapRevertError(err))
	}

	out, err := DecodeKLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: kLast: %w", err)
	}

	return out, nil
}

// PermitTypehash returns req.Contract's Permit EIP-712 typehash.
func (pr *Pair) PermitTypehash(ctx context.Context, opts *bind.CallOptions, req *PermitTypehashRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: permitTypehash: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, EncodePermitTypehash())
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: permitTypehash: %w", pr.wrapRevertError(err))
	}

	out, err := DecodePermitTypehash(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: permitTypehash: %w", err)
	}

	return out, nil
}

// CallMint runs mint(to) via eth_call as opts.From, returning the liquidity it would mint or the revert.
func (pr *Pair) CallMint(ctx context.Context, opts *bind.CallOptions, req *MintRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}

	calldata, err := EncodeMint(req.To)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", pr.wrapRevertError(err))
	}

	liquidity, err := DecodeMint(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}

	return liquidity, nil
}

// CallBurn runs burn(to) via eth_call as opts.From, returning the (amount0, amount1) it would withdraw or the revert.
func (pr *Pair) CallBurn(ctx context.Context, opts *bind.CallOptions, req *BurnRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}

	calldata, err := EncodeBurn(req.To)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}

	data, err := bind.Call(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", pr.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeBurn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}

	return amount0, amount1, nil
}

// Initialize sends initialize(token0, token1) on req.Contract, signed by opts.Signer, which must be the pair's factory or it reverts with ErrForbidden.
func (pr *Pair) Initialize(ctx context.Context, opts *bind.TransactOptions, req *InitializeRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: initialize: %w", err)
	}

	calldata, err := EncodeInitialize(req.Token0, req.Token1)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: initialize: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: initialize: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Mint sends mint(to) on req.Contract, signed by opts.Signer; the tokens must already be transferred to the pair.
func (pr *Pair) Mint(ctx context.Context, opts *bind.TransactOptions, req *MintRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: mint: %w", err)
	}

	calldata, err := EncodeMint(req.To)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: mint: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: mint: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Burn sends burn(to) on req.Contract, signed by opts.Signer; the LP tokens must already be transferred to the pair.
func (pr *Pair) Burn(ctx context.Context, opts *bind.TransactOptions, req *BurnRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: burn: %w", err)
	}

	calldata, err := EncodeBurn(req.To)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: burn: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: burn: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Swap sends swap(amount0Out, amount1Out, to, data) on req.Contract, signed by opts.Signer; a non-empty Data makes it a flash swap.
func (pr *Pair) Swap(ctx context.Context, opts *bind.TransactOptions, req *SwapRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swap: %w", err)
	}

	calldata, err := EncodeSwap(req.Amount0Out, req.Amount1Out, req.To, req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: swap: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: swap: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Skim sends skim(to) on req.Contract, signed by opts.Signer, sending any balance above the reserves to req.To.
func (pr *Pair) Skim(ctx context.Context, opts *bind.TransactOptions, req *SkimRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: skim: %w", err)
	}

	calldata, err := EncodeSkim(req.To)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: skim: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: skim: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Sync sends sync() on req.Contract, signed by opts.Signer, forcing the reserves to match the pair's balances.
func (pr *Pair) Sync(ctx context.Context, opts *bind.TransactOptions, req *SyncRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: sync: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, pr.c, opts, req.Contract, EncodeSync())
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv2: sync: %w", pr.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a known revert reason in err to its sentinel.
func (_ *Pair) wrapRevertError(err error) error {
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
	case reasonLocked:
		return fmt.Errorf("%w: %w", ErrLocked, err)
	case reasonTransferFailed:
		return fmt.Errorf("%w: %w", ErrTransferFailed, err)
	case reasonForbidden:
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	case reasonOverflow:
		return fmt.Errorf("%w: %w", ErrOverflow, err)
	case reasonInsufficientLiquidityMinted:
		return fmt.Errorf("%w: %w", ErrInsufficientLiquidityMinted, err)
	case reasonInsufficientLiquidityBurned:
		return fmt.Errorf("%w: %w", ErrInsufficientLiquidityBurned, err)
	case reasonInsufficientOutputAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientOutputAmount, err)
	case reasonInsufficientLiquidity:
		return fmt.Errorf("%w: %w", ErrInsufficientLiquidity, err)
	case reasonInvalidTo:
		return fmt.Errorf("%w: %w", ErrInvalidTo, err)
	case reasonInsufficientInputAmount:
		return fmt.Errorf("%w: %w", ErrInsufficientInputAmount, err)
	case reasonK:
		return fmt.Errorf("%w: %w", ErrK, err)
	case reasonAddOverflow:
		return fmt.Errorf("%w: %w", ErrAddOverflow, err)
	case reasonSubUnderflow:
		return fmt.Errorf("%w: %w", ErrSubUnderflow, err)
	case reasonMulOverflow:
		return fmt.Errorf("%w: %w", ErrMulOverflow, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
