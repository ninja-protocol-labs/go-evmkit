package v4

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
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type QuoterSuite struct {
	suite.Suite

	Enable                 bool
	RPCURL                 string
	PrivateKeyHex          string
	PoolManagerHex         string
	PositionManagerHex     string
	Permit2Hex             string
	QuoterHex              string
	PoolSwapTestHex        string
	MulticallHex           string
	NativeSymbol           string
	Interval               int
	Timeout                int
	ByteCode               string
	Constructor            *abi.Function
	Salt                   *types.Hash
	MintValue              *big.Int
	Permit2ApproveAmount   *big.Int
	Permit2ApproveDuration *big.Int
	MintLiquidity          *big.Int
	MintAmountMax          *big.Int
	SwapAmountIn           *big.Int
	SwapAmountOut          *big.Int
	ExcessiveAmountOut     *big.Int

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	deployer *deploy.NicksDeployer
	signer   core.Signer
	from     *types.Address

	quoter                  *Quoter
	positionManager         *PositionManager
	erc20                   *erc20.ERC20
	permit2                 *permit2.Permit2
	poolManagerContract     *types.Address
	positionManagerContract *types.Address
	permit2Contract         *types.Address
	quoterContract          *types.Address
	poolSwapTest            *types.Address
	multicall               *types.Address

	tokens [3]*types.Address
	keyAB  *PoolKey
	keyBC  *PoolKey
	keyAC  *PoolKey

	positionTokenIDs []*big.Int
	positionKeys     []*PoolKey

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestQuoterSuite(t *testing.T) {
	suite.Run(t, &QuoterSuite{
		Enable:                 false,
		RPCURL:                 "",
		PrivateKeyHex:          "",
		PoolManagerHex:         "",
		PositionManagerHex:     "",
		Permit2Hex:             "",
		QuoterHex:              "",
		PoolSwapTestHex:        "0xf3A39C86dbd13C45365E57FB90fe413371F65AF8",
		MulticallHex:           multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:           "",
		Interval:               1500,
		Timeout:                60,
		ByteCode:               erc20ByteCodeHex,
		Constructor:            abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                   core.Keccak256([]byte("go-evmkit-v4-quoter")),
		MintValue:              new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		Permit2ApproveAmount:   new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)),
		Permit2ApproveDuration: big.NewInt(30 * 24 * 60 * 60),
		MintLiquidity:          new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil),
		MintAmountMax:          new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
		SwapAmountIn:           new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		SwapAmountOut:          new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		ExcessiveAmountOut:     new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
	})
}

func (s *QuoterSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 15*time.Minute)

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

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [3]string{"evmkit v4 quoter test token A", "evmkit v4 quoter test token B", "evmkit v4 quoter test token C"}
	symbols := [3]string{"EVQA", "EVQB", "EVQC"}

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(deploy.DefaultNicksMethodFactory, deployment))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(deploy.DefaultNicksMethodFactory, deployment))
		s.Require().NoError(err)
		s.tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
		calldata, err := tokenMintFn.EncodeCall(s.from, s.MintValue)
		s.Require().NoError(err)

		mintUnsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
		s.Require().NoError(err)
		s.Require().NoError(mintUnsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintUnsigned.MaxFee())
		s.mu.Unlock()

		mintRaw, err := mintUnsigned.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, mintRaw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		for _, spender := range []*types.Address{s.permit2Contract, s.poolSwapTest} {
			approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(token, spender, s.MintValue))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, approveTx.MaxFee())
			s.mu.Unlock()

			approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "erc20 approve transaction reverted")
		}

		expiration := big.NewInt(time.Now().Add(time.Duration(s.Permit2ApproveDuration.Int64()) * time.Second).Unix())
		permit2ApproveTx, permit2ApproveHash, err := s.permit2.Approve(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), permit2.NewApproveRequest(s.permit2Contract, token, s.positionManagerContract, s.Permit2ApproveAmount, expiration))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permit2ApproveTx.MaxFee())
		s.mu.Unlock()

		permit2ApproveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permit2ApproveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(permit2ApproveReceipt.Status, "permit2 approve transaction reverted")

		s.T().Logf("SetupSuite: deployed, minted, and approved token %s %s (tx %s)", symbols[i], token, hash)
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
		initUnsigned, err := s.positionManager.CreateInitializePool(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, key, new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(0))
		s.Require().NoError(err)
		s.Require().NoError(initUnsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, initUnsigned.MaxFee())
		s.mu.Unlock()

		initHash, err := s.positionManager.InitializePool(s.ctx, initUnsigned)
		s.Require().NoError(err)

		initReceipt, err := tx.WaitForReceipt(s.ctx, s.client, initHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(initReceipt.Status, "initializePool transaction reverted")

		mintParams := NewMintPositionParams(key, pool.tickLower, pool.tickUpper, s.MintLiquidity, s.MintAmountMax, s.MintAmountMax, s.from, nil)
		mintData, err := EncodeActionMintPosition(mintParams)
		s.Require().NoError(err)
		settleData, err := EncodeActionSettlePair(key.Currency0, key.Currency1)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		mintUnsigned, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, []byte{ActionMintPosition, ActionSettlePair}, [][]byte{mintData, settleData}, deadline, big.NewInt(0))
		s.Require().NoError(err)
		s.Require().NoError(mintUnsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintUnsigned.MaxFee())
		s.mu.Unlock()

		mintRaw, err := mintUnsigned.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, mintRaw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
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

		s.T().Logf("SetupSuite: initialized pool %s/%s and minted liquidity %s in ticks [%s, %s] as tokenID %s (init tx %s, mint tx %s)", key.Currency0, key.Currency1, s.MintLiquidity, pool.tickLower, pool.tickUpper, transfers[0].TokenId, initHash, mintHash)
	}
}

