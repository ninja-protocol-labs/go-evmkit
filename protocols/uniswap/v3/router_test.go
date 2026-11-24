package v3

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/weth"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type RouterSwapSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	RouterHex           string
	QuoterHex           string
	MulticallHex        string
	TokenInHex          string
	TokenOutHex         string
	HopsHex             []string
	FeesHex             []string
	NativeIn            bool
	NativeOut           bool
	AmountIn            string
	RoundTrip           bool
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	PriceBuffer         uint64
	LimitBuffer         uint64

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	router    *types.Address
	quoter    *types.Address
	weth9Addr *types.Address
	tokenIn   *types.Address
	tokenOut  *types.Address
	hops      []*types.Address
	fees      []*big.Int
	multicall *types.Address
	amount    *big.Int

	v3Router *Router
	v3Quoter *Quoter
	erc20    *erc20.ERC20

	totalFee *big.Int
}

func TestRouterSwapSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V3.Router == "" || c.Addresses.Protocol.Uniswap.V3.Quoter == "" {
		t.Skip("testenv: no Uniswap V3 router/quoter configured for this network")
	}

	s := &RouterSwapSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		RouterHex:           c.Addresses.Protocol.Uniswap.V3.Router,
		QuoterHex:           c.Addresses.Protocol.Uniswap.V3.Quoter,
		MulticallHex:        c.Addresses.Standard.Multicall,
		TokenInHex:          "",
		TokenOutHex:         "",
		HopsHex:             []string{},
		FeesHex:             []string{strconv.Itoa(FeeLow)},
		NativeIn:            false,
		NativeOut:           false,
		AmountIn:            "",
		RoundTrip:           false,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		PriceBuffer:         150,
		LimitBuffer:         150,
	}
	if s.AmountIn == "" || (!s.NativeIn && s.TokenInHex == "") || (!s.NativeOut && s.TokenOutHex == "") {
		t.Skip("RouterSwapSuite: AmountIn/TokenInHex/TokenOutHex not set")
	}

	suite.Run(t, s)
}

func (s *RouterSwapSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.v3Router = NewRouter(s.client)
	s.v3Quoter = NewQuoter(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)

	s.quoter, err = types.NewAddressFromHex(s.QuoterHex)
	s.Require().NoError(err)

	s.Require().False(s.NativeIn && s.NativeOut, "NativeIn and NativeOut can't both be set -- that's not a swap")

	s.weth9Addr, err = s.v3Router.WETH9(s.ctx, nil, NewWETH9Request(s.router))
	s.Require().NoError(err)

	if s.NativeIn {
		s.tokenIn = s.weth9Addr
	} else {
		s.tokenIn, err = types.NewAddressFromHex(s.TokenInHex)
		s.Require().NoError(err)
	}

	if s.NativeOut {
		s.tokenOut = s.weth9Addr
	} else {
		s.tokenOut, err = types.NewAddressFromHex(s.TokenOutHex)
		s.Require().NoError(err)
	}

	s.hops = make([]*types.Address, 0, len(s.HopsHex))
	for _, h := range s.HopsHex {
		hop, err := types.NewAddressFromHex(h)
		s.Require().NoError(err)
		s.hops = append(s.hops, hop)
	}

	s.Require().Len(s.FeesHex, len(s.hops)+1, "need exactly one fee tier per pool in the path (len(hops)+1)")
	s.fees = make([]*big.Int, 0, len(s.FeesHex))
	for _, f := range s.FeesHex {
		fee, ok := new(big.Int).SetString(f, 10)
		s.Require().True(ok)
		s.fees = append(s.fees, fee)
	}

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	amount, ok := new(big.Int).SetString(s.AmountIn, 10)
	s.Require().True(ok)
	s.amount = amount

	path := append([]*types.Address{s.tokenIn}, s.hops...)
	path = append(path, s.tokenOut)
	previewPath, err := EncodePath(path, s.fees)
	s.Require().NoError(err)
	if s.NativeIn {
		previewParams := NewExactInputParams(previewPath, s.from, s.amount, big.NewInt(0))
		_, err = s.v3Router.CallExactInput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputRequest(s.router, previewParams, s.amount))
		s.Require().NoError(err, "no route from tokenIn to tokenOut via this router, at this fee tier")
	}
}

func (s *RouterSwapSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RouterSwapSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterSwapSuite) TestSwap() {
	tokenInMeta, err := s.erc20.MetadataWithBalance(s.ctx, nil, erc20.NewMetadataWithBalanceRequest(s.multicall, s.tokenIn, s.from, false))
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenIn %s", s.tokenIn)
	s.T().Logf("\t %s (%s), decimals: %d", tokenInMeta.Name, tokenInMeta.Symbol, tokenInMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenIn, tokenInMeta.Balance.String())

	tokenOutMeta, err := s.erc20.MetadataWithBalance(s.ctx, nil, erc20.NewMetadataWithBalanceRequest(s.multicall, s.tokenOut, s.from, false))
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenOut %s", s.tokenOut)
	s.T().Logf("\t %s (%s), decimals: %d", tokenOutMeta.Name, tokenOutMeta.Symbol, tokenOutMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenOut, tokenOutMeta.Balance.String())

	forwardPath := append([]*types.Address{s.tokenIn}, s.hops...)
	forwardPath = append(forwardPath, s.tokenOut)
	forwardFees := s.fees

	reversePath := make([]*types.Address, len(forwardPath))
	reverseFees := make([]*big.Int, len(forwardFees))
	for i, addr := range forwardPath {
		reversePath[len(forwardPath)-1-i] = addr
	}
	for i, fee := range forwardFees {
		reverseFees[len(forwardFees)-1-i] = fee
	}

	received := s.swap(forwardPath, forwardFees, s.amount, s.NativeIn, s.NativeOut)
	if !s.RoundTrip {
		return
	}

	final := s.swap(reversePath, reverseFees, received, s.NativeOut, s.NativeIn)
	s.T().Logf("TestSwap: round trip %s -> %s -> %s: sent %s, received back %s", s.tokenIn, s.tokenOut, s.tokenIn, s.amount, final)

	if s.NativeOut {
		return
	}
	tokenOutBalanceAfter, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenOut, s.from))
	s.Require().NoError(err)
	s.Require().Zero(tokenOutBalanceAfter.Cmp(tokenOutMeta.Balance), "wallet's TokenOut balance changed (%s -> %s) after swapping it all back", tokenOutMeta.Balance, tokenOutBalanceAfter)
}

func (s *RouterSwapSuite) swap(path []*types.Address, fees []*big.Int, amountIn *big.Int, nativeIn, nativeOut bool) *big.Int {
	tokenIn, tokenOut := path[0], path[len(path)-1]

	if nativeIn {
		var hex string
		s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &hex)))
		balance, ok := new(big.Int).SetString(strings.TrimPrefix(hex, "0x"), 16)
		s.Require().True(ok)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s native, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet native balance: %s", tokenIn, tokenOut, balance)
	} else {
		balance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(tokenIn, s.from))
		s.Require().NoError(err)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s of tokenIn, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet balance of tokenIn: %s", tokenIn, tokenOut, balance)

		approveBuilder := tx.NewDynamicFeeBuilder()
		approveBuilder.PriceBuffer(s.PriceBuffer)
		approveBuilder.LimitBuffer(s.LimitBuffer)
		approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, approveBuilder), erc20.NewApproveRequest(tokenIn, s.router, amountIn))
		s.Require().NoError(err)
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())

		s.T().Logf("swap %s -> %s: approve %s to spend %s, broadcast tx %s", tokenIn, tokenOut, s.router, amountIn, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(tokenIn, s.from, s.router))
			return err == nil && amountIn.Cmp(allowance) == 0
		}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
		s.T().Logf("swap %s -> %s: approve confirmed in block %d", tokenIn, tokenOut, approveReceipt.BlockNumber)
	}

	encodedPath, err := EncodePath(path, fees)
	s.Require().NoError(err)

	recipient := s.from
	if nativeOut {
		recipient = types.MustNewAddressFromHex("0x0000000000000000000000000000000000000002")
	}

	var previewValue *big.Int
	if nativeIn {
		previewValue = amountIn
	}

	var amountOutMin *big.Int
	var swapCalldata []byte
	var lastErr error
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		previewParams := NewExactInputParams(encodedPath, recipient, amountIn, big.NewInt(0))
		amountOutLast, err := s.v3Router.CallExactInput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputRequest(s.router, previewParams, previewValue))
		if err != nil {
			lastErr = err
			return false
		}
		amountOutMin = new(big.Int).Div(new(big.Int).Mul(amountOutLast, big.NewInt(995)), big.NewInt(1000))
		s.T().Logf("swap %s -> %s: attempt %d: CallExactInput=%s amountOutMin=%s", tokenIn, tokenOut, attempt, amountOutLast, amountOutMin)

		calldata, err := EncodeExactInput(encodedPath, recipient, amountIn, amountOutMin)
		if err != nil {
			lastErr = err
			return false
		}
		swapCalldata = calldata
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "swap not buildable: %v", lastErr)

	quoterQuote, err := s.v3Quoter.CallQuoteExactInput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewQuoteExactInputRequest(s.quoter, encodedPath, amountIn))
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: Quoter's own quote taken right before broadcast: amountOut=%s", tokenIn, tokenOut, quoterQuote.AmountOut)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	opts := bind.NewTransactOptions(s.signer, builder)

	var swapTx core.Transaction
	var swapHash *types.Hash
	if nativeOut {
		unwrapCalldata, encErr := EncodeRouterUnwrapWETH9(amountOutMin, s.from)
		s.Require().NoError(encErr)
		deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
		swapTx, swapHash, err = s.v3Router.Multicall(s.ctx, opts, NewRouterMulticallRequest(s.router, [][]byte{swapCalldata, unwrapCalldata}, deadline, previewValue))
	} else {
		swapTx, swapHash, err = s.v3Router.ExactInput(s.ctx, opts, NewExactInputRequest(s.router, NewExactInputParams(encodedPath, recipient, amountIn, amountOutMin), previewValue))
	}
	s.Require().NoError(err)
	s.totalFee.Add(s.totalFee, swapTx.MaxFee())
	s.T().Logf("swap %s -> %s: swap sent after %d attempt(s), amountOutMin=%s, nonce %d", tokenIn, tokenOut, attempt, amountOutMin, swapTx.AddressNonce())

	s.T().Logf("swap %s -> %s: broadcast swap, tx %s", tokenIn, tokenOut, swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("swap %s -> %s: swap confirmed in block %d", tokenIn, tokenOut, swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, len(path)-1)

	transfers, err := erc20.ExtractTransfers(swapReceipt.Logs)
	s.Require().NoError(err)

	lastSwap := swaps[len(swaps)-1]
	var actualAmountOut *big.Int
	if lastSwap.Amount0.Sign() < 0 {
		actualAmountOut = new(big.Int).Neg(lastSwap.Amount0)
	} else {
		actualAmountOut = new(big.Int).Neg(lastSwap.Amount1)
	}

	drift := new(big.Int).Sub(actualAmountOut, quoterQuote.AmountOut)
	driftPct := new(big.Float).Quo(new(big.Float).SetInt(drift), new(big.Float).SetInt(quoterQuote.AmountOut))
	driftPct.Mul(driftPct, big.NewFloat(100))

	s.T().Logf("Router swap %s -> %s Result", tokenIn, tokenOut)
	s.T().Logf("\tAmountIn: %s", amountIn)
	s.T().Logf("\tAmountOut: %s", actualAmountOut)
	s.T().Logf("\tQuoter's pre-swap quote: %s (drift %s, %s%%)", quoterQuote.AmountOut, drift, driftPct.Text('f', 6))
	s.T().Logf("\tTx: %s", swapHash)
	s.T().Logf("\tLogs:")
	for _, transfer := range transfers {
		s.T().Logf("\t\tTransfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	if nativeIn {
		deposits, err := weth.ExtractDeposits(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, d := range deposits {
			s.T().Logf("\t\tDeposit: dst=%s wad=%s", d.Dst, d.Wad)
		}
	}
	if nativeOut {
		withdrawals, err := weth.ExtractWithdrawals(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, w := range withdrawals {
			s.T().Logf("\t\tWithdrawal: src=%s wad=%s", w.Src, w.Wad)
		}
	}
	for _, sw := range swaps {
		s.T().Logf("\t\tSwap: contract=%s sender=%s recipient=%s amount0=%s amount1=%s", sw.Contract, sw.Sender, sw.Recipient, sw.Amount0, sw.Amount1)
	}

	return actualAmountOut
}

type RouterLifeCycleSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	RouterHex           string
	NfpmHex             string
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Fee                 *big.Int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Constructor         *abi.Function
	Salt                *types.Hash
	MintValue           string
	FeeRecipientHex     string

	ctx    context.Context
	cancel context.CancelFunc

	client       rpc.Client
	signer       core.Signer
	deployer     *deploy.NicksDeployer
	from         *types.Address
	router       *types.Address
	nfpm         *types.Address
	factory      *types.Address
	weth9        *types.Address
	multicall    *types.Address
	nicks        *types.Address
	tokenA       *types.Address
	tokenB       *types.Address
	tokenC       *types.Address
	pool         *types.Address
	poolBC       *types.Address
	poolWETH     *types.Address
	wethToken0   *types.Address
	wethToken1   *types.Address
	feeRecipient *types.Address
	chainID      *big.Int
	tokenId      *big.Int

	v3Router  *Router
	v3Nfpm    *NFPM
	v3Factory *Factory
	v3Pool    *Pool
	erc20     *erc20.ERC20

	mu            *sync.Mutex
	totalFee      *big.Int
	mintValue     *big.Int
	wethLiquidity *big.Int
}

func TestRouterLifeCycleSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V3.Router == "" || c.Addresses.Protocol.Uniswap.V3.NFPM == "" {
		t.Skip("testenv: no Uniswap V3 router/NFPM configured for this network")
	}

	suite.Run(t, &RouterLifeCycleSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		RouterHex:           c.Addresses.Protocol.Uniswap.V3.Router,
		NfpmHex:             c.Addresses.Protocol.Uniswap.V3.NFPM,
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Fee:                 big.NewInt(FeeMedium),
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20PermitFullByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.Types{abi.String, abi.String}, nil),
		Salt:                core.Keccak256([]byte("go-evmkit")),
		MintValue:           "1000000000000000000000000",
		FeeRecipientHex:     "0x000000000000000000000000000000000000dEaD",
	})
}

