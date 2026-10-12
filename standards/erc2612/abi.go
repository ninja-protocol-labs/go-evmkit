package erc2612

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Function signatures from IERC20Permit, for encoding calls and decoding results.
var (
	permitFn          = abi.NewFunction("permit", abi.NewTypes(abi.Address, abi.Address, abi.Uint256, abi.Uint256, abi.Uint8, abi.Bytes32, abi.Bytes32), nil)
	noncesFn          = abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	domainSeparatorFn = abi.NewFunction("DOMAIN_SEPARATOR", nil, abi.NewTypes(abi.Bytes32))
)

// Custom errors from ERC20Permit's own implementation (not part of IERC20Permit itself).
var (
	errExpiredSignature = abi.NewError("ERC2612ExpiredSignature", abi.NewTypes(abi.Uint256))
	errInvalidSigner    = abi.NewError("ERC2612InvalidSigner", abi.NewTypes(abi.Address, abi.Address))
)

// ErrExpiredSignature, etc. mirror ERC20Permit's custom errors, for errors.Is.
var (
	ErrExpiredSignature = errors.New("erc2612: expired signature")
	ErrInvalidSigner    = errors.New("erc2612: invalid signer")
	ErrUnknown          = errors.New("erc2612: unknown revert")
)

// EncodePermit returns the calldata for permit(owner, spender, value, deadline, v, r, s), where
// v, r, s come from sig -- the EIP-712 signature over Permit.Hash, not a transaction signature.
// Uses sig.LegacyV() (27/28), since ecrecover/ECDSA.recover expect that encoding.
func EncodePermit(owner, spender *types.Address, value, deadline *big.Int, sig *types.Signature) ([]byte, error) {
	v := uint8(sig.LegacyV().Uint64())
	r := make([]byte, 32)
	sig.R().FillBytes(r)
	s := make([]byte, 32)
	sig.S().FillBytes(s)

	data, err := permitFn.EncodeCall(owner, spender, value, deadline, v, r, s)
	if err != nil {
		return nil, fmt.Errorf("erc2612: encodePermit: %w", err)
	}
	return data, nil
}

// EncodeNonces returns the calldata for nonces(owner).
func EncodeNonces(owner *types.Address) ([]byte, error) {
	data, err := noncesFn.EncodeCall(owner)
	if err != nil {
		return nil, fmt.Errorf("erc2612: encodeNonces: %w", err)
	}
	return data, nil
}

// DecodeNonces decodes the return data of nonces.
func DecodeNonces(data []byte) (*big.Int, error) {
	n, err := noncesFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: decodeNonces: %w", err)
	}
	return n, nil
}

// EncodeDomainSeparator returns the calldata for DOMAIN_SEPARATOR().
func EncodeDomainSeparator() []byte {
	return domainSeparatorFn.SelectorBytes()
}

// DecodeDomainSeparator decodes the return data of DOMAIN_SEPARATOR.
func DecodeDomainSeparator(data []byte) ([]byte, error) {
	b, err := domainSeparatorFn.DecodeSingleReturn[[]byte](data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: decodeDomainSeparator: %w", err)
	}
	return b, nil
}
