// Package erc721 encodes calls to and decodes results from an ERC-721 token
// contract. Go types mirror the interfaces in solidity/interfaces/IERC721.sol
// and solidity/interfaces/IERC721Metadata.sol.
package erc721

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IERC721/IERC721Metadata, for encoding calls and decoding results.
var (
	nameFn                     = abi.NewFunction("name", nil, abi.NewTypes(abi.String))
	symbolFn                   = abi.NewFunction("symbol", nil, abi.NewTypes(abi.String))
	balanceOfFn                = abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	ownerOfFn                  = abi.NewFunction("ownerOf", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Address))
	getApprovedFn              = abi.NewFunction("getApproved", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.Address))
	isApprovedForAllFn         = abi.NewFunction("isApprovedForAll", abi.NewTypes(abi.Address, abi.Address), abi.NewTypes(abi.Bool))
	approveFn                  = abi.NewFunction("approve", abi.NewTypes(abi.Address, abi.Uint256), nil)
	setApprovalForAllFn        = abi.NewFunction("setApprovalForAll", abi.NewTypes(abi.Address, abi.Bool), nil)
	transferFromFn             = abi.NewFunction("transferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), nil)
	safeTransferFromFn         = abi.NewFunction("safeTransferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256), nil)
	safeTransferFromWithDataFn = abi.NewFunction("safeTransferFrom", abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Bytes), nil)
	tokenURIFn                 = abi.NewFunction("tokenURI", abi.NewTypes(abi.Uint256), abi.NewTypes(abi.String))
)

// Events from IERC721, for decoding logs via Event.Decode.
var (
	transferEvent = abi.NewEvent("Transfer", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256, Indexed: true},
	}, false)
	approvalEvent = abi.NewEvent("Approval", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256, Indexed: true},
	}, false)
	approvalForAllEvent = abi.NewEvent("ApprovalForAll", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Bool},
	}, false)
)

// Custom errors from IERC721Errors (ERC-6093), for decoding reverts via abi.Error.Decode.
var (
	errInvalidOwner         = abi.NewError("ERC721InvalidOwner", abi.NewTypes(abi.Address))
	errNonexistentToken     = abi.NewError("ERC721NonexistentToken", abi.NewTypes(abi.Uint256))
	errIncorrectOwner       = abi.NewError("ERC721IncorrectOwner", abi.NewTypes(abi.Address, abi.Uint256, abi.Address))
	errInvalidSender        = abi.NewError("ERC721InvalidSender", abi.NewTypes(abi.Address))
	errInvalidReceiver      = abi.NewError("ERC721InvalidReceiver", abi.NewTypes(abi.Address))
	errInsufficientApproval = abi.NewError("ERC721InsufficientApproval", abi.NewTypes(abi.Address, abi.Uint256))
	errInvalidApprover      = abi.NewError("ERC721InvalidApprover", abi.NewTypes(abi.Address))
	errInvalidOperator      = abi.NewError("ERC721InvalidOperator", abi.NewTypes(abi.Address))
)

// ErrInvalidOwner, etc. mirror IERC721Errors' custom errors, for errors.Is.
var (
	ErrInvalidOwner         = errors.New("erc721: invalid owner")
	ErrNonexistentToken     = errors.New("erc721: nonexistent token")
	ErrIncorrectOwner       = errors.New("erc721: incorrect owner")
	ErrInvalidSender        = errors.New("erc721: invalid sender")
	ErrInvalidReceiver      = errors.New("erc721: invalid receiver")
	ErrInsufficientApproval = errors.New("erc721: insufficient approval")
	ErrInvalidApprover      = errors.New("erc721: invalid approver")
	ErrInvalidOperator      = errors.New("erc721: invalid operator")
	ErrUnknown              = errors.New("erc721: unknown revert")
)

// EncodeName returns the calldata for name().
func EncodeName() []byte {
	return nameFn.SelectorBytes()
}

