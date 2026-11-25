package v4

import (
	"context"
	"errors"
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
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type QuoterSuite struct {
	suite.Suite

	RPCURL                 string
	PrivateKeyHex          string
	PoolManagerHex         string
	PositionManagerHex     string
	Permit2Hex             string
	QuoterHex              string
	PoolSwapTestHex        string
	MulticallHex           string
	NicksFactoryHex        string
	NativeSymbol           string
	NativeDecimals         int
	Timeout                time.Duration
	ReceiptPollInterval    time.Duration
	ReceiptTimeout         time.Duration
	ByteCode               string
	Constructor            *abi.Function
	Salt                   *types.Hash
	MintValue              *big.Int
	Permit2ApproveAmount   *big.Int
	Permit2ApproveDuration time.Duration
	MintLiquidity          *big.Int
	MintAmountMax          *big.Int
	SwapAmountIn           *big.Int
	SwapAmountOut          *big.Int
	ExcessiveAmountOut     *big.Int

	client                  rpc.Client
	deployer                *deploy.NicksDeployer
	signer                  core.Signer
	from                    *types.Address
	poolManagerContract     *types.Address
	positionManagerContract *types.Address
	permit2Contract         *types.Address
	quoterContract          *types.Address
	poolSwapTest            *types.Address
	multicall               *types.Address
	nicks                   *types.Address

	quoter          *Quoter
	positionManager *PositionManager
	erc20           *erc20.ERC20
	permit2         *permit2.Permit2

	tokens     [3]*types.Address
	keyAB      *PoolKey
	keyBC      *PoolKey
	keyAC      *PoolKey
	stateBlock string

	positionTokenIDs []*big.Int
	positionKeys     []*PoolKey

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestQuoterSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	v4 := c.Addresses.Protocol.Uniswap.V4
	if v4.PoolManager == "" || v4.PositionManager == "" || v4.Quoter == "" || v4.PoolSwapTest == "" {
		t.Skip("testenv: no Uniswap V4 PoolManager/PositionManager/Quoter/PoolSwapTest configured for this network")
	}
	if c.Addresses.Standard.Permit2 == "" {
		t.Skip("testenv: no Permit2 configured for this network")
	}

	suite.Run(t, &QuoterSuite{
		RPCURL:                 c.RPCURL,
		PrivateKeyHex:          c.ActorPrivateKeyHex,
		PoolManagerHex:         v4.PoolManager,
		PositionManagerHex:     v4.PositionManager,
		Permit2Hex:             c.Addresses.Standard.Permit2,
		QuoterHex:              v4.Quoter,
		PoolSwapTestHex:        v4.PoolSwapTest,
		MulticallHex:           c.Addresses.Standard.Multicall,
		NicksFactoryHex:        c.Addresses.Standard.NicksFactory,
		NativeSymbol:           c.Constant.NativeSymbol,
		NativeDecimals:         c.Constant.NativeDecimals,
		Timeout:                10 * time.Minute,
		ReceiptPollInterval:    c.Constant.ReceiptPollInterval,
		ReceiptTimeout:         c.Constant.ReceiptTimeout,
		ByteCode:               testenv.ERC20ByteCodeHex,
		Constructor:            abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                   core.Keccak256([]byte("go-evmkit-v4-quoter")),
		MintValue:              new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		Permit2ApproveAmount:   new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)),
		Permit2ApproveDuration: 30 * 24 * time.Hour,
		MintLiquidity:          new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil),
		MintAmountMax:          new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
		SwapAmountIn:           new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		SwapAmountOut:          new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		ExcessiveAmountOut:     new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
	})
}

