package v3

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type QuoterLifeCycleSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	QuoterHex     string
	RouterHex     string
	NfpmHex       string
	MulticallHex  string
	NativeSymbol  string
	Fee           *big.Int
	Interval      int
	Timeout       int
	ByteCode      string
	Constructor   *abi.Function
	Salt          *types.Hash
	MintValue     string

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	quoter    *types.Address
	router    *types.Address
	nfpm      *types.Address
	factory   *types.Address
	multicall *types.Address
	tokenA    *types.Address
	tokenB    *types.Address
	pool      *types.Address
	chainID   *big.Int
	tokenId   *big.Int
	poolBC    *types.Address
	tokenC    *types.Address

	v3Quoter  *Quoter
	v3Router  *Router
	v3Nfpm    *NFPM
	v3Factory *Factory
	v3Pool    *Pool
	erc20     *erc20.ERC20

	mu        *sync.Mutex
	totalFee  *big.Int
	mintValue *big.Int
}

func TestQuoterLifeCycleSuite(t *testing.T) {
	suite.Run(t, &QuoterLifeCycleSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		QuoterHex:     "",
		RouterHex:     "",
		NfpmHex:       "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:  "",
		Fee:           big.NewInt(FeeMedium),
		Interval:      1500,
		Timeout:       60,
		ByteCode:      erc20PermitFullByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String}, nil),
		Salt:          core.Keccak256([]byte("go-evmkit")),
		MintValue:     "1000000000000000000000000",
	})
}

func (s *QuoterLifeCycleSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Quoter = NewQuoter(s.client)
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

	s.quoter, err = types.NewAddressFromHex(s.QuoterHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target QuoterV2 %s", s.quoter)

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target SwapRouter %s (fee tier %s)", s.router, s.Fee)

	s.nfpm, err = types.NewAddressFromHex(s.NfpmHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target NonfungiblePositionManager %s", s.nfpm)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
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

	names := [3]string{"evkit test quoter token A", "evkit test quoter token B", "evkit test quoter token C"}
	symbols := [3]string{"EVKQA", "EVKQB", "EVKQC"}
	tokens := [3]*types.Address{}

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i]}),
			salt,
		)

		predicted, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)

		unsigned, err := s.deployer.CreateDeployment(
			s.ctx,
			builder,
			s.from,
			deploy.DefaultNicksMethodFactory,
			big.NewInt(0),
			deployment,
		)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], unsigned.AddressNonce(), unsigned.MaxFee())

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d", symbols[i], receipt.BlockNumber)

		token, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, unsigned.AddressNonce(), hash)

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

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s of %s to %s, confirmed in block %d (tx %s)", s.mintValue, token, s.from, mintReceipt.BlockNumber, mintHash)
	}

	sorted := tokens[:]
	slices.SortFunc(sorted, func(a, b *types.Address) int { return a.Big().Cmp(b.Big()) })
	s.tokenA, s.tokenB, s.tokenC = sorted[0], sorted[1], sorted[2]
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s token2=%s", s.tokenA, s.tokenB, s.tokenC)
}