func (s *RouterLifeCycleSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Router = NewRouter(s.client)
	s.v3Nfpm = NewNFPM(s.client)
	s.v3Factory = NewFactory(s.client)
	s.v3Pool = NewPool(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target SwapRouter %s (fee tier %s)", s.router, s.Fee)

	s.nfpm, err = types.NewAddressFromHex(s.NfpmHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target NonfungiblePositionManager %s", s.nfpm)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	s.feeRecipient, err = types.NewAddressFromHex(s.FeeRecipientHex)
	s.Require().NoError(err)

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID
	s.T().Logf("SetupSuite: chainID = %s", s.chainID)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [3]string{"evkit test router token A", "evkit test router token B", "evkit test router token C"}
	symbols := [3]string{"EVKRA", "EVKRB", "EVKRC"}
	tokens := [3]*types.Address{}

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i]}),
			salt,
		)

		predicted, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)

		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], deployTx.AddressNonce(), deployTx.MaxFee())

		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d", symbols[i], receipt.BlockNumber)

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, deployTx.AddressNonce(), hash)

		tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
		calldata, err := tokenMintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		mintBuilder := tx.NewDynamicFeeBuilder()
		mintBuilder.PriceBuffer(150)
		mintBuilder.LimitBuffer(300)
		unsignedMint, err := mintBuilder.Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
		s.Require().NoError(err)
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed mint of %s to %s (nonce %d, gasLimit %d, maxFee %s)", symbols[i], s.from, unsignedMint.AddressNonce(), unsignedMint.Gas(), unsignedMint.MaxFee())

		mintRaw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, mintRaw)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: broadcast mint of %s, tx %s", symbols[i], mintHash)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s of %s to %s, confirmed in block %d (tx %s)", s.mintValue, token, s.from, mintReceipt.BlockNumber, mintHash)
	}

	sorted := tokens[:]
	slices.SortFunc(sorted, func(a, b *types.Address) int { return a.Big().Cmp(b.Big()) })
	s.tokenA, s.tokenB, s.tokenC = sorted[0], sorted[1], sorted[2]
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s token2=%s", s.tokenA, s.tokenB, s.tokenC)

	weth9, err := s.v3Router.WETH9(s.ctx, nil, NewWETH9Request(s.router))
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.weth9 = weth9

	var nativeBalanceHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceHex)))
	nativeBalance, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceHex, "0x"), 16)
	s.Require().True(ok)
	s.wethLiquidity = new(big.Int).Div(nativeBalance, big.NewInt(10000))
	s.Require().True(s.wethLiquidity.Sign() > 0, "native balance %s is too small to spare 1/10000 for WETH liquidity", nativeBalance)
	s.T().Logf("SetupSuite: native balance of %s is %s, wrapping 1/10000 of it (%s) into WETH9 %s", s.from, nativeBalance, s.wethLiquidity, weth9)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, weth9, s.wethLiquidity, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	s.T().Logf("SetupSuite: signed WETH9.deposit() of %s (nonce %d, maxFee %s)", s.wethLiquidity, unsignedWrap.AddressNonce(), unsignedWrap.MaxFee())

	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: broadcast WETH9.deposit(), tx %s", wrapHash)

	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("SetupSuite: wrapped %s into WETH9, confirmed in block %d (tx %s)", s.wethLiquidity, wrapReceipt.BlockNumber, wrapHash)
}

func (s *RouterLifeCycleSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil && s.tokenC != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB, s.tokenC}, false))
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			s.T().Logf("TearDownSuite: leftover balance of %s is %s, burning it", tb.Token, tb.Amount)

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()
			s.T().Logf("TearDownSuite: signed burn of %s (nonce %d, maxFee %s)", tb.Token, unsignedBurn.AddressNonce(), unsignedBurn.MaxFee())

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)
			s.T().Logf("TearDownSuite: broadcast burn of %s, tx %s", tb.Token, burnHash)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s, confirmed in block %d (tx %s)", tb.Amount, tb.Token, s.from, burnReceipt.BlockNumber, burnHash)
		}
	}

	if s.feeRecipient != nil && s.tokenA != nil {
		feeRecipientBalance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.feeRecipient))
		s.Require().NoError(err)

		if feeRecipientBalance.Sign() > 0 {
			s.T().Logf("TearDownSuite: feeRecipient %s holds %s of %s (collected fees), burning it (this test token's burn() is unrestricted by design)", s.feeRecipient, feeRecipientBalance, s.tokenA)

			tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)
			calldata, err := tokenBurnFn.EncodeCall(s.feeRecipient, feeRecipientBalance)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, s.tokenA, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			s.T().Logf("TearDownSuite: burned %s of %s from feeRecipient %s, confirmed in block %d (tx %s)", feeRecipientBalance, s.tokenA, s.feeRecipient, burnReceipt.BlockNumber, burnHash)
		}

		var feeRecipientNativeHex string
		s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientNativeHex)))
		feeRecipientNative, ok := new(big.Int).SetString(strings.TrimPrefix(feeRecipientNativeHex, "0x"), 16)
		s.Require().True(ok)
		if feeRecipientNative.Sign() > 0 {
			s.T().Logf("TearDownSuite: feeRecipient %s holds %s native %s (collected WETH9 fee, now unwrapped) -- not reclaimable, we don't hold its private key", s.feeRecipient, feeRecipientNative, s.NativeSymbol)
		}
	}

	if s.weth9 != nil {
		wethBalance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.weth9, s.from))
		s.Require().NoError(err)

		if wethBalance.Sign() > 0 {
			s.T().Logf("TearDownSuite: leftover WETH9 balance is %s, withdrawing it back to native", wethBalance)

			withdrawFn := abi.NewFunction("withdraw", abi.NewTypes(abi.Uint256), nil)
			calldata, err := withdrawFn.EncodeCall(wethBalance)
			s.Require().NoError(err)

			unsignedWithdraw, err := tx.NewDynamicFeeBuilder().Build(s.from, s.weth9, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedWithdraw.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedWithdraw.MaxFee())
			s.mu.Unlock()
			s.T().Logf("TearDownSuite: signed WETH9.withdraw(%s) (nonce %d, maxFee %s)", wethBalance, unsignedWithdraw.AddressNonce(), unsignedWithdraw.MaxFee())

			raw, err := unsignedWithdraw.EncodeRLP()
			s.Require().NoError(err)
			withdrawHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)
			s.T().Logf("TearDownSuite: broadcast WETH9.withdraw(), tx %s", withdrawHash)

			withdrawReceipt, err := tx.WaitForReceipt(s.ctx, s.client, withdrawHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(withdrawReceipt.Status, "WETH9.withdraw() transaction reverted")
			s.T().Logf("TearDownSuite: withdrew %s of WETH9 back to native, confirmed in block %d (tx %s)", wethBalance, withdrawReceipt.BlockNumber, withdrawHash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RouterLifeCycleSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v3Router.Factory(s.ctx, nil, NewFactoryRequest(s.router))
	s.Require().NoError(err)
	s.Require().False(factory.IsZero())
	s.factory = factory
	s.T().Logf("Test01ReadMethods: factory() = %s", factory)

	weth9, err := s.v3Router.WETH9(s.ctx, nil, NewWETH9Request(s.router))
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.weth9 = weth9
	s.T().Logf("Test01ReadMethods: WETH9() = %s", weth9)

	nfpmFactory, err := s.v3Nfpm.Factory(s.ctx, nil, NewFactoryRequest(s.nfpm))
	s.Require().NoError(err)
	s.Require().True(nfpmFactory.Equal(factory), "Router and NFPM are different periphery contracts, but they must share the same core Factory")
	s.T().Logf("Test01ReadMethods: NFPM's factory() = %s matches Router's", nfpmFactory)

	nfpmWETH9, err := s.v3Nfpm.WETH9(s.ctx, nil, NewWETH9Request(s.nfpm))
	s.Require().NoError(err)
	s.Require().True(nfpmWETH9.Equal(weth9), "Router and NFPM are different periphery contracts, but they must share the same WETH9")
	s.T().Logf("Test01ReadMethods: NFPM's WETH9() = %s matches Router's", nfpmWETH9)

	balances, err := s.erc20.AddressBalances(s.ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB, s.tokenC}, false))
	s.Require().NoError(err)
	s.Require().Len(balances.Tokens, 3)
	for _, tb := range balances.Tokens {
		s.Require().Zero(s.mintValue.Cmp(tb.Amount), "freshly minted token %s must hold exactly mintValue before any swap spends it", tb.Token)
		s.T().Logf("Test01ReadMethods: deployed token %s balance=%s", tb.Token, tb.Amount)
	}
}