// DecodeName decodes the return data of name.
func DecodeName(data []byte) (string, error) {
	s, err := nameFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc721: decodeName: %w", err)
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
		return "", fmt.Errorf("erc721: decodeSymbol: %w", err)
	}
	return s, nil
}

// EncodeBalanceOf returns the calldata for balanceOf(owner).
func EncodeBalanceOf(owner *types.Address) ([]byte, error) {
	data, err := balanceOfFn.EncodeCall(owner)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeBalanceOf: %w", err)
	}
	return data, nil
}

// DecodeBalanceOf decodes the return data of balanceOf.
func DecodeBalanceOf(data []byte) (*big.Int, error) {
	n, err := balanceOfFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeBalanceOf: %w", err)
	}
	return n, nil
}

// EncodeOwnerOf returns the calldata for ownerOf(tokenId).
func EncodeOwnerOf(tokenId *big.Int) ([]byte, error) {
	data, err := ownerOfFn.EncodeCall(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeOwnerOf: %w", err)
	}
	return data, nil
}

// DecodeOwnerOf decodes the return data of ownerOf.
func DecodeOwnerOf(data []byte) (*types.Address, error) {
	addr, err := ownerOfFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeOwnerOf: %w", err)
	}
	return addr, nil
}

// EncodeGetApproved returns the calldata for getApproved(tokenId).
func EncodeGetApproved(tokenId *big.Int) ([]byte, error) {
	data, err := getApprovedFn.EncodeCall(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeGetApproved: %w", err)
	}
	return data, nil
}

// DecodeGetApproved decodes the return data of getApproved.
func DecodeGetApproved(data []byte) (*types.Address, error) {
	addr, err := getApprovedFn.DecodeSingleReturn[*types.Address](data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeGetApproved: %w", err)
	}
	return addr, nil
}

// EncodeIsApprovedForAll returns the calldata for isApprovedForAll(owner, operator).
func EncodeIsApprovedForAll(owner, operator *types.Address) ([]byte, error) {
	data, err := isApprovedForAllFn.EncodeCall(owner, operator)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeIsApprovedForAll: %w", err)
	}
	return data, nil
}

// DecodeIsApprovedForAll decodes the return data of isApprovedForAll.
func DecodeIsApprovedForAll(data []byte) (bool, error) {
	ok, err := isApprovedForAllFn.DecodeSingleReturn[bool](data)
	if err != nil {
		return false, fmt.Errorf("erc721: decodeIsApprovedForAll: %w", err)
	}
	return ok, nil
}

// EncodeApprove returns the calldata for approve(to, tokenId).
func EncodeApprove(to *types.Address, tokenId *big.Int) ([]byte, error) {
	data, err := approveFn.EncodeCall(to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeApprove: %w", err)
	}
	return data, nil
}

// EncodeSetApprovalForAll returns the calldata for setApprovalForAll(operator, approved).
func EncodeSetApprovalForAll(operator *types.Address, approved bool) ([]byte, error) {
	data, err := setApprovalForAllFn.EncodeCall(operator, approved)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeSetApprovalForAll: %w", err)
	}
	return data, nil
}

// EncodeTransferFrom returns the calldata for transferFrom(from, to, tokenId).
func EncodeTransferFrom(from, to *types.Address, tokenId *big.Int) ([]byte, error) {
	data, err := transferFromFn.EncodeCall(from, to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeTransferFrom: %w", err)
	}
	return data, nil
}

// EncodeSafeTransferFrom returns the calldata for safeTransferFrom(from, to, tokenId).
func EncodeSafeTransferFrom(from, to *types.Address, tokenId *big.Int) ([]byte, error) {
	data, err := safeTransferFromFn.EncodeCall(from, to, tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeSafeTransferFrom: %w", err)
	}
	return data, nil
}

