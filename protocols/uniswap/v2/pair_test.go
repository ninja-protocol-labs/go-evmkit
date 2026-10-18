package v2

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

type PairLifeCycleSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	FactoryHex    string
	Interval      int
	Timeout       int
	ByteCode      string
	Constructor   *abi.Function
	MintValue     string
	NativeSymbol  string
	Salt          *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	factory   *types.Address
	token0    *types.Address
	token1    *types.Address
	pair      *types.Address
	mintValue *big.Int

	v2Factory *Factory
	v2Pair    *Pair
	erc20     *erc20.ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPairLifeCycleSuite(t *testing.T) {
	suite.Run(t, &PairLifeCycleSuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		FactoryHex:    "",
		Interval:      1500,
		Timeout:       60,
		ByteCode:      erc20ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		MintValue:     "1000000000000000000000000",
		NativeSymbol:  "",
		Salt:          core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *PairLifeCycleSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v2Factory = NewFactory(s.client)
	s.v2Pair = NewPair(s.client)
	s.erc20 = erc20.New(s.client)

	s.mu = new(sync.Mutex)
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())
	names := [2]string{"evmkit test pair token A", "evmkit test pair token B"}
	symbols := [2]string{"EVKPA", "EVKPB"}
	tokens := [2]*types.Address{}

	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		unsigned, err := s.deployer.CreateDeployment(s.ctx, builder, s.from, deploy.DefaultNicksMethodFactory, big.NewInt(0), deployment)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

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
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, unsigned.AddressNonce(), hash)

		calldata, err := tokenMintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		var unsignedMint core.Transaction
		s.Require().Eventually(func() bool {
			u, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			if err != nil || u.Gas() < 40000 {
				return false
			}
			unsignedMint = u
			return true
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "mint gas estimate looked like a no-code call")
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()

		raw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		s.T().Logf("SetupSuite: minted %s of %s to owner %s (tx %s)", s.mintValue, token, s.from, mintHash)
	}

	s.token0, s.token1 = tokens[0], tokens[1]
	if s.token0.Big().Cmp(s.token1.Big()) > 0 {
		s.token0, s.token1 = s.token1, s.token0
	}

	createBuilder := tx.NewDynamicFeeBuilder()
	createBuilder.PriceBuffer(150)
	unsignedCreate, err := s.v2Factory.CreateCreatePair(s.ctx, createBuilder, s.factory, s.from, s.token0, s.token1)
	s.Require().NoError(err)
	s.Require().NoError(unsignedCreate.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedCreate.MaxFee())
	s.mu.Unlock()

	createHash, err := s.v2Factory.CreatePair(s.ctx, unsignedCreate)
	s.Require().NoError(err)

	createReceipt, err := tx.WaitForReceipt(s.ctx, s.client, createHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(createReceipt.Status, "createPair transaction reverted")

	created, err := ExtractPairCreated(createReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.pair = created[0].Pair

	s.T().Logf("SetupSuite: created pair %s for (%s, %s) (tx %s)", s.pair, s.token0, s.token1, createHash)
}

func (s *PairLifeCycleSuite) TearDownSuite() {
	if s.token0 != nil && s.token1 != nil {
		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, token := range []*types.Address{s.token0, s.token1} {
			balance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, balance)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, token, s.from, burnHash)
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

		s.T().Logf("PairSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PairLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v2Pair.Factory(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(factory.Equal(s.factory))

	token0, err := s.v2Pair.Token0(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(token0.Equal(s.token0))

	token1, err := s.v2Pair.Token1(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(token1.Equal(s.token1))

	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(reserves.Reserve0.Sign())
	s.Require().Zero(reserves.Reserve1.Sign())

	minimumLiquidity, err := s.v2Pair.MinimumLiquidity(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	kLast, err := s.v2Pair.KLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(kLast.Sign())

	permitTypehash, err := s.v2Pair.PermitTypehash(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	price0, err := s.v2Pair.Price0CumulativeLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(price0.Sign())

	price1, err := s.v2Pair.Price1CumulativeLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(price1.Sign())

	name, err := s.v2Pair.Name(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal("Uniswap V2", name)

	symbol, err := s.v2Pair.Symbol(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal("UNI-V2", symbol)

	decimals, err := s.v2Pair.Decimals(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	s.T().Logf("UniswapV2Pair %s Test01ReadMethods Result", s.pair)
	s.T().Logf("\tFactory: %s", factory)
	s.T().Logf("\tToken0: %s", token0)
	s.T().Logf("\tToken1: %s", token1)
	s.T().Logf("\tReserves: %s / %s (blockTimestampLast %d)", reserves.Reserve0, reserves.Reserve1, reserves.BlockTimestampLast)
	s.T().Logf("\tMinimumLiquidity: %s", minimumLiquidity)
	s.T().Logf("\tKLast: %s", kLast)
	s.T().Logf("\tPermitTypehash: %s", encoding.Hex.EncodePrefixed(permitTypehash))
	s.T().Logf("\tName/Symbol/Decimals: %s/%s/%d", name, symbol, decimals)
}

func (s *PairLifeCycleSuite) Test02Mint() {
	amount0 := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	amount1 := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	tokens := map[*types.Address]*big.Int{
		s.token0: amount0,
		s.token1: amount1,
	}
	for token, amount := range tokens {
		unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), token, s.from, s.pair, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedTransfer.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
		s.mu.Unlock()

		transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
		s.Require().NoError(err)

		transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	}

	var predicted *big.Int
	s.Require().Eventually(func() bool {
		var err error
		predicted, err = s.v2Pair.CallMint(s.ctx, s.pair, s.from, s.from, rpc.BlockTagLatest)
		return err == nil
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "deposited balances not visible before mint")

	unsignedMint, err := s.v2Pair.CreateMint(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, s.from)
	s.Require().NoError(err)
	s.Require().NoError(unsignedMint.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
	s.mu.Unlock()

	mintHash, err := s.v2Pair.Mint(s.ctx, unsignedMint)
	s.Require().NoError(err)

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")

	mints, err := ExtractMint(mintReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(mints, 1)
	s.Require().True(mints[0].Contract.Equal(s.pair))
	s.Require().True(mints[0].Sender.Equal(s.from))
	s.Require().Zero(amount0.Cmp(mints[0].Amount0))
	s.Require().Zero(amount1.Cmp(mints[0].Amount1))

	var reserves *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reserves, err = s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		return err == nil && amount0.Cmp(reserves.Reserve0) == 0 && amount1.Cmp(reserves.Reserve1) == 0
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after mint")

	lpBalance, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(predicted.Cmp(lpBalance))
	s.Require().True(lpBalance.Sign() > 0)

	s.T().Logf("UniswapV2Pair %s Test02Mint Result", s.pair)
	s.T().Logf("\tAmount0/Amount1: %s/%s", amount0, amount1)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tLP balance: %s", lpBalance)
	s.T().Logf("\tTx: %s", mintHash)
}

func (s *PairLifeCycleSuite) Test03ReadMethodsAfterMint() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(reserves.Reserve0.Sign() > 0)
	s.Require().True(reserves.Reserve1.Sign() > 0)

	kLast, err := s.v2Pair.KLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	lpBalance, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	minimumLiquidity, err := s.v2Pair.MinimumLiquidity(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	totalSupply, err := s.v2Pair.TotalSupply(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Add(lpBalance, minimumLiquidity).Cmp(totalSupply))

	s.T().Logf("UniswapV2Pair %s Test03ReadMethodsAfterMint Result", s.pair)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tKLast: %s", kLast)
	s.T().Logf("\tLP balance: %s", lpBalance)
	s.T().Logf("\tTotalSupply: %s", totalSupply)
}

func (s *PairLifeCycleSuite) Test04Swap() {
	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	amountIn := new(big.Int).Div(reservesBefore.Reserve0, big.NewInt(100))

	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reservesBefore.Reserve1)
	denominator := new(big.Int).Add(new(big.Int).Mul(reservesBefore.Reserve0, big.NewInt(1000)), amountInWithFee)
	amountOut := new(big.Int).Div(numerator, denominator)

	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.token0, s.from, s.pair, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()

	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	var unsignedSwap core.Transaction
	s.Require().Eventually(func() bool {
		u, err := s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), amountOut, s.from, nil)
		if err != nil {
			return false
		}
		unsignedSwap = u
		return true
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "deposited balance not visible before swap")
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.mu.Unlock()

	swapHash, err := s.v2Pair.Swap(s.ctx, unsignedSwap)
	s.Require().NoError(err)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pair))
	s.Require().True(swaps[0].Sender.Equal(s.from))
	s.Require().Zero(amountIn.Cmp(swaps[0].Amount0In))
	s.Require().Zero(big.NewInt(0).Cmp(swaps[0].Amount1In))
	s.Require().Zero(big.NewInt(0).Cmp(swaps[0].Amount0Out))
	s.Require().Zero(amountOut.Cmp(swaps[0].Amount1Out))
	s.Require().True(swaps[0].To.Equal(s.from))

	wantReserve0 := new(big.Int).Add(reservesBefore.Reserve0, amountIn)
	wantReserve1 := new(big.Int).Sub(reservesBefore.Reserve1, amountOut)
	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		return err == nil && wantReserve0.Cmp(reservesAfter.Reserve0) == 0 && wantReserve1.Cmp(reservesAfter.Reserve1) == 0
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after swap")

	s.T().Logf("UniswapV2Pair %s Test04Swap Result", s.pair)
	s.T().Logf("\tAmountIn (token0): %s", amountIn)
	s.T().Logf("\tAmountOut (token1): %s", amountOut)
	s.T().Logf("\tReserves: %s/%s -> %s/%s", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1)
	s.T().Logf("\tTx: %s", swapHash)
}

func (s *PairLifeCycleSuite) Test05ReadMethodsAfterSwap() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	price0, err := s.v2Pair.Price0CumulativeLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(price0.Sign() > 0)

	price1, err := s.v2Pair.Price1CumulativeLast(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(price1.Sign() > 0)

	s.T().Logf("UniswapV2Pair %s Test05ReadMethodsAfterSwap Result", s.pair)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tPrice0CumulativeLast: %s", price0)
	s.T().Logf("\tPrice1CumulativeLast: %s", price1)
}

func (s *PairLifeCycleSuite) Test06Burn() {
	lpBalanceBefore, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	burnAmount := new(big.Int).Div(lpBalanceBefore, big.NewInt(2))

	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	totalSupplyBefore, err := s.v2Pair.TotalSupply(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	wantAmount0 := new(big.Int).Div(new(big.Int).Mul(burnAmount, reservesBefore.Reserve0), totalSupplyBefore)
	wantAmount1 := new(big.Int).Div(new(big.Int).Mul(burnAmount, reservesBefore.Reserve1), totalSupplyBefore)

	unsignedTransfer, err := s.v2Pair.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, s.pair, burnAmount)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()

	transferHash, err := s.v2Pair.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	var unsignedBurn core.Transaction
	s.Require().Eventually(func() bool {
		u, err := s.v2Pair.CreateBurn(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, s.from)
		if err != nil {
			return false
		}
		unsignedBurn = u
		return true
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "deposited LP balance not visible before burn")
	s.Require().NoError(unsignedBurn.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
	s.mu.Unlock()

	burnHash, err := s.v2Pair.Burn(s.ctx, unsignedBurn)
	s.Require().NoError(err)

	burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(burnReceipt.Status, "burn transaction reverted")

	burns, err := ExtractBurn(burnReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(burns, 1)
	s.Require().True(burns[0].Contract.Equal(s.pair))
	s.Require().True(burns[0].Sender.Equal(s.from))
	s.Require().Zero(wantAmount0.Cmp(burns[0].Amount0))
	s.Require().Zero(wantAmount1.Cmp(burns[0].Amount1))
	s.Require().True(burns[0].To.Equal(s.from))

	wantReserve0 := new(big.Int).Sub(reservesBefore.Reserve0, wantAmount0)
	wantReserve1 := new(big.Int).Sub(reservesBefore.Reserve1, wantAmount1)
	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
		return err == nil && wantReserve0.Cmp(reservesAfter.Reserve0) == 0 && wantReserve1.Cmp(reservesAfter.Reserve1) == 0
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "reserves not visible after burn")

	lpBalanceAfter, err := s.v2Pair.BalanceOf(s.ctx, s.pair, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(lpBalanceBefore, burnAmount).Cmp(lpBalanceAfter))

	s.T().Logf("UniswapV2Pair %s Test06Burn Result", s.pair)
	s.T().Logf("\tBurnAmount: %s", burnAmount)
	s.T().Logf("\tAmount0/Amount1: %s/%s", wantAmount0, wantAmount1)
	s.T().Logf("\tReserves: %s/%s -> %s/%s", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1)
	s.T().Logf("\tLP balance: %s -> %s", lpBalanceBefore, lpBalanceAfter)
	s.T().Logf("\tTx: %s", burnHash)
}

func (s *PairLifeCycleSuite) Test07SwapInsufficientOutputAmount() {
	_, err := s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), big.NewInt(0), s.from, nil)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientOutputAmount)
}

func (s *PairLifeCycleSuite) Test08SwapInvalidTo() {
	_, err := s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), big.NewInt(1), s.token1, nil)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidTo)
}

func (s *PairLifeCycleSuite) Test09SwapInsufficientLiquidity() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)

	_, err = s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), reserves.Reserve1, s.from, nil)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientLiquidity)
}

