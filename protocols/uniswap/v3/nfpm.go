package v3

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// INFPMImmutables mirrors PeripheryImmutableState.
type INFPMImmutables interface {
	Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error)
	WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error)
}

// INFPMEnumerable mirrors IERC721Enumerable.
type INFPMEnumerable interface {
	TotalSupply(ctx context.Context, opts *bind.CallOptions, req *TotalSupplyRequest) (*big.Int, error)
	TokenOfOwnerByIndex(ctx context.Context, opts *bind.CallOptions, req *TokenOfOwnerByIndexRequest) (*big.Int, error)
	TokenByIndex(ctx context.Context, opts *bind.CallOptions, req *TokenByIndexRequest) (*big.Int, error)
}

// INFPMPermit mirrors IERC721Permit.
type INFPMPermit interface {
	PermitTypehash(ctx context.Context, opts *bind.CallOptions, req *PermitTypehashRequest) ([32]byte, error)
	DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) ([32]byte, error)

	Permit(ctx context.Context, opts *bind.TransactOptions, req *PermitRequest) (core.Transaction, *types.Hash, error)
}

// INFPMPoolInitializer mirrors IPoolInitializer.
type INFPMPoolInitializer interface {
	CallCreateAndInitializePoolIfNecessary(ctx context.Context, opts *bind.CallOptions, req *CreateAndInitializePoolIfNecessaryRequest) (pool *types.Address, err error)

	CreateAndInitializePoolIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *CreateAndInitializePoolIfNecessaryRequest) (core.Transaction, *types.Hash, error)
}

// INFPMPayments mirrors IPeripheryPayments.
type INFPMPayments interface {
	UnwrapWETH9(ctx context.Context, opts *bind.TransactOptions, req *NFPMUnwrapWETH9Request) (core.Transaction, *types.Hash, error)
	SweepToken(ctx context.Context, opts *bind.TransactOptions, req *NFPMSweepTokenRequest) (core.Transaction, *types.Hash, error)
	RefundETH(ctx context.Context, opts *bind.TransactOptions, req *NFPMRefundETHRequest) (core.Transaction, *types.Hash, error)
}

// INFPMMulticall mirrors Multicall.
type INFPMMulticall interface {
	CallMulticall(ctx context.Context, opts *bind.CallOptions, req *NFPMMulticallRequest) (results [][]byte, err error)

	Multicall(ctx context.Context, opts *bind.TransactOptions, req *NFPMMulticallRequest) (core.Transaction, *types.Hash, error)
}

// INFPMActions mirrors INonfungiblePositionManager's own position-management surface.
type INFPMActions interface {
	Positions(ctx context.Context, opts *bind.CallOptions, req *NFPMPositionsRequest) (*Position, error)
	CallMint(ctx context.Context, opts *bind.CallOptions, req *NFPMMintRequest) (tokenID, liquidity, amount0, amount1 *big.Int, err error)
	CallIncreaseLiquidity(ctx context.Context, opts *bind.CallOptions, req *IncreaseLiquidityRequest) (liquidity, amount0, amount1 *big.Int, err error)
	CallDecreaseLiquidity(ctx context.Context, opts *bind.CallOptions, req *DecreaseLiquidityRequest) (amount0, amount1 *big.Int, err error)
	CallCollect(ctx context.Context, opts *bind.CallOptions, req *NFPMCollectRequest) (amount0, amount1 *big.Int, err error)

	Mint(ctx context.Context, opts *bind.TransactOptions, req *NFPMMintRequest) (core.Transaction, *types.Hash, error)
	IncreaseLiquidity(ctx context.Context, opts *bind.TransactOptions, req *IncreaseLiquidityRequest) (core.Transaction, *types.Hash, error)
	DecreaseLiquidity(ctx context.Context, opts *bind.TransactOptions, req *DecreaseLiquidityRequest) (core.Transaction, *types.Hash, error)
	Collect(ctx context.Context, opts *bind.TransactOptions, req *NFPMCollectRequest) (core.Transaction, *types.Hash, error)
	Burn(ctx context.Context, opts *bind.TransactOptions, req *NFPMBurnRequest) (core.Transaction, *types.Hash, error)
}

// INFPM mirrors NonfungiblePositionManager: one small interface per Solidity base, plus the ERC721 (+ Metadata) surface positions live in.
type INFPM interface {
	erc721.IERC721Metadata
	INFPMImmutables
	INFPMEnumerable
	INFPMPermit
	INFPMPoolInitializer
	INFPMPayments
	INFPMMulticall
	INFPMActions
}

var _ INFPM = (*NFPM)(nil)

