// Package erc20 encodes calls to and decodes results from an ERC-20 token
// contract. Go types mirror the interfaces in solidity/interfaces/IERC20.sol
// and solidity/interfaces/IERC20Metadata.sol.
package erc20

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IERC20/IERC20Metadata, for encoding calls and decoding results.
var (
	nameFn         = abi.NewFunction("name", nil, abi.NewTypes(abi.String))
	symbolFn       = abi.NewFunction("symbol", nil, abi.NewTypes(abi.String))
	decimalsFn     = abi.NewFunction("decimals", nil, abi.NewTypes(abi.Uint8))
	totalSupplyFn  = abi.NewFunction("totalSupply", nil, abi.NewTypes(abi.Uint256))
	balanceOfFn    = abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	transferFn     = abi.NewFunction("transfer", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
	allowanceFn    = abi.NewFunction("allowance", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Uint256))
	approveFn      = abi.NewFunction("approve", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
	transferFromFn = abi.NewFunction("transferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), abi.NewTypes(abi.Bool))
)

// Events from IERC20, for decoding logs via Event.Decode.
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

// EncodeName returns the calldata for name().
func EncodeName() []byte {
	return nameFn.SelectorBytes()
}

// DecodeName decodes the return data of name.
func DecodeName(data []byte) (string, error) {
	s, err := nameFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc20: decodeName: %w", err)
	}
	return s, nil
}

// EncodeSymbol returns the calldata for symbol().
func EncodeSymbol() []byte {
	return symbolFn.SelectorBytes()
}

// DecodeSymbol decodes the return data of symbol.
func DecodeSymbol(data []byte) (string, error) {
	s, err := symbolFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc20: decodeSymbol: %w", err)
	}
	return s, nil
}

// EncodeDecimals returns the calldata for decimals().
func EncodeDecimals() []byte {
	return decimalsFn.SelectorBytes()
}

// DecodeDecimals decodes the return data of decimals.
func DecodeDecimals(data []byte) (uint8, error) {
	n, err := decimalsFn.DecodeSingleReturn[uint8](data)
	if err != nil {
		return 0, fmt.Errorf("erc20: decodeDecimals: %w", err)
	}
	return n, nil
}

// EncodeTotalSupply returns the calldata for totalSupply().
func EncodeTotalSupply() []byte {
	return totalSupplyFn.SelectorBytes()
}

// DecodeTotalSupply decodes the return data of totalSupply.
func DecodeTotalSupply(data []byte) (*big.Int, error) {
	n, err := totalSupplyFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeTotalSupply: %w", err)
	}
	return n, nil
}

// EncodeBalanceOf returns the calldata for balanceOf(account).
func EncodeBalanceOf(account *types.Address) ([]byte, error) {
	data, err := balanceOfFn.EncodeCall(account)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeBalanceOf: %w", err)
	}
	return data, nil
}

// DecodeBalanceOf decodes the return data of balanceOf.
func DecodeBalanceOf(data []byte) (*big.Int, error) {
	n, err := balanceOfFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeBalanceOf: %w", err)
	}
	return n, nil
}

// EncodeAllowance returns the calldata for allowance(owner, spender).
func EncodeAllowance(owner, spender *types.Address) ([]byte, error) {
	data, err := allowanceFn.EncodeCall(owner, spender)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAllowance: %w", err)
	}
	return data, nil
}

// DecodeAllowance decodes the return data of allowance.
func DecodeAllowance(data []byte) (*big.Int, error) {
	n, err := allowanceFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeAllowance: %w", err)
	}
	return n, nil
}

// EncodeTransfer returns the calldata for transfer(to, value).
func EncodeTransfer(to *types.Address, value *big.Int) ([]byte, error) {
	data, err := transferFn.EncodeCall(to, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeTransfer: %w", err)
	}
	return data, nil
}

// DecodeTransfer decodes the return data of transfer.
func DecodeTransfer(data []byte) (bool, error) {
	ok, err := transferFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc20: decodeTransfer: %w", err)
	}
	return ok, nil
}

// EncodeApprove returns the calldata for approve(spender, value).
func EncodeApprove(spender *types.Address, value *big.Int) ([]byte, error) {
	data, err := approveFn.EncodeCall(spender, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeApprove: %w", err)
	}
	return data, nil
}

// DecodeApprove decodes the return data of approve.
func DecodeApprove(data []byte) (bool, error) {
	ok, err := approveFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc20: decodeApprove: %w", err)
	}
	return ok, nil
}

// EncodeTransferFrom returns the calldata for transferFrom(from, to, value).
func EncodeTransferFrom(from, to *types.Address, value *big.Int) ([]byte, error) {
	data, err := transferFromFn.EncodeCall(from, to, value)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeTransferFrom: %w", err)
	}
	return data, nil
}

// DecodeTransferFrom decodes the return data of transferFrom.
func DecodeTransferFrom(data []byte) (bool, error) {
	ok, err := transferFromFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc20: decodeTransferFrom: %w", err)
	}
	return ok, nil
}

