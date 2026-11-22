package permit2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

func TestNewPermitDetails(t *testing.T) {
	token := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(100)
	expiration := big.NewInt(200)
	nonce := big.NewInt(1)

	d := NewPermitDetails(token, amount, expiration, nonce)
	require.Same(t, token, d.Token)
	require.Same(t, amount, d.Amount)
	require.Same(t, expiration, d.Expiration)
	require.Same(t, nonce, d.Nonce)
}

func TestNewPermitSingle(t *testing.T) {
	details := NewPermitDetails(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(100), big.NewInt(200), big.NewInt(1))
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	sigDeadline := big.NewInt(9999999999)

	p := NewPermitSingle(details, spender, sigDeadline)
	require.Same(t, details, p.Details)
	require.Same(t, spender, p.Spender)
	require.Same(t, sigDeadline, p.SigDeadline)
}

func TestNewPermitBatch(t *testing.T) {
	details := []PermitDetails{*NewPermitDetails(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(100), big.NewInt(200), big.NewInt(1))}
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	sigDeadline := big.NewInt(9999999999)

	p := NewPermitBatch(details, spender, sigDeadline)
	require.Equal(t, details, p.Details)
	require.Same(t, spender, p.Spender)
	require.Same(t, sigDeadline, p.SigDeadline)
}

func TestNewPackedAllowance(t *testing.T) {
	amount := big.NewInt(100)
	expiration := big.NewInt(200)
	nonce := big.NewInt(1)

	a := NewPackedAllowance(amount, expiration, nonce)
	require.Same(t, amount, a.Amount)
	require.Same(t, expiration, a.Expiration)
	require.Same(t, nonce, a.Nonce)
}

func TestNewTokenSpenderPair(t *testing.T) {
	token := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	p := NewTokenSpenderPair(token, spender)
	require.Same(t, token, p.Token)
	require.Same(t, spender, p.Spender)
}

func TestNewAllowanceTransferDetails(t *testing.T) {
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount := big.NewInt(100)
	token := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	d := NewAllowanceTransferDetails(from, to, amount, token)
	require.Same(t, from, d.From)
	require.Same(t, to, d.To)
	require.Same(t, amount, d.Amount)
	require.Same(t, token, d.Token)
}

func TestNewTokenPermissions(t *testing.T) {
	token := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(100)

	p := NewTokenPermissions(token, amount)
	require.Same(t, token, p.Token)
	require.Same(t, amount, p.Amount)
}

func TestNewPermitTransferFrom(t *testing.T) {
	permitted := NewTokenPermissions(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(100))
	nonce := big.NewInt(1)
	deadline := big.NewInt(9999999999)

	p := NewPermitTransferFrom(permitted, nonce, deadline)
	require.Same(t, permitted, p.Permitted)
	require.Same(t, nonce, p.Nonce)
	require.Same(t, deadline, p.Deadline)
}

func TestNewSignatureTransferDetails(t *testing.T) {
	to := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	requestedAmount := big.NewInt(100)

	d := NewSignatureTransferDetails(to, requestedAmount)
	require.Same(t, to, d.To)
	require.Same(t, requestedAmount, d.RequestedAmount)
}

func TestNewPermitBatchTransferFrom(t *testing.T) {
	permitted := []TokenPermissions{*NewTokenPermissions(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(100))}
	nonce := big.NewInt(1)
	deadline := big.NewInt(9999999999)

	p := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	require.Equal(t, permitted, p.Permitted)
	require.Same(t, nonce, p.Nonce)
	require.Same(t, deadline, p.Deadline)
}

func TestNewNonceInvalidationEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	newNonce := big.NewInt(2)
	oldNonce := big.NewInt(1)

	n := NewNonceInvalidationEvent(contract, testOwner, testToken, testSpender, newNonce, oldNonce)
	require.Same(t, contract, n.Contract)
	require.Same(t, testOwner, n.Owner)
	require.Same(t, testToken, n.Token)
	require.Same(t, testSpender, n.Spender)
	require.Same(t, newNonce, n.NewNonce)
	require.Same(t, oldNonce, n.OldNonce)
}

func TestNewApprovalEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(100)
	expiration := big.NewInt(200)

	a := NewApprovalEvent(contract, testOwner, testToken, testSpender, amount, expiration)
	require.Same(t, contract, a.Contract)
	require.Same(t, testOwner, a.Owner)
	require.Same(t, testToken, a.Token)
	require.Same(t, testSpender, a.Spender)
	require.Same(t, amount, a.Amount)
	require.Same(t, expiration, a.Expiration)
}

func TestNewPermitEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(100)
	expiration := big.NewInt(200)
	nonce := big.NewInt(1)

	p := NewPermitEvent(contract, testOwner, testToken, testSpender, amount, expiration, nonce)
	require.Same(t, contract, p.Contract)
	require.Same(t, testOwner, p.Owner)
	require.Same(t, testToken, p.Token)
	require.Same(t, testSpender, p.Spender)
	require.Same(t, amount, p.Amount)
	require.Same(t, expiration, p.Expiration)
	require.Same(t, nonce, p.Nonce)
}

func TestNewLockdownEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	l := NewLockdownEvent(contract, testOwner, testToken, testSpender)
	require.Same(t, contract, l.Contract)
	require.Same(t, testOwner, l.Owner)
	require.Same(t, testToken, l.Token)
	require.Same(t, testSpender, l.Spender)
}

func TestNewUnorderedNonceInvalidationEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	word := big.NewInt(1)
	mask := big.NewInt(2)

	n := NewUnorderedNonceInvalidationEvent(contract, testOwner, word, mask)
	require.Same(t, contract, n.Contract)
	require.Same(t, testOwner, n.Owner)
	require.Same(t, word, n.Word)
	require.Same(t, mask, n.Mask)
}