func (s *QuoterSuite) TearDownSuite() {
	for i, tokenID := range s.positionTokenIDs {
		key := s.positionKeys[i]

		burnData, err := EncodeActionBurnPosition(tokenID, big.NewInt(0), big.NewInt(0), nil)
		s.Require().NoError(err)
		takeData, err := EncodeActionTakePair(key.Currency0, key.Currency1, s.from)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		unsignedBurnPosition, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, []byte{ActionBurnPosition, ActionTakePair}, [][]byte{burnData, takeData}, deadline, big.NewInt(0))
		s.Require().NoError(err)
		s.Require().NoError(unsignedBurnPosition.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedBurnPosition.MaxFee())
		s.mu.Unlock()

		raw, err := unsignedBurnPosition.EncodeRLP()
		s.Require().NoError(err)
		burnPositionHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		burnPositionReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnPositionHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(burnPositionReceipt.Status, "modifyLiquidities (burn position) transaction reverted")

		var positionManagerLogs []tx.Log
		for _, log := range burnPositionReceipt.Logs {
			if log.Address.Equal(s.positionManagerContract) {
				positionManagerLogs = append(positionManagerLogs, log)
			}
		}
		transfers, err := erc721.ExtractTransfers(positionManagerLogs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].To.Equal(types.ZeroAddress()))
		s.Require().Zero(tokenID.Cmp(transfers[0].TokenId))

		s.T().Logf("TearDownSuite: burned position tokenID %s in pool %s/%s and took its liquidity back (tx %s)", tokenID, key.Currency0, key.Currency1, burnPositionHash)
	}

	if s.tokens[0] != nil && s.tokens[1] != nil && s.tokens[2] != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, s.tokens[:], false))
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

		s.T().Logf("QuoterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *QuoterSuite) Test01Immutables() {
	poolManager, err := s.quoter.PoolManager(s.ctx, s.quoterContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolManager.Equal(s.poolManagerContract))

	msgSender, err := s.quoter.MsgSender(s.ctx, s.quoterContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(msgSender.Equal(types.ZeroAddress()))

	s.T().Logf("QuoterSuite Test01Immutables Result")
	s.T().Logf("\tPoolManager: %s", poolManager)
	s.T().Logf("\tMsgSender: %s", msgSender)
}

func (s *QuoterSuite) Test02QuoteExactInputSingleMatchesSwap() {
	quote, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(quote.AmountOut.Sign())
	s.Require().Negative(quote.AmountOut.Cmp(s.SwapAmountIn))
	s.Require().Positive(quote.GasEstimate.Sign())

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

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

	s.T().Logf("QuoterSuite Test02QuoteExactInputSingleMatchesSwap Result")
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tQuoted AmountOut: %s", quote.AmountOut)
	s.T().Logf("\tSwap Amount1: %s", swaps[0].Amount1)
	s.T().Logf("\tGasEstimate: %s", quote.GasEstimate)
	s.T().Logf("\tTx: %s", hash)
}

func (s *QuoterSuite) Test03QuoteExactInputSingleOneForZeroMatchesSwap() {
	quote, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, false, s.SwapAmountIn, nil), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(quote.AmountOut.Sign())
	s.Require().Positive(quote.GasEstimate.Sign())

	maxSqrtPriceMinus1, ok := new(big.Int).SetString("1461446703485210103287273052203988822378723970341", 10)
	s.Require().True(ok)

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{false, new(big.Int).Neg(s.SwapAmountIn), maxSqrtPriceMinus1},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

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

	s.T().Logf("QuoterSuite Test03QuoteExactInputSingleOneForZeroMatchesSwap Result")
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tQuoted AmountOut: %s", quote.AmountOut)
	s.T().Logf("\tSwap Amount0: %s", swaps[0].Amount0)
	s.T().Logf("\tGasEstimate: %s", quote.GasEstimate)
	s.T().Logf("\tTx: %s", hash)
}