func (s *QuoterLifeCycleSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil && s.tokenC != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB, s.tokenC}, false, rpc.BlockTagLatest)
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

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s, confirmed in block %d (tx %s)", tb.Amount, tb.Token, s.from, burnReceipt.BlockNumber, burnHash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("QuoterLifeCycleSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *QuoterLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v3Quoter.Factory(s.ctx, s.quoter, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(factory.IsZero())
	s.factory = factory
	s.T().Logf("Test01ReadMethods: Quoter's factory() = %s", factory)

	weth9, err := s.v3Quoter.WETH9(s.ctx, s.quoter, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.T().Logf("Test01ReadMethods: Quoter's WETH9() = %s", weth9)

	routerFactory, err := s.v3Router.Factory(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(routerFactory.Equal(factory), "Quoter and Router are different periphery contracts, but they must share the same core Factory")
	s.T().Logf("Test01ReadMethods: Router's factory() = %s matches Quoter's", routerFactory)

	routerWETH9, err := s.v3Router.WETH9(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(routerWETH9.Equal(weth9), "Quoter and Router are different periphery contracts, but they must share the same WETH9")
	s.T().Logf("Test01ReadMethods: Router's WETH9() = %s matches Quoter's", routerWETH9)

	nfpmFactory, err := s.v3Nfpm.Factory(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(nfpmFactory.Equal(factory), "Quoter and NFPM are different periphery contracts, but they must share the same core Factory")
	s.T().Logf("Test01ReadMethods: NFPM's factory() = %s matches Quoter's", nfpmFactory)

	nfpmWETH9, err := s.v3Nfpm.WETH9(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(nfpmWETH9.Equal(weth9), "Quoter and NFPM are different periphery contracts, but they must share the same WETH9")
	s.T().Logf("Test01ReadMethods: NFPM's WETH9() = %s matches Quoter's", nfpmWETH9)
}

func (s *QuoterLifeCycleSuite) Test02CreateAndInitializePoolIfNecessary() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, NewCreateAndInitializePoolIfNecessaryParams(s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, NewCreateAndInitializePoolIfNecessaryParams(s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
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
		registeredPool, err = s.v3Factory.GetPool(s.ctx, NewGetPoolParams(s.factory, s.tokenA, s.tokenB, s.Fee), rpc.BlockTagLatest)
		return err == nil && registeredPool.Equal(s.pool)
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	poolToken0, err := s.v3Pool.Token0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken0.Equal(s.tokenA))
	poolToken1, err := s.v3Pool.Token1(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken1.Equal(s.tokenB))
	poolFee, err := s.v3Pool.Fee(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(poolFee))
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool confirms token0=%s token1=%s fee=%s", poolToken0, poolToken1, poolFee)

	slot0, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().True(slot0.Unlocked)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool slot0 confirms sqrtPriceX96=%s tick=%s unlocked=%t", slot0.SqrtPriceX96, slot0.Tick, slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool liquidity() = %s", liquidity)
}

func (s *QuoterLifeCycleSuite) Test03Mint() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test03Mint: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenA, s.tokenB, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenA, s.tokenB} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.nfpm, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test03Mint: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test03Mint: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test03Mint: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	params := NewMintParams(s.tokenA, s.tokenB, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test03Mint: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03Mint: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test03Mint: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
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
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after mint")
	s.T().Logf("Test03Mint: positions(%s) confirms liquidity=%s", s.tokenId, position.Liquidity)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's active liquidity not visible after mint")
	s.T().Logf("Test03Mint: pool liquidity() = %s", liquidity)
}

func (s *QuoterLifeCycleSuite) Test04QuoteExactInputSingleMatchesRouterPreview() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	s.T().Logf("Test04QuoteExactInputSingleMatchesRouterPreview: tokenIn=%s tokenOut=%s amountIn=%s", s.tokenA, s.tokenB, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test04QuoteExactInputSingleMatchesRouterPreview: approve confirmed in block %d", approveReceipt.BlockNumber)

	quoteParams := NewQuoteExactInputSingleParams(s.tokenA, s.tokenB, amountIn, s.Fee, big.NewInt(0))
	quote, err := s.v3Quoter.CallQuoteExactInputSingle(s.ctx, s.quoter, s.from, quoteParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountOut.Sign() > 0)
	s.T().Logf("Test04QuoteExactInputSingleMatchesRouterPreview: Quoter amountOut=%s sqrtPriceX96After=%s initializedTicksCrossed=%s gasEstimate=%s", quote.AmountOut, quote.SqrtPriceX96After, quote.InitializedTicksCrossed, quote.GasEstimate)

	swapParams := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))
	routerAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, nil, swapParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test04QuoteExactInputSingleMatchesRouterPreview: Router CallExactInputSingle amountOut=%s", routerAmountOut)

	s.Require().Zero(quote.AmountOut.Cmp(routerAmountOut), "Quoter's quote and Router's own preview must agree exactly -- both simulate the same swap against the same pool state")
}

func (s *QuoterLifeCycleSuite) Test05QuoteExactInputSingleMatchesActualSwap() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: tokenIn=%s tokenOut=%s amountIn=%s", s.tokenA, s.tokenB, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: approve confirmed in block %d", approveReceipt.BlockNumber)

	quoteParams := NewQuoteExactInputSingleParams(s.tokenA, s.tokenB, amountIn, s.Fee, big.NewInt(0))
	quote, err := s.v3Quoter.CallQuoteExactInputSingle(s.ctx, s.quoter, s.from, quoteParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountOut.Sign() > 0)
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: Quoter quoted amountOut=%s sqrtPriceX96After=%s before the real swap", quote.AmountOut, quote.SqrtPriceX96After)

	swapParams := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))
	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsignedSwap, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, nil, swapParams)
	s.Require().NoError(err)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.mu.Unlock()
	swapHash, err := s.v3Router.ExactInputSingle(s.ctx, unsignedSwap)
	s.Require().NoError(err)
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: broadcast swap tx %s", swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: swap confirmed in block %d", swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	actualAmountOut := new(big.Int).Neg(swaps[0].Amount1)
	s.Require().Zero(quote.AmountOut.Cmp(actualAmountOut), "Quoter's pre-swap quote must match the real swap's actual output exactly -- nothing else touched this pool in between")
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: actual amountOut=%s matches the quote", actualAmountOut)

	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && quote.SqrtPriceX96After.Cmp(slot0After.SqrtPriceX96) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's sqrtPriceX96 not yet matching the quote's predicted sqrtPriceX96After")
	s.Require().Zero(quote.InitializedTicksCrossed.Sign(), "the swap stayed within our wide position's single range, so it shouldn't have crossed any initialized ticks")
	s.T().Logf("Test05QuoteExactInputSingleMatchesActualSwap: pool's actual post-swap sqrtPriceX96=%s matches the quote's predicted sqrtPriceX96After exactly", slot0After.SqrtPriceX96)
}