func (s *QuoterSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.quoter = NewQuoter(s.client)
	s.positionManager = NewPositionManager(s.client)
	s.erc20 = erc20.New(s.client)
	s.permit2 = permit2.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.positionManagerContract, err = types.NewAddressFromHex(s.PositionManagerHex)
	s.Require().NoError(err)
	s.permit2Contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)
	s.quoterContract, err = types.NewAddressFromHex(s.QuoterHex)
	s.Require().NoError(err)
	s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: Quoter %s, PoolManager %s, PositionManager %s, PoolSwapTest %s", s.quoterContract, s.poolManagerContract, s.positionManagerContract, s.poolSwapTest)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [3]string{"evmkit v4 quoter test token A", "evmkit v4 quoter test token B", "evmkit v4 quoter test token C"}
	symbols := [3]string{"EVQA", "EVQB", "EVQC"}
	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	expiration := big.NewInt(time.Now().Add(s.Permit2ApproveDuration).Unix())

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		deployTx, hash, err := s.deployer.Deploy(s.T().Context(), bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()

		receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.T().Context(), rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)
		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s, block %d)", symbols[i], token, deployTx.AddressNonce(), hash, receipt.BlockNumber)

		calldata, err := tokenMintFn.EncodeCall(s.from, s.MintValue)
		s.Require().NoError(err)

		mintBuilder := tx.NewDynamicFeeBuilder()
		mintBuilder.PriceBuffer(150)
		mintTx, mintHash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, mintBuilder), token, calldata)
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintTx.MaxFee())
		s.mu.Unlock()

		mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s %s to %s (tx %s)", s.MintValue, symbols[i], s.from, mintHash)

		for _, spender := range []*types.Address{s.permit2Contract, s.poolSwapTest} {
			approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(token, spender, s.MintValue))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, approveTx.MaxFee())
			s.mu.Unlock()

			approveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "erc20 approve transaction reverted")
			s.T().Logf("SetupSuite: approved %s for %s (tx %s)", spender, symbols[i], approveHash)
		}

		permit2ApproveTx, permit2ApproveHash, err := s.permit2.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), permit2.NewApproveRequest(s.permit2Contract, token, s.positionManagerContract, s.Permit2ApproveAmount, expiration))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permit2ApproveTx.MaxFee())
		s.mu.Unlock()

		permit2ApproveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, permit2ApproveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(permit2ApproveReceipt.Status, "permit2 approve transaction reverted")
		s.T().Logf("SetupSuite: Permit2 allowance for PositionManager on %s set (tx %s)", symbols[i], permit2ApproveHash)
	}

	for i := range len(s.tokens) {
		for j := i + 1; j < len(s.tokens); j++ {
			if s.tokens[i].Big().Cmp(s.tokens[j].Big()) > 0 {
				s.tokens[i], s.tokens[j] = s.tokens[j], s.tokens[i]
			}
		}
	}
	s.T().Logf("SetupSuite: sorted A=%s B=%s C=%s", s.tokens[0], s.tokens[1], s.tokens[2])

	s.keyAB = NewPoolKey(s.tokens[0], s.tokens[1], big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.keyBC = NewPoolKey(s.tokens[1], s.tokens[2], big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.keyAC = NewPoolKey(s.tokens[0], s.tokens[2], big.NewInt(3000), big.NewInt(60), types.ZeroAddress())

	for _, pool := range []struct {
		key       *PoolKey
		tickLower *big.Int
		tickUpper *big.Int
	}{
		{s.keyAB, big.NewInt(-887220), big.NewInt(887220)},
		{s.keyBC, big.NewInt(-887220), big.NewInt(887220)},
		{s.keyAC, big.NewInt(-60), big.NewInt(60)},
	} {
		key := pool.key
		initTx, initHash, err := s.positionManager.InitializePool(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializePoolRequest(s.positionManagerContract, key, new(big.Int).Lsh(big.NewInt(1), 96)))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, initTx.MaxFee())
		s.mu.Unlock()

		initReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, initHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(initReceipt.Status, "initializePool transaction reverted")

		planner := NewPlanner()
		s.Require().NoError(planner.AppendMintPosition(NewMintPositionParams(key, pool.tickLower, pool.tickUpper, s.MintLiquidity, s.MintAmountMax, s.MintAmountMax, s.from, nil)))
		s.Require().NoError(planner.AppendSettlePair(key.Currency0, key.Currency1))
		deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

		var mintTx core.Transaction
		var mintHash *types.Hash
		s.Require().Eventually(func() bool {
			mintTx, mintHash, err = s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.positionManagerContract, planner.Actions(), planner.Params(), deadline, nil))
			return !errors.Is(err, ErrPoolNotInitialized)
		}, 20*time.Second, s.ReceiptPollInterval, "pool still not initialized on the estimating node")
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintTx.MaxFee())
		s.mu.Unlock()

		mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "modifyLiquidities (mint) transaction reverted")

		var positionManagerLogs []tx.Log
		for _, log := range mintReceipt.Logs {
			if log.Address.Equal(s.positionManagerContract) {
				positionManagerLogs = append(positionManagerLogs, log)
			}
		}
		transfers, err := erc721.ExtractTransfers(positionManagerLogs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].To.Equal(s.from))
		s.positionTokenIDs = append(s.positionTokenIDs, transfers[0].TokenId)
		s.positionKeys = append(s.positionKeys, key)
		s.stateBlock = fmt.Sprintf("0x%x", mintReceipt.BlockNumber)

		s.T().Logf("SetupSuite: initialized pool %s/%s and minted liquidity %s in ticks [%s, %s] as tokenID %s (init tx %s, mint tx %s, block %d)", key.Currency0, key.Currency1, s.MintLiquidity, pool.tickLower, pool.tickUpper, transfers[0].TokenId, initHash, mintHash, mintReceipt.BlockNumber)
	}
}

