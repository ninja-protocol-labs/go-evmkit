package feeds

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

type AggregatorSuite struct {
	suite.Suite

	RPCURL     string
	ETHUSDHex  string
	BTCUSDHex  string
	USDCUSDHex string
	LINKUSDHex string
	LINKETHHex string

	client     rpc.Client
	aggregator *Aggregator
	ethUSD     *types.Address
	btcUSD     *types.Address
	usdcUSD    *types.Address
	linkUSD    *types.Address
	linkETH    *types.Address
}

func TestAggregatorSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Chainlink.Feeds.ETHUSD == "" {
		t.Skip("testenv: no Chainlink ETH/USD feed configured for this network")
	}

	suite.Run(t, &AggregatorSuite{
		RPCURL:     c.RPCURL,
		ETHUSDHex:  c.Addresses.Protocol.Chainlink.Feeds.ETHUSD,
		BTCUSDHex:  c.Addresses.Protocol.Chainlink.Feeds.BTCUSD,
		USDCUSDHex: c.Addresses.Protocol.Chainlink.Feeds.USDCUSD,
		LINKUSDHex: c.Addresses.Protocol.Chainlink.Feeds.LINKUSD,
		LINKETHHex: c.Addresses.Protocol.Chainlink.Feeds.LINKETH,
	})
}

func (s *AggregatorSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.aggregator = NewAggregator(s.client)

	var err error
	s.ethUSD, err = types.NewAddressFromHex(s.ETHUSDHex)
	s.Require().NoError(err)
	if s.BTCUSDHex != "" {
		s.btcUSD, err = types.NewAddressFromHex(s.BTCUSDHex)
		s.Require().NoError(err)
	}
	if s.USDCUSDHex != "" {
		s.usdcUSD, err = types.NewAddressFromHex(s.USDCUSDHex)
		s.Require().NoError(err)
	}
	if s.LINKUSDHex != "" {
		s.linkUSD, err = types.NewAddressFromHex(s.LINKUSDHex)
		s.Require().NoError(err)
	}
	if s.LINKETHHex != "" {
		s.linkETH, err = types.NewAddressFromHex(s.LINKETHHex)
		s.Require().NoError(err)
	}
}

func (s *AggregatorSuite) Test01ETHUSD() {
	description, err := s.aggregator.Description(s.T().Context(), nil, NewDescriptionRequest(s.ethUSD))
	s.Require().NoError(err)
	s.Require().Equal("ETH / USD", description)
	decimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.ethUSD))
	s.Require().NoError(err)
	version, err := s.aggregator.Version(s.T().Context(), nil, NewVersionRequest(s.ethUSD))
	s.Require().NoError(err)
	s.Require().Positive(version.Sign())
	round, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	s.Require().Positive(round.Answer.Sign())
	s.Require().Positive(round.UpdatedAt.Sign())
	s.Require().LessOrEqual(round.StartedAt.Cmp(round.UpdatedAt), 0)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	price := new(big.Float).Quo(new(big.Float).SetInt(round.Answer), new(big.Float).SetInt(scale))

	s.T().Logf("AggregatorV3 %s Test01ETHUSD Result", s.ethUSD)
	s.T().Logf("\tDescription: %s", description)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tVersion: %s", version)
	phase := new(big.Int).Rsh(round.RoundID, 64)
	aggregatorRound := new(big.Int).And(round.RoundID, new(big.Int).SetUint64(^uint64(0)))
	s.T().Logf("\tRoundID: %s (phase=%s aggregatorRound=%s)", round.RoundID, phase, aggregatorRound)
	s.T().Logf("\tAnswer: %s (%s)", round.Answer, price.Text('f', int(decimals)))
	s.T().Logf("\tStartedAt: %s", time.Unix(round.StartedAt.Int64(), 0).UTC())
	s.T().Logf("\tUpdatedAt: %s", time.Unix(round.UpdatedAt.Int64(), 0).UTC())
}

