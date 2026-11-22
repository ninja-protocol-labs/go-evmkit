package erc20

import (
	"context"
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

var (
	fuzzToken1 = types.MustNewAddressFromHex("0x4200000000000000000000000000000000000006")
	fuzzToken2 = types.MustNewAddressFromHex("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913")
	fuzzAddr1  = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fuzzAddr2  = types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
)

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./standards/erc20
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	nameData, err := abi.Pack(abi.NewTypes(abi.String), "USD Coin")
	require.NoError(f, err)
	f.Add(nameData)

	uintData, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(123456789))
	require.NoError(f, err)
	f.Add(uintData)

	boolData, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(f, err)
	f.Add(boolData)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeName(data)
		_, _ = DecodeSymbol(data)
		_, _ = DecodeDecimals(data)
		_, _ = DecodeTotalSupply(data)
		_, _ = DecodeBalanceOf(data)
		_, _ = DecodeAllowance(data)
		_, _ = DecodeTransfer(data)
		_, _ = DecodeApprove(data)
		_, _ = DecodeTransferFrom(data)
	})
}

// go test -run '^$' -fuzz '^FuzzBatchDecodersDoNotPanic$' -fuzztime=10s ./standards/erc20
func FuzzBatchDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 192))

	if seed, err := EncodeMetadataWithBalance(fuzzToken1, fuzzAddr1, false); err == nil {
		f.Add(seed)
	}
	if seed, err := EncodeAllowanceWithBalance(fuzzToken1, fuzzAddr1, fuzzAddr2, false); err == nil {
		f.Add(seed)
	}

	addrs := []*types.Address{fuzzAddr1, fuzzAddr2}
	toks := []*types.Address{fuzzToken1, fuzzToken2}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeMetadata(data)
		_, _ = DecodeMetadataWithBalance(data)
		_, _ = DecodeAllowanceWithBalance(data)
		_, _ = DecodeTokenBalances(fuzzToken1, addrs, data)
		_, _ = DecodeAddressBalances(fuzzAddr1, toks, data)
		_, _ = DecodeBalancePairs(toks, addrs, data)
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeBalanceOfRoundTrip$' -fuzztime=10s ./standards/erc20
func FuzzDecodeBalanceOfRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32 {
			raw = raw[len(raw)-32:]
		}
		value := new(big.Int).SetBytes(raw)

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), value)
		require.NoError(t, err)

		got, err := DecodeBalanceOf(data)
		require.NoError(t, err)
		require.Zero(t, value.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersCrashSafety$' -fuzztime=10s ./standards/erc20
func FuzzExtractTransfersCrashSafety(f *testing.F) {
	t0, ok := transferEvent.Topic0()
	require.True(f, ok)

	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x2a})
	f.Add([]byte{}, []byte{}, []byte{}, []byte{})
	f.Add(make([]byte, 32), make([]byte, 20), make([]byte, 20), make([]byte, 32))
	f.Add(t0[:], fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x01})

	f.Fuzz(func(t *testing.T, topic0Raw, topic1Raw, topic2Raw, data []byte) {
		h0 := types.NewHashFromBytes(topic0Raw)
		h1 := types.NewHashFromBytes(topic1Raw)
		h2 := types.NewHashFromBytes(topic2Raw)

		variants := [][]*types.Hash{
			nil,
			{h0},
			{h0, h1},
			{h0, h1, h2},
		}
		for _, topics := range variants {
			logs := []tx.Log{{Address: fuzzToken1, Topics: topics, Data: data}}
			_, _ = ExtractTransfers(logs)
			_, _ = ExtractApprovals(logs)
		}
	})
}