func (s *QuoterSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	for i, tokenID := range s.positionTokenIDs {
		key := s.positionKeys[i]

		planner := NewPlanner()
		s.Require().NoError(planner.AppendBurnPosition(tokenID, big.NewInt(0), big.NewInt(0), nil))
		s.Require().NoError(planner.AppendTakePair(key.Currency0, key.Currency1, s.from))
		deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

		burnTx, burnHash, err := s.positionManager.ModifyLiquidities(ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.positionManagerContract, planner.Actions(), planner.Params(), deadline, nil))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, burnTx.MaxFee())
		s.mu.Unlock()

		burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(burnReceipt.Status, "modifyLiquidities (burn position) transaction reverted")

		s.T().Logf("TearDownSuite: burned position tokenID %s in pool %s/%s (tx %s, block %d, gas %d)", tokenID, key.Currency0, key.Currency1, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
	}

	if s.tokens[0] != nil && s.tokens[1] != nil && s.tokens[2] != nil {
		balances, err := s.erc20.AddressBalances(ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, s.tokens[:], false))
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			burnTx, burnHash, err := bind.Transact(ctx, s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), tb.Token, calldata)
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, burnTx.MaxFee())
			s.mu.Unlock()

			burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s, block %d, gas %d)", tb.Amount, tb.Token, s.from, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
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

		s.T().Logf("QuoterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *QuoterSuite) Test01Immutables() {
	atBlock := bind.NewCallOptions().WithBlock(s.stateBlock)

	var poolManager *types.Address
	var err error
	s.Require().Eventually(func() bool {
		poolManager, err = s.quoter.PoolManager(s.T().Context(), atBlock, NewPoolManagerRequest(s.quoterContract))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "poolManager not readable at block %s", s.stateBlock)
	s.Require().True(poolManager.Equal(s.poolManagerContract))

	msgSender, err := s.quoter.MsgSender(s.T().Context(), nil, NewMsgSenderRequest(s.quoterContract))
	s.Require().NoError(err)
	s.Require().True(msgSender.Equal(types.ZeroAddress()))

	s.T().Logf("Quoter %s Test01Immutables Result", s.quoterContract)
	s.T().Logf("\tPoolManager: %s", poolManager)
	s.T().Logf("\tMsgSender: %s", msgSender)
}

func (s *QuoterSuite) Test02QuoteExactInputSingleMatchesSwap() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var quote *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		quote, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
	s.Require().Positive(quote.AmountOut.Sign())
	s.Require().Negative(quote.AmountOut.Cmp(s.SwapAmountIn))
	s.Require().Positive(quote.GasEstimate.Sign())

	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02QuoteExactInputSingleMatchesSwap: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")
	s.stateBlock = fmt.Sprintf("0x%x", receipt.BlockNumber)

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	swaps, err := ExtractSwap(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(new(big.Int).Neg(s.SwapAmountIn).Cmp(swaps[0].Amount0))
	s.Require().Zero(quote.AmountOut.Cmp(swaps[0].Amount1), "quoted amountOut must equal the swap's actual output to the wei")

	s.T().Logf("Quoter %s Test02QuoteExactInputSingleMatchesSwap Result", s.quoterContract)
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tQuoted AmountOut: %s (gas estimate %s)", quote.AmountOut, quote.GasEstimate)
	s.T().Logf("\tSwap Amount1: %s", swaps[0].Amount1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *QuoterSuite) Test03QuoteExactInputSingleOneForZeroMatchesSwap() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var quote *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		quote, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, false, s.SwapAmountIn, nil)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
	s.Require().Positive(quote.AmountOut.Sign())
	s.Require().Positive(quote.GasEstimate.Sign())

	maxSqrtPriceMinus1, ok := new(big.Int).SetString("1461446703485210103287273052203988822378723970341", 10)
	s.Require().True(ok)

	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{false, new(big.Int).Neg(s.SwapAmountIn), maxSqrtPriceMinus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03QuoteExactInputSingleOneForZeroMatchesSwap: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")
	s.stateBlock = fmt.Sprintf("0x%x", receipt.BlockNumber)

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	swaps, err := ExtractSwap(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(new(big.Int).Neg(s.SwapAmountIn).Cmp(swaps[0].Amount1))
	s.Require().Zero(quote.AmountOut.Cmp(swaps[0].Amount0), "quoted amountOut must equal the swap's actual output to the wei")

	s.T().Logf("Quoter %s Test03QuoteExactInputSingleOneForZeroMatchesSwap Result", s.quoterContract)
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tQuoted AmountOut: %s (gas estimate %s)", quote.AmountOut, quote.GasEstimate)
	s.T().Logf("\tSwap Amount0: %s", swaps[0].Amount0)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *QuoterSuite) Test04QuoteExactOutputSingleMatchesSwap() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var quote *QuoteExactOutputResult
	var err error
	s.Require().Eventually(func() bool {
		quote, err = s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountOut, nil)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
	s.Require().Positive(quote.AmountIn.Cmp(s.SwapAmountOut))
	s.Require().Positive(quote.GasEstimate.Sign())

	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{true, s.SwapAmountOut, big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04QuoteExactOutputSingleMatchesSwap: sent PoolSwapTest.swap (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")
	s.stateBlock = fmt.Sprintf("0x%x", receipt.BlockNumber)

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	swaps, err := ExtractSwap(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(s.SwapAmountOut.Cmp(swaps[0].Amount1))
	s.Require().Zero(new(big.Int).Neg(quote.AmountIn).Cmp(swaps[0].Amount0), "quoted amountIn must equal the swap's actual input to the wei")

	s.T().Logf("Quoter %s Test04QuoteExactOutputSingleMatchesSwap Result", s.quoterContract)
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tQuoted AmountIn: %s (gas estimate %s)", quote.AmountIn, quote.GasEstimate)
	s.T().Logf("\tSwap Amount0: %s", swaps[0].Amount0)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *QuoterSuite) Test05QuoteExactInputMultiHopMatchesChainedSingles() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)
	path := []*PathKey{
		NewPathKey(s.tokens[1], s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks, nil),
		NewPathKey(s.tokens[2], s.keyBC.Fee, s.keyBC.TickSpacing, s.keyBC.Hooks, nil),
	}

	var multi *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		multi, err = s.quoter.CallQuoteExactInput(s.T().Context(), atBlock, NewQuoteExactInputRequest(s.quoterContract, NewQuoteExactParams(s.tokens[0], path, s.SwapAmountIn)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
	s.Require().Positive(multi.AmountOut.Sign())
	s.Require().Positive(multi.GasEstimate.Sign())

	hop1, err := s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil)))
	s.Require().NoError(err)
	hop2, err := s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, true, hop1.AmountOut, nil)))
	s.Require().NoError(err)

	s.Require().Zero(hop2.AmountOut.Cmp(multi.AmountOut), "multi-hop quote must equal two chained single-hop quotes")
	s.Require().Negative(multi.AmountOut.Cmp(hop1.AmountOut))
	s.Require().Positive(multi.GasEstimate.Cmp(hop1.GasEstimate))

	s.T().Logf("Quoter %s Test05QuoteExactInputMultiHopMatchesChainedSingles Result", s.quoterContract)
	s.T().Logf("\tBlock: %s", s.stateBlock)
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tHop1 AmountOut: %s", hop1.AmountOut)
	s.T().Logf("\tHop2 AmountOut: %s", hop2.AmountOut)
	s.T().Logf("\tMulti AmountOut: %s (gas estimate %s)", multi.AmountOut, multi.GasEstimate)
}