func TestNewDomain(t *testing.T) {
	chainID := big.NewInt(1)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	d := NewDomain(chainID, contract)
	require.Equal(t, "Permit2", d.Name)
	require.Empty(t, d.Version)
	require.Same(t, chainID, d.ChainID)
	require.Same(t, contract, d.VerifyingContract)
	require.Nil(t, d.Salt)
}

func TestPermitDetailsTuple(t *testing.T) {
	d := NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	require.Equal(t, []any{d.Token, d.Amount, d.Expiration, d.Nonce}, d.tuple())
}

func TestPermitDetailsMessage(t *testing.T) {
	d := NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	require.Equal(t, eip712.Message{
		tokenFieldName:      d.Token,
		amountFieldName:     d.Amount,
		expirationFieldName: d.Expiration,
		nonceFieldName:      d.Nonce,
	}, d.message())
}

func TestPermitSingleTuple(t *testing.T) {
	details := NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	p := NewPermitSingle(details, testSpender, big.NewInt(9999999999))
	require.Equal(t, []any{details.tuple(), p.Spender, p.SigDeadline}, p.tuple())
}

func TestPermitBatchTuple(t *testing.T) {
	details := []PermitDetails{
		*NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)),
		*NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2)),
	}
	p := NewPermitBatch(details, testSpender, big.NewInt(9999999999))
	got := p.tuple()
	gotDetails := got[0].([][]any)
	require.Len(t, gotDetails, 2)
	require.Equal(t, details[0].tuple(), gotDetails[0])
	require.Equal(t, details[1].tuple(), gotDetails[1])
	require.Same(t, p.Spender, got[1])
	require.Same(t, p.SigDeadline, got[2])
}

func TestTokenSpenderPairTuple(t *testing.T) {
	p := NewTokenSpenderPair(testToken, testSpender)
	require.Equal(t, []any{p.Token, p.Spender}, p.tuple())
}

func TestAllowanceTransferDetailsTuple(t *testing.T) {
	d := NewAllowanceTransferDetails(testOwner, testTo, big.NewInt(100), testToken)
	require.Equal(t, []any{d.From, d.To, d.Amount, d.Token}, d.tuple())
}

func TestPermitSingleTypeHashMatchesPermitHashSol(t *testing.T) {
	want := "PermitSingle(PermitDetails details,address spender,uint256 sigDeadline)" +
		"PermitDetails(address token,uint160 amount,uint48 expiration,uint48 nonce)"
	require.Equal(t, want, permitSingleTypes.EncodeType(permitSinglePrimaryType))
}

func TestPermitBatchTypeHashMatchesPermitHashSol(t *testing.T) {
	want := "PermitBatch(PermitDetails[] details,address spender,uint256 sigDeadline)" +
		"PermitDetails(address token,uint160 amount,uint48 expiration,uint48 nonce)"
	require.Equal(t, want, permitBatchTypes.EncodeType(permitBatchPrimaryType))
}

func TestPermitSingleHashDeterministic(t *testing.T) {
	details := NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	p := NewPermitSingle(details, testSpender, big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h1, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestPermitSingleHashSensitiveToEachField(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	base := NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variants := []*PermitSingle{
		NewPermitSingle(NewPermitDetails(testTokenB, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999999)),
		NewPermitSingle(NewPermitDetails(testToken, big.NewInt(101), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999999)),
		NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(201), big.NewInt(1)), testSpender, big.NewInt(9999999999)),
		NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(2)), testSpender, big.NewInt(9999999999)),
		NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testToken, big.NewInt(9999999999)),
		NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestPermitSingleHashSensitiveToDomain(t *testing.T) {
	p := NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999999))

	h1, err := p.Hash(NewDomain(big.NewInt(1), testToken))
	require.NoError(t, err)
	h2, err := p.Hash(NewDomain(big.NewInt(2), testToken))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())

	h3, err := p.Hash(NewDomain(big.NewInt(1), testTokenB))
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h3.Bytes())
}

func TestPermitSingleSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	p := NewPermitSingle(NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)), testSpender, big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	sig, err := p.Sign(domain, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPermitBatchHashDeterministic(t *testing.T) {
	details := []PermitDetails{
		*NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)),
		*NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2)),
	}
	p := NewPermitBatch(details, testSpender, big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h1, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestPermitBatchHashSensitiveToEachField(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	baseDetails := []PermitDetails{
		*NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)),
		*NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2)),
	}
	base := NewPermitBatch(baseDetails, testSpender, big.NewInt(9999999999))
	baseHash, err := base.Hash(domain)
	require.NoError(t, err)

	variantDetails := []PermitDetails{
		*NewPermitDetails(testToken, big.NewInt(101), big.NewInt(200), big.NewInt(1)),
		*NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2)),
	}
	variants := []*PermitBatch{
		NewPermitBatch(variantDetails, testSpender, big.NewInt(9999999999)),
		NewPermitBatch(baseDetails[:1], testSpender, big.NewInt(9999999999)),
		NewPermitBatch(baseDetails, testToken, big.NewInt(9999999999)),
		NewPermitBatch(baseDetails, testSpender, big.NewInt(9999999998)),
	}
	for i, v := range variants {
		h, err := v.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestPermitBatchHashSensitiveToDetailsOrder(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	d0 := *NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	d1 := *NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2))

	forward := NewPermitBatch([]PermitDetails{d0, d1}, testSpender, big.NewInt(9999999999))
	reversed := NewPermitBatch([]PermitDetails{d1, d0}, testSpender, big.NewInt(9999999999))

	hForward, err := forward.Hash(domain)
	require.NoError(t, err)
	hReversed, err := reversed.Hash(domain)
	require.NoError(t, err)
	require.NotEqual(t, hForward.Bytes(), hReversed.Bytes())
}