// EncodeMetadata batches name/symbol/decimals/totalSupply on token into one aggregate3 call.
func EncodeMetadata(token *types.Address, allowFailure bool) ([]byte, error) {
	calls := []multicall3.Call3{
		multicall3.NewCall3(token, allowFailure, EncodeName()),
		multicall3.NewCall3(token, allowFailure, EncodeSymbol()),
		multicall3.NewCall3(token, allowFailure, EncodeDecimals()),
		multicall3.NewCall3(token, allowFailure, EncodeTotalSupply()),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeMetadata: %w", err)
	}
	return data, nil
}

// DecodeMetadata decodes the return data of EncodeMetadata; a failed call leaves its field at zero value.
func DecodeMetadata(data []byte) (*Metadata, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeMetadata: %w", err)
	}
	if len(results) != 4 {
		return nil, fmt.Errorf("erc20: decodeMetadata: expected 4 results, got %d", len(results))
	}

	m, err := decodeMetadata(results)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeMetadata: %w", err)
	}
	return m, nil
}

// EncodeMetadataWithBalance is EncodeMetadata plus balanceOf(account) in the same call.
func EncodeMetadataWithBalance(token, account *types.Address, allowFailure bool) ([]byte, error) {
	balCalldata, err := EncodeBalanceOf(account)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeMetadataWithBalance: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(token, allowFailure, EncodeName()),
		multicall3.NewCall3(token, allowFailure, EncodeSymbol()),
		multicall3.NewCall3(token, allowFailure, EncodeDecimals()),
		multicall3.NewCall3(token, allowFailure, EncodeTotalSupply()),
		multicall3.NewCall3(token, allowFailure, balCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeMetadataWithBalance: %w", err)
	}
	return data, nil
}

// DecodeMetadataWithBalance decodes the return data of EncodeMetadataWithBalance.
func DecodeMetadataWithBalance(data []byte) (*MetadataWithBalance, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeMetadataWithBalance: %w", err)
	}
	if len(results) != 5 {
		return nil, fmt.Errorf("erc20: decodeMetadataWithBalance: expected 5 results, got %d", len(results))
	}

	m, err := decodeMetadata(results[:4])
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeMetadataWithBalance: %w", err)
	}

	mb := MetadataWithBalance{
		Metadata: *m,
	}
	if results[4].Success {
		if mb.Balance, err = DecodeBalanceOf(results[4].ReturnData); err != nil {
			return nil, fmt.Errorf("erc20: decodeMetadataWithBalance: %w", err)
		}
	}
	return &mb, nil
}

// EncodeAllowanceWithBalance batches allowance(owner, spender) and both their balances into one call.
func EncodeAllowanceWithBalance(token, owner, spender *types.Address, allowFailure bool) ([]byte, error) {
	allowanceCalldata, err := EncodeAllowance(owner, spender)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAllowanceWithBalance: %w", err)
	}
	ownerBalCalldata, err := EncodeBalanceOf(owner)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAllowanceWithBalance: %w", err)
	}
	spenderBalCalldata, err := EncodeBalanceOf(spender)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAllowanceWithBalance: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(token, allowFailure, allowanceCalldata),
		multicall3.NewCall3(token, allowFailure, ownerBalCalldata),
		multicall3.NewCall3(token, allowFailure, spenderBalCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAllowanceWithBalance: %w", err)
	}
	return data, nil
}

// DecodeAllowanceWithBalance decodes the return data of EncodeAllowanceWithBalance.
func DecodeAllowanceWithBalance(data []byte) (*AllowanceWithBalance, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeAllowanceWithBalance: %w", err)
	}
	if len(results) != 3 {
		return nil, fmt.Errorf("erc20: decodeAllowanceWithBalance: expected 3 results, got %d", len(results))
	}

	var awb AllowanceWithBalance
	if results[0].Success {
		if awb.Value, err = DecodeAllowance(results[0].ReturnData); err != nil {
			return nil, fmt.Errorf("erc20: decodeAllowanceWithBalance: %w", err)
		}
	}
	if results[1].Success {
		if awb.OwnerBalance, err = DecodeBalanceOf(results[1].ReturnData); err != nil {
			return nil, fmt.Errorf("erc20: decodeAllowanceWithBalance: %w", err)
		}
	}
	if results[2].Success {
		if awb.SpenderBalance, err = DecodeBalanceOf(results[2].ReturnData); err != nil {
			return nil, fmt.Errorf("erc20: decodeAllowanceWithBalance: %w", err)
		}
	}
	return &awb, nil
}

