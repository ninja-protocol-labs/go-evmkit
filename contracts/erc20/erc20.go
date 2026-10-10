package erc20

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

var (
	nameFn         = abi.NewFunction("name", abi.NewTypes(), abi.NewTypes(abi.String))
	symbolFn       = abi.NewFunction("symbol", abi.NewTypes(), abi.NewTypes(abi.String))
	decimalsFn     = abi.NewFunction("decimals", abi.NewTypes(), abi.NewTypes(abi.Uint8))
	totalSupplyFn  = abi.NewFunction("totalSupply", abi.NewTypes(), abi.NewTypes(abi.Uint256))
	balanceOfFn    = abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	transferFn     = abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
	allowanceFn    = abi.NewFunction("allowance", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Uint256))
	approveFn      = abi.NewFunction("approve", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
	transferFromFn = abi.NewFunction("transferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
)

var (
	transferEvent = abi.NewEvent("Transfer", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
	}, false)
	approvalEvent = abi.NewEvent("Approval", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
	}, false)
)

// Custom errors from IERC20Errors (ERC-6093), for decoding reverts via abi.Error.Decode.
var (
	errInsufficientBalance   = abi.NewError("ERC20InsufficientBalance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256))
	errInvalidSender         = abi.NewError("ERC20InvalidSender", abi.NewTypes(abi.Address))
	errInvalidReceiver       = abi.NewError("ERC20InvalidReceiver", abi.NewTypes(abi.Address))
	errInsufficientAllowance = abi.NewError("ERC20InsufficientAllowance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256))
	errInvalidApprover       = abi.NewError("ERC20InvalidApprover", abi.NewTypes(abi.Address))
	errInvalidSpender        = abi.NewError("ERC20InvalidSpender", abi.NewTypes(abi.Address))
)

// ErrInsufficientBalance, etc. mirror IERC20Errors' custom errors, for errors.Is.
var (
	ErrInsufficientBalance   = errors.New("erc20: insufficient balance")
	ErrInvalidSender         = errors.New("erc20: invalid sender")
	ErrInvalidReceiver       = errors.New("erc20: invalid receiver")
	ErrInsufficientAllowance = errors.New("erc20: insufficient allowance")
	ErrInvalidApprover       = errors.New("erc20: invalid approver")
	ErrInvalidSpender        = errors.New("erc20: invalid spender")
	ErrUnknown               = errors.New("erc20: unknown revert")
)

// IERC20 is implemented by *ERC20.
type IERC20 interface {
	TotalSupply(ctx context.Context, token *types.Address, block string) (*big.Int, error)
	BalanceOf(ctx context.Context, token, addr *types.Address, block string) (*big.Int, error)
	Allowance(ctx context.Context, token, owner, spender *types.Address, block string) (*big.Int, error)
	Transfer(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
	TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IERC20 = (*ERC20)(nil)

// IERC20Metadata mirrors IERC20Metadata.sol, including its IERC20 embedding.
// name/symbol/decimals are optional per EIP-20.
type IERC20Metadata interface {
	IERC20
	Name(ctx context.Context, token *types.Address, block string) (string, error)
	Symbol(ctx context.Context, token *types.Address, block string) (string, error)
	Decimals(ctx context.Context, token *types.Address, block string) (uint8, error)
}

var _ IERC20Metadata = (*ERC20)(nil)

// ERC20 calls any ERC-20 token contract over a Client; it holds no token
// address, so one instance serves every token.
type ERC20 struct {
	c rpc.Client
}

// New returns an ERC20 calling tokens via cli.
func New(cli rpc.Client) *ERC20 {
	return &ERC20{
		c: cli,
	}
}

// TotalSupply returns token's total supply.
func (e *ERC20) TotalSupply(ctx context.Context, token *types.Address, block string) (*big.Int, error) {
	data, err := e.call(ctx, nil, token, EncodeTotalSupply(), block)
	if err != nil {
		return nil, fmt.Errorf("erc20: totalSupply: %w", err)
	}
	n, err := DecodeTotalSupply(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: totalSupply: %w", err)
	}
	return n, nil
}

// BalanceOf returns address's balance of token.
func (e *ERC20) BalanceOf(ctx context.Context, token, addr *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeBalanceOf(addr)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}
	data, err := e.call(ctx, nil, token, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}
	n, err := DecodeBalanceOf(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: balanceOf: %w", err)
	}
	return n, nil
}

// Allowance returns how much spender may still spend of owner's token.
func (e *ERC20) Allowance(ctx context.Context, token, owner, spender *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeAllowance(owner, spender)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}
	data, err := e.call(ctx, nil, token, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}
	n, err := DecodeAllowance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowance: %w", err)
	}
	return n, nil
}

// Name returns token's name.
func (e *ERC20) Name(ctx context.Context, token *types.Address, block string) (string, error) {
	data, err := e.call(ctx, nil, token, EncodeName(), block)
	if err != nil {
		return "", fmt.Errorf("erc20: name: %w", err)
	}
	s, err := DecodeName(data)
	if err != nil {
		return "", fmt.Errorf("erc20: name: %w", err)
	}
	return s, nil
}

// Symbol returns token's symbol.
func (e *ERC20) Symbol(ctx context.Context, token *types.Address, block string) (string, error) {
	data, err := e.call(ctx, nil, token, EncodeSymbol(), block)
	if err != nil {
		return "", fmt.Errorf("erc20: symbol: %w", err)
	}
	s, err := DecodeSymbol(data)
	if err != nil {
		return "", fmt.Errorf("erc20: symbol: %w", err)
	}
	return s, nil
}

// Decimals returns token's decimals.
func (e *ERC20) Decimals(ctx context.Context, token *types.Address, block string) (uint8, error) {
	data, err := e.call(ctx, nil, token, EncodeDecimals(), block)
	if err != nil {
		return 0, fmt.Errorf("erc20: decimals: %w", err)
	}
	d, err := DecodeDecimals(data)
	if err != nil {
		return 0, fmt.Errorf("erc20: decimals: %w", err)
	}
	return d, nil
}

// CreateTransfer builds an unsigned transfer(to, value) transaction via b.
func (e *ERC20) CreateTransfer(ctx context.Context, b tx.Builder, token, from, to *types.Address, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTransfer(to, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: createTransfer: %w", err)
	}

	unsigned, err := b.Build(from, token, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc20: createTransfer: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferGas estimates the gas a transfer transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC20) TransferGas(ctx context.Context, b tx.Builder, token, from, to *types.Address, value *big.Int) (uint64, error) {
	calldata, err := EncodeTransfer(to, value)
	if err != nil {
		return 0, fmt.Errorf("erc20: transferGas: %w", err)
	}

	gas, err := b.Build(from, token, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc20: transferGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// Transfer broadcasts an already-signed transfer transaction.
func (e *ERC20) Transfer(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc20: transfer: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc20: transfer: %w", err)
	}
	return hash, nil
}

// CallTransfer simulates transfer(to, value) as from via eth_call, catching
// a revert before spending gas on a real transfer.
func (e *ERC20) CallTransfer(ctx context.Context, token, from, to *types.Address, value *big.Int, block string) (bool, error) {
	calldata, err := EncodeTransfer(to, value)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}

	data, err := e.call(ctx, from, token, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}

	ok, err := DecodeTransfer(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransfer: %w", err)
	}
	return ok, nil
}

// CreateApprove builds an unsigned approve(spender, value) transaction via b.
func (e *ERC20) CreateApprove(ctx context.Context, b tx.Builder, token, from, spender *types.Address, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeApprove(spender, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: createApprove: %w", err)
	}

	unsigned, err := b.Build(from, token, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc20: createApprove: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// ApproveGas estimates the gas an approve transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC20) ApproveGas(ctx context.Context, b tx.Builder, token, from, spender *types.Address, value *big.Int) (uint64, error) {
	calldata, err := EncodeApprove(spender, value)
	if err != nil {
		return 0, fmt.Errorf("erc20: approveGas: %w", err)
	}

	gas, err := b.Build(from, token, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc20: approveGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// Approve broadcasts an already-signed approve transaction.
func (e *ERC20) Approve(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc20: approve: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc20: approve: %w", err)
	}
	return hash, nil
}

// CallApprove simulates approve(spender, value) as from via eth_call,
// catching a revert before spending gas on a real approve.
func (e *ERC20) CallApprove(ctx context.Context, token, from, spender *types.Address, value *big.Int, block string) (bool, error) {
	calldata, err := EncodeApprove(spender, value)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}

	data, err := e.call(ctx, from, token, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}

	ok, err := DecodeApprove(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callApprove: %w", err)
	}
	return ok, nil
}

// CreateTransferFrom builds an unsigned transferFrom(from, to, value) transaction via b, sent by sender.
func (e *ERC20) CreateTransferFrom(ctx context.Context, b tx.Builder, token, sender, from, to *types.Address, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeTransferFrom(from, to, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: createTransferFrom: %w", err)
	}

	unsigned, err := b.Build(sender, token, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc20: createTransferFrom: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// TransferFromGas estimates the gas a transferFrom transaction would use, without
// resolving chainID/nonce/gasPrice or building the transaction itself.
func (e *ERC20) TransferFromGas(ctx context.Context, b tx.Builder, token, sender, from, to *types.Address, value *big.Int) (uint64, error) {
	calldata, err := EncodeTransferFrom(from, to, value)
	if err != nil {
		return 0, fmt.Errorf("erc20: transferFromGas: %w", err)
	}

	gas, err := b.Build(sender, token, nil, calldata).Gas(ctx, e.c)
	if err != nil {
		return 0, fmt.Errorf("erc20: transferFromGas: %w", e.wrapRevertError(err))
	}
	return gas, nil
}

// TransferFrom broadcasts an already-signed transferFrom transaction.
func (e *ERC20) TransferFrom(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc20: transferFrom: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc20: transferFrom: %w", err)
	}
	return hash, nil
}

// CallTransferFrom simulates transferFrom(from, to, value) as sender via eth_call,
// catching a revert before spending gas on a real transferFrom.
func (e *ERC20) CallTransferFrom(ctx context.Context, token, sender, from, to *types.Address, value *big.Int, block string) (bool, error) {
	calldata, err := EncodeTransferFrom(from, to, value)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}

	data, err := e.call(ctx, sender, token, calldata, block)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}

	ok, err := DecodeTransferFrom(data)
	if err != nil {
		return false, fmt.Errorf("erc20: callTransferFrom: %w", err)
	}
	return ok, nil
}

// Metadata batches name/symbol/decimals/totalSupply on token into one call
// via the Multicall3 contract deployed at multicall.
func (e *ERC20) Metadata(ctx context.Context, multicall, token *types.Address, allowFailure bool, block string) (*Metadata, error) {
	calldata, err := EncodeMetadata(token, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}

	m, err := DecodeMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadata: %w", err)
	}
	return m, nil
}

// MetadataWithBalance is Metadata plus addr's balance in the same call, via
// the Multicall3 contract deployed at multicall.
func (e *ERC20) MetadataWithBalance(ctx context.Context, multicall, token, addr *types.Address, allowFailure bool, block string) (*MetadataWithBalance, error) {
	calldata, err := EncodeMetadataWithBalance(token, addr, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}

	m, err := DecodeMetadataWithBalance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: metadataWithBalance: %w", err)
	}
	return m, nil
}