func TestPermitBatchHashEmptyDetails(t *testing.T) {
	p := NewPermitBatch(nil, testSpender, big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h, err := p.Hash(domain)
	require.NoError(t, err)
	require.Len(t, h.Bytes(), 32)
}

func TestPermitBatchSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	details := []PermitDetails{*NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))}
	p := NewPermitBatch(details, testSpender, big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	sig, err := p.Sign(domain, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestTokenPermissionsTuple(t *testing.T) {
	p := NewTokenPermissions(testToken, big.NewInt(100))
	require.Equal(t, []any{p.Token, p.Amount}, p.tuple())
}

func TestTokenPermissionsMessage(t *testing.T) {
	p := NewTokenPermissions(testToken, big.NewInt(100))
	require.Equal(t, eip712.Message{
		tokenFieldName:  p.Token,
		amountFieldName: p.Amount,
	}, p.message())
}

func TestPermitTransferFromTuple(t *testing.T) {
	permitted := NewTokenPermissions(testToken, big.NewInt(100))
	p := NewPermitTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	require.Equal(t, []any{permitted.tuple(), p.Nonce, p.Deadline}, p.tuple())
}

func TestSignatureTransferDetailsTuple(t *testing.T) {
	d := NewSignatureTransferDetails(testTo, big.NewInt(100))
	require.Equal(t, []any{d.To, d.RequestedAmount}, d.tuple())
}

func TestPermitBatchTransferFromTuple(t *testing.T) {
	permitted := []TokenPermissions{
		*NewTokenPermissions(testToken, big.NewInt(100)),
		*NewTokenPermissions(testTokenB, big.NewInt(200)),
	}
	p := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	got := p.tuple()
	gotPermitted := got[0].([][]any)
	require.Len(t, gotPermitted, 2)
	require.Equal(t, permitted[0].tuple(), gotPermitted[0])
	require.Equal(t, permitted[1].tuple(), gotPermitted[1])
	require.Same(t, p.Nonce, got[1])
	require.Same(t, p.Deadline, got[2])
}

func TestPermitTransferFromTypeHashMatchesPermitHashSol(t *testing.T) {
	want := "PermitTransferFrom(TokenPermissions permitted,address spender,uint256 nonce,uint256 deadline)" +
		"TokenPermissions(address token,uint256 amount)"
	require.Equal(t, want, permitTransferFromTypes.EncodeType(permitTransferFromPrimaryType))
}

func TestPermitBatchTransferFromTypeHashMatchesPermitHashSol(t *testing.T) {
	want := "PermitBatchTransferFrom(TokenPermissions[] permitted,address spender,uint256 nonce,uint256 deadline)" +
		"TokenPermissions(address token,uint256 amount)"
	require.Equal(t, want, permitBatchTransferFromTypes.EncodeType(permitBatchTransferFromPrimaryType))
}

func TestPermitTransferFromHashDeterministic(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h1, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestPermitTransferFromHashSensitiveToEachField(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	base := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain, testSpender)
	require.NoError(t, err)

	variants := []struct {
		p       *PermitTransferFrom
		spender *types.Address
	}{
		{NewPermitTransferFrom(NewTokenPermissions(testTokenB, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999)), testSpender},
		{NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(101)), big.NewInt(1), big.NewInt(9999999999)), testSpender},
		{NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(2), big.NewInt(9999999999)), testSpender},
		{NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999998)), testSpender},
		{NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999)), testTokenB},
	}
	for i, v := range variants {
		h, err := v.p.Hash(domain, v.spender)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestPermitTransferFromHashSensitiveToDomain(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))

	h1, err := p.Hash(NewDomain(big.NewInt(1), testToken), testSpender)
	require.NoError(t, err)
	h2, err := p.Hash(NewDomain(big.NewInt(2), testToken), testSpender)
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())
}

func TestPermitTransferFromSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	sig, err := p.Sign(domain, testSpender, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain, testSpender)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPermitTransferFromHashWithWitnessDeterministic(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))

	h1, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestPermitTransferFromHashWithWitnessDiffersFromHash(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))

	plain, err := p.Hash(domain, testSpender)
	require.NoError(t, err)

	withWitness, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)

	require.NotEqual(t, plain.Bytes(), withWitness.Bytes())
}

func TestPermitTransferFromHashWithWitnessSensitiveToWitness(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	typeString := "Witness witness)Witness(uint256 x)"

	h1, err := p.HashWithWitness(domain, testSpender, types.NewHashFromBytes([]byte("witness one, 32 bytes long here")), typeString)
	require.NoError(t, err)
	h2, err := p.HashWithWitness(domain, testSpender, types.NewHashFromBytes([]byte("witness two, 32 bytes long here")), typeString)
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())
}

func TestPermitTransferFromHashWithWitnessSensitiveToWitnessTypeString(t *testing.T) {
	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))

	h1, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)
	h2, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 y)")
	require.NoError(t, err)
	require.NotEqual(t, h1.Bytes(), h2.Bytes())
}

func TestPermitTransferFromSignWithWitness(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	p := NewPermitTransferFrom(NewTokenPermissions(testToken, big.NewInt(100)), big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))
	typeString := "Witness witness)Witness(uint256 x)"

	sig, err := p.SignWithWitness(domain, testSpender, witness, typeString, key)
	require.NoError(t, err)

	digest, err := p.HashWithWitness(domain, testSpender, witness, typeString)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPermitBatchTransferFromHashDeterministic(t *testing.T) {
	permitted := []TokenPermissions{*NewTokenPermissions(testToken, big.NewInt(100))}
	p := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h1, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())
}