func (s *QuoterLifeCycleSuite) Test06QuoteExactOutputSingleMatchesRouterPreview() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(500))
	s.T().Logf("Test06QuoteExactOutputSingleMatchesRouterPreview: tokenIn=%s tokenOut=%s amountOut=%s approveCap=%s", s.tokenB, s.tokenA, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenB, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test06QuoteExactOutputSingleMatchesRouterPreview: approve confirmed in block %d", approveReceipt.BlockNumber)

	quoteParams := NewQuoteExactOutputSingleParams(s.tokenB, s.tokenA, amountOut, s.Fee, big.NewInt(0))
	quote, err := s.v3Quoter.CallQuoteExactOutputSingle(s.ctx, s.quoter, s.from, quoteParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountIn.Sign() > 0)
	s.T().Logf("Test06QuoteExactOutputSingleMatchesRouterPreview: Quoter amountIn=%s sqrtPriceX96After=%s", quote.AmountIn, quote.SqrtPriceX96After)

	swapParams := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, approveCap, big.NewInt(0))
	routerAmountIn, err := s.v3Router.CallExactOutputSingle(s.ctx, s.router, s.from, nil, swapParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test06QuoteExactOutputSingleMatchesRouterPreview: Router CallExactOutputSingle amountIn=%s", routerAmountIn)

	s.Require().Zero(quote.AmountIn.Cmp(routerAmountIn), "Quoter's quote and Router's own preview must agree exactly -- both simulate the same swap against the same pool state")
}