func (s *QuoterSuite) Test06QuoteExactOutputMultiHopMatchesChainedSingles() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)
	path := []*PathKey{
		NewPathKey(s.tokens[0], s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks, nil),
		NewPathKey(s.tokens[1], s.keyBC.Fee, s.keyBC.TickSpacing, s.keyBC.Hooks, nil),
	}

	var multi *QuoteExactOutputResult
	var err error
	s.Require().Eventually(func() bool {
		multi, err = s.quoter.CallQuoteExactOutput(s.T().Context(), atBlock, NewQuoteExactOutputRequest(s.quoterContract, NewQuoteExactParams(s.tokens[2], path, s.SwapAmountOut)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
	s.Require().Positive(multi.AmountIn.Sign())
	s.Require().Positive(multi.GasEstimate.Sign())

	hop2, err := s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, true, s.SwapAmountOut, nil)))
	s.Require().NoError(err)
	hop1, err := s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, hop2.AmountIn, nil)))
	s.Require().NoError(err)

	s.Require().Zero(hop1.AmountIn.Cmp(multi.AmountIn), "multi-hop quote must equal two chained single-hop quotes")
	s.Require().Positive(multi.AmountIn.Cmp(hop2.AmountIn))

	s.T().Logf("Quoter %s Test06QuoteExactOutputMultiHopMatchesChainedSingles Result", s.quoterContract)
	s.T().Logf("\tBlock: %s", s.stateBlock)
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tHop2 AmountIn: %s", hop2.AmountIn)
	s.T().Logf("\tHop1 AmountIn: %s", hop1.AmountIn)
	s.T().Logf("\tMulti AmountIn: %s (gas estimate %s)", multi.AmountIn, multi.GasEstimate)
}