func TestPermitBatchTransferFromHashSensitiveToEachField(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	basePermitted := []TokenPermissions{
		*NewTokenPermissions(testToken, big.NewInt(100)),
		*NewTokenPermissions(testTokenB, big.NewInt(200)),
	}
	base := NewPermitBatchTransferFrom(basePermitted, big.NewInt(1), big.NewInt(9999999999))
	baseHash, err := base.Hash(domain, testSpender)
	require.NoError(t, err)

	variantPermitted := []TokenPermissions{
		*NewTokenPermissions(testToken, big.NewInt(101)),
		*NewTokenPermissions(testTokenB, big.NewInt(200)),
	}
	variants := []struct {
		p       *PermitBatchTransferFrom
		spender *types.Address
	}{
		{NewPermitBatchTransferFrom(variantPermitted, big.NewInt(1), big.NewInt(9999999999)), testSpender},
		{NewPermitBatchTransferFrom(basePermitted[:1], big.NewInt(1), big.NewInt(9999999999)), testSpender},
		{NewPermitBatchTransferFrom(basePermitted, big.NewInt(2), big.NewInt(9999999999)), testSpender},
		{NewPermitBatchTransferFrom(basePermitted, big.NewInt(1), big.NewInt(9999999998)), testSpender},
		{NewPermitBatchTransferFrom(basePermitted, big.NewInt(1), big.NewInt(9999999999)), testTokenB},
	}
	for i, v := range variants {
		h, err := v.p.Hash(domain, v.spender)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), h.Bytes(), "variant %d", i)
	}
}

func TestPermitBatchTransferFromHashSensitiveToOrder(t *testing.T) {
	domain := NewDomain(big.NewInt(1), testToken)
	t0 := *NewTokenPermissions(testToken, big.NewInt(100))
	t1 := *NewTokenPermissions(testTokenB, big.NewInt(200))

	forward := NewPermitBatchTransferFrom([]TokenPermissions{t0, t1}, big.NewInt(1), big.NewInt(9999999999))
	reversed := NewPermitBatchTransferFrom([]TokenPermissions{t1, t0}, big.NewInt(1), big.NewInt(9999999999))

	hForward, err := forward.Hash(domain, testSpender)
	require.NoError(t, err)
	hReversed, err := reversed.Hash(domain, testSpender)
	require.NoError(t, err)
	require.NotEqual(t, hForward.Bytes(), hReversed.Bytes())
}

func TestPermitBatchTransferFromHashEmptyPermitted(t *testing.T) {
	p := NewPermitBatchTransferFrom(nil, big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	h, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.Len(t, h.Bytes(), 32)
}

func TestPermitBatchTransferFromSign(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	permitted := []TokenPermissions{*NewTokenPermissions(testToken, big.NewInt(100))}
	p := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)

	sig, err := p.Sign(domain, testSpender, key)
	require.NoError(t, err)

	digest, err := p.Hash(domain, testSpender)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPermitBatchTransferFromHashWithWitnessDeterministic(t *testing.T) {
	permitted := []TokenPermissions{*NewTokenPermissions(testToken, big.NewInt(100))}
	p := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))

	h1, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)
	require.Len(t, h1.Bytes(), 32)

	h2, err := p.HashWithWitness(domain, testSpender, witness, "Witness witness)Witness(uint256 x)")
	require.NoError(t, err)
	require.Equal(t, h1.Bytes(), h2.Bytes())

	plain, err := p.Hash(domain, testSpender)
	require.NoError(t, err)
	require.NotEqual(t, plain.Bytes(), h1.Bytes())
}