// AllowanceWithBalance batches allowance(owner, spender) and both their
// balances into one call, via the Multicall3 contract deployed at multicall.
func (e *ERC20) AllowanceWithBalance(ctx context.Context, multicall, token, owner, spender *types.Address, allowFailure bool, block string) (*AllowanceWithBalance, error) {
	calldata, err := EncodeAllowanceWithBalance(token, owner, spender, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}

	awb, err := DecodeAllowanceWithBalance(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: allowanceWithBalance: %w", err)
	}
	return awb, nil
}

// TokenBalances batches balanceOf(address) for each address on token into
// one call, via the Multicall3 contract deployed at multicall.
func (e *ERC20) TokenBalances(ctx context.Context, multicall, token *types.Address, addresses []*types.Address, allowFailure bool, block string) (*TokenBalances, error) {
	calldata, err := EncodeTokenBalances(token, addresses, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}

	tb, err := DecodeTokenBalances(token, addresses, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: tokenBalances: %w", err)
	}
	return tb, nil
}

// AddressBalances batches balanceOf(address) across each token in tokens
// into one call, via the Multicall3 contract deployed at multicall.
func (e *ERC20) AddressBalances(ctx context.Context, multicall, address *types.Address, tokens []*types.Address, allowFailure bool, block string) (*AddressBalances, error) {
	calldata, err := EncodeAddressBalances(address, tokens, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}

	ab, err := DecodeAddressBalances(address, tokens, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: addressBalances: %w", err)
	}
	return ab, nil
}

