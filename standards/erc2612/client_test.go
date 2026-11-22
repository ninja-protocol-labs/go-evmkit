package erc2612

import (
	"context"
	"encoding/json/jsontext"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestReadsSendCallAndDecode(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewCallOptions().WithFrom(owner).WithBlock("0x1b4")
	separator := core.Keccak256([]byte("domain"))

	nonceWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(9))
	require.NoError(t, err)
	separatorWord, err := abi.Pack(abi.NewTypes(abi.Bytes32), separator.Bytes())
	require.NoError(t, err)
	noncesData, err := EncodeNonces(owner)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(e *ERC2612) (any, error)
		want     string
	}{
		{
			"nonces",
			noncesData,
			nonceWord,
			func(e *ERC2612) (any, error) {
				return e.Nonces(t.Context(), opts, NewNoncesRequest(contract, owner))
			},
			"9",
		},
		{
			"domainSeparator",
			EncodeDomainSeparator(),
			separatorWord,
			func(e *ERC2612) (any, error) {
				return e.DomainSeparator(t.Context(), opts, NewDomainSeparatorRequest(contract))
			},
			separator.String(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, contract.String(), call["to"])
				require.Equal(t, owner.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, tc.want, fmt.Sprint(got))
		})
	}
}

func TestReadsRejectMalformedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Twice()
			e := New(c)

			n, err := e.Nonces(t.Context(), nil, NewNoncesRequest(contract, owner))
			require.Error(t, err)
			require.Nil(t, n)
			require.ErrorContains(t, err, "erc2612: nonces: ")

			h, err := e.DomainSeparator(t.Context(), nil, NewDomainSeparatorRequest(contract))
			require.Error(t, err)
			require.Nil(t, h)
			require.ErrorContains(t, err, "erc2612: domainSeparator: ")
		})
	}
}

func TestClientMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	e := New(mocks.NewClient(t))

	n, err := e.Nonces(t.Context(), nil, NewNoncesRequest(contract, nil))
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "erc2612: nonces: "+NewNoncesRequest(contract, nil).Validate().Error())
	require.Nil(t, n)

	h, err := e.DomainSeparator(t.Context(), nil, &DomainSeparatorRequest{})
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "erc2612: domainSeparator: "+(&DomainSeparatorRequest{}).Validate().Error())
	require.Nil(t, h)

	signed, hash, err := e.Permit(t.Context(), nil, NewPermitRequest(contract, contract, contract, big.NewInt(1), big.NewInt(1), nil))
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "erc2612: permit: "+NewPermitRequest(contract, contract, contract, big.NewInt(1), big.NewInt(1), nil).Validate().Error())
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestPermitRelaysSignedPermit(t *testing.T) {
	ownerKey, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	relayerKey, err := types.NewPrivateKeyFromBig(big.NewInt(0xb0b))
	require.NoError(t, err)
	owner := core.PubkeyToAddress(ownerKey.PublicKey())
	relayer := core.PubkeyToAddress(relayerKey.PublicKey())
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount := big.NewInt(12345)
	deadline := big.NewInt(1_900_000_000)

	sig, err := NewPermit(owner, spender, amount, big.NewInt(0), deadline).Sign(eip712.NewDomain("Token", DefaultVersion, big.NewInt(1), contract, nil), ownerKey)
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(4)
	builder.Limit(90_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var raw []byte
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
		raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
		require.NoError(t, err)
		*e.Result.(*string) = core.Keccak256(raw).String()
		return nil
	}).Once()

	signed, hash, err := New(c).Permit(t.Context(), bind.NewTransactOptions(relayerKey, builder), NewPermitRequest(contract, owner, spender, amount, deadline, sig))
	require.NoError(t, err)
	require.True(t, hash.Equal(core.Keccak256(raw)))
	require.Equal(t, uint64(4), signed.AddressNonce())

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	dyn, ok := decoded.(*core.DynamicFeeTx)
	require.True(t, ok)
	require.True(t, dyn.To.Equal(contract))
	require.Zero(t, dyn.Value.Sign())
	from, err := decoded.Sender()
	require.NoError(t, err)
	require.True(t, from.Equal(relayer))
	require.False(t, from.Equal(owner))

	vals, err := permitFn.DecodeCall(dyn.Data)
	require.NoError(t, err)
	require.True(t, vals[0].(*types.Address).Equal(owner))
	require.True(t, vals[1].(*types.Address).Equal(spender))
	require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
	require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))
	require.Equal(t, uint8(sig.LegacyV().Uint64()), vals[4])
	require.Contains(t, []uint8{27, 28}, vals[4])
	require.Equal(t, sig.R().FillBytes(make([]byte, 32)), vals[5])
	require.Equal(t, sig.S().FillBytes(make([]byte, 32)), vals[6])
}

func TestPermitMapsEstimateRevertToSentinel(t *testing.T) {
	ownerKey, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	owner := core.PubkeyToAddress(ownerKey.PublicKey())
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	sig, err := NewPermit(owner, spender, big.NewInt(1), big.NewInt(0), big.NewInt(1)).Sign(eip712.NewDomain("Token", DefaultVersion, big.NewInt(1), contract, nil), ownerKey)
	require.NoError(t, err)

	expiredArgs, err := abi.Pack(errExpiredSignature.Inputs, big.NewInt(1))
	require.NoError(t, err)
	expiredSel := errExpiredSignature.Selector()
	invalidArgs, err := abi.Pack(errInvalidSigner.Inputs, spender, owner)
	require.NoError(t, err)
	invalidSel := errInvalidSigner.Selector()

	for _, tc := range []struct {
		name   string
		revert []byte
		want   error
		wantNo error
	}{
		{"expired signature", append(append([]byte{}, expiredSel[:]...), expiredArgs...), ErrExpiredSignature, ErrInvalidSigner},
		{"invalid signer", append(append([]byte{}, invalidSel[:]...), invalidArgs...), ErrInvalidSigner, ErrExpiredSignature},
		{"unknown selector", []byte{0xde, 0xad, 0xbe, 0xef}, ErrUnknown, ErrInvalidSigner},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(4)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			c := mocks.NewClient(t)
			c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
				require.Equal(t, 1, elems.Len())
				require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
				return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(tc.revert) + `"`)})
			}).Once()

			signed, hash, err := New(c).Permit(t.Context(), bind.NewTransactOptions(ownerKey, builder), NewPermitRequest(contract, owner, spender, big.NewInt(1), big.NewInt(1), sig))
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, tc.wantNo)
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
			require.Nil(t, signed)
			require.Nil(t, hash)
		})
	}
}

func TestPermitReturnsSignedTxOnBroadcastFailure(t *testing.T) {
	ownerKey, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	owner := core.PubkeyToAddress(ownerKey.PublicKey())
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	sig, err := NewPermit(owner, spender, big.NewInt(1), big.NewInt(0), big.NewInt(1)).Sign(eip712.NewDomain("Token", DefaultVersion, big.NewInt(1), contract, nil), ownerKey)
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(11)
	builder.Limit(90_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Once()

	signed, hash, err := New(c).Permit(t.Context(), bind.NewTransactOptions(ownerKey, builder), NewPermitRequest(contract, owner, spender, big.NewInt(1), big.NewInt(1), sig))
	require.ErrorContains(t, err, "erc2612: permit: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(11), signed.AddressNonce())
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "erc2612")
	contract := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	nonceWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "erc2612", got.Value(marker{}))
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(nonceWord)
		return nil
	}).Once()

	_, err = New(c).Nonces(ctx, nil, NewNoncesRequest(contract, owner))
	require.NoError(t, err)
}
