// Package erc2612 encodes calls to and decodes results from an ERC-2612
// (Permit) token contract: gasless ERC-20 approvals authorized by an
// off-chain EIP-712 signature instead of an on-chain approve() transaction.
// It has no erc165/erc20 embedding, since IERC20Permit doesn't extend
// either at the interface level; use it alongside an *erc20.ERC20 for the
// same contract address.
package erc2612

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

// IERC2612 mirrors IERC20Permit.sol. No erc165/erc20 embedding -- see the package doc above.
type IERC2612 interface {
	Nonces(ctx context.Context, contract, owner *types.Address, block string) (*big.Int, error)
	DomainSeparator(ctx context.Context, contract *types.Address, block string) ([]byte, error)
	Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IERC2612 = (*ERC2612)(nil)

// ERC2612 calls any ERC-2612 (Permit) token contract over a Client; it holds no
// contract address, so one instance serves every collection.
type ERC2612 struct {
	c rpc.Client
}

// New returns an ERC2612 calling tokens via cli.
func New(cli rpc.Client) *ERC2612 {
	return &ERC2612{
		c: cli,
	}
}

// Nonces returns owner's current nonce on contract.
func (e *ERC2612) Nonces(ctx context.Context, contract, owner *types.Address, block string) (*big.Int, error) {
	calldata, err := EncodeNonces(owner)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}
	data, err := e.call(ctx, nil, contract, calldata, block)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}
	n, err := DecodeNonces(data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: nonces: %w", err)
	}
	return n, nil
}

// DomainSeparator returns contract's EIP-712 domain separator.
func (e *ERC2612) DomainSeparator(ctx context.Context, contract *types.Address, block string) ([]byte, error) {
	data, err := e.call(ctx, nil, contract, EncodeDomainSeparator(), block)
	if err != nil {
		return nil, fmt.Errorf("erc2612: domainSeparator: %w", err)
	}
	b, err := DecodeDomainSeparator(data)
	if err != nil {
		return nil, fmt.Errorf("erc2612: domainSeparator: %w", err)
	}
	return b, nil
}

// CreatePermit builds an unsigned permit(owner, spender, value, deadline, v, r, s) transaction via
// b, sent by sender. sender and owner may differ -- anyone holding a valid Permit signature (sig,
// from (*Permit).Sign) can relay it; owner is only the signer the signature authorizes.
func (e *ERC2612) CreatePermit(
	ctx context.Context,
	b tx.Builder,
	contract, sender, owner, spender *types.Address,
	value, deadline *big.Int,
	sig *types.Signature,
) (core.Transaction, error) {
	calldata, err := EncodePermit(owner, spender, value, deadline, sig)
	if err != nil {
		return nil, fmt.Errorf("erc2612: createPermit: %w", err)
	}

	unsigned, err := b.Build(sender, contract, nil, calldata).Pack(ctx, e.c)
	if err != nil {
		return nil, fmt.Errorf("erc2612: createPermit: %w", e.wrapRevertError(err))
	}
	return unsigned, nil
}

// Permit broadcasts an already-signed permit transaction.
func (e *ERC2612) Permit(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("erc2612: permit: %w", err)
	}
	hash, err := tx.Broadcast(ctx, e.c, raw)
	if err != nil {
		return nil, fmt.Errorf("erc2612: permit: %w", err)
	}
	return hash, nil
}

// call eth_calls contract with calldata (from msg.sender from, if non-nil), decoding a reverted custom error to its sentinel.
func (e *ERC2612) call(ctx context.Context, from, contract *types.Address, calldata []byte, block string) ([]byte, error) {
	params := map[string]any{
		"to":   contract.String(),
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

// wrapRevertError maps a known ERC20Permit custom error in err to its sentinel.
func (_ *ERC2612) wrapRevertError(err error) error {
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
	case errExpiredSignature.Selector():
		return fmt.Errorf("%w: %w", ErrExpiredSignature, err)
	case errInvalidSigner.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidSigner, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
