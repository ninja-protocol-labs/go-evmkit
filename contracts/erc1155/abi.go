package erc1155

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IERC1155/IERC1155MetadataURI, for encoding calls and decoding results.
var (
	uriFn                   = abi.NewFunction("uri", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.String))
	balanceOfFn             = abi.NewFunction("balanceOf", abi.NewTypes(abi.Address, abi.Uint256), abi.NewTypes(abi.Uint256))
	balanceOfBatchFn        = abi.NewFunction("balanceOfBatch", abi.NewTypes(abi.Slice(abi.Address), abi.Slice(abi.Uint256)), abi.NewTypes(abi.Slice(abi.Uint256)))
	isApprovedForAllFn      = abi.NewFunction("isApprovedForAll", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Bool))
	setApprovalForAllFn     = abi.NewFunction("setApprovalForAll", abi.NewTypes(abi.Address, abi.Bool), nil)
	safeTransferFromFn      = abi.NewFunction("safeTransferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Bytes), nil)
	safeBatchTransferFromFn = abi.NewFunction("safeBatchTransferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Slice(abi.Uint256), abi.Slice(abi.Uint256), abi.Bytes), nil)
)

// Events from IERC1155, for decoding logs via Event.Decode.
var (
	transferSingleEvent = abi.NewEvent("TransferSingle", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)
	transferBatchEvent = abi.NewEvent("TransferBatch", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Slice(abi.Uint256)},
		{Type: abi.Slice(abi.Uint256)},
	}, false)
	approvalForAllEvent = abi.NewEvent("ApprovalForAll", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Bool},
	}, false)
	uriEvent = abi.NewEvent("URI", []abi.EventParam{
		{Type: abi.String},
		{Type: abi.Uint256, Indexed: true},
	}, false)
)

// Custom errors from IERC1155Errors (ERC-6093), for decoding reverts via abi.Error.Decode.
var (
	errInsufficientBalance   = abi.NewError("ERC1155InsufficientBalance", abi.NewTypes(abi.Address, abi.Uint256, abi.Uint256, abi.Uint256))
	errInvalidSender         = abi.NewError("ERC1155InvalidSender", abi.NewTypes(abi.Address))
	errInvalidReceiver       = abi.NewError("ERC1155InvalidReceiver", abi.NewTypes(abi.Address))
	errMissingApprovalForAll = abi.NewError("ERC1155MissingApprovalForAll", abi.NewTypes(abi.Address, abi.Address))
	errInvalidApprover       = abi.NewError("ERC1155InvalidApprover", abi.NewTypes(abi.Address))
	errInvalidOperator       = abi.NewError("ERC1155InvalidOperator", abi.NewTypes(abi.Address))
	errInvalidArrayLength    = abi.NewError("ERC1155InvalidArrayLength", abi.NewTypes(abi.Uint256, abi.Uint256))
)

// ErrInsufficientBalance, etc. mirror IERC1155Errors' custom errors, for errors.Is.
var (
	ErrInsufficientBalance   = errors.New("erc1155: insufficient balance")
	ErrInvalidSender         = errors.New("erc1155: invalid sender")
	ErrInvalidReceiver       = errors.New("erc1155: invalid receiver")
	ErrMissingApprovalForAll = errors.New("erc1155: missing approval for all")
	ErrInvalidApprover       = errors.New("erc1155: invalid approver")
	ErrInvalidOperator       = errors.New("erc1155: invalid operator")
	ErrInvalidArrayLength    = errors.New("erc1155: invalid array length")
	ErrUnknown               = errors.New("erc1155: unknown revert")
)

// EncodeBalanceOf returns the calldata for balanceOf(account, id).
func EncodeBalanceOf(account *types.Address, id *big.Int) ([]byte, error) {
	data, err := balanceOfFn.EncodeCall(account, id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeBalanceOf: %w", err)
	}
	return data, nil
}

// DecodeBalanceOf decodes the return data of balanceOf.
func DecodeBalanceOf(data []byte) (*big.Int, error) {
	n, err := balanceOfFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeBalanceOf: %w", err)
	}
	return n, nil
}

// EncodeBalanceOfBatch returns the calldata for balanceOfBatch(accounts, ids). accounts and ids must
// be the same length: each balanceOf(accounts[i], ids[i]) is returned at result[i].
func EncodeBalanceOfBatch(accounts []*types.Address, ids []*big.Int) ([]byte, error) {
	if len(accounts) != len(ids) {
		return nil, fmt.Errorf("erc1155: encodeBalanceOfBatch: accounts and ids must be the same length, got %d and %d", len(accounts), len(ids))
	}

	data, err := balanceOfBatchFn.EncodeCall(accounts, ids)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeBalanceOfBatch: %w", err)
	}
	return data, nil
}

