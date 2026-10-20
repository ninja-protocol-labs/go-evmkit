package permit2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestNewNonceInvalidation(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	newNonce := big.NewInt(2)
	oldNonce := big.NewInt(1)

	n := NewNonceInvalidation(contract, testOwner, testToken, testSpender, newNonce, oldNonce)
	require.Same(t, contract, n.Contract)
	require.Same(t, testOwner, n.Owner)
	require.Same(t, testToken, n.Token)
	require.Same(t, testSpender, n.Spender)
	require.Same(t, newNonce, n.NewNonce)
	require.Same(t, oldNonce, n.OldNonce)
}

func TestNewApproval(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(100)
	expiration := big.NewInt(200)

	a := NewApproval(contract, testOwner, testToken, testSpender, amount, expiration)
	require.Same(t, contract, a.Contract)
	require.Same(t, testOwner, a.Owner)
	require.Same(t, testToken, a.Token)
	require.Same(t, testSpender, a.Spender)
	require.Same(t, amount, a.Amount)
	require.Same(t, expiration, a.Expiration)
}

func TestNewPermit(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(100)
	expiration := big.NewInt(200)
	nonce := big.NewInt(1)

	p := NewPermit(contract, testOwner, testToken, testSpender, amount, expiration, nonce)
	require.Same(t, contract, p.Contract)
	require.Same(t, testOwner, p.Owner)
	require.Same(t, testToken, p.Token)
	require.Same(t, testSpender, p.Spender)
	require.Same(t, amount, p.Amount)
	require.Same(t, expiration, p.Expiration)
	require.Same(t, nonce, p.Nonce)
}

func TestNewLockdown(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	l := NewLockdown(contract, testOwner, testToken, testSpender)
	require.Same(t, contract, l.Contract)
	require.Same(t, testOwner, l.Owner)
	require.Same(t, testToken, l.Token)
	require.Same(t, testSpender, l.Spender)
}

func TestNewUnorderedNonceInvalidation(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	word := big.NewInt(1)
	mask := big.NewInt(2)

	n := NewUnorderedNonceInvalidation(contract, testOwner, word, mask)
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