func (s *AggregatorSuite) Test02BTCUSD() {
	if s.btcUSD == nil {
		s.T().Skip("testenv: no Chainlink BTC/USD feed configured for this network")
	}

	description, err := s.aggregator.Description(s.T().Context(), nil, NewDescriptionRequest(s.btcUSD))
	s.Require().NoError(err)
	s.Require().Equal("BTC / USD", description)
	decimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.btcUSD))
	s.Require().NoError(err)
	round, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.btcUSD))
	s.Require().NoError(err)
	s.Require().Positive(round.Answer.Sign())
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	price := new(big.Float).Quo(new(big.Float).SetInt(round.Answer), new(big.Float).SetInt(scale))

	s.T().Logf("AggregatorV3 %s Test02BTCUSD Result", s.btcUSD)
	s.T().Logf("\tDescription: %s", description)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tRoundID: %s", round.RoundID)
	s.T().Logf("\tAnswer: %s (%s)", round.Answer, price.Text('f', int(decimals)))
	s.T().Logf("\tUpdatedAt: %s", time.Unix(round.UpdatedAt.Int64(), 0).UTC())
}

func (s *AggregatorSuite) Test03USDCUSDIsNearOne() {
	if s.usdcUSD == nil {
		s.T().Skip("testenv: no Chainlink USDC/USD feed configured for this network")
	}

	description, err := s.aggregator.Description(s.T().Context(), nil, NewDescriptionRequest(s.usdcUSD))
	s.Require().NoError(err)
	s.Require().Equal("USDC / USD", description)
	decimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.usdcUSD))
	s.Require().NoError(err)
	round, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.usdcUSD))
	s.Require().NoError(err)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	lower := new(big.Int).Div(new(big.Int).Mul(scale, big.NewInt(9)), big.NewInt(10))
	upper := new(big.Int).Div(new(big.Int).Mul(scale, big.NewInt(11)), big.NewInt(10))
	s.Require().Positive(round.Answer.Cmp(lower))
	s.Require().Negative(round.Answer.Cmp(upper))
	price := new(big.Float).Quo(new(big.Float).SetInt(round.Answer), new(big.Float).SetInt(scale))

	s.T().Logf("AggregatorV3 %s Test03USDCUSDIsNearOne Result", s.usdcUSD)
	s.T().Logf("\tDescription: %s", description)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tAnswer: %s (%s)", round.Answer, price.Text('f', int(decimals)))
	s.T().Logf("\tBounds: (%s, %s)", lower, upper)
	s.T().Logf("\tUpdatedAt: %s", time.Unix(round.UpdatedAt.Int64(), 0).UTC())
}

func (s *AggregatorSuite) Test04LINKUSD() {
	if s.linkUSD == nil {
		s.T().Skip("testenv: no Chainlink LINK/USD feed configured for this network")
	}

	description, err := s.aggregator.Description(s.T().Context(), nil, NewDescriptionRequest(s.linkUSD))
	s.Require().NoError(err)
	s.Require().Equal("LINK / USD", description)
	decimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.linkUSD))
	s.Require().NoError(err)
	round, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.linkUSD))
	s.Require().NoError(err)
	s.Require().Positive(round.Answer.Sign())
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	price := new(big.Float).Quo(new(big.Float).SetInt(round.Answer), new(big.Float).SetInt(scale))

	s.T().Logf("AggregatorV3 %s Test04LINKUSD Result", s.linkUSD)
	s.T().Logf("\tDescription: %s", description)
	s.T().Logf("\tDecimals: %d", decimals)
	s.T().Logf("\tAnswer: %s (%s)", round.Answer, price.Text('f', int(decimals)))
	s.T().Logf("\tUpdatedAt: %s", time.Unix(round.UpdatedAt.Int64(), 0).UTC())
}