// DecodeBalanceOfBatch decodes the return data of balanceOfBatch, in the same order as the accounts/ids it was called with.
func DecodeBalanceOfBatch(data []byte) ([]*big.Int, error) {
	vals, err := balanceOfBatchFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeBalanceOfBatch: %w", err)
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("erc1155: decodeBalanceOfBatch: no return values")
	}

	elems, ok := vals[0].([]any)
	if !ok {
		return nil, fmt.Errorf("erc1155: decodeBalanceOfBatch: unexpected type %T", vals[0])
	}

	amounts := make([]*big.Int, len(elems))
	for i, e := range elems {
		n, ok := e.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("erc1155: decodeBalanceOfBatch: amounts[%d]: unexpected type %T", i, e)
		}
		amounts[i] = n
	}
	return amounts, nil
}

// DecodeIdBalances decodes the return data of a balanceOfBatch call made with a single account
// repeated once per id, pairing each result with its id. ids must be the same ids (in the same
// order) the call was made with.
func DecodeIdBalances(ids []*big.Int, data []byte) ([]IdBalance, error) {
	amounts, err := DecodeBalanceOfBatch(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeIdBalances: %w", err)
	}
	if len(amounts) != len(ids) {
		return nil, fmt.Errorf("erc1155: decodeIdBalances: expected %d results, got %d", len(ids), len(amounts))
	}

	balances := make([]IdBalance, len(ids))
	for i, id := range ids {
		balances[i] = *NewIdBalance(id, amounts[i])
	}
	return balances, nil
}

// DecodeIdBalancePairs decodes the return data of a balanceOfBatch(accounts, ids) call, pairing
// each result with the (account, id) pair that produced it. accounts and ids must be the same
// accounts/ids (in the same order) the call was made with.
func DecodeIdBalancePairs(accounts []*types.Address, ids []*big.Int, data []byte) ([]IdBalancePair, error) {
	if len(accounts) != len(ids) {
		return nil, fmt.Errorf("erc1155: decodeIdBalancePairs: accounts and ids must be the same length, got %d and %d", len(accounts), len(ids))
	}

	amounts, err := DecodeBalanceOfBatch(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeIdBalancePairs: %w", err)
	}
	if len(amounts) != len(accounts) {
		return nil, fmt.Errorf("erc1155: decodeIdBalancePairs: expected %d results, got %d", len(accounts), len(amounts))
	}

	pairs := make([]IdBalancePair, len(accounts))
	for i := range accounts {
		pairs[i] = *NewIdBalancePair(accounts[i], ids[i], amounts[i])
	}
	return pairs, nil
}

// EncodeIsApprovedForAll returns the calldata for isApprovedForAll(account, operator).
func EncodeIsApprovedForAll(account, operator *types.Address) ([]byte, error) {
	data, err := isApprovedForAllFn.EncodeCall(account, operator)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeIsApprovedForAll: %w", err)
	}
	return data, nil
}

// DecodeIsApprovedForAll decodes the return data of isApprovedForAll.
func DecodeIsApprovedForAll(data []byte) (bool, error) {
	ok, err := isApprovedForAllFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc1155: decodeIsApprovedForAll: %w", err)
	}
	return ok, nil
}

// EncodeSetApprovalForAll returns the calldata for setApprovalForAll(operator, approved).
func EncodeSetApprovalForAll(operator *types.Address, approved bool) ([]byte, error) {
	data, err := setApprovalForAllFn.EncodeCall(operator, approved)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeSetApprovalForAll: %w", err)
	}
	return data, nil
}

// EncodeSafeTransferFrom returns the calldata for safeTransferFrom(from, to, id, value, data).
func EncodeSafeTransferFrom(from, to *types.Address, id, value *big.Int, callData []byte) ([]byte, error) {
	data, err := safeTransferFromFn.EncodeCall(from, to, id, value, callData)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeSafeTransferFrom: %w", err)
	}
	return data, nil
}

