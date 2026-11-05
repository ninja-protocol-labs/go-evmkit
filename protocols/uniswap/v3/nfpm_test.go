package v3

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type NFPMSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
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

	client        rpc.Client
	signer        core.Signer
	deployer      *deploy.NicksDeployer
	from          *types.Address
	nfpm          *types.Address
	factory       *types.Address
	weth9         *types.Address
	multicall     *types.Address
	tokenA        *types.Address
	tokenB        *types.Address
	pool          *types.Address
	tokenId       *big.Int
	liquidity     *big.Int
	tokensOwed0   *big.Int
	tokensOwed1   *big.Int
	chainID       *big.Int
	permitTokenId *big.Int

	v3Nfpm *NFPM
	erc20  *erc20.ERC20

	mu        *sync.Mutex
	totalFee  *big.Int
	mintValue *big.Int
}

func TestNFPMSuite(t *testing.T) {
	suite.Run(t, &NFPMSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		NfpmHex:       "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:  "",
		Fee:           big.NewInt(FeeMedium),
		Interval:      1500,
		Timeout:       60,
		ByteCode:      erc20ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:          core.Keccak256([]byte("go-evmkit")),
		MintValue:     "1000000000000000000000000",
	})
}

func (s *NFPMSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Nfpm = NewNFPM(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.nfpm, err = types.NewAddressFromHex(s.NfpmHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target NonfungiblePositionManager %s (fee tier %s)", s.nfpm, s.Fee)

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

	names := [2]string{"evkit test nfpm token A", "evkit test nfpm token B"}
	symbols := [2]string{"EVKNA", "EVKNB"}
	tokens := [2]*types.Address{}

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
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

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s", s.tokenA, s.tokenB)
}

func (s *NFPMSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB}, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			s.T().Logf("TearDownSuite: leftover balance of %s is %s, burning it", tb.Token, tb.Amount)

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

		s.T().Logf("NFPMSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *NFPMSuite) Test01ReadMethods() {
	factory, err := s.v3Nfpm.Factory(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(factory.IsZero())
	s.factory = factory
	s.T().Logf("Test01ReadMethods: factory() = %s", factory)

	weth9, err := s.v3Nfpm.WETH9(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.weth9 = weth9
	s.T().Logf("Test01ReadMethods: WETH9() = %s", weth9)
}

func (s *NFPMSuite) Test02CreateAndInitializePoolIfNecessary() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)
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
}

func (s *NFPMSuite) Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predicted.Equal(s.pool))
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: confirmed in block %d", receipt.BlockNumber)

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Empty(initialized, "pool was already initialized, so no second Initialize event should have been emitted")
	s.T().Logf("Test03CreateAndInitializePoolIfNecessaryAlreadyInitialized: no Initialize event emitted, as expected")
}

func (s *NFPMSuite) Test04Mint() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test04Mint: token0=%s token1=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenA, s.tokenB, amount, amount, deadline)

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
		s.T().Logf("Test04Mint: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test04Mint: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test04Mint: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	params := NewMintParams(s.tokenA, s.tokenB, s.Fee, big.NewInt(-600), big.NewInt(600), amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)
	s.T().Logf("Test04Mint: tickLower=%s tickUpper=%s amount0Min=%s amount1Min=%s recipient=%s", params.TickLower, params.TickUpper, params.Amount0Min, params.Amount1Min, params.Recipient)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test04Mint: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04Mint: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test04Mint: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test04Mint: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.tokenId = increased[0].TokenId
	s.liquidity = increased[0].Liquidity
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test04Mint: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var nfpmLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.nfpm) {
			nfpmLogs = append(nfpmLogs, log)
		}
	}
	transfers, err := erc721.ExtractTransfers(nfpmLogs)
	s.Require().NoError(err)
	var minted []erc721.Transfer
	for _, transfer := range transfers {
		if transfer.From.IsZero() {
			minted = append(minted, transfer)
		}
	}
	s.Require().Len(minted, 1)
	s.Require().Zero(s.tokenId.Cmp(minted[0].TokenId))
	s.Require().True(minted[0].To.Equal(s.from))
	s.T().Logf("Test04Mint: Transfer (mint) tokenId=%s to=%s", minted[0].TokenId, minted[0].To)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after mint")
	s.Require().True(position.Token0.Equal(s.tokenA))
	s.Require().True(position.Token1.Equal(s.tokenB))
	s.Require().Zero(s.Fee.Cmp(position.Fee))
	s.T().Logf("Test04Mint: positions(%s) confirms token0=%s token1=%s fee=%s liquidity=%s", s.tokenId, position.Token0, position.Token1, position.Fee, position.Liquidity)
}

