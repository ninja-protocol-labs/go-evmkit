package v3

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// INFPMImmutables mirrors PeripheryImmutableState, the only state NFPM itself holds outside its
// positions.
type INFPMImmutables interface {
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
}

// INFPMEnumerable mirrors IERC721Enumerable.
type INFPMEnumerable interface {
	TotalSupply(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	TokenOfOwnerByIndex(ctx context.Context, p *TokenOfOwnerByIndexParams, block string) (*big.Int, error)
	TokenByIndex(ctx context.Context, contract *types.Address, index *big.Int, block string) (*big.Int, error)
}

// INFPMPermit mirrors IERC721Permit. No tx.Builder-taking Create* constructor here -- see
// CreatePermit.
type INFPMPermit interface {
	PermitTypehash(ctx context.Context, contract *types.Address, block string) ([32]byte, error)
	DomainSeparator(ctx context.Context, contract *types.Address, block string) ([32]byte, error)
	Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// INFPMPoolInitializer mirrors IPoolInitializer. No tx.Builder-taking Create* constructor here --
// see CreateCreateAndInitializePoolIfNecessary.
type INFPMPoolInitializer interface {
	CreateAndInitializePoolIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// INFPMPayments mirrors IPeripheryPayments. No tx.Builder-taking Create* constructors here -- see
// CreateUnwrapWETH9, CreateSweepToken, and CreateRefundETH.
type INFPMPayments interface {
	UnwrapWETH9(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	SweepToken(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	RefundETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// INFPMMulticall mirrors Multicall. No tx.Builder-taking Create* constructor here -- see
// CreateMulticall.
type INFPMMulticall interface {
	Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// INFPMActions mirrors INonfungiblePositionManager's own position-management surface. No
// tx.Builder-taking Create* constructors here -- see CreateMint, CreateIncreaseLiquidity,
// CreateDecreaseLiquidity, CreateCollect, and CreateBurn.
type INFPMActions interface {
	Positions(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*Position, error)
	Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	IncreaseLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	DecreaseLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Collect(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

// INFPM mirrors NonfungiblePositionManager, composed the same way its Solidity base contracts
// are: one small interface per concern, plus the full ERC721 (+ Metadata) surface it wraps
// positions in.
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

// NFPM calls any NonfungiblePositionManager contract over a Client; it holds no contract
// address, so one instance serves every deployment on every network. The embedded *erc721.ERC721
// serves the plain ERC721 (+ Metadata) surface NFPM wraps positions in.
type NFPM struct {
	*erc721.ERC721

	c rpc.Client
}

// NewNFPM returns an NFPM calling position managers via c.
func NewNFPM(c rpc.Client) *NFPM {
	return &NFPM{
		ERC721: erc721.New(c),
		c:      c,
	}
}

// Factory returns the UniswapV3Factory contract's address.
func (n *NFPM) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := n.call(ctx, nil, contract, nil, EncodeNfpmFactory(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	addr, err := DecodeNfpmFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: factory: %w", err)
	}
	return addr, nil
}

// WETH9 returns the WETH9 contract's address.
func (n *NFPM) WETH9(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := n.call(ctx, nil, contract, nil, EncodeWETH9(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	addr, err := DecodeWETH9(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: weth9: %w", err)
	}
	return addr, nil
}

// TotalSupply returns the number of position NFTs in existence.
func (n *NFPM) TotalSupply(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := n.call(ctx, nil, contract, nil, EncodeTotalSupply(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: totalSupply: %w", err)
	}
	n_, err := DecodeTotalSupply(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: totalSupply: %w", err)
	}
	return n_, nil
}

// TokenOfOwnerByIndex returns the tokenId owned by owner at index.
func (n *NFPM) TokenOfOwnerByIndex(ctx context.Context, p *TokenOfOwnerByIndexParams, block string) (*big.Int, error) {
	calldata, err := EncodeTokenOfOwnerByIndex(p.Owner, p.Index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}
	data, err := n.call(ctx, nil, p.NFPM, nil, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}
	tokenId, err := DecodeTokenOfOwnerByIndex(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenOfOwnerByIndex: %w", err)
	}
	return tokenId, nil
}

// TokenByIndex returns the tokenId at index, among every position NFT in existence.
func (n *NFPM) TokenByIndex(ctx context.Context, contract *types.Address, index *big.Int, block string) (*big.Int, error) {
	calldata, err := EncodeTokenByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}
	data, err := n.call(ctx, nil, contract, nil, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}
	tokenId, err := DecodeTokenByIndex(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: tokenByIndex: %w", err)
	}
	return tokenId, nil
}

// PermitTypehash returns the EIP-712 typehash used in the token permit signature.
func (n *NFPM) PermitTypehash(ctx context.Context, contract *types.Address, block string) ([32]byte, error) {
	data, err := n.call(ctx, nil, contract, nil, EncodePermitTypehash(), block)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: permitTypehash: %w", err)
	}
	h, err := DecodePermitTypehash(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: permitTypehash: %w", err)
	}
	return h, nil
}

// DomainSeparator returns the EIP-712 domain separator used in the token permit signature.
func (n *NFPM) DomainSeparator(ctx context.Context, contract *types.Address, block string) ([32]byte, error) {
	data, err := n.call(ctx, nil, contract, nil, EncodeDomainSeparator(), block)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: domainSeparator: %w", err)
	}
	h, err := DecodeDomainSeparator(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("uniswapv3: domainSeparator: %w", err)
	}
	return h, nil
}

// CreatePermit builds an unsigned permit(spender, tokenId, deadline, v, r, s) transaction via b,
// sent by sender.
func (n *NFPM) CreatePermit(ctx context.Context, b tx.Builder, p *PermitParams) (core.Transaction, error) {
	calldata, err := EncodeNFTPermit(p.Spender, p.TokenId, p.Deadline, p.Sig)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createPermit: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createPermit: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// Permit broadcasts an already-signed permit transaction.
func (n *NFPM) Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: permit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: permit: %w", err)
	}
	return hash, nil
}

// CreateCreateAndInitializePoolIfNecessary builds an unsigned
// createAndInitializePoolIfNecessary(token0, token1, fee, sqrtPriceX96) transaction via b, sent
// by sender.
func (n *NFPM) CreateCreateAndInitializePoolIfNecessary(ctx context.Context, b tx.Builder, p *CreateAndInitializePoolIfNecessaryParams) (core.Transaction, error) {
	calldata, err := EncodeCreateAndInitializePoolIfNecessary(p.Token0, p.Token1, p.Fee, p.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCreateAndInitializePoolIfNecessary: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCreateAndInitializePoolIfNecessary: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCreateAndInitializePoolIfNecessary simulates createAndInitializePoolIfNecessary(token0,
// token1, fee, sqrtPriceX96) via eth_call, as sender, returning the pool's address -- without
// broadcasting anything.
func (n *NFPM) CallCreateAndInitializePoolIfNecessary(ctx context.Context, p *CreateAndInitializePoolIfNecessaryParams, block string) (*types.Address, error) {
	calldata, err := EncodeCreateAndInitializePoolIfNecessary(p.Token0, p.Token1, p.Fee, p.SqrtPriceX96)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}
	data, err := n.call(ctx, p.Sender, p.NFPM, nil, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}
	addr, err := DecodeCreateAndInitializePoolIfNecessary(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callCreateAndInitializePoolIfNecessary: %w", err)
	}
	return addr, nil
}

// CreateAndInitializePoolIfNecessary broadcasts an already-signed
// createAndInitializePoolIfNecessary transaction.
func (n *NFPM) CreateAndInitializePoolIfNecessary(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createAndInitializePoolIfNecessary: %w", err)
	}
	return hash, nil
}

// CreateUnwrapWETH9 builds an unsigned unwrapWETH9(amountMinimum, recipient) transaction via b,
// sent by sender.
func (n *NFPM) CreateUnwrapWETH9(ctx context.Context, b tx.Builder, p *NFPMUnwrapWETH9Params) (core.Transaction, error) {
	calldata, err := EncodeUnwrapWETH9(p.AmountMinimum, p.Recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createUnwrapWETH9: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// UnwrapWETH9 broadcasts an already-signed unwrapWETH9 transaction.
func (n *NFPM) UnwrapWETH9(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: unwrapWETH9: %w", err)
	}
	return hash, nil
}

// CreateSweepToken builds an unsigned sweepToken(token, amountMinimum, recipient) transaction
// via b, sent by sender.
func (n *NFPM) CreateSweepToken(ctx context.Context, b tx.Builder, p *NFPMSweepTokenParams) (core.Transaction, error) {
	calldata, err := EncodeSweepToken(p.Token, p.AmountMinimum, p.Recipient)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepToken: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createSweepToken: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// SweepToken broadcasts an already-signed sweepToken transaction.
func (n *NFPM) SweepToken(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sweepToken: %w", err)
	}
	return hash, nil
}

// CreateRefundETH builds an unsigned refundETH() transaction via b, sent by sender.
func (n *NFPM) CreateRefundETH(ctx context.Context, b tx.Builder, p *NFPMRefundETHParams) (core.Transaction, error) {
	unsigned, err := b.Build(p.Sender, p.NFPM, nil, EncodeRefundETH()).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createRefundETH: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// RefundETH broadcasts an already-signed refundETH transaction.
func (n *NFPM) RefundETH(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: refundETH: %w", err)
	}
	return hash, nil
}

// CreateMulticall builds an unsigned multicall(data) transaction via b, sent by sender with
// value wei attached (value may be nil).
func (n *NFPM) CreateMulticall(ctx context.Context, b tx.Builder, p *NFPMMulticallParams) (core.Transaction, error) {
	calldata, err := EncodeMulticall(p.Data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMulticall: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, p.Value, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMulticall: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallMulticall simulates multicall(data) via eth_call, as sender with value wei attached,
// returning each sub-call's own return data -- without broadcasting anything.
func (n *NFPM) CallMulticall(ctx context.Context, p *NFPMMulticallParams, block string) ([][]byte, error) {
	calldata, err := EncodeMulticall(p.Data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	raw, err := n.call(ctx, p.Sender, p.NFPM, p.Value, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	results, err := DecodeMulticall(raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: callMulticall: %w", err)
	}
	return results, nil
}

// Multicall broadcasts an already-signed multicall transaction.
func (n *NFPM) Multicall(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: multicall: %w", err)
	}
	return hash, nil
}

// Positions returns tokenId's position information.
func (n *NFPM) Positions(ctx context.Context, contract *types.Address, tokenId *big.Int, block string) (*Position, error) {
	calldata, err := EncodeNfpmPositions(tokenId)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	data, err := n.call(ctx, nil, contract, nil, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	position, err := DecodeNfpmPositions(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: positions: %w", err)
	}
	return position, nil
}

// CreateMint builds an unsigned mint(params) transaction via b, sent by sender with value wei
// attached (value may be nil for a non-ETH-input mint).
func (n *NFPM) CreateMint(ctx context.Context, b tx.Builder, contract, sender *types.Address, value *big.Int, params *MintParams) (core.Transaction, error) {
	calldata, err := EncodeNfpmMint(params.Token0, params.Token1, params.Fee, params.TickLower, params.TickUpper, params.Amount0Desired, params.Amount1Desired, params.Amount0Min, params.Amount1Min, params.Recipient, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMint: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createMint: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallMint simulates mint(params) via eth_call, as sender with value wei attached, returning the
// tokenId/liquidity/amount0/amount1 that would result -- without broadcasting anything.
func (n *NFPM) CallMint(ctx context.Context, contract, sender *types.Address, value *big.Int, params *MintParams, block string) (tokenId, liquidity, amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeNfpmMint(params.Token0, params.Token1, params.Fee, params.TickLower, params.TickUpper, params.Amount0Desired, params.Amount1Desired, params.Amount0Min, params.Amount1Min, params.Recipient, params.Deadline)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	data, err := n.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	tokenId, liquidity, amount0, amount1, err = DecodeNfpmMint(data)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("uniswapv3: callMint: %w", err)
	}
	return tokenId, liquidity, amount0, amount1, nil
}

// Mint broadcasts an already-signed mint transaction.
func (n *NFPM) Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: mint: %w", err)
	}
	return hash, nil
}

// CreateIncreaseLiquidity builds an unsigned increaseLiquidity(params) transaction via b, sent by
// sender with value wei attached (value may be nil for a non-ETH-input increase).
func (n *NFPM) CreateIncreaseLiquidity(ctx context.Context, b tx.Builder, contract, sender *types.Address, value *big.Int, params *IncreaseLiquidityParams) (core.Transaction, error) {
	calldata, err := EncodeIncreaseLiquidity(params.TokenId, params.Amount0Desired, params.Amount1Desired, params.Amount0Min, params.Amount1Min, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createIncreaseLiquidity: %w", err)
	}

	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createIncreaseLiquidity: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallIncreaseLiquidity simulates increaseLiquidity(params) via eth_call, as sender with value
// wei attached, returning the liquidity/amount0/amount1 that would result -- without
// broadcasting anything.
func (n *NFPM) CallIncreaseLiquidity(ctx context.Context, contract, sender *types.Address, value *big.Int, params *IncreaseLiquidityParams, block string) (liquidity, amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeIncreaseLiquidity(params.TokenId, params.Amount0Desired, params.Amount1Desired, params.Amount0Min, params.Amount1Min, params.Deadline)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}
	data, err := n.call(ctx, sender, contract, value, calldata, block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}
	liquidity, amount0, amount1, err = DecodeIncreaseLiquidity(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("uniswapv3: callIncreaseLiquidity: %w", err)
	}
	return liquidity, amount0, amount1, nil
}

// IncreaseLiquidity broadcasts an already-signed increaseLiquidity transaction.
func (n *NFPM) IncreaseLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: increaseLiquidity: %w", err)
	}
	return hash, nil
}

// CreateDecreaseLiquidity builds an unsigned decreaseLiquidity(params) transaction via b, sent by
// sender -- who must be the position's owner or operator.
func (n *NFPM) CreateDecreaseLiquidity(ctx context.Context, b tx.Builder, contract, sender *types.Address, params *DecreaseLiquidityParams) (core.Transaction, error) {
	calldata, err := EncodeDecreaseLiquidity(params.TokenId, params.Liquidity, params.Amount0Min, params.Amount1Min, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createDecreaseLiquidity: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createDecreaseLiquidity: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallDecreaseLiquidity simulates decreaseLiquidity(params) via eth_call, as sender, returning
// the amount0/amount1 that would be accounted to the position -- without broadcasting anything.
func (n *NFPM) CallDecreaseLiquidity(ctx context.Context, contract, sender *types.Address, params *DecreaseLiquidityParams, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeDecreaseLiquidity(params.TokenId, params.Liquidity, params.Amount0Min, params.Amount1Min, params.Deadline)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}
	data, err := n.call(ctx, sender, contract, nil, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}
	amount0, amount1, err = DecodeDecreaseLiquidity(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callDecreaseLiquidity: %w", err)
	}
	return amount0, amount1, nil
}

// DecreaseLiquidity broadcasts an already-signed decreaseLiquidity transaction.
func (n *NFPM) DecreaseLiquidity(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decreaseLiquidity: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: decreaseLiquidity: %w", err)
	}
	return hash, nil
}

// CreateCollect builds an unsigned collect(params) transaction via b, sent by sender -- who must
// be the position's owner or operator.
func (n *NFPM) CreateCollect(ctx context.Context, b tx.Builder, contract, sender *types.Address, params *CollectParams) (core.Transaction, error) {
	calldata, err := EncodeNfpmCollect(params.TokenId, params.Recipient, params.Amount0Max, params.Amount1Max)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollect: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createCollect: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// CallCollect simulates collect(params) via eth_call, as sender, returning the amount0/amount1
// that would be collected -- without broadcasting anything.
func (n *NFPM) CallCollect(ctx context.Context, contract, sender *types.Address, params *CollectParams, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeNfpmCollect(params.TokenId, params.Recipient, params.Amount0Max, params.Amount1Max)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	data, err := n.call(ctx, sender, contract, nil, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	amount0, amount1, err = DecodeNfpmCollect(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv3: callCollect: %w", err)
	}
	return amount0, amount1, nil
}

// Collect broadcasts an already-signed collect transaction.
func (n *NFPM) Collect(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: collect: %w", err)
	}
	return hash, nil
}

// CreateBurn builds an unsigned burn(tokenId) transaction via b, sent by sender -- who must be
// the position's owner or operator. The position must have 0 liquidity and 0 tokens owed.
func (n *NFPM) CreateBurn(ctx context.Context, b tx.Builder, p *NFPMBurnParams) (core.Transaction, error) {
	calldata, err := EncodeNfpmBurn(p.TokenId)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createBurn: %w", err)
	}

	unsigned, err := b.Build(p.Sender, p.NFPM, nil, calldata).Pack(ctx, n.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: createBurn: %w", n.wrapRevertError(err))
	}
	return unsigned, nil
}

// Burn broadcasts an already-signed burn transaction.
func (n *NFPM) Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}
	hash, err := tx.Broadcast(ctx, n.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: burn: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, optionally simulating from a specific sender with value
// wei attached.
func (n *NFPM) call(ctx context.Context, from, contract *types.Address, value *big.Int, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}
	if value != nil && value.Sign() != 0 {
		params["value"] = "0x" + value.Text(16)
	}

	var raw string
	if err := n.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, n.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
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