func (s *RouterLifeCycleSuite) Test02CreateAndInitializePoolIfNecessary() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: signed tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: confirmed in block %d", receipt.BlockNumber)

	s.pool = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.pool))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, nil, NewGetPoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
		return err == nil && registeredPool.Equal(s.pool)
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	expectedTickSpacing, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, nil, NewFeeAmountTickSpacingRequest(s.factory, s.Fee))
	s.Require().NoError(err)
	s.Require().True(expectedTickSpacing.Sign() > 0, "fee tier %s isn't enabled on this Factory", s.Fee)

	poolToken0, err := s.v3Pool.Token0(s.ctx, nil, NewToken0Request(s.pool))
	s.Require().NoError(err)
	s.Require().True(poolToken0.Equal(s.tokenA))
	poolToken1, err := s.v3Pool.Token1(s.ctx, nil, NewToken1Request(s.pool))
	s.Require().NoError(err)
	s.Require().True(poolToken1.Equal(s.tokenB))
	poolFee, err := s.v3Pool.Fee(s.ctx, nil, NewFeeRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(poolFee))
	poolTickSpacing, err := s.v3Pool.TickSpacing(s.ctx, nil, NewTickSpacingRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(expectedTickSpacing.Cmp(poolTickSpacing), "pool's own tickSpacing must match what the Factory advertises for this fee tier")
	poolFactory, err := s.v3Pool.Factory(s.ctx, nil, NewFactoryRequest(s.pool))
	s.Require().NoError(err)
	s.Require().True(poolFactory.Equal(s.factory))
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool confirms token0=%s token1=%s fee=%s tickSpacing=%s factory=%s", poolToken0, poolToken1, poolFee, poolTickSpacing, poolFactory)

	slot0, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().True(slot0.Unlocked)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool slot0 confirms sqrtPriceX96=%s tick=%s unlocked=%t", slot0.SqrtPriceX96, slot0.Tick, slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test03Mint() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test03Mint: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenA, s.tokenB, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenA, s.tokenB} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(token, s.nfpm, amount))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test03Mint: signed approve of %s (nonce %d, maxFee %s)", token, approveTx.AddressNonce(), approveTx.MaxFee())

		s.T().Logf("Test03Mint: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.nfpm))
			return err == nil && amount.Cmp(allowance) == 0
		}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
		s.T().Logf("Test03Mint: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	balance0Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test03Mint: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	params := NewMintParams(s.tokenA, s.tokenB, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test03Mint: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.Mint(s.ctx, bind.NewTransactOptions(s.signer, builder), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03Mint: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test03Mint: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test03Mint: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.tokenId = increased[0].TokenId
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test03Mint: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, nil, NewNFPMPositionsRequest(s.nfpm, s.tokenId))
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "position not visible after mint")
	s.Require().True(position.Token0.Equal(s.tokenA))
	s.Require().True(position.Token1.Equal(s.tokenB))
	s.Require().Zero(s.Fee.Cmp(position.Fee))
	s.Require().Zero(tickLower.Cmp(position.TickLower))
	s.Require().Zero(tickUpper.Cmp(position.TickUpper))
	s.T().Logf("Test03Mint: positions(%s) confirms token0=%s token1=%s fee=%s tickLower=%s tickUpper=%s liquidity=%s", s.tokenId, position.Token0, position.Token1, position.Fee, position.TickLower, position.TickUpper, position.Liquidity)

	owner, err := s.v3Nfpm.OwnerOf(s.ctx, nil, erc721.NewOwnerOfRequest(s.nfpm, s.tokenId))
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))
	s.T().Logf("Test03Mint: ownerOf(%s) = %s", s.tokenId, owner)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool's active liquidity not visible after mint")
	s.T().Logf("Test03Mint: pool liquidity() = %s, matching the position's own liquidity since it's the pool's sole LP and the current tick sits inside its range", liquidity)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		return err == nil && new(big.Int).Sub(balance0Before, balance0After).Cmp(predictedAmount0) == 0 && new(big.Int).Sub(balance1Before, balance1After).Cmp(predictedAmount1) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after mint")
	s.T().Logf("Test03Mint: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test04ExactInputSingle() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(100))
	s.T().Logf("Test04ExactInputSingle: tokenIn=%s tokenOut=%s amountIn=%s", s.tokenA, s.tokenB, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenA, s.router, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04ExactInputSingle: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, approveTx.AddressNonce(), approveTx.MaxFee())

	s.T().Logf("Test04ExactInputSingle: approve %s to spend %s of %s, broadcast tx %s", s.router, amountIn, s.tokenA, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && amountIn.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test04ExactInputSingle: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	balance0Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	slot0Before, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	liquidityBefore, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: pool before: sqrtPriceX96=%s tick=%s liquidity=%s", slot0Before.SqrtPriceX96, slot0Before.Tick, liquidityBefore)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))

	var predictedAmountOut *big.Int
	s.Require().Eventually(func() bool {
		predictedAmountOut, err = s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, nil))
		ok := err == nil && predictedAmountOut.Sign() > 0
		if !ok {
			allowance, allowanceErr := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
			balance, balanceErr := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
			slot0, slot0Err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
			liquidity, liquidityErr := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
			s.T().Logf("Test04ExactInputSingle: retry diagnostics: callErr=%v amountOut=%v allowance=%v (err=%v) balance=%v (err=%v) slot0=%+v (err=%v) liquidity=%v (err=%v)", err, predictedAmountOut, allowance, allowanceErr, balance, balanceErr, slot0, slot0Err, liquidity, liquidityErr)
		}
		return ok
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "callExactInputSingle not yet succeeding, last error: %v", err)
	s.T().Logf("Test04ExactInputSingle: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactInputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04ExactInputSingle: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test04ExactInputSingle: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test04ExactInputSingle: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().Zero(amountIn.Cmp(swaps[0].Amount0), "pool's own Amount0 (token0 in) must equal our exact amountIn, since tokenA is token0")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(swaps[0].Amount1)), "pool's own -Amount1 (token1 out) must equal the predicted amountOut")
	s.T().Logf("Test04ExactInputSingle: Swap sender=%s recipient=%s amount0=%s amount1=%s sqrtPriceX96=%s tick=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Tick)

	var liquidityAfter *big.Int
	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
		if err != nil {
			return false
		}
		liquidityAfter, err = s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
		return err == nil && slot0Before.SqrtPriceX96.Cmp(slot0After.SqrtPriceX96) != 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool state not visible after swap")
	s.Require().True(slot0After.SqrtPriceX96.Cmp(slot0Before.SqrtPriceX96) < 0, "selling token0 for token1 must push the price (token1 per token0) down")
	s.Require().Zero(liquidityBefore.Cmp(liquidityAfter), "swap stayed within our wide position's range, so active liquidity shouldn't have changed")
	s.T().Logf("Test04ExactInputSingle: pool after: sqrtPriceX96=%s (%s -> %s) tick=%s liquidity=%s", slot0After.SqrtPriceX96, slot0Before.SqrtPriceX96, slot0After.SqrtPriceX96, slot0After.Tick, liquidityAfter)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		return err == nil && new(big.Int).Sub(balance0Before, balance0After).Cmp(amountIn) == 0 && new(big.Int).Sub(balance1After, balance1Before).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.T().Logf("Test04ExactInputSingle: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test05ExactOutputSingle() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(200))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(100))
	s.T().Logf("Test05ExactOutputSingle: tokenIn=%s tokenOut=%s amountOut=%s approveCap=%s", s.tokenB, s.tokenA, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenB, s.router, approveCap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05ExactOutputSingle: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, approveTx.AddressNonce(), approveTx.MaxFee())

	s.T().Logf("Test05ExactOutputSingle: approve %s to spend %s of %s, broadcast tx %s", s.router, approveCap, s.tokenB, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
		return err == nil && approveCap.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test05ExactOutputSingle: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	previewParams := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, approveCap, big.NewInt(0))
	predictedAmountIn, err := s.v3Router.CallExactOutputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactOutputSingleRequest(s.router, previewParams, nil))
	s.Require().NoError(err)
	s.Require().True(predictedAmountIn.Sign() > 0)
	s.Require().True(predictedAmountIn.Cmp(approveCap) <= 0, "predicted amountIn must fit within what we already approved")
	s.T().Logf("Test05ExactOutputSingle: CallExactOutputSingle predicted amountIn=%s", predictedAmountIn)

	balance0Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	slot0Before, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	liquidityBefore, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: pool before: sqrtPriceX96=%s tick=%s liquidity=%s", slot0Before.SqrtPriceX96, slot0Before.Tick, liquidityBefore)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, predictedAmountIn, big.NewInt(0))

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactOutputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactOutputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05ExactOutputSingle: signed exactOutputSingle tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test05ExactOutputSingle: broadcast exactOutputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutputSingle transaction reverted")
	s.T().Logf("Test05ExactOutputSingle: exactOutputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().Zero(amountOut.Cmp(new(big.Int).Neg(swaps[0].Amount0)), "pool's own -Amount0 (token0 out) must equal our exact amountOut, since tokenA is token0")
	actualAmountIn := swaps[0].Amount1
	s.Require().True(actualAmountIn.Sign() > 0)
	s.Require().True(actualAmountIn.Cmp(predictedAmountIn) <= 0, "actual amountIn must not exceed the predicted amountInMaximum we approved")
	s.T().Logf("Test05ExactOutputSingle: Swap sender=%s recipient=%s amount0=%s amount1=%s sqrtPriceX96=%s tick=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Tick)

	var liquidityAfter *big.Int
	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
		if err != nil {
			return false
		}
		liquidityAfter, err = s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.pool))
		return err == nil && slot0Before.SqrtPriceX96.Cmp(slot0After.SqrtPriceX96) != 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool state not visible after swap")
	s.Require().True(slot0After.SqrtPriceX96.Cmp(slot0Before.SqrtPriceX96) > 0, "buying token0 with token1 must push the price (token1 per token0) up")
	s.Require().Zero(liquidityBefore.Cmp(liquidityAfter), "swap stayed within our wide position's range, so active liquidity shouldn't have changed")
	s.T().Logf("Test05ExactOutputSingle: pool after: sqrtPriceX96=%s (%s -> %s) tick=%s liquidity=%s", slot0After.SqrtPriceX96, slot0Before.SqrtPriceX96, slot0After.SqrtPriceX96, slot0After.Tick, liquidityAfter)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		return err == nil && new(big.Int).Sub(balance0After, balance0Before).Cmp(amountOut) == 0 && new(big.Int).Sub(balance1Before, balance1After).Cmp(actualAmountIn) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.T().Logf("Test05ExactOutputSingle: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test06CreateAndInitializePoolBC() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test06CreateAndInitializePoolBC: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenB, s.tokenC, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test06CreateAndInitializePoolBC: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test06CreateAndInitializePoolBC: signed tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test06CreateAndInitializePoolBC: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test06CreateAndInitializePoolBC: confirmed in block %d", receipt.BlockNumber)

	s.poolBC = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.poolBC))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test06CreateAndInitializePoolBC: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, nil, NewGetPoolRequest(s.factory, s.tokenB, s.tokenC, s.Fee))
		return err == nil && registeredPool.Equal(s.poolBC)
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test06CreateAndInitializePoolBC: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	poolToken0, err := s.v3Pool.Token0(s.ctx, nil, NewToken0Request(s.poolBC))
	s.Require().NoError(err)
	s.Require().True(poolToken0.Equal(s.tokenB))
	poolToken1, err := s.v3Pool.Token1(s.ctx, nil, NewToken1Request(s.poolBC))
	s.Require().NoError(err)
	s.Require().True(poolToken1.Equal(s.tokenC))
	poolFee, err := s.v3Pool.Fee(s.ctx, nil, NewFeeRequest(s.poolBC))
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(poolFee))
	s.T().Logf("Test06CreateAndInitializePoolBC: pool confirms token0=%s token1=%s fee=%s", poolToken0, poolToken1, poolFee)

	slot0, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.poolBC))
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().True(slot0.Unlocked)
	s.T().Logf("Test06CreateAndInitializePoolBC: pool slot0 confirms sqrtPriceX96=%s tick=%s unlocked=%t", slot0.SqrtPriceX96, slot0.Tick, slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.poolBC))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test06CreateAndInitializePoolBC: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test07MintBC() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test07MintBC: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenB, s.tokenC, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenB, s.tokenC} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(token, s.nfpm, amount))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test07MintBC: signed approve of %s (nonce %d, maxFee %s)", token, approveTx.AddressNonce(), approveTx.MaxFee())

		s.T().Logf("Test07MintBC: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.nfpm))
			return err == nil && amount.Cmp(allowance) == 0
		}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
		s.T().Logf("Test07MintBC: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	balance1Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	balance2Before, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test07MintBC: balances before: token1=%s token2=%s", balance1Before, balance2Before)

	params := NewMintParams(s.tokenB, s.tokenC, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test07MintBC: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.Mint(s.ctx, bind.NewTransactOptions(s.signer, builder), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test07MintBC: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test07MintBC: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test07MintBC: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	tokenIdBC := increased[0].TokenId
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test07MintBC: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, nil, NewNFPMPositionsRequest(s.nfpm, tokenIdBC))
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "position not visible after mint")
	s.Require().True(position.Token0.Equal(s.tokenB))
	s.Require().True(position.Token1.Equal(s.tokenC))
	s.Require().Zero(s.Fee.Cmp(position.Fee))
	s.Require().Zero(tickLower.Cmp(position.TickLower))
	s.Require().Zero(tickUpper.Cmp(position.TickUpper))
	s.T().Logf("Test07MintBC: positions(%s) confirms token0=%s token1=%s fee=%s tickLower=%s tickUpper=%s liquidity=%s", tokenIdBC, position.Token0, position.Token1, position.Fee, position.TickLower, position.TickUpper, position.Liquidity)

	owner, err := s.v3Nfpm.OwnerOf(s.ctx, nil, erc721.NewOwnerOfRequest(s.nfpm, tokenIdBC))
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))
	s.T().Logf("Test07MintBC: ownerOf(%s) = %s", tokenIdBC, owner)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.poolBC))
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool's active liquidity not visible after mint")
	s.T().Logf("Test07MintBC: pool liquidity() = %s, matching the position's own liquidity since it's the pool's sole LP and the current tick sits inside its range", liquidity)

	var balance1After, balance2After *big.Int
	s.Require().Eventually(func() bool {
		balance1After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		if err != nil {
			return false
		}
		balance2After, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
		return err == nil && new(big.Int).Sub(balance1Before, balance1After).Cmp(predictedAmount0) == 0 && new(big.Int).Sub(balance2Before, balance2After).Cmp(predictedAmount1) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after mint")
	s.T().Logf("Test07MintBC: balances after: token1=%s (%s -> %s) token2=%s (%s -> %s)", balance1After, balance1Before, balance1After, balance2After, balance2Before, balance2After)
}

func (s *RouterLifeCycleSuite) Test08ExactInput() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	path, err := EncodePath([]*types.Address{s.tokenA, s.tokenB, s.tokenC}, []*big.Int{s.Fee, s.Fee})
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: path=tokenA(%s) -> tokenB(%s) -> tokenC(%s) fee=%s amountIn=%s", s.tokenA, s.tokenB, s.tokenC, s.Fee, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenA, s.router, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08ExactInput: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, approveTx.AddressNonce(), approveTx.MaxFee())

	s.T().Logf("Test08ExactInput: approve %s to spend %s of %s, broadcast tx %s", s.router, amountIn, s.tokenA, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && amountIn.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test08ExactInput: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	balanceCBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: balances before: tokenA=%s tokenB=%s tokenC=%s", balanceABefore, balanceBBefore, balanceCBefore)

	slot0ABBefore, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	slot0BCBefore, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.poolBC))
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: pools before: AB sqrtPriceX96=%s BC sqrtPriceX96=%s", slot0ABBefore.SqrtPriceX96, slot0BCBefore.SqrtPriceX96)

	params := NewExactInputParams(path, s.from, amountIn, big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test08ExactInput: CallExactInput predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactInput(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08ExactInput: signed exactInput tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test08ExactInput: broadcast exactInput tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInput transaction reverted")
	s.T().Logf("Test08ExactInput: exactInput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactInput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *SwapEvent
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	s.Require().Zero(amountIn.Cmp(hopAB.Amount0), "A-B pool's own Amount0 (token0 in) must equal our exact amountIn, since tokenA is token0")
	intermediateAmount := new(big.Int).Neg(hopAB.Amount1)
	s.Require().True(intermediateAmount.Sign() > 0)
	s.Require().Zero(intermediateAmount.Cmp(hopBC.Amount0), "the tokenB amount the A-B pool paid out must equal exactly what the B-C pool received as input")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(hopBC.Amount1)), "B-C pool's own -Amount1 (token1 out) must equal the predicted final amountOut")
	s.T().Logf("Test08ExactInput: hop A-B amount0=%s amount1=%s; hop B-C amount0=%s amount1=%s; intermediate tokenB=%s", hopAB.Amount0, hopAB.Amount1, hopBC.Amount0, hopBC.Amount1, intermediateAmount)

	var slot0ABAfter, slot0BCAfter *Slot0
	s.Require().Eventually(func() bool {
		slot0ABAfter, err = s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
		if err != nil {
			return false
		}
		slot0BCAfter, err = s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.poolBC))
		return err == nil && slot0ABBefore.SqrtPriceX96.Cmp(slot0ABAfter.SqrtPriceX96) != 0 && slot0BCBefore.SqrtPriceX96.Cmp(slot0BCAfter.SqrtPriceX96) != 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool state not visible after swap")
	s.Require().True(slot0ABAfter.SqrtPriceX96.Cmp(slot0ABBefore.SqrtPriceX96) < 0, "selling token0 for token1 in the A-B pool must push its price down")
	s.Require().True(slot0BCAfter.SqrtPriceX96.Cmp(slot0BCBefore.SqrtPriceX96) < 0, "selling token0 for token1 in the B-C pool must push its price down")
	s.T().Logf("Test08ExactInput: pools after: AB sqrtPriceX96=%s (%s -> %s) BC sqrtPriceX96=%s (%s -> %s)", slot0ABAfter.SqrtPriceX96, slot0ABBefore.SqrtPriceX96, slot0ABAfter.SqrtPriceX96, slot0BCAfter.SqrtPriceX96, slot0BCBefore.SqrtPriceX96, slot0BCAfter.SqrtPriceX96)

	var balanceAAfter, balanceBAfter, balanceCAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		if err != nil {
			return false
		}
		balanceCAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
		return err == nil &&
			new(big.Int).Sub(balanceABefore, balanceAAfter).Cmp(amountIn) == 0 &&
			balanceBBefore.Cmp(balanceBAfter) == 0 &&
			new(big.Int).Sub(balanceCAfter, balanceCBefore).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.Require().Zero(balanceBBefore.Cmp(balanceBAfter), "tokenB is only an intermediate hop, so our own tokenB balance must be unchanged")
	s.T().Logf("Test08ExactInput: balances after: tokenA=%s (%s -> %s) tokenB=%s (unchanged) tokenC=%s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter, balanceBAfter, balanceCAfter, balanceCBefore, balanceCAfter)
}

func (s *RouterLifeCycleSuite) Test09ExactInputSingleTooLittleReceived() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountOutMinimum := new(big.Int).Mul(amountIn, big.NewInt(2))
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: amountIn=%s amountOutMinimum=%s", amountIn, unrealisticAmountOutMinimum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenA, s.router, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && amountIn.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, unrealisticAmountOutMinimum, big.NewInt(0))

	_, err = s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooLittleReceived)
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: callExactInputSingle correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test10ExactOutputSingleTooMuchRequested() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountInMaximum := big.NewInt(1)
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: amountOut=%s amountInMaximum=%s", amountOut, unrealisticAmountInMaximum)

	approveCap := new(big.Int).Mul(amountOut, big.NewInt(2))
	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenB, s.router, approveCap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
		return err == nil && approveCap.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, unrealisticAmountInMaximum, big.NewInt(0))

	_, err = s.v3Router.CallExactOutputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactOutputSingleRequest(s.router, params, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooMuchRequested)
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: callExactOutputSingle correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test11ExactOutput() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(2000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	path, err := EncodePath([]*types.Address{s.tokenA, s.tokenB, s.tokenC}, []*big.Int{s.Fee, s.Fee})
	s.Require().NoError(err)
	s.T().Logf("Test11ExactOutput: path=tokenC(%s) -> tokenB(%s) -> tokenA(%s) fee=%s amountOut=%s approveCap=%s", s.tokenC, s.tokenB, s.tokenA, s.Fee, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenC, s.router, approveCap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11ExactOutput: signed approve of %s (nonce %d, maxFee %s)", s.tokenC, approveTx.AddressNonce(), approveTx.MaxFee())

	s.T().Logf("Test11ExactOutput: approve %s to spend %s of %s, broadcast tx %s", s.router, approveCap, s.tokenC, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenC, s.from, s.router))
		return err == nil && approveCap.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test11ExactOutput: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	previewParams := NewExactOutputParams(path, s.from, amountOut, approveCap)
	predictedAmountIn, err := s.v3Router.CallExactOutput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactOutputRequest(s.router, previewParams, nil))
	s.Require().NoError(err)
	s.Require().True(predictedAmountIn.Sign() > 0)
	s.Require().True(predictedAmountIn.Cmp(approveCap) <= 0, "predicted amountIn must fit within what we already approved")
	s.T().Logf("Test11ExactOutput: CallExactOutput predicted amountIn=%s", predictedAmountIn)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	balanceCBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test11ExactOutput: balances before: tokenA=%s tokenB=%s tokenC=%s", balanceABefore, balanceBBefore, balanceCBefore)

	params := NewExactOutputParams(path, s.from, amountOut, predictedAmountIn)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactOutput(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactOutputRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11ExactOutput: signed exactOutput tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test11ExactOutput: broadcast exactOutput tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutput transaction reverted")
	s.T().Logf("Test11ExactOutput: exactOutput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactOutput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *SwapEvent
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	s.Require().Zero(amountOut.Cmp(new(big.Int).Neg(hopAB.Amount0)), "A-B pool's own -Amount0 (token0 out) must equal our exact amountOut, since tokenA is token0")
	intermediateAmount := hopAB.Amount1
	s.Require().True(intermediateAmount.Sign() > 0)
	s.Require().Zero(intermediateAmount.Cmp(new(big.Int).Neg(hopBC.Amount0)), "the tokenB amount the A-B pool pulled in must equal exactly what the B-C pool paid out")
	actualAmountIn := hopBC.Amount1
	s.Require().True(actualAmountIn.Sign() > 0)
	s.Require().True(actualAmountIn.Cmp(predictedAmountIn) <= 0, "actual amountIn must not exceed the predicted amountInMaximum we approved")
	s.T().Logf("Test11ExactOutput: hop A-B amount0=%s amount1=%s; hop B-C amount0=%s amount1=%s; intermediate tokenB=%s", hopAB.Amount0, hopAB.Amount1, hopBC.Amount0, hopBC.Amount1, intermediateAmount)

	var balanceAAfter, balanceBAfter, balanceCAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		if err != nil {
			return false
		}
		balanceCAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenC, s.from))
		return err == nil &&
			new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(amountOut) == 0 &&
			balanceBBefore.Cmp(balanceBAfter) == 0 &&
			new(big.Int).Sub(balanceCBefore, balanceCAfter).Cmp(actualAmountIn) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.Require().Zero(balanceBBefore.Cmp(balanceBAfter), "tokenB is only an intermediate hop, so our own tokenB balance must be unchanged")
	s.T().Logf("Test11ExactOutput: balances after: tokenA=%s (%s -> %s) tokenB=%s (unchanged) tokenC=%s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter, balanceBAfter, balanceCAfter, balanceCBefore, balanceCAfter)
}

