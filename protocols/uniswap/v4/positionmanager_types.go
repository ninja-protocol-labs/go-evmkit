package v4

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

// positionManagerDomainName is the name EIP712_v4's constructor is called with
// (ERC721Permit_v4("Uniswap v4 Positions NFT", "UNI-V4-POSM")); PositionManager's EIP-712 domain
// has no version field, unlike permit2's.
const positionManagerDomainName = "Uniswap v4 Positions NFT"

const (
	permitPrimaryType       = "Permit"
	permitForAllPrimaryType = "PermitForAll"
)

var (
	// permitTypes matches ERC721PermitHash.sol's PERMIT_TYPEHASH exactly:
	// keccak256("Permit(address spender,uint256 tokenId,uint256 nonce,uint256 deadline)").
	permitTypes = eip712.Types{
		permitPrimaryType: {
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("tokenId", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}

	// permitForAllTypes matches ERC721PermitHash.sol's PERMIT_FOR_ALL_TYPEHASH exactly:
	// keccak256("PermitForAll(address operator,bool approved,uint256 nonce,uint256 deadline)").
	permitForAllTypes = eip712.Types{
		permitForAllPrimaryType: {
			eip712.NewField("operator", abi.Address.String()),
			eip712.NewField("approved", abi.Bool.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
)

// NewPositionManagerDomain returns the EIP-712 domain a given PositionManager deployment signs
// permit/permitForAll under: name, chainID, and verifyingContract only -- no version, no salt.
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

// MintPositionFromDeltasParams mirrors MINT_POSITION_FROM_DELTAS's decoding. Deprecated
// upstream (sandwich-attack vulnerable); prefer MintPositionParams.
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

// PositionInfo mirrors libraries/PositionInfoLibrary.sol's packed PositionInfo: a truncated
// 25-byte PoolId plus a position's tick range and subscription flag, packed into one uint256.
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

// ModifyPosition mirrors IPositionManager's ModifyPosition event: emitted on every
// modifyLiquidity call, like PoolManager's ModifyLiquidity except Sender is the end user
// (the unlock locker), not the position manager itself.
type ModifyPosition struct {
	PositionManager *types.Address
	Id              PoolID
	Sender          *types.Address
	TickLower       *big.Int
	TickUpper       *big.Int
	LiquidityDelta  *big.Int
	Salt            [32]byte
}

func NewModifyPosition(positionManager *types.Address, id PoolID, sender *types.Address, tickLower, tickUpper, liquidityDelta *big.Int, salt [32]byte) *ModifyPosition {
	return &ModifyPosition{
		PositionManager: positionManager,
		Id:              id,
		Sender:          sender,
		TickLower:       tickLower,
		TickUpper:       tickUpper,
		LiquidityDelta:  liquidityDelta,
		Salt:            salt,
	}
}

// Subscription mirrors INotifier's Subscription event.
type Subscription struct {
	PositionManager *types.Address
	TokenID         *big.Int
	Subscriber      *types.Address
}

func NewSubscription(positionManager *types.Address, tokenID *big.Int, subscriber *types.Address) *Subscription {
	return &Subscription{
		PositionManager: positionManager,
		TokenID:         tokenID,
		Subscriber:      subscriber,
	}
}

// Unsubscription mirrors INotifier's Unsubscription event.
type Unsubscription struct {
	PositionManager *types.Address
	TokenID         *big.Int
	Subscriber      *types.Address
}

func NewUnsubscription(positionManager *types.Address, tokenID *big.Int, subscriber *types.Address) *Unsubscription {
	return &Unsubscription{
		PositionManager: positionManager,
		TokenID:         tokenID,
		Subscriber:      subscriber,
	}
}

// ERC721Permit is the EIP-712 message signed to approve Spender for a single position NFT via
// ERC721Permit_v4.permit() -- see ERC721PermitHash.sol's hashPermit.
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

// ERC721PermitForAll is the EIP-712 message signed to set Operator over all of Owner's position
// NFTs via ERC721Permit_v4.permitForAll() -- see ERC721PermitHash.sol's hashPermitForAll.
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
