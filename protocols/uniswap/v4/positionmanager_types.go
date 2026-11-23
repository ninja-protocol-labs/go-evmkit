package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

// positionManagerDomainName is the EIP-712 name ERC721Permit_v4 is constructed with; the domain has no version.
const positionManagerDomainName = "Uniswap v4 Positions NFT"

const (
	permitPrimaryType       = "Permit"
	permitForAllPrimaryType = "PermitForAll"
)

var (
	// permitTypes matches ERC721PermitHash.sol's PERMIT_TYPEHASH.
	permitTypes = eip712.Types{
		permitPrimaryType: {
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("tokenId", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}

	// permitForAllTypes matches ERC721PermitHash.sol's PERMIT_FOR_ALL_TYPEHASH.
	permitForAllTypes = eip712.Types{
		permitForAllPrimaryType: {
			eip712.NewField("operator", abi.Address.String()),
			eip712.NewField("approved", abi.Bool.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
)

// NewPositionManagerDomain returns the EIP-712 domain (name, chainID, verifyingContract) PositionManager permits are signed under.
func NewPositionManagerDomain(chainID *big.Int, verifyingContract *types.Address) *eip712.Domain {
	return eip712.NewDomain(positionManagerDomainName, "", chainID, verifyingContract, nil)
}

// MintPositionParams mirrors MINT_POSITION's decoding (CalldataDecoder.decodeMintParams).
type MintPositionParams struct {
	PoolKey    *PoolKey
	TickLower  *big.Int
	TickUpper  *big.Int
	Liquidity  *big.Int
	Amount0Max *big.Int
	Amount1Max *big.Int
	Owner      *types.Address
	HookData   []byte
}

func NewMintPositionParams(poolKey *PoolKey, tickLower, tickUpper, liquidity, amount0Max, amount1Max *big.Int, owner *types.Address, hookData []byte) *MintPositionParams {
	return &MintPositionParams{
		PoolKey:    poolKey,
		TickLower:  tickLower,
		TickUpper:  tickUpper,
		Liquidity:  liquidity,
		Amount0Max: amount0Max,
		Amount1Max: amount1Max,
		Owner:      owner,
		HookData:   hookData,
	}
}

// MintPositionFromDeltasParams mirrors MINT_POSITION_FROM_DELTAS's decoding; deprecated upstream, prefer MintPositionParams.
type MintPositionFromDeltasParams struct {
	PoolKey    *PoolKey
	TickLower  *big.Int
	TickUpper  *big.Int
	Amount0Max *big.Int
	Amount1Max *big.Int
	Owner      *types.Address
	HookData   []byte
}

func NewMintPositionFromDeltasParams(poolKey *PoolKey, tickLower, tickUpper, amount0Max, amount1Max *big.Int, owner *types.Address, hookData []byte) *MintPositionFromDeltasParams {
	return &MintPositionFromDeltasParams{
		PoolKey:    poolKey,
		TickLower:  tickLower,
		TickUpper:  tickUpper,
		Amount0Max: amount0Max,
		Amount1Max: amount1Max,
		Owner:      owner,
		HookData:   hookData,
	}
}

// PositionInfo mirrors PositionInfoLibrary's packed word: truncated 25-byte PoolId, tick range and subscription flag.
type PositionInfo struct {
	PoolID        [25]byte
	TickLower     *big.Int
	TickUpper     *big.Int
	HasSubscriber bool
}

func NewPositionInfo(poolID [25]byte, tickLower, tickUpper *big.Int, hasSubscriber bool) *PositionInfo {
	return &PositionInfo{
		PoolID:        poolID,
		TickLower:     tickLower,
		TickUpper:     tickUpper,
		HasSubscriber: hasSubscriber,
	}
}

// ModifyPositionEvent mirrors IPositionManager's ModifyPosition event; Sender is the unlock locker, not the position manager.
type ModifyPositionEvent struct {
	Contract       *types.Address
	Id             PoolID
	Sender         *types.Address
	TickLower      *big.Int
	TickUpper      *big.Int
	LiquidityDelta *big.Int
	Salt           [32]byte
}

func NewModifyPositionEvent(contract *types.Address, id PoolID, sender *types.Address, tickLower, tickUpper, liquidityDelta *big.Int, salt [32]byte) *ModifyPositionEvent {
	return &ModifyPositionEvent{
		Contract:       contract,
		Id:             id,
		Sender:         sender,
		TickLower:      tickLower,
		TickUpper:      tickUpper,
		LiquidityDelta: liquidityDelta,
		Salt:           salt,
	}
}

// SubscriptionEvent mirrors INotifier's Subscription event.
type SubscriptionEvent struct {
	Contract   *types.Address
	TokenID    *big.Int
	Subscriber *types.Address
}

func NewSubscriptionEvent(contract *types.Address, tokenID *big.Int, subscriber *types.Address) *SubscriptionEvent {
	return &SubscriptionEvent{
		Contract:   contract,
		TokenID:    tokenID,
		Subscriber: subscriber,
	}
}

// UnsubscriptionEvent mirrors INotifier's Unsubscription event.
type UnsubscriptionEvent struct {
	Contract   *types.Address
	TokenID    *big.Int
	Subscriber *types.Address
}

func NewUnsubscriptionEvent(contract *types.Address, tokenID *big.Int, subscriber *types.Address) *UnsubscriptionEvent {
	return &UnsubscriptionEvent{
		Contract:   contract,
		TokenID:    tokenID,
		Subscriber: subscriber,
	}
}

// ERC721Permit is the EIP-712 message signed to approve Spender for one position NFT via permit().
type ERC721Permit struct {
	Spender  *types.Address
	TokenID  *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

func NewERC721Permit(spender *types.Address, tokenID, nonce, deadline *big.Int) *ERC721Permit {
	return &ERC721Permit{
		Spender:  spender,
		TokenID:  tokenID,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *ERC721Permit) Hash(domain *eip712.Domain) (*types.Hash, error) {
	message := eip712.Message{
		"spender":  p.Spender,
		"tokenId":  p.TokenID,
		"nonce":    p.Nonce,
		"deadline": p.Deadline,
	}
	h, err := eip712.Hash(domain, permitPrimaryType, permitTypes, message)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721Permit: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain.
func (p *ERC721Permit) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}
	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721Permit: sign: %w", err)
	}
	return sig, nil
}

// ERC721PermitForAll is the EIP-712 message signed to set Operator over all of the signer's position NFTs via permitForAll().
type ERC721PermitForAll struct {
	Operator *types.Address
	Approved bool
	Nonce    *big.Int
	Deadline *big.Int
}

func NewERC721PermitForAll(operator *types.Address, approved bool, nonce, deadline *big.Int) *ERC721PermitForAll {
	return &ERC721PermitForAll{
		Operator: operator,
		Approved: approved,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *ERC721PermitForAll) Hash(domain *eip712.Domain) (*types.Hash, error) {
	message := eip712.Message{
		"operator": p.Operator,
		"approved": p.Approved,
		"nonce":    p.Nonce,
		"deadline": p.Deadline,
	}
	h, err := eip712.Hash(domain, permitForAllPrimaryType, permitForAllTypes, message)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721PermitForAll: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain.
func (p *ERC721PermitForAll) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}
	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: erc721PermitForAll: sign: %w", err)
	}
	return sig, nil
}

// WETH9Request is WETH9()'s input on Contract.
type WETH9Request struct {
	Contract *types.Address
}

func NewWETH9Request(contract *types.Address) *WETH9Request {
	return &WETH9Request{
		Contract: contract,
	}
}

func (r *WETH9Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// Permit2Request is permit2()'s input on Contract.
type Permit2Request struct {
	Contract *types.Address
}

func NewPermit2Request(contract *types.Address) *Permit2Request {
	return &Permit2Request{
		Contract: contract,
	}
}

func (r *Permit2Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NextTokenIDRequest is nextTokenId()'s input on Contract.
type NextTokenIDRequest struct {
	Contract *types.Address
}

func NewNextTokenIDRequest(contract *types.Address) *NextTokenIDRequest {
	return &NextTokenIDRequest{
		Contract: contract,
	}
}

func (r *NextTokenIDRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPositionLiquidityRequest is getPositionLiquidity(tokenId)'s input on Contract.
type GetPositionLiquidityRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewGetPositionLiquidityRequest(contract *types.Address, tokenID *big.Int) *GetPositionLiquidityRequest {
	return &GetPositionLiquidityRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *GetPositionLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetPoolAndPositionInfoRequest is getPoolAndPositionInfo(tokenId)'s input on Contract.
type GetPoolAndPositionInfoRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewGetPoolAndPositionInfoRequest(contract *types.Address, tokenID *big.Int) *GetPoolAndPositionInfoRequest {
	return &GetPoolAndPositionInfoRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *GetPoolAndPositionInfoRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PositionInfoRequest is positionInfo(tokenId)'s input on Contract.
type PositionInfoRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewPositionInfoRequest(contract *types.Address, tokenID *big.Int) *PositionInfoRequest {
	return &PositionInfoRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *PositionInfoRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PoolKeysRequest is poolKeys(poolId)'s input on Contract; PoolID is PositionInfo's truncated 25-byte pool ID.
type PoolKeysRequest struct {
	Contract *types.Address
	PoolID   [25]byte
}

func NewPoolKeysRequest(contract *types.Address, poolID [25]byte) *PoolKeysRequest {
	return &PoolKeysRequest{
		Contract: contract,
		PoolID:   poolID,
	}
}

func (r *PoolKeysRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// SubscriberRequest is subscriber(tokenId)'s input on Contract.
type SubscriberRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewSubscriberRequest(contract *types.Address, tokenID *big.Int) *SubscriberRequest {
	return &SubscriberRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *SubscriberRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// UnsubscribeGasLimitRequest is unsubscribeGasLimit()'s input on Contract.
type UnsubscribeGasLimitRequest struct {
	Contract *types.Address
}

func NewUnsubscribeGasLimitRequest(contract *types.Address) *UnsubscribeGasLimitRequest {
	return &UnsubscribeGasLimitRequest{
		Contract: contract,
	}
}

func (r *UnsubscribeGasLimitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NoncesRequest is nonces(owner, word)'s input on Contract.
type NoncesRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Word     *big.Int
}

func NewNoncesRequest(contract, owner *types.Address, word *big.Int) *NoncesRequest {
	return &NoncesRequest{
		Contract: contract,
		Owner:    owner,
		Word:     word,
	}
}

func (r *NoncesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Word == nil {
		return fmt.Errorf("%w: word is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ModifyLiquiditiesRequest is modifyLiquidities(abi.encode(actions, params), deadline)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ModifyLiquiditiesRequest struct {
	Contract     *types.Address
	Actions      []byte
	Params       [][]byte
	Deadline     *big.Int
	NativeAmount *big.Int
}

func NewModifyLiquiditiesRequest(contract *types.Address, actions []byte, params [][]byte, deadline, nativeAmount *big.Int) *ModifyLiquiditiesRequest {
	return &ModifyLiquiditiesRequest{
		Contract:     contract,
		Actions:      actions,
		Params:       params,
		Deadline:     deadline,
		NativeAmount: nativeAmount,
	}
}

func (r *ModifyLiquiditiesRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Actions) == 0 {
		return fmt.Errorf("%w: actions is empty", bind.ErrInvalidRequest)
	}

	if len(r.Params) != len(r.Actions) {
		return fmt.Errorf("%w: params length %d does not match actions length %d", bind.ErrInvalidRequest, len(r.Params), len(r.Actions))
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

// ModifyLiquiditiesWithoutUnlockRequest is modifyLiquiditiesWithoutUnlock(actions, params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type ModifyLiquiditiesWithoutUnlockRequest struct {
	Contract     *types.Address
	Actions      []byte
	Params       [][]byte
	NativeAmount *big.Int
}

func NewModifyLiquiditiesWithoutUnlockRequest(contract *types.Address, actions []byte, params [][]byte, nativeAmount *big.Int) *ModifyLiquiditiesWithoutUnlockRequest {
	return &ModifyLiquiditiesWithoutUnlockRequest{
		Contract:     contract,
		Actions:      actions,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *ModifyLiquiditiesWithoutUnlockRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Actions) == 0 {
		return fmt.Errorf("%w: actions is empty", bind.ErrInvalidRequest)
	}

	if len(r.Params) != len(r.Actions) {
		return fmt.Errorf("%w: params length %d does not match actions length %d", bind.ErrInvalidRequest, len(r.Params), len(r.Actions))
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// InitializePoolRequest is initializePool(key, sqrtPriceX96)'s input on Contract.
type InitializePoolRequest struct {
	Contract     *types.Address
	Key          *PoolKey
	SqrtPriceX96 *big.Int
}

func NewInitializePoolRequest(contract *types.Address, key *PoolKey, sqrtPriceX96 *big.Int) *InitializePoolRequest {
	return &InitializePoolRequest{
		Contract:     contract,
		Key:          key,
		SqrtPriceX96: sqrtPriceX96,
	}
}

func (r *InitializePoolRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validatePoolKey(r.Key); err != nil {
		return err
	}

	if r.SqrtPriceX96 == nil {
		return fmt.Errorf("%w: sqrtPriceX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// MulticallRequest is multicall(data)'s input on Contract; NativeAmount, if set, is sent as msg.value and shared by every sub-call.
type MulticallRequest struct {
	Contract     *types.Address
	Data         [][]byte
	NativeAmount *big.Int
}

func NewMulticallRequest(contract *types.Address, data [][]byte, nativeAmount *big.Int) *MulticallRequest {
	return &MulticallRequest{
		Contract:     contract,
		Data:         data,
		NativeAmount: nativeAmount,
	}
}

func (r *MulticallRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Data) == 0 {
		return fmt.Errorf("%w: data is empty", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.Sign() < 0 {
		return fmt.Errorf("%w: nativeAmount is negative", bind.ErrInvalidRequest)
	}

	if r.NativeAmount != nil && r.NativeAmount.BitLen() > 256 {
		return fmt.Errorf("%w: nativeAmount exceeds uint256", bind.ErrInvalidRequest)
	}

	return nil
}

// SubscribeRequest is subscribe(tokenId, newSubscriber, data)'s input on Contract.
type SubscribeRequest struct {
	Contract   *types.Address
	TokenID    *big.Int
	Subscriber *types.Address
	Data       []byte
}

func NewSubscribeRequest(contract *types.Address, tokenID *big.Int, subscriber *types.Address, data []byte) *SubscribeRequest {
	return &SubscribeRequest{
		Contract:   contract,
		TokenID:    tokenID,
		Subscriber: subscriber,
		Data:       data,
	}
}

func (r *SubscribeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	if r.Subscriber == nil {
		return fmt.Errorf("%w: subscriber is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// UnsubscribeRequest is unsubscribe(tokenId)'s input on Contract.
type UnsubscribeRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewUnsubscribeRequest(contract *types.Address, tokenID *big.Int) *UnsubscribeRequest {
	return &UnsubscribeRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *UnsubscribeRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RevokeNonceRequest is revokeNonce(nonce)'s input on Contract.
type RevokeNonceRequest struct {
	Contract *types.Address
	Nonce    *big.Int
}

func NewRevokeNonceRequest(contract *types.Address, nonce *big.Int) *RevokeNonceRequest {
	return &RevokeNonceRequest{
		Contract: contract,
		Nonce:    nonce,
	}
}

func (r *RevokeNonceRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Nonce == nil {
		return fmt.Errorf("%w: nonce is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// ERC721PermitRequest is permit(spender, tokenId, deadline, nonce, signature)'s input on Contract.
type ERC721PermitRequest struct {
	Contract  *types.Address
	Spender   *types.Address
	TokenID   *big.Int
	Deadline  *big.Int
	Nonce     *big.Int
	Signature []byte
}

func NewERC721PermitRequest(contract, spender *types.Address, tokenID, deadline, nonce *big.Int, signature []byte) *ERC721PermitRequest {
	return &ERC721PermitRequest{
		Contract:  contract,
		Spender:   spender,
		TokenID:   tokenID,
		Deadline:  deadline,
		Nonce:     nonce,
		Signature: signature,
	}
}

func (r *ERC721PermitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Spender == nil {
		return fmt.Errorf("%w: spender is nil", bind.ErrInvalidRequest)
	}

	if r.TokenID == nil {
		return fmt.Errorf("%w: tokenID is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Nonce == nil {
		return fmt.Errorf("%w: nonce is nil", bind.ErrInvalidRequest)
	}

	if len(r.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// ERC721PermitForAllRequest is permitForAll(owner, operator, approved, deadline, nonce, signature)'s input on Contract.
type ERC721PermitForAllRequest struct {
	Contract  *types.Address
	Owner     *types.Address
	Operator  *types.Address
	Approved  bool
	Deadline  *big.Int
	Nonce     *big.Int
	Signature []byte
}

func NewERC721PermitForAllRequest(contract, owner, operator *types.Address, approved bool, deadline, nonce *big.Int, signature []byte) *ERC721PermitForAllRequest {
	return &ERC721PermitForAllRequest{
		Contract:  contract,
		Owner:     owner,
		Operator:  operator,
		Approved:  approved,
		Deadline:  deadline,
		Nonce:     nonce,
		Signature: signature,
	}
}

func (r *ERC721PermitForAllRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Operator == nil {
		return fmt.Errorf("%w: operator is nil", bind.ErrInvalidRequest)
	}

	if r.Deadline == nil {
		return fmt.Errorf("%w: deadline is nil", bind.ErrInvalidRequest)
	}

	if r.Nonce == nil {
		return fmt.Errorf("%w: nonce is nil", bind.ErrInvalidRequest)
	}

	if len(r.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// Permit2PermitRequest is Permit2Forwarder.permit(owner, permitSingle, signature)'s input on Contract.
type Permit2PermitRequest struct {
	Contract     *types.Address
	Owner        *types.Address
	PermitSingle *permit2.PermitSingle
	Signature    []byte
}

func NewPermit2PermitRequest(contract, owner *types.Address, permitSingle *permit2.PermitSingle, signature []byte) *Permit2PermitRequest {
	return &Permit2PermitRequest{
		Contract:     contract,
		Owner:        owner,
		PermitSingle: permitSingle,
		Signature:    signature,
	}
}

func (r *Permit2PermitRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.PermitSingle == nil {
		return fmt.Errorf("%w: permitSingle is nil", bind.ErrInvalidRequest)
	}

	if r.PermitSingle.Details == nil {
		return fmt.Errorf("%w: permitSingle.details is nil", bind.ErrInvalidRequest)
	}

	if err := validatePermitDetails("permitSingle.details", r.PermitSingle.Details); err != nil {
		return err
	}

	if r.PermitSingle.Spender == nil {
		return fmt.Errorf("%w: permitSingle.spender is nil", bind.ErrInvalidRequest)
	}

	if r.PermitSingle.SigDeadline == nil {
		return fmt.Errorf("%w: permitSingle.sigDeadline is nil", bind.ErrInvalidRequest)
	}

	if len(r.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// Permit2PermitBatchRequest is Permit2Forwarder.permitBatch(owner, permitBatch, signature)'s input on Contract.
type Permit2PermitBatchRequest struct {
	Contract    *types.Address
	Owner       *types.Address
	PermitBatch *permit2.PermitBatch
	Signature   []byte
}

func NewPermit2PermitBatchRequest(contract, owner *types.Address, permitBatch *permit2.PermitBatch, signature []byte) *Permit2PermitBatchRequest {
	return &Permit2PermitBatchRequest{
		Contract:    contract,
		Owner:       owner,
		PermitBatch: permitBatch,
		Signature:   signature,
	}
}

func (r *Permit2PermitBatchRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.PermitBatch == nil {
		return fmt.Errorf("%w: permitBatch is nil", bind.ErrInvalidRequest)
	}

	if len(r.PermitBatch.Details) == 0 {
		return fmt.Errorf("%w: permitBatch.details is empty", bind.ErrInvalidRequest)
	}

	for i := range r.PermitBatch.Details {
		if err := validatePermitDetails(fmt.Sprintf("permitBatch.details[%d]", i), &r.PermitBatch.Details[i]); err != nil {
			return err
		}
	}

	if r.PermitBatch.Spender == nil {
		return fmt.Errorf("%w: permitBatch.spender is nil", bind.ErrInvalidRequest)
	}

	if r.PermitBatch.SigDeadline == nil {
		return fmt.Errorf("%w: permitBatch.sigDeadline is nil", bind.ErrInvalidRequest)
	}

	if len(r.Signature) == 0 {
		return fmt.Errorf("%w: signature is empty", bind.ErrInvalidRequest)
	}

	return nil
}

func validatePermitDetails(path string, d *permit2.PermitDetails) error {
	if d.Token == nil {
		return fmt.Errorf("%w: %s.token is nil", bind.ErrInvalidRequest, path)
	}

	if d.Amount == nil {
		return fmt.Errorf("%w: %s.amount is nil", bind.ErrInvalidRequest, path)
	}

	if d.Expiration == nil {
		return fmt.Errorf("%w: %s.expiration is nil", bind.ErrInvalidRequest, path)
	}

	if d.Nonce == nil {
		return fmt.Errorf("%w: %s.nonce is nil", bind.ErrInvalidRequest, path)
	}

	return nil
}
