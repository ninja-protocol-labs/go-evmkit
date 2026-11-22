package v2

import (
	"context"
	"fmt"
	"math/big"
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
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type PairLifeCycleSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	FactoryHex          string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Constructor         *abi.Function
	MintValue           string
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	factory   *types.Address
	nicks     *types.Address
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
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V2.Factory == "" {
		t.Skip("testenv: no Uniswap V2 factory configured for this network")
	}

	suite.Run(t, &PairLifeCycleSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		FactoryHex:          c.Addresses.Protocol.Uniswap.V2.Factory,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		MintValue:           "1000000000000000000000000",
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *PairLifeCycleSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v2Factory = NewFactory(s.client)
	s.v2Pair = NewPair(s.client)
	s.erc20 = erc20.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
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

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)

		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

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

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s, block %d, gas %d)", symbols[i], token, deployTx.AddressNonce(), hash, receipt.BlockNumber, receipt.GasUsed)

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
		}, 20*time.Second, s.ReceiptPollInterval, "mint gas estimate looked like a no-code call")
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()

		raw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		s.T().Logf("SetupSuite: minted %s of %s to owner %s (tx %s, block %d, gas %d)", s.mintValue, token, s.from, mintHash, mintReceipt.BlockNumber, mintReceipt.GasUsed)
	}

	s.token0, s.token1 = tokens[0], tokens[1]
	if s.token0.Big().Cmp(s.token1.Big()) > 0 {
		s.token0, s.token1 = s.token1, s.token0
	}

	createBuilder := tx.NewDynamicFeeBuilder()
	createBuilder.PriceBuffer(150)
	createTx, createHash, err := s.v2Factory.CreatePair(s.ctx, bind.NewTransactOptions(s.signer, createBuilder), NewCreatePairRequest(s.factory, s.token0, s.token1))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, createTx.MaxFee())
	s.mu.Unlock()

	createReceipt, err := tx.WaitForReceipt(s.ctx, s.client, createHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(createReceipt.Status, "createPair transaction reverted")

	created, err := ExtractPairCreated(createReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.pair = created[0].Pair

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", s.pair.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 20*time.Second, s.ReceiptPollInterval, "no code visible at created pair %s", s.pair)

	s.T().Logf("SetupSuite: created pair %s for (%s, %s) (nonce %d, tx %s, block %d, gas %d)", s.pair, s.token0, s.token1, createTx.AddressNonce(), createHash, createReceipt.BlockNumber, createReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) TearDownSuite() {
	if s.token0 != nil && s.token1 != nil {
		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, token := range []*types.Address{s.token0, s.token1} {
			balance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(token, s.from))
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

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s, block %d, gas %d)", balance, token, s.from, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
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

		s.T().Logf("PairLifeCycleSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PairLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v2Pair.Factory(s.ctx, nil, NewFactoryRequest(s.pair))
	s.Require().NoError(err)
	s.Require().True(factory.Equal(s.factory))

	token0, err := s.v2Pair.Token0(s.ctx, nil, NewToken0Request(s.pair))
	s.Require().NoError(err)
	s.Require().True(token0.Equal(s.token0))

	token1, err := s.v2Pair.Token1(s.ctx, nil, NewToken1Request(s.pair))
	s.Require().NoError(err)
	s.Require().True(token1.Equal(s.token1))

	reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(reserves.Reserve0.Sign())
	s.Require().Zero(reserves.Reserve1.Sign())

	minimumLiquidity, err := s.v2Pair.MinimumLiquidity(s.ctx, nil, NewMinimumLiquidityRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(1000).Cmp(minimumLiquidity))

	kLast, err := s.v2Pair.KLast(s.ctx, nil, NewKLastRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(kLast.Sign())

	permitTypehash, err := s.v2Pair.PermitTypehash(s.ctx, nil, NewPermitTypehashRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Equal(core.Keccak256([]byte("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)")).Bytes(), permitTypehash)

	price0, err := s.v2Pair.Price0CumulativeLast(s.ctx, nil, NewPrice0CumulativeLastRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(price0.Sign())

	price1, err := s.v2Pair.Price1CumulativeLast(s.ctx, nil, NewPrice1CumulativeLastRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(price1.Sign())

	name, err := s.v2Pair.Name(s.ctx, nil, erc20.NewNameRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Equal("Uniswap V2", name)

	symbol, err := s.v2Pair.Symbol(s.ctx, nil, erc20.NewSymbolRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Equal("UNI-V2", symbol)

	decimals, err := s.v2Pair.Decimals(s.ctx, nil, erc20.NewDecimalsRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	s.T().Logf("UniswapV2Pair %s Test01ReadMethods Result", s.pair)
	s.T().Logf("\tFactory: %s", factory)
	s.T().Logf("\tToken0: %s", token0)
	s.T().Logf("\tToken1: %s", token1)
	s.T().Logf("\tReserves: %s / %s (blockTimestampLast %d)", reserves.Reserve0, reserves.Reserve1, reserves.BlockTimestampLast)
	s.T().Logf("\tMinimumLiquidity: %s", minimumLiquidity)
	s.T().Logf("\tKLast: %s", kLast)
	s.T().Logf("\tPrice0CumulativeLast: %s", price0)
	s.T().Logf("\tPrice1CumulativeLast: %s", price1)
	s.T().Logf("\tPermitTypehash: %s", encoding.Hex.EncodePrefixed(permitTypehash))
	s.T().Logf("\tName: %s", name)
	s.T().Logf("\tSymbol: %s", symbol)
	s.T().Logf("\tDecimals: %d", decimals)
}

func (s *PairLifeCycleSuite) Test02Mint() {
	amount0 := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	amount1 := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	for token, amount := range map[*types.Address]*big.Int{s.token0: amount0, s.token1: amount1} {
		transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(token, s.pair, amount))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, transferTx.MaxFee())
		s.mu.Unlock()

		transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(transferReceipt.Status, "transfer transaction reverted")

		s.T().Logf("Test02Mint: transferred %s of %s to pair (nonce %d, tx %s, block %d, gas %d)", amount, token, transferTx.AddressNonce(), transferHash, transferReceipt.BlockNumber, transferReceipt.GasUsed)
	}

	var predicted *big.Int
	s.Require().Eventually(func() bool {
		var err error
		predicted, err = s.v2Pair.CallMint(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewMintRequest(s.pair, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "deposited balances not visible before mint")

	mintTx, mintHash, err := s.v2Pair.Mint(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewMintRequest(s.pair, s.from))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, mintTx.MaxFee())
	s.mu.Unlock()

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")

	mints, err := ExtractMint(mintReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(mints, 1)
	s.Require().True(mints[0].Contract.Equal(s.pair))
	s.Require().True(mints[0].Sender.Equal(s.from))
	s.Require().Zero(amount0.Cmp(mints[0].Amount0))
	s.Require().Zero(amount1.Cmp(mints[0].Amount1))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", mintReceipt.BlockNumber))

	var reserves *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reserves, err = s.v2Pair.GetReserves(s.ctx, atBlock, NewGetReservesRequest(s.pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reserves not readable at the receipt block")
	s.Require().Zero(amount0.Cmp(reserves.Reserve0))
	s.Require().Zero(amount1.Cmp(reserves.Reserve1))

	var lpBalance *big.Int
	s.Require().Eventually(func() bool {
		var err error
		lpBalance, err = s.v2Pair.BalanceOf(s.ctx, atBlock, erc20.NewBalanceOfRequest(s.pair, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "lpBalance not readable at the receipt block")
	s.Require().Zero(predicted.Cmp(lpBalance))
	s.Require().True(lpBalance.Sign() > 0)

	s.T().Logf("UniswapV2Pair %s Test02Mint Result", s.pair)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tDeposited Amount0/Amount1: %s/%s", amount0, amount1)
	s.T().Logf("\tCallMint predicted liquidity: %s", predicted)
	s.T().Logf("\tMint event: sender=%s amount0=%s amount1=%s", mints[0].Sender, mints[0].Amount0, mints[0].Amount1)
	s.T().Logf("\tReserves: %s/%s (blockTimestampLast %d)", reserves.Reserve0, reserves.Reserve1, reserves.BlockTimestampLast)
	s.T().Logf("\tLP balance: %s", lpBalance)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", mintHash, mintTx.AddressNonce(), mintReceipt.BlockNumber, mintReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) Test03ReadMethodsAfterMint() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)
	s.Require().True(reserves.Reserve0.Sign() > 0)
	s.Require().True(reserves.Reserve1.Sign() > 0)

	kLast, err := s.v2Pair.KLast(s.ctx, nil, NewKLastRequest(s.pair))
	s.Require().NoError(err)

	lpBalance, err := s.v2Pair.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.pair, s.from))
	s.Require().NoError(err)

	minimumLiquidity, err := s.v2Pair.MinimumLiquidity(s.ctx, nil, NewMinimumLiquidityRequest(s.pair))
	s.Require().NoError(err)

	totalSupply, err := s.v2Pair.TotalSupply(s.ctx, nil, erc20.NewTotalSupplyRequest(s.pair))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Add(lpBalance, minimumLiquidity).Cmp(totalSupply))

	s.T().Logf("UniswapV2Pair %s Test03ReadMethodsAfterMint Result", s.pair)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tKLast: %s", kLast)
	s.T().Logf("\tLP balance: %s", lpBalance)
	s.T().Logf("\tMinimumLiquidity (locked at zero address): %s", minimumLiquidity)
	s.T().Logf("\tTotalSupply: %s (= LP balance + MinimumLiquidity)", totalSupply)
}

func (s *PairLifeCycleSuite) Test04Swap() {
	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	amountIn := new(big.Int).Div(reservesBefore.Reserve0, big.NewInt(100))

	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reservesBefore.Reserve1)
	denominator := new(big.Int).Add(new(big.Int).Mul(reservesBefore.Reserve0, big.NewInt(1000)), amountInWithFee)
	amountOut := new(big.Int).Div(numerator, denominator)

	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.token0, s.pair, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	s.T().Logf("Test04Swap: transferred %s of token0 %s to pair (nonce %d, tx %s, block %d, gas %d)", amountIn, s.token0, transferTx.AddressNonce(), transferHash, transferReceipt.BlockNumber, transferReceipt.GasUsed)

	var swapTx core.Transaction
	var swapHash *types.Hash
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		sent, hash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), amountOut, s.from, nil))
		if err != nil {
			s.T().Logf("Test04Swap: attempt %d failed: %v", attempt, err)
			return false
		}
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, sent.MaxFee())
		s.mu.Unlock()
		swapTx, swapHash = sent, hash
		return true
	}, 20*time.Second, s.ReceiptPollInterval, "deposited balance not visible before swap")

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pair))
	s.Require().True(swaps[0].Sender.Equal(s.from))
	s.Require().Zero(amountIn.Cmp(swaps[0].Amount0In))
	s.Require().Zero(swaps[0].Amount1In.Sign())
	s.Require().Zero(swaps[0].Amount0Out.Sign())
	s.Require().Zero(amountOut.Cmp(swaps[0].Amount1Out))
	s.Require().True(swaps[0].To.Equal(s.from))

	syncs, err := ExtractSync(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(syncs, 1)
	s.Require().True(syncs[0].Contract.Equal(s.pair))

	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", swapReceipt.BlockNumber)), NewGetReservesRequest(s.pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reservesAfter not readable at the receipt block")
	s.Require().Zero(new(big.Int).Add(reservesBefore.Reserve0, amountIn).Cmp(reservesAfter.Reserve0))
	s.Require().Zero(new(big.Int).Sub(reservesBefore.Reserve1, amountOut).Cmp(reservesAfter.Reserve1))
	s.Require().Zero(syncs[0].Reserve0.Cmp(reservesAfter.Reserve0))
	s.Require().Zero(syncs[0].Reserve1.Cmp(reservesAfter.Reserve1))

	s.T().Logf("UniswapV2Pair %s Test04Swap Result", s.pair)
	s.T().Logf("\tAmountIn (token0): %s", amountIn)
	s.T().Logf("\tAmountOut (token1): %s (x*y=k with 0.3%% fee)", amountOut)
	s.T().Logf("\tSwap event: sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s", swaps[0].Sender, swaps[0].To, swaps[0].Amount0In, swaps[0].Amount1In, swaps[0].Amount0Out, swaps[0].Amount1Out)
	s.T().Logf("\tSync event: reserve0=%s reserve1=%s", syncs[0].Reserve0, syncs[0].Reserve1)
	s.T().Logf("\tReserves: %s/%s -> %s/%s", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1)
	s.T().Logf("\tTx: %s (nonce %d, attempts %d, block %d, gas %d)", swapHash, swapTx.AddressNonce(), attempt, swapReceipt.BlockNumber, swapReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) Test05ReadMethodsAfterSwap() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	kLast, err := s.v2Pair.KLast(s.ctx, nil, NewKLastRequest(s.pair))
	s.Require().NoError(err)

	price0, err := s.v2Pair.Price0CumulativeLast(s.ctx, nil, NewPrice0CumulativeLastRequest(s.pair))
	s.Require().NoError(err)
	s.Require().True(price0.Sign() > 0)

	price1, err := s.v2Pair.Price1CumulativeLast(s.ctx, nil, NewPrice1CumulativeLastRequest(s.pair))
	s.Require().NoError(err)
	s.Require().True(price1.Sign() > 0)

	s.T().Logf("UniswapV2Pair %s Test05ReadMethodsAfterSwap Result", s.pair)
	s.T().Logf("\tReserves: %s/%s (blockTimestampLast %d)", reserves.Reserve0, reserves.Reserve1, reserves.BlockTimestampLast)
	s.T().Logf("\tKLast: %s", kLast)
	s.T().Logf("\tPrice0CumulativeLast: %s", price0)
	s.T().Logf("\tPrice1CumulativeLast: %s", price1)
}

func (s *PairLifeCycleSuite) Test06Burn() {
	lpBalanceBefore, err := s.v2Pair.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.pair, s.from))
	s.Require().NoError(err)
	burnAmount := new(big.Int).Div(lpBalanceBefore, big.NewInt(2))

	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	totalSupplyBefore, err := s.v2Pair.TotalSupply(s.ctx, nil, erc20.NewTotalSupplyRequest(s.pair))
	s.Require().NoError(err)

	wantAmount0 := new(big.Int).Div(new(big.Int).Mul(burnAmount, reservesBefore.Reserve0), totalSupplyBefore)
	wantAmount1 := new(big.Int).Div(new(big.Int).Mul(burnAmount, reservesBefore.Reserve1), totalSupplyBefore)

	transferTx, transferHash, err := s.v2Pair.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.pair, s.pair, burnAmount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	s.T().Logf("Test06Burn: transferred %s LP to pair (nonce %d, tx %s, block %d, gas %d)", burnAmount, transferTx.AddressNonce(), transferHash, transferReceipt.BlockNumber, transferReceipt.GasUsed)

	var amount0, amount1 *big.Int
	s.Require().Eventually(func() bool {
		var err error
		amount0, amount1, err = s.v2Pair.CallBurn(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewBurnRequest(s.pair, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "deposited LP balance not visible before burn")
	s.Require().Zero(wantAmount0.Cmp(amount0))
	s.Require().Zero(wantAmount1.Cmp(amount1))

	burnTx, burnHash, err := s.v2Pair.Burn(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.pair, s.from))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, burnTx.MaxFee())
	s.mu.Unlock()

	burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
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

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", burnReceipt.BlockNumber))

	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, atBlock, NewGetReservesRequest(s.pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reservesAfter not readable at the receipt block")
	s.Require().Zero(new(big.Int).Sub(reservesBefore.Reserve0, wantAmount0).Cmp(reservesAfter.Reserve0))
	s.Require().Zero(new(big.Int).Sub(reservesBefore.Reserve1, wantAmount1).Cmp(reservesAfter.Reserve1))

	var lpBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		var err error
		lpBalanceAfter, err = s.v2Pair.BalanceOf(s.ctx, atBlock, erc20.NewBalanceOfRequest(s.pair, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "lpBalanceAfter not readable at the receipt block")
	s.Require().Zero(new(big.Int).Sub(lpBalanceBefore, burnAmount).Cmp(lpBalanceAfter))

	s.T().Logf("UniswapV2Pair %s Test06Burn Result", s.pair)
	s.T().Logf("\tBurnAmount: %s of TotalSupply %s", burnAmount, totalSupplyBefore)
	s.T().Logf("\tCallBurn predicted Amount0/Amount1: %s/%s", amount0, amount1)
	s.T().Logf("\tBurn event: sender=%s to=%s amount0=%s amount1=%s", burns[0].Sender, burns[0].To, burns[0].Amount0, burns[0].Amount1)
	s.T().Logf("\tReserves: %s/%s -> %s/%s", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1)
	s.T().Logf("\tLP balance: %s -> %s", lpBalanceBefore, lpBalanceAfter)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", burnHash, burnTx.AddressNonce(), burnReceipt.BlockNumber, burnReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) Test07Skim() {
	excess := big.NewInt(1_000_000)

	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	balanceBefore, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.token1, s.from))
	s.Require().NoError(err)

	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.token1, s.pair, excess))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	skimTx, skimHash, err := s.v2Pair.Skim(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSkimRequest(s.pair, s.from))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, skimTx.MaxFee())
	s.mu.Unlock()

	skimReceipt, err := tx.WaitForReceipt(s.ctx, s.client, skimHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(skimReceipt.Status, "skim transaction reverted")

	transfers, err := erc20.ExtractTransfers(skimReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 2)
	s.Require().True(transfers[0].Contract.Equal(s.token0))
	s.Require().True(transfers[0].From.Equal(s.pair))
	s.Require().True(transfers[0].To.Equal(s.from))
	s.Require().Zero(transfers[0].Value.Sign())
	s.Require().True(transfers[1].Contract.Equal(s.token1))
	s.Require().True(transfers[1].From.Equal(s.pair))
	s.Require().True(transfers[1].To.Equal(s.from))
	s.Require().Zero(excess.Cmp(transfers[1].Value))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", skimReceipt.BlockNumber))

	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, atBlock, NewGetReservesRequest(s.pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reservesAfter not readable at the receipt block")
	s.Require().Zero(reservesBefore.Reserve0.Cmp(reservesAfter.Reserve0))
	s.Require().Zero(reservesBefore.Reserve1.Cmp(reservesAfter.Reserve1))

	var balanceAfter *big.Int
	s.Require().Eventually(func() bool {
		var err error
		balanceAfter, err = s.erc20.BalanceOf(s.ctx, atBlock, erc20.NewBalanceOfRequest(s.token1, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balanceAfter not readable at the receipt block")
	s.Require().Zero(balanceBefore.Cmp(balanceAfter))

	s.T().Logf("UniswapV2Pair %s Test07Skim Result", s.pair)
	s.T().Logf("\tExcess (token1): %s", excess)
	s.T().Logf("\tTransfer in: tx %s (nonce %d, block %d, gas %d)", transferHash, transferTx.AddressNonce(), transferReceipt.BlockNumber, transferReceipt.GasUsed)
	for _, transfer := range transfers {
		s.T().Logf("\tSkim Transfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	s.T().Logf("\tReserves: %s/%s -> %s/%s (unchanged)", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1)
	s.T().Logf("\tToken1 balance: %s -> %s", balanceBefore, balanceAfter)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", skimHash, skimTx.AddressNonce(), skimReceipt.BlockNumber, skimReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) Test08Sync() {
	donation := big.NewInt(1_000_000)

	reservesBefore, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.token0, s.pair, donation))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	syncTx, syncHash, err := s.v2Pair.Sync(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSyncRequest(s.pair))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, syncTx.MaxFee())
	s.mu.Unlock()

	syncReceipt, err := tx.WaitForReceipt(s.ctx, s.client, syncHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(syncReceipt.Status, "sync transaction reverted")

	wantReserve0 := new(big.Int).Add(reservesBefore.Reserve0, donation)

	syncs, err := ExtractSync(syncReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(syncs, 1)
	s.Require().True(syncs[0].Contract.Equal(s.pair))
	s.Require().Zero(wantReserve0.Cmp(syncs[0].Reserve0))
	s.Require().Zero(reservesBefore.Reserve1.Cmp(syncs[0].Reserve1))

	var reservesAfter *Reserves
	s.Require().Eventually(func() bool {
		var err error
		reservesAfter, err = s.v2Pair.GetReserves(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", syncReceipt.BlockNumber)), NewGetReservesRequest(s.pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reservesAfter not readable at the receipt block")
	s.Require().Zero(wantReserve0.Cmp(reservesAfter.Reserve0))
	s.Require().Zero(reservesBefore.Reserve1.Cmp(reservesAfter.Reserve1))

	s.T().Logf("UniswapV2Pair %s Test08Sync Result", s.pair)
	s.T().Logf("\tDonation (token0): %s", donation)
	s.T().Logf("\tTransfer in: tx %s (nonce %d, block %d, gas %d)", transferHash, transferTx.AddressNonce(), transferReceipt.BlockNumber, transferReceipt.GasUsed)
	s.T().Logf("\tSync event: reserve0=%s reserve1=%s", syncs[0].Reserve0, syncs[0].Reserve1)
	s.T().Logf("\tReserves: %s/%s -> %s/%s (blockTimestampLast %d)", reservesBefore.Reserve0, reservesBefore.Reserve1, reservesAfter.Reserve0, reservesAfter.Reserve1, reservesAfter.BlockTimestampLast)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", syncHash, syncTx.AddressNonce(), syncReceipt.BlockNumber, syncReceipt.GasUsed)
}

func (s *PairLifeCycleSuite) Test09SwapInsufficientOutputAmount() {
	revertTx, revertHash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), big.NewInt(0), s.from, nil))
	s.Require().ErrorIs(err, ErrInsufficientOutputAmount)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test09SwapInsufficientOutputAmount Result", s.pair)
	s.T().Logf("\tRequest: amount0Out=0 amount1Out=0 to=%s", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *PairLifeCycleSuite) Test10SwapInvalidTo() {
	revertTx, revertHash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), big.NewInt(1), s.token1, nil))
	s.Require().ErrorIs(err, ErrInvalidTo)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test10SwapInvalidTo Result", s.pair)
	s.T().Logf("\tRequest: amount0Out=0 amount1Out=1 to=%s (token1)", s.token1)
	s.T().Logf("\tError: %v", err)
}

func (s *PairLifeCycleSuite) Test11SwapInsufficientLiquidity() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)

	revertTx, revertHash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), reserves.Reserve1, s.from, nil))
	s.Require().ErrorIs(err, ErrInsufficientLiquidity)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test11SwapInsufficientLiquidity Result", s.pair)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tRequest: amount0Out=0 amount1Out=%s (= reserve1) to=%s", reserves.Reserve1, s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *PairLifeCycleSuite) Test12SwapInsufficientInputAmount() {
	revertTx, revertHash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), big.NewInt(1), s.from, nil))
	s.Require().ErrorIs(err, ErrInsufficientInputAmount)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test12SwapInsufficientInputAmount Result", s.pair)
	s.T().Logf("\tRequest: amount0Out=0 amount1Out=1 to=%s (nothing deposited)", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *PairLifeCycleSuite) Test13SwapK() {
	reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
	s.Require().NoError(err)
	s.Require().True(reserves.Reserve0.Sign() > 0, "pair has no liquidity yet")

	amountIn := new(big.Int).Div(reserves.Reserve0, big.NewInt(100))
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reserves.Reserve1)
	denominator := new(big.Int).Add(new(big.Int).Mul(reserves.Reserve0, big.NewInt(1000)), amountInWithFee)
	correctAmountOut := new(big.Int).Div(numerator, denominator)
	tooMuchAmountOut := new(big.Int).Add(correctAmountOut, big.NewInt(1))

	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(s.token0, s.pair, amountIn))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	expectedBalance := new(big.Int).Add(reserves.Reserve0, amountIn)
	s.Require().Eventually(func() bool {
		balance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.token0, s.pair))
		return err == nil && expectedBalance.Cmp(balance) == 0
	}, 20*time.Second, s.ReceiptPollInterval, "deposited balance not visible before swap")

	revertTx, revertHash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, big.NewInt(0), tooMuchAmountOut, s.from, nil))
	s.Require().ErrorIs(err, ErrK)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test13SwapK Result", s.pair)
	s.T().Logf("\tReserves: %s/%s", reserves.Reserve0, reserves.Reserve1)
	s.T().Logf("\tAmountIn (token0): %s, tx %s (nonce %d, block %d)", amountIn, transferHash, transferTx.AddressNonce(), transferReceipt.BlockNumber)
	s.T().Logf("\tAmountOut: correct %s, requested %s", correctAmountOut, tooMuchAmountOut)
	s.T().Logf("\tError: %v", err)
}

