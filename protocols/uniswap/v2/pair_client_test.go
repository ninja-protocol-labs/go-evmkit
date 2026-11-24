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

func TestPairReadsSendCallAndDecode(t *testing.T) {
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")

	typehash := make([]byte, 32)
	typehash[0] = 0x6e
	typehash[31] = 0xc9

	uintWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1000))
	require.NoError(t, err)
	aWord, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)
	bWord, err := abi.Pack(abi.NewTypes(abi.Address), b)
	require.NoError(t, err)
	reservesWords, err := abi.Pack(abi.NewTypes(abi.Uint112, abi.Uint112, abi.Uint32), big.NewInt(5_000), big.NewInt(7_000), uint32(1_700_000_000))
	require.NoError(t, err)
	price0Word, err := abi.Pack(abi.NewTypes(abi.Uint256), new(big.Int).Lsh(big.NewInt(3), 200))
	require.NoError(t, err)
	price1Word, err := abi.Pack(abi.NewTypes(abi.Uint256), new(big.Int).Lsh(big.NewInt(5), 180))
	require.NoError(t, err)
	kLastWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(35_000_000))
	require.NoError(t, err)
	typehashWord, err := abi.Pack(abi.NewTypes(abi.Bytes32), typehash)
	require.NoError(t, err)
	liquidityWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(4_000))
	require.NoError(t, err)
	burnWords, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	mintData, err := EncodeMint(b)
	require.NoError(t, err)
	burnData, err := EncodeBurn(b)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(p *Pair) (any, error)
		want     any
	}{
		{
			"minimumLiquidity",
			EncodeMinimumLiquidity(),
			uintWord,
			func(p *Pair) (any, error) {
				return p.MinimumLiquidity(t.Context(), opts, NewMinimumLiquidityRequest(pair))
			},
			big.NewInt(1000),
		},
		{
			"factory",
			EncodeFactory(),
			aWord,
			func(p *Pair) (any, error) {
				return p.Factory(t.Context(), opts, NewFactoryRequest(pair))
			},
			a,
		},
		{
			"token0",
			EncodeToken0(),
			aWord,
			func(p *Pair) (any, error) {
				return p.Token0(t.Context(), opts, NewToken0Request(pair))
			},
			a,
		},
		{
			"token1",
			EncodeToken1(),
			bWord,
			func(p *Pair) (any, error) {
				return p.Token1(t.Context(), opts, NewToken1Request(pair))
			},
			b,
		},
		{
			"getReserves",
			EncodeGetReserves(),
			reservesWords,
			func(p *Pair) (any, error) {
				return p.GetReserves(t.Context(), opts, NewGetReservesRequest(pair))
			},
			NewReserves(big.NewInt(5_000), big.NewInt(7_000), 1_700_000_000),
		},
		{
			"price0CumulativeLast",
			EncodePrice0CumulativeLast(),
			price0Word,
			func(p *Pair) (any, error) {
				return p.Price0CumulativeLast(t.Context(), opts, NewPrice0CumulativeLastRequest(pair))
			},
			new(big.Int).Lsh(big.NewInt(3), 200),
		},
		{
			"price1CumulativeLast",
			EncodePrice1CumulativeLast(),
			price1Word,
			func(p *Pair) (any, error) {
				return p.Price1CumulativeLast(t.Context(), opts, NewPrice1CumulativeLastRequest(pair))
			},
			new(big.Int).Lsh(big.NewInt(5), 180),
		},
		{
			"kLast",
			EncodeKLast(),
			kLastWord,
			func(p *Pair) (any, error) {
				return p.KLast(t.Context(), opts, NewKLastRequest(pair))
			},
			big.NewInt(35_000_000),
		},
		{
			"permitTypehash",
			EncodePermitTypehash(),
			typehashWord,
			func(p *Pair) (any, error) {
				return p.PermitTypehash(t.Context(), opts, NewPermitTypehashRequest(pair))
			},
			typehash,
		},
		{
			"callMint",
			mintData,
			liquidityWord,
			func(p *Pair) (any, error) {
				return p.CallMint(t.Context(), opts, NewMintRequest(pair, b))
			},
			big.NewInt(4_000),
		},
		{
			"callBurn",
			burnData,
			burnWords,
			func(p *Pair) (any, error) {
				amount0, amount1, err := p.CallBurn(t.Context(), opts, NewBurnRequest(pair, b))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(100), big.NewInt(200)},
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
				require.Equal(t, pair.String(), call["to"])
				require.Equal(t, a.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewPair(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestPairReadsMapRevertToSentinel(t *testing.T) {
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}
	all := []error{
		ErrLocked, ErrTransferFailed, ErrForbidden, ErrOverflow,
		ErrInsufficientLiquidityMinted, ErrInsufficientLiquidityBurned,
		ErrInsufficientOutputAmount, ErrInsufficientLiquidity, ErrInvalidTo,
		ErrInsufficientInputAmount, ErrK, ErrAddOverflow, ErrSubUnderflow,
		ErrMulOverflow, ErrUnknown,
	}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"locked", reasonLocked, ErrLocked},
		{"transfer failed", reasonTransferFailed, ErrTransferFailed},
		{"forbidden", reasonForbidden, ErrForbidden},
		{"overflow", reasonOverflow, ErrOverflow},
		{"insufficient liquidity minted", reasonInsufficientLiquidityMinted, ErrInsufficientLiquidityMinted},
		{"insufficient liquidity burned", reasonInsufficientLiquidityBurned, ErrInsufficientLiquidityBurned},
		{"insufficient output amount", reasonInsufficientOutputAmount, ErrInsufficientOutputAmount},
		{"insufficient liquidity", reasonInsufficientLiquidity, ErrInsufficientLiquidity},
		{"invalid to", reasonInvalidTo, ErrInvalidTo},
		{"insufficient input amount", reasonInsufficientInputAmount, ErrInsufficientInputAmount},
		{"k", reasonK, ErrK},
		{"add overflow", reasonAddOverflow, ErrAddOverflow},
		{"sub underflow", reasonSubUnderflow, ErrSubUnderflow},
		{"mul overflow", reasonMulOverflow, ErrMulOverflow},
		{"other reason", "UniswapV2: SOMETHING_ELSE", ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Twice()
			p := NewPair(c)

			liquidity, err := p.CallMint(t.Context(), bind.NewCallOptions().WithFrom(a), NewMintRequest(pair, a))
			require.Nil(t, liquidity)
			require.ErrorIs(t, err, tc.want)
			for _, other := range all {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)

			amount0, amount1, err := p.CallBurn(t.Context(), nil, NewBurnRequest(pair, a))
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv2: callBurn: ")
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`), jsontext.Value(`"0xzz"`)} {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		reserves, err := NewPair(c).GetReserves(t.Context(), nil, NewGetReservesRequest(pair))
		require.Nil(t, reserves)
		require.ErrorContains(t, err, "uniswapv2: getReserves: ")
		require.NotErrorIs(t, err, ErrUnknown)
	}
}

func TestPairReadsRejectMalformedResult(t *testing.T) {
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(5)
			p := NewPair(c)

			reserves, err := p.GetReserves(t.Context(), nil, NewGetReservesRequest(pair))
			require.Nil(t, reserves)
			require.ErrorContains(t, err, "uniswapv2: getReserves: ")

			token0, err := p.Token0(t.Context(), nil, NewToken0Request(pair))
			require.Nil(t, token0)
			require.ErrorContains(t, err, "uniswapv2: token0: ")

			typehash, err := p.PermitTypehash(t.Context(), nil, NewPermitTypehashRequest(pair))
			require.Nil(t, typehash)
			require.ErrorContains(t, err, "uniswapv2: permitTypehash: ")

			liquidity, err := p.CallMint(t.Context(), nil, NewMintRequest(pair, a))
			require.Nil(t, liquidity)
			require.ErrorContains(t, err, "uniswapv2: callMint: ")

			amount0, amount1, err := p.CallBurn(t.Context(), nil, NewBurnRequest(pair, a))
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorContains(t, err, "uniswapv2: callBurn: ")
		})
	}
}

func TestPairMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	p := NewPair(mocks.NewClient(t))
	ctx := t.Context()

	_, err = p.MinimumLiquidity(ctx, nil, &MinimumLiquidityRequest{})
	require.EqualError(t, err, "uniswapv2: minimumLiquidity: "+(&MinimumLiquidityRequest{}).Validate().Error())
	_, err = p.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv2: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = p.Token0(ctx, nil, &Token0Request{})
	require.EqualError(t, err, "uniswapv2: token0: "+(&Token0Request{}).Validate().Error())
	_, err = p.Token1(ctx, nil, &Token1Request{})
	require.EqualError(t, err, "uniswapv2: token1: "+(&Token1Request{}).Validate().Error())
	_, err = p.GetReserves(ctx, nil, &GetReservesRequest{})
	require.EqualError(t, err, "uniswapv2: getReserves: "+(&GetReservesRequest{}).Validate().Error())
	_, err = p.Price0CumulativeLast(ctx, nil, &Price0CumulativeLastRequest{})
	require.EqualError(t, err, "uniswapv2: price0CumulativeLast: "+(&Price0CumulativeLastRequest{}).Validate().Error())
	_, err = p.Price1CumulativeLast(ctx, nil, &Price1CumulativeLastRequest{})
	require.EqualError(t, err, "uniswapv2: price1CumulativeLast: "+(&Price1CumulativeLastRequest{}).Validate().Error())
	_, err = p.KLast(ctx, nil, &KLastRequest{})
	require.EqualError(t, err, "uniswapv2: kLast: "+(&KLastRequest{}).Validate().Error())
	_, err = p.PermitTypehash(ctx, nil, &PermitTypehashRequest{})
	require.EqualError(t, err, "uniswapv2: permitTypehash: "+(&PermitTypehashRequest{}).Validate().Error())
	_, err = p.CallMint(ctx, nil, NewMintRequest(pair, nil))
	require.EqualError(t, err, "uniswapv2: callMint: "+NewMintRequest(pair, nil).Validate().Error())
	_, _, err = p.CallBurn(ctx, nil, NewBurnRequest(nil, a))
	require.EqualError(t, err, "uniswapv2: callBurn: "+NewBurnRequest(nil, a).Validate().Error())

	_, _, err = p.Initialize(ctx, opts, NewInitializeRequest(pair, a, nil))
	require.EqualError(t, err, "uniswapv2: initialize: "+NewInitializeRequest(pair, a, nil).Validate().Error())
	_, _, err = p.Mint(ctx, opts, NewMintRequest(nil, a))
	require.EqualError(t, err, "uniswapv2: mint: "+NewMintRequest(nil, a).Validate().Error())
	_, _, err = p.Burn(ctx, opts, NewBurnRequest(pair, nil))
	require.EqualError(t, err, "uniswapv2: burn: "+NewBurnRequest(pair, nil).Validate().Error())
	_, _, err = p.Swap(ctx, opts, NewSwapRequest(pair, big.NewInt(0), nil, a, nil))
	require.EqualError(t, err, "uniswapv2: swap: "+NewSwapRequest(pair, big.NewInt(0), nil, a, nil).Validate().Error())
	_, _, err = p.Skim(ctx, opts, NewSkimRequest(pair, nil))
	require.EqualError(t, err, "uniswapv2: skim: "+NewSkimRequest(pair, nil).Validate().Error())
	_, _, err = p.Sync(ctx, opts, &SyncRequest{})
	require.EqualError(t, err, "uniswapv2: sync: "+(&SyncRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = p.Swap(ctx, opts, NewSwapRequest(pair, big.NewInt(-1), big.NewInt(0), a, nil))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = p.Sync(ctx, nil, NewSyncRequest(pair))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv2: sync: bind: transact: invalid options: options is nil")
}

func TestPairWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	initializeData, err := EncodeInitialize(a, b)
	require.NoError(t, err)
	mintData, err := EncodeMint(a)
	require.NoError(t, err)
	burnData, err := EncodeBurn(b)
	require.NoError(t, err)
	swapData, err := EncodeSwap(big.NewInt(0), big.NewInt(95), a, []byte{0xca, 0xfe})
	require.NoError(t, err)
	skimData, err := EncodeSkim(b)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"initialize",
			initializeData,
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Initialize(t.Context(), opts, NewInitializeRequest(pair, a, b))
			},
		},
		{
			"mint",
			mintData,
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Mint(t.Context(), opts, NewMintRequest(pair, a))
			},
		},
		{
			"burn",
			burnData,
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Burn(t.Context(), opts, NewBurnRequest(pair, b))
			},
		},
		{
			"swap",
			swapData,
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Swap(t.Context(), opts, NewSwapRequest(pair, big.NewInt(0), big.NewInt(95), a, []byte{0xca, 0xfe}))
			},
		},
		{
			"skim",
			skimData,
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Skim(t.Context(), opts, NewSkimRequest(pair, b))
			},
		},
		{
			"sync",
			EncodeSync(),
			func(p *Pair, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Sync(t.Context(), opts, NewSyncRequest(pair))
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

			signed, hash, err := tc.send(NewPair(c), bind.NewTransactOptions(key, builder))
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(pair))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestPairWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonK)
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

	signed, hash, err := NewPair(c).Swap(t.Context(), bind.NewTransactOptions(key, builder), NewSwapRequest(pair, big.NewInt(0), big.NewInt(1), a, nil))
	require.ErrorIs(t, err, ErrK)
	require.NotErrorIs(t, err, ErrUnknown)
	require.ErrorContains(t, err, "uniswapv2: swap: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestPairWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Twice()

	signed, hash, err := NewPair(c).Mint(t.Context(), bind.NewTransactOptions(key, builder), NewMintRequest(pair, a))
	require.ErrorContains(t, err, "uniswapv2: mint: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestPairCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "pair")
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "pair", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewPair(c).KLast(ctx, nil, NewKLastRequest(pair))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewPair(c).Sync(ctx, bind.NewTransactOptions(key, builder), NewSyncRequest(pair))
	require.NoError(t, err)
}
