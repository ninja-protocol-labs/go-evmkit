package weth

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	amount := big.NewInt(1_000_000_000_000_000_000)

	withdrawData, err := EncodeWithdraw(amount)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		wantValue string
		send      func(w *WETH, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"deposit",
			EncodeDeposit(),
			"1000000000000000000",
			func(w *WETH, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return w.Deposit(t.Context(), opts, NewDepositRequest(contract, amount))
			},
		},
		{
			"deposit zero",
			EncodeDeposit(),
			"0",
			func(w *WETH, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return w.Deposit(t.Context(), opts, NewDepositRequest(contract, big.NewInt(0)))
			},
		},
		{
			"withdraw",
			withdrawData,
			"0",
			func(w *WETH, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return w.Withdraw(t.Context(), opts, NewWithdrawRequest(contract, amount))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(60_000)
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

			opts := bind.NewTransactOptions(key, builder)
			signed, hash, err := tc.send(New(c), opts)
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())
			require.Nil(t, opts.Value)

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(contract))
			require.Equal(t, tc.wantValue, dyn.Value.String())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestWithdrawPassesOptionsValue(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(60_000)
	builder.Price(big.NewInt(2), big.NewInt(1))

	var raw []byte
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
		require.NoError(t, err)
		*e.Result.(*string) = core.Keccak256(raw).String()
		return nil
	}).Once()

	_, _, err = New(c).Withdraw(t.Context(), bind.NewTransactOptions(key, builder).WithValue(big.NewInt(3)), NewWithdrawRequest(contract, big.NewInt(1)))
	require.NoError(t, err)

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	require.Equal(t, "3", decoded.(*core.DynamicFeeTx).Value.String())
}

func TestDepositRejectsOptionsValue(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	w := New(mocks.NewClient(t))

	for _, value := range []*big.Int{big.NewInt(1), big.NewInt(0)} {
		signed, hash, err := w.Deposit(t.Context(), bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(value), NewDepositRequest(contract, big.NewInt(1)))
		require.Nil(t, signed)
		require.Nil(t, hash)
		require.ErrorIs(t, err, bind.ErrInvalidOptions)
		require.EqualError(t, err, "weth: deposit: invalid options: value must be nil")
	}

	signed, hash, err := w.Deposit(t.Context(), nil, NewDepositRequest(contract, big.NewInt(1)))
	require.Nil(t, signed)
	require.Nil(t, hash)
	require.EqualError(t, err, "weth: deposit: bind: transact: invalid options: options is nil")
}

func TestClientMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	w := New(mocks.NewClient(t))
	ctx := t.Context()

	_, _, err = w.Deposit(ctx, opts, &DepositRequest{})
	require.EqualError(t, err, "weth: deposit: "+(&DepositRequest{}).Validate().Error())
	_, _, err = w.Deposit(ctx, opts, NewDepositRequest(contract, nil))
	require.EqualError(t, err, "weth: deposit: "+NewDepositRequest(contract, nil).Validate().Error())
	_, _, err = w.Deposit(ctx, opts, NewDepositRequest(contract, big.NewInt(-5)))
	require.EqualError(t, err, "weth: deposit: "+NewDepositRequest(contract, big.NewInt(-5)).Validate().Error())
	_, _, err = w.Deposit(ctx, bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewDepositRequest(nil, big.NewInt(1)))
	require.EqualError(t, err, "weth: deposit: "+NewDepositRequest(nil, big.NewInt(1)).Validate().Error())

	_, _, err = w.Withdraw(ctx, opts, &WithdrawRequest{})
	require.EqualError(t, err, "weth: withdraw: "+(&WithdrawRequest{}).Validate().Error())
	_, _, err = w.Withdraw(ctx, opts, NewWithdrawRequest(contract, nil))
	require.EqualError(t, err, "weth: withdraw: "+NewWithdrawRequest(contract, nil).Validate().Error())
	_, _, err = w.Withdraw(ctx, opts, NewWithdrawRequest(contract, big.NewInt(-1)))
	require.ErrorContains(t, err, "weth: withdraw: weth: encodeWithdraw: ")
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = w.Withdraw(ctx, nil, NewWithdrawRequest(contract, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "weth: withdraw: bind: transact: invalid options: options is nil")
}

func TestWithdrawSurfacesEstimateRevert(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	revert := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x"`)}
	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, revert)
	}).Once()

	signed, hash, err := New(c).Withdraw(t.Context(), bind.NewTransactOptions(key, builder), NewWithdrawRequest(contract, big.NewInt(1)))
	require.ErrorIs(t, err, revert)
	require.ErrorContains(t, err, "weth: withdraw: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(60_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "insufficient funds for gas * price + value"}).Once()

	signed, hash, err := New(c).Deposit(t.Context(), bind.NewTransactOptions(key, builder), NewDepositRequest(contract, big.NewInt(77)))
	require.ErrorContains(t, err, "weth: deposit: ")
	require.ErrorContains(t, err, "insufficient funds")
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	require.Equal(t, "77", signed.(*core.DynamicFeeTx).Value.String())
}

func TestERC20MethodsArePromoted(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	balanceOfData, err := erc20.EncodeBalanceOf(owner)
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(42))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		call := e.Params.([]any)[0].(map[string]any)
		require.Equal(t, contract.String(), call["to"])
		require.Equal(t, encoding.Hex.EncodePrefixed(balanceOfData), call["data"])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	balance, err := New(c).BalanceOf(t.Context(), nil, erc20.NewBalanceOfRequest(contract, owner))
	require.NoError(t, err)
	require.Equal(t, "42", balance.String())
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "weth")
	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "weth", got.Value(marker{}))
		return nil
	}).Twice()

	for i, send := range []func(opts *bind.TransactOptions) error{
		func(opts *bind.TransactOptions) error {
			_, _, err := New(c).Deposit(ctx, opts, NewDepositRequest(contract, big.NewInt(1)))
			return err
		},
		func(opts *bind.TransactOptions) error {
			_, _, err := New(c).Withdraw(ctx, opts, NewWithdrawRequest(contract, big.NewInt(1)))
			return err
		},
	} {
		builder := tx.NewDynamicFeeBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(uint64(i))
		builder.Limit(60_000)
		builder.Price(big.NewInt(2), big.NewInt(1))
		require.NoError(t, send(bind.NewTransactOptions(key, builder)))
	}
}
