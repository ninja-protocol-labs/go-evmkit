package multicall3

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// multicall3Address is Multicall3's deployment address: the same on every chain that
// has it (https://github.com/mds1/multicall#new-deployments).
const multicall3Address = "0xcA11bde05977b3631167028862bE2a173976CA11"

var (
	ethereum = []string{
		"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
		"0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
		"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		"0xae7ab96520de3a18e5e111b5eaab095312d7fe84", // stETH
		"0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE", // SHIB
		"0x514910771af9ca656af840dff83e8264ecf986ca", // LINK
		"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984", // UNI
		"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599", // WBTC
		"0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9", // AAVE
		"0x6982508145454ce325ddbe47a25d4ec3d2311933", // PEPE
	}
	base = []string{
		"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC
		"0x4200000000000000000000000000000000000006", // WETH
		"0xcbB7C0000aB88B473b1f5aFd9ef808440eed33Bf", // cbBTC
		"0x940181a94A35A4569E4529A3CDfB74e38FD98631", // AERO
		"0x532f27101965dd16442e59d40670faf5ebb142e4", // BRETT
		"0x0b3e328455c4059eeb9e3f84b5543f74e24e7e1b", // VIRTUAL
		"0x4ed4e862860bed51a9570b96d89af5e1b0efefed", // DEGEN
		"0xAC1Bd2486aAf3B5C0fc3Fd868558b082a531B2B4", // TOSHI
		"0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA", // USDbC
		"0x50c5725949A6F0c72E6C4a641F24049A917DB0Cb", // DAI
	}
	robinhood = []string{
		"0x0Bd7D308f8E1639FAb988df18A8011f41EAcAD73", // WETH
		"0x492641F648a4986844848E0beFE66D14817bCE34", // LINK
		"0xcec185eb182c47d1ba1efc84e6959e18cd620be4", // cbBTC
		"0x5d3a1ff2b6bab83b63cd9ad0787074081a52ef34", // USDe
		"0x39dBED3a2bd333467115dE45665cC57F813C4571", // PONS
		"0xd0601CE157Db5bdC3162BbaC2a2C8aF5320D9EEC", // NVDA
		"0xaF3D76f1834A1d425780943C99Ea8A608f8a93f9", // AAPL
		"0x12f190a9F9d7D37a250758b26824B97CE941bF54", // AMZN
		"0x86923f96303D656E4aa86D9d42D1e57ad2023fdC", // AMD
		"0x2F62fC9fAbb470C690f141c28340eD832bB27020", // BND
	}
	arbitrum = []string{
		"0x912CE59144191C1204E64559FE8253a0e49E6548", // ARB
		"0xaf88d065e77c8cC2239327C5EDb3A432268e5831", // USDC
		"0x82aF49447D8a07e3bd95BD0d56f35241523fBab1", // WETH
		"0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9", // USDT
		"0x2f2a2543B76A4166549F7aaB2e75Bef0aefC5B0f", // WBTC
		"0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8", // USDC.e
		"0x5979D7b546E38E414F7E9822514be443A4800529", // wstETH
		"0xf97f4df75117a78c1A5a0DBb814Af92458539FB4", // LINK
		"0xFa7F8980b0f1E64A2062791cc3b0871572f1F7f0", // UNI
		"0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1", // DAI
	}
	avalanche = []string{
		"0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7", // WAVAX
		"0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7", // USDt
		"0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E", // USDC
		"0x6e84a6216eA6dACC71eE8E6b0a5B7322EEbC0fDd", // JOE
		"0x50b7545627a5162f82a992c33b87adc75187b218", // WBTC.e
		"0x49D5c2BdFfac6CE2BFdB6640F4F80f226bc10bAB", // WETH.e
		"0x2b2C81e08f1Af8835a78Bb2A90AE924ACE0eA4bE", // sAVAX
		"0x5947bb275c521040051d82396192181b413227a3", // LINK.e
		"0xA7D7079b0FEaD91F3e65f86E8915Cb59c1a4C664", // USDC.e
		"0x420FcA0121DC28039145009570975747295f2329", // COQ
	}
)

// network is one chain to query: its default public RPC endpoint and the
// tokens to look up on it.
type network struct {
	name   string
	rpcURL string
	tokens []string
}

var networks = []network{
	{name: "ethereum", rpcURL: "https://ethereum-rpc.publicnode.com", tokens: ethereum},
	{name: "base", rpcURL: "https://mainnet.base.org", tokens: base},
	{name: "arbitrum", rpcURL: "https://arb1.arbitrum.io/rpc", tokens: arbitrum},
	{name: "avalanche", rpcURL: "https://api.avax.network/ext/bc/C/rpc", tokens: avalanche},
	{name: "robinhood", rpcURL: "https://rpc.mainnet.chain.robinhood.com", tokens: robinhood},
}

type Multicall3Suite struct {
	suite.Suite

	Enable    bool
	HolderHex string

	ctx    context.Context
	cancel context.CancelFunc

	holder *types.Address
}

