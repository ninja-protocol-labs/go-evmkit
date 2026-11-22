package v3

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

const (
	nftPermitDomainName        = "Uniswap V3 Positions NFT-V1"
	nftPermitDomainVersion     = "1"
	nftPermitPrimaryType       = "Permit"
	nftPermitSpenderFieldName  = "spender"
	nftPermitTokenIdFieldName  = "tokenId"
	nftPermitNonceFieldName    = "nonce"
	nftPermitDeadlineFieldName = "deadline"
)

// nftPermitTypes is the EIP-712 schema for a position NFT's permit, matching ERC721Permit.sol's
// PERMIT_TYPEHASH exactly.
var nftPermitTypes = eip712.Types{
	nftPermitPrimaryType: {
		eip712.NewField(nftPermitSpenderFieldName, abi.Address.String()),
		eip712.NewField(nftPermitTokenIdFieldName, abi.Uint256.String()),
		eip712.NewField(nftPermitNonceFieldName, abi.Uint256.String()),
		eip712.NewField(nftPermitDeadlineFieldName, abi.Uint256.String()),
	},
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

// TotalSupplyRequest is totalSupply()'s input on Contract.
type TotalSupplyRequest struct {
	Contract *types.Address
}

func NewTotalSupplyRequest(contract *types.Address) *TotalSupplyRequest {
	return &TotalSupplyRequest{
		Contract: contract,
	}
}

func (r *TotalSupplyRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TokenOfOwnerByIndexRequest is tokenOfOwnerByIndex(owner, index)'s input on Contract.
type TokenOfOwnerByIndexRequest struct {
	Contract *types.Address
	Owner    *types.Address
	Index    *big.Int
}

func NewTokenOfOwnerByIndexRequest(contract, owner *types.Address, index *big.Int) *TokenOfOwnerByIndexRequest {
	return &TokenOfOwnerByIndexRequest{
		Contract: contract,
		Owner:    owner,
		Index:    index,
	}
}

func (r *TokenOfOwnerByIndexRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Owner == nil {
		return fmt.Errorf("%w: owner is nil", bind.ErrInvalidRequest)
	}

	if r.Index == nil {
		return fmt.Errorf("%w: index is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// TokenByIndexRequest is tokenByIndex(index)'s input on Contract.
type TokenByIndexRequest struct {
	Contract *types.Address
	Index    *big.Int
}

func NewTokenByIndexRequest(contract *types.Address, index *big.Int) *TokenByIndexRequest {
	return &TokenByIndexRequest{
		Contract: contract,
		Index:    index,
	}
}

func (r *TokenByIndexRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Index == nil {
		return fmt.Errorf("%w: index is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// PermitTypehashRequest is PERMIT_TYPEHASH()'s input on Contract.
type PermitTypehashRequest struct {
	Contract *types.Address
}

func NewPermitTypehashRequest(contract *types.Address) *PermitTypehashRequest {
	return &PermitTypehashRequest{
		Contract: contract,
	}
}

func (r *PermitTypehashRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// DomainSeparatorRequest is DOMAIN_SEPARATOR()'s input on Contract.
type DomainSeparatorRequest struct {
	Contract *types.Address
}

func NewDomainSeparatorRequest(contract *types.Address) *DomainSeparatorRequest {
	return &DomainSeparatorRequest{
		Contract: contract,
	}
}

func (r *DomainSeparatorRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMPositionsRequest is positions(tokenId)'s input on Contract.
type NFPMPositionsRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewNFPMPositionsRequest(contract *types.Address, tokenID *big.Int) *NFPMPositionsRequest {
	return &NFPMPositionsRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *NFPMPositionsRequest) Validate() error {
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

// PermitRequest is permit(spender, tokenId, deadline, v, r, s)'s input on Contract.
type PermitRequest struct {
	Contract *types.Address
	Spender  *types.Address
	TokenID  *big.Int
	Deadline *big.Int
	Sig      *types.Signature
}

func NewPermitRequest(contract, spender *types.Address, tokenID, deadline *big.Int, sig *types.Signature) *PermitRequest {
	return &PermitRequest{
		Contract: contract,
		Spender:  spender,
		TokenID:  tokenID,
		Deadline: deadline,
		Sig:      sig,
	}
}

func (r *PermitRequest) Validate() error {
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

	if r.Sig == nil {
		return fmt.Errorf("%w: sig is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// CreateAndInitializePoolIfNecessaryRequest is createAndInitializePoolIfNecessary(token0, token1, fee, sqrtPriceX96)'s input on Contract.
type CreateAndInitializePoolIfNecessaryRequest struct {
	Contract     *types.Address
	Token0       *types.Address
	Token1       *types.Address
	Fee          *big.Int
	SqrtPriceX96 *big.Int
}

func NewCreateAndInitializePoolIfNecessaryRequest(contract, token0, token1 *types.Address, fee, sqrtPriceX96 *big.Int) *CreateAndInitializePoolIfNecessaryRequest {
	return &CreateAndInitializePoolIfNecessaryRequest{
		Contract:     contract,
		Token0:       token0,
		Token1:       token1,
		Fee:          fee,
		SqrtPriceX96: sqrtPriceX96,
	}
}

func (r *CreateAndInitializePoolIfNecessaryRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token0 == nil {
		return fmt.Errorf("%w: token0 is nil", bind.ErrInvalidRequest)
	}

	if r.Token1 == nil {
		return fmt.Errorf("%w: token1 is nil", bind.ErrInvalidRequest)
	}

	if r.Fee == nil {
		return fmt.Errorf("%w: fee is nil", bind.ErrInvalidRequest)
	}

	if r.SqrtPriceX96 == nil {
		return fmt.Errorf("%w: sqrtPriceX96 is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMUnwrapWETH9Request is unwrapWETH9(amountMinimum, recipient)'s input on Contract.
type NFPMUnwrapWETH9Request struct {
	Contract      *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewNFPMUnwrapWETH9Request(contract *types.Address, amountMinimum *big.Int, recipient *types.Address) *NFPMUnwrapWETH9Request {
	return &NFPMUnwrapWETH9Request{
		Contract:      contract,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

func (r *NFPMUnwrapWETH9Request) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMSweepTokenRequest is sweepToken(token, amountMinimum, recipient)'s input on Contract.
type NFPMSweepTokenRequest struct {
	Contract      *types.Address
	Token         *types.Address
	AmountMinimum *big.Int
	Recipient     *types.Address
}

func NewNFPMSweepTokenRequest(contract, token *types.Address, amountMinimum *big.Int, recipient *types.Address) *NFPMSweepTokenRequest {
	return &NFPMSweepTokenRequest{
		Contract:      contract,
		Token:         token,
		AmountMinimum: amountMinimum,
		Recipient:     recipient,
	}
}

func (r *NFPMSweepTokenRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.Token == nil {
		return fmt.Errorf("%w: token is nil", bind.ErrInvalidRequest)
	}

	if r.AmountMinimum == nil {
		return fmt.Errorf("%w: amountMinimum is nil", bind.ErrInvalidRequest)
	}

	if r.Recipient == nil {
		return fmt.Errorf("%w: recipient is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMRefundETHRequest is refundETH()'s input on Contract.
type NFPMRefundETHRequest struct {
	Contract *types.Address
}

func NewNFPMRefundETHRequest(contract *types.Address) *NFPMRefundETHRequest {
	return &NFPMRefundETHRequest{
		Contract: contract,
	}
}

func (r *NFPMRefundETHRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMMulticallRequest is multicall(data)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type NFPMMulticallRequest struct {
	Contract     *types.Address
	Data         [][]byte
	NativeAmount *big.Int
}

func NewNFPMMulticallRequest(contract *types.Address, data [][]byte, nativeAmount *big.Int) *NFPMMulticallRequest {
	return &NFPMMulticallRequest{
		Contract:     contract,
		Data:         data,
		NativeAmount: nativeAmount,
	}
}

func (r *NFPMMulticallRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if len(r.Data) == 0 {
		return fmt.Errorf("%w: data is empty", bind.ErrInvalidRequest)
	}

	return nil
}

// NFPMMintRequest is mint(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type NFPMMintRequest struct {
	Contract     *types.Address
	Params       *MintParams
	NativeAmount *big.Int
}

func NewNFPMMintRequest(contract *types.Address, params *MintParams, nativeAmount *big.Int) *NFPMMintRequest {
	return &NFPMMintRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *NFPMMintRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateMintParams(r.Params); err != nil {
		return err
	}

	return nil
}

// IncreaseLiquidityRequest is increaseLiquidity(params)'s input on Contract; NativeAmount, if set, is sent as msg.value.
type IncreaseLiquidityRequest struct {
	Contract     *types.Address
	Params       *IncreaseLiquidityParams
	NativeAmount *big.Int
}

func NewIncreaseLiquidityRequest(contract *types.Address, params *IncreaseLiquidityParams, nativeAmount *big.Int) *IncreaseLiquidityRequest {
	return &IncreaseLiquidityRequest{
		Contract:     contract,
		Params:       params,
		NativeAmount: nativeAmount,
	}
}

func (r *IncreaseLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateIncreaseLiquidityParams(r.Params); err != nil {
		return err
	}

	return nil
}

// DecreaseLiquidityRequest is decreaseLiquidity(params)'s input on Contract.
type DecreaseLiquidityRequest struct {
	Contract *types.Address
	Params   *DecreaseLiquidityParams
}

func NewDecreaseLiquidityRequest(contract *types.Address, params *DecreaseLiquidityParams) *DecreaseLiquidityRequest {
	return &DecreaseLiquidityRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *DecreaseLiquidityRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateDecreaseLiquidityParams(r.Params); err != nil {
		return err
	}

	return nil
}

// NFPMCollectRequest is collect(params)'s input on Contract.
type NFPMCollectRequest struct {
	Contract *types.Address
	Params   *CollectParams
}

func NewNFPMCollectRequest(contract *types.Address, params *CollectParams) *NFPMCollectRequest {
	return &NFPMCollectRequest{
		Contract: contract,
		Params:   params,
	}
}

func (r *NFPMCollectRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if err := validateCollectParams(r.Params); err != nil {
		return err
	}

	return nil
}

// NFPMBurnRequest is burn(tokenId)'s input on Contract.
type NFPMBurnRequest struct {
	Contract *types.Address
	TokenID  *big.Int
}

func NewNFPMBurnRequest(contract *types.Address, tokenID *big.Int) *NFPMBurnRequest {
	return &NFPMBurnRequest{
		Contract: contract,
		TokenID:  tokenID,
	}
}

func (r *NFPMBurnRequest) Validate() error {
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

func validateMintParams(p *MintParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.Token0 == nil {
		return fmt.Errorf("%w: params.token0 is nil", bind.ErrInvalidRequest)
	}

	if p.Token1 == nil {
		return fmt.Errorf("%w: params.token1 is nil", bind.ErrInvalidRequest)
	}

	if p.Fee == nil {
		return fmt.Errorf("%w: params.fee is nil", bind.ErrInvalidRequest)
	}

	if p.TickLower == nil {
		return fmt.Errorf("%w: params.tickLower is nil", bind.ErrInvalidRequest)
	}

	if p.TickUpper == nil {
		return fmt.Errorf("%w: params.tickUpper is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Desired == nil {
		return fmt.Errorf("%w: params.amount0Desired is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Desired == nil {
		return fmt.Errorf("%w: params.amount1Desired is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Min == nil {
		return fmt.Errorf("%w: params.amount0Min is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Min == nil {
		return fmt.Errorf("%w: params.amount1Min is nil", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.Deadline == nil {
		return fmt.Errorf("%w: params.deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateIncreaseLiquidityParams(p *IncreaseLiquidityParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenId == nil {
		return fmt.Errorf("%w: params.tokenId is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Desired == nil {
		return fmt.Errorf("%w: params.amount0Desired is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Desired == nil {
		return fmt.Errorf("%w: params.amount1Desired is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Min == nil {
		return fmt.Errorf("%w: params.amount0Min is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Min == nil {
		return fmt.Errorf("%w: params.amount1Min is nil", bind.ErrInvalidRequest)
	}

	if p.Deadline == nil {
		return fmt.Errorf("%w: params.deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateDecreaseLiquidityParams(p *DecreaseLiquidityParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenId == nil {
		return fmt.Errorf("%w: params.tokenId is nil", bind.ErrInvalidRequest)
	}

	if p.Liquidity == nil {
		return fmt.Errorf("%w: params.liquidity is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Min == nil {
		return fmt.Errorf("%w: params.amount0Min is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Min == nil {
		return fmt.Errorf("%w: params.amount1Min is nil", bind.ErrInvalidRequest)
	}

	if p.Deadline == nil {
		return fmt.Errorf("%w: params.deadline is nil", bind.ErrInvalidRequest)
	}

	return nil
}

func validateCollectParams(p *CollectParams) error {
	if p == nil {
		return fmt.Errorf("%w: params is nil", bind.ErrInvalidRequest)
	}

	if p.TokenId == nil {
		return fmt.Errorf("%w: params.tokenId is nil", bind.ErrInvalidRequest)
	}

	if p.Recipient == nil {
		return fmt.Errorf("%w: params.recipient is nil", bind.ErrInvalidRequest)
	}

	if p.Amount0Max == nil {
		return fmt.Errorf("%w: params.amount0Max is nil", bind.ErrInvalidRequest)
	}

	if p.Amount1Max == nil {
		return fmt.Errorf("%w: params.amount1Max is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// NewNFTPermitDomain returns the EIP-712 domain a NonfungiblePositionManager deployment signs
// permits under: name, version, chainID, and verifyingContract -- no salt.
func NewNFTPermitDomain(chainID *big.Int, verifyingContract *types.Address) *eip712.Domain {
	return eip712.NewDomain(nftPermitDomainName, nftPermitDomainVersion, chainID, verifyingContract, nil)
}

// NFTPermit is the EIP-712 message signed to approve spender for tokenId via permit().
type NFTPermit struct {
	Spender  *types.Address
	TokenId  *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

func NewNFTPermit(spender *types.Address, tokenId, nonce, deadline *big.Int) *NFTPermit {
	return &NFTPermit{
		Spender:  spender,
		TokenId:  tokenId,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Hash returns the signable EIP-712 digest for p under domain.
func (p *NFTPermit) Hash(domain *eip712.Domain) (*types.Hash, error) {
	message := eip712.Message{
		nftPermitSpenderFieldName:  p.Spender,
		nftPermitTokenIdFieldName:  p.TokenId,
		nftPermitNonceFieldName:    p.Nonce,
		nftPermitDeadlineFieldName: p.Deadline,
	}

	h, err := eip712.Hash(domain, nftPermitPrimaryType, nftPermitTypes, message)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: hash: %w", err)
	}
	return h, nil
}

// Sign returns signer's signature over p's EIP-712 digest under domain.
func (p *NFTPermit) Sign(domain *eip712.Domain, signer core.Signer) (*types.Signature, error) {
	digest, err := p.Hash(domain)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("uniswapv3: sign: %w", err)
	}
	return sig, nil
}

// Position mirrors INonfungiblePositionManager's positions(tokenId) tuple.
type Position struct {
	Nonce                    *big.Int
	Operator                 *types.Address
	Token0                   *types.Address
	Token1                   *types.Address
	Fee                      *big.Int
	TickLower                *big.Int
	TickUpper                *big.Int
	Liquidity                *big.Int
	FeeGrowthInside0LastX128 *big.Int
	FeeGrowthInside1LastX128 *big.Int
	TokensOwed0              *big.Int
	TokensOwed1              *big.Int
}

func NewPosition(
	nonce *big.Int,
	operator, token0, token1 *types.Address,
	fee, tickLower, tickUpper, liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1 *big.Int,
) *Position {
	return &Position{
		Nonce:                    nonce,
		Operator:                 operator,
		Token0:                   token0,
		Token1:                   token1,
		Fee:                      fee,
		TickLower:                tickLower,
		TickUpper:                tickUpper,
		Liquidity:                liquidity,
		FeeGrowthInside0LastX128: feeGrowthInside0LastX128,
		FeeGrowthInside1LastX128: feeGrowthInside1LastX128,
		TokensOwed0:              tokensOwed0,
		TokensOwed1:              tokensOwed1,
	}
}

// MintParams mirrors INonfungiblePositionManager.MintParams.
type MintParams struct {
	Token0         *types.Address
	Token1         *types.Address
	Fee            *big.Int
	TickLower      *big.Int
	TickUpper      *big.Int
	Amount0Desired *big.Int
	Amount1Desired *big.Int
	Amount0Min     *big.Int
	Amount1Min     *big.Int
	Recipient      *types.Address
	Deadline       *big.Int
}

func NewMintParams(
	token0, token1 *types.Address,
	fee, tickLower, tickUpper, amount0Desired, amount1Desired, amount0Min, amount1Min *big.Int,
	recipient *types.Address,
	deadline *big.Int,
) *MintParams {
	return &MintParams{
		Token0:         token0,
		Token1:         token1,
		Fee:            fee,
		TickLower:      tickLower,
		TickUpper:      tickUpper,
		Amount0Desired: amount0Desired,
		Amount1Desired: amount1Desired,
		Amount0Min:     amount0Min,
		Amount1Min:     amount1Min,
		Recipient:      recipient,
		Deadline:       deadline,
	}
}

// IncreaseLiquidityParams mirrors INonfungiblePositionManager.IncreaseLiquidityParams.
type IncreaseLiquidityParams struct {
	TokenId        *big.Int
	Amount0Desired *big.Int
	Amount1Desired *big.Int
	Amount0Min     *big.Int
	Amount1Min     *big.Int
	Deadline       *big.Int
}

func NewIncreaseLiquidityParams(tokenId, amount0Desired, amount1Desired, amount0Min, amount1Min, deadline *big.Int) *IncreaseLiquidityParams {
	return &IncreaseLiquidityParams{
		TokenId:        tokenId,
		Amount0Desired: amount0Desired,
		Amount1Desired: amount1Desired,
		Amount0Min:     amount0Min,
		Amount1Min:     amount1Min,
		Deadline:       deadline,
	}
}

// DecreaseLiquidityParams mirrors INonfungiblePositionManager.DecreaseLiquidityParams.
type DecreaseLiquidityParams struct {
	TokenId    *big.Int
	Liquidity  *big.Int
	Amount0Min *big.Int
	Amount1Min *big.Int
	Deadline   *big.Int
}

func NewDecreaseLiquidityParams(tokenId, liquidity, amount0Min, amount1Min, deadline *big.Int) *DecreaseLiquidityParams {
	return &DecreaseLiquidityParams{
		TokenId:    tokenId,
		Liquidity:  liquidity,
		Amount0Min: amount0Min,
		Amount1Min: amount1Min,
		Deadline:   deadline,
	}
}

// CollectParams mirrors INonfungiblePositionManager.CollectParams.
type CollectParams struct {
	TokenId    *big.Int
	Recipient  *types.Address
	Amount0Max *big.Int
	Amount1Max *big.Int
}

func NewCollectParams(tokenId *big.Int, recipient *types.Address, amount0Max, amount1Max *big.Int) *CollectParams {
	return &CollectParams{
		TokenId:    tokenId,
		Recipient:  recipient,
		Amount0Max: amount0Max,
		Amount1Max: amount1Max,
	}
}

// IncreaseLiquidityEvent mirrors INonfungiblePositionManager's IncreaseLiquidity event.
type IncreaseLiquidityEvent struct {
	Contract  *types.Address
	TokenId   *big.Int
	Liquidity *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewIncreaseLiquidityEvent(contract *types.Address, tokenId, liquidity, amount0, amount1 *big.Int) *IncreaseLiquidityEvent {
	return &IncreaseLiquidityEvent{
		Contract:  contract,
		TokenId:   tokenId,
		Liquidity: liquidity,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// DecreaseLiquidityEvent mirrors INonfungiblePositionManager's DecreaseLiquidity event.
type DecreaseLiquidityEvent struct {
	Contract  *types.Address
	TokenId   *big.Int
	Liquidity *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewDecreaseLiquidityEvent(contract *types.Address, tokenId, liquidity, amount0, amount1 *big.Int) *DecreaseLiquidityEvent {
	return &DecreaseLiquidityEvent{
		Contract:  contract,
		TokenId:   tokenId,
		Liquidity: liquidity,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// PositionCollectEvent mirrors INonfungiblePositionManager's Collect event (named to avoid colliding
// with Pool's own Collect event type).
type PositionCollectEvent struct {
	Contract  *types.Address
	TokenId   *big.Int
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewPositionCollectEvent(contract, recipient *types.Address, tokenId, amount0, amount1 *big.Int) *PositionCollectEvent {
	return &PositionCollectEvent{
		Contract:  contract,
		TokenId:   tokenId,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}
