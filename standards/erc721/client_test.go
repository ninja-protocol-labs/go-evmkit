package erc721

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
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tokenID := big.NewInt(42)
	opts := bind.NewCallOptions().WithFrom(owner).WithBlock("0x1b4")

	countWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(3))
	require.NoError(t, err)
	ownerWord, err := abi.Pack(abi.NewTypes(abi.Address), owner)
	require.NoError(t, err)
	operatorWord, err := abi.Pack(abi.NewTypes(abi.Address), operator)
	require.NoError(t, err)
	trueWord, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	nameWord, err := abi.Pack(abi.NewTypes(abi.String), "Bored Ape Yacht Club")
	require.NoError(t, err)
	symbolWord, err := abi.Pack(abi.NewTypes(abi.String), "BAYC")
	require.NoError(t, err)
	uriWord, err := abi.Pack(abi.NewTypes(abi.String), "ipfs://42")
	require.NoError(t, err)

	balanceOfData, err := EncodeBalanceOf(owner)
	require.NoError(t, err)
	ownerOfData, err := EncodeOwnerOf(tokenID)
	require.NoError(t, err)
	getApprovedData, err := EncodeGetApproved(tokenID)
	require.NoError(t, err)
	isApprovedForAllData, err := EncodeIsApprovedForAll(owner, operator)
	require.NoError(t, err)
	tokenURIData, err := EncodeTokenURI(tokenID)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(e *ERC721) (any, error)
		want     string
	}{
		{
			"balanceOf",
			balanceOfData,
			countWord,
			func(e *ERC721) (any, error) {
				return e.BalanceOf(t.Context(), opts, NewBalanceOfRequest(contract, owner))
			},
			"3",
		},
		{
			"ownerOf",
			ownerOfData,
			ownerWord,
			func(e *ERC721) (any, error) {
				return e.OwnerOf(t.Context(), opts, NewOwnerOfRequest(contract, tokenID))
			},
			owner.String(),
		},
		{
			"getApproved",
			getApprovedData,
			operatorWord,
			func(e *ERC721) (any, error) {
				return e.GetApproved(t.Context(), opts, NewGetApprovedRequest(contract, tokenID))
			},
			operator.String(),
		},
		{
			"isApprovedForAll",
			isApprovedForAllData,
			trueWord,
			func(e *ERC721) (any, error) {
				return e.IsApprovedForAll(t.Context(), opts, NewIsApprovedForAllRequest(contract, owner, operator))
			},
			"true",
		},
		{
			"name",
			EncodeName(),
			nameWord,
			func(e *ERC721) (any, error) {
				return e.Name(t.Context(), opts, NewNameRequest(contract))
			},
			"Bored Ape Yacht Club",
		},
		{
			"symbol",
			EncodeSymbol(),
			symbolWord,
			func(e *ERC721) (any, error) {
				return e.Symbol(t.Context(), opts, NewSymbolRequest(contract))
			},
			"BAYC",
		},
		{
			"tokenURI",
			tokenURIData,
			uriWord,
			func(e *ERC721) (any, error) {
				return e.TokenURI(t.Context(), opts, NewTokenURIRequest(contract, tokenID))
			},
			"ipfs://42",
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

	args, err := abi.Pack(errNonexistentToken.Inputs, big.NewInt(42))
	require.NoError(t, err)
	sel := errNonexistentToken.Selector()
	nonexistent := append(append([]byte{}, sel[:]...), args...)

	for _, tc := range []struct {
		name   string
		data   string
		want   error
		wantNo error
	}{
		{"nonexistent token", `"` + encoding.Hex.EncodePrefixed(nonexistent) + `"`, ErrNonexistentToken, ErrUnknown},
		{"unknown selector", `"0xdeadbeef"`, ErrUnknown, ErrNonexistentToken},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(tc.data)}).Once()

			owner, err := New(c).OwnerOf(t.Context(), nil, NewOwnerOfRequest(contract, big.NewInt(42)))
			require.Nil(t, owner)
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
	tokenID := big.NewInt(42)

	approveData, err := EncodeApprove(to, tokenID)
	require.NoError(t, err)
	setApprovalForAllData, err := EncodeSetApprovalForAll(to, true)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(from, to, tokenID)
	require.NoError(t, err)
	safeTransferFromData, err := EncodeSafeTransferFrom(from, to, tokenID)
	require.NoError(t, err)
	safeTransferFromWithData, err := EncodeSafeTransferFromWithData(from, to, tokenID, []byte{0xca, 0xfe})
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"approve",
			approveData,
			func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.Approve(t.Context(), opts, NewApproveRequest(contract, to, tokenID))
			},
		},
		{
			"setApprovalForAll",
			setApprovalForAllData,
			func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SetApprovalForAll(t.Context(), opts, NewSetApprovalForAllRequest(contract, to, true))
			},
		},
		{
			"transferFrom",
			transferFromData,
			func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.TransferFrom(t.Context(), opts, NewTransferFromRequest(contract, from, to, tokenID))
			},
		},
		{
			"safeTransferFrom",
			safeTransferFromData,
			func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SafeTransferFrom(t.Context(), opts, NewSafeTransferFromRequest(contract, from, to, tokenID))
			},
		},
		{
			"safeTransferFromWithData",
			safeTransferFromWithData,
			func(e *ERC721, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return e.SafeTransferFromWithData(t.Context(), opts, NewSafeTransferFromWithDataRequest(contract, from, to, tokenID, []byte{0xca, 0xfe}))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(120_000)
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

	require.NotEqual(t, safeTransferFromData[:4], safeTransferFromWithData[:4])
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	args, err := abi.Pack(errInsufficientApproval.Inputs, to, big.NewInt(42))
	require.NoError(t, err)
	sel := errInsufficientApproval.Selector()
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

	signed, hash, err := New(c).TransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewTransferFromRequest(contract, to, to, big.NewInt(42)))
	require.ErrorIs(t, err, ErrInsufficientApproval)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestTokenOwnersBatchesThroughMulticall(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokenIDs := []*big.Int{big.NewInt(1), big.NewInt(2)}

	ownerWord, err := abi.Pack(abi.NewTypes(abi.Address), owner)
	require.NoError(t, err)
	result, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Tuple(abi.Bool, abi.Bytes))), []any{
		[]any{true, ownerWord},
		[]any{false, []byte{}},
	})
	require.NoError(t, err)
	calldata, err := EncodeTokenOwners(contract, tokenIDs, true)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		call := e.Params.([]any)[0].(map[string]any)
		require.Equal(t, multicall.String(), call["to"])
		require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
		return nil
	}).Once()

	owners, err := New(c).TokenOwners(t.Context(), nil, NewTokenOwnersRequest(multicall, contract, tokenIDs, true))
	require.NoError(t, err)
	require.True(t, owners.Contract.Equal(contract))
	require.Len(t, owners.Tokens, 2)
	require.Zero(t, big.NewInt(1).Cmp(owners.Tokens[0].TokenId))
	require.True(t, owners.Tokens[0].Owner.Equal(owner))
	require.Zero(t, big.NewInt(2).Cmp(owners.Tokens[1].TokenId))
	require.Nil(t, owners.Tokens[1].Owner)
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
			}).Times(4)
			e := New(c)

			addr, err := e.OwnerOf(t.Context(), nil, NewOwnerOfRequest(contract, big.NewInt(1)))
			require.Error(t, err)
			require.Nil(t, addr)
			require.ErrorContains(t, err, "erc721: ownerOf: ")

			n, err := e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(contract, owner))
			require.Error(t, err)
			require.Nil(t, n)

			uri, err := e.TokenURI(t.Context(), nil, NewTokenURIRequest(contract, big.NewInt(1)))
			require.Error(t, err)
			require.Empty(t, uri)

			owners, err := e.TokenOwners(t.Context(), nil, NewTokenOwnersRequest(owner, contract, []*big.Int{big.NewInt(1)}, false))
			require.Error(t, err)
			require.Nil(t, owners)
		})
	}
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "erc721")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	ownerWord, err := abi.Pack(abi.NewTypes(abi.Address), owner)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "erc721", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(ownerWord)
		}
		return nil
	}).Twice()

	_, err = New(c).OwnerOf(ctx, nil, NewOwnerOfRequest(contract, big.NewInt(1)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(90_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).SafeTransferFrom(ctx, bind.NewTransactOptions(key, builder), NewSafeTransferFromRequest(contract, owner, owner, big.NewInt(1)))
	require.NoError(t, err)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(90_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	var raw []byte
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
		require.NoError(t, err)
		return &rpc.ResponseError{Code: -32000, Message: "already known"}
	}).Once()

	signed, hash, err := New(c).Approve(t.Context(), bind.NewTransactOptions(key, builder), NewApproveRequest(contract, to, big.NewInt(3)))
	require.Error(t, err)
	require.ErrorContains(t, err, "erc721: approve: ")
	require.ErrorContains(t, err, "already known")
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