func (s *RouterLifeCycleSuite) Test12ExactInputTooLittleReceived() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountOutMinimum := new(big.Int).Mul(amountIn, big.NewInt(2))

	path, err := EncodePath([]*types.Address{s.tokenA, s.tokenB, s.tokenC}, []*big.Int{s.Fee, s.Fee})
	s.Require().NoError(err)
	s.T().Logf("Test12ExactInputTooLittleReceived: path=tokenA(%s) -> tokenB(%s) -> tokenC(%s) amountIn=%s amountOutMinimum=%s", s.tokenA, s.tokenB, s.tokenC, amountIn, unrealisticAmountOutMinimum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenA, s.router, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test12ExactInputTooLittleReceived: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && amountIn.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test12ExactInputTooLittleReceived: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactInputParams(path, s.from, amountIn, unrealisticAmountOutMinimum)

	_, err = s.v3Router.CallExactInput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputRequest(s.router, params, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooLittleReceived)
	s.T().Logf("Test12ExactInputTooLittleReceived: callExactInput correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test13ExactOutputTooMuchRequested() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(2000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountInMaximum := big.NewInt(1)

	path, err := EncodePath([]*types.Address{s.tokenA, s.tokenB, s.tokenC}, []*big.Int{s.Fee, s.Fee})
	s.Require().NoError(err)
	s.T().Logf("Test13ExactOutputTooMuchRequested: path=tokenC(%s) -> tokenB(%s) -> tokenA(%s) amountOut=%s amountInMaximum=%s", s.tokenC, s.tokenB, s.tokenA, amountOut, unrealisticAmountInMaximum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenC, s.router, approveCap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test13ExactOutputTooMuchRequested: signed approve of %s (nonce %d, maxFee %s)", s.tokenC, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenC, s.from, s.router))
		return err == nil && approveCap.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test13ExactOutputTooMuchRequested: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactOutputParams(path, s.from, amountOut, unrealisticAmountInMaximum)

	_, err = s.v3Router.CallExactOutput(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactOutputRequest(s.router, params, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooMuchRequested)
	s.T().Logf("Test13ExactOutputTooMuchRequested: callExactOutput correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test14ExactInputSingleContractBalanceSentinel() {
	amountToSend := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	contractBalanceSentinel := big.NewInt(0)
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: sending %s of %s directly to router %s, then swapping with amountIn=CONTRACT_BALANCE(%s)", amountToSend, s.tokenA, s.router, contractBalanceSentinel)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewTransferRequest(s.tokenA, s.router, amountToSend))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: signed transfer to router (nonce %d, maxFee %s)", transferTx.AddressNonce(), transferTx.MaxFee())

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		return err == nil && amountToSend.Cmp(routerBalanceBefore) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's token balance not visible after transfer")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, contractBalanceSentinel, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: CallExactInputSingle predicted amountOut=%s (sourced from router's own balance, no approval needed from us)", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactInputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().True(swaps[0].Sender.Equal(s.router), "the router itself initiated the pool swap, since it paid from its own balance")
	s.Require().True(swaps[0].Recipient.Equal(s.from))
	s.Require().Zero(amountToSend.Cmp(swaps[0].Amount0), "the pool must have pulled exactly the router's entire tokenA balance, not just some of it")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(swaps[0].Amount1)))
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: Swap sender=%s recipient=%s amount0=%s amount1=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1)

	var routerBalanceAfter, balanceBAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		return err == nil && routerBalanceAfter.Sign() == 0 && new(big.Int).Sub(balanceBAfter, balanceBBefore).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: router's tokenA balance now %s (fully drained), our tokenB balance %s (%s -> %s)", routerBalanceAfter, balanceBAfter, balanceBBefore, balanceBAfter)
}

