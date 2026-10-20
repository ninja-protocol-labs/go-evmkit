package permit2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testOwner   = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	testUser    = testOwner
	testSpender = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	testToken   = types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	testTokenB  = types.MustNewAddressFromHex("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f")
	testTo      = testTokenB
)

func testPermit2Signature(t *testing.T) *types.Signature {
	t.Helper()
	key, err := types.GeneratePrivateKey()
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
	require.NoError(t, err)
	return sig
}

func TestEncodeDomainSeparator(t *testing.T) {
	require.Equal(t, []byte{0x36, 0x44, 0xe5, 0x15}, EncodeDomainSeparator())
}

func TestDecodeDomainSeparator(t *testing.T) {
	want := types.NewHashFromBytes([]byte("some 32 byte digest padded out!"))
	data, err := abi.Pack(abi.Types{abi.Bytes32}, want.Bytes())
	require.NoError(t, err)

	got, err := DecodeDomainSeparator(data)
	require.NoError(t, err)
	require.Equal(t, want.Bytes(), got.Bytes())
}

func TestEncodeAllowance(t *testing.T) {
	data, err := EncodeAllowance(testUser, testToken, testSpender)
	require.NoError(t, err)
	require.Equal(t, []byte{0x92, 0x7d, 0xa1, 0x05}, data[:4])

	vals, err := allowanceFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testUser))
	require.True(t, vals[1].(*types.Address).Equal(testToken))
	require.True(t, vals[2].(*types.Address).Equal(testSpender))
}

func TestDecodeAllowance(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Uint48, abi.Uint48}, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	require.NoError(t, err)

	got, err := DecodeAllowance(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(100).Cmp(got.Amount))
	require.Zero(t, big.NewInt(200).Cmp(got.Expiration))
	require.Zero(t, big.NewInt(1).Cmp(got.Nonce))
}

func TestEncodeApprove(t *testing.T) {
	data, err := EncodeApprove(testToken, testSpender, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)
	require.Equal(t, []byte{0x87, 0x51, 0x7c, 0x45}, data[:4])

	vals, err := approveFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testToken))
	require.True(t, vals[1].(*types.Address).Equal(testSpender))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(vals[3].(*big.Int)))
}

func TestEncodePermitSingle(t *testing.T) {
	details := NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	permitSingle := NewPermitSingle(details, testSpender, big.NewInt(9999999999))
	sig := testPermit2Signature(t)

	data, err := EncodePermitSingle(testOwner, permitSingle, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x2b, 0x67, 0xb5, 0x70}, data[:4])

	vals, err := permitSingleFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))

	got := vals[1].([]any)
	gotDetails := got[0].([]any)
	require.True(t, gotDetails[0].(*types.Address).Equal(testToken))
	require.Zero(t, big.NewInt(100).Cmp(gotDetails[1].(*big.Int)))
	require.Zero(t, big.NewInt(200).Cmp(gotDetails[2].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(gotDetails[3].(*big.Int)))
	require.True(t, got[1].(*types.Address).Equal(testSpender))
	require.Zero(t, big.NewInt(9999999999).Cmp(got[2].(*big.Int)))

	require.Equal(t, sig.LegacyBytes(), vals[2].([]byte))
}

func TestEncodePermitBatch(t *testing.T) {
	details := []PermitDetails{
		*NewPermitDetails(testToken, big.NewInt(100), big.NewInt(200), big.NewInt(1)),
		*NewPermitDetails(testTokenB, big.NewInt(300), big.NewInt(400), big.NewInt(2)),
	}
	permitBatch := NewPermitBatch(details, testSpender, big.NewInt(9999999999))
	sig := testPermit2Signature(t)

	data, err := EncodePermitBatch(testOwner, permitBatch, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x2a, 0x2d, 0x80, 0xd1}, data[:4])

	vals, err := permitBatchFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))

	got := vals[1].([]any)
	gotDetails := got[0].([]any)
	require.Len(t, gotDetails, 2)
	first := gotDetails[0].([]any)
	require.True(t, first[0].(*types.Address).Equal(testToken))
	second := gotDetails[1].([]any)
	require.True(t, second[0].(*types.Address).Equal(testTokenB))
	require.True(t, got[1].(*types.Address).Equal(testSpender))
}