func (s *NFPMSuite) Test05Enumerable() {
	totalSupply, err := s.v3Nfpm.TotalSupply(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(totalSupply.Sign() > 0)
	s.T().Logf("Test05Enumerable: totalSupply() = %s", totalSupply)

	balance, err := s.v3Nfpm.BalanceOf(s.ctx, s.nfpm, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(balance.Sign() > 0)
	s.T().Logf("Test05Enumerable: balanceOf(%s) = %s", s.from, balance)

	var ownedTokenIds []*big.Int
	for i := int64(0); i < balance.Int64(); i++ {
		tokenId, err := s.v3Nfpm.TokenOfOwnerByIndex(s.ctx, s.nfpm, s.from, big.NewInt(i), rpc.BlockTagLatest)
		s.Require().NoError(err)
		ownedTokenIds = append(ownedTokenIds, tokenId)
	}
	s.T().Logf("Test05Enumerable: tokenOfOwnerByIndex(%s, 0..%d) = %v", s.from, balance.Int64()-1, ownedTokenIds)

	var found bool
	for _, tokenId := range ownedTokenIds {
		if tokenId.Cmp(s.tokenId) == 0 {
			found = true
			break
		}
	}
	s.Require().True(found, "minted tokenId %s not found among owner's tokens", s.tokenId)

	globalTokenId, err := s.v3Nfpm.TokenByIndex(s.ctx, s.nfpm, new(big.Int).Sub(totalSupply, big.NewInt(1)), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test05Enumerable: tokenByIndex(totalSupply-1) = %s", globalTokenId)
}

func (s *NFPMSuite) Test06IncreaseLiquidity() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(10))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test06IncreaseLiquidity: tokenId=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenId, amount, amount, deadline)

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
		s.T().Logf("Test06IncreaseLiquidity: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test06IncreaseLiquidity: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test06IncreaseLiquidity: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	params := NewIncreaseLiquidityParams(s.tokenId, amount, amount, big.NewInt(0), big.NewInt(0), deadline)

	predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallIncreaseLiquidity(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test06IncreaseLiquidity: CallIncreaseLiquidity predicted liquidity=%s amount0=%s amount1=%s", predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateIncreaseLiquidity(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test06IncreaseLiquidity: signed increaseLiquidity tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.IncreaseLiquidity(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test06IncreaseLiquidity: broadcast increaseLiquidity tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "increaseLiquidity transaction reverted")
	s.T().Logf("Test06IncreaseLiquidity: increaseLiquidity confirmed in block %d", receipt.BlockNumber)

	var nfpmLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.nfpm) {
			nfpmLogs = append(nfpmLogs, log)
		}
	}
	increased, err := ExtractIncreaseLiquidity(nfpmLogs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().Zero(s.tokenId.Cmp(increased[0].TokenId))
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test06IncreaseLiquidity: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	expectedLiquidity := new(big.Int).Add(s.liquidity, predictedLiquidity)
	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && expectedLiquidity.Cmp(position.Liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after increaseLiquidity")
	s.liquidity = position.Liquidity
	s.T().Logf("Test06IncreaseLiquidity: positions(%s) confirms liquidity=%s", s.tokenId, position.Liquidity)
}

func (s *NFPMSuite) Test07DecreaseLiquidity() {
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test07DecreaseLiquidity: tokenId=%s liquidity=%s deadline=%s", s.tokenId, s.liquidity, deadline)

	params := NewDecreaseLiquidityParams(s.tokenId, s.liquidity, big.NewInt(0), big.NewInt(0), deadline)

	predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallDecreaseLiquidity(s.ctx, s.nfpm, s.from, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmount0.Sign() > 0)
	s.Require().True(predictedAmount1.Sign() > 0)
	s.T().Logf("Test07DecreaseLiquidity: CallDecreaseLiquidity predicted amount0=%s amount1=%s", predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateDecreaseLiquidity(s.ctx, builder, s.nfpm, s.from, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test07DecreaseLiquidity: signed decreaseLiquidity tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.DecreaseLiquidity(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test07DecreaseLiquidity: broadcast decreaseLiquidity tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "decreaseLiquidity transaction reverted")
	s.T().Logf("Test07DecreaseLiquidity: decreaseLiquidity confirmed in block %d", receipt.BlockNumber)

	decreased, err := ExtractDecreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(decreased, 1)
	s.Require().Zero(s.tokenId.Cmp(decreased[0].TokenId))
	s.Require().Zero(s.liquidity.Cmp(decreased[0].Liquidity))
	s.Require().Zero(predictedAmount0.Cmp(decreased[0].Amount0))
	s.Require().Zero(predictedAmount1.Cmp(decreased[0].Amount1))
	s.T().Logf("Test07DecreaseLiquidity: DecreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", decreased[0].TokenId, decreased[0].Liquidity, decreased[0].Amount0, decreased[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && position.Liquidity.Sign() == 0 && predictedAmount0.Cmp(position.TokensOwed0) == 0 && predictedAmount1.Cmp(position.TokensOwed1) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after decreaseLiquidity")
	s.liquidity = position.Liquidity
	s.tokensOwed0 = position.TokensOwed0
	s.tokensOwed1 = position.TokensOwed1
	s.T().Logf("Test07DecreaseLiquidity: positions(%s) confirms liquidity=%s tokensOwed0=%s tokensOwed1=%s", s.tokenId, position.Liquidity, position.TokensOwed0, position.TokensOwed1)
}

func (s *NFPMSuite) Test08Collect() {
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	s.T().Logf("Test08Collect: tokenId=%s tokensOwed0=%s tokensOwed1=%s (collecting everything via MaxUint128, not these exact figures -- NFPM's own tokensOwed bookkeeping can be off by up to 1 wei from the pool's actual payout due to independent rounding)", s.tokenId, s.tokensOwed0, s.tokensOwed1)

	params := NewCollectParams(s.tokenId, s.from, maxUint128, maxUint128)

	predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallCollect(s.ctx, s.nfpm, s.from, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmount0.Sign() > 0)
	s.Require().True(predictedAmount1.Sign() > 0)
	s.T().Logf("Test08Collect: CallCollect predicted amount0=%s amount1=%s", predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCollect(s.ctx, builder, s.nfpm, s.from, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08Collect: signed collect tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Collect(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test08Collect: broadcast collect tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "collect transaction reverted")
	s.T().Logf("Test08Collect: collect confirmed in block %d", receipt.BlockNumber)

	collected, err := ExtractPositionCollect(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(collected, 1)
	s.Require().Zero(s.tokenId.Cmp(collected[0].TokenId))
	s.Require().True(collected[0].Recipient.Equal(s.from))
	s.Require().Zero(predictedAmount0.Cmp(collected[0].Amount0))
	s.Require().Zero(predictedAmount1.Cmp(collected[0].Amount1))
	s.T().Logf("Test08Collect: Collect tokenId=%s recipient=%s amount0=%s amount1=%s", collected[0].TokenId, collected[0].Recipient, collected[0].Amount0, collected[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && position.Liquidity.Sign() == 0 && position.TokensOwed0.Sign() == 0 && position.TokensOwed1.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after collect")
	s.tokensOwed0 = position.TokensOwed0
	s.tokensOwed1 = position.TokensOwed1
	s.T().Logf("Test08Collect: positions(%s) confirms liquidity=%s tokensOwed0=%s tokensOwed1=%s", s.tokenId, position.Liquidity, position.TokensOwed0, position.TokensOwed1)
}

func (s *NFPMSuite) Test09Burn() {
	s.T().Logf("Test09Burn: tokenId=%s", s.tokenId)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateBurn(s.ctx, builder, s.nfpm, s.from, s.tokenId)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09Burn: signed burn tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Burn(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test09Burn: broadcast burn tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "burn transaction reverted")
	s.T().Logf("Test09Burn: burn confirmed in block %d", receipt.BlockNumber)

	var nfpmLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.nfpm) {
			nfpmLogs = append(nfpmLogs, log)
		}
	}
	transfers, err := erc721.ExtractTransfers(nfpmLogs)
	s.Require().NoError(err)
	var burned []erc721.Transfer
	for _, transfer := range transfers {
		if transfer.To.IsZero() {
			burned = append(burned, transfer)
		}
	}
	s.Require().Len(burned, 1)
	s.Require().Zero(s.tokenId.Cmp(burned[0].TokenId))
	s.Require().True(burned[0].From.Equal(s.from))
	s.T().Logf("Test09Burn: Transfer (burn) tokenId=%s from=%s", burned[0].TokenId, burned[0].From)

	_, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidTokenID)
	s.T().Logf("Test09Burn: positions(%s) reverted as expected after burn: %v", s.tokenId, err)
}

func (s *NFPMSuite) Test10Multicall() {
	calls := [][]byte{EncodeRefundETH()}
	s.T().Logf("Test10Multicall: calls=%d (refundETH)", len(calls))

	results, err := s.v3Nfpm.CallMulticall(s.ctx, s.nfpm, s.from, nil, calls, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.T().Logf("Test10Multicall: CallMulticall results=%v", results)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMulticall(s.ctx, builder, s.nfpm, s.from, nil, calls)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test10Multicall: signed multicall tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Multicall(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test10Multicall: broadcast multicall tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "multicall transaction reverted")
	s.T().Logf("Test10Multicall: multicall confirmed in block %d", receipt.BlockNumber)
}

func (s *NFPMSuite) Test11Permit() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(20))
	mintDeadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test11Permit: minting a fresh position to permit over, amount0Desired=%s amount1Desired=%s", amount, amount)

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
		s.T().Logf("Test11Permit: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test11Permit: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test11Permit: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	mintParams := NewMintParams(s.tokenA, s.tokenB, s.Fee, big.NewInt(-600), big.NewInt(600), amount, amount, big.NewInt(0), big.NewInt(0), s.from, mintDeadline)

	mintBuilder := tx.NewDynamicFeeBuilder()
	mintBuilder.PriceBuffer(150)
	mintBuilder.LimitBuffer(300)
	unsignedMint, err := s.v3Nfpm.CreateMint(s.ctx, mintBuilder, s.nfpm, s.from, nil, mintParams)
	s.Require().NoError(err)
	s.Require().NoError(unsignedMint.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11Permit: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsignedMint.AddressNonce(), unsignedMint.Gas(), unsignedMint.MaxFee())

	mintHash, err := s.v3Nfpm.Mint(s.ctx, unsignedMint)
	s.Require().NoError(err)
	s.T().Logf("Test11Permit: broadcast mint tx %s", mintHash)

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")
	s.T().Logf("Test11Permit: mint confirmed in block %d", mintReceipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(mintReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	permitTokenId := increased[0].TokenId
	s.permitTokenId = permitTokenId
	s.T().Logf("Test11Permit: minted a fresh position tokenId=%s liquidity=%s for the permit test", permitTokenId, increased[0].Liquidity)

	spenderKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	spender := core.PubkeyToAddress(spenderKey.PublicKey())
	s.T().Logf("Test11Permit: approving spender=%s (a fresh unrelated address, != owner %s)", spender, s.from)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, permitTokenId, rpc.BlockTagLatest)
		return err == nil
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after mint")
	nonce := position.Nonce
	s.T().Logf("Test11Permit: positions(%s).nonce = %s", permitTokenId, nonce)

	onchainDomainSeparator, err := s.v3Nfpm.DomainSeparator(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test11Permit: DOMAIN_SEPARATOR() = 0x%x", onchainDomainSeparator)
	onchainPermitTypehash, err := s.v3Nfpm.PermitTypehash(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test11Permit: PERMIT_TYPEHASH() = 0x%x", onchainPermitTypehash)
	permitDeadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	domain := NewNFTPermitDomain(s.chainID, s.nfpm)

	localDomainSeparator, err := eip712.DomainSeparator(domain)
	s.Require().NoError(err)
	s.Require().Equal(onchainDomainSeparator[:], localDomainSeparator.Bytes())
	localPermitTypehash := nftPermitTypes.TypeHash(nftPermitPrimaryType)
	s.Require().Equal(onchainPermitTypehash[:], localPermitTypehash.Bytes())
	s.T().Logf("Test11Permit: locally-derived domain separator and PERMIT_TYPEHASH match the on-chain values")

	permit := NewNFTPermit(spender, permitTokenId, nonce, permitDeadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)
	s.T().Logf("Test11Permit: tokenId=%s spender=%s nonce=%s deadline=%s signature=%s", permitTokenId, spender, nonce, permitDeadline, sig)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedPermit, err := s.v3Nfpm.CreatePermit(s.ctx, builder, s.nfpm, s.from, spender, permitTokenId, permitDeadline, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsignedPermit.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedPermit.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11Permit: signed permit tx (nonce %d, gasLimit %d, maxFee %s)", unsignedPermit.AddressNonce(), unsignedPermit.Gas(), unsignedPermit.MaxFee())

	hash, err := s.v3Nfpm.Permit(s.ctx, unsignedPermit)
	s.Require().NoError(err)
	s.T().Logf("Test11Permit: broadcast permit tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")
	s.T().Logf("Test11Permit: permit confirmed in block %d", receipt.BlockNumber)

	approved, err := s.v3Nfpm.GetApproved(s.ctx, s.nfpm, permitTokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(approved.Equal(spender))
	s.T().Logf("Test11Permit: getApproved(%s) = %s", permitTokenId, approved)
}

func (s *NFPMSuite) Test12DecreaseLiquidityNotApproved() {
	position, err := s.v3Nfpm.Positions(s.ctx, s.nfpm, s.permitTokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test12DecreaseLiquidityNotApproved: tokenId=%s liquidity=%s", s.permitTokenId, position.Liquidity)

	strangerKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	stranger := core.PubkeyToAddress(strangerKey.PublicKey())
	s.T().Logf("Test12DecreaseLiquidityNotApproved: stranger=%s (neither owner %s nor the approved spender)", stranger, s.from)

	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	params := NewDecreaseLiquidityParams(s.permitTokenId, position.Liquidity, big.NewInt(0), big.NewInt(0), deadline)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	_, err = s.v3Nfpm.CreateDecreaseLiquidity(s.ctx, builder, s.nfpm, stranger, params)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrNotApproved)
	s.T().Logf("Test12DecreaseLiquidityNotApproved: createDecreaseLiquidity(%s, as %s) reverted as expected: %v", s.permitTokenId, stranger, err)
}

func (s *NFPMSuite) Test13PermitExpired() {
	expiredDeadline := big.NewInt(time.Now().Add(-10 * time.Minute).Unix())
	s.T().Logf("Test13PermitExpired: tokenId=%s deadline=%s (already expired)", s.permitTokenId, expiredDeadline)

	spenderKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	spender := core.PubkeyToAddress(spenderKey.PublicKey())

	position, err := s.v3Nfpm.Positions(s.ctx, s.nfpm, s.permitTokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	nonce := position.Nonce

	domain := NewNFTPermitDomain(s.chainID, s.nfpm)
	permit := NewNFTPermit(spender, s.permitTokenId, nonce, expiredDeadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	_, err = s.v3Nfpm.CreatePermit(s.ctx, builder, s.nfpm, s.from, spender, s.permitTokenId, expiredDeadline, sig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrPermitExpired)
	s.T().Logf("Test13PermitExpired: createPermit(%s, deadline=%s) reverted as expected: %v", s.permitTokenId, expiredDeadline, err)
}

func (s *NFPMSuite) Test14BurnNotCleared() {
	position, err := s.v3Nfpm.Positions(s.ctx, s.nfpm, s.permitTokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(position.Liquidity.Sign() > 0)
	s.T().Logf("Test14BurnNotCleared: tokenId=%s still has liquidity=%s", s.permitTokenId, position.Liquidity)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	_, err = s.v3Nfpm.CreateBurn(s.ctx, builder, s.nfpm, s.from, s.permitTokenId)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrNotCleared)
	s.T().Logf("Test14BurnNotCleared: createBurn(%s) reverted as expected: %v", s.permitTokenId, err)
}

func (s *NFPMSuite) Test15SweepTokenInsufficientToken() {
	amountMinimum := big.NewInt(1)
	s.T().Logf("Test15SweepTokenInsufficientToken: token=%s amountMinimum=%s (NFPM holds none of it)", s.tokenA, amountMinimum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	_, err := s.v3Nfpm.CreateSweepToken(s.ctx, builder, s.nfpm, s.from, s.tokenA, amountMinimum, s.from)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientToken)
	s.T().Logf("Test15SweepTokenInsufficientToken: createSweepToken(%s, %s) reverted as expected: %v", s.tokenA, amountMinimum, err)
}

func (s *NFPMSuite) Test16MintDeadlineExpired() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(20))
	expiredDeadline := big.NewInt(time.Now().Add(-10 * time.Minute).Unix())
	s.T().Logf("Test16MintDeadlineExpired: deadline=%s (already expired)", expiredDeadline)

	params := NewMintParams(s.tokenA, s.tokenB, s.Fee, big.NewInt(-600), big.NewInt(600), amount, amount, big.NewInt(0), big.NewInt(0), s.from, expiredDeadline)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	_, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTransactionTooOld)
	s.T().Logf("Test16MintDeadlineExpired: createMint(deadline=%s) reverted as expected: %v", expiredDeadline, err)
}
