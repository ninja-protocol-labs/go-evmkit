package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
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

	e := NewIncreaseLiquidityEvent(contract, tokenId, liquidity, amount0, amount1)
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

	e := NewDecreaseLiquidityEvent(contract, tokenId, liquidity, amount0, amount1)
	require.Same(t, contract, e.Contract)
	require.Same(t, tokenId, e.TokenId)
	require.Same(t, liquidity, e.Liquidity)
	require.Same(t, amount0, e.Amount0)
	require.Same(t, amount1, e.Amount1)
}

func TestNewPositionCollectEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	recipient := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	tokenId := big.NewInt(1)
	amount0 := big.NewInt(2)
	amount1 := big.NewInt(3)

	e := NewPositionCollectEvent(contract, recipient, tokenId, amount0, amount1)
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

func TestNFPMRequestsValidate(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	mintParams := NewMintParams(a, a, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1), big.NewInt(2), big.NewInt(0), big.NewInt(0), a, big.NewInt(1_700_000_000))
	increaseParams := NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1), big.NewInt(2), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	decreaseParams := NewDecreaseLiquidityParams(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	collectParams := NewCollectParams(big.NewInt(7), a, big.NewInt(5), big.NewInt(6))

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"wETH9 nil", (*WETH9Request)(nil).Validate(), "invalid request: request is nil"},
		{"wETH9 empty", (&WETH9Request{}).Validate(), "invalid request: contract is nil"},
		{"wETH9 nil contract", (&WETH9Request{}).Validate(), "invalid request: contract is nil"},
		{"wETH9 valid", (&WETH9Request{Contract: nfpm}).Validate(), ""},
		{"totalSupply nil", (*TotalSupplyRequest)(nil).Validate(), "invalid request: request is nil"},
		{"totalSupply empty", (&TotalSupplyRequest{}).Validate(), "invalid request: contract is nil"},
		{"totalSupply nil contract", (&TotalSupplyRequest{}).Validate(), "invalid request: contract is nil"},
		{"totalSupply valid", (&TotalSupplyRequest{Contract: nfpm}).Validate(), ""},
		{"tokenOfOwnerByIndex nil", (*TokenOfOwnerByIndexRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenOfOwnerByIndex empty", (&TokenOfOwnerByIndexRequest{}).Validate(), "invalid request: contract is nil"},
		{"tokenOfOwnerByIndex nil contract", (&TokenOfOwnerByIndexRequest{Owner: a, Index: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"tokenOfOwnerByIndex nil owner", (&TokenOfOwnerByIndexRequest{Contract: nfpm, Index: big.NewInt(21)}).Validate(), "invalid request: owner is nil"},
		{"tokenOfOwnerByIndex nil index", (&TokenOfOwnerByIndexRequest{Contract: nfpm, Owner: a}).Validate(), "invalid request: index is nil"},
		{"tokenOfOwnerByIndex valid", (&TokenOfOwnerByIndexRequest{Contract: nfpm, Owner: a, Index: big.NewInt(21)}).Validate(), ""},
		{"tokenByIndex nil", (*TokenByIndexRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenByIndex empty", (&TokenByIndexRequest{}).Validate(), "invalid request: contract is nil"},
		{"tokenByIndex nil contract", (&TokenByIndexRequest{Index: big.NewInt(11)}).Validate(), "invalid request: contract is nil"},
		{"tokenByIndex nil index", (&TokenByIndexRequest{Contract: nfpm}).Validate(), "invalid request: index is nil"},
		{"tokenByIndex valid", (&TokenByIndexRequest{Contract: nfpm, Index: big.NewInt(11)}).Validate(), ""},
		{"permitTypehash nil", (*PermitTypehashRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitTypehash empty", (&PermitTypehashRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitTypehash nil contract", (&PermitTypehashRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitTypehash valid", (&PermitTypehashRequest{Contract: nfpm}).Validate(), ""},
		{"domainSeparator nil", (*DomainSeparatorRequest)(nil).Validate(), "invalid request: request is nil"},
		{"domainSeparator empty", (&DomainSeparatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"domainSeparator nil contract", (&DomainSeparatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"domainSeparator valid", (&DomainSeparatorRequest{Contract: nfpm}).Validate(), ""},
		{"nFPMPositions nil", (*NFPMPositionsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMPositions empty", (&NFPMPositionsRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMPositions nil contract", (&NFPMPositionsRequest{TokenID: big.NewInt(11)}).Validate(), "invalid request: contract is nil"},
		{"nFPMPositions nil tokenID", (&NFPMPositionsRequest{Contract: nfpm}).Validate(), "invalid request: tokenID is nil"},
		{"nFPMPositions valid", (&NFPMPositionsRequest{Contract: nfpm, TokenID: big.NewInt(11)}).Validate(), ""},
		{"permit nil", (*PermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permit empty", (&PermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"permit nil contract", (&PermitRequest{Spender: a, TokenID: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"permit nil spender", (&PermitRequest{Contract: nfpm, TokenID: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: spender is nil"},
		{"permit nil tokenID", (&PermitRequest{Contract: nfpm, Spender: a, Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: tokenID is nil"},
		{"permit nil deadline", (&PermitRequest{Contract: nfpm, Spender: a, TokenID: big.NewInt(21), Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"permit nil sig", (&PermitRequest{Contract: nfpm, Spender: a, TokenID: big.NewInt(21), Deadline: big.NewInt(31)}).Validate(), "invalid request: sig is nil"},
		{"permit valid", (&PermitRequest{Contract: nfpm, Spender: a, TokenID: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), ""},
		{"createAndInitializePoolIfNecessary nil", (*CreateAndInitializePoolIfNecessaryRequest)(nil).Validate(), "invalid request: request is nil"},
		{"createAndInitializePoolIfNecessary empty", (&CreateAndInitializePoolIfNecessaryRequest{}).Validate(), "invalid request: contract is nil"},
		{"createAndInitializePoolIfNecessary nil contract", (&CreateAndInitializePoolIfNecessaryRequest{Token0: a, Token1: a, Fee: big.NewInt(31), SqrtPriceX96: big.NewInt(41)}).Validate(), "invalid request: contract is nil"},
		{"createAndInitializePoolIfNecessary nil token0", (&CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token1: a, Fee: big.NewInt(31), SqrtPriceX96: big.NewInt(41)}).Validate(), "invalid request: token0 is nil"},
		{"createAndInitializePoolIfNecessary nil token1", (&CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token0: a, Fee: big.NewInt(31), SqrtPriceX96: big.NewInt(41)}).Validate(), "invalid request: token1 is nil"},
		{"createAndInitializePoolIfNecessary nil fee", (&CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token0: a, Token1: a, SqrtPriceX96: big.NewInt(41)}).Validate(), "invalid request: fee is nil"},
		{"createAndInitializePoolIfNecessary nil sqrtPriceX96", (&CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token0: a, Token1: a, Fee: big.NewInt(31)}).Validate(), "invalid request: sqrtPriceX96 is nil"},
		{"createAndInitializePoolIfNecessary valid", (&CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token0: a, Token1: a, Fee: big.NewInt(31), SqrtPriceX96: big.NewInt(41)}).Validate(), ""},
		{"nFPMUnwrapWETH9 nil", (*NFPMUnwrapWETH9Request)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMUnwrapWETH9 empty", (&NFPMUnwrapWETH9Request{}).Validate(), "invalid request: contract is nil"},
		{"nFPMUnwrapWETH9 nil contract", (&NFPMUnwrapWETH9Request{AmountMinimum: big.NewInt(11), Recipient: a}).Validate(), "invalid request: contract is nil"},
		{"nFPMUnwrapWETH9 nil amountMinimum", (&NFPMUnwrapWETH9Request{Contract: nfpm, Recipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"nFPMUnwrapWETH9 nil recipient", (&NFPMUnwrapWETH9Request{Contract: nfpm, AmountMinimum: big.NewInt(11)}).Validate(), "invalid request: recipient is nil"},
		{"nFPMUnwrapWETH9 valid", (&NFPMUnwrapWETH9Request{Contract: nfpm, AmountMinimum: big.NewInt(11), Recipient: a}).Validate(), ""},
		{"nFPMSweepToken nil", (*NFPMSweepTokenRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMSweepToken empty", (&NFPMSweepTokenRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMSweepToken nil contract", (&NFPMSweepTokenRequest{Token: a, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), "invalid request: contract is nil"},
		{"nFPMSweepToken nil token", (&NFPMSweepTokenRequest{Contract: nfpm, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), "invalid request: token is nil"},
		{"nFPMSweepToken nil amountMinimum", (&NFPMSweepTokenRequest{Contract: nfpm, Token: a, Recipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"nFPMSweepToken nil recipient", (&NFPMSweepTokenRequest{Contract: nfpm, Token: a, AmountMinimum: big.NewInt(21)}).Validate(), "invalid request: recipient is nil"},
		{"nFPMSweepToken valid", (&NFPMSweepTokenRequest{Contract: nfpm, Token: a, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), ""},
		{"nFPMRefundETH nil", (*NFPMRefundETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMRefundETH empty", (&NFPMRefundETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMRefundETH nil contract", (&NFPMRefundETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMRefundETH valid", (&NFPMRefundETHRequest{Contract: nfpm}).Validate(), ""},
		{"nFPMMulticall nil", (*NFPMMulticallRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMMulticall empty", (&NFPMMulticallRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMMulticall nil contract", (&NFPMMulticallRequest{Data: [][]byte{{0x01}}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"nFPMMulticall empty data", (&NFPMMulticallRequest{Contract: nfpm, Data: [][]byte{}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: data is empty"},
		{"nFPMMulticall valid", (&NFPMMulticallRequest{Contract: nfpm, Data: [][]byte{{0x01}}, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"nFPMMulticall nil nativeAmount", (&NFPMMulticallRequest{Contract: nfpm, Data: [][]byte{{0x01}}}).Validate(), ""},
		{"nFPMMint nil", (*NFPMMintRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMMint empty", (&NFPMMintRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMMint nil contract", (&NFPMMintRequest{Params: mintParams, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"nFPMMint nil params", (&NFPMMintRequest{Contract: nfpm, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"nFPMMint nil params.token0", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.token0 is nil"},
		{"nFPMMint nil params.token1", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.token1 is nil"},
		{"nFPMMint nil params.fee", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.fee is nil"},
		{"nFPMMint nil params.tickLower", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tickLower is nil"},
		{"nFPMMint nil params.tickUpper", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tickUpper is nil"},
		{"nFPMMint nil params.amount0Desired", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount0Desired is nil"},
		{"nFPMMint nil params.amount1Desired", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount1Desired is nil"},
		{"nFPMMint nil params.amount0Min", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount0Min is nil"},
		{"nFPMMint nil params.amount1Min", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Recipient: mintParams.Recipient, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount1Min is nil"},
		{"nFPMMint nil params.recipient", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Deadline: mintParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.recipient is nil"},
		{"nFPMMint nil params.deadline", (&NFPMMintRequest{Contract: nfpm, Params: &MintParams{Token0: mintParams.Token0, Token1: mintParams.Token1, Fee: mintParams.Fee, TickLower: mintParams.TickLower, TickUpper: mintParams.TickUpper, Amount0Desired: mintParams.Amount0Desired, Amount1Desired: mintParams.Amount1Desired, Amount0Min: mintParams.Amount0Min, Amount1Min: mintParams.Amount1Min, Recipient: mintParams.Recipient}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.deadline is nil"},
		{"nFPMMint valid", (&NFPMMintRequest{Contract: nfpm, Params: mintParams, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"nFPMMint nil nativeAmount", (&NFPMMintRequest{Contract: nfpm, Params: mintParams}).Validate(), ""},
		{"increaseLiquidity nil", (*IncreaseLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"increaseLiquidity empty", (&IncreaseLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"increaseLiquidity nil contract", (&IncreaseLiquidityRequest{Params: increaseParams, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"increaseLiquidity nil params", (&IncreaseLiquidityRequest{Contract: nfpm, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"increaseLiquidity nil params.tokenId", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{Amount0Desired: increaseParams.Amount0Desired, Amount1Desired: increaseParams.Amount1Desired, Amount0Min: increaseParams.Amount0Min, Amount1Min: increaseParams.Amount1Min, Deadline: increaseParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tokenId is nil"},
		{"increaseLiquidity nil params.amount0Desired", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{TokenId: increaseParams.TokenId, Amount1Desired: increaseParams.Amount1Desired, Amount0Min: increaseParams.Amount0Min, Amount1Min: increaseParams.Amount1Min, Deadline: increaseParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount0Desired is nil"},
		{"increaseLiquidity nil params.amount1Desired", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{TokenId: increaseParams.TokenId, Amount0Desired: increaseParams.Amount0Desired, Amount0Min: increaseParams.Amount0Min, Amount1Min: increaseParams.Amount1Min, Deadline: increaseParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount1Desired is nil"},
		{"increaseLiquidity nil params.amount0Min", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{TokenId: increaseParams.TokenId, Amount0Desired: increaseParams.Amount0Desired, Amount1Desired: increaseParams.Amount1Desired, Amount1Min: increaseParams.Amount1Min, Deadline: increaseParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount0Min is nil"},
		{"increaseLiquidity nil params.amount1Min", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{TokenId: increaseParams.TokenId, Amount0Desired: increaseParams.Amount0Desired, Amount1Desired: increaseParams.Amount1Desired, Amount0Min: increaseParams.Amount0Min, Deadline: increaseParams.Deadline}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amount1Min is nil"},
		{"increaseLiquidity nil params.deadline", (&IncreaseLiquidityRequest{Contract: nfpm, Params: &IncreaseLiquidityParams{TokenId: increaseParams.TokenId, Amount0Desired: increaseParams.Amount0Desired, Amount1Desired: increaseParams.Amount1Desired, Amount0Min: increaseParams.Amount0Min, Amount1Min: increaseParams.Amount1Min}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.deadline is nil"},
		{"increaseLiquidity valid", (&IncreaseLiquidityRequest{Contract: nfpm, Params: increaseParams, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"increaseLiquidity nil nativeAmount", (&IncreaseLiquidityRequest{Contract: nfpm, Params: increaseParams}).Validate(), ""},
		{"decreaseLiquidity nil", (*DecreaseLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"decreaseLiquidity empty", (&DecreaseLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"decreaseLiquidity nil contract", (&DecreaseLiquidityRequest{Params: decreaseParams}).Validate(), "invalid request: contract is nil"},
		{"decreaseLiquidity nil params", (&DecreaseLiquidityRequest{Contract: nfpm}).Validate(), "invalid request: params is nil"},
		{"decreaseLiquidity nil params.tokenId", (&DecreaseLiquidityRequest{Contract: nfpm, Params: &DecreaseLiquidityParams{Liquidity: decreaseParams.Liquidity, Amount0Min: decreaseParams.Amount0Min, Amount1Min: decreaseParams.Amount1Min, Deadline: decreaseParams.Deadline}}).Validate(), "invalid request: params.tokenId is nil"},
		{"decreaseLiquidity nil params.liquidity", (&DecreaseLiquidityRequest{Contract: nfpm, Params: &DecreaseLiquidityParams{TokenId: decreaseParams.TokenId, Amount0Min: decreaseParams.Amount0Min, Amount1Min: decreaseParams.Amount1Min, Deadline: decreaseParams.Deadline}}).Validate(), "invalid request: params.liquidity is nil"},
		{"decreaseLiquidity nil params.amount0Min", (&DecreaseLiquidityRequest{Contract: nfpm, Params: &DecreaseLiquidityParams{TokenId: decreaseParams.TokenId, Liquidity: decreaseParams.Liquidity, Amount1Min: decreaseParams.Amount1Min, Deadline: decreaseParams.Deadline}}).Validate(), "invalid request: params.amount0Min is nil"},
		{"decreaseLiquidity nil params.amount1Min", (&DecreaseLiquidityRequest{Contract: nfpm, Params: &DecreaseLiquidityParams{TokenId: decreaseParams.TokenId, Liquidity: decreaseParams.Liquidity, Amount0Min: decreaseParams.Amount0Min, Deadline: decreaseParams.Deadline}}).Validate(), "invalid request: params.amount1Min is nil"},
		{"decreaseLiquidity nil params.deadline", (&DecreaseLiquidityRequest{Contract: nfpm, Params: &DecreaseLiquidityParams{TokenId: decreaseParams.TokenId, Liquidity: decreaseParams.Liquidity, Amount0Min: decreaseParams.Amount0Min, Amount1Min: decreaseParams.Amount1Min}}).Validate(), "invalid request: params.deadline is nil"},
		{"decreaseLiquidity valid", (&DecreaseLiquidityRequest{Contract: nfpm, Params: decreaseParams}).Validate(), ""},
		{"nFPMCollect nil", (*NFPMCollectRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMCollect empty", (&NFPMCollectRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMCollect nil contract", (&NFPMCollectRequest{Params: collectParams}).Validate(), "invalid request: contract is nil"},
		{"nFPMCollect nil params", (&NFPMCollectRequest{Contract: nfpm}).Validate(), "invalid request: params is nil"},
		{"nFPMCollect nil params.tokenId", (&NFPMCollectRequest{Contract: nfpm, Params: &CollectParams{Recipient: collectParams.Recipient, Amount0Max: collectParams.Amount0Max, Amount1Max: collectParams.Amount1Max}}).Validate(), "invalid request: params.tokenId is nil"},
		{"nFPMCollect nil params.recipient", (&NFPMCollectRequest{Contract: nfpm, Params: &CollectParams{TokenId: collectParams.TokenId, Amount0Max: collectParams.Amount0Max, Amount1Max: collectParams.Amount1Max}}).Validate(), "invalid request: params.recipient is nil"},
		{"nFPMCollect nil params.amount0Max", (&NFPMCollectRequest{Contract: nfpm, Params: &CollectParams{TokenId: collectParams.TokenId, Recipient: collectParams.Recipient, Amount1Max: collectParams.Amount1Max}}).Validate(), "invalid request: params.amount0Max is nil"},
		{"nFPMCollect nil params.amount1Max", (&NFPMCollectRequest{Contract: nfpm, Params: &CollectParams{TokenId: collectParams.TokenId, Recipient: collectParams.Recipient, Amount0Max: collectParams.Amount0Max}}).Validate(), "invalid request: params.amount1Max is nil"},
		{"nFPMCollect valid", (&NFPMCollectRequest{Contract: nfpm, Params: collectParams}).Validate(), ""},
		{"nFPMBurn nil", (*NFPMBurnRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nFPMBurn empty", (&NFPMBurnRequest{}).Validate(), "invalid request: contract is nil"},
		{"nFPMBurn nil contract", (&NFPMBurnRequest{TokenID: big.NewInt(11)}).Validate(), "invalid request: contract is nil"},
		{"nFPMBurn nil tokenID", (&NFPMBurnRequest{Contract: nfpm}).Validate(), "invalid request: tokenID is nil"},
		{"nFPMBurn valid", (&NFPMBurnRequest{Contract: nfpm, TokenID: big.NewInt(11)}).Validate(), ""},
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

func TestNewNFPMRequestsSetFields(t *testing.T) {
	nfpm := types.MustNewAddressFromHex("0xC36442b4a4522E871399CD717aBDD847Ab11FE88")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	mintParams := NewMintParams(a, a, big.NewInt(3000), big.NewInt(-60), big.NewInt(60), big.NewInt(1), big.NewInt(2), big.NewInt(0), big.NewInt(0), a, big.NewInt(1_700_000_000))
	increaseParams := NewIncreaseLiquidityParams(big.NewInt(7), big.NewInt(1), big.NewInt(2), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	decreaseParams := NewDecreaseLiquidityParams(big.NewInt(7), big.NewInt(100), big.NewInt(0), big.NewInt(0), big.NewInt(1_700_000_000))
	collectParams := NewCollectParams(big.NewInt(7), a, big.NewInt(5), big.NewInt(6))

	require.Equal(t, &WETH9Request{Contract: nfpm}, NewWETH9Request(nfpm))
	require.Equal(t, &TotalSupplyRequest{Contract: nfpm}, NewTotalSupplyRequest(nfpm))
	require.Equal(t, &TokenOfOwnerByIndexRequest{Contract: nfpm, Owner: a, Index: big.NewInt(21)}, NewTokenOfOwnerByIndexRequest(nfpm, a, big.NewInt(21)))
	require.Equal(t, &TokenByIndexRequest{Contract: nfpm, Index: big.NewInt(11)}, NewTokenByIndexRequest(nfpm, big.NewInt(11)))
	require.Equal(t, &PermitTypehashRequest{Contract: nfpm}, NewPermitTypehashRequest(nfpm))
	require.Equal(t, &DomainSeparatorRequest{Contract: nfpm}, NewDomainSeparatorRequest(nfpm))
	require.Equal(t, &NFPMPositionsRequest{Contract: nfpm, TokenID: big.NewInt(11)}, NewNFPMPositionsRequest(nfpm, big.NewInt(11)))
	require.Equal(t, &PermitRequest{Contract: nfpm, Spender: a, TokenID: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}, NewPermitRequest(nfpm, a, big.NewInt(21), big.NewInt(31), sig))
	require.Equal(t, &CreateAndInitializePoolIfNecessaryRequest{Contract: nfpm, Token0: a, Token1: a, Fee: big.NewInt(31), SqrtPriceX96: big.NewInt(41)}, NewCreateAndInitializePoolIfNecessaryRequest(nfpm, a, a, big.NewInt(31), big.NewInt(41)))
	require.Equal(t, &NFPMUnwrapWETH9Request{Contract: nfpm, AmountMinimum: big.NewInt(11), Recipient: a}, NewNFPMUnwrapWETH9Request(nfpm, big.NewInt(11), a))
	require.Equal(t, &NFPMSweepTokenRequest{Contract: nfpm, Token: a, AmountMinimum: big.NewInt(21), Recipient: a}, NewNFPMSweepTokenRequest(nfpm, a, big.NewInt(21), a))
	require.Equal(t, &NFPMRefundETHRequest{Contract: nfpm}, NewNFPMRefundETHRequest(nfpm))
	require.Equal(t, &NFPMMulticallRequest{Contract: nfpm, Data: [][]byte{{0x01}}, NativeAmount: big.NewInt(21)}, NewNFPMMulticallRequest(nfpm, [][]byte{{0x01}}, big.NewInt(21)))
	require.Equal(t, &NFPMMintRequest{Contract: nfpm, Params: mintParams, NativeAmount: big.NewInt(21)}, NewNFPMMintRequest(nfpm, mintParams, big.NewInt(21)))
	require.Equal(t, &IncreaseLiquidityRequest{Contract: nfpm, Params: increaseParams, NativeAmount: big.NewInt(21)}, NewIncreaseLiquidityRequest(nfpm, increaseParams, big.NewInt(21)))
	require.Equal(t, &DecreaseLiquidityRequest{Contract: nfpm, Params: decreaseParams}, NewDecreaseLiquidityRequest(nfpm, decreaseParams))
	require.Equal(t, &NFPMCollectRequest{Contract: nfpm, Params: collectParams}, NewNFPMCollectRequest(nfpm, collectParams))
	require.Equal(t, &NFPMBurnRequest{Contract: nfpm, TokenID: big.NewInt(11)}, NewNFPMBurnRequest(nfpm, big.NewInt(11)))
}