func (s *QuoterSuite) Test04QuoteExactOutputSingleMatchesSwap() {
	quote, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountOut, nil), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(quote.AmountIn.Cmp(s.SwapAmountOut))
	s.Require().Positive(quote.GasEstimate.Sign())

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.keyAB.Currency0, s.keyAB.Currency1, s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks},
		[]any{true, s.SwapAmountOut, big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	unsigned, err := tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	raw, err := unsigned.EncodeRLP()
	s.Require().NoError(err)
	hash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

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

	s.T().Logf("QuoterSuite Test04QuoteExactOutputSingleMatchesSwap Result")
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tQuoted AmountIn: %s", quote.AmountIn)
	s.T().Logf("\tSwap Amount0: %s", swaps[0].Amount0)
	s.T().Logf("\tGasEstimate: %s", quote.GasEstimate)
	s.T().Logf("\tTx: %s", hash)
}

func (s *QuoterSuite) Test05QuoteExactInputMultiHopMatchesChainedSingles() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	path := []*PathKey{
		NewPathKey(s.tokens[1], s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks, nil),
		NewPathKey(s.tokens[2], s.keyBC.Fee, s.keyBC.TickSpacing, s.keyBC.Hooks, nil),
	}
	multi, err := s.quoter.CallQuoteExactInput(s.ctx, s.quoterContract, s.from, NewQuoteExactParams(s.tokens[0], path, s.SwapAmountIn), blockHex)
	s.Require().NoError(err)
	s.Require().Positive(multi.AmountOut.Sign())
	s.Require().Positive(multi.GasEstimate.Sign())

	hop1, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil), blockHex)
	s.Require().NoError(err)
	hop2, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, true, hop1.AmountOut, nil), blockHex)
	s.Require().NoError(err)

	s.Require().Zero(hop2.AmountOut.Cmp(multi.AmountOut), "multi-hop quote must equal two chained single-hop quotes")
	s.Require().Negative(multi.AmountOut.Cmp(hop1.AmountOut))
	s.Require().Positive(multi.GasEstimate.Cmp(hop1.GasEstimate))

	s.T().Logf("QuoterSuite Test05QuoteExactInputMultiHopMatchesChainedSingles Result")
	s.T().Logf("\tBlock: %s", blockHex)
	s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
	s.T().Logf("\tHop1 AmountOut: %s", hop1.AmountOut)
	s.T().Logf("\tHop2 AmountOut: %s", hop2.AmountOut)
	s.T().Logf("\tMulti AmountOut: %s", multi.AmountOut)
	s.T().Logf("\tMulti GasEstimate: %s", multi.GasEstimate)
}