func TestMulticall3LiveSuite(t *testing.T) {
	suite.Run(t, &Multicall3Suite{
		Enable:    false,
		HolderHex: "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	})
}

func (s *Multicall3Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	var err error
	s.holder, err = types.NewAddressFromHex(s.HolderHex)
	s.Require().NoError(err)
}

func (s *Multicall3Suite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Multicall3Suite) decodeOrUnknown(fn *abi.Function, r Result) string {
	if !r.Success || len(r.ReturnData) == 0 {
		return "?"
	}
	vals, err := fn.DecodeReturn(r.ReturnData)
	s.Require().NoError(err)
	return fmt.Sprint(vals[0])
}

func decodeValueOrUnknown[T any](t require.TestingT, decode func([]byte) (T, error), r Result) string {
	if !r.Success || len(r.ReturnData) == 0 {
		return "?"
	}
	v, err := decode(r.ReturnData)
	require.NoError(t, err)
	return fmt.Sprint(v)
}

// call sends calls to the Multicall3 contract at multicall via Aggregate3
// and returns the decoded results.
func (s *Multicall3Suite) call(m IMulticall3, multicall *types.Address, calls []Call3) []Result {
	results, err := m.Aggregate3(s.ctx, multicall, calls, rpc.BlockTagLatest)
	s.Require().NoError(err)
	return results
}

func (s *Multicall3Suite) TestAcrossNetworks() {
	multicall, err := types.NewAddressFromHex(multicall3Address)
	s.Require().NoError(err)

	balanceOf, err := abi.ParseFunction("balanceOf(address) returns (uint256)", nil)
	s.Require().NoError(err)
	symbolFn, err := abi.ParseFunction("symbol() returns (string)", nil)
	s.Require().NoError(err)
	totalSupplyFn, err := abi.ParseFunction("totalSupply() returns (uint256)", nil)
	s.Require().NoError(err)
	decimalsFn, err := abi.ParseFunction("decimals() returns (uint8)", nil)
	s.Require().NoError(err)

	for _, n := range networks {
		s.Run(n.name, func() {
			client := rpc.NewClient(n.rpcURL, rpc.WithTimeout(15*time.Second))
			m := New(client)

			tokens := make([]*types.Address, len(n.tokens))
			var calls []Call3
			for i, tk := range n.tokens {
				token, err := types.NewAddressFromHex(tk)
				s.Require().NoError(err)
				tokens[i] = token

				b, err := balanceOf.EncodeCall(s.holder)
				s.Require().NoError(err)
				calls = append(calls, NewCall3(token, true, b))

				b, err = symbolFn.EncodeCall()
				s.Require().NoError(err)
				calls = append(calls, NewCall3(token, false, b))

				b, err = totalSupplyFn.EncodeCall()
				s.Require().NoError(err)
				calls = append(calls, NewCall3(token, true, b))

				b, err = decimalsFn.EncodeCall()
				s.Require().NoError(err)
				calls = append(calls, NewCall3(token, true, b))
			}

			results := s.call(m, multicall, calls)
			s.Require().Len(results, len(calls))

			for i, token := range tokens {
				base := i * 4
				balance := s.decodeOrUnknown(balanceOf, results[base])
				symbol := s.decodeOrUnknown(symbolFn, results[base+1])
				totalSupply := s.decodeOrUnknown(totalSupplyFn, results[base+2])
				decimals := s.decodeOrUnknown(decimalsFn, results[base+3])

				s.T().Logf("[%s] %s(%s): decimals(%s), totalSupply(%s), balanceOf(%s)",
					n.name, token, symbol, decimals, totalSupply, balance)
			}
		})
	}
}

func (s *Multicall3Suite) TestChainInfoAcrossNetworks() {
	multicall, err := types.NewAddressFromHex(multicall3Address)
	s.Require().NoError(err)

	ethBalanceData, err := EncodeGetEthBalance(s.holder)
	s.Require().NoError(err)

	calls := []Call3{
		NewCall3(multicall, true, EncodeGetChainID()),
		NewCall3(multicall, true, EncodeGetBlockNumber()),
		NewCall3(multicall, true, EncodeGetCurrentBlockCoinbase()),
		NewCall3(multicall, true, EncodeGetCurrentBlockDifficulty()),
		NewCall3(multicall, true, EncodeGetCurrentBlockGasLimit()),
		NewCall3(multicall, true, EncodeGetCurrentBlockTimestamp()),
		NewCall3(multicall, true, EncodeGetBasefee()), // can revert if BASEFEE isn't implemented
		NewCall3(multicall, true, EncodeGetLastBlockHash()),
		NewCall3(multicall, true, ethBalanceData),
	}

	for _, n := range networks {
		s.Run(n.name, func() {
			client := rpc.NewClient(n.rpcURL, rpc.WithTimeout(15*time.Second))
			m := New(client)

			results := s.call(m, multicall, calls)
			s.Require().Len(results, len(calls))

			chainID := decodeValueOrUnknown(s.T(), DecodeGetChainID, results[0])
			blockNumber := decodeValueOrUnknown(s.T(), DecodeGetBlockNumber, results[1])
			coinbase := decodeValueOrUnknown(s.T(), DecodeGetCurrentBlockCoinbase, results[2])
			difficulty := decodeValueOrUnknown(s.T(), DecodeGetCurrentBlockDifficulty, results[3])
			gasLimit := decodeValueOrUnknown(s.T(), DecodeGetCurrentBlockGasLimit, results[4])
			timestamp := decodeValueOrUnknown(s.T(), DecodeGetCurrentBlockTimestamp, results[5])
			basefee := decodeValueOrUnknown(s.T(), DecodeGetBasefee, results[6])
			lastBlockHash := decodeValueOrUnknown(s.T(), DecodeGetLastBlockHash, results[7])
			ethBalance := decodeValueOrUnknown(s.T(), DecodeGetEthBalance, results[8])

			s.T().Logf("[%s] chainId(%s), blockNumber(%s), coinbase(%s), difficulty(%s), gasLimit(%s), timestamp(%s), basefee(%s), lastBlockHash(%s), ethBalance(%s)",
				n.name, chainID, blockNumber, coinbase, difficulty, gasLimit, timestamp, basefee, lastBlockHash, ethBalance)
		})
	}
}
