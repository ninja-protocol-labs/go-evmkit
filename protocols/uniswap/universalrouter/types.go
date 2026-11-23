package universalrouter

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

const (
	universalRouterDomainName    = "UniversalRouter"
	universalRouterDomainVersion = "2"
	executeSignedPrimaryType     = "ExecuteSigned"
)

// executeSignedTypes matches RouteSigner's EXECUTE_SIGNED_TYPEHASH.
var executeSignedTypes = eip712.Types{
	executeSignedPrimaryType: {
		eip712.NewField("commands", abi.Bytes.String()),
		eip712.NewField("inputs", "bytes[]"),
		eip712.NewField("intent", abi.Bytes32.String()),
		eip712.NewField("data", abi.Bytes32.String()),
		eip712.NewField("sender", abi.Address.String()),
		eip712.NewField("nonce", abi.Bytes32.String()),
		eip712.NewField("deadline", abi.Uint256.String()),
	},
}

// NewUniversalRouterDomain returns the EIP-712 domain executeSigned routes are signed under: name "UniversalRouter", version "2".
func NewUniversalRouterDomain(chainID *big.Int, verifyingContract *types.Address) *eip712.Domain {
	return eip712.NewDomain(universalRouterDomainName, universalRouterDomainVersion, chainID, verifyingContract, nil)
}

// UncheckedNonce returns bytes32(type(uint256).max), the nonce executeSigned does not record, so the signature can be replayed until Deadline.
func UncheckedNonce() [32]byte {
	var nonce [32]byte
	for i := range nonce {
		nonce[i] = 0xff
	}
	return nonce
}

// ExecuteSignedMessage is the EIP-712 message an executeSigned route is signed over; Sender is the zero address unless the route is sent with verifySender.
type ExecuteSignedMessage struct {
	Plan     *Plan
	Intent   [32]byte
	Data     [32]byte
	Sender   *types.Address
	Nonce    [32]byte
	Deadline *big.Int
}

func NewExecuteSignedMessage(plan *Plan, intent, data [32]byte, sender *types.Address, nonce [32]byte, deadline *big.Int) *ExecuteSignedMessage {
	return &ExecuteSignedMessage{
		Plan:     plan,
		Intent:   intent,
		Data:     data,
		Sender:   sender,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Hash returns the signable EIP-712 digest for m under domain.
func (m *ExecuteSignedMessage) Hash(domain *eip712.Domain) (*types.Hash, error) {
	if m == nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: hash: %w: message is nil", bind.ErrInvalidRequest)
	}

	if m.Sender == nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: hash: %w: sender is nil", bind.ErrInvalidRequest)
	}

	if m.Deadline == nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: hash: %w: deadline is nil", bind.ErrInvalidRequest)
	}

	commands, inputs, err := m.Plan.Encode()
	if err != nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: hash: %w", err)
	}

	message := eip712.Message{
		"commands": commands,
		"inputs":   inputs,
		"intent":   m.Intent[:],
		"data":     m.Data[:],
		"sender":   m.Sender,
		"nonce":    m.Nonce[:],
		"deadline": m.Deadline,
	}

	h, err := eip712.Hash(domain, executeSignedPrimaryType, executeSignedTypes, message)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: hash: %w", err)
	}

	return h, nil
}

// Sign returns signer's signature over m's EIP-712 digest under domain; pass LegacyBytes() as executeSigned's signature.
func (m *ExecuteSignedMessage) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := m.Hash(domain)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("universalrouter: executeSignedMessage: sign: %w", err)
	}

	return sig, nil
}

// RouteContext is signedRouteContext()'s result: the signer, intent and data of the executeSigned call in progress.
type RouteContext struct {
	Signer *types.Address
	Intent [32]byte
	Data   [32]byte
}

func NewRouteContext(signer *types.Address, intent, data [32]byte) *RouteContext {
	return &RouteContext{
		Signer: signer,
		Intent: intent,
		Data:   data,
	}
}

// ExecutionFailedError is a decoded ExecutionFailed(commandIndex, message) revert; errors.Is also matches the sentinels mapped from Message.
type ExecutionFailedError struct {
	CommandIndex *big.Int
	Message      []byte
	cause        error
}

func NewExecutionFailedError(commandIndex *big.Int, message []byte, cause error) *ExecutionFailedError {
	return &ExecutionFailedError{
		CommandIndex: commandIndex,
		Message:      message,
		cause:        cause,
	}
}

func (e *ExecutionFailedError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: command %s: %s", ErrExecutionFailed, e.CommandIndex, e.cause)
	}
	return fmt.Sprintf("%s: command %s: 0x%x", ErrExecutionFailed, e.CommandIndex, e.Message)
}

func (e *ExecutionFailedError) Unwrap() error {
	return ErrExecutionFailed
}

func (e *ExecutionFailedError) Is(target error) bool {
	return e.cause != nil && errors.Is(e.cause, target)
}

// ExecuteRequest is execute(commands, inputs[, deadline])'s input on Contract; a nil Deadline selects the deadline-less overload, and NativeAmount, if set, is sent as msg.value.
type ExecuteRequest struct {
	Contract     *types.Address
	Plan         *Plan
	Deadline     *big.Int
	NativeAmount *big.Int
}

func NewExecuteRequest(contract *types.Address, plan *Plan, deadline, nativeAmount *big.Int) *ExecuteRequest {
	return &ExecuteRequest{
		Contract:     contract,
		Plan:         plan,
		Deadline:     deadline,
		NativeAmount: nativeAmount,
	}
}

func (r *ExecuteRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := r.Plan.Validate(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// ExecuteSignedRequest is executeSigned's input on Contract; Signature is the route signer's 65-byte signature over the matching ExecuteSignedMessage.
type ExecuteSignedRequest struct {
	Contract     *types.Address
	Plan         *Plan
	Intent       [32]byte
	Data         [32]byte
	VerifySender bool
	Nonce        [32]byte
	Signature    []byte
	Deadline     *big.Int
	NativeAmount *big.Int
}

func NewExecuteSignedRequest(contract *types.Address, plan *Plan, intent, data [32]byte, verifySender bool, nonce [32]byte, signature []byte, deadline, nativeAmount *big.Int) *ExecuteSignedRequest {
	return &ExecuteSignedRequest{
		Contract:     contract,
		Plan:         plan,
		Intent:       intent,
		Data:         data,
		VerifySender: verifySender,
		Nonce:        nonce,
		Signature:    signature,
		Deadline:     deadline,
		NativeAmount: nativeAmount,
	}
}

func (r *ExecuteSignedRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := r.Plan.Validate(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}

	if len(r.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// NoncesUsedRequest is noncesUsed(user, nonce)'s input on Contract.
type NoncesUsedRequest struct {
	Contract *types.Address
	User     *types.Address
	Nonce    [32]byte
}

func NewNoncesUsedRequest(contract, user *types.Address, nonce [32]byte) *NoncesUsedRequest {
	return &NoncesUsedRequest{
		Contract: contract,
		User:     user,
		Nonce:    nonce,
	}
}

func (r *NoncesUsedRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.User == nil {
		return fmt.Errorf("%w: user is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SignedRouteContextRequest is signedRouteContext()'s input on Contract.
type SignedRouteContextRequest struct {
	Contract *types.Address
}

func NewSignedRouteContextRequest(contract *types.Address) *SignedRouteContextRequest {
	return &SignedRouteContextRequest{
		Contract: contract,
	}
}

func (r *SignedRouteContextRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}