// EncodeSafeTransferFromWithData returns the calldata for safeTransferFrom(from, to, tokenId, data).
func EncodeSafeTransferFromWithData(from, to *types.Address, tokenId *big.Int, callData []byte) ([]byte, error) {
	encoded, err := safeTransferFromWithDataFn.EncodeCall(from, to, tokenId, callData)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeSafeTransferFromWithData: %w", err)
	}
	return encoded, nil
}

// EncodeTokenURI returns the calldata for tokenURI(tokenId).
func EncodeTokenURI(tokenId *big.Int) ([]byte, error) {
	data, err := tokenURIFn.EncodeCall(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeTokenURI: %w", err)
	}
	return data, nil
}

// DecodeTokenURI decodes the return data of tokenURI.
func DecodeTokenURI(data []byte) (string, error) {
	s, err := tokenURIFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("erc721: decodeTokenURI: %w", err)
	}
	return s, nil
}

// EncodeMetadata batches name/symbol on contract into one aggregate3 call.
func EncodeMetadata(contract *types.Address, allowFailure bool) ([]byte, error) {
	calls := []multicall3.Call3{
		multicall3.NewCall3(contract, allowFailure, EncodeName()),
		multicall3.NewCall3(contract, allowFailure, EncodeSymbol()),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeMetadata: %w", err)
	}
	return data, nil
}

// DecodeMetadata decodes the return data of EncodeMetadata; a failed call leaves its field at zero value.
func DecodeMetadata(data []byte) (*Metadata, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeMetadata: %w", err)
	}
	if len(results) != 2 {
		return nil, fmt.Errorf("erc721: decodeMetadata: expected 2 results, got %d", len(results))
	}

	m, err := decodeMetadata(results)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeMetadata: %w", err)
	}
	return m, nil
}

// decodeMetadata decodes results[0:2] into a Metadata, zeroing failed fields.
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
	return &m, nil
}

// EncodeMetadataWithTokenID is EncodeMetadata plus tokenURI(tokenId) and ownerOf(tokenId) in the same call.
func EncodeMetadataWithTokenID(contract *types.Address, tokenId *big.Int, allowFailure bool) ([]byte, error) {
	uriCalldata, err := EncodeTokenURI(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeMetadataWithTokenID: %w", err)
	}
	ownerCalldata, err := EncodeOwnerOf(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeMetadataWithTokenID: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(contract, allowFailure, EncodeName()),
		multicall3.NewCall3(contract, allowFailure, EncodeSymbol()),
		multicall3.NewCall3(contract, allowFailure, uriCalldata),
		multicall3.NewCall3(contract, allowFailure, ownerCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeMetadataWithTokenID: %w", err)
	}
	return data, nil
}

// DecodeMetadataWithTokenID decodes the return data of EncodeMetadataWithTokenID. tokenId must be
// the same value passed to EncodeMetadataWithTokenID, since it isn't itself returned on-chain.
func DecodeMetadataWithTokenID(tokenId *big.Int, data []byte) (*MetadataWithTokenID, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeMetadataWithTokenID: %w", err)
	}
	if len(results) != 4 {
		return nil, fmt.Errorf("erc721: decodeMetadataWithTokenID: expected 4 results, got %d", len(results))
	}

	m, err := decodeMetadata(results[:2])
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeMetadataWithTokenID: %w", err)
	}

	mt := MetadataWithTokenID{
		Metadata: *m,
		TokenId:  tokenId,
	}
	if results[2].Success {
		if mt.TokenURI, err = DecodeTokenURI(results[2].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeMetadataWithTokenID: %w", err)
		}
	}
	if results[3].Success {
		if mt.Owner, err = DecodeOwnerOf(results[3].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeMetadataWithTokenID: %w", err)
		}
	}
	return &mt, nil
}