func (s *PairLifeCycleSuite) Test10SwapInsufficientInputAmount() {
	_, err := s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), big.NewInt(1), s.from, nil)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientInputAmount)
}

func (s *PairLifeCycleSuite) Test11SwapK() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, s.pair, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(reserves.Reserve0.Sign() > 0, "pair has no liquidity yet")

	amountIn := new(big.Int).Div(reserves.Reserve0, big.NewInt(100))
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reserves.Reserve1)
	denominator := new(big.Int).Add(new(big.Int).Mul(reserves.Reserve0, big.NewInt(1000)), amountInWithFee)
	correctAmountOut := new(big.Int).Div(numerator, denominator)
	tooMuchAmountOut := new(big.Int).Add(correctAmountOut, big.NewInt(1))

	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.token0, s.from, s.pair, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()

	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	s.Require().Eventually(func() bool {
		_, err = s.v2Pair.CreateSwap(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, big.NewInt(0), tooMuchAmountOut, s.from, nil)
		return err != nil
	}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "deposited balance not visible before swap")
	s.Require().ErrorIs(err, ErrK)
}

func (s *PairLifeCycleSuite) Test12InitializeForbidden() {
	_, err := s.v2Pair.CreateInitialize(s.ctx, tx.NewDynamicFeeBuilder(), s.pair, s.from, s.token0, s.token1)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrForbidden)
}