func TestEncodeTransferFrom(t *testing.T) {
	data, err := EncodeTransferFrom(testOwner, testTo, big.NewInt(100), testToken)
	require.NoError(t, err)
	require.Equal(t, []byte{0x36, 0xc7, 0x85, 0x16}, data[:4])

	vals, err := transferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))
	require.True(t, vals[1].(*types.Address).Equal(testTo))
	require.Zero(t, big.NewInt(100).Cmp(vals[2].(*big.Int)))
	require.True(t, vals[3].(*types.Address).Equal(testToken))
}

func TestEncodeTransferFromBatch(t *testing.T) {
	details := []AllowanceTransferDetails{
		*NewAllowanceTransferDetails(testOwner, testTo, big.NewInt(100), testToken),
	}

	data, err := EncodeTransferFromBatch(details)
	require.NoError(t, err)
	require.Equal(t, []byte{0x0d, 0x58, 0xb1, 0xdb}, data[:4])

	vals, err := transferFromBatchFn.DecodeCall(data)
	require.NoError(t, err)
	got := vals[0].([]any)
	require.Len(t, got, 1)
	first := got[0].([]any)
	require.True(t, first[0].(*types.Address).Equal(testOwner))
	require.True(t, first[1].(*types.Address).Equal(testTo))
	require.Zero(t, big.NewInt(100).Cmp(first[2].(*big.Int)))
	require.True(t, first[3].(*types.Address).Equal(testToken))
}

func TestEncodeLockdown(t *testing.T) {
	approvals := []TokenSpenderPair{
		*NewTokenSpenderPair(testToken, testSpender),
	}

	data, err := EncodeLockdown(approvals)
	require.NoError(t, err)
	require.Equal(t, []byte{0xcc, 0x53, 0x28, 0x7f}, data[:4])

	vals, err := lockdownFn.DecodeCall(data)
	require.NoError(t, err)
	got := vals[0].([]any)
	require.Len(t, got, 1)
	first := got[0].([]any)
	require.True(t, first[0].(*types.Address).Equal(testToken))
	require.True(t, first[1].(*types.Address).Equal(testSpender))
}

func TestEncodeInvalidateNonces(t *testing.T) {
	data, err := EncodeInvalidateNonces(testToken, testSpender, big.NewInt(5))
	require.NoError(t, err)
	require.Equal(t, []byte{0x65, 0xd9, 0x72, 0x3c}, data[:4])

	vals, err := invalidateNoncesFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testToken))
	require.True(t, vals[1].(*types.Address).Equal(testSpender))
	require.Zero(t, big.NewInt(5).Cmp(vals[2].(*big.Int)))
}

