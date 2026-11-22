package weth

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzEncodeWithdrawRoundTrip$' -fuzztime=10s ./protocols/weth
func FuzzEncodeWithdrawRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, wadRaw []byte) {
		if len(wadRaw) > 32 {
			wadRaw = wadRaw[len(wadRaw)-32:]
		}
		wad := new(big.Int).SetBytes(wadRaw)

		data, err := EncodeWithdraw(wad)
		require.NoError(t, err)
		require.Equal(t, []byte{0x2e, 0x1a, 0x7d, 0x4d}, data[:4])

		vals, err := withdrawFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, wad.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzWETHSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/weth
func FuzzWETHSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = withdrawFn.DecodeCall(data)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractEventsDoNotPanic$' -fuzztime=10s ./protocols/weth
func FuzzExtractEventsDoNotPanic(f *testing.F) {
	f.Add(uint8(0), []byte{}, []byte{})
	f.Add(uint8(2), make([]byte, 64), make([]byte, 32))
	f.Add(uint8(3), []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x01})

	f.Fuzz(func(t *testing.T, topicCount uint8, topicsRaw, data []byte) {
		n := int(topicCount % 4)
		topics := make([]*types.Hash, n)
		for i := range topics {
			start := i * 32
			if start >= len(topicsRaw) {
				topics[i] = types.NewHashFromBytes(nil)
				continue
			}
			end := min(start+32, len(topicsRaw))
			topics[i] = types.NewHashFromBytes(topicsRaw[start:end])
		}
		logs := []tx.Log{{Address: testWETH, Topics: topics, Data: data}}

		_, _ = ExtractDeposits(logs)
		_, _ = ExtractWithdrawals(logs)
	})
}

// go test -run '^$' -fuzz '^FuzzDepositRequestValidate$' -fuzztime=10s ./protocols/weth
func FuzzDepositRequestValidate(f *testing.F) {
	f.Add(uint8(0), int64(1))
	f.Add(uint8(0x03), int64(0))
	f.Add(uint8(0x03), int64(-1))
	f.Add(uint8(0x04), int64(1))

	contract := types.MustNewAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	f.Fuzz(func(t *testing.T, mask uint8, amount int64) {
		var req *DepositRequest
		if mask&0x04 == 0 {
			req = &DepositRequest{}
			if mask&0x01 != 0 {
				req.Contract = contract
			}
			if mask&0x02 != 0 {
				req.Amount = big.NewInt(amount)
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case req.Amount == nil:
			want = "invalid request: amount is nil"
		case req.Amount.Sign() < 0:
			want = "invalid request: amount is negative"
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

// go test -run '^$' -fuzz '^FuzzDepositThroughClient$' -fuzztime=10s ./protocols/weth
func FuzzDepositThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x0d, 0xe0, 0xb6, 0xb3}, uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 31), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, contractRaw, amountRaw []byte, nonce uint64, dynamic bool) {
		if len(amountRaw) > 31 {
			amountRaw = amountRaw[len(amountRaw)-31:]
		}
		contract := types.NewAddressFromBytes(contractRaw)
		amount := new(big.Int).SetBytes(amountRaw)

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(60_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		opts := bind.NewTransactOptions(key, builder)
		signed, hash, err := New(c).Deposit(t.Context(), opts, NewDepositRequest(contract, amount))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Nil(t, opts.Value)

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotValue *big.Int
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(contract))
		require.Equal(t, amount.String(), gotValue.String())
		require.Equal(t, EncodeDeposit(), gotData)
	})
}