// EncodeSafeBatchTransferFrom returns the calldata for safeBatchTransferFrom(from, to, ids, values, data).
// ids and values must be the same length.
func EncodeSafeBatchTransferFrom(from, to *types.Address, ids, values []*big.Int, callData []byte) ([]byte, error) {
	if len(ids) != len(values) {
		return nil, fmt.Errorf("erc1155: encodeSafeBatchTransferFrom: ids and values must be the same length, got %d and %d", len(ids), len(values))
	}

	data, err := safeBatchTransferFromFn.EncodeCall(from, to, ids, values, callData)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeSafeBatchTransferFrom: %w", err)
	}
	return data, nil
}

// EncodeURI returns the calldata for uri(id).
func EncodeURI(id *big.Int) ([]byte, error) {
	data, err := uriFn.EncodeCall(id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeURI: %w", err)
	}
	return data, nil
}

// DecodeURI decodes the return data of uri.
func DecodeURI(data []byte) (string, error) {
	s, err := uriFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc1155: decodeURI: %w", err)
	}
	return s, nil
}

// EncodeApprovalForAllWithBalance batches isApprovedForAll(account, operator) and both their
// balanceOf(id) results into one call.
func EncodeApprovalForAllWithBalance(contract, account, operator *types.Address, id *big.Int, allowFailure bool) ([]byte, error) {
	approvedCalldata, err := EncodeIsApprovedForAll(account, operator)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeApprovalForAllWithBalance: %w", err)
	}
	accountBalCalldata, err := EncodeBalanceOf(account, id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeApprovalForAllWithBalance: %w", err)
	}
	operatorBalCalldata, err := EncodeBalanceOf(operator, id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeApprovalForAllWithBalance: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(contract, allowFailure, approvedCalldata),
		multicall3.NewCall3(contract, allowFailure, accountBalCalldata),
		multicall3.NewCall3(contract, allowFailure, operatorBalCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeApprovalForAllWithBalance: %w", err)
	}
	return data, nil
}

// DecodeApprovalForAllWithBalance decodes the return data of EncodeApprovalForAllWithBalance.
// account, operator and id must be the same values passed to EncodeApprovalForAllWithBalance,
// since none of them is itself returned on-chain.
func DecodeApprovalForAllWithBalance(account, operator *types.Address, id *big.Int, data []byte) (*ApprovalForAllWithBalance, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeApprovalForAllWithBalance: %w", err)
	}
	if len(results) != 3 {
		return nil, fmt.Errorf("erc1155: decodeApprovalForAllWithBalance: expected 3 results, got %d", len(results))
	}

	afawb := ApprovalForAllWithBalance{Account: account, Operator: operator, Id: id}
	if results[0].Success {
		if afawb.Approved, err = DecodeIsApprovedForAll(results[0].ReturnData); err != nil {
			return nil, fmt.Errorf("erc1155: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	if results[1].Success {
		if afawb.AccountBalance, err = DecodeBalanceOf(results[1].ReturnData); err != nil {
			return nil, fmt.Errorf("erc1155: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	if results[2].Success {
		if afawb.OperatorBalance, err = DecodeBalanceOf(results[2].ReturnData); err != nil {
			return nil, fmt.Errorf("erc1155: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	return &afawb, nil
}

// EncodeAccountIdBalances batches balanceOf(account, id) across each contract in contracts into one call.
func EncodeAccountIdBalances(account *types.Address, id *big.Int, contracts []*types.Address, allowFailure bool) ([]byte, error) {
	calldata, err := EncodeBalanceOf(account, id)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeAccountIdBalances: %w", err)
	}

	calls := make([]multicall3.Call3, len(contracts))
	for i, contract := range contracts {
		calls[i] = multicall3.NewCall3(contract, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc1155: encodeAccountIdBalances: %w", err)
	}
	return data, nil
}

// DecodeAccountIdBalances decodes the return data of EncodeAccountIdBalances, in the same order as contracts.
func DecodeAccountIdBalances(account *types.Address, id *big.Int, contracts []*types.Address, data []byte) (*AccountIdBalances, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc1155: decodeAccountIdBalances: %w", err)
	}
	if len(results) != len(contracts) {
		return nil, fmt.Errorf("erc1155: decodeAccountIdBalances: expected %d results, got %d", len(contracts), len(results))
	}

	balances := make([]ContractBalance, len(contracts))
	for i, contract := range contracts {
		balances[i].Contract = contract
		if results[i].Success {
			if balances[i].Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc1155: decodeAccountIdBalances: contracts[%d] (%s): %w", i, contract, err)
			}
		}
	}

	return NewAccountIdBalances(account, id, balances), nil
}
