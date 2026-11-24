package v3

import (
	"context"
	"encoding/json/jsontext"
	"errors"
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
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")

	ownerWord, err := abi.Pack(abi.NewTypes(abi.Address), b)
	require.NoError(t, err)
	tickSpacingWord, err := abi.Pack(abi.NewTypes(abi.Int24), big.NewInt(60))
	require.NoError(t, err)
	negativeTickSpacingWord, err := abi.Pack(abi.NewTypes(abi.Int24), big.NewInt(-1))
	require.NoError(t, err)
	poolWord, err := abi.Pack(abi.NewTypes(abi.Address), pool)
	require.NoError(t, err)

	tickSpacingData, err := EncodeFeeAmountTickSpacing(big.NewInt(3000))
	require.NoError(t, err)
	negativeTickSpacingData, err := EncodeFeeAmountTickSpacing(big.NewInt(100))
	require.NoError(t, err)
	getPoolData, err := EncodeGetPool(a, b, big.NewInt(3000))
	require.NoError(t, err)
	createPoolData, err := EncodeCreatePool(a, b, big.NewInt(500))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(f *Factory) (any, error)
		want     any
	}{
		{
			"owner",
			EncodeOwner(),
			ownerWord,
			func(f *Factory) (any, error) {
				return f.Owner(t.Context(), opts, NewOwnerRequest(factory))
			},
			b,
		},
		{
			"feeAmountTickSpacing",
			tickSpacingData,
			tickSpacingWord,
			func(f *Factory) (any, error) {
				return f.FeeAmountTickSpacing(t.Context(), opts, NewFeeAmountTickSpacingRequest(factory, big.NewInt(3000)))
			},
			big.NewInt(60),
		},
		{
			"feeAmountTickSpacing negative",
			negativeTickSpacingData,
			negativeTickSpacingWord,
			func(f *Factory) (any, error) {
				return f.FeeAmountTickSpacing(t.Context(), opts, NewFeeAmountTickSpacingRequest(factory, big.NewInt(100)))
			},
			big.NewInt(-1),
		},
		{
			"getPool",
			getPoolData,
			poolWord,
			func(f *Factory) (any, error) {
				return f.GetPool(t.Context(), opts, NewGetPoolRequest(factory, a, b, big.NewInt(3000)))
			},
			pool,
		},
		{
			"callCreatePool",
			createPoolData,
			poolWord,
			func(f *Factory) (any, error) {
				return f.CallCreatePool(t.Context(), opts, NewCreatePoolRequest(factory, a, b, big.NewInt(500)))
			},
			pool,
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
				require.NotContains(t, call, "value")
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

func TestFactoryReadsMapRevertToErrReverted(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0xdeadbeef"`), jsontext.Value(`{"x":1}`)} {
		t.Run(string(data), func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

			pool, err := NewFactory(c).CallCreatePool(t.Context(), bind.NewCallOptions().WithFrom(a), NewCreatePoolRequest(factory, a, a, big.NewInt(3000)))
			require.Nil(t, pool)
			require.ErrorIs(t, err, ErrReverted)
			require.ErrorContains(t, err, "uniswapv3: callCreatePool: reverted: ")
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	owner, err := NewFactory(c).Owner(t.Context(), nil, NewOwnerRequest(factory))
	require.Nil(t, owner)
	require.EqualError(t, err, "uniswapv3: owner: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestFactoryReadsRejectMalformedResult(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(4)
			f := NewFactory(c)

			owner, err := f.Owner(t.Context(), nil, NewOwnerRequest(factory))
			require.Nil(t, owner)
			require.ErrorContains(t, err, "uniswapv3: owner: ")

			tickSpacing, err := f.FeeAmountTickSpacing(t.Context(), nil, NewFeeAmountTickSpacingRequest(factory, big.NewInt(3000)))
			require.Nil(t, tickSpacing)
			require.ErrorContains(t, err, "uniswapv3: feeAmountTickSpacing: ")

			pool, err := f.GetPool(t.Context(), nil, NewGetPoolRequest(factory, a, a, big.NewInt(3000)))
			require.Nil(t, pool)
			require.ErrorContains(t, err, "uniswapv3: getPool: ")

			created, err := f.CallCreatePool(t.Context(), nil, NewCreatePoolRequest(factory, a, a, big.NewInt(3000)))
			require.Nil(t, created)
			require.ErrorContains(t, err, "uniswapv3: callCreatePool: ")
		})
	}
}

func TestFactoryMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	f := NewFactory(mocks.NewClient(t))
	ctx := t.Context()

	_, err = f.Owner(ctx, nil, &OwnerRequest{})
	require.EqualError(t, err, "uniswapv3: owner: "+(&OwnerRequest{}).Validate().Error())
	_, err = f.FeeAmountTickSpacing(ctx, nil, NewFeeAmountTickSpacingRequest(factory, nil))
	require.EqualError(t, err, "uniswapv3: feeAmountTickSpacing: "+NewFeeAmountTickSpacingRequest(factory, nil).Validate().Error())
	_, err = f.GetPool(ctx, nil, NewGetPoolRequest(factory, a, nil, big.NewInt(3000)))
	require.EqualError(t, err, "uniswapv3: getPool: "+NewGetPoolRequest(factory, a, nil, big.NewInt(3000)).Validate().Error())
	_, err = f.CallCreatePool(ctx, nil, NewCreatePoolRequest(nil, a, a, big.NewInt(3000)))
	require.EqualError(t, err, "uniswapv3: callCreatePool: "+NewCreatePoolRequest(nil, a, a, big.NewInt(3000)).Validate().Error())

	_, _, err = f.CreatePool(ctx, opts, NewCreatePoolRequest(factory, nil, a, big.NewInt(3000)))
	require.EqualError(t, err, "uniswapv3: createPool: "+NewCreatePoolRequest(factory, nil, a, big.NewInt(3000)).Validate().Error())
	_, _, err = f.SetOwner(ctx, opts, NewSetOwnerRequest(factory, nil))
	require.EqualError(t, err, "uniswapv3: setOwner: "+NewSetOwnerRequest(factory, nil).Validate().Error())
	_, _, err = f.EnableFeeAmount(ctx, opts, &EnableFeeAmountRequest{})
	require.EqualError(t, err, "uniswapv3: enableFeeAmount: "+(&EnableFeeAmountRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = f.FeeAmountTickSpacing(ctx, nil, NewFeeAmountTickSpacingRequest(factory, big.NewInt(1<<24)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = f.EnableFeeAmount(ctx, opts, NewEnableFeeAmountRequest(factory, big.NewInt(100), big.NewInt(1<<23)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = f.CreatePool(ctx, nil, NewCreatePoolRequest(factory, a, a, big.NewInt(3000)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: createPool: bind: transact: invalid options: options is nil")
}

func TestFactoryWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	createPoolData, err := EncodeCreatePool(a, b, big.NewInt(3000))
	require.NoError(t, err)
	setOwnerData, err := EncodeSetOwner(b)
	require.NoError(t, err)
	enableFeeAmountData, err := EncodeEnableFeeAmount(big.NewInt(100), big.NewInt(1))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"createPool",
			createPoolData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.CreatePool(t.Context(), opts, NewCreatePoolRequest(factory, a, b, big.NewInt(3000)))
			},
		},
		{
			"setOwner",
			setOwnerData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.SetOwner(t.Context(), opts, NewSetOwnerRequest(factory, b))
			},
		},
		{
			"enableFeeAmount",
			enableFeeAmountData,
			func(f *Factory, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return f.EnableFeeAmount(t.Context(), opts, NewEnableFeeAmountRequest(factory, big.NewInt(100), big.NewInt(1)))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(6_000_000)
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

func TestFactoryWritesMapEstimateRevertToErrReverted(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x"`)})
	}).Once()

	signed, hash, err := NewFactory(c).SetOwner(t.Context(), bind.NewTransactOptions(key, builder), NewSetOwnerRequest(factory, a))
	require.ErrorIs(t, err, ErrReverted)
	require.ErrorContains(t, err, "uniswapv3: setOwner: reverted: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestFactoryWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(6_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Twice()

	signed, hash, err := NewFactory(c).CreatePool(t.Context(), bind.NewTransactOptions(key, builder), NewCreatePoolRequest(factory, a, b, big.NewInt(3000)))
	require.ErrorContains(t, err, "uniswapv3: createPool: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrReverted)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestFactoryCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "factory")
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "factory", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewFactory(c).Owner(ctx, nil, NewOwnerRequest(factory))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewFactory(c).SetOwner(ctx, bind.NewTransactOptions(key, builder), NewSetOwnerRequest(factory, a))
	require.NoError(t, err)
}