// EncodeAllowanceWithTokenID batches ownerOf(tokenId) and getApproved(tokenId) into one call.
func EncodeAllowanceWithTokenID(contract *types.Address, tokenId *big.Int, allowFailure bool) ([]byte, error) {
	ownerCalldata, err := EncodeOwnerOf(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeAllowanceWithTokenID: %w", err)
	}
	approvedCalldata, err := EncodeGetApproved(tokenId)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeAllowanceWithTokenID: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(contract, allowFailure, ownerCalldata),
		multicall3.NewCall3(contract, allowFailure, approvedCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeAllowanceWithTokenID: %w", err)
	}
	return data, nil
}

// DecodeAllowanceWithTokenID decodes the return data of EncodeAllowanceWithTokenID. tokenId must be
// the same value passed to EncodeAllowanceWithTokenID, since it isn't itself returned on-chain.
func DecodeAllowanceWithTokenID(tokenId *big.Int, data []byte) (*AllowanceWithTokenID, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeAllowanceWithTokenID: %w", err)
	}
	if len(results) != 2 {
		return nil, fmt.Errorf("erc721: decodeAllowanceWithTokenID: expected 2 results, got %d", len(results))
	}

	awt := AllowanceWithTokenID{TokenId: tokenId}
	if results[0].Success {
		if awt.Owner, err = DecodeOwnerOf(results[0].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeAllowanceWithTokenID: %w", err)
		}
	}
	if results[1].Success {
		if awt.Operator, err = DecodeGetApproved(results[1].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeAllowanceWithTokenID: %w", err)
		}
	}
	return &awt, nil
}

// EncodeApprovalForAllWithBalance batches isApprovedForAll(owner, operator) and both their balances into one call.
func EncodeApprovalForAllWithBalance(contract, owner, operator *types.Address, allowFailure bool) ([]byte, error) {
	approvedCalldata, err := EncodeIsApprovedForAll(owner, operator)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeApprovalForAllWithBalance: %w", err)
	}
	ownerBalCalldata, err := EncodeBalanceOf(owner)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeApprovalForAllWithBalance: %w", err)
	}
	operatorBalCalldata, err := EncodeBalanceOf(operator)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeApprovalForAllWithBalance: %w", err)
	}

	calls := []multicall3.Call3{
		multicall3.NewCall3(contract, allowFailure, approvedCalldata),
		multicall3.NewCall3(contract, allowFailure, ownerBalCalldata),
		multicall3.NewCall3(contract, allowFailure, operatorBalCalldata),
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeApprovalForAllWithBalance: %w", err)
	}
	return data, nil
}

