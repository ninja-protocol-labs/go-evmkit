package erc20

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
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount := big.NewInt(12345)
	opts := bind.NewCallOptions().WithFrom(owner).WithBlock("0x1b4")

	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
	require.NoError(t, err)
	trueWord, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	nameWord, err := abi.Pack(abi.NewTypes(abi.String), "USD Coin")
	require.NoError(t, err)
	symbolWord, err := abi.Pack(abi.NewTypes(abi.String), "USDC")
	require.NoError(t, err)
	decimalsWord, err := abi.Pack(abi.NewTypes(abi.Uint8), uint8(6))
	require.NoError(t, err)

	balanceOfData, err := EncodeBalanceOf(owner)
	require.NoError(t, err)
	allowanceData, err := EncodeAllowance(owner, spender)
	require.NoError(t, err)
	transferData, err := EncodeTransfer(spender, amount)
	require.NoError(t, err)
	approveData, err := EncodeApprove(spender, amount)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(owner, spender, amount)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(e *ERC20) (any, error)
		want     string
	}{
		{
			"totalSupply",
			EncodeTotalSupply(),
			amountWord,
			func(e *ERC20) (any, error) {
				return e.TotalSupply(t.Context(), opts, NewTotalSupplyRequest(token))
			},
			"12345",
		},
		{
			"balanceOf",
			balanceOfData,
			amountWord,
			func(e *ERC20) (any, error) {
				return e.BalanceOf(t.Context(), opts, NewBalanceOfRequest(token, owner))
			},
			"12345",
		},
		{
			"allowance",
			allowanceData,
			amountWord,
			func(e *ERC20) (any, error) {
				return e.Allowance(t.Context(), opts, NewAllowanceRequest(token, owner, spender))
			},
			"12345",
		},
		{
			"name",
			EncodeName(),
			nameWord,
			func(e *ERC20) (any, error) {
				return e.Name(t.Context(), opts, NewNameRequest(token))
			},
			"USD Coin",
		},
		{
			"symbol",
			EncodeSymbol(),
			symbolWord,
			func(e *ERC20) (any, error) {
				return e.Symbol(t.Context(), opts, NewSymbolRequest(token))
			},
			"USDC",
		},
		{
			"decimals",
			EncodeDecimals(),
			decimalsWord,
			func(e *ERC20) (any, error) {
				return e.Decimals(t.Context(), opts, NewDecimalsRequest(token))
			},
			"6",
		},
		{
			"callTransfer",
			transferData,
			trueWord,
			func(e *ERC20) (any, error) {
				return e.CallTransfer(t.Context(), opts, NewTransferRequest(token, spender, amount))
			},
			"true",
		},
		{
			"callApprove",
			approveData,
			trueWord,
			func(e *ERC20) (any, error) {
				return e.CallApprove(t.Context(), opts, NewApproveRequest(token, spender, amount))
			},
			"true",
		},
		{
			"callTransferFrom",
			transferFromData,
			trueWord,
			func(e *ERC20) (any, error) {
				return e.CallTransferFrom(t.Context(), opts, NewTransferFromRequest(token, owner, spender, amount))
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
				require.Equal(t, token.String(), call["to"])
				require.Equal(t, owner.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.Equal(t, "0x1b4", params[1])

				result, ok := e.Result.(*string)
				require.True(t, ok)
				*result = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, tc.want, fmt.Sprint(got))
		})
	}
}