func (s *QuoterSuite) Test07ExactInputAndExactOutputAreConsistent() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	for _, zeroForOne := range []bool{true, false} {
		var in *QuoteExactInputResult
		var err error
		s.Require().Eventually(func() bool {
			in, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, zeroForOne, s.SwapAmountIn, nil)))
			return err == nil
		}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)

		out, err := s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, zeroForOne, in.AmountOut, nil)))
		s.Require().NoError(err)
		s.Require().LessOrEqual(out.AmountIn.Cmp(s.SwapAmountIn), 0, "input needed for the quoted output must not exceed the original input")

		back, err := s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, zeroForOne, out.AmountIn, nil)))
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(back.AmountOut.Cmp(in.AmountOut), 0, "the exact-output quote's input must buy at least the requested output")

		s.T().Logf("Quoter %s Test07ExactInputAndExactOutputAreConsistent Result (zeroForOne=%t)", s.quoterContract, zeroForOne)
		s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
		s.T().Logf("\tExactIn AmountOut: %s", in.AmountOut)
		s.T().Logf("\tExactOut AmountIn: %s", out.AmountIn)
		s.T().Logf("\tRe-quoted AmountOut: %s", back.AmountOut)
	}
}

func (s *QuoterSuite) Test08LargerInputHasMorePriceImpact() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)
	amounts := []*big.Int{
		new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(19), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil),
	}
	outs := make([]*big.Int, len(amounts))
	for i, amountIn := range amounts {
		var quote *QuoteExactInputResult
		var err error
		s.Require().Eventually(func() bool {
			quote, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyBC, true, amountIn, nil)))
			return err == nil
		}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)
		outs[i] = quote.AmountOut
	}

	for i := 1; i < len(amounts); i++ {
		s.Require().Positive(outs[i].Cmp(outs[i-1]), "amountOut must grow with amountIn")
		lhs := new(big.Int).Mul(outs[i], amounts[i-1])
		rhs := new(big.Int).Mul(outs[i-1], amounts[i])
		s.Require().Negative(lhs.Cmp(rhs), "amountOut/amountIn must shrink as amountIn grows")
	}

	s.T().Logf("Quoter %s Test08LargerInputHasMorePriceImpact Result", s.quoterContract)
	for i := range amounts {
		s.T().Logf("\tAmountIn: %s -> AmountOut: %s", amounts[i], outs[i])
	}
}