func (s *QuoterLifeCycleSuite) Test07QuoteExactOutputSingleMatchesActualSwap() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(500))
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: tokenIn=%s tokenOut=%s amountOut=%s approveCap=%s", s.tokenB, s.tokenA, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenB, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: approve confirmed in block %d", approveReceipt.BlockNumber)

	quoteParams := NewQuoteExactOutputSingleParams(s.tokenB, s.tokenA, amountOut, s.Fee, big.NewInt(0))
	quote, err := s.v3Quoter.CallQuoteExactOutputSingle(s.ctx, s.quoter, s.from, quoteParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountIn.Sign() > 0)
	s.Require().True(quote.AmountIn.Cmp(approveCap) <= 0, "quoted amountIn must fit within what we already approved")
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: Quoter quoted amountIn=%s before the real swap", quote.AmountIn)

	swapParams := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, quote.AmountIn, big.NewInt(0))
	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsignedSwap, err := s.v3Router.CreateExactOutputSingle(s.ctx, swapBuilder, s.router, s.from, nil, swapParams)
	s.Require().NoError(err)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.mu.Unlock()
	swapHash, err := s.v3Router.ExactOutputSingle(s.ctx, unsignedSwap)
	s.Require().NoError(err)
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: broadcast swap tx %s", swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: swap confirmed in block %d", swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	actualAmountIn := swaps[0].Amount1
	s.Require().Zero(quote.AmountIn.Cmp(actualAmountIn), "Quoter's pre-swap quote must match the real swap's actual input exactly -- nothing else touched this pool in between")
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: actual amountIn=%s matches the quote", actualAmountIn)

	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && quote.SqrtPriceX96After.Cmp(slot0After.SqrtPriceX96) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's sqrtPriceX96 not yet matching the quote's predicted sqrtPriceX96After")
	s.T().Logf("Test07QuoteExactOutputSingleMatchesActualSwap: pool's actual post-swap sqrtPriceX96=%s matches the quote's predicted sqrtPriceX96After exactly", slot0After.SqrtPriceX96)
}

func (s *QuoterLifeCycleSuite) Test08CreateAndInitializePoolBC() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test08CreateAndInitializePoolBC: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenB, s.tokenC, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, NewCreateAndInitializePoolIfNecessaryParams(s.nfpm, s.from, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test08CreateAndInitializePoolBC: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, NewCreateAndInitializePoolIfNecessaryParams(s.nfpm, s.from, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08CreateAndInitializePoolBC: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test08CreateAndInitializePoolBC: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test08CreateAndInitializePoolBC: confirmed in block %d", receipt.BlockNumber)

	s.poolBC = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.poolBC))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test08CreateAndInitializePoolBC: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, NewGetPoolParams(s.factory, s.tokenB, s.tokenC, s.Fee), rpc.BlockTagLatest)
		return err == nil && registeredPool.Equal(s.poolBC)
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test08CreateAndInitializePoolBC: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)
}

func (s *QuoterLifeCycleSuite) Test09MintBC() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(4))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test09MintBC: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenB, s.tokenC, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenB, s.tokenC} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.nfpm, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test09MintBC: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test09MintBC: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test09MintBC: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	params := NewMintParams(s.tokenB, s.tokenC, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test09MintBC: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09MintBC: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test09MintBC: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test09MintBC: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test09MintBC: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, s.poolBC, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's active liquidity not visible after mint")
	s.T().Logf("Test09MintBC: pool liquidity() = %s", liquidity)
}

