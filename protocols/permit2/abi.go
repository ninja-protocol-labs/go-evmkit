package permit2

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Tuple types for IAllowanceTransfer's structs.
var (
	permitDetailsT           = abi.Tuple(abi.Address, abi.Uint160, abi.Uint48, abi.Uint48)
	permitSingleT            = abi.Tuple(permitDetailsT, abi.Address, abi.Uint256)
	permitBatchT             = abi.Tuple(abi.Slice(permitDetailsT), abi.Address, abi.Uint256)
	tokenSpenderPairT        = abi.Tuple(abi.Address, abi.Address)
	allowanceTransferDetailT = abi.Tuple(abi.Address, abi.Address, abi.Uint160, abi.Address)
)

// Tuple types for ISignatureTransfer's structs.
var (
	tokenPermissionsT        = abi.Tuple(abi.Address, abi.Uint256)
	permitTransferFromT      = abi.Tuple(tokenPermissionsT, abi.Uint256, abi.Uint256)
	permitBatchTransferFromT = abi.Tuple(abi.Slice(tokenPermissionsT), abi.Uint256, abi.Uint256)
	signatureTransferDetailT = abi.Tuple(abi.Address, abi.Uint256)
)

// Function signatures from IEIP712/IAllowanceTransfer.
var (
	domainSeparatorFn = abi.NewFunction(
		"DOMAIN_SEPARATOR",
		nil,
		abi.NewTypes(abi.Bytes32),
	)
	allowanceFn = abi.NewFunction(
		"allowance",
		abi.NewTypes(abi.Address, abi.Address, abi.Address),
		abi.NewTypes(abi.Uint160, abi.Uint48, abi.Uint48),
	)
	approveFn = abi.NewFunction(
		"approve",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint160, abi.Uint48),
		nil,
	)
	permitSingleFn = abi.NewFunction(
		"permit",
		abi.NewTypes(abi.Address, permitSingleT, abi.Bytes),
		nil,
	)
	permitBatchFn = abi.NewFunction(
		"permit",
		abi.NewTypes(abi.Address, permitBatchT, abi.Bytes),
		nil,
	)
	transferFromFn = abi.NewFunction(
		"transferFrom",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint160, abi.Address),
		nil,
	)
	transferFromBatchFn = abi.NewFunction(
		"transferFrom",
		abi.NewTypes(abi.Slice(allowanceTransferDetailT)),
		nil,
	)
	lockdownFn = abi.NewFunction(
		"lockdown",
		abi.NewTypes(abi.Slice(tokenSpenderPairT)),
		nil,
	)
	invalidateNoncesFn = abi.NewFunction(
		"invalidateNonces",
		abi.NewTypes(abi.Address, abi.Address, abi.Uint48),
		nil,
	)
)

// Function signatures from ISignatureTransfer.
var (
	nonceBitmapFn = abi.NewFunction(
		"nonceBitmap",
		abi.NewTypes(abi.Address, abi.Uint256),
		abi.NewTypes(abi.Uint256),
	)
	permitTransferFromFn = abi.NewFunction(
		"permitTransferFrom",
		abi.NewTypes(permitTransferFromT, signatureTransferDetailT, abi.Address, abi.Bytes),
		nil,
	)
	permitWitnessTransferFromFn = abi.NewFunction(
		"permitWitnessTransferFrom",
		abi.NewTypes(permitTransferFromT, signatureTransferDetailT, abi.Address, abi.Bytes32, abi.String, abi.Bytes),
		nil,
	)
	permitTransferFromBatchFn = abi.NewFunction(
		"permitTransferFrom",
		abi.NewTypes(permitBatchTransferFromT, abi.Slice(signatureTransferDetailT), abi.Address, abi.Bytes),
		nil,
	)
	permitWitnessTransferFromBatchFn = abi.NewFunction(
		"permitWitnessTransferFrom",
		abi.NewTypes(permitBatchTransferFromT, abi.Slice(signatureTransferDetailT), abi.Address, abi.Bytes32, abi.String, abi.Bytes),
		nil,
	)
	invalidateUnorderedNoncesFn = abi.NewFunction(
		"invalidateUnorderedNonces",
		abi.NewTypes(abi.Uint256, abi.Uint256),
		nil,
	)
)

// Events from IAllowanceTransfer's own surface.
var (
	nonceInvalidationEvent = abi.NewEvent("NonceInvalidation", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint48},
		{Type: abi.Uint48},
	}, false)
	approvalEvent = abi.NewEvent("Approval", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint160},
		{Type: abi.Uint48},
	}, false)
	permitEvent = abi.NewEvent("Permit", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint160},
		{Type: abi.Uint48},
		{Type: abi.Uint48},
	}, false)
	lockdownEvent = abi.NewEvent("Lockdown", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Address},
		{Type: abi.Address},
	}, false)
	unorderedNonceInvalidationEvent = abi.NewEvent("UnorderedNonceInvalidation", []abi.EventParam{
		{Type: abi.Address, Indexed: true},
		{Type: abi.Uint256},
		{Type: abi.Uint256},
	}, false)
)

