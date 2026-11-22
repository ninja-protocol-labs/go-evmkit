package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPairCreatedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token1 := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	allPairsLength := big.NewInt(7)

	pc := NewPairCreatedEvent(contract, token0, token1, pair, allPairsLength)
	require.Same(t, contract, pc.Contract)
	require.Same(t, token0, pc.Token0)
	require.Same(t, token1, pc.Token1)
	require.Same(t, pair, pc.Pair)
	require.Same(t, allPairsLength, pc.AllPairsLength)
}

func TestFactoryRequestsValidate(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"feeTo nil", (*FeeToRequest)(nil).Validate(), "invalid request: request is nil"},
		{"feeTo empty", (&FeeToRequest{}).Validate(), "invalid request: contract is nil"},
		{"feeTo nil contract", (&FeeToRequest{}).Validate(), "invalid request: contract is nil"},
		{"feeTo valid", (&FeeToRequest{Contract: factory}).Validate(), ""},
		{"feeToSetter nil", (*FeeToSetterRequest)(nil).Validate(), "invalid request: request is nil"},
		{"feeToSetter empty", (&FeeToSetterRequest{}).Validate(), "invalid request: contract is nil"},
		{"feeToSetter nil contract", (&FeeToSetterRequest{}).Validate(), "invalid request: contract is nil"},
		{"feeToSetter valid", (&FeeToSetterRequest{Contract: factory}).Validate(), ""},
		{"allPairsLength nil", (*AllPairsLengthRequest)(nil).Validate(), "invalid request: request is nil"},
		{"allPairsLength empty", (&AllPairsLengthRequest{}).Validate(), "invalid request: contract is nil"},
		{"allPairsLength nil contract", (&AllPairsLengthRequest{}).Validate(), "invalid request: contract is nil"},
		{"allPairsLength valid", (&AllPairsLengthRequest{Contract: factory}).Validate(), ""},
		{"getPair nil", (*GetPairRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPair empty", (&GetPairRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPair nil contract", (&GetPairRequest{TokenA: a, TokenB: b}).Validate(), "invalid request: contract is nil"},
		{"getPair nil tokenA", (&GetPairRequest{Contract: factory, TokenB: b}).Validate(), "invalid request: tokenA is nil"},
		{"getPair nil tokenB", (&GetPairRequest{Contract: factory, TokenA: a}).Validate(), "invalid request: tokenB is nil"},
		{"getPair valid", (&GetPairRequest{Contract: factory, TokenA: a, TokenB: b}).Validate(), ""},
		{"allPairs nil", (*AllPairsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"allPairs empty", (&AllPairsRequest{}).Validate(), "invalid request: contract is nil"},
		{"allPairs nil contract", (&AllPairsRequest{Index: big.NewInt(0)}).Validate(), "invalid request: contract is nil"},
		{"allPairs nil index", (&AllPairsRequest{Contract: factory}).Validate(), "invalid request: index is nil"},
		{"allPairs valid", (&AllPairsRequest{Contract: factory, Index: big.NewInt(0)}).Validate(), ""},
		{"createPair nil", (*CreatePairRequest)(nil).Validate(), "invalid request: request is nil"},
		{"createPair empty", (&CreatePairRequest{}).Validate(), "invalid request: contract is nil"},
		{"createPair nil contract", (&CreatePairRequest{TokenA: a, TokenB: b}).Validate(), "invalid request: contract is nil"},
		{"createPair nil tokenA", (&CreatePairRequest{Contract: factory, TokenB: b}).Validate(), "invalid request: tokenA is nil"},
		{"createPair nil tokenB", (&CreatePairRequest{Contract: factory, TokenA: a}).Validate(), "invalid request: tokenB is nil"},
		{"createPair valid", (&CreatePairRequest{Contract: factory, TokenA: a, TokenB: b}).Validate(), ""},
		{"setFeeTo nil", (*SetFeeToRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setFeeTo empty", (&SetFeeToRequest{}).Validate(), "invalid request: contract is nil"},
		{"setFeeTo nil contract", (&SetFeeToRequest{FeeTo: a}).Validate(), "invalid request: contract is nil"},
		{"setFeeTo nil feeTo", (&SetFeeToRequest{Contract: factory}).Validate(), "invalid request: feeTo is nil"},
		{"setFeeTo valid", (&SetFeeToRequest{Contract: factory, FeeTo: a}).Validate(), ""},
		{"setFeeToSetter nil", (*SetFeeToSetterRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setFeeToSetter empty", (&SetFeeToSetterRequest{}).Validate(), "invalid request: contract is nil"},
		{"setFeeToSetter nil contract", (&SetFeeToSetterRequest{FeeToSetter: a}).Validate(), "invalid request: contract is nil"},
		{"setFeeToSetter nil feeToSetter", (&SetFeeToSetterRequest{Contract: factory}).Validate(), "invalid request: feeToSetter is nil"},
		{"setFeeToSetter valid", (&SetFeeToSetterRequest{Contract: factory, FeeToSetter: a}).Validate(), ""},
		{"getPair identical tokens", (&GetPairRequest{Contract: factory, TokenA: a, TokenB: a}).Validate(), ""},
		{"createPair zero token", (&CreatePairRequest{Contract: factory, TokenA: a, TokenB: types.ZeroAddress()}).Validate(), ""},
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

func TestNewFactoryRequestsSetFields(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	require.Equal(t, &FeeToRequest{Contract: factory}, NewFeeToRequest(factory))
	require.Equal(t, &FeeToSetterRequest{Contract: factory}, NewFeeToSetterRequest(factory))
	require.Equal(t, &GetPairRequest{Contract: factory, TokenA: a, TokenB: b}, NewGetPairRequest(factory, a, b))
	require.Equal(t, &AllPairsRequest{Contract: factory, Index: big.NewInt(3)}, NewAllPairsRequest(factory, big.NewInt(3)))
	require.Equal(t, &AllPairsLengthRequest{Contract: factory}, NewAllPairsLengthRequest(factory))
	require.Equal(t, &CreatePairRequest{Contract: factory, TokenA: a, TokenB: b}, NewCreatePairRequest(factory, a, b))
	require.Equal(t, &SetFeeToRequest{Contract: factory, FeeTo: a}, NewSetFeeToRequest(factory, a))
	require.Equal(t, &SetFeeToSetterRequest{Contract: factory, FeeToSetter: b}, NewSetFeeToSetterRequest(factory, b))
}