// DecodeApprovalForAllWithBalance decodes the return data of EncodeApprovalForAllWithBalance. owner
// and operator must be the same values passed to EncodeApprovalForAllWithBalance, since neither is
// itself returned on-chain.
func DecodeApprovalForAllWithBalance(owner, operator *types.Address, data []byte) (*ApprovalForAllWithBalance, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeApprovalForAllWithBalance: %w", err)
	}
	if len(results) != 3 {
		return nil, fmt.Errorf("erc721: decodeApprovalForAllWithBalance: expected 3 results, got %d", len(results))
	}

	afawb := ApprovalForAllWithBalance{Owner: owner, Operator: operator}
	if results[0].Success {
		if afawb.Approved, err = DecodeIsApprovedForAll(results[0].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	if results[1].Success {
		if afawb.OwnerBalance, err = DecodeBalanceOf(results[1].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	if results[2].Success {
		if afawb.OperatorBalance, err = DecodeBalanceOf(results[2].ReturnData); err != nil {
			return nil, fmt.Errorf("erc721: decodeApprovalForAllWithBalance: %w", err)
		}
	}
	return &afawb, nil
}

// EncodeTokenOwners batches ownerOf(tokenIds[i]) on contract for each i into one call.
func EncodeTokenOwners(contract *types.Address, tokenIds []*big.Int, allowFailure bool) ([]byte, error) {
	calls := make([]multicall3.Call3, len(tokenIds))
	for i, tokenId := range tokenIds {
		calldata, err := EncodeOwnerOf(tokenId)
		if err != nil {
			return nil, fmt.Errorf("erc721: encodeTokenOwners: tokenIds[%d] (%s): %w", i, tokenId, err)
		}
		calls[i] = multicall3.NewCall3(contract, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeTokenOwners: %w", err)
	}
	return data, nil
}

// DecodeTokenOwners decodes the return data of EncodeTokenOwners, in the same order as tokenIds.
func DecodeTokenOwners(contract *types.Address, tokenIds []*big.Int, data []byte) (*TokenOwners, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeTokenOwners: %w", err)
	}
	if len(results) != len(tokenIds) {
		return nil, fmt.Errorf("erc721: decodeTokenOwners: expected %d results, got %d", len(tokenIds), len(results))
	}

	owners := make([]TokenOwner, len(tokenIds))
	for i, tokenId := range tokenIds {
		owners[i].TokenId = tokenId
		if results[i].Success {
			if owners[i].Owner, err = DecodeOwnerOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc721: decodeTokenOwners: tokenIds[%d] (%s): %w", i, tokenId, err)
			}
		}
	}

	return NewTokenOwners(contract, owners), nil
}

// EncodeTokenOperators batches getApproved(tokenIds[i]) on contract for each i into one call.
func EncodeTokenOperators(contract *types.Address, tokenIds []*big.Int, allowFailure bool) ([]byte, error) {
	calls := make([]multicall3.Call3, len(tokenIds))
	for i, tokenId := range tokenIds {
		calldata, err := EncodeGetApproved(tokenId)
		if err != nil {
			return nil, fmt.Errorf("erc721: encodeTokenOperators: tokenIds[%d] (%s): %w", i, tokenId, err)
		}
		calls[i] = multicall3.NewCall3(contract, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeTokenOperators: %w", err)
	}
	return data, nil
}

// DecodeTokenOperators decodes the return data of EncodeTokenOperators, in the same order as tokenIds.
func DecodeTokenOperators(contract *types.Address, tokenIds []*big.Int, data []byte) (*TokenOperators, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeTokenOperators: %w", err)
	}
	if len(results) != len(tokenIds) {
		return nil, fmt.Errorf("erc721: decodeTokenOperators: expected %d results, got %d", len(tokenIds), len(results))
	}

	operators := make([]TokenOperator, len(tokenIds))
	for i, tokenId := range tokenIds {
		operators[i].TokenId = tokenId
		if results[i].Success {
			if operators[i].Operator, err = DecodeGetApproved(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc721: decodeTokenOperators: tokenIds[%d] (%s): %w", i, tokenId, err)
			}
		}
	}

	return NewTokenOperators(contract, operators), nil
}

// EncodeAddressBalances batches balanceOf(address) across each contract in contracts into one call.
func EncodeAddressBalances(address *types.Address, contracts []*types.Address, allowFailure bool) ([]byte, error) {
	calldata, err := EncodeBalanceOf(address)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeAddressBalances: %w", err)
	}

	calls := make([]multicall3.Call3, len(contracts))
	for i, contract := range contracts {
		calls[i] = multicall3.NewCall3(contract, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeAddressBalances: %w", err)
	}
	return data, nil
}

// DecodeAddressBalances decodes the return data of EncodeAddressBalances, in the same order as contracts.
func DecodeAddressBalances(address *types.Address, contracts []*types.Address, data []byte) (*AddressBalances, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeAddressBalances: %w", err)
	}
	if len(results) != len(contracts) {
		return nil, fmt.Errorf("erc721: decodeAddressBalances: expected %d results, got %d", len(contracts), len(results))
	}

	balances := make([]TokenBalance, len(contracts))
	for i, contract := range contracts {
		balances[i].Token = contract
		if results[i].Success {
			if balances[i].Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc721: decodeAddressBalances: contracts[%d] (%s): %w", i, contract, err)
			}
		}
	}

	return NewAddressBalances(address, balances), nil
}