// Custom errors from IAllowanceTransfer, PermitErrors (shared with SignatureTransfer), and the
// SignatureVerification library -- Permit2 predates nothing, every revert here is a Solidity
// custom error, never a require(cond, "string") reason.
var (
	errAllowanceExpired         = abi.NewError("AllowanceExpired", abi.NewTypes(abi.Uint256))
	errInsufficientAllowance    = abi.NewError("InsufficientAllowance", abi.NewTypes(abi.Uint256))
	errExcessiveInvalidation    = abi.NewError("ExcessiveInvalidation", nil)
	errSignatureExpired         = abi.NewError("SignatureExpired", abi.NewTypes(abi.Uint256))
	errInvalidNonce             = abi.NewError("InvalidNonce", nil)
	errInvalidSignatureLength   = abi.NewError("InvalidSignatureLength", nil)
	errInvalidSignature         = abi.NewError("InvalidSignature", nil)
	errInvalidSigner            = abi.NewError("InvalidSigner", nil)
	errInvalidContractSignature = abi.NewError("InvalidContractSignature", nil)
	errUnsafeCast               = abi.NewError("UnsafeCast", nil)
	errInvalidAmount            = abi.NewError("InvalidAmount", abi.NewTypes(abi.Uint256))
	errLengthMismatch           = abi.NewError("LengthMismatch", nil)
)

// ErrAllowanceExpired, etc. are sentinel errors wrapRevertError maps known Permit2 custom
// errors to, for errors.Is.
var (
	ErrAllowanceExpired         = errors.New("permit2: allowance expired")
	ErrInsufficientAllowance    = errors.New("permit2: insufficient allowance")
	ErrExcessiveInvalidation    = errors.New("permit2: excessive invalidation")
	ErrSignatureExpired         = errors.New("permit2: signature expired")
	ErrInvalidNonce             = errors.New("permit2: invalid nonce")
	ErrInvalidSignatureLength   = errors.New("permit2: invalid signature length")
	ErrInvalidSignature         = errors.New("permit2: invalid signature")
	ErrInvalidSigner            = errors.New("permit2: invalid signer")
	ErrInvalidContractSignature = errors.New("permit2: invalid contract signature")
	ErrUnsafeCast               = errors.New("permit2: unsafe cast")
	ErrInvalidAmount            = errors.New("permit2: invalid amount")
	ErrLengthMismatch           = errors.New("permit2: length mismatch")
	ErrUnknown                  = errors.New("permit2: unknown revert")
)

// EncodeDomainSeparator returns the calldata for DOMAIN_SEPARATOR().
func EncodeDomainSeparator() []byte {
	return domainSeparatorFn.SelectorBytes()
}

// DecodeDomainSeparator decodes the return data of DOMAIN_SEPARATOR.
func DecodeDomainSeparator(data []byte) (*types.Hash, error) {
	h, err := domainSeparatorFn.DecodeSingleReturnAs(data, types.NewHashFromBytes)
	if err != nil {
		return nil, fmt.Errorf("permit2: decodeDomainSeparator: %w", err)
	}
	return h, nil
}

// EncodeAllowance returns the calldata for allowance(user, token, spender).
func EncodeAllowance(user, token, spender *types.Address) ([]byte, error) {
	data, err := allowanceFn.EncodeCall(user, token, spender)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeAllowance: %w", err)
	}
	return data, nil
}

// DecodeAllowance decodes the return data of allowance.
func DecodeAllowance(data []byte) (*PackedAllowance, error) {
	vals, err := allowanceFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("permit2: decodeAllowance: %w", err)
	}
	amount, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("permit2: decodeAllowance: amount: unexpected type %T", vals[0])
	}
	expiration, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("permit2: decodeAllowance: expiration: unexpected type %T", vals[1])
	}
	nonce, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("permit2: decodeAllowance: nonce: unexpected type %T", vals[2])
	}
	return NewPackedAllowance(amount, expiration, nonce), nil
}

// EncodeApprove returns the calldata for approve(token, spender, amount, expiration).
func EncodeApprove(token, spender *types.Address, amount, expiration *big.Int) ([]byte, error) {
	data, err := approveFn.EncodeCall(token, spender, amount, expiration)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeApprove: %w", err)
	}
	return data, nil
}

// EncodePermitSingle returns the calldata for permit(owner, permitSingle, signature).
func EncodePermitSingle(owner *types.Address, permitSingle *PermitSingle, sig *types.Signature) ([]byte, error) {
	data, err := permitSingleFn.EncodeCall(owner, permitSingle.tuple(), sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitSingle: %w", err)
	}
	return data, nil
}

// EncodePermitBatch returns the calldata for permit(owner, permitBatch, signature).
func EncodePermitBatch(owner *types.Address, permitBatch *PermitBatch, sig *types.Signature) ([]byte, error) {
	data, err := permitBatchFn.EncodeCall(owner, permitBatch.tuple(), sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitBatch: %w", err)
	}
	return data, nil
}

