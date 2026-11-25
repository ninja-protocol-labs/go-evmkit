package feeds

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestAggregatorReadsSendCallAndDecode(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4")
	roundID := new(big.Int).Or(new(big.Int).Lsh(big.NewInt(7), 64), big.NewInt(36348))
	round := NewRoundData(roundID, big.NewInt(165_327_740_000), big.NewInt(1_795_586_656), big.NewInt(1_795_586_675), roundID)

	decimalsResult, err := decimalsFn.EncodeReturn(uint8(8))
	require.NoError(t, err)
	descriptionResult, err := descriptionFn.EncodeReturn("ETH / USD")
	require.NoError(t, err)
	versionResult, err := versionFn.EncodeReturn(big.NewInt(6))
	require.NoError(t, err)
	roundResult, err := latestRoundDataFn.EncodeReturn(
		round.RoundID,
		round.Answer,
		round.StartedAt,
		round.UpdatedAt,
		round.AnsweredInRound,
	)
	require.NoError(t, err)
	getRoundDataCalldata, err := EncodeGetRoundData(roundID)
	require.NoError(t, err)

	var wantCalldata, result []byte
	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		require.Equal(t, rpc.MethodETHCall, e.Method)
		params, ok := e.Params.([]any)
		require.True(t, ok)
		require.Len(t, params, 2)
		call, ok := params[0].(map[string]any)
		require.True(t, ok)
		require.Equal(t, feed.String(), call["to"])
		require.Equal(t, from.String(), call["from"])
		require.Equal(t, encoding.Hex.EncodePrefixed(wantCalldata), call["data"])
		require.NotContains(t, call, "value")
		require.Equal(t, "0x1b4", params[1])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
		return nil
	}).Times(5)
	a := NewAggregator(cli)

	wantCalldata, result = EncodeDecimals(), decimalsResult
	decimals, err := a.Decimals(t.Context(), opts, NewDecimalsRequest(feed))
	require.NoError(t, err)
	require.Equal(t, uint8(8), decimals)

	wantCalldata, result = EncodeDescription(), descriptionResult
	description, err := a.Description(t.Context(), opts, NewDescriptionRequest(feed))
	require.NoError(t, err)
	require.Equal(t, "ETH / USD", description)

	wantCalldata, result = EncodeVersion(), versionResult
	version, err := a.Version(t.Context(), opts, NewVersionRequest(feed))
	require.NoError(t, err)
	require.Zero(t, big.NewInt(6).Cmp(version))

	wantCalldata, result = EncodeLatestRoundData(), roundResult
	latest, err := a.LatestRoundData(t.Context(), opts, NewLatestRoundDataRequest(feed))
	require.NoError(t, err)
	require.Equal(t, round, latest)

	wantCalldata, result = getRoundDataCalldata, roundResult
	got, err := a.GetRoundData(t.Context(), opts, NewGetRoundDataRequest(feed, roundID))
	require.NoError(t, err)
	require.Equal(t, round, got)
}

func TestAggregatorRoundReadsRejectZeroUpdatedAt(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	roundID := new(big.Int).Or(new(big.Int).Lsh(big.NewInt(7), 64), big.NewInt(36349))
	zero, err := latestRoundDataFn.EncodeReturn(
		new(big.Int).Lsh(big.NewInt(7), 64),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
	)
	require.NoError(t, err)

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(zero)
		return nil
	}).Times(2)
	a := NewAggregator(cli)

	round, err := a.GetRoundData(t.Context(), nil, NewGetRoundDataRequest(feed, roundID))
	require.Nil(t, round)
	require.ErrorIs(t, err, ErrNoDataPresent)
	require.EqualError(t, err, "chainlinkfeeds: getRoundData: no data present: round "+roundID.String()+" has zero updatedAt")

	latest, err := a.LatestRoundData(t.Context(), nil, NewLatestRoundDataRequest(feed))
	require.Nil(t, latest)
	require.ErrorIs(t, err, ErrNoDataPresent)
	require.EqualError(t, err, "chainlinkfeeds: latestRoundData: no data present: latest round has zero updatedAt")
}