// go test -run '^$' -fuzz '^FuzzExtractTransfersRoundTrip$' -fuzztime=10s ./standards/erc20
func FuzzExtractTransfersRoundTrip(f *testing.F) {
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x01, 0x02, 0x03})
	f.Add(make([]byte, 20), make([]byte, 20), []byte{})

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, valueRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		value := new(big.Int).SetBytes(valueRaw)

		data, err := abi.Pack(abi.NewTypes(abi.Uint256), value)
		require.NoError(t, err)

		t0, ok := transferEvent.Topic0()
		require.True(t, ok)

		logs := []tx.Log{
			{
				Address: fuzzToken1,
				Topics: []*types.Hash{
					types.NewHashFromBytes(t0[:]),
					types.NewHashFromBytes(from.Bytes()),
					types.NewHashFromBytes(to.Bytes()),
				},
				Data: data,
			},
		}

		transfers, err := ExtractTransfers(logs)
		require.NoError(t, err)
		require.Len(t, transfers, 1)
		require.True(t, transfers[0].Contract.Equal(fuzzToken1))
		require.True(t, transfers[0].From.Equal(from))
		require.True(t, transfers[0].To.Equal(to))
		require.Zero(t, value.Cmp(transfers[0].Value))

		approvals, err := ExtractApprovals(logs)
		require.NoError(t, err)
		require.Empty(t, approvals)
	})
}

// go test -run '^$' -fuzz '^FuzzTransferFromRequestValidate$' -fuzztime=10s ./standards/erc20
func FuzzTransferFromRequestValidate(f *testing.F) {
	f.Add(uint8(0), []byte{})
	f.Add(uint8(0x0f), []byte{0x01})
	f.Add(uint8(0x1f), []byte{0xff, 0xff})
	f.Add(uint8(0x07), []byte{})
	f.Add(uint8(0x10), []byte{0x00})

	f.Fuzz(func(t *testing.T, mask uint8, amountRaw []byte) {
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[:32]
		}

		var req *TransferFromRequest
		if mask&0x10 == 0 {
			req = &TransferFromRequest{}
			if mask&0x01 != 0 {
				req.Contract = fuzzToken1
			}
			if mask&0x02 != 0 {
				req.From = fuzzAddr1
			}
			if mask&0x04 != 0 {
				req.To = fuzzAddr2
			}
			if mask&0x08 != 0 {
				req.Amount = new(big.Int).SetBytes(amountRaw)
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case req.From == nil:
			want = "invalid request: from is nil"
		case req.To == nil:
			want = "invalid request: to is nil"
		case req.Amount == nil:
			want = "invalid request: amount is nil"
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzBalancePairsRequestValidate$' -fuzztime=10s ./standards/erc20
func FuzzBalancePairsRequestValidate(f *testing.F) {
	f.Add(true, uint8(0), uint8(0), uint8(0), uint8(0))
	f.Add(true, uint8(2), uint8(2), uint8(9), uint8(9))
	f.Add(true, uint8(3), uint8(2), uint8(9), uint8(9))
	f.Add(true, uint8(3), uint8(3), uint8(1), uint8(9))
	f.Add(true, uint8(3), uint8(3), uint8(9), uint8(0))
	f.Add(false, uint8(1), uint8(1), uint8(9), uint8(9))

	f.Fuzz(func(t *testing.T, withMulticall bool, tokenLen, accountLen, nilToken, nilAccount uint8) {
		tokenLen %= 8
		accountLen %= 8

		tokens := make([]*types.Address, tokenLen)
		for i := range tokens {
			tokens[i] = fuzzToken1
			if uint8(i) == nilToken {
				tokens[i] = nil
			}
		}
		accounts := make([]*types.Address, accountLen)
		for i := range accounts {
			accounts[i] = fuzzAddr1
			if uint8(i) == nilAccount {
				accounts[i] = nil
			}
		}

		var multicall *types.Address
		if withMulticall {
			multicall = fuzzToken2
		}
		req := NewBalancePairsRequest(multicall, tokens, accounts, nilToken%2 == 0)

		want := ""
		switch {
		case multicall == nil:
			want = "invalid request: multicall is nil"
		case tokenLen != accountLen:
			want = "invalid request: tokens and accounts differ in length"
		case tokenLen == 0:
			want = "invalid request: tokens is empty"
		case nilToken < tokenLen && nilToken <= nilAccount:
			want = fmt.Sprintf("invalid request: tokens[%d] is nil", nilToken)
		case nilAccount < accountLen:
			want = fmt.Sprintf("invalid request: accounts[%d] is nil", nilAccount)
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzBalanceOfThroughClient$' -fuzztime=10s ./standards/erc20
func FuzzBalanceOfThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, uint64(0))
	f.Add(fuzzToken1.Bytes(), fuzzAddr1.Bytes(), []byte{0x01}, uint64(1))
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 32), ^uint64(0))
	f.Add([]byte{0xff}, []byte{0xee}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, uint64(1<<40))

	f.Fuzz(func(t *testing.T, tokenRaw, accountRaw, amountRaw []byte, block uint64) {
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		token := types.NewAddressFromBytes(tokenRaw)
		account := types.NewAddressFromBytes(accountRaw)
		amount := new(big.Int).SetBytes(amountRaw)
		blockHex := "0x" + new(big.Int).SetUint64(block).Text(16)

		calldata, err := EncodeBalanceOf(account)
		require.NoError(t, err)
		result, err := abi.Pack(abi.NewTypes(abi.Uint256), amount)
		require.NoError(t, err)

		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			params := e.Params.([]any)
			call := params[0].(map[string]any)
			require.Equal(t, token.String(), call["to"])
			require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
			require.NotContains(t, call, "from")
			require.Equal(t, blockHex, params[1])
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Once()

		got, err := New(c).BalanceOf(t.Context(), bind.NewCallOptions().WithBlock(blockHex), NewBalanceOfRequest(token, account))
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzTransferFromThroughClient$' -fuzztime=10s ./standards/erc20
func FuzzTransferFromThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, uint64(0), uint64(21_000), true)
	f.Add(fuzzAddr1.Bytes(), fuzzAddr2.Bytes(), []byte{0x01}, uint64(42), uint64(100_000), false)
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 32), ^uint64(0)-1, uint64(30_000_000), true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, amountRaw []byte, nonce, gasLimit uint64, dynamic bool) {
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		amount := new(big.Int).SetBytes(amountRaw)

		calldata, err := EncodeTransferFrom(from, to, amount)
		require.NoError(t, err)

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(gasLimit)
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

		signed, hash, err := New(c).TransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewTransferFromRequest(fuzzToken1, from, to, amount))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Equal(t, gasLimit, signed.Gas())

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		got, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, got.Equal(sender))

		var gotTo *types.Address
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotData = v.To, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotData = v.To, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(fuzzToken1))
		require.Equal(t, calldata, gotData)
	})
}