func (s *RouterLifeCycleSuite) Test15ExactOutputSinglePriceLimitPartialFill() {
	requestedAmountOut := new(big.Int).Div(s.mintValue, big.NewInt(500))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(100))

	slot0Before, err := s.v3Pool.Slot0(s.ctx, nil, NewSlot0Request(s.pool))
	s.Require().NoError(err)
	sqrtPriceLimitX96 := new(big.Int).Add(slot0Before.SqrtPriceX96, new(big.Int).Div(slot0Before.SqrtPriceX96, big.NewInt(10000)))
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: requestedAmountOut=%s sqrtPriceX96Before=%s sqrtPriceLimitX96=%s (a tight +0.01%% cap)", requestedAmountOut, slot0Before.SqrtPriceX96, sqrtPriceLimitX96)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenB, s.router, approveCap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
		return err == nil && approveCap.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: approve confirmed in block %d", approveReceipt.BlockNumber)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, requestedAmountOut, approveCap, sqrtPriceLimitX96)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactOutputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactOutputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: signed exactOutputSingle tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: broadcast exactOutputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutputSingle transaction must not revert even though the price limit prevents it from fully filling amountOut")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: exactOutputSingle confirmed in block %d despite the price limit", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	actualAmountOut := new(big.Int).Neg(swaps[0].Amount0)
	s.Require().True(actualAmountOut.Sign() > 0)
	s.Require().True(actualAmountOut.Cmp(requestedAmountOut) < 0, "the tight price limit must have capped the actual output below what we asked for")
	s.Require().Zero(sqrtPriceLimitX96.Cmp(swaps[0].SqrtPriceX96), "the pool must land exactly on our price limit when the limit (not the amount) is what stopped the swap")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: actualAmountOut=%s (< requested %s), pool landed exactly at sqrtPriceX96=%s", actualAmountOut, requestedAmountOut, swaps[0].SqrtPriceX96)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(actualAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: tokenA balance %s (%s -> %s), exactly the partial amount actually delivered", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test16ExactInputSingleMsgSenderRecipientSentinel() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	msgSenderSentinel := types.MustNewAddressFromHex("0x0000000000000000000000000000000000000001")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: amountIn=%s recipient=%s (MSG_SENDER sentinel, should resolve to %s)", amountIn, msgSenderSentinel, s.from)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(s.tokenA, s.router, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, approveTx.AddressNonce(), approveTx.MaxFee())

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && amountIn.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: approve confirmed in block %d", approveReceipt.BlockNumber)

	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
	s.Require().NoError(err)
	sentinelBalanceBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, msgSenderSentinel))
	s.Require().NoError(err)
	s.Require().Zero(sentinelBalanceBefore.Sign(), "nothing should ever have been sent to the literal sentinel address")

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, msgSenderSentinel, amountIn, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactInputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputSingleRequest(s.router, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().True(swaps[0].Recipient.Equal(s.from), "the MSG_SENDER sentinel must resolve to the actual caller, not the literal address(1)")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: Swap recipient=%s (resolved from the sentinel)", swaps[0].Recipient)

	var balanceBAfter, sentinelBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, s.from))
		if err != nil {
			return false
		}
		sentinelBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenB, msgSenderSentinel))
		return err == nil && new(big.Int).Sub(balanceBAfter, balanceBBefore).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.Require().Zero(sentinelBalanceAfter.Sign(), "the literal sentinel address must still hold nothing")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: our tokenB balance %s (%s -> %s), sentinel address still holds %s", balanceBAfter, balanceBBefore, balanceBAfter, sentinelBalanceAfter)
}