func TestAllowanceTransferFunctionSelectors(t *testing.T) {
	tests := []struct {
		name string
		fn   *abi.Function
		sig  string
		sel  []byte
	}{
		{
			"DOMAIN_SEPARATOR",
			domainSeparatorFn,
			"DOMAIN_SEPARATOR()",
			[]byte{0x36, 0x44, 0xe5, 0x15},
		},
		{
			"allowance",
			allowanceFn,
			"allowance(address,address,address)",
			[]byte{0x92, 0x7d, 0xa1, 0x05},
		},
		{
			"approve",
			approveFn,
			"approve(address,address,uint160,uint48)",
			[]byte{0x87, 0x51, 0x7c, 0x45},
		},
		{
			"permit(single)",
			permitSingleFn,
			"permit(address,((address,uint160,uint48,uint48),address,uint256),bytes)",
			[]byte{0x2b, 0x67, 0xb5, 0x70},
		},
		{
			"permit(batch)",
			permitBatchFn,
			"permit(address,((address,uint160,uint48,uint48)[],address,uint256),bytes)",
			[]byte{0x2a, 0x2d, 0x80, 0xd1},
		},
		{
			"transferFrom(single)",
			transferFromFn,
			"transferFrom(address,address,uint160,address)",
			[]byte{0x36, 0xc7, 0x85, 0x16},
		},
		{
			"transferFrom(batch)",
			transferFromBatchFn,
			"transferFrom((address,address,uint160,address)[])",
			[]byte{0x0d, 0x58, 0xb1, 0xdb},
		},
		{
			"lockdown",
			lockdownFn,
			"lockdown((address,address)[])",
			[]byte{0xcc, 0x53, 0x28, 0x7f},
		},
		{
			"invalidateNonces",
			invalidateNoncesFn,
			"invalidateNonces(address,address,uint48)",
			[]byte{0x65, 0xd9, 0x72, 0x3c},
		},
	}

	seen := map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.sig, tt.fn.Signature())
			require.Equal(t, tt.sel, tt.fn.SelectorBytes())
		})
		sel := string(tt.sel)
		require.Empty(t, seen[sel], "selector collision between %s and %s", tt.name, seen[sel])
		seen[sel] = tt.name
	}
}

func TestDecodeDomainSeparatorRejectsGarbage(t *testing.T) {
	_, err := DecodeDomainSeparator([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestDecodeAllowanceRejectsGarbage(t *testing.T) {
	_, err := DecodeAllowance([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodeApproveRejectsUint160Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 160)
	_, err := EncodeApprove(testToken, testSpender, overflow, big.NewInt(1))
	require.Error(t, err)
}

func TestEncodeApproveRejectsUint48Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 48)
	_, err := EncodeApprove(testToken, testSpender, big.NewInt(1), overflow)
	require.Error(t, err)
}

func TestEncodeApproveAcceptsMaxUint160AndUint48(t *testing.T) {
	maxUint160 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1))
	maxUint48 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 48), big.NewInt(1))

	data, err := EncodeApprove(testToken, testSpender, maxUint160, maxUint48)
	require.NoError(t, err)

	vals, err := approveFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, maxUint160.Cmp(vals[2].(*big.Int)))
	require.Zero(t, maxUint48.Cmp(vals[3].(*big.Int)))
}

func TestEncodePermitSingleRejectsUint48Overflow(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 48)
	details := NewPermitDetails(testToken, big.NewInt(100), overflow, big.NewInt(1))
	permitSingle := NewPermitSingle(details, testSpender, big.NewInt(9999999999))

	_, err := EncodePermitSingle(testOwner, permitSingle, testPermit2Signature(t))
	require.Error(t, err)
}

func TestEncodePermitBatchEmptyDetails(t *testing.T) {
	permitBatch := NewPermitBatch(nil, testSpender, big.NewInt(9999999999))

	data, err := EncodePermitBatch(testOwner, permitBatch, testPermit2Signature(t))
	require.NoError(t, err)

	vals, err := permitBatchFn.DecodeCall(data)
	require.NoError(t, err)
	got := vals[1].([]any)
	require.Empty(t, got[0].([]any))
}

func TestEncodeTransferFromBatchEmpty(t *testing.T) {
	data, err := EncodeTransferFromBatch(nil)
	require.NoError(t, err)

	vals, err := transferFromBatchFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]any))
}

func TestEncodeLockdownEmpty(t *testing.T) {
	data, err := EncodeLockdown(nil)
	require.NoError(t, err)

	vals, err := lockdownFn.DecodeCall(data)
	require.NoError(t, err)
	require.Empty(t, vals[0].([]any))
}

func TestEncodeNonceBitmap(t *testing.T) {
	data, err := EncodeNonceBitmap(testOwner, big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, []byte{0x4f, 0xe0, 0x2b, 0x44}, data[:4])

	vals, err := nonceBitmapFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(testOwner))
	require.Zero(t, big.NewInt(1).Cmp(vals[1].(*big.Int)))
}