func TestAggregatorReadsMapRevertToSentinel(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x694AA1769357215DE4FAC081bf1f309aDC325306")
	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonNoDataPresent)
	require.NoError(t, err)
	noDataRevert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

	cli := mocks.NewClient(t)
	noDataErr := &rpc.ResponseError{
		Code:    3,
		Message: "execution reverted: No data present",
		Data:    jsontext.Value(`"` + encoding.Hex.EncodePrefixed(noDataRevert) + `"`),
	}
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(noDataErr).Once()
	req := NewGetRoundDataRequest(feed, new(big.Int).Lsh(big.NewInt(1), 70))
	round, err := NewAggregator(cli).GetRoundData(t.Context(), nil, req)
	require.Nil(t, round)
	require.ErrorIs(t, err, ErrNoDataPresent)
	require.ErrorContains(t, err, "chainlinkfeeds: getRoundData: no data present: ")
	var rpcErr *rpc.ResponseError
	require.ErrorAs(t, err, &rpcErr)

	cli = mocks.NewClient(t)
	emptyErr := &rpc.ResponseError{
		Code:    3,
		Message: "execution reverted",
		Data:    jsontext.Value(`"0x"`),
	}
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(emptyErr).Once()
	round, err = NewAggregator(cli).GetRoundData(t.Context(), nil, NewGetRoundDataRequest(feed, big.NewInt(1)))
	require.Nil(t, round)
	require.ErrorIs(t, err, ErrReverted)
	require.NotErrorIs(t, err, ErrNoDataPresent)
	require.ErrorContains(t, err, "chainlinkfeeds: getRoundData: reverted: ")

	cli = mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted"}).Once()
	description, err := NewAggregator(cli).Description(t.Context(), nil, NewDescriptionRequest(feed))
	require.Empty(t, description)
	require.ErrorIs(t, err, ErrReverted)
	require.ErrorContains(t, err, "chainlinkfeeds: description: reverted: ")

	cli = mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()
	decimals, err := NewAggregator(cli).Decimals(t.Context(), nil, NewDecimalsRequest(feed))
	require.Zero(t, decimals)
	require.EqualError(t, err, "chainlinkfeeds: decimals: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestAggregatorReadsRejectMalformedResult(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			cli := mocks.NewClient(t)
			cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(5)
			a := NewAggregator(cli)

			decimals, err := a.Decimals(t.Context(), nil, NewDecimalsRequest(feed))
			require.Zero(t, decimals)
			require.ErrorContains(t, err, "chainlinkfeeds: decimals: ")

			description, err := a.Description(t.Context(), nil, NewDescriptionRequest(feed))
			require.Empty(t, description)
			require.ErrorContains(t, err, "chainlinkfeeds: description: ")

			version, err := a.Version(t.Context(), nil, NewVersionRequest(feed))
			require.Nil(t, version)
			require.ErrorContains(t, err, "chainlinkfeeds: version: ")

			round, err := a.GetRoundData(t.Context(), nil, NewGetRoundDataRequest(feed, big.NewInt(1)))
			require.Nil(t, round)
			require.ErrorContains(t, err, "chainlinkfeeds: getRoundData: ")

			latest, err := a.LatestRoundData(t.Context(), nil, NewLatestRoundDataRequest(feed))
			require.Nil(t, latest)
			require.ErrorContains(t, err, "chainlinkfeeds: latestRoundData: ")
		})
	}
}

func TestAggregatorMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	a := NewAggregator(mocks.NewClient(t))
	ctx := t.Context()

	_, err := a.Decimals(ctx, nil, &DecimalsRequest{})
	require.EqualError(t, err, "chainlinkfeeds: decimals: invalid request: contract is nil")
	_, err = a.Description(ctx, nil, nil)
	require.EqualError(t, err, "chainlinkfeeds: description: invalid request: request is nil")
	_, err = a.Version(ctx, nil, &VersionRequest{})
	require.EqualError(t, err, "chainlinkfeeds: version: invalid request: contract is nil")
	_, err = a.GetRoundData(ctx, nil, NewGetRoundDataRequest(feed, nil))
	require.EqualError(t, err, "chainlinkfeeds: getRoundData: invalid request: roundId is nil")
	_, err = a.GetRoundData(ctx, nil, NewGetRoundDataRequest(feed, new(big.Int).Lsh(big.NewInt(1), 80)))
	require.EqualError(t, err, "chainlinkfeeds: getRoundData: invalid request: roundId exceeds uint80")
	_, err = a.LatestRoundData(ctx, nil, &LatestRoundDataRequest{})
	require.EqualError(t, err, "chainlinkfeeds: latestRoundData: invalid request: contract is nil")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
}

func TestAggregatorCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "aggregator")
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	word, err := decimalsFn.EncodeReturn(uint8(8))
	require.NoError(t, err)

	cli := mocks.NewClient(t)
	cli.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "aggregator", got.Value(marker{}))
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		return nil
	}).Once()

	_, err = NewAggregator(cli).Decimals(ctx, nil, NewDecimalsRequest(feed))
	require.NoError(t, err)
}