func (s *RouterLifeCycleSuite) Test17CreateAndInitializePoolWETH() {
	s.wethToken0, s.wethToken1 = s.weth9, s.tokenA
	if s.wethToken0.Big().Cmp(s.wethToken1.Big()) > 0 {
		s.wethToken0, s.wethToken1 = s.wethToken1, s.wethToken0
	}
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test17CreateAndInitializePoolWETH: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test17CreateAndInitializePoolWETH: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder), NewCreateAndInitializePoolIfNecessaryRequest(s.nfpm, s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test17CreateAndInitializePoolWETH: signed tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test17CreateAndInitializePoolWETH: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test17CreateAndInitializePoolWETH: confirmed in block %d", receipt.BlockNumber)

	s.poolWETH = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.poolWETH))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test17CreateAndInitializePoolWETH: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, nil, NewGetPoolRequest(s.factory, s.wethToken0, s.wethToken1, s.Fee))
		return err == nil && registeredPool.Equal(s.poolWETH)
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test17CreateAndInitializePoolWETH: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.poolWETH))
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test17CreateAndInitializePoolWETH: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test18MintWETH() {
	amount := s.wethLiquidity
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test18MintWETH: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.wethToken0, s.wethToken1, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.wethToken0, s.wethToken1} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewApproveRequest(token, s.nfpm, amount))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test18MintWETH: signed approve of %s (nonce %d, maxFee %s)", token, approveTx.AddressNonce(), approveTx.MaxFee())

		s.T().Logf("Test18MintWETH: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		s.Require().Eventually(func() bool {
			allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(token, s.from, s.nfpm))
			return err == nil && amount.Cmp(allowance) == 0
		}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after approve")
		s.T().Logf("Test18MintWETH: approve of %s confirmed in block %d", token, approveReceipt.BlockNumber)
	}

	params := NewMintParams(s.wethToken0, s.wethToken1, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test18MintWETH: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Nfpm.Mint(s.ctx, bind.NewTransactOptions(s.signer, builder), NewNFPMMintRequest(s.nfpm, params, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test18MintWETH: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test18MintWETH: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test18MintWETH: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test18MintWETH: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, nil, NewLiquidityRequest(s.poolWETH))
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "pool's active liquidity not visible after mint")
	s.T().Logf("Test18MintWETH: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test19ExactInputSingleNativeETH() {
	value := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	s.T().Logf("Test19ExactInputSingleNativeETH: swapping %s native %s directly (tokenIn=WETH9=%s, tokenOut=tokenA=%s)", value, s.NativeSymbol, s.weth9, s.tokenA)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, value, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, value))
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test19ExactInputSingleNativeETH: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.ExactInputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputSingleRequest(s.router, params, value))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test19ExactInputSingleNativeETH: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s, value %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee(), value)

	s.T().Logf("Test19ExactInputSingleNativeETH: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test19ExactInputSingleNativeETH: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolWETH))
	s.Require().True(swaps[0].Recipient.Equal(s.from))
	wethIsToken0 := s.wethToken0.Equal(s.weth9)
	actualAmountIn := swaps[0].Amount0
	actualAmountOut := new(big.Int).Neg(swaps[0].Amount1)
	if !wethIsToken0 {
		actualAmountIn, actualAmountOut = swaps[0].Amount1, new(big.Int).Neg(swaps[0].Amount0)
	}
	s.Require().Zero(value.Cmp(actualAmountIn), "the pool must have received exactly the native value we attached, auto-wrapped into WETH9 by the router")
	s.Require().Zero(predictedAmountOut.Cmp(actualAmountOut))
	s.T().Logf("Test19ExactInputSingleNativeETH: Swap confirms amountIn(WETH9)=%s amountOut(tokenA)=%s", actualAmountIn, actualAmountOut)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after swap")
	s.T().Logf("Test19ExactInputSingleNativeETH: tokenA balance %s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test20SweepToken() {
	amountToSweep := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	s.T().Logf("Test20SweepToken: sending %s of %s directly to router %s, then sweeping it back", amountToSweep, s.tokenA, s.router)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewTransferRequest(s.tokenA, s.router, amountToSweep))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		return err == nil && amountToSweep.Cmp(routerBalanceBefore) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's token balance not visible after transfer")
	s.T().Logf("Test20SweepToken: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)

	sweepBuilder := tx.NewDynamicFeeBuilder()
	sweepBuilder.PriceBuffer(150)
	sweepBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.SweepToken(s.ctx, bind.NewTransactOptions(s.signer, sweepBuilder), NewRouterSweepTokenRequest(s.router, s.tokenA, amountToSweep, s.from))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test20SweepToken: signed sweepToken tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test20SweepToken: broadcast sweepToken tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "sweepToken transaction reverted")
	s.T().Logf("Test20SweepToken: sweepToken confirmed in block %d", receipt.BlockNumber)

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	var swept *erc20.TransferEvent
	for i := range transfers {
		if transfers[i].Contract.Equal(s.tokenA) && transfers[i].From.Equal(s.router) {
			swept = &transfers[i]
		}
	}
	s.Require().NotNil(swept, "sweepToken must emit the token's own Transfer event, from the router")
	s.Require().True(swept.To.Equal(s.from))
	s.Require().Zero(amountToSweep.Cmp(swept.Value))
	s.T().Logf("Test20SweepToken: Transfer event confirms from=%s to=%s value=%s", swept.From, swept.To, swept.Value)

	var routerBalanceAfter, balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		if err != nil {
			return false
		}
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		return err == nil && routerBalanceAfter.Sign() == 0 && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(amountToSweep) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after sweep")
	s.T().Logf("Test20SweepToken: router's tokenA balance now %s (fully swept), our tokenA balance %s (%s -> %s)", routerBalanceAfter, balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test21SweepTokenWithFee() {
	amountToSweep := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	feeBips := big.NewInt(100)
	s.T().Logf("Test21SweepTokenWithFee: sending %s of %s directly to router %s, sweeping with feeBips=%s to feeRecipient=%s", amountToSweep, s.tokenA, s.router, feeBips, s.feeRecipient)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewTransferRequest(s.tokenA, s.router, amountToSweep))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		return err == nil && amountToSweep.Cmp(routerBalanceBefore) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's token balance not visible after transfer")
	s.T().Logf("Test21SweepTokenWithFee: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)
	feeRecipientBalanceBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.feeRecipient))
	s.Require().NoError(err)

	sweepBuilder := tx.NewDynamicFeeBuilder()
	sweepBuilder.PriceBuffer(150)
	sweepBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.SweepTokenWithFee(s.ctx, bind.NewTransactOptions(s.signer, sweepBuilder), NewRouterSweepTokenWithFeeRequest(s.router, s.tokenA, amountToSweep, s.from, feeBips, s.feeRecipient))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test21SweepTokenWithFee: signed sweepTokenWithFee tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test21SweepTokenWithFee: broadcast sweepTokenWithFee tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "sweepTokenWithFee transaction reverted")
	s.T().Logf("Test21SweepTokenWithFee: sweepTokenWithFee confirmed in block %d", receipt.BlockNumber)

	expectedFee := new(big.Int).Div(new(big.Int).Mul(amountToSweep, feeBips), big.NewInt(10000))
	expectedToRecipient := new(big.Int).Sub(amountToSweep, expectedFee)

	var routerBalanceAfter, balanceAAfter, feeRecipientBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.router))
		if err != nil {
			return false
		}
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		if err != nil {
			return false
		}
		feeRecipientBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.feeRecipient))
		return err == nil && routerBalanceAfter.Sign() == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after sweep")
	s.Require().Zero(expectedToRecipient.Cmp(new(big.Int).Sub(balanceAAfter, balanceABefore)), "recipient must receive amount minus the fee")
	s.Require().Zero(expectedFee.Cmp(new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore)), "feeRecipient must receive exactly feeBips/10000 of the swept amount")
	s.T().Logf("Test21SweepTokenWithFee: router's tokenA balance now %s (fully swept), recipient got %s, feeRecipient got %s (expected fee %s)", routerBalanceAfter, new(big.Int).Sub(balanceAAfter, balanceABefore), new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore), expectedFee)
}