// EncodeTransferFrom returns the calldata for transferFrom(from, to, amount, token).
func EncodeTransferFrom(from, to *types.Address, amount *big.Int, token *types.Address) ([]byte, error) {
	data, err := transferFromFn.EncodeCall(from, to, amount, token)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeTransferFrom: %w", err)
	}
	return data, nil
}

// EncodeTransferFromBatch returns the calldata for transferFrom(transferDetails).
func EncodeTransferFromBatch(transferDetails []AllowanceTransferDetails) ([]byte, error) {
	tuples := make([][]any, len(transferDetails))
	for i, d := range transferDetails {
		tuples[i] = d.tuple()
	}
	data, err := transferFromBatchFn.EncodeCall(tuples)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeTransferFromBatch: %w", err)
	}
	return data, nil
}

// EncodeLockdown returns the calldata for lockdown(approvals).
func EncodeLockdown(approvals []TokenSpenderPair) ([]byte, error) {
	tuples := make([][]any, len(approvals))
	for i, a := range approvals {
		tuples[i] = a.tuple()
	}
	data, err := lockdownFn.EncodeCall(tuples)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeLockdown: %w", err)
	}
	return data, nil
}

// EncodeInvalidateNonces returns the calldata for invalidateNonces(token, spender, newNonce).
func EncodeInvalidateNonces(token, spender *types.Address, newNonce *big.Int) ([]byte, error) {
	data, err := invalidateNoncesFn.EncodeCall(token, spender, newNonce)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeInvalidateNonces: %w", err)
	}
	return data, nil
}

// EncodeNonceBitmap returns the calldata for nonceBitmap(owner, wordPos).
func EncodeNonceBitmap(owner *types.Address, wordPos *big.Int) ([]byte, error) {
	data, err := nonceBitmapFn.EncodeCall(owner, wordPos)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeNonceBitmap: %w", err)
	}
	return data, nil
}

// DecodeNonceBitmap decodes the return data of nonceBitmap.
func DecodeNonceBitmap(data []byte) (*big.Int, error) {
	n, err := nonceBitmapFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("permit2: decodeNonceBitmap: %w", err)
	}
	return n, nil
}

// EncodePermitTransferFrom returns the calldata for
// permitTransferFrom(permit, transferDetails, owner, signature).
func EncodePermitTransferFrom(permit *PermitTransferFrom, transferDetails *SignatureTransferDetails, owner *types.Address, sig *types.Signature) ([]byte, error) {
	data, err := permitTransferFromFn.EncodeCall(permit.tuple(), transferDetails.tuple(), owner, sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitTransferFrom: %w", err)
	}
	return data, nil
}

// EncodePermitWitnessTransferFrom returns the calldata for
// permitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, signature).
func EncodePermitWitnessTransferFrom(permit *PermitTransferFrom, transferDetails *SignatureTransferDetails, owner *types.Address, witness *types.Hash, witnessTypeString string, sig *types.Signature) ([]byte, error) {
	data, err := permitWitnessTransferFromFn.EncodeCall(permit.tuple(), transferDetails.tuple(), owner, witness.Bytes(), witnessTypeString, sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitWitnessTransferFrom: %w", err)
	}
	return data, nil
}

// EncodePermitTransferFromBatch returns the calldata for
// permitTransferFrom(permit, transferDetails, owner, signature) (batch overload).
func EncodePermitTransferFromBatch(permit *PermitBatchTransferFrom, transferDetails []SignatureTransferDetails, owner *types.Address, sig *types.Signature) ([]byte, error) {
	tuples := make([][]any, len(transferDetails))
	for i, d := range transferDetails {
		tuples[i] = d.tuple()
	}
	data, err := permitTransferFromBatchFn.EncodeCall(permit.tuple(), tuples, owner, sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitTransferFromBatch: %w", err)
	}
	return data, nil
}

// EncodePermitWitnessTransferFromBatch returns the calldata for
// permitWitnessTransferFrom(permit, transferDetails, owner, witness, witnessTypeString, signature)
// (batch overload).
func EncodePermitWitnessTransferFromBatch(permit *PermitBatchTransferFrom, transferDetails []SignatureTransferDetails, owner *types.Address, witness *types.Hash, witnessTypeString string, sig *types.Signature) ([]byte, error) {
	tuples := make([][]any, len(transferDetails))
	for i, d := range transferDetails {
		tuples[i] = d.tuple()
	}
	data, err := permitWitnessTransferFromBatchFn.EncodeCall(permit.tuple(), tuples, owner, witness.Bytes(), witnessTypeString, sig.LegacyBytes())
	if err != nil {
		return nil, fmt.Errorf("permit2: encodePermitWitnessTransferFromBatch: %w", err)
	}
	return data, nil
}

// EncodeInvalidateUnorderedNonces returns the calldata for invalidateUnorderedNonces(wordPos, mask).
func EncodeInvalidateUnorderedNonces(wordPos, mask *big.Int) ([]byte, error) {
	data, err := invalidateUnorderedNoncesFn.EncodeCall(wordPos, mask)
	if err != nil {
		return nil, fmt.Errorf("permit2: encodeInvalidateUnorderedNonces: %w", err)
	}
	return data, nil
}
