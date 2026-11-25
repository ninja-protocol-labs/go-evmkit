package feeds

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestAggregatorRequestsValidate(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
	maxRoundID := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 80), big.NewInt(1))

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"decimals nil", (*DecimalsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"decimals empty", (&DecimalsRequest{}).Validate(), "invalid request: contract is nil"},
		{"decimals valid", (&DecimalsRequest{Contract: feed}).Validate(), ""},
		{"description nil", (*DescriptionRequest)(nil).Validate(), "invalid request: request is nil"},
		{"description empty", (&DescriptionRequest{}).Validate(), "invalid request: contract is nil"},
		{"description valid", (&DescriptionRequest{Contract: feed}).Validate(), ""},
		{"version nil", (*VersionRequest)(nil).Validate(), "invalid request: request is nil"},
		{"version empty", (&VersionRequest{}).Validate(), "invalid request: contract is nil"},
		{"version valid", (&VersionRequest{Contract: feed}).Validate(), ""},
		{"getRoundData nil", (*GetRoundDataRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getRoundData empty", (&GetRoundDataRequest{}).Validate(), "invalid request: contract is nil"},
		{"getRoundData nil contract", (&GetRoundDataRequest{RoundID: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"getRoundData nil roundId", (&GetRoundDataRequest{Contract: feed}).Validate(), "invalid request: roundId is nil"},
		{"getRoundData negative roundId", (&GetRoundDataRequest{Contract: feed, RoundID: big.NewInt(-1)}).Validate(), "invalid request: roundId is negative"},
		{"getRoundData roundId over uint80", (&GetRoundDataRequest{Contract: feed, RoundID: new(big.Int).Lsh(big.NewInt(1), 80)}).Validate(), "invalid request: roundId exceeds uint80"},
		{"getRoundData zero roundId", (&GetRoundDataRequest{Contract: feed, RoundID: big.NewInt(0)}).Validate(), ""},
		{"getRoundData max roundId", (&GetRoundDataRequest{Contract: feed, RoundID: maxRoundID}).Validate(), ""},
		{"latestRoundData nil", (*LatestRoundDataRequest)(nil).Validate(), "invalid request: request is nil"},
		{"latestRoundData empty", (&LatestRoundDataRequest{}).Validate(), "invalid request: contract is nil"},
		{"latestRoundData valid", (&LatestRoundDataRequest{Contract: feed}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewAggregatorRequestsSetFields(t *testing.T) {
	feed := types.MustNewAddressFromHex("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")

	require.Equal(t, &DecimalsRequest{Contract: feed}, NewDecimalsRequest(feed))
	require.Equal(t, &DescriptionRequest{Contract: feed}, NewDescriptionRequest(feed))
	require.Equal(t, &VersionRequest{Contract: feed}, NewVersionRequest(feed))
	require.Equal(t, &GetRoundDataRequest{Contract: feed, RoundID: big.NewInt(9)}, NewGetRoundDataRequest(feed, big.NewInt(9)))
	require.Equal(t, &LatestRoundDataRequest{Contract: feed}, NewLatestRoundDataRequest(feed))
}

func TestNewRoundData(t *testing.T) {
	got := NewRoundData(big.NewInt(1), big.NewInt(-2), big.NewInt(3), big.NewInt(4), big.NewInt(5))
	require.Equal(t, &RoundData{
		RoundID:         big.NewInt(1),
		Answer:          big.NewInt(-2),
		StartedAt:       big.NewInt(3),
		UpdatedAt:       big.NewInt(4),
		AnsweredInRound: big.NewInt(5),
	}, got)
}