func (s *QuoterSuite) Test09EmptyPathReturnsExactAmount() {
	in, err := s.quoter.CallQuoteExactInput(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewQuoteExactInputRequest(s.quoterContract, NewQuoteExactParams(s.tokens[0], []*PathKey{}, s.SwapAmountIn)))
	s.Require().NoError(err)
	s.Require().Zero(s.SwapAmountIn.Cmp(in.AmountOut))

	out, err := s.quoter.CallQuoteExactOutput(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewQuoteExactOutputRequest(s.quoterContract, NewQuoteExactParams(s.tokens[0], nil, s.SwapAmountOut)))
	s.Require().NoError(err)
	s.Require().Zero(s.SwapAmountOut.Cmp(out.AmountIn))

	s.T().Logf("Quoter %s Test09EmptyPathReturnsExactAmount Result", s.quoterContract)
	s.T().Logf("\tExactInput AmountOut: %s (gas estimate %s)", in.AmountOut, in.GasEstimate)
	s.T().Logf("\tExactOutput AmountIn: %s (gas estimate %s)", out.AmountIn, out.GasEstimate)
}

func (s *QuoterSuite) Test10PoolNotInitialized() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)
	key := NewPoolKey(s.tokens[0], s.tokens[2], big.NewInt(500), big.NewInt(10), types.ZeroAddress())

	var single *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		single, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(key, true, s.SwapAmountIn, nil)))
		return errors.Is(err, ErrUnexpectedRevertBytes)
	}, 20*time.Second, s.ReceiptPollInterval, "no quoter revert at block %s", s.stateBlock)
	s.Require().Nil(single)
	s.Require().ErrorIs(err, ErrPoolNotInitialized)
	s.Require().NotErrorIs(err, ErrUnknown)
	singleErr := err

	path := []*PathKey{NewPathKey(s.tokens[2], key.Fee, key.TickSpacing, key.Hooks, nil)}
	multi, err := s.quoter.CallQuoteExactInput(s.T().Context(), atBlock, NewQuoteExactInputRequest(s.quoterContract, NewQuoteExactParams(s.tokens[0], path, s.SwapAmountIn)))
	s.Require().Nil(multi)
	s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
	s.Require().ErrorIs(err, ErrPoolNotInitialized)

	s.T().Logf("Quoter %s Test10PoolNotInitialized Result", s.quoterContract)
	s.T().Logf("\tSingle: %v", singleErr)
	s.T().Logf("\tMulti: %v", err)
}