func (s *QuoterSuite) Test06QuoteExactOutputMultiHopMatchesChainedSingles() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	path := []*PathKey{
		NewPathKey(s.tokens[0], s.keyAB.Fee, s.keyAB.TickSpacing, s.keyAB.Hooks, nil),
		NewPathKey(s.tokens[1], s.keyBC.Fee, s.keyBC.TickSpacing, s.keyBC.Hooks, nil),
	}
	multi, err := s.quoter.CallQuoteExactOutput(s.ctx, s.quoterContract, s.from, NewQuoteExactParams(s.tokens[2], path, s.SwapAmountOut), blockHex)
	s.Require().NoError(err)
	s.Require().Positive(multi.AmountIn.Sign())
	s.Require().Positive(multi.GasEstimate.Sign())

	hop2, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, true, s.SwapAmountOut, nil), blockHex)
	s.Require().NoError(err)
	hop1, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, hop2.AmountIn, nil), blockHex)
	s.Require().NoError(err)

	s.Require().Zero(hop1.AmountIn.Cmp(multi.AmountIn), "multi-hop quote must equal two chained single-hop quotes")
	s.Require().Positive(multi.AmountIn.Cmp(hop2.AmountIn))

	s.T().Logf("QuoterSuite Test06QuoteExactOutputMultiHopMatchesChainedSingles Result")
	s.T().Logf("\tBlock: %s", blockHex)
	s.T().Logf("\tAmountOut: %s", s.SwapAmountOut)
	s.T().Logf("\tHop2 AmountIn: %s", hop2.AmountIn)
	s.T().Logf("\tHop1 AmountIn: %s", hop1.AmountIn)
	s.T().Logf("\tMulti AmountIn: %s", multi.AmountIn)
	s.T().Logf("\tMulti GasEstimate: %s", multi.GasEstimate)
}

func (s *QuoterSuite) Test07ExactInputAndExactOutputAreConsistent() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	for _, zeroForOne := range []bool{true, false} {
		in, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, zeroForOne, s.SwapAmountIn, nil), blockHex)
		s.Require().NoError(err)

		out, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, zeroForOne, in.AmountOut, nil), blockHex)
		s.Require().NoError(err)
		s.Require().LessOrEqual(out.AmountIn.Cmp(s.SwapAmountIn), 0, "input needed for the quoted output must not exceed the original input")

		back, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, zeroForOne, out.AmountIn, nil), blockHex)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(back.AmountOut.Cmp(in.AmountOut), 0, "the exact-output quote's input must buy at least the requested output")

		s.T().Logf("QuoterSuite Test07ExactInputAndExactOutputAreConsistent Result (zeroForOne=%t)", zeroForOne)
		s.T().Logf("\tAmountIn: %s", s.SwapAmountIn)
		s.T().Logf("\tExactIn AmountOut: %s", in.AmountOut)
		s.T().Logf("\tExactOut AmountIn: %s", out.AmountIn)
		s.T().Logf("\tRe-quoted AmountOut: %s", back.AmountOut)
	}
}

func (s *QuoterSuite) Test08LargerInputHasMorePriceImpact() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	amounts := []*big.Int{
		new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(19), nil),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil),
	}
	outs := make([]*big.Int, len(amounts))
	for i, amountIn := range amounts {
		quote, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyBC, true, amountIn, nil), blockHex)
		s.Require().NoError(err)
		outs[i] = quote.AmountOut
	}

	for i := 1; i < len(amounts); i++ {
		s.Require().Positive(outs[i].Cmp(outs[i-1]), "amountOut must grow with amountIn")
		lhs := new(big.Int).Mul(outs[i], amounts[i-1])
		rhs := new(big.Int).Mul(outs[i-1], amounts[i])
		s.Require().Negative(lhs.Cmp(rhs), "amountOut/amountIn must shrink as amountIn grows")
	}

	s.T().Logf("QuoterSuite Test08LargerInputHasMorePriceImpact Result")
	for i := range amounts {
		s.T().Logf("\tAmountIn: %s -> AmountOut: %s", amounts[i], outs[i])
	}
}

func (s *QuoterSuite) Test09EmptyPathReturnsExactAmount() {
	in, err := s.quoter.CallQuoteExactInput(s.ctx, s.quoterContract, s.from, NewQuoteExactParams(s.tokens[0], []*PathKey{}, s.SwapAmountIn), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.SwapAmountIn.Cmp(in.AmountOut))

	out, err := s.quoter.CallQuoteExactOutput(s.ctx, s.quoterContract, s.from, NewQuoteExactParams(s.tokens[0], nil, s.SwapAmountOut), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.SwapAmountOut.Cmp(out.AmountIn))

	s.T().Logf("QuoterSuite Test09EmptyPathReturnsExactAmount Result")
	s.T().Logf("\tExactInput AmountOut: %s", in.AmountOut)
	s.T().Logf("\tExactOutput AmountIn: %s", out.AmountIn)
}