// NFPM calls any NonfungiblePositionManager over a Client; the embedded *erc721.ERC721 serves the ERC721 (+ Metadata) surface.
type NFPM struct {
	*erc721.ERC721

	c rpc.Client
}

// NewNFPM returns an NFPM calling position managers via cli.
func NewNFPM(cli rpc.Client) *NFPM {
	return &NFPM{
		ERC721: erc721.New(cli),
		c:      cli,
	}
}

// Factory returns the UniswapV3Factory req.Contract mints positions against.
func (n *NFPM) Factory(ctx context.Context, opts *bind.CallOptions, req *FactoryRequest) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, EncodeNfpmFactory())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", n.wrapRevertError(err))
	}

	out, err := DecodeNfpmFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}

	return out, nil
}

// WETH9 returns the wrapped native token req.Contract unwraps to and from.
func (n *NFPM) WETH9(ctx context.Context, opts *bind.CallOptions, req *WETH9Request) (*types.Address, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, EncodeWETH9())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", n.wrapRevertError(err))
	}

	out, err := DecodeWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}

	return out, nil
}

// TotalSupply returns how many position NFTs req.Contract has outstanding.
func (n *NFPM) TotalSupply(ctx context.Context, opts *bind.CallOptions, req *TotalSupplyRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: totalSupply: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, EncodeTotalSupply())
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: totalSupply: %w", n.wrapRevertError(err))
	}

	out, err := DecodeTotalSupply(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: totalSupply: %w", err)
	}

	return out, nil
}

// TokenOfOwnerByIndex returns the req.Index-th position token ID owned by req.Owner.
func (n *NFPM) TokenOfOwnerByIndex(ctx context.Context, opts *bind.CallOptions, req *TokenOfOwnerByIndexRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}

	calldata, err := EncodeTokenOfOwnerByIndex(req.Owner, req.Index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", n.wrapRevertError(err))
	}

	out, err := DecodeTokenOfOwnerByIndex(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}

	return out, nil
}

// TokenByIndex returns the req.Index-th position token ID across all owners.
func (n *NFPM) TokenByIndex(ctx context.Context, opts *bind.CallOptions, req *TokenByIndexRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}

	calldata, err := EncodeTokenByIndex(req.Index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", n.wrapRevertError(err))
	}

	out, err := DecodeTokenByIndex(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}

	return out, nil
}

// PermitTypehash returns req.Contract's ERC721 permit EIP-712 typehash.
func (n *NFPM) PermitTypehash(ctx context.Context, opts *bind.CallOptions, req *PermitTypehashRequest) ([32]byte, error) {
	if err := req.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: permitTypehash: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, EncodePermitTypehash())
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: permitTypehash: %w", n.wrapRevertError(err))
	}

	out, err := DecodePermitTypehash(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: permitTypehash: %w", err)
	}

	return out, nil
}

// DomainSeparator returns req.Contract's EIP-712 domain separator.
func (n *NFPM) DomainSeparator(ctx context.Context, opts *bind.CallOptions, req *DomainSeparatorRequest) ([32]byte, error) {
	if err := req.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: domainSeparator: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, EncodeDomainSeparator())
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: domainSeparator: %w", n.wrapRevertError(err))
	}

	out, err := DecodeDomainSeparator(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: domainSeparator: %w", err)
	}

	return out, nil
}

// Positions returns the position behind token req.TokenID.
func (n *NFPM) Positions(ctx context.Context, opts *bind.CallOptions, req *NFPMPositionsRequest) (*Position, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	calldata, err := EncodeNfpmPositions(req.TokenID)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", n.wrapRevertError(err))
	}

	out, err := DecodeNfpmPositions(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}

	return out, nil
}

// CallCreateAndInitializePoolIfNecessary runs createAndInitializePoolIfNecessary via eth_call as opts.From, returning its result or the revert.
func (n *NFPM) CallCreateAndInitializePoolIfNecessary(ctx context.Context, opts *bind.CallOptions, req *CreateAndInitializePoolIfNecessaryRequest) (pool *types.Address, err error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeCreateAndInitializePoolIfNecessary(req.Token0, req.Token1, req.Fee, req.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", n.wrapRevertError(err))
	}

	pool, err = DecodeCreateAndInitializePoolIfNecessary(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}

	return pool, nil
}

// CallMulticall runs multicall via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (n *NFPM) CallMulticall(ctx context.Context, opts *bind.CallOptions, req *NFPMMulticallRequest) (results [][]byte, err error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeMulticall(req.Data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, n.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", n.wrapRevertError(err))
	}

	results, err = DecodeMulticall(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}

	return results, nil
}

