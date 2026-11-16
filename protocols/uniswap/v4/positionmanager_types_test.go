package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewMintPositionParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	tickLower := big.NewInt(-887220)
	tickUpper := big.NewInt(887220)
	liquidity := big.NewInt(1000000)
	amount0Max := big.NewInt(1000)
	amount1Max := big.NewInt(1100)
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	hookData := []byte{0x05}

	p := NewMintPositionParams(poolKey, tickLower, tickUpper, liquidity, amount0Max, amount1Max, owner, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, liquidity, p.Liquidity)
	require.Same(t, amount0Max, p.Amount0Max)
	require.Same(t, amount1Max, p.Amount1Max)
	require.Same(t, owner, p.Owner)
	require.Equal(t, hookData, p.HookData)
}

func TestNewMintPositionFromDeltasParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	tickLower := big.NewInt(-887220)
	tickUpper := big.NewInt(887220)
	amount0Max := big.NewInt(1000)
	amount1Max := big.NewInt(1100)
	owner := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	hookData := []byte{0x06}

	p := NewMintPositionFromDeltasParams(poolKey, tickLower, tickUpper, amount0Max, amount1Max, owner, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, amount0Max, p.Amount0Max)
	require.Same(t, amount1Max, p.Amount1Max)
	require.Same(t, owner, p.Owner)
	require.Equal(t, hookData, p.HookData)
}

func TestNewPositionInfo(t *testing.T) {
	var poolID [25]byte
	copy(poolID[:], []byte{0x11, 0x22, 0x33})
	tickLower := big.NewInt(-887220)
	tickUpper := big.NewInt(887220)

	p := NewPositionInfo(poolID, tickLower, tickUpper, true)
	require.Equal(t, poolID, p.PoolID)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.True(t, p.HasSubscriber)
}

func TestERC721PermitTypeHashMatchesERC721PermitHashSol(t *testing.T) {
	want := "Permit(address spender,uint256 tokenId,uint256 nonce,uint256 deadline)"
	require.Equal(t, want, permitTypes.EncodeType(permitPrimaryType))
}

func TestERC721PermitForAllTypeHashMatchesERC721PermitHashSol(t *testing.T) {
	want := "PermitForAll(address operator,bool approved,uint256 nonce,uint256 deadline)"
	require.Equal(t, want, permitForAllTypes.EncodeType(permitForAllPrimaryType))
}

func TestERC721PermitHashDeterministic(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	p := NewERC721Permit(spender, big.NewInt(42), big.NewInt(1), big.NewInt(9999999999))
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	h1, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestERC721PermitHashSensitiveToEachField(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	other := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	base := NewERC721Permit(spender, big.NewInt(42), big.NewInt(1), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variants := []*ERC721Permit{
		NewERC721Permit(other, big.NewInt(42), big.NewInt(1), big.NewInt(9999999999)),
		NewERC721Permit(spender, big.NewInt(43), big.NewInt(1), big.NewInt(9999999999)),
		NewERC721Permit(spender, big.NewInt(42), big.NewInt(2), big.NewInt(9999999999)),
		NewERC721Permit(spender, big.NewInt(42), big.NewInt(1), big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestERC721PermitHashSensitiveToDomain(t *testing.T) {
	spender := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	other := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	p := NewERC721Permit(spender, big.NewInt(42), big.NewInt(1), big.NewInt(9999999999))

	h1, err := p.Hash(NewPositionManagerDomain(big.NewInt(1), contract))
	require.NoError(t, err)
	h2, err := p.Hash(NewPositionManagerDomain(big.NewInt(2), contract))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())

	h3, err := p.Hash(NewPositionManagerDomain(big.NewInt(1), other))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h3.Bytes())
}

func TestERC721PermitSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	spender := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	p := NewERC721Permit(spender, big.NewInt(42), big.NewInt(1), big.NewInt(9999999999))
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	sig, err := p.Sign(domain, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestERC721PermitForAllHashDeterministic(t *testing.T) {
	operator := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	p := NewERC721PermitForAll(operator, true, big.NewInt(1), big.NewInt(9999999999))
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	h1, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestERC721PermitForAllHashSensitiveToEachField(t *testing.T) {
	operator := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	other := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	base := NewERC721PermitForAll(operator, true, big.NewInt(1), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variants := []*ERC721PermitForAll{
		NewERC721PermitForAll(other, true, big.NewInt(1), big.NewInt(9999999999)),
		NewERC721PermitForAll(operator, false, big.NewInt(1), big.NewInt(9999999999)),
		NewERC721PermitForAll(operator, true, big.NewInt(2), big.NewInt(9999999999)),
		NewERC721PermitForAll(operator, true, big.NewInt(1), big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestERC721PermitForAllSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	operator := types.MustNewAddressFromHex("0x9a8ca723f5dccb7926d00b71dec55c2fea1f50f7")
	contract := types.MustNewAddressFromHex("0x4b2c77d209d3405f41a037ec6c77f7f5b8e2ca80")
	p := NewERC721PermitForAll(operator, true, big.NewInt(1), big.NewInt(9999999999))
	domain := NewPositionManagerDomain(big.NewInt(1), contract)

	sig, err := p.Sign(domain, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}