func (s *RouterLifeCycleSuite) Test22UnwrapWETH9() {
	amountToUnwrap := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	s.T().Logf("Test22UnwrapWETH9: wrapping %s fresh native into WETH9, sending it to router, then unwrapping it back", amountToUnwrap)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, s.weth9, amountToUnwrap, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("Test22UnwrapWETH9: wrapped %s into WETH9, confirmed in block %d", amountToUnwrap, wrapReceipt.BlockNumber)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewTransferRequest(s.weth9, s.router, amountToUnwrap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.weth9, s.router))
		return err == nil && amountToUnwrap.Cmp(routerBalanceBefore) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's WETH9 balance not visible after transfer")
	s.T().Logf("Test22UnwrapWETH9: sent %s of WETH9 to router, confirmed in block %d, router's WETH9 balance now %s", amountToUnwrap, transferReceipt.BlockNumber, routerBalanceBefore)

	var nativeBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceBeforeHex)))
	nativeBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	unwrapBuilder := tx.NewDynamicFeeBuilder()
	unwrapBuilder.PriceBuffer(150)
	unwrapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.UnwrapWETH9(s.ctx, bind.NewTransactOptions(s.signer, unwrapBuilder), NewRouterUnwrapWETH9Request(s.router, amountToUnwrap, s.from))
	s.Require().NoError(err)
	s.T().Logf("Test22UnwrapWETH9: signed unwrapWETH9 tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test22UnwrapWETH9: broadcast unwrapWETH9 tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "unwrapWETH9 transaction reverted")
	s.T().Logf("Test22UnwrapWETH9: unwrapWETH9 confirmed in block %d, gasUsed=%d", receipt.BlockNumber, receipt.GasUsed)

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()
	s.T().Logf("Test22UnwrapWETH9: actual gas cost paid = %s (gasUsed=%d, effectiveGasPrice=%s, including any L1 data fee on OP-stack chains)", gasCost, receipt.GasUsed, effectiveGasPrice)

	var routerBalanceAfter, actualReceived *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.weth9, s.router))
		if err != nil {
			return false
		}
		var nativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceAfterHex)); err != nil {
			return false
		}
		nativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		actualReceived = new(big.Int).Add(new(big.Int).Sub(nativeBalanceAfter, nativeBalanceBefore), gasCost)
		return routerBalanceAfter.Sign() == 0 && amountToUnwrap.Cmp(actualReceived) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after unwrap")
	s.T().Logf("Test22UnwrapWETH9: router's WETH9 balance now %s (fully unwrapped), we actually received %s native %s (after backing out gas cost)", routerBalanceAfter, actualReceived, s.NativeSymbol)
}

func (s *RouterLifeCycleSuite) Test23UnwrapWETH9WithFee() {
	amountToUnwrap := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	feeBips := big.NewInt(100)
	s.T().Logf("Test23UnwrapWETH9WithFee: wrapping %s fresh native into WETH9, sending it to router, unwrapping with feeBips=%s to feeRecipient=%s", amountToUnwrap, feeBips, s.feeRecipient)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, s.weth9, amountToUnwrap, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("Test23UnwrapWETH9WithFee: wrapped %s into WETH9, confirmed in block %d", amountToUnwrap, wrapReceipt.BlockNumber)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, builder), erc20.NewTransferRequest(s.weth9, s.router, amountToUnwrap))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.weth9, s.router))
		return err == nil && amountToUnwrap.Cmp(routerBalanceBefore) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's WETH9 balance not visible after transfer")
	s.T().Logf("Test23UnwrapWETH9WithFee: sent %s of WETH9 to router, confirmed in block %d", amountToUnwrap, transferReceipt.BlockNumber)

	var feeRecipientBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientBalanceBeforeHex)))
	feeRecipientBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(feeRecipientBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	unwrapBuilder := tx.NewDynamicFeeBuilder()
	unwrapBuilder.PriceBuffer(150)
	unwrapBuilder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.UnwrapWETH9WithFee(s.ctx, bind.NewTransactOptions(s.signer, unwrapBuilder), NewRouterUnwrapWETH9WithFeeRequest(s.router, amountToUnwrap, s.from, feeBips, s.feeRecipient))
	s.Require().NoError(err)
	s.T().Logf("Test23UnwrapWETH9WithFee: signed unwrapWETH9WithFee tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test23UnwrapWETH9WithFee: broadcast unwrapWETH9WithFee tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "unwrapWETH9WithFee transaction reverted")
	s.T().Logf("Test23UnwrapWETH9WithFee: unwrapWETH9WithFee confirmed in block %d", receipt.BlockNumber)

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	expectedFee := new(big.Int).Div(new(big.Int).Mul(amountToUnwrap, feeBips), big.NewInt(10000))
	expectedToRecipient := new(big.Int).Sub(amountToUnwrap, expectedFee)

	var routerBalanceAfter, feeRecipientBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.weth9, s.router))
		if err != nil {
			return false
		}
		var feeRecipientBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientBalanceAfterHex)); err != nil {
			return false
		}
		feeRecipientBalanceAfter, ok = new(big.Int).SetString(strings.TrimPrefix(feeRecipientBalanceAfterHex, "0x"), 16)
		return ok && routerBalanceAfter.Sign() == 0 && expectedFee.Cmp(new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore)) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after unwrap")
	s.T().Logf("Test23UnwrapWETH9WithFee: router's WETH9 balance now %s (fully unwrapped), feeRecipient received %s native %s (expected %s), recipient should have received %s", routerBalanceAfter, new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore), s.NativeSymbol, expectedFee, expectedToRecipient)
}