func TestDecodeNonceBitmap(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(12345))
	require.NoError(t, err)

	got, err := DecodeNonceBitmap(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(12345).Cmp(got))
}

func TestDecodeNonceBitmapRejectsGarbage(t *testing.T) {
	_, err := DecodeNonceBitmap([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncodePermitTransferFrom(t *testing.T) {
	permitted := NewTokenPermissions(testToken, big.NewInt(100))
	permit := NewPermitTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	details := NewSignatureTransferDetails(testTo, big.NewInt(50))
	sig := testPermit2Signature(t)

	data, err := EncodePermitTransferFrom(permit, details, testOwner, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x30, 0xf2, 0x8b, 0x7a}, data[:4])

	vals, err := permitTransferFromFn.DecodeCall(data)
	require.NoError(t, err)

	gotPermit := vals[0].([]any)
	gotPermitted := gotPermit[0].([]any)
	require.True(t, gotPermitted[0].(*types.Address).Equal(testToken))
	require.Zero(t, big.NewInt(100).Cmp(gotPermitted[1].(*big.Int)))
	require.Zero(t, big.NewInt(1).Cmp(gotPermit[1].(*big.Int)))
	require.Zero(t, big.NewInt(9999999999).Cmp(gotPermit[2].(*big.Int)))

	gotDetails := vals[1].([]any)
	require.True(t, gotDetails[0].(*types.Address).Equal(testTo))
	require.Zero(t, big.NewInt(50).Cmp(gotDetails[1].(*big.Int)))

	require.True(t, vals[2].(*types.Address).Equal(testOwner))
	require.Equal(t, sig.LegacyBytes(), vals[3].([]byte))
}

func TestEncodePermitWitnessTransferFrom(t *testing.T) {
	permitted := NewTokenPermissions(testToken, big.NewInt(100))
	permit := NewPermitTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	details := NewSignatureTransferDetails(testTo, big.NewInt(50))
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))
	witnessTypeString := "Witness witness)Witness(uint256 x)"
	sig := testPermit2Signature(t)

	data, err := EncodePermitWitnessTransferFrom(permit, details, testOwner, witness, witnessTypeString, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0x13, 0x7c, 0x29, 0xfe}, data[:4])

	vals, err := permitWitnessTransferFromFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[2].(*types.Address).Equal(testOwner))
	require.Equal(t, witness.Bytes(), vals[3].([]byte))
	require.Equal(t, witnessTypeString, vals[4].(string))
	require.Equal(t, sig.LegacyBytes(), vals[5].([]byte))
}

func TestEncodePermitTransferFromBatch(t *testing.T) {
	permitted := []TokenPermissions{
		*NewTokenPermissions(testToken, big.NewInt(100)),
		*NewTokenPermissions(testTokenB, big.NewInt(200)),
	}
	permit := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	details := []SignatureTransferDetails{
		*NewSignatureTransferDetails(testTo, big.NewInt(50)),
		*NewSignatureTransferDetails(testOwner, big.NewInt(60)),
	}
	sig := testPermit2Signature(t)

	data, err := EncodePermitTransferFromBatch(permit, details, testOwner, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xed, 0xd9, 0x44, 0x4b}, data[:4])

	vals, err := permitTransferFromBatchFn.DecodeCall(data)
	require.NoError(t, err)

	gotPermit := vals[0].([]any)
	gotPermitted := gotPermit[0].([]any)
	require.Len(t, gotPermitted, 2)

	gotDetails := vals[1].([]any)
	require.Len(t, gotDetails, 2)
	first := gotDetails[0].([]any)
	require.True(t, first[0].(*types.Address).Equal(testTo))
}

