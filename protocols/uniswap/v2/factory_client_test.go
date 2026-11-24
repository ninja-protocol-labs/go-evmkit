package v2

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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestFactoryReadsSendCallAndDecode(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")

	pairWord, err := abi.Pack(abi.NewTypes(abi.Address), pair)
	require.NoError(t, err)
	aWord, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)
	lengthWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(528403))
	require.NoError(t, err)

	getPairData, err := EncodeGetPair(a, b)
	require.NoError(t, err)
	allPairsData, err := EncodeAllPairs(big.NewInt(9))
	require.NoError(t, err)
	createPairData, err := EncodeCreatePair(a, b)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(f *Factory) (any, error)
		want     any
	}{
		{
			"feeTo",
			EncodeFeeTo(),
			aWord,
			func(f *Factory) (any, error) {
				return f.FeeTo(t.Context(), opts, NewFeeToRequest(factory))
			},
			a,
		},
		{
			"feeToSetter",
			EncodeFeeToSetter(),
			aWord,
			func(f *Factory) (any, error) {
				return f.FeeToSetter(t.Context(), opts, NewFeeToSetterRequest(factory))
			},
			a,
		},
		{
			"getPair",
			getPairData,
			pairWord,
			func(f *Factory) (any, error) {
				return f.GetPair(t.Context(), opts, NewGetPairRequest(factory, a, b))
			},
			pair,
		},
		{
			"allPairs",
			allPairsData,
			pairWord,
			func(f *Factory) (any, error) {
				return f.AllPairs(t.Context(), opts, NewAllPairsRequest(factory, big.NewInt(9)))
			},
			pair,
		},
		{
			"allPairsLength",
			EncodeAllPairsLength(),
			lengthWord,
			func(f *Factory) (any, error) {
				return f.AllPairsLength(t.Context(), opts, NewAllPairsLengthRequest(factory))
			},
			big.NewInt(528403),
		},
		{
			"callCreatePair",
			createPairData,
			pairWord,
			func(f *Factory) (any, error) {
				return f.CallCreatePair(t.Context(), opts, NewCreatePairRequest(factory, a, b))
			},
			pair,
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
				require.Equal(t, factory.String(), call["to"])
				require.Equal(t, a.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewFactory(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestFactoryReadsMapRevertToSentinel(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"identical addresses", reasonIdenticalAddresses, ErrIdenticalAddresses},
		{"zero address", reasonZeroAddress, ErrZeroAddress},
		{"pair exists", reasonPairExists, ErrPairExists},
		{"forbidden", reasonForbidden, ErrForbidden},
		{"other reason", "UniswapV2: SOMETHING_ELSE", ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Once()

			pair, err := NewFactory(c).CallCreatePair(t.Context(), bind.NewCallOptions().WithFrom(a), NewCreatePairRequest(factory, a, a))
			require.Nil(t, pair)
			require.ErrorIs(t, err, tc.want)
			for _, other := range []error{ErrIdenticalAddresses, ErrZeroAddress, ErrPairExists, ErrForbidden, ErrUnknown} {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`), jsontext.Value(`"0xzz"`)} {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		n, err := NewFactory(c).AllPairsLength(t.Context(), nil, NewAllPairsLengthRequest(factory))
		require.Nil(t, n)
		require.ErrorContains(t, err, "uniswapv2: allPairsLength: ")
		require.NotErrorIs(t, err, ErrUnknown)
	}
}

func TestFactoryReadsRejectMalformedResult(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(3)
			f := NewFactory(c)

			pair, err := f.GetPair(t.Context(), nil, NewGetPairRequest(factory, a, a))
			require.Nil(t, pair)
			require.ErrorContains(t, err, "uniswapv2: getPair: ")

			n, err := f.AllPairsLength(t.Context(), nil, NewAllPairsLengthRequest(factory))
			require.Nil(t, n)
			require.ErrorContains(t, err, "uniswapv2: allPairsLength: ")

			created, err := f.CallCreatePair(t.Context(), nil, NewCreatePairRequest(factory, a, a))
			require.Nil(t, created)
			require.ErrorContains(t, err, "uniswapv2: callCreatePair: ")
		})
	}
}

func TestFactoryMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	f := NewFactory(mocks.NewClient(t))
	ctx := t.Context()

	_, err = f.FeeTo(ctx, nil, &FeeToRequest{})
	require.EqualError(t, err, "uniswapv2: feeTo: "+(&FeeToRequest{}).Validate().Error())
	_, err = f.FeeToSetter(ctx, nil, NewFeeToSetterRequest(nil))
	require.EqualError(t, err, "uniswapv2: feeToSetter: "+NewFeeToSetterRequest(nil).Validate().Error())
	_, err = f.GetPair(ctx, nil, NewGetPairRequest(factory, a, nil))
	require.EqualError(t, err, "uniswapv2: getPair: "+NewGetPairRequest(factory, a, nil).Validate().Error())
	_, err = f.AllPairs(ctx, nil, NewAllPairsRequest(factory, nil))
	require.EqualError(t, err, "uniswapv2: allPairs: "+NewAllPairsRequest(factory, nil).Validate().Error())
	_, err = f.AllPairsLength(ctx, nil, &AllPairsLengthRequest{})
	require.EqualError(t, err, "uniswapv2: allPairsLength: "+(&AllPairsLengthRequest{}).Validate().Error())
	_, err = f.CallCreatePair(ctx, nil, NewCreatePairRequest(factory, nil, a))
	require.EqualError(t, err, "uniswapv2: callCreatePair: "+NewCreatePairRequest(factory, nil, a).Validate().Error())

	_, _, err = f.CreatePair(ctx, opts, NewCreatePairRequest(nil, a, a))
	require.EqualError(t, err, "uniswapv2: createPair: "+NewCreatePairRequest(nil, a, a).Validate().Error())
	_, _, err = f.SetFeeTo(ctx, opts, NewSetFeeToRequest(factory, nil))
	require.EqualError(t, err, "uniswapv2: setFeeTo: "+NewSetFeeToRequest(factory, nil).Validate().Error())
	_, _, err = f.SetFeeToSetter(ctx, opts, &SetFeeToSetterRequest{})
	require.EqualError(t, err, "uniswapv2: setFeeToSetter: "+(&SetFeeToSetterRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = f.AllPairs(ctx, nil, NewAllPairsRequest(factory, big.NewInt(-1)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = f.CreatePair(ctx, nil, NewCreatePairRequest(factory, a, a))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: createPair: bind: transact: invalid options: options is nil")
}

func TestFactoryWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	createPairData, err := EncodeCreatePair(a, b)
	require.NoError(t, err)
	setFeeToData, err := EncodeSetFeeTo(a)
	require.NoError(t, err)
	setFeeToSetterData, err := EncodeSetFeeToSetter(b)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"createPair",
			createPairData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.CreatePair(t.Context(), opts, NewCreatePairRequest(factory, a, b))
			},
		},
		{
			"setFeeTo",
			setFeeToData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.SetFeeTo(t.Context(), opts, NewSetFeeToRequest(factory, a))
			},
		},
		{
			"setFeeToSetter",
			setFeeToSetterData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.SetFeeToSetter(t.Context(), opts, NewSetFeeToSetterRequest(factory, b))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(3_000_000)
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

			signed, hash, err := tc.send(NewFactory(c), bind.NewTransactOptions(key, builder))
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(factory))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestFactoryWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonForbidden)
	require.NoError(t, err)
	revert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := NewFactory(c).SetFeeTo(t.Context(), bind.NewTransactOptions(key, builder), NewSetFeeToRequest(factory, a))
	require.ErrorIs(t, err, ErrForbidden)
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestFactoryWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Twice()

	signed, hash, err := NewFactory(c).CreatePair(t.Context(), bind.NewTransactOptions(key, builder), NewCreatePairRequest(factory, a, b))
	require.ErrorContains(t, err, "uniswapv2: createPair: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestFactoryCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "factory")
	factory := types.MustNewAddressFromHex("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "factory", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewFactory(c).AllPairsLength(ctx, nil, NewAllPairsLengthRequest(factory))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewFactory(c).SetFeeTo(ctx, bind.NewTransactOptions(key, builder), NewSetFeeToRequest(factory, a))
	require.NoError(t, err)
}
