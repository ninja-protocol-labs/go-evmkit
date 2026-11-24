package erc6909

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

func TestReadsSendCallAndDecode(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(7)
	amount := big.NewInt(12345)
	opts := bind.NewCallOptions().WithFrom(owner).WithBlock("0x1b4")

	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
	require.NoError(t, err)
	trueWord, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	nameWord, err := abi.Pack(abi.NewTypes(abi.String), "Claim 7")
	require.NoError(t, err)
	symbolWord, err := abi.Pack(abi.NewTypes(abi.String), "C7")
	require.NoError(t, err)
	decimalsWord, err := abi.Pack(abi.NewTypes(abi.Uint8), uint8(6))
	require.NoError(t, err)

	balanceOfData, err := EncodeBalanceOf(owner, id)
	require.NoError(t, err)
	allowanceData, err := EncodeAllowance(owner, spender, id)
	require.NoError(t, err)
	isOperatorData, err := EncodeIsOperator(owner, spender)
	require.NoError(t, err)
	nameData, err := EncodeName(id)
	require.NoError(t, err)
	symbolData, err := EncodeSymbol(id)
	require.NoError(t, err)
	decimalsData, err := EncodeDecimals(id)
	require.NoError(t, err)
	approveData, err := EncodeApprove(spender, id, amount)
	require.NoError(t, err)
	setOperatorData, err := EncodeSetOperator(spender, true)
	require.NoError(t, err)
	transferData, err := EncodeTransfer(spender, id, amount)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(owner, spender, id, amount)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(e *ERC6909) (any, error)
		want     string
	}{
		{
			"balanceOf",
			balanceOfData,
			amountWord,
			func(e *ERC6909) (any, error) {
				return e.BalanceOf(t.Context(), opts, NewBalanceOfRequest(contract, owner, id))
			},
			"12345",
		},
		{
			"allowance",
			allowanceData,
			amountWord,
			func(e *ERC6909) (any, error) {
				return e.Allowance(t.Context(), opts, NewAllowanceRequest(contract, owner, spender, id))
			},
			"12345",
		},
		{
			"isOperator",
			isOperatorData,
			trueWord,
			func(e *ERC6909) (any, error) {
				return e.IsOperator(t.Context(), opts, NewIsOperatorRequest(contract, owner, spender))
			},
			"true",
		},
		{
			"name",
			nameData,
			nameWord,
			func(e *ERC6909) (any, error) {
				return e.Name(t.Context(), opts, NewNameRequest(contract, id))
			},
			"Claim 7",
		},
		{
			"symbol",
			symbolData,
			symbolWord,
			func(e *ERC6909) (any, error) {
				return e.Symbol(t.Context(), opts, NewSymbolRequest(contract, id))
			},
			"C7",
		},
		{
			"decimals",
			decimalsData,
			decimalsWord,
			func(e *ERC6909) (any, error) {
				return e.Decimals(t.Context(), opts, NewDecimalsRequest(contract, id))
			},
			"6",
		},
		{
			"callApprove",
			approveData,
			trueWord,
			func(e *ERC6909) (any, error) {
				return e.CallApprove(t.Context(), opts, NewApproveRequest(contract, spender, id, amount))
			},
			"true",
		},
		{
			"callSetOperator",
			setOperatorData,
			trueWord,
			func(e *ERC6909) (any, error) {
				return e.CallSetOperator(t.Context(), opts, NewSetOperatorRequest(contract, spender, true))
			},
			"true",
		},
		{
			"callTransfer",
			transferData,
			trueWord,
			func(e *ERC6909) (any, error) {
				return e.CallTransfer(t.Context(), opts, NewTransferRequest(contract, spender, id, amount))
			},
			"true",
		},
		{
			"callTransferFrom",
			transferFromData,
			trueWord,
			func(e *ERC6909) (any, error) {
				return e.CallTransferFrom(t.Context(), opts, NewTransferFromRequest(contract, owner, spender, id, amount))
			},
			"true",
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

func TestReadsMapRevertToSentinel(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	balanceArgs, err := abi.Pack(errInsufficientBalance.Inputs, owner, big.NewInt(0), big.NewInt(1), big.NewInt(7))
	require.NoError(t, err)
	balanceSel := errInsufficientBalance.Selector()
	allowanceArgs, err := abi.Pack(errInsufficientAllowance.Inputs, spender, big.NewInt(0), big.NewInt(1), big.NewInt(7))
	require.NoError(t, err)
	allowanceSel := errInsufficientAllowance.Selector()

	for _, tc := range []struct {
		name   string
		revert []byte
		want   error
		wantNo error
	}{
		{"insufficient balance", append(append([]byte{}, balanceSel[:]...), balanceArgs...), ErrInsufficientBalance, ErrInsufficientAllowance},
		{"insufficient allowance", append(append([]byte{}, allowanceSel[:]...), allowanceArgs...), ErrInsufficientAllowance, ErrInsufficientBalance},
		{"unknown selector", []byte{0xde, 0xad, 0xbe, 0xef}, ErrUnknown, ErrInsufficientBalance},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(tc.revert) + `"`)}).Once()

			ok, err := New(c).CallTransferFrom(t.Context(), bind.NewCallOptions().WithFrom(spender), NewTransferFromRequest(contract, owner, spender, big.NewInt(7), big.NewInt(1)))
			require.False(t, ok)
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, tc.wantNo)
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}
}

func TestReadsRejectMalformedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(3)
			e := New(c)

			n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(contract, owner, big.NewInt(1)))
			require.Error(t, err)
			require.Nil(t, n)
			require.ErrorContains(t, err, "erc6909: balanceOf: ")

			name, err := e.Name(t.Context(), nil, NewNameRequest(contract, big.NewInt(1)))
			require.Error(t, err)
			require.Empty(t, name)

			ok, err := e.CallTransfer(t.Context(), nil, NewTransferRequest(contract, owner, big.NewInt(1), big.NewInt(1)))
			require.Error(t, err)
			require.False(t, ok)
		})
	}
}

func TestClientMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	e := New(mocks.NewClient(t))
	ctx := t.Context()

	_, err := e.BalanceOf(ctx, nil, NewBalanceOfRequest(contract, contract, nil))
	require.EqualError(t, err, "erc6909: balanceOf: "+NewBalanceOfRequest(contract, contract, nil).Validate().Error())
	_, err = e.Allowance(ctx, nil, NewAllowanceRequest(contract, contract, nil, big.NewInt(1)))
	require.EqualError(t, err, "erc6909: allowance: "+NewAllowanceRequest(contract, contract, nil, big.NewInt(1)).Validate().Error())
	_, err = e.IsOperator(ctx, nil, &IsOperatorRequest{})
	require.EqualError(t, err, "erc6909: isOperator: "+(&IsOperatorRequest{}).Validate().Error())
	_, err = e.Name(ctx, nil, NewNameRequest(nil, big.NewInt(1)))
	require.EqualError(t, err, "erc6909: name: "+NewNameRequest(nil, big.NewInt(1)).Validate().Error())
	_, err = e.Symbol(ctx, nil, NewSymbolRequest(contract, nil))
	require.EqualError(t, err, "erc6909: symbol: "+NewSymbolRequest(contract, nil).Validate().Error())
	_, err = e.Decimals(ctx, nil, &DecimalsRequest{})
	require.EqualError(t, err, "erc6909: decimals: "+(&DecimalsRequest{}).Validate().Error())
	_, err = e.CallApprove(ctx, nil, NewApproveRequest(contract, contract, big.NewInt(1), nil))
	require.EqualError(t, err, "erc6909: callApprove: "+NewApproveRequest(contract, contract, big.NewInt(1), nil).Validate().Error())
	_, err = e.CallSetOperator(ctx, nil, NewSetOperatorRequest(contract, nil, true))
	require.EqualError(t, err, "erc6909: callSetOperator: "+NewSetOperatorRequest(contract, nil, true).Validate().Error())
	_, err = e.CallTransfer(ctx, nil, NewTransferRequest(contract, nil, big.NewInt(1), big.NewInt(1)))
	require.EqualError(t, err, "erc6909: callTransfer: "+NewTransferRequest(contract, nil, big.NewInt(1), big.NewInt(1)).Validate().Error())
	_, err = e.CallTransferFrom(ctx, nil, NewTransferFromRequest(contract, nil, contract, big.NewInt(1), big.NewInt(1)))
	require.EqualError(t, err, "erc6909: callTransferFrom: "+NewTransferFromRequest(contract, nil, contract, big.NewInt(1), big.NewInt(1)).Validate().Error())
	_, _, err = e.Approve(ctx, nil, NewApproveRequest(nil, contract, big.NewInt(1), big.NewInt(1)))
	require.EqualError(t, err, "erc6909: approve: "+NewApproveRequest(nil, contract, big.NewInt(1), big.NewInt(1)).Validate().Error())
	_, _, err = e.SetOperator(ctx, nil, &SetOperatorRequest{})
	require.EqualError(t, err, "erc6909: setOperator: "+(&SetOperatorRequest{}).Validate().Error())
	_, _, err = e.Transfer(ctx, nil, NewTransferRequest(contract, contract, nil, big.NewInt(1)))
	require.EqualError(t, err, "erc6909: transfer: "+NewTransferRequest(contract, contract, nil, big.NewInt(1)).Validate().Error())
	_, _, err = e.TransferFrom(ctx, nil, NewTransferFromRequest(contract, contract, contract, big.NewInt(1), nil))
	require.EqualError(t, err, "erc6909: transferFrom: "+NewTransferFromRequest(contract, contract, contract, big.NewInt(1), nil).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = e.Transfer(ctx, nil, NewTransferRequest(contract, contract, big.NewInt(1), big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "erc6909: transfer: bind: transact: invalid options: options is nil")
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	receiver := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(7)
	amount := big.NewInt(12345)

	approveData, err := EncodeApprove(receiver, id, amount)
	require.NoError(t, err)
	setOperatorData, err := EncodeSetOperator(receiver, false)
	require.NoError(t, err)
	transferData, err := EncodeTransfer(receiver, id, amount)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(owner, receiver, id, amount)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(e *ERC6909, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"approve",
			approveData,
			func(e *ERC6909, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.Approve(t.Context(), opts, NewApproveRequest(contract, receiver, id, amount))
			},
		},
		{
			"setOperator",
			setOperatorData,
			func(e *ERC6909, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SetOperator(t.Context(), opts, NewSetOperatorRequest(contract, receiver, false))
			},
		},
		{
			"transfer",
			transferData,
			func(e *ERC6909, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.Transfer(t.Context(), opts, NewTransferRequest(contract, receiver, id, amount))
			},
		},
		{
			"transferFrom",
			transferFromData,
			func(e *ERC6909, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.TransferFrom(t.Context(), opts, NewTransferFromRequest(contract, owner, receiver, id, amount))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
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

			signed, hash, err := tc.send(New(c), bind.NewTransactOptions(key, builder))
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(contract))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	args, err := abi.Pack(errInvalidReceiver.Inputs, types.ZeroAddress())
	require.NoError(t, err)
	sel := errInvalidReceiver.Selector()
	revert := append(append([]byte{}, sel[:]...), args...)

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

	signed, hash, err := New(c).Transfer(t.Context(), bind.NewTransactOptions(key, builder), NewTransferRequest(contract, types.ZeroAddress(), big.NewInt(1), big.NewInt(1)))
	require.ErrorIs(t, err, ErrInvalidReceiver)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(90_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Twice()

	signed, hash, err := New(c).SetOperator(t.Context(), bind.NewTransactOptions(key, builder), NewSetOperatorRequest(contract, contract, true))
	require.ErrorContains(t, err, "erc6909: setOperator: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "erc6909")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "erc6909", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(amountWord)
		}
		return nil
	}).Twice()

	_, err = New(c).BalanceOf(ctx, nil, NewBalanceOfRequest(contract, owner, big.NewInt(1)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(90_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).Approve(ctx, bind.NewTransactOptions(key, builder), NewApproveRequest(contract, owner, big.NewInt(1), big.NewInt(1)))
	require.NoError(t, err)
}

func TestMapRevertErrorPassesThroughNilAndUndecodable(t *testing.T) {
	require.NoError(t, MapRevertError(nil))
	var success *rpc.ResponseError
	require.NoError(t, MapRevertError(success))

	plain := fmt.Errorf("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
	empty := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	require.Same(t, empty, MapRevertError(empty))
}