func (s *RouterLifeCycleSuite) Test24RefundETH() {
	amountIn := new(big.Int).Div(s.wethLiquidity, big.NewInt(100))
	excess := new(big.Int).Div(amountIn, big.NewInt(2))
	value := new(big.Int).Add(amountIn, excess)
	s.T().Logf("Test24RefundETH: swapping with amountIn=%s but attaching value=%s (excess=%s should be left in the router)", amountIn, value, excess)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewExactInputSingleRequest(s.router, params, value))
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test24RefundETH: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	swapTx, swapHash, err := s.v3Router.ExactInputSingle(s.ctx, bind.NewTransactOptions(s.signer, swapBuilder), NewExactInputSingleRequest(s.router, params, value))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, swapTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test24RefundETH: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s, value %s)", swapTx.AddressNonce(), swapTx.Gas(), swapTx.MaxFee(), value)

	s.T().Logf("Test24RefundETH: broadcast exactInputSingle tx %s", swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test24RefundETH: exactInputSingle confirmed in block %d", swapReceipt.BlockNumber)

	var routerNativeBalance *big.Int
	s.Require().Eventually(func() bool {
		var routerNativeBalanceHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), rpc.BlockTagLatest, &routerNativeBalanceHex)); err != nil {
			return false
		}
		var ok bool
		routerNativeBalance, ok = new(big.Int).SetString(strings.TrimPrefix(routerNativeBalanceHex, "0x"), 16)
		return ok && excess.Cmp(routerNativeBalance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "router's native balance not visible after the overpaid swap")
	s.T().Logf("Test24RefundETH: router's native balance is now %s (the leftover excess), %d WETH9 wrapped and spent", routerNativeBalance, 1)

	var nativeBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceBeforeHex)))
	nativeBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	refundBuilder := tx.NewDynamicFeeBuilder()
	refundBuilder.PriceBuffer(150)
	refundBuilder.LimitBuffer(300)
	refundTx, refundHash, err := s.v3Router.RefundETH(s.ctx, bind.NewTransactOptions(s.signer, refundBuilder), NewRouterRefundETHRequest(s.router))
	s.Require().NoError(err)
	s.T().Logf("Test24RefundETH: signed refundETH tx (nonce %d, gasLimit %d, maxFee %s)", refundTx.AddressNonce(), refundTx.Gas(), refundTx.MaxFee())

	s.T().Logf("Test24RefundETH: broadcast refundETH tx %s", refundHash)

	refundReceipt, err := tx.WaitForReceipt(s.ctx, s.client, refundHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(refundReceipt.Status, "refundETH transaction reverted")
	s.T().Logf("Test24RefundETH: refundETH confirmed in block %d", refundReceipt.BlockNumber)

	effectiveGasPriceHex, ok := refundReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(refundReceipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := refundReceipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	var routerNativeBalanceAfter, actualRefund *big.Int
	s.Require().Eventually(func() bool {
		var routerNativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), rpc.BlockTagLatest, &routerNativeBalanceAfterHex)); err != nil {
			return false
		}
		var ok bool
		routerNativeBalanceAfter, ok = new(big.Int).SetString(strings.TrimPrefix(routerNativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		var nativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceAfterHex)); err != nil {
			return false
		}
		nativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		actualRefund = new(big.Int).Add(new(big.Int).Sub(nativeBalanceAfter, nativeBalanceBefore), gasCost)
		return routerNativeBalanceAfter.Sign() == 0 && excess.Cmp(actualRefund) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "balances not visible after refund")
	s.T().Logf("Test24RefundETH: router's native balance now %s (fully refunded), we actually received %s native %s back (after backing out gas cost), matching the excess", routerNativeBalanceAfter, actualRefund, s.NativeSymbol)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenA, s.from))
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(predictedAmountOut) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "tokenA balance not visible after swap")
	s.T().Logf("Test24RefundETH: tokenA balance %s (%s -> %s), matching the swap's predicted amountOut regardless of the overpaid value", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test25SelfPermit() {
	value := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(s.tokenA))
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenA.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, nil, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test25SelfPermit: token=%s name=%s nonce=%s value=%s deadline=%s", s.tokenA, name, nonce, value, deadline)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenA, nil)
	permitTypes := eip712.Types{
		"Permit": {
			eip712.NewField("owner", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("value", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
	message := eip712.Message{
		"owner":    s.from,
		"spender":  s.router,
		"value":    value,
		"nonce":    nonce,
		"deadline": deadline,
	}
	digest, err := eip712.Hash(domain, "Permit", permitTypes, message)
	s.Require().NoError(err)
	sig, err := s.signer.Sign(digest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.SelfPermit(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSelfPermitRequest(s.router, s.tokenA, value, deadline, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test25SelfPermit: signed selfPermit tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test25SelfPermit: broadcast selfPermit tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermit transaction reverted")
	s.T().Logf("Test25SelfPermit: selfPermit confirmed in block %d", receipt.BlockNumber)

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenA, s.from, s.router))
		return err == nil && value.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after selfPermit")
	s.T().Logf("Test25SelfPermit: allowance(from, router) = %s, set purely from the holder's signature, no approve tx sent", allowance)
}

func (s *RouterLifeCycleSuite) Test26SelfPermitIfNecessary() {
	value1 := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	resetBuilder := tx.NewDynamicFeeBuilder()
	resetBuilder.PriceBuffer(150)
	resetBuilder.LimitBuffer(300)
	resetTx, resetHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, resetBuilder), erc20.NewApproveRequest(s.tokenB, s.router, big.NewInt(0)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, resetTx.MaxFee())
	s.mu.Unlock()
	resetReceipt, err := tx.WaitForReceipt(s.ctx, s.client, resetHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(resetReceipt.Status, "allowance-reset transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
		return err == nil && allowance.Sign() == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible as 0 after reset")
	s.T().Logf("Test26SelfPermitIfNecessary: reset tokenB's leftover router allowance (from Test05) back to 0, confirmed in block %d", resetReceipt.BlockNumber)

	name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(s.tokenB))
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenB.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, nil, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test26SelfPermitIfNecessary: token=%s name=%s nonce=%s value1=%s deadline=%s", s.tokenB, name, nonce, value1, deadline)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenB, nil)
	permitTypes := eip712.Types{
		"Permit": {
			eip712.NewField("owner", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("value", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
	message1 := eip712.Message{
		"owner":    s.from,
		"spender":  s.router,
		"value":    value1,
		"nonce":    nonce,
		"deadline": deadline,
	}
	digest1, err := eip712.Hash(domain, "Permit", permitTypes, message1)
	s.Require().NoError(err)
	sig1, err := s.signer.Sign(digest1)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx1, hash1, err := s.v3Router.SelfPermitIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSelfPermitIfNecessaryRequest(s.router, s.tokenB, value1, deadline, sig1))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx1.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test26SelfPermitIfNecessary: signed first selfPermitIfNecessary tx (nonce %d, maxFee %s)", sentTx1.AddressNonce(), sentTx1.MaxFee())

	receipt1, err := tx.WaitForReceipt(s.ctx, s.client, hash1, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt1.Status, "first selfPermitIfNecessary transaction reverted")

	var allowanceAfterFirst *big.Int
	s.Require().Eventually(func() bool {
		allowanceAfterFirst, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
		return err == nil && value1.Cmp(allowanceAfterFirst) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after first selfPermitIfNecessary")
	s.T().Logf("Test26SelfPermitIfNecessary: first call (allowance 0 < value1) executed the permit, allowance now %s, confirmed in block %d", allowanceAfterFirst, receipt1.BlockNumber)

	value2 := new(big.Int).Div(value1, big.NewInt(2))
	s.T().Logf("Test26SelfPermitIfNecessary: second call with value2=%s (< current allowance) and a stale/wrong signature -- must skip permit entirely and still succeed", value2)

	builder2 := tx.NewDynamicFeeBuilder()
	builder2.PriceBuffer(150)
	builder2.LimitBuffer(300)
	sentTx2, hash2, err := s.v3Router.SelfPermitIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder2), NewSelfPermitIfNecessaryRequest(s.router, s.tokenB, value2, deadline, sig1))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx2.MaxFee())
	s.mu.Unlock()

	receipt2, err := tx.WaitForReceipt(s.ctx, s.client, hash2, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt2.Status, "second selfPermitIfNecessary transaction reverted -- it should have skipped permit and succeeded trivially")

	allowanceAfterSecond, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenB, s.from, s.router))
	s.Require().NoError(err)
	s.Require().Zero(allowanceAfterFirst.Cmp(allowanceAfterSecond), "allowance must be unchanged -- confirms permit (with the stale signature) was never actually invoked")
	s.T().Logf("Test26SelfPermitIfNecessary: second call succeeded despite the wrong signature, allowance still %s, confirmed in block %d", allowanceAfterSecond, receipt2.BlockNumber)
}

func (s *RouterLifeCycleSuite) Test27SelfPermitAllowed() {
	expiry := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	name, err := s.erc20.Name(s.ctx, nil, erc20.NewNameRequest(s.tokenC))
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenC.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, nil, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test27SelfPermitAllowed: token=%s name=%s nonce=%s expiry=%s", s.tokenC, name, nonce, expiry)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenC, nil)
	permitAllowedTypes := eip712.Types{
		"Permit": {
			eip712.NewField("holder", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("expiry", abi.Uint256.String()),
			eip712.NewField("allowed", abi.Bool.String()),
		},
	}
	message := eip712.Message{
		"holder":  s.from,
		"spender": s.router,
		"nonce":   nonce,
		"expiry":  expiry,
		"allowed": true,
	}
	digest, err := eip712.Hash(domain, "Permit", permitAllowedTypes, message)
	s.Require().NoError(err)
	sig, err := s.signer.Sign(digest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.SelfPermitAllowed(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSelfPermitAllowedRequest(s.router, s.tokenC, nonce, expiry, sig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test27SelfPermitAllowed: signed selfPermitAllowed tx (nonce %d, gasLimit %d, maxFee %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee())

	s.T().Logf("Test27SelfPermitAllowed: broadcast selfPermitAllowed tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermitAllowed transaction reverted")
	s.T().Logf("Test27SelfPermitAllowed: selfPermitAllowed confirmed in block %d", receipt.BlockNumber)

	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenC, s.from, s.router))
		return err == nil && maxUint256.Cmp(allowance) == 0
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "allowance not visible after selfPermitAllowed")
	s.T().Logf("Test27SelfPermitAllowed: allowance(from, router) = %s (max, DAI-style allowed=true), set purely from the holder's signature", allowance)
}

func (s *RouterLifeCycleSuite) Test28SelfPermitAllowedIfNecessary() {
	nextExpiry := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: token=%s allowance should already be max from Test27, so a stale/wrong signature here must be skipped and still succeed", s.tokenC)

	wrongNonce := big.NewInt(999999)
	wrongSig, err := s.signer.Sign(types.NewHashFromBytes([]byte("deliberately the wrong 32-byte digest!!")))
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.SelfPermitAllowedIfNecessary(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSelfPermitAllowedIfNecessaryRequest(s.router, s.tokenC, wrongNonce, nextExpiry, wrongSig))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: signed selfPermitAllowedIfNecessary tx (nonce %d, maxFee %s)", sentTx.AddressNonce(), sentTx.MaxFee())

	s.T().Logf("Test28SelfPermitAllowedIfNecessary: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermitAllowedIfNecessary transaction reverted -- it should have skipped permit (allowance already max) and succeeded trivially")
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: succeeded despite the bogus nonce/signature, confirming allowance already being max skipped the permit call, confirmed in block %d", receipt.BlockNumber)

	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	allowance, err := s.erc20.Allowance(s.ctx, nil, erc20.NewAllowanceRequest(s.tokenC, s.from, s.router))
	s.Require().NoError(err)
	s.Require().Zero(maxUint256.Cmp(allowance), "allowance must remain max, unchanged")
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: allowance(from, router) still %s", allowance)
}

func (s *RouterLifeCycleSuite) Test29Multicall() {
	amountIn := new(big.Int).Div(s.wethLiquidity, big.NewInt(100))
	excess := new(big.Int).Div(amountIn, big.NewInt(2))
	value := new(big.Int).Add(amountIn, excess)
	s.T().Logf("Test29Multicall: bundling exactInputSingle(amountIn=%s, value=%s) + refundETH into one multicall tx (excess=%s expected back)", amountIn, value, excess)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))
	swapCalldata, err := EncodeExactInputSingle(params.TokenIn, params.TokenOut, params.Fee, params.Recipient, params.AmountIn, params.AmountOutMinimum, params.SqrtPriceLimitX96)
	s.Require().NoError(err)
	refundCalldata := EncodeRouterRefundETH()
	calls := [][]byte{swapCalldata, refundCalldata}

	results, err := s.v3Router.CallMulticall(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewRouterMulticallRequest(s.router, calls, nil, value))
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	predictedAmountOut, err := DecodeExactInputSingle(results[0])
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.Require().Empty(results[1], "refundETH has no return data")
	s.T().Logf("Test29Multicall: CallMulticall predicted amountOut=%s", predictedAmountOut)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	sentTx, hash, err := s.v3Router.Multicall(s.ctx, bind.NewTransactOptions(s.signer, builder), NewRouterMulticallRequest(s.router, calls, nil, value))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test29Multicall: signed multicall tx (nonce %d, gasLimit %d, maxFee %s, value %s)", sentTx.AddressNonce(), sentTx.Gas(), sentTx.MaxFee(), value)

	s.T().Logf("Test29Multicall: broadcast multicall tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "multicall transaction reverted")
	s.T().Logf("Test29Multicall: multicall confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolWETH))

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	beforeBlock := fmt.Sprintf("0x%x", receipt.BlockNumber-1)
	afterBlock := fmt.Sprintf("0x%x", receipt.BlockNumber)
	var tokenABefore, tokenAAfter, fromNativeBefore, fromNativeAfter, routerNativeBefore, routerNativeAfter *big.Int
	s.Require().Eventually(func() bool {
		var err error
		if tokenABefore, err = s.erc20.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(beforeBlock), erc20.NewBalanceOfRequest(s.tokenA, s.from)); err != nil {
			return false
		}
		if tokenAAfter, err = s.erc20.BalanceOf(s.ctx, bind.NewCallOptions().WithBlock(afterBlock), erc20.NewBalanceOfRequest(s.tokenA, s.from)); err != nil {
			return false
		}
		var fromBeforeHex, fromAfterHex, routerBeforeHex, routerAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), beforeBlock, &fromBeforeHex)); err != nil {
			return false
		}
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), afterBlock, &fromAfterHex)); err != nil {
			return false
		}
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), beforeBlock, &routerBeforeHex)); err != nil {
			return false
		}
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), afterBlock, &routerAfterHex)); err != nil {
			return false
		}
		var ok1, ok2, ok3, ok4 bool
		fromNativeBefore, ok1 = new(big.Int).SetString(strings.TrimPrefix(fromBeforeHex, "0x"), 16)
		fromNativeAfter, ok2 = new(big.Int).SetString(strings.TrimPrefix(fromAfterHex, "0x"), 16)
		routerNativeBefore, ok3 = new(big.Int).SetString(strings.TrimPrefix(routerBeforeHex, "0x"), 16)
		routerNativeAfter, ok4 = new(big.Int).SetString(strings.TrimPrefix(routerAfterHex, "0x"), 16)
		return ok1 && ok2 && ok3 && ok4
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable around the multicall block")

	actualRefund := new(big.Int).Sub(fromNativeAfter, fromNativeBefore)
	actualRefund.Add(actualRefund, gasCost)
	actualRefund.Add(actualRefund, value)
	expectedRefund := new(big.Int).Add(excess, routerNativeBefore)

	s.T().Logf("UniswapV3Router %s Test29Multicall Result", s.router)
	s.T().Logf("\tTokenA: %s -> %s (got %s, predicted %s)", tokenABefore, tokenAAfter, new(big.Int).Sub(tokenAAfter, tokenABefore), predictedAmountOut)
	s.T().Logf("\tNative: %s -> %s (sent %s, gas %s)", fromNativeBefore, fromNativeAfter, value, gasCost)
	s.T().Logf("\tRefund: got %s, want excess %s + router's prior balance %s = %s", actualRefund, excess, routerNativeBefore, expectedRefund)
	s.T().Logf("\tRouter native balance: %s -> %s", routerNativeBefore, routerNativeAfter)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)

	s.Require().Zero(new(big.Int).Sub(tokenAAfter, tokenABefore).Cmp(predictedAmountOut))
	s.Require().Zero(expectedRefund.Cmp(actualRefund))
	s.Require().Zero(routerNativeAfter.Sign(), "router must not retain any native balance after the bundled refund")
}