func (s *QuoterSuite) Test11NotEnoughLiquidity() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	for _, zeroForOne := range []bool{true, false} {
		var quote *QuoteExactOutputResult
		var err error
		s.Require().Eventually(func() bool {
			quote, err = s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAC, zeroForOne, s.ExcessiveAmountOut, nil)))
			return errors.Is(err, ErrUnexpectedRevertBytes)
		}, 20*time.Second, s.ReceiptPollInterval, "no quoter revert at block %s", s.stateBlock)
		s.Require().Nil(quote)
		s.Require().ErrorIs(err, ErrNotEnoughLiquidity)
		s.Require().NotErrorIs(err, ErrSafeCastOverflow)
		s.Require().NotErrorIs(err, ErrUnknown)

		s.T().Logf("Quoter %s Test11NotEnoughLiquidity Result (zeroForOne=%t)", s.quoterContract, zeroForOne)
		s.T().Logf("\tExactAmount: %s", s.ExcessiveAmountOut)
		s.T().Logf("\tError: %v", err)
	}

	quote, err := s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAC, true, s.SwapAmountOut, nil)))
	s.Require().NoError(err)
	s.Require().Positive(quote.AmountIn.Cmp(s.SwapAmountOut))

	s.T().Logf("\tIn-range ExactAmount: %s -> AmountIn: %s", s.SwapAmountOut, quote.AmountIn)
}

func (s *QuoterSuite) Test12SafeCastOverflow() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var quote *QuoteExactOutputResult
	var err error
	s.Require().Eventually(func() bool {
		quote, err = s.quoter.CallQuoteExactOutputSingle(s.T().Context(), atBlock, NewQuoteExactOutputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.ExcessiveAmountOut, nil)))
		return errors.Is(err, ErrUnexpectedRevertBytes)
	}, 20*time.Second, s.ReceiptPollInterval, "no quoter revert at block %s", s.stateBlock)
	s.Require().Nil(quote)
	s.Require().ErrorIs(err, ErrSafeCastOverflow)
	s.Require().NotErrorIs(err, ErrNotEnoughLiquidity)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("Quoter %s Test12SafeCastOverflow Result", s.quoterContract)
	s.T().Logf("\tExactAmount: %s", s.ExcessiveAmountOut)
	s.T().Logf("\tError: %v", err)
}

func (s *QuoterSuite) Test13SwapAmountCannotBeZero() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var quote *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		quote, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, big.NewInt(0), nil)))
		return errors.Is(err, ErrUnexpectedRevertBytes)
	}, 20*time.Second, s.ReceiptPollInterval, "no quoter revert at block %s", s.stateBlock)
	s.Require().Nil(quote)
	s.Require().ErrorIs(err, ErrSwapAmountCannotBeZero)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("Quoter %s Test13SwapAmountCannotBeZero Result", s.quoterContract)
	s.T().Logf("\tError: %v", err)
}

func (s *QuoterSuite) Test14QuotingDoesNotChangeState() {
	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(s.stateBlock)

	var first *QuoteExactInputResult
	var err error
	s.Require().Eventually(func() bool {
		first, err = s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil)))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "quote not available at block %s", s.stateBlock)

	second, err := s.quoter.CallQuoteExactInputSingle(s.T().Context(), atBlock, NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil)))
	s.Require().NoError(err)
	s.Require().Zero(first.AmountOut.Cmp(second.AmountOut))

	withoutFrom, err := s.quoter.CallQuoteExactInputSingle(s.T().Context(), bind.NewCallOptions().WithBlock(s.stateBlock), NewQuoteExactInputSingleRequest(s.quoterContract, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil)))
	s.Require().NoError(err)
	s.Require().Zero(first.AmountOut.Cmp(withoutFrom.AmountOut))

	msgSender, err := s.quoter.MsgSender(s.T().Context(), bind.NewCallOptions().WithBlock(s.stateBlock), NewMsgSenderRequest(s.quoterContract))
	s.Require().NoError(err)
	s.Require().True(msgSender.Equal(types.ZeroAddress()))

	s.T().Logf("Quoter %s Test14QuotingDoesNotChangeState Result", s.quoterContract)
	s.T().Logf("\tBlock: %s", s.stateBlock)
	s.T().Logf("\tAmountOut (from / again / no from): %s / %s / %s", first.AmountOut, second.AmountOut, withoutFrom.AmountOut)
	s.T().Logf("\tMsgSender after quotes: %s", msgSender)
}
