package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
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

func TestPositionManagerRequestsValidate(t *testing.T) {
	pm := testPositionManager
	owner := testSender
	tokenID := big.NewInt(7)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	sig := []byte{0x01, 0x02}
	actions := []byte{ActionMintPosition, ActionSettlePair}
	params := [][]byte{{0x01}, {0x02}}
	deadline := big.NewInt(1_700_000_000)
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)
	details := permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3))
	single := permit2.NewPermitSingle(details, testHooks, big.NewInt(4))
	batch := permit2.NewPermitBatch([]permit2.PermitDetails{*details, *details}, testHooks, big.NewInt(4))

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"poolManager nil", (*PoolManagerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"poolManager empty", (&PoolManagerRequest{}).Validate(), "invalid request: contract is nil"},
		{"poolManager valid", (&PoolManagerRequest{Contract: pm}).Validate(), ""},
		{"weth9 nil", (*WETH9Request)(nil).Validate(), "invalid request: request is nil"},
		{"weth9 empty", (&WETH9Request{}).Validate(), "invalid request: contract is nil"},
		{"weth9 valid", (&WETH9Request{Contract: pm}).Validate(), ""},
		{"permit2 nil", (*Permit2Request)(nil).Validate(), "invalid request: request is nil"},
		{"permit2 empty", (&Permit2Request{}).Validate(), "invalid request: contract is nil"},
		{"permit2 valid", (&Permit2Request{Contract: pm}).Validate(), ""},
		{"nextTokenId nil", (*NextTokenIDRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nextTokenId empty", (&NextTokenIDRequest{}).Validate(), "invalid request: contract is nil"},
		{"nextTokenId valid", (&NextTokenIDRequest{Contract: pm}).Validate(), ""},
		{"getPositionLiquidity nil", (*GetPositionLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPositionLiquidity empty", (&GetPositionLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPositionLiquidity nil tokenID", (&GetPositionLiquidityRequest{Contract: pm}).Validate(), "invalid request: tokenID is nil"},
		{"getPositionLiquidity valid", (&GetPositionLiquidityRequest{Contract: pm, TokenID: tokenID}).Validate(), ""},
		{"getPoolAndPositionInfo nil", (*GetPoolAndPositionInfoRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPoolAndPositionInfo empty", (&GetPoolAndPositionInfoRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPoolAndPositionInfo nil tokenID", (&GetPoolAndPositionInfoRequest{Contract: pm}).Validate(), "invalid request: tokenID is nil"},
		{"getPoolAndPositionInfo valid", (&GetPoolAndPositionInfoRequest{Contract: pm, TokenID: tokenID}).Validate(), ""},
		{"positionInfo nil", (*PositionInfoRequest)(nil).Validate(), "invalid request: request is nil"},
		{"positionInfo empty", (&PositionInfoRequest{}).Validate(), "invalid request: contract is nil"},
		{"positionInfo nil tokenID", (&PositionInfoRequest{Contract: pm}).Validate(), "invalid request: tokenID is nil"},
		{"positionInfo valid", (&PositionInfoRequest{Contract: pm, TokenID: big.NewInt(0)}).Validate(), ""},
		{"poolKeys nil", (*PoolKeysRequest)(nil).Validate(), "invalid request: request is nil"},
		{"poolKeys empty", (&PoolKeysRequest{}).Validate(), "invalid request: contract is nil"},
		{"poolKeys zero poolId", (&PoolKeysRequest{Contract: pm}).Validate(), ""},
		{"subscriber nil", (*SubscriberRequest)(nil).Validate(), "invalid request: request is nil"},
		{"subscriber empty", (&SubscriberRequest{}).Validate(), "invalid request: contract is nil"},
		{"subscriber nil tokenID", (&SubscriberRequest{Contract: pm}).Validate(), "invalid request: tokenID is nil"},
		{"subscriber valid", (&SubscriberRequest{Contract: pm, TokenID: tokenID}).Validate(), ""},
		{"unsubscribeGasLimit nil", (*UnsubscribeGasLimitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"unsubscribeGasLimit empty", (&UnsubscribeGasLimitRequest{}).Validate(), "invalid request: contract is nil"},
		{"unsubscribeGasLimit valid", (&UnsubscribeGasLimitRequest{Contract: pm}).Validate(), ""},
		{"nonces nil", (*NoncesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nonces empty", (&NoncesRequest{}).Validate(), "invalid request: contract is nil"},
		{"nonces nil owner", (&NoncesRequest{Contract: pm, Word: big.NewInt(0)}).Validate(), "invalid request: owner is nil"},
		{"nonces nil word", (&NoncesRequest{Contract: pm, Owner: owner}).Validate(), "invalid request: word is nil"},
		{"nonces valid", (&NoncesRequest{Contract: pm, Owner: owner, Word: big.NewInt(0)}).Validate(), ""},
		{"modifyLiquidities nil", (*ModifyLiquiditiesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"modifyLiquidities empty", (&ModifyLiquiditiesRequest{}).Validate(), "invalid request: contract is nil"},
		{"modifyLiquidities empty actions", (&ModifyLiquiditiesRequest{Contract: pm, Deadline: deadline}).Validate(), "invalid request: actions is empty"},
		{"modifyLiquidities params mismatch", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params[:1], Deadline: deadline}).Validate(), "invalid request: params length 1 does not match actions length 2"},
		{"modifyLiquidities nil deadline", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params}).Validate(), "invalid request: deadline is nil"},
		{"modifyLiquidities negative native", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"modifyLiquidities overflow native", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline, NativeAmount: overflow}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"modifyLiquidities nil native", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline}).Validate(), ""},
		{"modifyLiquidities zero native", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline, NativeAmount: big.NewInt(0)}).Validate(), ""},
		{"modifyLiquidities max native", (&ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline, NativeAmount: new(big.Int).Sub(overflow, big.NewInt(1))}).Validate(), ""},
		{"modifyLiquiditiesWithoutUnlock nil", (*ModifyLiquiditiesWithoutUnlockRequest)(nil).Validate(), "invalid request: request is nil"},
		{"modifyLiquiditiesWithoutUnlock empty", (&ModifyLiquiditiesWithoutUnlockRequest{}).Validate(), "invalid request: contract is nil"},
		{"modifyLiquiditiesWithoutUnlock empty actions", (&ModifyLiquiditiesWithoutUnlockRequest{Contract: pm}).Validate(), "invalid request: actions is empty"},
		{"modifyLiquiditiesWithoutUnlock params mismatch", (&ModifyLiquiditiesWithoutUnlockRequest{Contract: pm, Actions: actions[:1], Params: params}).Validate(), "invalid request: params length 2 does not match actions length 1"},
		{"modifyLiquiditiesWithoutUnlock negative native", (&ModifyLiquiditiesWithoutUnlockRequest{Contract: pm, Actions: actions, Params: params, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"modifyLiquiditiesWithoutUnlock overflow native", (&ModifyLiquiditiesWithoutUnlockRequest{Contract: pm, Actions: actions, Params: params, NativeAmount: overflow}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"modifyLiquiditiesWithoutUnlock valid", (&ModifyLiquiditiesWithoutUnlockRequest{Contract: pm, Actions: actions, Params: params, NativeAmount: big.NewInt(1)}).Validate(), ""},
		{"initializePool nil", (*InitializePoolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"initializePool empty", (&InitializePoolRequest{}).Validate(), "invalid request: contract is nil"},
		{"initializePool nil key", (&InitializePoolRequest{Contract: pm, SqrtPriceX96: big.NewInt(1)}).Validate(), "invalid request: key is nil"},
		{"initializePool nil key.currency0", (&InitializePoolRequest{Contract: pm, Key: &PoolKey{Currency1: testCurrency1, Fee: big.NewInt(1), TickSpacing: big.NewInt(1), Hooks: testHooks}, SqrtPriceX96: big.NewInt(1)}).Validate(), "invalid request: key.currency0 is nil"},
		{"initializePool nil key.hooks", (&InitializePoolRequest{Contract: pm, Key: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, Fee: big.NewInt(1), TickSpacing: big.NewInt(1)}, SqrtPriceX96: big.NewInt(1)}).Validate(), "invalid request: key.hooks is nil"},
		{"initializePool nil sqrtPrice", (&InitializePoolRequest{Contract: pm, Key: key}).Validate(), "invalid request: sqrtPriceX96 is nil"},
		{"initializePool valid", (&InitializePoolRequest{Contract: pm, Key: key, SqrtPriceX96: big.NewInt(1)}).Validate(), ""},
		{"multicall nil", (*MulticallRequest)(nil).Validate(), "invalid request: request is nil"},
		{"multicall empty", (&MulticallRequest{}).Validate(), "invalid request: contract is nil"},
		{"multicall empty data", (&MulticallRequest{Contract: pm}).Validate(), "invalid request: data is empty"},
		{"multicall negative native", (&MulticallRequest{Contract: pm, Data: params, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"multicall overflow native", (&MulticallRequest{Contract: pm, Data: params, NativeAmount: overflow}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"multicall valid", (&MulticallRequest{Contract: pm, Data: params}).Validate(), ""},
		{"subscribe nil", (*SubscribeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"subscribe empty", (&SubscribeRequest{}).Validate(), "invalid request: contract is nil"},
		{"subscribe nil tokenID", (&SubscribeRequest{Contract: pm, Subscriber: testHooks}).Validate(), "invalid request: tokenID is nil"},
		{"subscribe nil subscriber", (&SubscribeRequest{Contract: pm, TokenID: tokenID}).Validate(), "invalid request: subscriber is nil"},
		{"subscribe nil data", (&SubscribeRequest{Contract: pm, TokenID: tokenID, Subscriber: testHooks}).Validate(), ""},
		{"unsubscribe nil", (*UnsubscribeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"unsubscribe empty", (&UnsubscribeRequest{}).Validate(), "invalid request: contract is nil"},
		{"unsubscribe nil tokenID", (&UnsubscribeRequest{Contract: pm}).Validate(), "invalid request: tokenID is nil"},
		{"unsubscribe valid", (&UnsubscribeRequest{Contract: pm, TokenID: tokenID}).Validate(), ""},
		{"revokeNonce nil", (*RevokeNonceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"revokeNonce empty", (&RevokeNonceRequest{}).Validate(), "invalid request: contract is nil"},
		{"revokeNonce nil nonce", (&RevokeNonceRequest{Contract: pm}).Validate(), "invalid request: nonce is nil"},
		{"revokeNonce valid", (&RevokeNonceRequest{Contract: pm, Nonce: big.NewInt(0)}).Validate(), ""},
		{"erc721Permit nil", (*ERC721PermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"erc721Permit empty", (&ERC721PermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"erc721Permit nil spender", (&ERC721PermitRequest{Contract: pm, TokenID: tokenID, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: spender is nil"},
		{"erc721Permit nil tokenID", (&ERC721PermitRequest{Contract: pm, Spender: owner, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: tokenID is nil"},
		{"erc721Permit nil deadline", (&ERC721PermitRequest{Contract: pm, Spender: owner, TokenID: tokenID, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: deadline is nil"},
		{"erc721Permit nil nonce", (&ERC721PermitRequest{Contract: pm, Spender: owner, TokenID: tokenID, Deadline: deadline, Signature: sig}).Validate(), "invalid request: nonce is nil"},
		{"erc721Permit empty signature", (&ERC721PermitRequest{Contract: pm, Spender: owner, TokenID: tokenID, Deadline: deadline, Nonce: big.NewInt(1), Signature: []byte{}}).Validate(), "invalid request: signature is empty"},
		{"erc721Permit valid", (&ERC721PermitRequest{Contract: pm, Spender: owner, TokenID: tokenID, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), ""},
		{"erc721PermitForAll nil", (*ERC721PermitForAllRequest)(nil).Validate(), "invalid request: request is nil"},
		{"erc721PermitForAll empty", (&ERC721PermitForAllRequest{}).Validate(), "invalid request: contract is nil"},
		{"erc721PermitForAll nil owner", (&ERC721PermitForAllRequest{Contract: pm, Operator: testHooks, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"erc721PermitForAll nil operator", (&ERC721PermitForAllRequest{Contract: pm, Owner: owner, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: operator is nil"},
		{"erc721PermitForAll nil deadline", (&ERC721PermitForAllRequest{Contract: pm, Owner: owner, Operator: testHooks, Nonce: big.NewInt(1), Signature: sig}).Validate(), "invalid request: deadline is nil"},
		{"erc721PermitForAll nil nonce", (&ERC721PermitForAllRequest{Contract: pm, Owner: owner, Operator: testHooks, Deadline: deadline, Signature: sig}).Validate(), "invalid request: nonce is nil"},
		{"erc721PermitForAll nil signature", (&ERC721PermitForAllRequest{Contract: pm, Owner: owner, Operator: testHooks, Deadline: deadline, Nonce: big.NewInt(1)}).Validate(), "invalid request: signature is empty"},
		{"erc721PermitForAll revoke valid", (&ERC721PermitForAllRequest{Contract: pm, Owner: owner, Operator: testHooks, Approved: false, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}).Validate(), ""},
		{"permit2Permit nil", (*Permit2PermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permit2Permit empty", (&Permit2PermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"permit2Permit nil owner", (&Permit2PermitRequest{Contract: pm, PermitSingle: single, Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permit2Permit nil permitSingle", (&Permit2PermitRequest{Contract: pm, Owner: owner, Signature: sig}).Validate(), "invalid request: permitSingle is nil"},
		{"permit2Permit nil details", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(nil, testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.details is nil"},
		{"permit2Permit nil details.token", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(nil, big.NewInt(1), big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.details.token is nil"},
		{"permit2Permit nil details.amount", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, nil, big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.details.amount is nil"},
		{"permit2Permit nil details.expiration", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), nil, big.NewInt(3)), testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.details.expiration is nil"},
		{"permit2Permit nil details.nonce", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), nil), testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.details.nonce is nil"},
		{"permit2Permit nil spender", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(details, nil, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitSingle.spender is nil"},
		{"permit2Permit nil sigDeadline", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: permit2.NewPermitSingle(details, testHooks, nil), Signature: sig}).Validate(), "invalid request: permitSingle.sigDeadline is nil"},
		{"permit2Permit empty signature", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: single}).Validate(), "invalid request: signature is empty"},
		{"permit2Permit valid", (&Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: single, Signature: sig}).Validate(), ""},
		{"permit2PermitBatch nil", (*Permit2PermitBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permit2PermitBatch empty", (&Permit2PermitBatchRequest{}).Validate(), "invalid request: contract is nil"},
		{"permit2PermitBatch nil owner", (&Permit2PermitBatchRequest{Contract: pm, PermitBatch: batch, Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permit2PermitBatch nil permitBatch", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, Signature: sig}).Validate(), "invalid request: permitBatch is nil"},
		{"permit2PermitBatch empty details", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: permit2.NewPermitBatch(nil, testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitBatch.details is empty"},
		{"permit2PermitBatch nil details[1].token", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details, {Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitBatch.details[1].token is nil"},
		{"permit2PermitBatch nil details[0].nonce", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{{Token: testCurrency0, Amount: big.NewInt(1), Expiration: big.NewInt(2)}}, testHooks, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitBatch.details[0].nonce is nil"},
		{"permit2PermitBatch nil spender", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details}, nil, big.NewInt(4)), Signature: sig}).Validate(), "invalid request: permitBatch.spender is nil"},
		{"permit2PermitBatch nil sigDeadline", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details}, testHooks, nil), Signature: sig}).Validate(), "invalid request: permitBatch.sigDeadline is nil"},
		{"permit2PermitBatch empty signature", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: batch, Signature: []byte{}}).Validate(), "invalid request: signature is empty"},
		{"permit2PermitBatch valid", (&Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: batch, Signature: sig}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewPositionManagerRequestsSetFields(t *testing.T) {
	pm := testPositionManager
	owner := testSender
	tokenID := big.NewInt(9)
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(10), testHooks)
	actions := []byte{ActionDecreaseLiquidity, ActionTakePair}
	params := [][]byte{{0x0a}, {0x0b}}
	deadline := big.NewInt(1_700_000_000)
	native := big.NewInt(42)
	sig := []byte{0xaa, 0xbb}
	single := permit2.NewPermitSingle(permit2.NewPermitDetails(testCurrency0, big.NewInt(1), big.NewInt(2), big.NewInt(3)), testHooks, big.NewInt(4))
	batch := permit2.NewPermitBatch([]permit2.PermitDetails{*permit2.NewPermitDetails(testCurrency1, big.NewInt(5), big.NewInt(6), big.NewInt(7))}, testHooks, big.NewInt(8))
	poolID := [25]byte{0x01, 0x02}

	require.Equal(t, &WETH9Request{Contract: pm}, NewWETH9Request(pm))
	require.Equal(t, &Permit2Request{Contract: pm}, NewPermit2Request(pm))
	require.Equal(t, &NextTokenIDRequest{Contract: pm}, NewNextTokenIDRequest(pm))
	require.Equal(t, &GetPositionLiquidityRequest{Contract: pm, TokenID: tokenID}, NewGetPositionLiquidityRequest(pm, tokenID))
	require.Equal(t, &GetPoolAndPositionInfoRequest{Contract: pm, TokenID: tokenID}, NewGetPoolAndPositionInfoRequest(pm, tokenID))
	require.Equal(t, &PositionInfoRequest{Contract: pm, TokenID: tokenID}, NewPositionInfoRequest(pm, tokenID))
	require.Equal(t, &PoolKeysRequest{Contract: pm, PoolID: poolID}, NewPoolKeysRequest(pm, poolID))
	require.Equal(t, &SubscriberRequest{Contract: pm, TokenID: tokenID}, NewSubscriberRequest(pm, tokenID))
	require.Equal(t, &UnsubscribeGasLimitRequest{Contract: pm}, NewUnsubscribeGasLimitRequest(pm))
	require.Equal(t, &NoncesRequest{Contract: pm, Owner: owner, Word: big.NewInt(3)}, NewNoncesRequest(pm, owner, big.NewInt(3)))
	require.Equal(t, &ModifyLiquiditiesRequest{Contract: pm, Actions: actions, Params: params, Deadline: deadline, NativeAmount: native}, NewModifyLiquiditiesRequest(pm, actions, params, deadline, native))
	require.Equal(t, &ModifyLiquiditiesWithoutUnlockRequest{Contract: pm, Actions: actions, Params: params, NativeAmount: native}, NewModifyLiquiditiesWithoutUnlockRequest(pm, actions, params, native))
	require.Equal(t, &InitializePoolRequest{Contract: pm, Key: key, SqrtPriceX96: big.NewInt(1)}, NewInitializePoolRequest(pm, key, big.NewInt(1)))
	require.Equal(t, &MulticallRequest{Contract: pm, Data: params, NativeAmount: native}, NewMulticallRequest(pm, params, native))
	require.Equal(t, &SubscribeRequest{Contract: pm, TokenID: tokenID, Subscriber: testHooks, Data: []byte{0x01}}, NewSubscribeRequest(pm, tokenID, testHooks, []byte{0x01}))
	require.Equal(t, &UnsubscribeRequest{Contract: pm, TokenID: tokenID}, NewUnsubscribeRequest(pm, tokenID))
	require.Equal(t, &RevokeNonceRequest{Contract: pm, Nonce: big.NewInt(5)}, NewRevokeNonceRequest(pm, big.NewInt(5)))
	require.Equal(t, &ERC721PermitRequest{Contract: pm, Spender: owner, TokenID: tokenID, Deadline: deadline, Nonce: big.NewInt(1), Signature: sig}, NewERC721PermitRequest(pm, owner, tokenID, deadline, big.NewInt(1), sig))
	require.Equal(t, &ERC721PermitForAllRequest{Contract: pm, Owner: owner, Operator: testHooks, Approved: true, Deadline: deadline, Nonce: big.NewInt(2), Signature: sig}, NewERC721PermitForAllRequest(pm, owner, testHooks, true, deadline, big.NewInt(2), sig))
	require.Equal(t, &Permit2PermitRequest{Contract: pm, Owner: owner, PermitSingle: single, Signature: sig}, NewPermit2PermitRequest(pm, owner, single, sig))
	require.Equal(t, &Permit2PermitBatchRequest{Contract: pm, Owner: owner, PermitBatch: batch, Signature: sig}, NewPermit2PermitBatchRequest(pm, owner, batch, sig))

	req := NewModifyLiquiditiesRequest(pm, actions, params, deadline, native)
	require.Same(t, native, req.NativeAmount)
	require.Same(t, deadline, req.Deadline)
	require.Same(t, single, NewPermit2PermitRequest(pm, owner, single, sig).PermitSingle)
}

func TestNewPositionManagerEventsSetFields(t *testing.T) {
	pm := testPositionManager
	id := PoolID{0x01}
	salt := [32]byte{0x02}

	require.Equal(t, &ModifyPositionEvent{Contract: pm, Id: id, Sender: testSender, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), LiquidityDelta: big.NewInt(-5), Salt: salt}, NewModifyPositionEvent(pm, id, testSender, big.NewInt(-60), big.NewInt(60), big.NewInt(-5), salt))
	require.Equal(t, &SubscriptionEvent{Contract: pm, TokenID: big.NewInt(1), Subscriber: testHooks}, NewSubscriptionEvent(pm, big.NewInt(1), testHooks))
	require.Equal(t, &UnsubscriptionEvent{Contract: pm, TokenID: big.NewInt(2), Subscriber: testHooks}, NewUnsubscriptionEvent(pm, big.NewInt(2), testHooks))
}
