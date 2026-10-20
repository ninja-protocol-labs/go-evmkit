package permit2

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

const (
	domainName                                = "Permit2"
	permitDetailsPrimaryType                  = "PermitDetails"
	permitSinglePrimaryType                   = "PermitSingle"
	permitBatchPrimaryType                    = "PermitBatch"
	tokenPermissionsPrimaryType               = "TokenPermissions"
	permitTransferFromPrimaryType             = "PermitTransferFrom"
	permitBatchTransferFromPrimaryType        = "PermitBatchTransferFrom"
	permitWitnessTransferFromPrimaryType      = "PermitWitnessTransferFrom"
	permitBatchWitnessTransferFromPrimaryType = "PermitBatchWitnessTransferFrom"
	tokenFieldName                            = "token"
	amountFieldName                           = "amount"
	expirationFieldName                       = "expiration"
	nonceFieldName                            = "nonce"
	detailsFieldName                          = "details"
	spenderFieldName                          = "spender"
	sigDeadlineFieldName                      = "sigDeadline"
	permittedFieldName                        = "permitted"
	deadlineFieldName                         = "deadline"
	witnessFieldName                          = "witness"

	// PermitHash.sol's witness typehash stubs -- caller's witnessTypeString completes them.
	permitTransferFromWitnessStub      = "PermitWitnessTransferFrom(TokenPermissions permitted,address spender,uint256 nonce,uint256 deadline,"
	permitBatchTransferFromWitnessStub = "PermitBatchWitnessTransferFrom(TokenPermissions[] permitted,address spender,uint256 nonce,uint256 deadline,"
)

// NewDomain returns the one EIP-712 domain every Permit2 deployment signs under: just name,
// chainID, and verifyingContract -- no version, no salt.
func NewDomain(chainID *big.Int, verifyingContract *types.Address) *eip712.Domain {
	return eip712.NewDomain(domainName, "", chainID, verifyingContract, nil)
}

var (
	// permitDetailsTypeFields is PermitDetails' own field list, nested under both PermitSingle and
	// PermitBatch's schemas below.
	permitDetailsTypeFields = []eip712.Field{
		eip712.NewField(tokenFieldName, abi.Address.String()),
		eip712.NewField(amountFieldName, abi.Uint160.String()),
		eip712.NewField(expirationFieldName, abi.Uint48.String()),
		eip712.NewField(nonceFieldName, abi.Uint48.String()),
	}

	// permitSingleTypes is the EIP-712 schema for PermitSingle, matching PermitHash.sol's
	// _PERMIT_SINGLE_TYPEHASH exactly.
	permitSingleTypes = eip712.Types{
		permitSinglePrimaryType: {
			eip712.NewField(detailsFieldName, permitDetailsPrimaryType),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(sigDeadlineFieldName, abi.Uint256.String()),
		},
		permitDetailsPrimaryType: permitDetailsTypeFields,
	}

	// permitBatchTypes is the EIP-712 schema for PermitBatch, matching PermitHash.sol's
	// _PERMIT_BATCH_TYPEHASH exactly.
	permitBatchTypes = eip712.Types{
		permitBatchPrimaryType: {
			eip712.NewField(detailsFieldName, permitDetailsPrimaryType+"[]"),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(sigDeadlineFieldName, abi.Uint256.String()),
		},
		permitDetailsPrimaryType: permitDetailsTypeFields,
	}

	// tokenPermissionsTypeFields is TokenPermissions' own field list, nested under
	// PermitTransferFrom/PermitBatchTransferFrom's schemas below.
	tokenPermissionsTypeFields = []eip712.Field{
		eip712.NewField(tokenFieldName, abi.Address.String()),
		eip712.NewField(amountFieldName, abi.Uint256.String()),
	}

	// permitTransferFromTypes matches PermitHash.sol's _PERMIT_TRANSFER_FROM_TYPEHASH. "spender"
	// isn't a struct field -- it's msg.sender at call time, signed over anyway.
	permitTransferFromTypes = eip712.Types{
		permitTransferFromPrimaryType: {
			eip712.NewField(permittedFieldName, tokenPermissionsPrimaryType),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(nonceFieldName, abi.Uint256.String()),
			eip712.NewField(deadlineFieldName, abi.Uint256.String()),
		},
		tokenPermissionsPrimaryType: tokenPermissionsTypeFields,
	}

	// permitBatchTransferFromTypes is the EIP-712 schema for PermitBatchTransferFrom, matching
	// PermitHash.sol's _PERMIT_BATCH_TRANSFER_FROM_TYPEHASH exactly.
	permitBatchTransferFromTypes = eip712.Types{
		permitBatchTransferFromPrimaryType: {
			eip712.NewField(permittedFieldName, tokenPermissionsPrimaryType+"[]"),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(nonceFieldName, abi.Uint256.String()),
			eip712.NewField(deadlineFieldName, abi.Uint256.String()),
		},
		tokenPermissionsPrimaryType: tokenPermissionsTypeFields,
	}

	// permitTransferFromWitnessTypes describes PermitWitnessTransferFrom's fields for value
	// encoding only -- its typeHash comes from the caller's witnessTypeString, not this schema.
	permitTransferFromWitnessTypes = eip712.Types{
		permitWitnessTransferFromPrimaryType: {
			eip712.NewField(permittedFieldName, tokenPermissionsPrimaryType),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(nonceFieldName, abi.Uint256.String()),
			eip712.NewField(deadlineFieldName, abi.Uint256.String()),
			eip712.NewField(witnessFieldName, abi.Bytes32.String()),
		},
		tokenPermissionsPrimaryType: tokenPermissionsTypeFields,
	}

	// permitBatchTransferFromWitnessTypes is PermitBatchWitnessTransferFrom's value-encoding
	// counterpart to permitTransferFromWitnessTypes above.
	permitBatchTransferFromWitnessTypes = eip712.Types{
		permitBatchWitnessTransferFromPrimaryType: {
			eip712.NewField(permittedFieldName, tokenPermissionsPrimaryType+"[]"),
			eip712.NewField(spenderFieldName, abi.Address.String()),
			eip712.NewField(nonceFieldName, abi.Uint256.String()),
			eip712.NewField(deadlineFieldName, abi.Uint256.String()),
			eip712.NewField(witnessFieldName, abi.Bytes32.String()),
		},
		tokenPermissionsPrimaryType: tokenPermissionsTypeFields,
	}
)