func TestEncodePermitWitnessTransferFromBatch(t *testing.T) {
	permitted := []TokenPermissions{*NewTokenPermissions(testToken, big.NewInt(100))}
	permit := NewPermitBatchTransferFrom(permitted, big.NewInt(1), big.NewInt(9999999999))
	details := []SignatureTransferDetails{*NewSignatureTransferDetails(testTo, big.NewInt(50))}
	witness := types.NewHashFromBytes([]byte("some witness data, 32 bytes!!!!"))
	witnessTypeString := "Witness witness)Witness(uint256 x)"
	sig := testPermit2Signature(t)

	data, err := EncodePermitWitnessTransferFromBatch(permit, details, testOwner, witness, witnessTypeString, sig)
	require.NoError(t, err)
	require.Equal(t, []byte{0xfe, 0x8e, 0xc1, 0xa7}, data[:4])

	vals, err := permitWitnessTransferFromBatchFn.DecodeCall(data)
	require.NoError(t, err)
	require.True(t, vals[2].(*types.Address).Equal(testOwner))
	require.Equal(t, witness.Bytes(), vals[3].([]byte))
	require.Equal(t, witnessTypeString, vals[4].(string))
}

func TestEncodeInvalidateUnorderedNonces(t *testing.T) {
	data, err := EncodeInvalidateUnorderedNonces(big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, []byte{0x3f, 0xf9, 0xdc, 0xb1}, data[:4])

	vals, err := invalidateUnorderedNoncesFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1).Cmp(vals[0].(*big.Int)))
	require.Zero(t, big.NewInt(2).Cmp(vals[1].(*big.Int)))
}

func TestEncodePermitTransferFromBatchEmptyPermitted(t *testing.T) {
	permit := NewPermitBatchTransferFrom(nil, big.NewInt(1), big.NewInt(9999999999))
	data, err := EncodePermitTransferFromBatch(permit, nil, testOwner, testPermit2Signature(t))
	require.NoError(t, err)

	vals, err := permitTransferFromBatchFn.DecodeCall(data)
	require.NoError(t, err)
	gotPermit := vals[0].([]any)
	require.Empty(t, gotPermit[0].([]any))
	require.Empty(t, vals[1].([]any))
}

func TestSignatureTransferFunctionSelectors(t *testing.T) {
	tests := []struct {
		name string
		fn   *abi.Function
		sig  string
		sel  []byte
	}{
		{
			"nonceBitmap",
			nonceBitmapFn,
			"nonceBitmap(address,uint256)",
			[]byte{0x4f, 0xe0, 0x2b, 0x44},
		},
		{
			"permitTransferFrom",
			permitTransferFromFn,
			"permitTransferFrom(((address,uint256),uint256,uint256),(address,uint256),address,bytes)",
			[]byte{0x30, 0xf2, 0x8b, 0x7a},
		},
		{
			"permitWitnessTransferFrom",
			permitWitnessTransferFromFn,
			"permitWitnessTransferFrom(((address,uint256),uint256,uint256),(address,uint256),address,bytes32,string,bytes)",
			[]byte{0x13, 0x7c, 0x29, 0xfe},
		},
		{
			"permitTransferFrom(batch)",
			permitTransferFromBatchFn,
			"permitTransferFrom(((address,uint256)[],uint256,uint256),(address,uint256)[],address,bytes)",
			[]byte{0xed, 0xd9, 0x44, 0x4b},
		},
		{
			"permitWitnessTransferFrom(batch)",
			permitWitnessTransferFromBatchFn,
			"permitWitnessTransferFrom(((address,uint256)[],uint256,uint256),(address,uint256)[],address,bytes32,string,bytes)",
			[]byte{0xfe, 0x8e, 0xc1, 0xa7},
		},
		{
			"invalidateUnorderedNonces",
			invalidateUnorderedNoncesFn,
			"invalidateUnorderedNonces(uint256,uint256)",
			[]byte{0x3f, 0xf9, 0xdc, 0xb1},
		},
	}

	seen := map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.sig, tt.fn.Signature())
			require.Equal(t, tt.sel, tt.fn.SelectorBytes())
		})
		sel := string(tt.sel)
		require.Empty(t, seen[sel], "selector collision between %s and %s", tt.name, seen[sel])
		seen[sel] = tt.name
	}
}
