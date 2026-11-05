package v3

import (
	"fmt"
	"math/big"

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

// IncreaseLiquidity mirrors INonfungiblePositionManager's IncreaseLiquidity event.
type IncreaseLiquidity struct {
	Contract  *types.Address
	TokenId   *big.Int
	Liquidity *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewIncreaseLiquidity(contract *types.Address, tokenId, liquidity, amount0, amount1 *big.Int) *IncreaseLiquidity {
	return &IncreaseLiquidity{
		Contract:  contract,
		TokenId:   tokenId,
		Liquidity: liquidity,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// DecreaseLiquidity mirrors INonfungiblePositionManager's DecreaseLiquidity event.
type DecreaseLiquidity struct {
	Contract  *types.Address
	TokenId   *big.Int
	Liquidity *big.Int
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewDecreaseLiquidity(contract *types.Address, tokenId, liquidity, amount0, amount1 *big.Int) *DecreaseLiquidity {
	return &DecreaseLiquidity{
		Contract:  contract,
		TokenId:   tokenId,
		Liquidity: liquidity,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}

// PositionCollect mirrors INonfungiblePositionManager's Collect event (named to avoid colliding
// with Pool's own Collect event type).
type PositionCollect struct {
	Contract  *types.Address
	TokenId   *big.Int
	Recipient *types.Address
	Amount0   *big.Int
	Amount1   *big.Int
}

func NewPositionCollect(contract, recipient *types.Address, tokenId, amount0, amount1 *big.Int) *PositionCollect {
	return &PositionCollect{
		Contract:  contract,
		TokenId:   tokenId,
		Recipient: recipient,
		Amount0:   amount0,
		Amount1:   amount1,
	}
}