func (s *QuoterLifeCycleSuite) Test10QuoteExactInputMatchesActualMultihopSwap() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: path=tokenA(%s) -> tokenB(%s) -> tokenC(%s) amountIn=%s", s.tokenA, s.tokenB, s.tokenC, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: approve confirmed in block %d", approveReceipt.BlockNumber)

	quote, err := s.v3Quoter.CallQuoteExactInput(s.ctx, s.quoter, s.from, path, amountIn, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountOut.Sign() > 0)
	s.Require().Len(quote.SqrtPriceX96AfterList, 2)
	s.Require().Len(quote.InitializedTicksCrossedList, 2)
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: Quoter quoted amountOut=%s sqrtPriceX96AfterList=%v before the real swap", quote.AmountOut, quote.SqrtPriceX96AfterList)

	params := NewExactInputParams(path, s.from, amountIn, big.NewInt(0))
	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInput(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	hash, err := s.v3Router.ExactInput(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: broadcast swap tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInput transaction reverted")
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: exactInput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactInput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *Swap
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	actualAmountOut := new(big.Int).Neg(hopBC.Amount1)
	s.Require().Zero(quote.AmountOut.Cmp(actualAmountOut), "Quoter's pre-swap multihop quote must match the real swap's actual final output exactly")
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: actual amountOut=%s matches the quote", actualAmountOut)

	var slot0ABAfter, slot0BCAfter *Slot0
	s.Require().Eventually(func() bool {
		slot0ABAfter, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		slot0BCAfter, err = s.v3Pool.Slot0(s.ctx, s.poolBC, rpc.BlockTagLatest)
		return err == nil && quote.SqrtPriceX96AfterList[0].Cmp(slot0ABAfter.SqrtPriceX96) == 0 && quote.SqrtPriceX96AfterList[1].Cmp(slot0BCAfter.SqrtPriceX96) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool states not yet matching the quote's predicted sqrtPriceX96AfterList")
	s.T().Logf("Test10QuoteExactInputMatchesActualMultihopSwap: both pools' actual post-swap sqrtPriceX96 match the quote's predicted list exactly: AB=%s BC=%s", slot0ABAfter.SqrtPriceX96, slot0BCAfter.SqrtPriceX96)
}

func (s *QuoterLifeCycleSuite) Test11QuoteExactOutputMatchesActualMultihopSwap() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(2000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: path=tokenC(%s) -> tokenB(%s) -> tokenA(%s) amountOut=%s approveCap=%s", s.tokenC, s.tokenB, s.tokenA, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenC, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenC, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: approve confirmed in block %d", approveReceipt.BlockNumber)

	quote, err := s.v3Quoter.CallQuoteExactOutput(s.ctx, s.quoter, s.from, path, amountOut, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(quote.AmountIn.Sign() > 0)
	s.Require().True(quote.AmountIn.Cmp(approveCap) <= 0, "quoted amountIn must fit within what we already approved")
	s.Require().Len(quote.SqrtPriceX96AfterList, 2)
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: Quoter quoted amountIn=%s sqrtPriceX96AfterList=%v before the real swap", quote.AmountIn, quote.SqrtPriceX96AfterList)

	params := NewExactOutputParams(path, s.from, amountOut, quote.AmountIn)
	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactOutput(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	hash, err := s.v3Router.ExactOutput(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: broadcast swap tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutput transaction reverted")
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: exactOutput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactOutput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *Swap
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	actualAmountIn := hopBC.Amount1
	s.Require().Zero(quote.AmountIn.Cmp(actualAmountIn), "Quoter's pre-swap multihop quote must match the real swap's actual required input exactly")
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: actual amountIn=%s matches the quote", actualAmountIn)

	var slot0ABAfter, slot0BCAfter *Slot0
	s.Require().Eventually(func() bool {
		slot0ABAfter, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		slot0BCAfter, err = s.v3Pool.Slot0(s.ctx, s.poolBC, rpc.BlockTagLatest)
		return err == nil && quote.SqrtPriceX96AfterList[0].Cmp(slot0ABAfter.SqrtPriceX96) == 0 && quote.SqrtPriceX96AfterList[1].Cmp(slot0BCAfter.SqrtPriceX96) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool states not yet matching the quote's predicted sqrtPriceX96AfterList")
	s.T().Logf("Test11QuoteExactOutputMatchesActualMultihopSwap: both pools' actual post-swap sqrtPriceX96 match the quote's predicted list exactly: AB=%s BC=%s", slot0ABAfter.SqrtPriceX96, slot0BCAfter.SqrtPriceX96)
}

func (s *QuoterLifeCycleSuite) Test12QuoteExactInputSingleNonexistentPool() {
	nonexistentFee := big.NewInt(FeeHigh)
	s.T().Logf("Test12QuoteExactInputSingleNonexistentPool: tokenIn=%s tokenOut=%s fee=%s (no pool created at this fee tier)", s.tokenA, s.tokenB, nonexistentFee)

	params := NewQuoteExactInputSingleParams(s.tokenA, s.tokenB, big.NewInt(1000), nonexistentFee, big.NewInt(0))
	_, err := s.v3Quoter.CallQuoteExactInputSingle(s.ctx, s.quoter, s.from, params, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.T().Logf("Test12QuoteExactInputSingleNonexistentPool: correctly rejected with %v", err)
}