// EncodeTokenBalances batches balanceOf(address) for each address on token into one call.
func EncodeTokenBalances(token *types.Address, addresses []*types.Address, allowFailure bool) ([]byte, error) {
	calls := make([]multicall3.Call3, len(addresses))
	for i, addr := range addresses {
		calldata, err := EncodeBalanceOf(addr)
		if err != nil {
			return nil, fmt.Errorf("erc20: encodeTokenBalances: addresses[%d] (%s): %w", i, addr, err)
		}
		calls[i] = multicall3.NewCall3(token, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeTokenBalances: %w", err)
	}
	return data, nil
}

// DecodeTokenBalances decodes the return data of EncodeTokenBalances, in the same order as addresses.
func DecodeTokenBalances(token *types.Address, addresses []*types.Address, data []byte) (*TokenBalances, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeTokenBalances: %w", err)
	}
	if len(results) != len(addresses) {
		return nil, fmt.Errorf("erc20: decodeTokenBalances: expected %d results, got %d", len(addresses), len(results))
	}

	balances := make([]AddressBalance, len(addresses))
	for i, addr := range addresses {
		balances[i].Address = addr
		if results[i].Success {
			if balances[i].Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc20: decodeTokenBalances: addresses[%d] (%s): %w", i, addr, err)
			}
		}
	}

	return NewTokenBalances(token, balances), nil
}

// EncodeAddressBalances batches balanceOf(address) across each token in tokens into one call.
func EncodeAddressBalances(address *types.Address, tokens []*types.Address, allowFailure bool) ([]byte, error) {
	calldata, err := EncodeBalanceOf(address)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAddressBalances: %w", err)
	}

	calls := make([]multicall3.Call3, len(tokens))
	for i, token := range tokens {
		calls[i] = multicall3.NewCall3(token, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeAddressBalances: %w", err)
	}
	return data, nil
}

// DecodeAddressBalances decodes the return data of EncodeAddressBalances, in the same order as tokens.
func DecodeAddressBalances(address *types.Address, tokens []*types.Address, data []byte) (*AddressBalances, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeAddressBalances: %w", err)
	}
	if len(results) != len(tokens) {
		return nil, fmt.Errorf("erc20: decodeAddressBalances: expected %d results, got %d", len(tokens), len(results))
	}

	balances := make([]TokenBalance, len(tokens))
	for i, token := range tokens {
		balances[i].Token = token
		if results[i].Success {
			if balances[i].Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc20: decodeAddressBalances: tokens[%d] (%s): %w", i, token, err)
			}
		}
	}

	return NewAddressBalances(address, balances), nil
}

// decodeMetadata decodes results[0:4] into a Metadata, zeroing failed fields.
func decodeMetadata(results []multicall3.Result) (*Metadata, error) {
	var (
		m   Metadata
		err error
	)
	if results[0].Success {
		if m.Name, err = DecodeName(results[0].ReturnData); err != nil {
			return nil, err
		}
	}
	if results[1].Success {
		if m.Symbol, err = DecodeSymbol(results[1].ReturnData); err != nil {
			return nil, err
		}
	}
	if results[2].Success {
		if m.Decimals, err = DecodeDecimals(results[2].ReturnData); err != nil {
			return nil, err
		}
	}
	if results[3].Success {
		if m.TotalSupply, err = DecodeTotalSupply(results[3].ReturnData); err != nil {
			return nil, err
		}
	}
	return &m, nil
}

// EncodeBalancePairs batches balanceOf(addresses[i]) on tokens[i] for each i into one call (paired positionally, not every combination).
func EncodeBalancePairs(tokens, addresses []*types.Address, allowFailure bool) ([]byte, error) {
	if len(tokens) != len(addresses) {
		return nil, fmt.Errorf("erc20: encodeBalancePairs: tokens and addresses must be the same length, got %d and %d", len(tokens), len(addresses))
	}

	calls := make([]multicall3.Call3, len(tokens))
	for i := range tokens {
		calldata, err := EncodeBalanceOf(addresses[i])
		if err != nil {
			return nil, fmt.Errorf("erc20: encodeBalancePairs: pairs[%d] (token %s, address %s): %w", i, tokens[i], addresses[i], err)
		}
		calls[i] = multicall3.NewCall3(tokens[i], allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc20: encodeBalancePairs: %w", err)
	}
	return data, nil
}

// DecodeBalancePairs decodes the return data of EncodeBalancePairs, in the same order and with the same tokens/addresses.
func DecodeBalancePairs(tokens, addresses []*types.Address, data []byte) ([]TokenAddressBalance, error) {
	if len(tokens) != len(addresses) {
		return nil, fmt.Errorf("erc20: decodeBalancePairs: tokens and addresses must be the same length, got %d and %d", len(tokens), len(addresses))
	}

	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc20: decodeBalancePairs: %w", err)
	}
	if len(results) != len(tokens) {
		return nil, fmt.Errorf("erc20: decodeBalancePairs: expected %d results, got %d", len(tokens), len(results))
	}

	balances := make([]TokenAddressBalance, len(tokens))
	for i := range tokens {
		tab := *NewTokenAddressBalance(tokens[i], addresses[i], nil)
		if results[i].Success {
			if tab.Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc20: decodeBalancePairs: pairs[%d] (token %s, address %s): %w", i, tokens[i], addresses[i], err)
			}
		}
		balances[i] = tab
	}
	return balances, nil
}