func TestReadsMapRevertToSentinel(t *testing.T) {
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	insufficientArgs, err := abi.Pack(errInsufficientBalance.Inputs, spender, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)
	insufficientSel := errInsufficientBalance.Selector()
	insufficient := append(append([]byte{}, insufficientSel[:]...), insufficientArgs...)

	for _, tc := range []struct {
		name   string
		data   string
		want   error
		wantNo error
	}{
		{"insufficient balance", `"` + encoding.Hex.EncodePrefixed(insufficient) + `"`, ErrInsufficientBalance, ErrUnknown},
		{"unknown selector", `"0xdeadbeef"`, ErrUnknown, ErrInsufficientBalance},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(tc.data)}).Once()

			ok, err := New(c).CallTransfer(t.Context(), nil, NewTransferRequest(token, spender, big.NewInt(1)))
			require.False(t, ok)
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, tc.wantNo)
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "header not found"}).Once()

	n, err := New(c).BalanceOf(t.Context(), nil, NewBalanceOfRequest(token, spender))
	require.Nil(t, n)
	require.NotErrorIs(t, err, ErrUnknown)
	require.ErrorContains(t, err, "header not found")
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount := big.NewInt(12345)

	transferData, err := EncodeTransfer(spender, amount)
	require.NoError(t, err)
	approveData, err := EncodeApprove(spender, amount)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(owner, spender, amount)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(e *ERC20, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"transfer",
			transferData,
			func(e *ERC20, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.Transfer(t.Context(), opts, NewTransferRequest(token, spender, amount))
			},
		},
		{
			"approve",
			approveData,
			func(e *ERC20, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.Approve(t.Context(), opts, NewApproveRequest(token, spender, amount))
			},
		},
		{
			"transferFrom",
			transferFromData,
			func(e *ERC20, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.TransferFrom(t.Context(), opts, NewTransferFromRequest(token, owner, spender, amount))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(80_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 1)
				rawHex, ok := params[0].(string)
				require.True(t, ok)
				raw, err = encoding.Hex.Decode(rawHex)
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
			require.True(t, dyn.To.Equal(token))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			from, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, from.Equal(sender))
		})
	}
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	args, err := abi.Pack(errInsufficientBalance.Inputs, spender, big.NewInt(0), big.NewInt(1))
	require.NoError(t, err)
	sel := errInsufficientBalance.Selector()
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

	signed, hash, err := New(c).Transfer(t.Context(), bind.NewTransactOptions(key, builder), NewTransferRequest(token, spender, big.NewInt(1)))
	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestMetadataBatchesThroughMulticall(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")

	nameWord, err := abi.Pack(abi.NewTypes(abi.String), "USD Coin")
	require.NoError(t, err)
	symbolWord, err := abi.Pack(abi.NewTypes(abi.String), "USDC")
	require.NoError(t, err)
	decimalsWord, err := abi.Pack(abi.NewTypes(abi.Uint8), uint8(6))
	require.NoError(t, err)
	supplyWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1_000_000))
	require.NoError(t, err)
	result, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Tuple(abi.Bool, abi.Bytes))), []any{
		[]any{true, nameWord},
		[]any{true, symbolWord},
		[]any{true, decimalsWord},
		[]any{true, supplyWord},
	})
	require.NoError(t, err)
	calldata, err := EncodeMetadata(token, false)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		call := e.Params.([]any)[0].(map[string]any)
		require.Equal(t, multicall.String(), call["to"])
		require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
		return nil
	}).Once()

	m, err := New(c).Metadata(t.Context(), nil, NewMetadataRequest(multicall, token, false))
	require.NoError(t, err)
	require.Equal(t, "USD Coin", m.Name)
	require.Equal(t, "USDC", m.Symbol)
	require.Equal(t, uint8(6), m.Decimals)
	require.Zero(t, big.NewInt(1_000_000).Cmp(m.TotalSupply))
}

func TestReadsRejectMalformedResult(t *testing.T) {
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x" + "00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(4)
			e := New(c)

			n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(token, owner))
			require.Error(t, err)
			require.Nil(t, n)
			require.ErrorContains(t, err, "erc20: balanceOf: ")

			name, err := e.Name(t.Context(), nil, NewNameRequest(token))
			require.Error(t, err)
			require.Empty(t, name)

			ok, err := e.CallTransfer(t.Context(), nil, NewTransferRequest(token, owner, big.NewInt(1)))
			require.Error(t, err)
			require.False(t, ok)

			m, err := e.Metadata(t.Context(), nil, NewMetadataRequest(owner, token, false))
			require.Error(t, err)
			require.Nil(t, m)
		})
	}
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "erc20")
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "erc20", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(amountWord)
		}
		return nil
	}).Twice()

	_, err = New(c).BalanceOf(ctx, nil, NewBalanceOfRequest(token, owner))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(60_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).Transfer(ctx, bind.NewTransactOptions(key, builder), NewTransferRequest(token, owner, big.NewInt(1)))
	require.NoError(t, err)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(60_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var raw []byte
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		if e.Method == rpc.MethodETHGetTransactionByHash {
			return nil
		}
		raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
		require.NoError(t, err)
		return &rpc.ResponseError{Code: -32000, Message: "nonce too low"}
	}).Twice()

	signed, hash, err := New(c).Transfer(t.Context(), bind.NewTransactOptions(key, builder), NewTransferRequest(token, to, big.NewInt(5)))
	require.Error(t, err)
	require.ErrorContains(t, err, "erc20: transfer: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	signedHash, err := signed.Hash()
	require.NoError(t, err)
	require.True(t, signedHash.Equal(core.Keccak256(raw)))
}

func TestWritesHonourBuilderTypeAndValue(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	token := types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	calldata, err := EncodeApprove(to, big.NewInt(0))
	require.NoError(t, err)

	builder := tx.NewLegacyBuilder()
	builder.ChainID(big.NewInt(56))
	builder.Nonce(3)
	builder.Limit(50_000)
	builder.Price(big.NewInt(3_000_000_000), nil)

	var raw []byte
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
		require.NoError(t, err)
		*e.Result.(*string) = core.Keccak256(raw).String()
		return nil
	}).Once()

	_, _, err = New(c).Approve(t.Context(), bind.NewTransactOptions(key, builder).WithValue(big.NewInt(77)), NewApproveRequest(token, to, big.NewInt(0)))
	require.NoError(t, err)

	decoded, err := core.DecodeTransaction(raw)
	require.NoError(t, err)
	legacy, ok := decoded.(*core.LegacyTx)
	require.True(t, ok, "got %T", decoded)
	require.Zero(t, big.NewInt(56).Cmp(legacy.ChainID))
	require.Equal(t, uint64(3), legacy.Nonce)
	require.Equal(t, uint64(50_000), legacy.GasLimit)
	require.Zero(t, big.NewInt(3_000_000_000).Cmp(legacy.GasPrice))
	require.Zero(t, big.NewInt(77).Cmp(legacy.Value))
	require.True(t, legacy.To.Equal(token))
	require.Equal(t, calldata, legacy.Data)
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