// BalancePairs batches balanceOf(addresses[i]) on tokens[i] for each i into
// one call (paired positionally, not every combination), via the Multicall3
// contract deployed at multicall.
func (e *ERC20) BalancePairs(ctx context.Context, multicall *types.Address, tokens, addresses []*types.Address, allowFailure bool, block string) ([]TokenAddressBalance, error) {
	calldata, err := EncodeBalancePairs(tokens, addresses, allowFailure)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}

	data, err := e.call(ctx, nil, multicall, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}

	pairs, err := DecodeBalancePairs(tokens, addresses, data)
	if err != nil {
		return nil, fmt.Errorf("erc20: balancePairs: %w", err)
	}
	return pairs, nil
}

// call eth_calls token with calldata (from msg.sender from, if non-nil), decoding a reverted custom error to its sentinel.
func (e *ERC20) call(ctx context.Context, from, token *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   token.String(),
		"data": encoding.Hex.EncodePrefixed(calldata),
	}
	if from != nil {
		params["from"] = from.String()
	}

	var raw string
	if err := e.c.Call(ctx, rpc.ETHCall("call", params, block, &raw)); err != nil {
		return nil, e.wrapRevertError(err)
	}
	return encoding.Hex.Decode(raw)
}

// wrapRevertError maps a known IERC20Errors custom error in err to its sentinel.
func (_ *ERC20) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errInsufficientBalance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientBalance, err)
	case errInvalidSender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSender, err)
	case errInvalidReceiver.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidReceiver, err)
	case errInsufficientAllowance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientAllowance, err)
	case errInvalidApprover.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidApprover, err)
	case errInvalidSpender.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSpender, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