// go test -run '^$' -fuzz '^FuzzReadsSurviveArbitraryResult$' -fuzztime=10s ./standards/erc20
func FuzzReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 96))
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(8)
		e := New(c)

		n, err := e.TotalSupply(t.Context(), nil, NewTotalSupplyRequest(fuzzToken1))
		require.True(t, (err == nil) != (n == nil))
		n, err = e.BalanceOf(t.Context(), nil, NewBalanceOfRequest(fuzzToken1, fuzzAddr1))
		require.True(t, (err == nil) != (n == nil))
		_, _ = e.Name(t.Context(), nil, NewNameRequest(fuzzToken1))
		_, _ = e.Decimals(t.Context(), nil, NewDecimalsRequest(fuzzToken1))
		_, _ = e.CallTransfer(t.Context(), nil, NewTransferRequest(fuzzToken1, fuzzAddr1, big.NewInt(1)))
		m, err := e.Metadata(t.Context(), nil, NewMetadataRequest(fuzzAddr2, fuzzToken1, true))
		require.True(t, (err == nil) != (m == nil))
		tb, err := e.TokenBalances(t.Context(), nil, NewTokenBalancesRequest(fuzzAddr2, fuzzToken1, []*types.Address{fuzzAddr1, fuzzAddr2}, true))
		require.True(t, (err == nil) != (tb == nil))
		pairs, err := e.BalancePairs(t.Context(), nil, NewBalancePairsRequest(fuzzAddr2, []*types.Address{fuzzToken1}, []*types.Address{fuzzAddr1}, false))
		if err == nil {
			require.Len(t, pairs, 1)
		}
	})
}
