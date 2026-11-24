package permit2

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
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

// ErrAllowanceExpired, etc. are sentinel errors MapRevertError maps known Permit2 custom
// errors to, for errors.Is.
var (
	ErrAllowanceExpired         = errors.New("allowance expired")
	ErrInsufficientAllowance    = errors.New("insufficient allowance")
	ErrExcessiveInvalidation    = errors.New("excessive invalidation")
	ErrSignatureExpired         = errors.New("signature expired")
	ErrInvalidNonce             = errors.New("invalid nonce")
	ErrInvalidSignatureLength   = errors.New("invalid signature length")
	ErrInvalidSignature         = errors.New("invalid signature")
	ErrInvalidSigner            = errors.New("invalid signer")
	ErrInvalidContractSignature = errors.New("invalid contract signature")
	ErrUnsafeCast               = errors.New("unsafe cast")
	ErrInvalidAmount            = errors.New("invalid amount")
	ErrLengthMismatch           = errors.New("length mismatch")
	ErrUnknown                  = errors.New("unknown revert")
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

// MapRevertError maps a known Permit2 custom error in err to its sentinel.
func MapRevertError(err error) error {
	if rpcErr, ok := errors.AsType[*rpc.ResponseError](err); ok && rpcErr == nil {
		return nil
	}

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
	case errAllowanceExpired.Selector():
		return fmt.Errorf("%w: %w", ErrAllowanceExpired, err)
	case errInsufficientAllowance.Selector():
		return fmt.Errorf("%w: %w", ErrInsufficientAllowance, err)
	case errExcessiveInvalidation.Selector():
		return fmt.Errorf("%w: %w", ErrExcessiveInvalidation, err)
	case errSignatureExpired.Selector():
		return fmt.Errorf("%w: %w", ErrSignatureExpired, err)
	case errInvalidNonce.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidNonce, err)
	case errInvalidSignatureLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignatureLength, err)
	case errInvalidSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	case errInvalidSigner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSigner, err)
	case errInvalidContractSignature.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidContractSignature, err)
	case errUnsafeCast.Selector():
		return fmt.Errorf("%w: %w", ErrUnsafeCast, err)
	case errInvalidAmount.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidAmount, err)
	case errLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrLengthMismatch, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