func TestPermitBatchTransferFromSignWithWitness(t *testing.T) {
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	signerAddr := core.PubkeyToAddress(key.PublicKey())

	permitted := []TokenPermissions{*NewTokenPermissions(testToken, big.NewInt(100))}
	p := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	domain := NewDomain(big.NewInt(1), testToken)
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))
	typeString := "Witness witness)Witness(uint256 x)"

	sig, err := p.SignWithWitness(domain, testSpender, witness, typeString, key)
	require.NoError(t, err)

	digest, err := p.HashWithWitness(domain, testSpender, witness, typeString)
	require.NoError(t, err)

	ok, err := core.VerifyAddress(digest, sig, signerAddr)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	witness := types.NewHashFromBytes([]byte{0x01})
	sig := &types.Signature{}

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"domainSeparator nil", (*DomainSeparatorRequest)(nil).Validate(), "invalid request: request is nil"},
		{"domainSeparator empty", (&DomainSeparatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"domainSeparator nil Contract", (&DomainSeparatorRequest{}).Validate(), "invalid request: contract is nil"},
		{"domainSeparator valid", (&DomainSeparatorRequest{Contract: contract}).Validate(), ""},
		{"allowance nil", (*AllowanceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"allowance empty", (&AllowanceRequest{}).Validate(), "invalid request: contract is nil"},
		{"allowance nil Contract", (&AllowanceRequest{Owner: owner, Token: token, Spender: spender}).Validate(), "invalid request: contract is nil"},
		{"allowance nil Owner", (&AllowanceRequest{Contract: contract, Token: token, Spender: spender}).Validate(), "invalid request: owner is nil"},
		{"allowance nil Token", (&AllowanceRequest{Contract: contract, Owner: owner, Spender: spender}).Validate(), "invalid request: token is nil"},
		{"allowance nil Spender", (&AllowanceRequest{Contract: contract, Owner: owner, Token: token}).Validate(), "invalid request: spender is nil"},
		{"allowance valid", (&AllowanceRequest{Contract: contract, Owner: owner, Token: token, Spender: spender}).Validate(), ""},
		{"nonceBitmap nil", (*NonceBitmapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"nonceBitmap empty", (&NonceBitmapRequest{}).Validate(), "invalid request: contract is nil"},
		{"nonceBitmap nil Contract", (&NonceBitmapRequest{Owner: owner, WordPos: big.NewInt(0)}).Validate(), "invalid request: contract is nil"},
		{"nonceBitmap nil Owner", (&NonceBitmapRequest{Contract: contract, WordPos: big.NewInt(0)}).Validate(), "invalid request: owner is nil"},
		{"nonceBitmap nil WordPos", (&NonceBitmapRequest{Contract: contract, Owner: owner}).Validate(), "invalid request: wordPos is nil"},
		{"nonceBitmap valid", (&NonceBitmapRequest{Contract: contract, Owner: owner, WordPos: big.NewInt(0)}).Validate(), ""},
		{"approve nil", (*ApproveRequest)(nil).Validate(), "invalid request: request is nil"},
		{"approve empty", (&ApproveRequest{}).Validate(), "invalid request: contract is nil"},
		{"approve nil Contract", (&ApproveRequest{Token: token, Spender: spender, Amount: big.NewInt(1), Expiration: big.NewInt(0)}).Validate(), "invalid request: contract is nil"},
		{"approve nil Token", (&ApproveRequest{Contract: contract, Spender: spender, Amount: big.NewInt(1), Expiration: big.NewInt(0)}).Validate(), "invalid request: token is nil"},
		{"approve nil Spender", (&ApproveRequest{Contract: contract, Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(0)}).Validate(), "invalid request: spender is nil"},
		{"approve nil Amount", (&ApproveRequest{Contract: contract, Token: token, Spender: spender, Expiration: big.NewInt(0)}).Validate(), "invalid request: amount is nil"},
		{"approve nil Expiration", (&ApproveRequest{Contract: contract, Token: token, Spender: spender, Amount: big.NewInt(1)}).Validate(), "invalid request: expiration is nil"},
		{"approve valid", (&ApproveRequest{Contract: contract, Token: token, Spender: spender, Amount: big.NewInt(1), Expiration: big.NewInt(0)}).Validate(), ""},
		{"transferFrom nil", (*TransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"transferFrom empty", (&TransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil Contract", (&TransferFromRequest{From: owner, To: spender, Amount: big.NewInt(1), Token: token}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil From", (&TransferFromRequest{Contract: contract, To: spender, Amount: big.NewInt(1), Token: token}).Validate(), "invalid request: from is nil"},
		{"transferFrom nil To", (&TransferFromRequest{Contract: contract, From: owner, Amount: big.NewInt(1), Token: token}).Validate(), "invalid request: to is nil"},
		{"transferFrom nil Amount", (&TransferFromRequest{Contract: contract, From: owner, To: spender, Token: token}).Validate(), "invalid request: amount is nil"},
		{"transferFrom nil Token", (&TransferFromRequest{Contract: contract, From: owner, To: spender, Amount: big.NewInt(1)}).Validate(), "invalid request: token is nil"},
		{"transferFrom valid", (&TransferFromRequest{Contract: contract, From: owner, To: spender, Amount: big.NewInt(1), Token: token}).Validate(), ""},
		{"invalidateNonces nil", (*InvalidateNoncesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"invalidateNonces empty", (&InvalidateNoncesRequest{}).Validate(), "invalid request: contract is nil"},
		{"invalidateNonces nil Contract", (&InvalidateNoncesRequest{Token: token, Spender: spender, NewNonce: big.NewInt(2)}).Validate(), "invalid request: contract is nil"},
		{"invalidateNonces nil Token", (&InvalidateNoncesRequest{Contract: contract, Spender: spender, NewNonce: big.NewInt(2)}).Validate(), "invalid request: token is nil"},
		{"invalidateNonces nil Spender", (&InvalidateNoncesRequest{Contract: contract, Token: token, NewNonce: big.NewInt(2)}).Validate(), "invalid request: spender is nil"},
		{"invalidateNonces nil NewNonce", (&InvalidateNoncesRequest{Contract: contract, Token: token, Spender: spender}).Validate(), "invalid request: newNonce is nil"},
		{"invalidateNonces valid", (&InvalidateNoncesRequest{Contract: contract, Token: token, Spender: spender, NewNonce: big.NewInt(2)}).Validate(), ""},
		{"invalidateUnorderedNonces nil", (*InvalidateUnorderedNoncesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"invalidateUnorderedNonces empty", (&InvalidateUnorderedNoncesRequest{}).Validate(), "invalid request: contract is nil"},
		{"invalidateUnorderedNonces nil Contract", (&InvalidateUnorderedNoncesRequest{WordPos: big.NewInt(0), Mask: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"invalidateUnorderedNonces nil WordPos", (&InvalidateUnorderedNoncesRequest{Contract: contract, Mask: big.NewInt(1)}).Validate(), "invalid request: wordPos is nil"},
		{"invalidateUnorderedNonces nil Mask", (&InvalidateUnorderedNoncesRequest{Contract: contract, WordPos: big.NewInt(0)}).Validate(), "invalid request: mask is nil"},
		{"invalidateUnorderedNonces valid", (&InvalidateUnorderedNoncesRequest{Contract: contract, WordPos: big.NewInt(0), Mask: big.NewInt(1)}).Validate(), ""},
		{"permitSingle nil", (*PermitSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitSingle empty", (&PermitSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitSingle nil Contract", (&PermitSingleRequest{Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitSingle nil Owner", (&PermitSingleRequest{Contract: contract, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitSingle nil PermitSingle", (&PermitSingleRequest{Contract: contract, Owner: owner, Signature: sig}).Validate(), "invalid request: permitSingle is nil"},
		{"permitSingle nil Details", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.details is nil"},
		{"permitSingle nil Spender", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.spender is nil"},
		{"permitSingle nil SigDeadline", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender}, Signature: sig}).Validate(), "invalid request: permitSingle.sigDeadline is nil"},
		{"permitSingle nil details token", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.details.token is nil"},
		{"permitSingle nil details amount", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.details.amount is nil"},
		{"permitSingle nil details expiration", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.details.expiration is nil"},
		{"permitSingle nil details nonce", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitSingle.details.nonce is nil"},
		{"permitSingle nil signature", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}}).Validate(), "invalid request: signature is nil"},
		{"permitSingle valid", (&PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: &PermitSingle{Details: &PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), ""},
		{"permitBatch nil", (*PermitBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitBatch empty", (&PermitBatchRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitBatch nil Contract", (&PermitBatchRequest{Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitBatch nil Owner", (&PermitBatchRequest{Contract: contract, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitBatch nil PermitBatch", (&PermitBatchRequest{Contract: contract, Owner: owner, Signature: sig}).Validate(), "invalid request: permitBatch is nil"},
		{"permitBatch nil details", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details is empty"},
		{"permitBatch empty details", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details is empty"},
		{"permitBatch nil details[1] token", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details[1].token is nil"},
		{"permitBatch nil details[1] amount", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details[1].amount is nil"},
		{"permitBatch nil details[1] expiration", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details[1].expiration is nil"},
		{"permitBatch nil details[1] nonce", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details[1].nonce is nil"},
		{"permitBatch first bad detail wins", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2)}, *&PermitDetails{Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.details[0].nonce is nil"},
		{"permitBatch nil Spender", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), "invalid request: permitBatch.spender is nil"},
		{"permitBatch nil SigDeadline", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender}, Signature: sig}).Validate(), "invalid request: permitBatch.sigDeadline is nil"},
		{"permitBatch nil signature", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}}).Validate(), "invalid request: signature is nil"},
		{"permitBatch valid", (&PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: &PermitBatch{Details: []PermitDetails{*&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}, *&PermitDetails{Token: token, Amount: big.NewInt(1), Expiration: big.NewInt(2), Nonce: big.NewInt(3)}}, Spender: spender, SigDeadline: big.NewInt(9)}, Signature: sig}).Validate(), ""},
		{"transferFromBatch nil", (*TransferFromBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"transferFromBatch empty", (&TransferFromBatchRequest{}).Validate(), "invalid request: contract is nil"},
		{"transferFromBatch nil contract", (&TransferFromBatchRequest{TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token)}}).Validate(), "invalid request: contract is nil"},
		{"transferFromBatch nil transferDetails", (&TransferFromBatchRequest{Contract: contract}).Validate(), "invalid request: transferDetails is empty"},
		{"transferFromBatch nil [1] from", (&TransferFromBatchRequest{Contract: contract, TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(nil, spender, big.NewInt(1), token)}}).Validate(), "invalid request: transferDetails[1].from is nil"},
		{"transferFromBatch nil [1] to", (&TransferFromBatchRequest{Contract: contract, TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, nil, big.NewInt(1), token)}}).Validate(), "invalid request: transferDetails[1].to is nil"},
		{"transferFromBatch nil [1] amount", (&TransferFromBatchRequest{Contract: contract, TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, spender, nil, token)}}).Validate(), "invalid request: transferDetails[1].amount is nil"},
		{"transferFromBatch nil [1] token", (&TransferFromBatchRequest{Contract: contract, TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, spender, big.NewInt(1), nil)}}).Validate(), "invalid request: transferDetails[1].token is nil"},
		{"transferFromBatch valid", (&TransferFromBatchRequest{Contract: contract, TransferDetails: []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token)}}).Validate(), ""},
		{"lockdown nil", (*LockdownRequest)(nil).Validate(), "invalid request: request is nil"},
		{"lockdown empty", (&LockdownRequest{}).Validate(), "invalid request: contract is nil"},
		{"lockdown nil contract", (&LockdownRequest{Approvals: []TokenSpenderPair{*NewTokenSpenderPair(token, spender)}}).Validate(), "invalid request: contract is nil"},
		{"lockdown empty approvals", (&LockdownRequest{Contract: contract, Approvals: []TokenSpenderPair{}}).Validate(), "invalid request: approvals is empty"},
		{"lockdown nil [1] token", (&LockdownRequest{Contract: contract, Approvals: []TokenSpenderPair{*NewTokenSpenderPair(token, spender), *NewTokenSpenderPair(nil, spender)}}).Validate(), "invalid request: approvals[1].token is nil"},
		{"lockdown nil [0] spender", (&LockdownRequest{Contract: contract, Approvals: []TokenSpenderPair{*NewTokenSpenderPair(token, nil)}}).Validate(), "invalid request: approvals[0].spender is nil"},
		{"lockdown valid", (&LockdownRequest{Contract: contract, Approvals: []TokenSpenderPair{*NewTokenSpenderPair(token, spender)}}).Validate(), ""},
		{"permitTransferFrom nil", (*PermitTransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitTransferFrom empty", (&PermitTransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitTransferFrom nil contract", (&PermitTransferFromRequest{Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitTransferFrom nil permit", (&PermitTransferFromRequest{Contract: contract, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit is nil"},
		{"permitTransferFrom nil permitted", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted is nil"},
		{"permitTransferFrom nil permitted token", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(nil, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted.token is nil"},
		{"permitTransferFrom nil permitted amount", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, nil), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted.amount is nil"},
		{"permitTransferFrom nil permit nonce", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit.nonce is nil"},
		{"permitTransferFrom nil permit deadline", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: permit.deadline is nil"},
		{"permitTransferFrom nil transferDetails", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails is nil"},
		{"permitTransferFrom nil transferDetails to", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(nil, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails.to is nil"},
		{"permitTransferFrom nil transferDetails requestedAmount", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, nil), Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails.requestedAmount is nil"},
		{"permitTransferFrom nil owner", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitTransferFrom nil signature", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner}).Validate(), "invalid request: signature is nil"},
		{"permitTransferFrom valid", (&PermitTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Signature: sig}).Validate(), ""},
		{"permitWitnessTransferFrom nil", (*PermitWitnessTransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitWitnessTransferFrom empty", (&PermitWitnessTransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitWitnessTransferFrom nil contract", (&PermitWitnessTransferFromRequest{Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitWitnessTransferFrom nil permit", (&PermitWitnessTransferFromRequest{Contract: contract, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit is nil"},
		{"permitWitnessTransferFrom nil permitted", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted is nil"},
		{"permitWitnessTransferFrom nil permitted token", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(nil, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted.token is nil"},
		{"permitWitnessTransferFrom nil permitted amount", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, nil), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted.amount is nil"},
		{"permitWitnessTransferFrom nil permit nonce", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.nonce is nil"},
		{"permitWitnessTransferFrom nil permit deadline", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.deadline is nil"},
		{"permitWitnessTransferFrom nil transferDetails", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails is nil"},
		{"permitWitnessTransferFrom nil transferDetails to", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(nil, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails.to is nil"},
		{"permitWitnessTransferFrom nil transferDetails requestedAmount", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, nil), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails.requestedAmount is nil"},
		{"permitWitnessTransferFrom nil owner", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitWitnessTransferFrom nil witness", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: witness is nil"},
		{"permitWitnessTransferFrom empty witnessTypeString", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "", Signature: sig}).Validate(), "invalid request: witnessTypeString is empty"},
		{"permitWitnessTransferFrom nil signature", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)"}).Validate(), "invalid request: signature is nil"},
		{"permitWitnessTransferFrom valid", (&PermitWitnessTransferFromRequest{Contract: contract, Permit: &PermitTransferFrom{Permitted: NewTokenPermissions(token, big.NewInt(5)), Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: NewSignatureTransferDetails(spender, big.NewInt(5)), Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), ""},
		{"permitTransferFromBatch nil", (*PermitTransferFromBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitTransferFromBatch empty", (&PermitTransferFromBatchRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitTransferFromBatch nil contract", (&PermitTransferFromBatchRequest{Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitTransferFromBatch nil permit", (&PermitTransferFromBatchRequest{Contract: contract, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit is nil"},
		{"permitTransferFromBatch empty permitted", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted is empty"},
		{"permitTransferFromBatch length mismatch", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5))}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails length 1 does not match permit.permitted length 2"},
		{"permitTransferFromBatch nil transferDetails", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails length 0 does not match permit.permitted length 2"},
		{"permitTransferFromBatch nil permitted[1] token", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(nil, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted[1].token is nil"},
		{"permitTransferFromBatch nil permitted[0] amount", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, nil), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit.permitted[0].amount is nil"},
		{"permitTransferFromBatch nil transferDetails[1] to", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(nil, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails[1].to is nil"},
		{"permitTransferFromBatch nil transferDetails[0] requestedAmount", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, nil), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails[0].requestedAmount is nil"},
		{"permitTransferFromBatch lower index wins", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(nil, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(nil, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: transferDetails[0].to is nil"},
		{"permitTransferFromBatch nil permit nonce", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit.nonce is nil"},
		{"permitTransferFromBatch nil permit deadline", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), "invalid request: permit.deadline is nil"},
		{"permitTransferFromBatch nil owner", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitTransferFromBatch nil signature", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner}).Validate(), "invalid request: signature is nil"},
		{"permitTransferFromBatch valid", (&PermitTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Signature: sig}).Validate(), ""},
		{"permitWitnessTransferFromBatch nil", (*PermitWitnessTransferFromBatchRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitWitnessTransferFromBatch empty", (&PermitWitnessTransferFromBatchRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitWitnessTransferFromBatch nil contract", (&PermitWitnessTransferFromBatchRequest{Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: contract is nil"},
		{"permitWitnessTransferFromBatch nil permit", (&PermitWitnessTransferFromBatchRequest{Contract: contract, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit is nil"},
		{"permitWitnessTransferFromBatch empty permitted", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted is empty"},
		{"permitWitnessTransferFromBatch length mismatch", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails length 1 does not match permit.permitted length 2"},
		{"permitWitnessTransferFromBatch nil transferDetails", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails length 0 does not match permit.permitted length 2"},
		{"permitWitnessTransferFromBatch nil permitted[1] token", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(nil, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted[1].token is nil"},
		{"permitWitnessTransferFromBatch nil permitted[0] amount", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, nil), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.permitted[0].amount is nil"},
		{"permitWitnessTransferFromBatch nil transferDetails[1] to", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(nil, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails[1].to is nil"},
		{"permitWitnessTransferFromBatch nil transferDetails[0] requestedAmount", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, nil), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails[0].requestedAmount is nil"},
		{"permitWitnessTransferFromBatch lower index wins", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(nil, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(nil, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: transferDetails[0].to is nil"},
		{"permitWitnessTransferFromBatch nil permit nonce", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.nonce is nil"},
		{"permitWitnessTransferFromBatch nil permit deadline", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: permit.deadline is nil"},
		{"permitWitnessTransferFromBatch nil owner", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: owner is nil"},
		{"permitWitnessTransferFromBatch nil witness", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), "invalid request: witness is nil"},
		{"permitWitnessTransferFromBatch empty witnessTypeString", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "", Signature: sig}).Validate(), "invalid request: witnessTypeString is empty"},
		{"permitWitnessTransferFromBatch nil signature", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)"}).Validate(), "invalid request: signature is nil"},
		{"permitWitnessTransferFromBatch valid", (&PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: &PermitBatchTransferFrom{Permitted: []TokenPermissions{*NewTokenPermissions(token, big.NewInt(5)), *NewTokenPermissions(token, big.NewInt(6))}, Nonce: big.NewInt(1), Deadline: big.NewInt(9)}, TransferDetails: []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5)), *NewSignatureTransferDetails(spender, big.NewInt(6))}, Owner: owner, Witness: witness, WitnessTypeString: "Witness witness)Witness(uint256 x)", Signature: sig}).Validate(), ""},
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

func TestNewRequestsSetFields(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	witness := types.NewHashFromBytes([]byte{0x01})
	sig := &types.Signature{}
	permitSingle := NewPermitSingle(NewPermitDetails(token, big.NewInt(1), big.NewInt(2), big.NewInt(3)), spender, big.NewInt(9))
	permitBatch := NewPermitBatch([]PermitDetails{*NewPermitDetails(token, big.NewInt(1), big.NewInt(2), big.NewInt(3))}, spender, big.NewInt(9))
	transferDetails := []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token)}
	approvals := []TokenSpenderPair{*NewTokenSpenderPair(token, spender)}
	permit := NewPermitTransferFrom(NewTokenPermissions(token, big.NewInt(5)), big.NewInt(1), big.NewInt(9))
	batchPermit := NewPermitBatchTransferFrom([]TokenPermissions{*NewTokenPermissions(token, big.NewInt(5))}, big.NewInt(1), big.NewInt(9))
	single := NewSignatureTransferDetails(spender, big.NewInt(5))
	batch := []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(5))}

	require.Equal(t, &DomainSeparatorRequest{Contract: contract}, NewDomainSeparatorRequest(contract))
	require.Equal(t, &AllowanceRequest{Contract: contract, Owner: owner, Token: token, Spender: spender}, NewAllowanceRequest(contract, owner, token, spender))
	require.Equal(t, &NonceBitmapRequest{Contract: contract, Owner: owner, WordPos: big.NewInt(4)}, NewNonceBitmapRequest(contract, owner, big.NewInt(4)))
	require.Equal(t, &ApproveRequest{Contract: contract, Token: token, Spender: spender, Amount: big.NewInt(1), Expiration: big.NewInt(2)}, NewApproveRequest(contract, token, spender, big.NewInt(1), big.NewInt(2)))
	require.Equal(t, &PermitSingleRequest{Contract: contract, Owner: owner, PermitSingle: permitSingle, Signature: sig}, NewPermitSingleRequest(contract, owner, permitSingle, sig))
	require.Equal(t, &PermitBatchRequest{Contract: contract, Owner: owner, PermitBatch: permitBatch, Signature: sig}, NewPermitBatchRequest(contract, owner, permitBatch, sig))
	require.Equal(t, &TransferFromRequest{Contract: contract, From: owner, To: spender, Amount: big.NewInt(1), Token: token}, NewTransferFromRequest(contract, owner, spender, big.NewInt(1), token))
	require.Equal(t, &TransferFromBatchRequest{Contract: contract, TransferDetails: transferDetails}, NewTransferFromBatchRequest(contract, transferDetails))
	require.Equal(t, &LockdownRequest{Contract: contract, Approvals: approvals}, NewLockdownRequest(contract, approvals))
	require.Equal(t, &InvalidateNoncesRequest{Contract: contract, Token: token, Spender: spender, NewNonce: big.NewInt(2)}, NewInvalidateNoncesRequest(contract, token, spender, big.NewInt(2)))
	require.Equal(t, &PermitTransferFromRequest{Contract: contract, Permit: permit, TransferDetails: single, Owner: owner, Signature: sig}, NewPermitTransferFromRequest(contract, permit, single, owner, sig))
	require.Equal(t, &PermitWitnessTransferFromRequest{Contract: contract, Permit: permit, TransferDetails: single, Owner: owner, Witness: witness, WitnessTypeString: "w", Signature: sig}, NewPermitWitnessTransferFromRequest(contract, permit, single, owner, witness, "w", sig))
	require.Equal(t, &PermitTransferFromBatchRequest{Contract: contract, Permit: batchPermit, TransferDetails: batch, Owner: owner, Signature: sig}, NewPermitTransferFromBatchRequest(contract, batchPermit, batch, owner, sig))
	require.Equal(t, &PermitWitnessTransferFromBatchRequest{Contract: contract, Permit: batchPermit, TransferDetails: batch, Owner: owner, Witness: witness, WitnessTypeString: "w", Signature: sig}, NewPermitWitnessTransferFromBatchRequest(contract, batchPermit, batch, owner, witness, "w", sig))
	require.Equal(t, &InvalidateUnorderedNoncesRequest{Contract: contract, WordPos: big.NewInt(4), Mask: big.NewInt(5)}, NewInvalidateUnorderedNoncesRequest(contract, big.NewInt(4), big.NewInt(5)))
}