// EncodeTokenBalances batches balanceOf(addresses[i]) on contract for each i into one call.
func EncodeTokenBalances(contract *types.Address, addresses []*types.Address, allowFailure bool) ([]byte, error) {
	calls := make([]multicall3.Call3, len(addresses))
	for i, addr := range addresses {
		calldata, err := EncodeBalanceOf(addr)
		if err != nil {
			return nil, fmt.Errorf("erc721: encodeTokenBalances: addresses[%d] (%s): %w", i, addr, err)
		}
		calls[i] = multicall3.NewCall3(contract, allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeTokenBalances: %w", err)
	}
	return data, nil
}

// DecodeTokenBalances decodes the return data of EncodeTokenBalances, in the same order as addresses.
func DecodeTokenBalances(contract *types.Address, addresses []*types.Address, data []byte) (*TokenBalances, error) {
	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeTokenBalances: %w", err)
	}
	if len(results) != len(addresses) {
		return nil, fmt.Errorf("erc721: decodeTokenBalances: expected %d results, got %d", len(addresses), len(results))
	}

	balances := make([]AddressBalance, len(addresses))
	for i, addr := range addresses {
		balances[i].Address = addr
		if results[i].Success {
			if balances[i].Amount, err = DecodeBalanceOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc721: decodeTokenBalances: addresses[%d] (%s): %w", i, addr, err)
			}
		}
	}

	return NewTokenBalances(contract, balances), nil
}

// EncodeOwnerPairs batches ownerOf(tokenIds[i]) on contracts[i] for each i into one call (paired positionally, not every combination).
func EncodeOwnerPairs(contracts []*types.Address, tokenIds []*big.Int, allowFailure bool) ([]byte, error) {
	if len(contracts) != len(tokenIds) {
		return nil, fmt.Errorf("erc721: encodeOwnerPairs: contracts and tokenIds must be the same length, got %d and %d", len(contracts), len(tokenIds))
	}

	calls := make([]multicall3.Call3, len(contracts))
	for i := range contracts {
		calldata, err := EncodeOwnerOf(tokenIds[i])
		if err != nil {
			return nil, fmt.Errorf("erc721: encodeOwnerPairs: pairs[%d] (contract %s, tokenId %s): %w", i, contracts[i], tokenIds[i], err)
		}
		calls[i] = multicall3.NewCall3(contracts[i], allowFailure, calldata)
	}

	data, err := multicall3.EncodeAggregate3(calls)
	if err != nil {
		return nil, fmt.Errorf("erc721: encodeOwnerPairs: %w", err)
	}
	return data, nil
}

// DecodeOwnerPairs decodes the return data of EncodeOwnerPairs, in the same order and with the same contracts/tokenIds.
func DecodeOwnerPairs(contracts []*types.Address, tokenIds []*big.Int, data []byte) ([]TokenOwnerPair, error) {
	if len(contracts) != len(tokenIds) {
		return nil, fmt.Errorf("erc721: decodeOwnerPairs: contracts and tokenIds must be the same length, got %d and %d", len(contracts), len(tokenIds))
	}

	results, err := multicall3.DecodeResults(data)
	if err != nil {
		return nil, fmt.Errorf("erc721: decodeOwnerPairs: %w", err)
	}
	if len(results) != len(contracts) {
		return nil, fmt.Errorf("erc721: decodeOwnerPairs: expected %d results, got %d", len(contracts), len(results))
	}

	owners := make([]TokenOwnerPair, len(contracts))
	for i := range contracts {
		owners[i] = *NewTokenOwnerPair(contracts[i], nil, tokenIds[i])
		if results[i].Success {
			if owners[i].Owner, err = DecodeOwnerOf(results[i].ReturnData); err != nil {
				return nil, fmt.Errorf("erc721: decodeOwnerPairs: pairs[%d] (contract %s, tokenId %s): %w", i, contracts[i], tokenIds[i], err)
			}
		}
	}

	return owners, nil
}