func (s *PairLifeCycleSuite) Test14InitializeForbidden() {
	revertTx, revertHash, err := s.v2Pair.Initialize(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.pair, s.token0, s.token1))
	s.Require().ErrorIs(err, ErrForbidden)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Pair %s Test14InitializeForbidden Result", s.pair)
	s.T().Logf("\tCaller: %s (factory is %s)", s.from, s.factory)
	s.T().Logf("\tError: %v", err)
}

type PairSwapSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	FactoryHex          string
	MulticallHex        string
	TokenInHex          string
	TokenOutHex         string
	AmountIn            string
	RoundTrip           bool
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	factory   *types.Address
	pair      *types.Address
	tokenIn   *types.Address
	tokenOut  *types.Address
	multicall *types.Address
	amount    *big.Int

	v2Factory *Factory
	v2Pair    *Pair
	erc20     *erc20.ERC20

	totalFee *big.Int
}

func TestPairSwapSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V2.Factory == "" {
		t.Skip("testenv: no Uniswap V2 factory configured for this network")
	}

	s := &PairSwapSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		FactoryHex:          c.Addresses.Protocol.Uniswap.V2.Factory,
		MulticallHex:        c.Addresses.Standard.Multicall,
		TokenInHex:          "",
		TokenOutHex:         "",
		AmountIn:            "1000000000000000",
		RoundTrip:           false,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
	}
	if s.TokenInHex == "" || s.TokenOutHex == "" {
		t.Skip("PairSwapSuite: TokenInHex/TokenOutHex not set")
	}

	suite.Run(t, s)
}

