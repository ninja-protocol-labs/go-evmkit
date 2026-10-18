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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc2612"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IPair mirrors IUniswapV2Pair, embedding erc20/erc2612 for the ERC20+Permit surface they share
// byte-for-byte. No tx.Builder-taking Create* constructors here -- see CreateMint etc.
type IPair interface {
	erc20.IERC20Metadata
	erc2612.IERC2612

	MinimumLiquidity(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Token0(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	Token1(ctx context.Context, contract *types.Address, block string) (*types.Address, error)
	GetReserves(ctx context.Context, contract *types.Address, block string) (*Reserves, error)
	Price0CumulativeLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	Price1CumulativeLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	KLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error)
	PermitTypehash(ctx context.Context, contract *types.Address, block string) ([]byte, error)

	Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Skim(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Sync(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IPair = (*Pair)(nil)

// Pair calls any UniswapV2Pair contract over a Client; it holds no contract
// address, so one instance serves every pair.
type Pair struct {
	*erc20.ERC20
	*erc2612.ERC2612

	c rpc.Client
}

// NewPair returns a Pair calling pairs via c.
func NewPair(c rpc.Client) *Pair {
	return &Pair{
		ERC20:   erc20.New(c),
		ERC2612: erc2612.New(c),
		c:       c,
	}
}

// MinimumLiquidity returns contract's minimum liquidity burned to the zero address on first mint.
func (p *Pair) MinimumLiquidity(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodeMinimumLiquidity(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: minimumLiquidity: %w", err)
	}
	n, err := DecodeMinimumLiquidity(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: minimumLiquidity: %w", err)
	}
	return n, nil
}

// Factory returns the UniswapV2Factory that deployed contract.
func (p *Pair) Factory(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeFactory(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}
	addr, err := DecodeFactory(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: factory: %w", err)
	}
	return addr, nil
}

// Token0 returns contract's token0 (the lower-valued of its two tokens).
func (p *Pair) Token0(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeToken0(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token0: %w", err)
	}
	addr, err := DecodeToken0(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token0: %w", err)
	}
	return addr, nil
}

// Token1 returns contract's token1 (the higher-valued of its two tokens).
func (p *Pair) Token1(ctx context.Context, contract *types.Address, block string) (*types.Address, error) {
	data, err := p.call(ctx, nil, contract, EncodeToken1(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token1: %w", err)
	}
	addr, err := DecodeToken1(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: token1: %w", err)
	}
	return addr, nil
}

// GetReserves returns contract's current (reserve0, reserve1, blockTimestampLast).
func (p *Pair) GetReserves(ctx context.Context, contract *types.Address, block string) (*Reserves, error) {
	data, err := p.call(ctx, nil, contract, EncodeGetReserves(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getReserves: %w", err)
	}
	reserves, err := DecodeGetReserves(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: getReserves: %w", err)
	}
	return reserves, nil
}

// Price0CumulativeLast returns contract's cumulative price of token0, as of the last update.
func (p *Pair) Price0CumulativeLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodePrice0CumulativeLast(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price0CumulativeLast: %w", err)
	}
	n, err := DecodePrice0CumulativeLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price0CumulativeLast: %w", err)
	}
	return n, nil
}

// Price1CumulativeLast returns contract's cumulative price of token1, as of the last update.
func (p *Pair) Price1CumulativeLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodePrice1CumulativeLast(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price1CumulativeLast: %w", err)
	}
	n, err := DecodePrice1CumulativeLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: price1CumulativeLast: %w", err)
	}
	return n, nil
}

// KLast returns contract's reserve0*reserve1 as of the most recent liquidity event.
func (p *Pair) KLast(ctx context.Context, contract *types.Address, block string) (*big.Int, error) {
	data, err := p.call(ctx, nil, contract, EncodeKLast(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: kLast: %w", err)
	}
	n, err := DecodeKLast(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: kLast: %w", err)
	}
	return n, nil
}

// PermitTypehash returns contract's Permit EIP-712 typehash.
func (p *Pair) PermitTypehash(ctx context.Context, contract *types.Address, block string) ([]byte, error) {
	data, err := p.call(ctx, nil, contract, EncodePermitTypehash(), block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: permitTypehash: %w", err)
	}
	b, err := DecodePermitTypehash(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: permitTypehash: %w", err)
	}
	return b, nil
}

// CreateInitialize builds an unsigned initialize(token0, token1) transaction via b, sent by sender.
func (p *Pair) CreateInitialize(ctx context.Context, b tx.Builder, contract, sender, token0, token1 *types.Address) (core.Transaction, error) {
	calldata, err := EncodeInitialize(token0, token1)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createInitialize: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createInitialize: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Initialize broadcasts an already-signed initialize transaction.
func (p *Pair) Initialize(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: initialize: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: initialize: %w", err)
	}
	return hash, nil
}

// CallMint simulates mint(to) via eth_call, returning the liquidity it would mint.
func (p *Pair) CallMint(ctx context.Context, contract, sender, to *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeMint(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}
	liquidity, err := DecodeMint(data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: callMint: %w", err)
	}
	return liquidity, nil
}

// CreateMint builds an unsigned mint(to) transaction via b, sent by sender.
func (p *Pair) CreateMint(ctx context.Context, b tx.Builder, contract, sender, to *types.Address) (core.Transaction, error) {
	calldata, err := EncodeMint(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createMint: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createMint: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Mint broadcasts an already-signed mint transaction.
func (p *Pair) Mint(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: mint: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: mint: %w", err)
	}
	return hash, nil
}

// CallBurn simulates burn(to) via eth_call, returning the (amount0, amount1) it would withdraw.
func (p *Pair) CallBurn(ctx context.Context, contract, sender, to *types.Address, block string) (amount0, amount1 *big.Int, err error) {
	calldata, err := EncodeBurn(to)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}
	data, err := p.call(ctx, sender, contract, calldata, block)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}
	amount0, amount1, err = DecodeBurn(data)
	if err != nil {
		return nil, nil, fmt.Errorf("uniswapv2: callBurn: %w", err)
	}
	return amount0, amount1, nil
}

// CreateBurn builds an unsigned burn(to) transaction via b, sent by sender.
func (p *Pair) CreateBurn(ctx context.Context, b tx.Builder, contract, sender, to *types.Address) (core.Transaction, error) {
	calldata, err := EncodeBurn(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createBurn: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createBurn: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Burn broadcasts an already-signed burn transaction.
func (p *Pair) Burn(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: burn: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: burn: %w", err)
	}
	return hash, nil
}

// CreateSwap builds an unsigned swap(amount0Out, amount1Out, to, data) transaction via b, sent
// by sender.
func (p *Pair) CreateSwap(ctx context.Context, b tx.Builder, contract, sender *types.Address, amount0Out, amount1Out *big.Int, to *types.Address, data []byte) (core.Transaction, error) {
	calldata, err := EncodeSwap(amount0Out, amount1Out, to, data)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwap: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSwap: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Swap broadcasts an already-signed swap transaction.
func (p *Pair) Swap(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swap: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: swap: %w", err)
	}
	return hash, nil
}

// CreateSkim builds an unsigned skim(to) transaction via b, sent by sender.
func (p *Pair) CreateSkim(ctx context.Context, b tx.Builder, contract, sender, to *types.Address) (core.Transaction, error) {
	calldata, err := EncodeSkim(to)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSkim: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSkim: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Skim broadcasts an already-signed skim transaction.
func (p *Pair) Skim(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: skim: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: skim: %w", err)
	}
	return hash, nil
}

// CreateSync builds an unsigned sync() transaction via b, sent by sender.
func (p *Pair) CreateSync(ctx context.Context, b tx.Builder, contract, sender *types.Address) (core.Transaction, error) {
	unsigned, err := b.Build(sender, contract, nil, EncodeSync()).Pack(ctx, p.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: createSync: %w", p.wrapRevertError(err))
	}
	return unsigned, nil
}

// Sync broadcasts an already-signed sync transaction.
func (p *Pair) Sync(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: sync: %w", err)
	}
	hash, err := tx.Broadcast(ctx, p.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv2: sync: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata, decoding a reverted require(cond, reason) to its sentinel.
func (p *Pair) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := p.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, p.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
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