// CallMint runs mint via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (n *NFPM) CallMint(ctx context.Context, opts *bind.CallOptions, req *NFPMMintRequest) (tokenID, liquidity, amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeNfpmMint(req.Params.Token0, req.Params.Token1, req.Params.Fee, req.Params.TickLower, req.Params.TickUpper, req.Params.Amount0Desired, req.Params.Amount1Desired, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Recipient, req.Params.Deadline)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, n.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", n.wrapRevertError(err))
	}

	tokenID, liquidity, amount0, amount1, err = DecodeNfpmMint(data)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}

	return tokenID, liquidity, amount0, amount1, nil
}

// CallIncreaseLiquidity runs increaseLiquidity via eth_call as opts.From with req.NativeAmount as msg.value, returning its result or the revert.
func (n *NFPM) CallIncreaseLiquidity(ctx context.Context, opts *bind.CallOptions, req *IncreaseLiquidityRequest) (liquidity, amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeIncreaseLiquidity(req.Params.TokenId, req.Params.Amount0Desired, req.Params.Amount1Desired, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Deadline)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}

	callOpts := bind.NewCallOptions()
	if opts != nil {
		*callOpts = *opts
	}
	callOpts.Value = req.NativeAmount

	data, err := bind.Call(ctx, n.c, callOpts, req.Contract, calldata)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", n.wrapRevertError(err))
	}

	liquidity, amount0, amount1, err = DecodeIncreaseLiquidity(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}

	return liquidity, amount0, amount1, nil
}

// CallDecreaseLiquidity runs decreaseLiquidity via eth_call as opts.From, returning its result or the revert.
func (n *NFPM) CallDecreaseLiquidity(ctx context.Context, opts *bind.CallOptions, req *DecreaseLiquidityRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}

	calldata, err := EncodeDecreaseLiquidity(req.Params.TokenId, req.Params.Liquidity, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", n.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeDecreaseLiquidity(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}

	return amount0, amount1, nil
}

// CallCollect runs collect via eth_call as opts.From, returning its result or the revert.
func (n *NFPM) CallCollect(ctx context.Context, opts *bind.CallOptions, req *NFPMCollectRequest) (amount0, amount1 *big.Int, err error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	calldata, err := EncodeNfpmCollect(req.Params.TokenId, req.Params.Recipient, req.Params.Amount0Max, req.Params.Amount1Max)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	data, err := bind.Call(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", n.wrapRevertError(err))
	}

	amount0, amount1, err = DecodeNfpmCollect(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}

	return amount0, amount1, nil
}

// Permit approves req.Spender for req.TokenID with the owner's EIP-712 signature req.Sig, signed by opts.Signer.
func (n *NFPM) Permit(ctx context.Context, opts *bind.TransactOptions, req *PermitRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: permit: %w", err)
	}

	calldata, err := EncodeNFTPermit(req.Spender, req.TokenID, req.Deadline, req.Sig)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: permit: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: permit: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// CreateAndInitializePoolIfNecessary creates and initializes the (Token0, Token1, Fee) pool unless it already exists, signed by opts.Signer.