func (s *PairSwapSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.v2Factory = NewFactory(s.client)
	s.v2Pair = NewPair(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	s.tokenIn, err = types.NewAddressFromHex(s.TokenInHex)
	s.Require().NoError(err)

	s.tokenOut, err = types.NewAddressFromHex(s.TokenOutHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	amount, ok := new(big.Int).SetString(s.AmountIn, 10)
	s.Require().True(ok)
	s.amount = amount

	s.pair, err = s.v2Factory.GetPair(s.ctx, nil, NewGetPairRequest(s.factory, s.tokenIn, s.tokenOut))
	s.Require().NoError(err)
	s.Require().False(s.pair.IsZero(), "no pair exists for (%s, %s)", s.tokenIn, s.tokenOut)
}

func (s *PairSwapSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("PairSwapSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PairSwapSuite) TestSwap() {
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

	received := s.swap(s.tokenIn, s.tokenOut, s.amount)
	if !s.RoundTrip {
		return
	}

	final := s.swap(s.tokenOut, s.tokenIn, received)
	s.T().Logf("TestSwap: round trip %s -> %s -> %s: sent %s, received back %s", s.tokenIn, s.tokenOut, s.tokenIn, s.amount, final)

	tokenOutBalanceAfter, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(s.tokenOut, s.from))
	s.Require().NoError(err)
	s.Require().Zero(tokenOutBalanceAfter.Cmp(tokenOutMeta.Balance), "wallet's TokenOut balance changed (%s -> %s) after swapping it all back", tokenOutMeta.Balance, tokenOutBalanceAfter)
}

func (s *PairSwapSuite) swap(tokenIn, tokenOut *types.Address, amountIn *big.Int) *big.Int {
	token0, err := s.v2Pair.Token0(s.ctx, nil, NewToken0Request(s.pair))
	s.Require().NoError(err)
	token1, err := s.v2Pair.Token1(s.ctx, nil, NewToken1Request(s.pair))
	s.Require().NoError(err)

	inIsToken0 := tokenIn.Equal(token0)
	if inIsToken0 {
		s.Require().True(tokenOut.Equal(token1), "tokenOut %s doesn't match the pair's token1 %s", tokenOut, token1)
	} else {
		s.Require().True(tokenIn.Equal(token1), "tokenIn %s is neither of the pair's tokens (%s, %s)", tokenIn, token0, token1)
		s.Require().True(tokenOut.Equal(token0), "tokenOut %s doesn't match the pair's token0 %s", tokenOut, token0)
	}
	s.T().Logf("swap %s -> %s: token0=%s token1=%s (tokenIn is token0: %v)", tokenIn, tokenOut, token0, token1, inIsToken0)

	balance, err := s.erc20.BalanceOf(s.ctx, nil, erc20.NewBalanceOfRequest(tokenIn, s.from))
	s.Require().NoError(err)
	s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s of tokenIn, less than amountIn %s", balance, amountIn)
	s.T().Logf("swap %s -> %s: wallet balance of tokenIn: %s", tokenIn, tokenOut, balance)

	transferTx, transferHash, err := s.erc20.Transfer(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewTransferRequest(tokenIn, s.pair, amountIn))
	s.Require().NoError(err)
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())

	s.T().Logf("swap %s -> %s: broadcast transfer of %s tokenIn to pair, tx %s", tokenIn, tokenOut, amountIn, transferHash)

	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	s.T().Logf("swap %s -> %s: transfer confirmed in block %d", tokenIn, tokenOut, transferReceipt.BlockNumber)

	var amountOut, amount0Out, amount1Out *big.Int
	var swapHash *types.Hash
	var lastErr error
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		reserves, err := s.v2Pair.GetReserves(s.ctx, nil, NewGetReservesRequest(s.pair))
		if err != nil {
			lastErr = err
			return false
		}

		reserveIn, reserveOut := reserves.Reserve0, reserves.Reserve1
		if !inIsToken0 {
			reserveIn, reserveOut = reserves.Reserve1, reserves.Reserve0
		}

		amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
		numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
		denominator := new(big.Int).Add(new(big.Int).Mul(reserveIn, big.NewInt(1000)), amountInWithFee)
		amountOut = new(big.Int).Div(numerator, denominator)
		amountOut = new(big.Int).Div(new(big.Int).Mul(amountOut, big.NewInt(995)), big.NewInt(1000))

		amount0Out, amount1Out = big.NewInt(0), amountOut
		if !inIsToken0 {
			amount0Out, amount1Out = amountOut, big.NewInt(0)
		}

		s.T().Logf("swap %s -> %s: attempt %d: reserveIn=%s reserveOut=%s amountOut=%s", tokenIn, tokenOut, attempt, reserveIn, reserveOut, amountOut)

		swapTx, hash, err := s.v2Pair.Swap(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.pair, amount0Out, amount1Out, s.from, nil))
		if err != nil {
			lastErr = err
			s.T().Logf("swap %s -> %s: attempt %d failed: %v", tokenIn, tokenOut, attempt, err)
			return false
		}
		s.totalFee.Add(s.totalFee, swapTx.MaxFee())
		swapHash = hash
		return true
	}, s.ReceiptTimeout, s.ReceiptPollInterval, "deposited balance not visible before swap: %v", lastErr)
	s.T().Logf("swap %s -> %s: swap sent after %d attempt(s), estimated amountOut=%s, tx %s", tokenIn, tokenOut, attempt, amountOut, swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("swap %s -> %s: swap confirmed in block %d", tokenIn, tokenOut, swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)

	syncs, err := ExtractSync(swapReceipt.Logs)
	s.Require().NoError(err)

	transfers, err := erc20.ExtractTransfers(swapReceipt.Logs)
	s.Require().NoError(err)

	actualAmountOut := swaps[0].Amount1Out
	if !inIsToken0 {
		actualAmountOut = swaps[0].Amount0Out
	}

	s.T().Logf("UniswapV2Pair %s swap %s -> %s Result", s.pair, tokenIn, tokenOut)
	s.T().Logf("\tAmountIn: %s", amountIn)
	s.T().Logf("\tAmountOut: %s", actualAmountOut)
	s.T().Logf("\tTx: %s", swapHash)

	s.T().Logf("\tLogs:")
	for _, transfer := range transfers {
		s.T().Logf("\t\tTransfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	for _, syc := range syncs {
		s.T().Logf("\t\tSync: reserve0=%s reserve1=%s", syc.Reserve0, syc.Reserve1)
	}
	for _, swap := range swaps {
		s.T().Logf("\t\tSwap: sender=%s to=%s amount0In=%s amount1In=%s amount0Out=%s amount1Out=%s",
			swap.Sender, swap.To, swap.Amount0In, swap.Amount1In, swap.Amount0Out, swap.Amount1Out)
	}

	return actualAmountOut
}
