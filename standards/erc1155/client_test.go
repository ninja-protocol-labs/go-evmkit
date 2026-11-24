package erc1155

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
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(7)
	opts := bind.NewCallOptions().WithFrom(account).WithBlock("0x1b4")

	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(500))
	require.NoError(t, err)
	amountsWord, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(500), big.NewInt(0)})
	require.NoError(t, err)
	trueWord, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	uriWord, err := abi.Pack(abi.NewTypes(abi.String), "ipfs://{id}.json")
	require.NoError(t, err)

	uriData, err := EncodeURI(id)
	require.NoError(t, err)
	balanceOfData, err := EncodeBalanceOf(account, id)
	require.NoError(t, err)
	balanceOfBatchData, err := EncodeBalanceOfBatch([]*types.Address{account, operator}, []*big.Int{id, id})
	require.NoError(t, err)
	isApprovedForAllData, err := EncodeIsApprovedForAll(account, operator)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(e *ERC1155) (any, error)
		want     string
	}{
		{
			"uri",
			uriData,
			uriWord,
			func(e *ERC1155) (any, error) {
				return e.URI(t.Context(), opts, NewURIRequest(contract, id))
			},
			"ipfs://{id}.json",
		},
		{
			"balanceOf",
			balanceOfData,
			amountWord,
			func(e *ERC1155) (any, error) {
				return e.BalanceOf(t.Context(), opts, NewBalanceOfRequest(contract, account, id))
			},
			"500",
		},
		{
			"balanceOfBatch",
			balanceOfBatchData,
			amountsWord,
			func(e *ERC1155) (any, error) {
				return e.BalanceOfBatch(t.Context(), opts, NewBalanceOfBatchRequest(contract, []*types.Address{account, operator}, []*big.Int{id, id}))
			},
			"[500 0]",
		},
		{
			"isApprovedForAll",
			isApprovedForAllData,
			trueWord,
			func(e *ERC1155) (any, error) {
				return e.IsApprovedForAll(t.Context(), opts, NewIsApprovedForAllRequest(contract, account, operator))
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
				require.Equal(t, account.String(), call["from"])
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
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	args, err := abi.Pack(errInvalidArrayLength.Inputs, big.NewInt(2), big.NewInt(1))
	require.NoError(t, err)
	sel := errInvalidArrayLength.Selector()
	invalidLength := append(append([]byte{}, sel[:]...), args...)

	for _, tc := range []struct {
		name   string
		data   string
		want   error
		wantNo error
	}{
		{"invalid array length", `"` + encoding.Hex.EncodePrefixed(invalidLength) + `"`, ErrInvalidArrayLength, ErrUnknown},
		{"unknown selector", `"0xdeadbeef"`, ErrUnknown, ErrInvalidArrayLength},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(tc.data)}).Once()

			amounts, err := New(c).BalanceOfBatch(t.Context(), nil, NewBalanceOfBatchRequest(contract, []*types.Address{account}, []*big.Int{big.NewInt(1)}))
			require.Nil(t, amounts)
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, tc.wantNo)
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	amounts := []*big.Int{big.NewInt(10), big.NewInt(20)}

	setApprovalForAllData, err := EncodeSetApprovalForAll(to, true)
	require.NoError(t, err)
	safeTransferFromData, err := EncodeSafeTransferFrom(from, to, ids[0], amounts[0], []byte{0xca, 0xfe})
	require.NoError(t, err)
	safeBatchTransferFromData, err := EncodeSafeBatchTransferFrom(from, to, ids, amounts, nil)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(e *ERC1155, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"setApprovalForAll",
			setApprovalForAllData,
			func(e *ERC1155, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SetApprovalForAll(t.Context(), opts, NewSetApprovalForAllRequest(contract, to, true))
			},
		},
		{
			"safeTransferFrom",
			safeTransferFromData,
			func(e *ERC1155, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SafeTransferFrom(t.Context(), opts, NewSafeTransferFromRequest(contract, from, to, ids[0], amounts[0], []byte{0xca, 0xfe}))
			},
		},
		{
			"safeBatchTransferFrom",
			safeBatchTransferFromData,
			func(e *ERC1155, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SafeBatchTransferFrom(t.Context(), opts, NewSafeBatchTransferFromRequest(contract, from, to, ids, amounts, nil))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(150_000)
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
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	args, err := abi.Pack(errInsufficientBalance.Inputs, from, big.NewInt(0), big.NewInt(10), big.NewInt(1))
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

	signed, hash, err := New(c).SafeTransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewSafeTransferFromRequest(contract, from, to, big.NewInt(1), big.NewInt(10), nil))
	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestIdBalancesRepeatsAccountPerID(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	ids := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}

	calldata, err := EncodeBalanceOfBatch([]*types.Address{account, account, account}, ids)
	require.NoError(t, err)
	result, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(10), big.NewInt(0), big.NewInt(30)})
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		call := e.Params.([]any)[0].(map[string]any)
		require.Equal(t, contract.String(), call["to"])
		require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
		return nil
	}).Once()

	balances, err := New(c).IdBalances(t.Context(), nil, NewIdBalancesRequest(contract, account, ids))
	require.NoError(t, err)
	require.Len(t, balances, 3)
	for i, want := range []int64{10, 0, 30} {
		require.Zero(t, ids[i].Cmp(balances[i].Id))
		require.Zero(t, big.NewInt(want).Cmp(balances[i].Amount))
	}
}

func TestReadsRejectMalformedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(4)
			e := New(c)

			n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(contract, account, big.NewInt(1)))
			require.Error(t, err)
			require.Nil(t, n)
			require.ErrorContains(t, err, "erc1155: balanceOf: ")

			amounts, err := e.BalanceOfBatch(t.Context(), nil, NewBalanceOfBatchRequest(contract, []*types.Address{account}, []*big.Int{big.NewInt(1)}))
			require.Error(t, err)
			require.Nil(t, amounts)

			uri, err := e.URI(t.Context(), nil, NewURIRequest(contract, big.NewInt(1)))
			require.Error(t, err)
			require.Empty(t, uri)

			balances, err := e.IdBalances(t.Context(), nil, NewIdBalancesRequest(contract, account, []*big.Int{big.NewInt(1)}))
			require.Error(t, err)
			require.Nil(t, balances)
		})
	}
}

func TestBalanceOfBatchRejectsLengthMismatchedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	result, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Uint256)), []*big.Int{big.NewInt(1)})
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
		return nil
	}).Once()

	balances, err := New(c).IdBalances(t.Context(), nil, NewIdBalancesRequest(contract, account, []*big.Int{big.NewInt(1), big.NewInt(2)}))
	require.Error(t, err)
	require.Nil(t, balances)
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "erc1155")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	account := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	amountWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "erc1155", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(amountWord)
		}
		return nil
	}).Twice()

	_, err = New(c).BalanceOf(ctx, nil, NewBalanceOfRequest(contract, account, big.NewInt(1)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(90_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).SetApprovalForAll(ctx, bind.NewTransactOptions(key, builder), NewSetApprovalForAllRequest(contract, account, true))
	require.NoError(t, err)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(150_000)
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

	signed, hash, err := New(c).SafeBatchTransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewSafeBatchTransferFromRequest(contract, from, to, []*big.Int{big.NewInt(1)}, []*big.Int{big.NewInt(2)}, nil))
	require.Error(t, err)
	require.ErrorContains(t, err, "erc1155: safeBatchTransferFrom: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	signedHash, err := signed.Hash()
	require.NoError(t, err)
	require.True(t, signedHash.Equal(core.Keccak256(raw)))
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
