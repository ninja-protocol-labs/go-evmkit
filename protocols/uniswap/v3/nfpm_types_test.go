package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPosition(t *testing.T) {
	nonce := big.NewInt(1)
	operator := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	token0 := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	token1 := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	fee := big.NewInt(2)
	tickLower := big.NewInt(3)
	tickUpper := big.NewInt(4)
	liquidity := big.NewInt(5)
	feeGrowthInside0LastX128 := big.NewInt(6)
	feeGrowthInside1LastX128 := big.NewInt(7)
	tokensOwed0 := big.NewInt(8)
	tokensOwed1 := big.NewInt(9)

	p := NewPosition(nonce, operator, token0, token1, fee, tickLower, tickUpper, liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128, tokensOwed0, tokensOwed1)
	require.Same(t, nonce, p.Nonce)
	require.Same(t, operator, p.Operator)
	require.Same(t, token0, p.Token0)
	require.Same(t, token1, p.Token1)
	require.Same(t, fee, p.Fee)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, liquidity, p.Liquidity)
	require.Same(t, feeGrowthInside0LastX128, p.FeeGrowthInside0LastX128)
	require.Same(t, feeGrowthInside1LastX128, p.FeeGrowthInside1LastX128)
	require.Same(t, tokensOwed0, p.TokensOwed0)
	require.Same(t, tokensOwed1, p.TokensOwed1)
}

func TestNewMintParams(t *testing.T) {
	token0 := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	token1 := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	fee := big.NewInt(1)
	tickLower := big.NewInt(2)
	tickUpper := big.NewInt(3)
	amount0Desired := big.NewInt(4)
	amount1Desired := big.NewInt(5)
	amount0Min := big.NewInt(6)
	amount1Min := big.NewInt(7)
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	deadline := big.NewInt(8)

	p := NewMintParams(token0, token1, fee, tickLower, tickUpper, amount0Desired, amount1Desired, amount0Min, amount1Min, recipient, deadline)
	require.Same(t, token0, p.Token0)
	require.Same(t, token1, p.Token1)
	require.Same(t, fee, p.Fee)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, amount0Desired, p.Amount0Desired)
	require.Same(t, amount1Desired, p.Amount1Desired)
	require.Same(t, amount0Min, p.Amount0Min)
	require.Same(t, amount1Min, p.Amount1Min)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, deadline, p.Deadline)
}

func TestNewIncreaseLiquidityParams(t *testing.T) {
	tokenId := big.NewInt(1)
	amount0Desired := big.NewInt(2)
	amount1Desired := big.NewInt(3)
	amount0Min := big.NewInt(4)
	amount1Min := big.NewInt(5)
	deadline := big.NewInt(6)

	p := NewIncreaseLiquidityParams(tokenId, amount0Desired, amount1Desired, amount0Min, amount1Min, deadline)
	require.Same(t, tokenId, p.TokenId)
	require.Same(t, amount0Desired, p.Amount0Desired)
	require.Same(t, amount1Desired, p.Amount1Desired)
	require.Same(t, amount0Min, p.Amount0Min)
	require.Same(t, amount1Min, p.Amount1Min)
	require.Same(t, deadline, p.Deadline)
}

func TestNewDecreaseLiquidityParams(t *testing.T) {
	tokenId := big.NewInt(1)
	liquidity := big.NewInt(2)
	amount0Min := big.NewInt(3)
	amount1Min := big.NewInt(4)
	deadline := big.NewInt(5)

	p := NewDecreaseLiquidityParams(tokenId, liquidity, amount0Min, amount1Min, deadline)
	require.Same(t, tokenId, p.TokenId)
	require.Same(t, liquidity, p.Liquidity)
	require.Same(t, amount0Min, p.Amount0Min)
	require.Same(t, amount1Min, p.Amount1Min)
	require.Same(t, deadline, p.Deadline)
}

func TestNewCollectParams(t *testing.T) {
	tokenId := big.NewInt(1)
	recipient := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	amount0Max := big.NewInt(2)
	amount1Max := big.NewInt(3)

	p := NewCollectParams(tokenId, recipient, amount0Max, amount1Max)
	require.Same(t, tokenId, p.TokenId)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amount0Max, p.Amount0Max)
	require.Same(t, amount1Max, p.Amount1Max)
}

func TestNewIncreaseLiquidityEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenId := big.NewInt(1)
	liquidity := big.NewInt(2)
	amount0 := big.NewInt(3)
	amount1 := big.NewInt(4)

	e := NewIncreaseLiquidity(contract, tokenId, liquidity, amount0, amount1)
	require.Same(t, contract, e.Contract)
	require.Same(t, tokenId, e.TokenId)
	require.Same(t, liquidity, e.Liquidity)
	require.Same(t, amount0, e.Amount0)
	require.Same(t, amount1, e.Amount1)
}

func TestNewDecreaseLiquidityEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenId := big.NewInt(1)
	liquidity := big.NewInt(2)
	amount0 := big.NewInt(3)
	amount1 := big.NewInt(4)

	e := NewDecreaseLiquidity(contract, tokenId, liquidity, amount0, amount1)
	require.Same(t, contract, e.Contract)
	require.Same(t, tokenId, e.TokenId)
	require.Same(t, liquidity, e.Liquidity)
	require.Same(t, amount0, e.Amount0)
	require.Same(t, amount1, e.Amount1)
}

func TestNewPositionCollect(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	recipient := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	tokenId := big.NewInt(1)
	amount0 := big.NewInt(2)
	amount1 := big.NewInt(3)

	e := NewPositionCollect(contract, recipient, tokenId, amount0, amount1)
	require.Same(t, contract, e.Contract)
	require.Same(t, recipient, e.Recipient)
	require.Same(t, tokenId, e.TokenId)
	require.Same(t, amount0, e.Amount0)
	require.Same(t, amount1, e.Amount1)
}

func TestNewNFTPermit(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenId := big.NewInt(1)
	nonce := big.NewInt(2)
	deadline := big.NewInt(3)

	p := NewNFTPermit(spender, tokenId, nonce, deadline)
	require.Same(t, spender, p.Spender)
	require.Same(t, tokenId, p.TokenId)
	require.Same(t, nonce, p.Nonce)
	require.Same(t, deadline, p.Deadline)
}

func TestNewNFTPermitDomain(t *testing.T) {
	chainID := big.NewInt(1)
	verifyingContract := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")

	domain := NewNFTPermitDomain(chainID, verifyingContract)
	require.Equal(t, "Uniswap V3 Positions NFT-V1", domain.Name)
	require.Equal(t, "1", domain.Version)
	require.Same(t, chainID, domain.ChainID)
	require.Same(t, verifyingContract, domain.VerifyingContract)
	require.Nil(t, domain.Salt)
}

func TestNFTPermitTypeHashMatchesERC721PermitSol(t *testing.T) {
	want := types.MustNewHashFromHex("0x49ecf333e5b8c95c40fdafc95c1ad136e8914a8fb55e9dc8bb01eaa83a2df9ad")
	got := nftPermitTypes.TypeHash(nftPermitPrimaryType)
	require.Equal(t, want.Bytes(), got.Bytes())
}

func TestNFTPermitHashDeterministic(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	verifyingContract := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	domain := NewNFTPermitDomain(big.NewInt(1), verifyingContract)
	p := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))

	h1, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestNFTPermitHashSensitiveToEachField(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	spenderB := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	verifyingContract := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	domain := NewNFTPermitDomain(big.NewInt(1), verifyingContract)

	base := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variants := []*NFTPermit{
		NewNFTPermit(spenderB, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999)),
		NewNFTPermit(spender, big.NewInt(2), big.NewInt(0), big.NewInt(9999999999)),
		NewNFTPermit(spender, big.NewInt(1), big.NewInt(1), big.NewInt(9999999999)),
		NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestNFTPermitHashSensitiveToDomain(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	verifyingContract := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	verifyingContractB := types.MustNewAddressFromHex("0x4444444444444444444444444444444444444444")
	p := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))

	h1, err := p.Hash(NewNFTPermitDomain(big.NewInt(1), verifyingContract))
	require.NoError(t, err)
	h2, err := p.Hash(NewNFTPermitDomain(big.NewInt(2), verifyingContract))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())

	h3, err := p.Hash(NewNFTPermitDomain(big.NewInt(1), verifyingContractB))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h3.Bytes())
}

func TestNFTPermitHashRejectsNilDomain(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	p := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))

	_, err := p.Hash(nil)
	require.Error(t, err)
}

func TestNFTPermitSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	verifyingContract := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	domain := NewNFTPermitDomain(big.NewInt(1), verifyingContract)
	p := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))

	sig, err := p.Sign(domain, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNFTPermitSignRejectsNilDomain(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)

	spender := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	p := NewNFTPermit(spender, big.NewInt(1), big.NewInt(0), big.NewInt(9999999999))

	_, err = p.Sign(nil, key)
	require.Error(t, err)
}