// PermitDetails is AllowanceTransfer's per-token permit data.
type PermitDetails struct {
	Token      *types.Address
	Amount     *big.Int // uint160
	Expiration *big.Int // uint48
	Nonce      *big.Int // uint48
}

func NewPermitDetails(token *types.Address, amount, expiration, nonce *big.Int) *PermitDetails {
	return &PermitDetails{
		Token:      token,
		Amount:     amount,
		Expiration: expiration,
		Nonce:      nonce,
	}
}

func (d *PermitDetails) message() eip712.Message {
	return eip712.Message{
		tokenFieldName:      d.Token,
		amountFieldName:     d.Amount,
		expirationFieldName: d.Expiration,
		nonceFieldName:      d.Nonce,
	}
}

// tuple returns d as the []any ABI encoding expects for a permitDetailsT value.
func (d *PermitDetails) tuple() []any {
	return []any{d.Token, d.Amount, d.Expiration, d.Nonce}
}

// PermitSingle is the EIP-712 message signed to grant a single-token allowance via permit().
type PermitSingle struct {
	Details     *PermitDetails
	Spender     *types.Address
	SigDeadline *big.Int
}

func NewPermitSingle(details *PermitDetails, spender *types.Address, sigDeadline *big.Int) *PermitSingle {
	return &PermitSingle{
		Details:     details,
		Spender:     spender,
		SigDeadline: sigDeadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *PermitSingle) Hash(domain *eip712.Domain) (*types.Hash, error) {
	message := eip712.Message{
		detailsFieldName:     p.Details.message(),
		spenderFieldName:     p.Spender,
		sigDeadlineFieldName: p.SigDeadline,
	}

	h, err := eip712.Hash(domain, permitSinglePrimaryType, permitSingleTypes, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain.
func (p *PermitSingle) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: sign: %w", err)
	}
	return sig, nil
}

// tuple returns p as the []any ABI encoding expects for a permitSingleT value.
func (p *PermitSingle) tuple() []any {
	return []any{p.Details.tuple(), p.Spender, p.SigDeadline}
}

// PermitBatch is the EIP-712 message signed to grant multiple token allowances via permitBatch().
type PermitBatch struct {
	Details     []PermitDetails
	Spender     *types.Address
	SigDeadline *big.Int
}

func NewPermitBatch(details []PermitDetails, spender *types.Address, sigDeadline *big.Int) *PermitBatch {
	return &PermitBatch{
		Details:     details,
		Spender:     spender,
		SigDeadline: sigDeadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *PermitBatch) Hash(domain *eip712.Domain) (*types.Hash, error) {
	details := make([]any, len(p.Details))
	for i, d := range p.Details {
		details[i] = d.message()
	}
	message := eip712.Message{
		detailsFieldName:     details,
		spenderFieldName:     p.Spender,
		sigDeadlineFieldName: p.SigDeadline,
	}

	h, err := eip712.Hash(domain, permitBatchPrimaryType, permitBatchTypes, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain.
func (p *PermitBatch) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: sign: %w", err)
	}
	return sig, nil
}

// tuple returns p as the []any ABI encoding expects for a permitBatchT value.
func (p *PermitBatch) tuple() []any {
	details := make([][]any, len(p.Details))
	for i, d := range p.Details {
		details[i] = d.tuple()
	}
	return []any{details, p.Spender, p.SigDeadline}
}

// PackedAllowance is allowance()'s (amount, expiration, nonce) return value.
type PackedAllowance struct {
	Amount     *big.Int // uint160
	Expiration *big.Int // uint48
	Nonce      *big.Int // uint48
}

func NewPackedAllowance(amount, expiration, nonce *big.Int) *PackedAllowance {
	return &PackedAllowance{
		Amount:     amount,
		Expiration: expiration,
		Nonce:      nonce,
	}
}

// TokenSpenderPair identifies one (token, spender) allowance to zero out via lockdown().
type TokenSpenderPair struct {
	Token   *types.Address
	Spender *types.Address
}

func NewTokenSpenderPair(token, spender *types.Address) *TokenSpenderPair {
	return &TokenSpenderPair{
		Token:   token,
		Spender: spender,
	}
}

// tuple returns p as the []any ABI encoding expects for a tokenSpenderPairT value.
func (p *TokenSpenderPair) tuple() []any {
	return []any{p.Token, p.Spender}
}

// AllowanceTransferDetails is one leg of a batched transferFrom() under AllowanceTransfer.
type AllowanceTransferDetails struct {
	From   *types.Address
	To     *types.Address
	Amount *big.Int // uint160
	Token  *types.Address
}

func NewAllowanceTransferDetails(from, to *types.Address, amount *big.Int, token *types.Address) *AllowanceTransferDetails {
	return &AllowanceTransferDetails{
		From:   from,
		To:     to,
		Amount: amount,
		Token:  token,
	}
}

// tuple returns d as the []any ABI encoding expects for an allowanceTransferDetailsT value.
func (d *AllowanceTransferDetails) tuple() []any {
	return []any{d.From, d.To, d.Amount, d.Token}
}

// TokenPermissions is the token and amount signed over in a SignatureTransfer permit.
type TokenPermissions struct {
	Token  *types.Address
	Amount *big.Int // uint256
}

func NewTokenPermissions(token *types.Address, amount *big.Int) *TokenPermissions {
	return &TokenPermissions{
		Token:  token,
		Amount: amount,
	}
}

func (p *TokenPermissions) message() eip712.Message {
	return eip712.Message{
		tokenFieldName:  p.Token,
		amountFieldName: p.Amount,
	}
}

// tuple returns p as the []any ABI encoding expects for a tokenPermissionsT value.
func (p *TokenPermissions) tuple() []any {
	return []any{p.Token, p.Amount}
}

// PermitTransferFrom is the EIP-712 message signed for a single one-time SignatureTransfer.
type PermitTransferFrom struct {
	Permitted *TokenPermissions
	Nonce     *big.Int // uint256
	Deadline  *big.Int
}

func NewPermitTransferFrom(permitted *TokenPermissions, nonce, deadline *big.Int) *PermitTransferFrom {
	return &PermitTransferFrom{
		Permitted: permitted,
		Nonce:     nonce,
		Deadline:  deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain, bound to spender -- the address
// that will call permitTransferFrom, signed over even though it isn't a field of p itself.
func (p *PermitTransferFrom) Hash(domain *eip712.Domain, spender *types.Address) (*types.Hash, error) {
	message := eip712.Message{
		permittedFieldName: p.Permitted.message(),
		spenderFieldName:   spender,
		nonceFieldName:     p.Nonce,
		deadlineFieldName:  p.Deadline,
	}

	h, err := eip712.Hash(domain, permitTransferFromPrimaryType, permitTransferFromTypes, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain, bound to spender.
func (p *PermitTransferFrom) Sign(domain *eip712.Domain, spender *types.Address, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain, spender)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: sign: %w", err)
	}
	return sig, nil
}

// HashWithWitness is like Hash, but extended with witness per witnessTypeString -- the caller's
// own type-string fragment completing PermitHash.sol's witness typehash stub.
func (p *PermitTransferFrom) HashWithWitness(domain *eip712.Domain, spender *types.Address, witness *types.Hash, witnessTypeString string) (*types.Hash, error) {
	typeHash := core.Keccak256([]byte(permitTransferFromWitnessStub + witnessTypeString))
	message := eip712.Message{
		permittedFieldName: p.Permitted.message(),
		spenderFieldName:   spender,
		nonceFieldName:     p.Nonce,
		deadlineFieldName:  p.Deadline,
		witnessFieldName:   witness.Bytes(),
	}

	structHash, err := permitTransferFromWitnessTypes.HashStructWithTypeHash(typeHash, permitWitnessTransferFromPrimaryType, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hashWithWitness: %w", err)
	}

	h, err := eip712.FinalizeHash(domain, structHash)
	if err != nil {
		return nil, fmt.Errorf("permit2: hashWithWitness: %w", err)
	}
	return h, nil
}

// SignWithWitness returns signer's signature over p's witness-extended EIP-712 digest under domain.
func (p *PermitTransferFrom) SignWithWitness(domain *eip712.Domain, spender *types.Address, witness *types.Hash, witnessTypeString string, signer core.Signer) (*types.Signature, error) {
	digest, err := p.HashWithWitness(domain, spender, witness, witnessTypeString)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: signWithWitness: %w", err)
	}
	return sig, nil
}

// tuple returns p as the []any ABI encoding expects for a permitTransferFromT value.
func (p *PermitTransferFrom) tuple() []any {
	return []any{p.Permitted.tuple(), p.Nonce, p.Deadline}
}

// SignatureTransferDetails is the recipient and requested amount for one leg of a transfer.
type SignatureTransferDetails struct {
	To              *types.Address
	RequestedAmount *big.Int
}

func NewSignatureTransferDetails(to *types.Address, requestedAmount *big.Int) *SignatureTransferDetails {
	return &SignatureTransferDetails{
		To:              to,
		RequestedAmount: requestedAmount,
	}
}

// tuple returns d as the []any ABI encoding expects for a signatureTransferDetailsT value.
func (d *SignatureTransferDetails) tuple() []any {
	return []any{d.To, d.RequestedAmount}
}

// PermitBatchTransferFrom is the EIP-712 message signed for a one-time SignatureTransfer of
// multiple tokens.
type PermitBatchTransferFrom struct {
	Permitted []TokenPermissions
	Nonce     *big.Int // uint256
	Deadline  *big.Int
}

func NewPermitBatchTransferFrom(permitted []TokenPermissions, nonce, deadline *big.Int) *PermitBatchTransferFrom {
	return &PermitBatchTransferFrom{
		Permitted: permitted,
		Nonce:     nonce,
		Deadline:  deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain, bound to spender -- the address
// that will call permitTransferFrom, signed over even though it isn't a field of p itself.
func (p *PermitBatchTransferFrom) Hash(domain *eip712.Domain, spender *types.Address) (*types.Hash, error) {
	permitted := make([]any, len(p.Permitted))
	for i, t := range p.Permitted {
		permitted[i] = t.message()
	}
	message := eip712.Message{
		permittedFieldName: permitted,
		spenderFieldName:   spender,
		nonceFieldName:     p.Nonce,
		deadlineFieldName:  p.Deadline,
	}

	h, err := eip712.Hash(domain, permitBatchTransferFromPrimaryType, permitBatchTransferFromTypes, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain, bound to spender.
func (p *PermitBatchTransferFrom) Sign(domain *eip712.Domain, spender *types.Address, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain, spender)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: sign: %w", err)
	}
	return sig, nil
}

// HashWithWitness returns the signable EIP-712 digest for p under domain, bound to spender and
// extended with witness per witnessTypeString (see PermitTransferFrom.HashWithWitness).
func (p *PermitBatchTransferFrom) HashWithWitness(domain *eip712.Domain, spender *types.Address, witness *types.Hash, witnessTypeString string) (*types.Hash, error) {
	typeHash := core.Keccak256([]byte(permitBatchTransferFromWitnessStub + witnessTypeString))

	permitted := make([]any, len(p.Permitted))
	for i, t := range p.Permitted {
		permitted[i] = t.message()
	}
	message := eip712.Message{
		permittedFieldName: permitted,
		spenderFieldName:   spender,
		nonceFieldName:     p.Nonce,
		deadlineFieldName:  p.Deadline,
		witnessFieldName:   witness.Bytes(),
	}

	structHash, err := permitBatchTransferFromWitnessTypes.HashStructWithTypeHash(typeHash, permitBatchWitnessTransferFromPrimaryType, message)
	if err != nil {
		return nil, fmt.Errorf("permit2: hashWithWitness: %w", err)
	}

	h, err := eip712.FinalizeHash(domain, structHash)
	if err != nil {
		return nil, fmt.Errorf("permit2: hashWithWitness: %w", err)
	}
	return h, nil
}

// SignWithWitness returns signer's signature over p's witness-extended EIP-712 digest under domain.
func (p *PermitBatchTransferFrom) SignWithWitness(domain *eip712.Domain, spender *types.Address, witness *types.Hash, witnessTypeString string, signer core.Signer) (*types.Signature, error) {
	digest, err := p.HashWithWitness(domain, spender, witness, witnessTypeString)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("permit2: signWithWitness: %w", err)
	}
	return sig, nil
}

// tuple returns p as the []any ABI encoding expects for a permitBatchTransferFromT value.
func (p *PermitBatchTransferFrom) tuple() []any {
	permitted := make([][]any, len(p.Permitted))
	for i, t := range p.Permitted {
		permitted[i] = t.tuple()
	}
	return []any{permitted, p.Nonce, p.Deadline}
}

// NonceInvalidation mirrors AllowanceTransfer's NonceInvalidation event.
type NonceInvalidation struct {
	Contract *types.Address
	Owner    *types.Address
	Token    *types.Address
	Spender  *types.Address
	NewNonce *big.Int
	OldNonce *big.Int
}

func NewNonceInvalidation(contract, owner, token, spender *types.Address, newNonce, oldNonce *big.Int) *NonceInvalidation {
	return &NonceInvalidation{
		Contract: contract,
		Owner:    owner,
		Token:    token,
		Spender:  spender,
		NewNonce: newNonce,
		OldNonce: oldNonce,
	}
}

// Approval mirrors AllowanceTransfer's Approval event.
type Approval struct {
	Contract   *types.Address
	Owner      *types.Address
	Token      *types.Address
	Spender    *types.Address
	Amount     *big.Int
	Expiration *big.Int
}

func NewApproval(contract, owner, token, spender *types.Address, amount, expiration *big.Int) *Approval {
	return &Approval{
		Contract:   contract,
		Owner:      owner,
		Token:      token,
		Spender:    spender,
		Amount:     amount,
		Expiration: expiration,
	}
}

// Permit mirrors AllowanceTransfer's Permit event.
type Permit struct {
	Contract   *types.Address
	Owner      *types.Address
	Token      *types.Address
	Spender    *types.Address
	Amount     *big.Int
	Expiration *big.Int
	Nonce      *big.Int
}

func NewPermit(contract, owner, token, spender *types.Address, amount, expiration, nonce *big.Int) *Permit {
	return &Permit{
		Contract:   contract,
		Owner:      owner,
		Token:      token,
		Spender:    spender,
		Amount:     amount,
		Expiration: expiration,
		Nonce:      nonce,
	}
}

// Lockdown mirrors AllowanceTransfer's Lockdown event.
type Lockdown struct {
	Contract *types.Address
	Owner    *types.Address
	Token    *types.Address
	Spender  *types.Address
}

func NewLockdown(contract, owner, token, spender *types.Address) *Lockdown {
	return &Lockdown{
		Contract: contract,
		Owner:    owner,
		Token:    token,
		Spender:  spender,
	}
}

// UnorderedNonceInvalidation mirrors SignatureTransfer's UnorderedNonceInvalidation event.
type UnorderedNonceInvalidation struct {
	Contract *types.Address
	Owner    *types.Address
	Word     *big.Int
	Mask     *big.Int
}

func NewUnorderedNonceInvalidation(contract, owner *types.Address, word, mask *big.Int) *UnorderedNonceInvalidation {
	return &UnorderedNonceInvalidation{
		Contract: contract,
		Owner:    owner,
		Word:     word,
		Mask:     mask,
	}
}