func (s *QuoterSuite) Test10PoolNotInitialized() {
	key := NewPoolKey(s.tokens[0], s.tokens[2], big.NewInt(500), big.NewInt(10), types.ZeroAddress())

	_, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(key, true, s.SwapAmountIn, nil), rpc.BlockTagLatest)
	s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
	s.Require().ErrorIs(err, ErrPoolNotInitialized)
	s.Require().NotErrorIs(err, ErrUnknown)

	path := []*PathKey{NewPathKey(s.tokens[2], key.Fee, key.TickSpacing, key.Hooks, nil)}
	_, err = s.quoter.CallQuoteExactInput(s.ctx, s.quoterContract, s.from, NewQuoteExactParams(s.tokens[0], path, s.SwapAmountIn), rpc.BlockTagLatest)
	s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
	s.Require().ErrorIs(err, ErrPoolNotInitialized)

	s.T().Logf("QuoterSuite Test10PoolNotInitialized Result")
	s.T().Logf("\tError: %v", err)
}

func (s *QuoterSuite) Test11NotEnoughLiquidity() {
	for _, zeroForOne := range []bool{true, false} {
		_, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAC, zeroForOne, s.ExcessiveAmountOut, nil), rpc.BlockTagLatest)
		s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
		s.Require().ErrorIs(err, ErrNotEnoughLiquidity)
		s.Require().NotErrorIs(err, ErrSafeCastOverflow)
		s.Require().NotErrorIs(err, ErrUnknown)

		s.T().Logf("QuoterSuite Test11NotEnoughLiquidity Result (zeroForOne=%t)", zeroForOne)
		s.T().Logf("\tExactAmount: %s", s.ExcessiveAmountOut)
		s.T().Logf("\tError: %v", err)
	}

	quote, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAC, true, s.SwapAmountOut, nil), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(quote.AmountIn.Cmp(s.SwapAmountOut))

	s.T().Logf("\tIn-range ExactAmount: %s -> AmountIn: %s", s.SwapAmountOut, quote.AmountIn)
}

func (s *QuoterSuite) Test12SafeCastOverflow() {
	_, err := s.quoter.CallQuoteExactOutputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.ExcessiveAmountOut, nil), rpc.BlockTagLatest)
	s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
	s.Require().ErrorIs(err, ErrSafeCastOverflow)
	s.Require().NotErrorIs(err, ErrNotEnoughLiquidity)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("QuoterSuite Test12SafeCastOverflow Result")
	s.T().Logf("\tExactAmount: %s", s.ExcessiveAmountOut)
	s.T().Logf("\tError: %v", err)
}

func (s *QuoterSuite) Test13SwapAmountCannotBeZero() {
	_, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, big.NewInt(0), nil), rpc.BlockTagLatest)
	s.Require().ErrorIs(err, ErrUnexpectedRevertBytes)
	s.Require().ErrorIs(err, ErrSwapAmountCannotBeZero)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("QuoterSuite Test13SwapAmountCannotBeZero Result")
	s.T().Logf("\tError: %v", err)
}

func (s *QuoterSuite) Test14QuotingDoesNotChangeState() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	first, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil), blockHex)
	s.Require().NoError(err)
	second, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, s.from, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil), blockHex)
	s.Require().NoError(err)
	s.Require().Zero(first.AmountOut.Cmp(second.AmountOut))

	nilSender, err := s.quoter.CallQuoteExactInputSingle(s.ctx, s.quoterContract, nil, NewQuoteExactSingleParams(s.keyAB, true, s.SwapAmountIn, nil), blockHex)
	s.Require().NoError(err)
	s.Require().Zero(first.AmountOut.Cmp(nilSender.AmountOut))

	msgSender, err := s.quoter.MsgSender(s.ctx, s.quoterContract, blockHex)
	s.Require().NoError(err)
	s.Require().True(msgSender.Equal(types.ZeroAddress()))

	s.T().Logf("QuoterSuite Test14QuotingDoesNotChangeState Result")
	s.T().Logf("\tBlock: %s", blockHex)
	s.T().Logf("\tAmountOut: %s", first.AmountOut)
}