func (s *AggregatorSuite) Test05LINKETHMatchesUSDCrossRate() {
	if s.linkETH == nil || s.linkUSD == nil {
		s.T().Skip("testenv: no Chainlink LINK/ETH and LINK/USD feeds configured for this network")
	}

	description, err := s.aggregator.Description(s.T().Context(), nil, NewDescriptionRequest(s.linkETH))
	s.Require().NoError(err)
	s.Require().Equal("LINK / ETH", description)
	linkETHDecimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.linkETH))
	s.Require().NoError(err)
	linkUSDDecimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.linkUSD))
	s.Require().NoError(err)
	ethUSDDecimals, err := s.aggregator.Decimals(s.T().Context(), nil, NewDecimalsRequest(s.ethUSD))
	s.Require().NoError(err)
	linkETH, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.linkETH))
	s.Require().NoError(err)
	linkUSD, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.linkUSD))
	s.Require().NoError(err)
	ethUSD, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)

	linkETHScale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(linkETHDecimals)), nil)
	linkUSDScale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(linkUSDDecimals)), nil)
	ethUSDScale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(ethUSDDecimals)), nil)
	direct := new(big.Float).Quo(new(big.Float).SetInt(linkETH.Answer), new(big.Float).SetInt(linkETHScale))
	linkPrice := new(big.Float).Quo(new(big.Float).SetInt(linkUSD.Answer), new(big.Float).SetInt(linkUSDScale))
	ethPrice := new(big.Float).Quo(new(big.Float).SetInt(ethUSD.Answer), new(big.Float).SetInt(ethUSDScale))
	cross := new(big.Float).Quo(linkPrice, ethPrice)
	ratio, _ := new(big.Float).Quo(direct, cross).Float64()
	s.Require().InDelta(1.0, ratio, 0.05)

	s.T().Logf("AggregatorV3 %s Test05LINKETHMatchesUSDCrossRate Result", s.linkETH)
	s.T().Logf(
		"\tLINK/ETH: answer=%s decimals=%d (%s)",
		linkETH.Answer,
		linkETHDecimals,
		direct.Text('f', int(linkETHDecimals)),
	)
	s.T().Logf(
		"\tLINK/USD ÷ ETH/USD: %s ÷ %s = %s",
		linkPrice.Text('f', int(linkUSDDecimals)),
		ethPrice.Text('f', int(ethUSDDecimals)),
		cross.Text('f', 18),
	)
	s.T().Logf("\tDirect/Cross: %.6f", ratio)
}

func (s *AggregatorSuite) Test06GetRoundDataMatchesLatest() {
	latest, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	round, err := s.aggregator.GetRoundData(s.T().Context(), nil, NewGetRoundDataRequest(s.ethUSD, latest.RoundID))
	s.Require().NoError(err)
	s.Require().Zero(latest.RoundID.Cmp(round.RoundID))
	s.Require().Zero(latest.Answer.Cmp(round.Answer))
	s.Require().Zero(latest.StartedAt.Cmp(round.StartedAt))
	s.Require().Zero(latest.UpdatedAt.Cmp(round.UpdatedAt))

	s.T().Logf("AggregatorV3 %s Test06GetRoundDataMatchesLatest Result", s.ethUSD)
	s.T().Logf("\tLatestRoundData: roundId=%s answer=%s updatedAt=%s", latest.RoundID, latest.Answer, latest.UpdatedAt)
	s.T().Logf("\tGetRoundData: roundId=%s answer=%s updatedAt=%s", round.RoundID, round.Answer, round.UpdatedAt)
}

func (s *AggregatorSuite) Test07GetRoundDataPreviousRound() {
	latest, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	s.Require().Positive(new(big.Int).And(latest.RoundID, new(big.Int).SetUint64(^uint64(0))).Cmp(big.NewInt(1)))
	previousID := new(big.Int).Sub(latest.RoundID, big.NewInt(1))
	previous, err := s.aggregator.GetRoundData(s.T().Context(), nil, NewGetRoundDataRequest(s.ethUSD, previousID))
	s.Require().NoError(err)
	s.Require().Zero(previousID.Cmp(previous.RoundID))
	s.Require().Positive(previous.Answer.Sign())
	s.Require().LessOrEqual(previous.UpdatedAt.Cmp(latest.UpdatedAt), 0)

	s.T().Logf("AggregatorV3 %s Test07GetRoundDataPreviousRound Result", s.ethUSD)
	latestAt := time.Unix(latest.UpdatedAt.Int64(), 0).UTC()
	previousAt := time.Unix(previous.UpdatedAt.Int64(), 0).UTC()
	s.T().Logf("\tLatest: roundId=%s answer=%s updatedAt=%s", latest.RoundID, latest.Answer, latestAt)
	s.T().Logf("\tPrevious: roundId=%s answer=%s updatedAt=%s", previous.RoundID, previous.Answer, previousAt)
	s.T().Logf("\tGap: %s", time.Duration(new(big.Int).Sub(latest.UpdatedAt, previous.UpdatedAt).Int64())*time.Second)
}