func (n *NFPM) CreateAndInitializePoolIfNecessary(ctx context.Context, opts *bind.TransactOptions, req *CreateAndInitializePoolIfNecessaryRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeCreateAndInitializePoolIfNecessary(req.Token0, req.Token1, req.Fee, req.SqrtPriceX96)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// UnwrapWETH9 unwraps req.Contract's whole WETH9 balance to req.Recipient as native currency, signed by opts.Signer.
func (n *NFPM) UnwrapWETH9(ctx context.Context, opts *bind.TransactOptions, req *NFPMUnwrapWETH9Request) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeUnwrapWETH9(req.AmountMinimum, req.Recipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// SweepToken sends req.Contract's whole req.Token balance to req.Recipient, signed by opts.Signer.
func (n *NFPM) SweepToken(ctx context.Context, opts *bind.TransactOptions, req *NFPMSweepTokenRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeSweepToken(req.Token, req.AmountMinimum, req.Recipient)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: sweepToken: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// RefundETH returns req.Contract's native balance to the sender, signed by opts.Signer.
func (n *NFPM) RefundETH(ctx context.Context, opts *bind.TransactOptions, req *NFPMRefundETHRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: refundETH: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata := EncodeRefundETH()

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: refundETH: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Multicall runs req.Data as one transaction on req.Contract, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (n *NFPM) Multicall(ctx context.Context, opts *bind.TransactOptions, req *NFPMMulticallRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeMulticall(req.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, n.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: multicall: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Mint mints a new position NFT for req.Params, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (n *NFPM) Mint(ctx context.Context, opts *bind.TransactOptions, req *NFPMMintRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: mint: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeNfpmMint(req.Params.Token0, req.Params.Token1, req.Params.Fee, req.Params.TickLower, req.Params.TickUpper, req.Params.Amount0Desired, req.Params.Amount1Desired, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Recipient, req.Params.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, n.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: mint: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// IncreaseLiquidity adds liquidity to an existing position, signed by opts.Signer with req.NativeAmount (nil sends none) as msg.value.
func (n *NFPM) IncreaseLiquidity(ctx context.Context, opts *bind.TransactOptions, req *IncreaseLiquidityRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w", err)
	}

	if opts != nil && opts.Value != nil {
		return nil, nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w: value must be nil", bind.ErrInvalidOptions)
	}

	calldata, err := EncodeIncreaseLiquidity(req.Params.TokenId, req.Params.Amount0Desired, req.Params.Amount1Desired, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w", err)
	}

	var txOpts *bind.TransactOptions
	if opts != nil {
		o := *opts
		o.Value = req.NativeAmount
		txOpts = &o
	}

	signed, hash, err := bind.Transact(ctx, n.c, txOpts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// DecreaseLiquidity removes liquidity from a position, crediting it as tokens owed, signed by opts.Signer.
func (n *NFPM) DecreaseLiquidity(ctx context.Context, opts *bind.TransactOptions, req *DecreaseLiquidityRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decreaseLiquidity: %w", err)
	}

	calldata, err := EncodeDecreaseLiquidity(req.Params.TokenId, req.Params.Liquidity, req.Params.Amount0Min, req.Params.Amount1Min, req.Params.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: decreaseLiquidity: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: decreaseLiquidity: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Collect collects tokens owed to a position, up to the requested maxima, signed by opts.Signer.
func (n *NFPM) Collect(ctx context.Context, opts *bind.TransactOptions, req *NFPMCollectRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}

	calldata, err := EncodeNfpmCollect(req.Params.TokenId, req.Params.Recipient, req.Params.Amount0Max, req.Params.Amount1Max)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: collect: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// Burn burns an empty, fully collected position NFT, signed by opts.Signer.
func (n *NFPM) Burn(ctx context.Context, opts *bind.TransactOptions, req *NFPMBurnRequest) (core.Transaction, *types.Hash, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}

	calldata, err := EncodeNfpmBurn(req.TokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}

	signed, hash, err := bind.Transact(ctx, n.c, opts, req.Contract, calldata)
	if err != nil {
		return signed, nil, fmt.Errorf("uniswapv3: burn: %w", n.wrapRevertError(err))
	}

	return signed, hash, nil
}

// wrapRevertError maps a known NonfungiblePositionManager revert reason in err to its sentinel.
func (_ *NFPM) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) {
		return err
	}
	if len(rpcErr.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	switch reason {
	case reasonTransactionTooOld:
		return fmt.Errorf("%w: %w", ErrTransactionTooOld, err)
	case reasonInsufficientWETH9:
		return fmt.Errorf("%w: %w", ErrInsufficientWETH9, err)
	case reasonInsufficientToken:
		return fmt.Errorf("%w: %w", ErrInsufficientToken, err)
	case reasonNotWETH9:
		return fmt.Errorf("%w: %w", ErrNotWETH9, err)
	case reasonTransferFromFailed:
		return fmt.Errorf("%w: %w", ErrTransferFromFailed, err)
	case reasonTokenTransferFailed:
		return fmt.Errorf("%w: %w", ErrTokenTransferFailed, err)
	case reasonApproveFailed:
		return fmt.Errorf("%w: %w", ErrApproveFailed, err)
	case reasonTransferETHFailed:
		return fmt.Errorf("%w: %w", ErrTransferETHFailed, err)
	case reasonInvalidTokenID:
		return fmt.Errorf("%w: %w", ErrInvalidTokenID, err)
	case reasonNotApproved:
		return fmt.Errorf("%w: %w", ErrNotApproved, err)
	case reasonPriceSlippageCheck:
		return fmt.Errorf("%w: %w", ErrPriceSlippageCheck, err)
	case reasonNotCleared:
		return fmt.Errorf("%w: %w", ErrNotCleared, err)
	case reasonPermitExpired:
		return fmt.Errorf("%w: %w", ErrPermitExpired, err)
	case reasonApprovalToCurrentOwner:
		return fmt.Errorf("%w: %w", ErrApprovalToCurrentOwner, err)
	case reasonUnauthorized:
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	case reasonInvalidSignature:
		return fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	default:
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}
}
