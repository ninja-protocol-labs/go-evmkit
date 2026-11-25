package feeds

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzEncodeGetRoundDataRoundTrip$' -fuzztime=10s ./protocols/chainlink/feeds
func FuzzEncodeGetRoundDataRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x8d, 0xfc})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, roundIDRaw []byte) {
		if len(roundIDRaw) > 10 {
			roundIDRaw = roundIDRaw[len(roundIDRaw)-10:]
		}
		roundID := new(big.Int).SetBytes(roundIDRaw)

		data, err := EncodeGetRoundData(roundID)
		require.NoError(t, err)

		vals, err := getRoundDataFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, roundID.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzRoundDataRoundTrip$' -fuzztime=10s ./protocols/chainlink/feeds
func FuzzRoundDataRoundTrip(f *testing.F) {
	f.Add(uint64(1), int64(165_327_740_000), uint64(1_795_586_656), uint64(1_795_586_675), uint16(7))
	f.Add(uint64(0), int64(-1), uint64(0), uint64(0), uint16(0))

	f.Fuzz(func(t *testing.T, aggregatorRound uint64, answerRaw int64, startedAtRaw, updatedAtRaw uint64, phase uint16) {
		roundID := new(big.Int).Or(new(big.Int).Lsh(new(big.Int).SetUint64(uint64(phase)), 64), new(big.Int).SetUint64(aggregatorRound))
		answer := big.NewInt(answerRaw)
		startedAt := new(big.Int).SetUint64(startedAtRaw)
		updatedAt := new(big.Int).SetUint64(updatedAtRaw)

		data, err := latestRoundDataFn.EncodeReturn(roundID, answer, startedAt, updatedAt, roundID)
		require.NoError(t, err)

		got, err := DecodeLatestRoundData(data)
		require.NoError(t, err)
		require.Zero(t, roundID.Cmp(got.RoundID))
		require.Zero(t, answer.Cmp(got.Answer))
		require.Zero(t, startedAt.Cmp(got.StartedAt))
		require.Zero(t, updatedAt.Cmp(got.UpdatedAt))
		require.Zero(t, roundID.Cmp(got.AnsweredInRound))
	})
}

// go test -run '^$' -fuzz '^FuzzAggregatorDecodersDoNotPanic$' -fuzztime=10s ./protocols/chainlink/feeds
func FuzzAggregatorDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 5*32))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeDecimals(data)
		_, _ = DecodeDescription(data)
		_, _ = DecodeVersion(data)
		_, _ = DecodeGetRoundData(data)
		_, _ = DecodeLatestRoundData(data)
	})
}

// go test -run '^$' -fuzz '^FuzzGetRoundDataRequestValidate$' -fuzztime=10s ./protocols/chainlink/feeds
func FuzzGetRoundDataRequestValidate(f *testing.F) {
	f.Add(uint8(0), []byte{0x01}, false)
	f.Add(uint8(0x03), []byte{0x01}, true)
	f.Add(uint8(0x03), make([]byte, 11), false)
	f.Add(uint8(0x04), []byte{}, false)

	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")

	f.Fuzz(func(t *testing.T, mask uint8, roundIDRaw []byte, negative bool) {
		var req *GetRoundDataRequest
		if mask&0x04 == 0 {
			req = &GetRoundDataRequest{}
			if mask&0x01 != 0 {
				req.Contract = feed
			}
			if mask&0x02 != 0 {
				req.RoundID = new(big.Int).SetBytes(roundIDRaw)
				if negative {
					req.RoundID.Neg(req.RoundID)
				}
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Contract == nil:
			want = "invalid request: contract is nil"
		case req.RoundID == nil:
			want = "invalid request: roundId is nil"
		case req.RoundID.Sign() < 0:
			want = "invalid request: roundId is negative"
		case req.RoundID.BitLen() > 80:
			want = "invalid request: roundId exceeds uint80"
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			_, err = EncodeGetRoundData(req.RoundID)
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzAggregatorReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/chainlink/feeds
func FuzzAggregatorReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 5*32))

	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")

	f.Fuzz(func(t *testing.T, result []byte) {
		cli := mocks.NewClient(t)
		cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(4)
		a := NewAggregator(cli)

		version, err := a.Version(t.Context(), nil, NewVersionRequest(feed))
		require.True(t, (err == nil) != (version == nil))
		round, err := a.GetRoundData(t.Context(), nil, NewGetRoundDataRequest(feed, big.NewInt(1)))
		require.True(t, (err == nil) != (round == nil))
		latest, err := a.LatestRoundData(t.Context(), nil, NewLatestRoundDataRequest(feed))
		require.True(t, (err == nil) != (latest == nil))
		if latest != nil {
			require.NotZero(t, latest.UpdatedAt.Sign())
		}
		_, _ = a.Description(t.Context(), nil, NewDescriptionRequest(feed))
	})
}