func (s *AggregatorSuite) Test08GetRoundDataFutureRoundHasNoData() {
	latest, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	futureID := new(big.Int).Add(latest.RoundID, big.NewInt(1_000_000))
	round, err := s.aggregator.GetRoundData(s.T().Context(), nil, NewGetRoundDataRequest(s.ethUSD, futureID))
	s.Require().Nil(round)
	s.Require().ErrorIs(err, ErrNoDataPresent)

	s.T().Logf("AggregatorV3 %s Test08GetRoundDataFutureRoundHasNoData Result", s.ethUSD)
	s.T().Logf("\tLatest RoundID: %s", latest.RoundID)
	s.T().Logf("\tFuture RoundID: %s", futureID)
	s.T().Logf("\tError: %v", err)
}

func (s *AggregatorSuite) Test09GetRoundDataBeyondUint32RoundHasNoData() {
	latest, err := s.aggregator.LatestRoundData(s.T().Context(), nil, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	phase := new(big.Int).Rsh(latest.RoundID, 64)
	beyondID := new(big.Int).Or(new(big.Int).Lsh(phase, 64), new(big.Int).Lsh(big.NewInt(1), 33))
	round, err := s.aggregator.GetRoundData(s.T().Context(), nil, NewGetRoundDataRequest(s.ethUSD, beyondID))
	s.Require().Nil(round)
	s.Require().ErrorIs(err, ErrNoDataPresent)

	s.T().Logf("AggregatorV3 %s Test09GetRoundDataBeyondUint32RoundHasNoData Result", s.ethUSD)
	s.T().Logf("\tRoundID: %s (phase=%s aggregatorRound=2^33)", beyondID, phase)
	s.T().Logf("\tError: %v", err)
}

func (s *AggregatorSuite) Test10GetRoundDataUnknownPhaseReverts() {
	unknownID := new(big.Int).Or(new(big.Int).Lsh(big.NewInt(0xffff), 64), big.NewInt(1))
	round, err := s.aggregator.GetRoundData(s.T().Context(), nil, NewGetRoundDataRequest(s.ethUSD, unknownID))
	s.Require().Nil(round)
	s.Require().ErrorIs(err, ErrReverted)

	s.T().Logf("AggregatorV3 %s Test10GetRoundDataUnknownPhaseReverts Result", s.ethUSD)
	s.T().Logf("\tRoundID: %s (phase=0xffff aggregatorRound=1)", unknownID)
	s.T().Logf("\tError: %v", err)
}

func (s *AggregatorSuite) Test11LatestRoundDataAtPinnedBlock() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHBlockNumber("1", &blockHex)))
	opts := bind.NewCallOptions().WithBlock(blockHex)
	first, err := s.aggregator.LatestRoundData(s.T().Context(), opts, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	second, err := s.aggregator.LatestRoundData(s.T().Context(), opts, NewLatestRoundDataRequest(s.ethUSD))
	s.Require().NoError(err)
	s.Require().Zero(first.RoundID.Cmp(second.RoundID))
	s.Require().Zero(first.Answer.Cmp(second.Answer))
	s.Require().Zero(first.UpdatedAt.Cmp(second.UpdatedAt))

	s.T().Logf("AggregatorV3 %s Test11LatestRoundDataAtPinnedBlock Result", s.ethUSD)
	s.T().Logf("\tBlock: %s", blockHex)
	s.T().Logf("\tRoundID: %s", first.RoundID)
	s.T().Logf("\tAnswer: %s", first.Answer)
}
