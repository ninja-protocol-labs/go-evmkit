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

type StateViewSuite struct {
	suite.Suite

	Enable                 bool
	RPCURL                 string
	PrivateKeyHex          string
	PoolManagerHex         string
	PositionManagerHex     string
	Permit2Hex             string
	StateViewHex           string
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
	FullRangeLiquidity     *big.Int
	NarrowLiquidity        *big.Int
	MintAmountMax          *big.Int
	SwapAmountIn           *big.Int

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	deployer *deploy.NicksDeployer
	signer   core.Signer
	from     *types.Address

	stateView               *StateView
	positionManager         *PositionManager
	erc20                   *erc20.ERC20
	permit2                 *permit2.Permit2
	poolManagerContract     *types.Address
	positionManagerContract *types.Address
	permit2Contract         *types.Address
	stateViewContract       *types.Address
	poolSwapTest            *types.Address
	multicall               *types.Address

	tokens [2]*types.Address
	key    *PoolKey
	poolID PoolID

	fullRangeTokenID *big.Int
	narrowTokenID    *big.Int
	fullRangeLower   *big.Int
	fullRangeUpper   *big.Int
	narrowLower      *big.Int
	narrowUpper      *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestStateViewSuite(t *testing.T) {
	suite.Run(t, &StateViewSuite{
		Enable:                 false,
		RPCURL:                 "",
		PrivateKeyHex:          "",
		PoolManagerHex:         "",
		PositionManagerHex:     "",
		Permit2Hex:             "",
		StateViewHex:           "",
		PoolSwapTestHex:        "",
		MulticallHex:           multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:           "",
		Interval:               1500,
		Timeout:                60,
		ByteCode:               erc20ByteCodeHex,
		Constructor:            abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                   core.Keccak256([]byte("go-evmkit-v4-stateview")),
		MintValue:              new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		Permit2ApproveAmount:   new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)),
		Permit2ApproveDuration: big.NewInt(30 * 24 * 60 * 60),
		FullRangeLiquidity:     new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil),
		NarrowLiquidity:        new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
		MintAmountMax:          new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil),
		SwapAmountIn:           new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
	})
}

func (s *StateViewSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 15*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.stateView = NewStateView(s.client)
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
	s.stateViewContract, err = types.NewAddressFromHex(s.StateViewHex)
	s.Require().NoError(err)
	s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 stateview test token A", "evmkit v4 stateview test token B"}
	symbols := [2]string{"EVSA", "EVSB"}

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
			unsignedApprove, err := s.erc20.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), token, s.from, spender, s.MintValue)
			s.Require().NoError(err)
			s.Require().NoError(unsignedApprove.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
			s.mu.Unlock()

			approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
			s.Require().NoError(err)

			approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "erc20 approve transaction reverted")
		}

		expiration := big.NewInt(time.Now().Add(time.Duration(s.Permit2ApproveDuration.Int64()) * time.Second).Unix())
		permit2Params := permit2.NewApproveParams(s.permit2Contract, s.from, token, s.positionManagerContract, s.Permit2ApproveAmount, expiration)

		unsignedPermit2Approve, err := s.permit2.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), permit2Params)
		s.Require().NoError(err)
		s.Require().NoError(unsignedPermit2Approve.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedPermit2Approve.MaxFee())
		s.mu.Unlock()

		permit2ApproveHash, err := s.permit2.Approve(s.ctx, unsignedPermit2Approve)
		s.Require().NoError(err)

		permit2ApproveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permit2ApproveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(permit2ApproveReceipt.Status, "permit2 approve transaction reverted")

		s.T().Logf("SetupSuite: deployed, minted, and approved token %s %s (tx %s)", symbols[i], token, hash)
	}

	if s.tokens[0].Big().Cmp(s.tokens[1].Big()) > 0 {
		s.tokens[0], s.tokens[1] = s.tokens[1], s.tokens[0]
	}
	s.key = NewPoolKey(s.tokens[0], s.tokens[1], big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.poolID, err = s.key.PoolID()
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s poolId=%s", s.tokens[0], s.tokens[1], encoding.Hex.EncodePrefixed(s.poolID[:]))

	initUnsigned, err := s.positionManager.CreateInitializePool(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, s.key, new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(0))
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

	s.fullRangeLower, s.fullRangeUpper = big.NewInt(-887220), big.NewInt(887220)
	s.narrowLower, s.narrowUpper = big.NewInt(-120), big.NewInt(120)

	for i, position := range []struct {
		tickLower *big.Int
		tickUpper *big.Int
		liquidity *big.Int
	}{
		{s.fullRangeLower, s.fullRangeUpper, s.FullRangeLiquidity},
		{s.narrowLower, s.narrowUpper, s.NarrowLiquidity},
	} {
		mintParams := NewMintPositionParams(s.key, position.tickLower, position.tickUpper, position.liquidity, s.MintAmountMax, s.MintAmountMax, s.from, nil)
		mintData, err := EncodeActionMintPosition(mintParams)
		s.Require().NoError(err)
		settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
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
		if i == 0 {
			s.fullRangeTokenID = transfers[0].TokenId
		} else {
			s.narrowTokenID = transfers[0].TokenId
		}

		s.T().Logf("SetupSuite: minted liquidity %s in ticks [%s, %s] as tokenID %s (tx %s)", position.liquidity, position.tickLower, position.tickUpper, transfers[0].TokenId, mintHash)
	}
}

func (s *StateViewSuite) TearDownSuite() {
	for _, tokenID := range []*big.Int{s.fullRangeTokenID, s.narrowTokenID} {
		if tokenID == nil {
			continue
		}

		burnData, err := EncodeActionBurnPosition(tokenID, big.NewInt(0), big.NewInt(0), nil)
		s.Require().NoError(err)
		takeData, err := EncodeActionTakePair(s.key.Currency0, s.key.Currency1, s.from)
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

		s.T().Logf("TearDownSuite: burned position tokenID %s and took its liquidity back (tx %s)", tokenID, burnPositionHash)
	}

	if s.tokens[0] != nil && s.tokens[1] != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, s.tokens[:], false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
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

			s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", tb.Amount, tb.Token, s.from, burnHash)
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

		s.T().Logf("StateViewSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *StateViewSuite) Test01PoolManager() {
	poolManager, err := s.stateView.PoolManager(s.ctx, s.stateViewContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolManager.Equal(s.poolManagerContract))

	s.T().Logf("StateViewSuite Test01PoolManager Result")
	s.T().Logf("\tPoolManager: %s", poolManager)
}

func (s *StateViewSuite) Test02Slot0AfterInitialize() {
	slot0, err := s.stateView.GetSlot0(s.ctx, s.stateViewContract, s.poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Lsh(big.NewInt(1), 96).Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().Zero(slot0.ProtocolFee.Sign())
	s.Require().Zero(s.key.Fee.Cmp(slot0.LPFee))

	s.T().Logf("StateViewSuite Test02Slot0AfterInitialize Result")
	s.T().Logf("\tSqrtPriceX96: %s", slot0.SqrtPriceX96)
	s.T().Logf("\tTick: %s", slot0.Tick)
	s.T().Logf("\tProtocolFee: %s", slot0.ProtocolFee)
	s.T().Logf("\tLPFee: %s", slot0.LPFee)
}

func (s *StateViewSuite) Test03LiquidityIsSumOfInRangePositions() {
	liquidity, err := s.stateView.GetLiquidity(s.ctx, s.stateViewContract, s.poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	want := new(big.Int).Add(s.FullRangeLiquidity, s.NarrowLiquidity)
	s.Require().Zero(want.Cmp(liquidity))

	s.T().Logf("StateViewSuite Test03LiquidityIsSumOfInRangePositions Result")
	s.T().Logf("\tLiquidity: %s", liquidity)
}

func (s *StateViewSuite) Test04PositionsMatchPositionManager() {
	for _, position := range []struct {
		tokenID   *big.Int
		tickLower *big.Int
		tickUpper *big.Int
		liquidity *big.Int
	}{
		{s.fullRangeTokenID, s.fullRangeLower, s.fullRangeUpper, s.FullRangeLiquidity},
		{s.narrowTokenID, s.narrowLower, s.narrowUpper, s.NarrowLiquidity},
	} {
		var salt [32]byte
		position.tokenID.FillBytes(salt[:])

		byFields, err := s.stateView.GetPositionInfo(s.ctx, s.stateViewContract, s.poolID, s.positionManagerContract, position.tickLower, position.tickUpper, salt, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(position.liquidity.Cmp(byFields.Liquidity))
		s.Require().Zero(byFields.FeeGrowthInside0LastX128.Sign())
		s.Require().Zero(byFields.FeeGrowthInside1LastX128.Sign())

		positionID, err := CalculatePositionKey(s.positionManagerContract, position.tickLower, position.tickUpper, salt)
		s.Require().NoError(err)

		byID, err := s.stateView.GetPositionInfoByID(s.ctx, s.stateViewContract, s.poolID, positionID, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(byFields.Liquidity.Cmp(byID.Liquidity))
		s.Require().Zero(byFields.FeeGrowthInside0LastX128.Cmp(byID.FeeGrowthInside0LastX128))
		s.Require().Zero(byFields.FeeGrowthInside1LastX128.Cmp(byID.FeeGrowthInside1LastX128))

		liquidity, err := s.stateView.GetPositionLiquidity(s.ctx, s.stateViewContract, s.poolID, positionID, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(position.liquidity.Cmp(liquidity))

		pmLiquidity, err := s.positionManager.GetPositionLiquidity(s.ctx, s.positionManagerContract, position.tokenID, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(pmLiquidity.Cmp(liquidity), "StateView and PositionManager must agree on the position's liquidity")

		var wrongOwnerSalt [32]byte
		position.tokenID.FillBytes(wrongOwnerSalt[:])
		wrongOwner, err := s.stateView.GetPositionInfo(s.ctx, s.stateViewContract, s.poolID, s.from, position.tickLower, position.tickUpper, wrongOwnerSalt, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(wrongOwner.Liquidity.Sign(), "positions are owned by PositionManager, not the minting EOA")

		s.T().Logf("StateViewSuite Test04PositionsMatchPositionManager Result (tokenID %s)", position.tokenID)
		s.T().Logf("\tPositionID: %s", encoding.Hex.EncodePrefixed(positionID[:]))
		s.T().Logf("\tLiquidity: %s", liquidity)
	}
}

func (s *StateViewSuite) Test05TickInfoReflectsPositionBoundaries() {
	for _, boundary := range []struct {
		tick  *big.Int
		gross *big.Int
		net   *big.Int
	}{
		{s.fullRangeLower, s.FullRangeLiquidity, s.FullRangeLiquidity},
		{s.fullRangeUpper, s.FullRangeLiquidity, new(big.Int).Neg(s.FullRangeLiquidity)},
		{s.narrowLower, s.NarrowLiquidity, s.NarrowLiquidity},
		{s.narrowUpper, s.NarrowLiquidity, new(big.Int).Neg(s.NarrowLiquidity)},
	} {
		info, err := s.stateView.GetTickInfo(s.ctx, s.stateViewContract, s.poolID, boundary.tick, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(boundary.gross.Cmp(info.LiquidityGross), "tick %s liquidityGross", boundary.tick)
		s.Require().Zero(boundary.net.Cmp(info.LiquidityNet), "tick %s liquidityNet", boundary.tick)

		liquidity, err := s.stateView.GetTickLiquidity(s.ctx, s.stateViewContract, s.poolID, boundary.tick, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(info.LiquidityGross.Cmp(liquidity.LiquidityGross))
		s.Require().Zero(info.LiquidityNet.Cmp(liquidity.LiquidityNet))

		outside, err := s.stateView.GetTickFeeGrowthOutside(s.ctx, s.stateViewContract, s.poolID, boundary.tick, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Zero(info.FeeGrowthOutside0X128.Cmp(outside.FeeGrowthOutside0X128))
		s.Require().Zero(info.FeeGrowthOutside1X128.Cmp(outside.FeeGrowthOutside1X128))

		s.T().Logf("StateViewSuite Test05TickInfoReflectsPositionBoundaries Result (tick %s)", boundary.tick)
		s.T().Logf("\tLiquidityGross: %s", info.LiquidityGross)
		s.T().Logf("\tLiquidityNet: %s", info.LiquidityNet)
	}

	empty, err := s.stateView.GetTickInfo(s.ctx, s.stateViewContract, s.poolID, big.NewInt(60), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(empty.LiquidityGross.Sign())
	s.Require().Zero(empty.LiquidityNet.Sign())
}

func (s *StateViewSuite) Test06TickBitmapHasInitializedTicks() {
	for _, tick := range []*big.Int{s.fullRangeLower, s.fullRangeUpper, s.narrowLower, s.narrowUpper} {
		compressed := tick.Int64() / s.key.TickSpacing.Int64()
		word := int16(compressed >> 8)
		bit := uint(compressed & 0xff)

		bitmap, err := s.stateView.GetTickBitmap(s.ctx, s.stateViewContract, s.poolID, word, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Equal(uint(1), bitmap.Bit(int(bit)), "tick %s must be flagged in word %d bit %d", tick, word, bit)

		s.T().Logf("StateViewSuite Test06TickBitmapHasInitializedTicks Result (tick %s)", tick)
		s.T().Logf("\tWord: %d Bit: %d", word, bit)
		s.T().Logf("\tBitmap: 0x%s", bitmap.Text(16))
	}

	bitmap, err := s.stateView.GetTickBitmap(s.ctx, s.stateViewContract, s.poolID, 1, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(bitmap.Sign())
}

func (s *StateViewSuite) Test07SwapUpdatesSlot0AndFeeGrowth() {
	var blockBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockBeforeHex)))

	globalsBefore, err := s.stateView.GetFeeGrowthGlobals(s.ctx, s.stateViewContract, s.poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(globalsBefore.FeeGrowthGlobal0.Sign())
	s.Require().Zero(globalsBefore.FeeGrowthGlobal1.Sign())

	testSettingsTuple := abi.Tuple(abi.Bool, abi.Bool)
	poolSwapTestFn := abi.NewFunction("swap", abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, testSettingsTuple, abi.Bytes), abi.NewTypes(abi.Bytes32))
	calldata, err := poolSwapTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
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

	blockAfterHex := fmt.Sprintf("0x%x", receipt.BlockNumber)

	slot0, err := s.stateView.GetSlot0(s.ctx, s.stateViewContract, s.poolID, blockAfterHex)
	s.Require().NoError(err)
	s.Require().Zero(swaps[0].SqrtPriceX96.Cmp(slot0.SqrtPriceX96), "slot0 sqrtPrice must equal the Swap event's")
	s.Require().Zero(swaps[0].Tick.Cmp(slot0.Tick), "slot0 tick must equal the Swap event's")
	s.Require().Negative(slot0.Tick.Sign())
	s.Require().GreaterOrEqual(slot0.Tick.Cmp(s.narrowLower), 0, "the swap must stay inside the narrow range")

	liquidity, err := s.stateView.GetLiquidity(s.ctx, s.stateViewContract, s.poolID, blockAfterHex)
	s.Require().NoError(err)
	s.Require().Zero(swaps[0].Liquidity.Cmp(liquidity), "in-range liquidity must equal the Swap event's")

	globals, err := s.stateView.GetFeeGrowthGlobals(s.ctx, s.stateViewContract, s.poolID, blockAfterHex)
	s.Require().NoError(err)
	s.Require().Positive(globals.FeeGrowthGlobal0.Sign(), "zeroForOne swap fees accrue in token0")
	s.Require().Zero(globals.FeeGrowthGlobal1.Sign())

	before, err := s.stateView.GetSlot0(s.ctx, s.stateViewContract, s.poolID, blockBeforeHex)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Lsh(big.NewInt(1), 96).Cmp(before.SqrtPriceX96), "the pre-swap block must still read the initial price")

	s.T().Logf("StateViewSuite Test07SwapUpdatesSlot0AndFeeGrowth Result")
	s.T().Logf("\tSqrtPriceX96: %s", slot0.SqrtPriceX96)
	s.T().Logf("\tTick: %s", slot0.Tick)
	s.T().Logf("\tLiquidity: %s", liquidity)
	s.T().Logf("\tFeeGrowthGlobal0: %s", globals.FeeGrowthGlobal0)
	s.T().Logf("\tTx: %s", hash)
}

func (s *StateViewSuite) Test08FeeGrowthInsideMatchesGlobalsWhileInRange() {
	var blockHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHBlockNumber("blockNumber", &blockHex)))

	globals, err := s.stateView.GetFeeGrowthGlobals(s.ctx, s.stateViewContract, s.poolID, blockHex)
	s.Require().NoError(err)

	for _, r := range []struct {
		tickLower *big.Int
		tickUpper *big.Int
	}{
		{s.fullRangeLower, s.fullRangeUpper},
		{s.narrowLower, s.narrowUpper},
	} {
		inside, err := s.stateView.GetFeeGrowthInside(s.ctx, s.stateViewContract, s.poolID, r.tickLower, r.tickUpper, blockHex)
		s.Require().NoError(err)
		s.Require().Zero(globals.FeeGrowthGlobal0.Cmp(inside.FeeGrowthInside0X128), "range [%s, %s] has held the price since genesis, so inside == global", r.tickLower, r.tickUpper)
		s.Require().Zero(globals.FeeGrowthGlobal1.Cmp(inside.FeeGrowthInside1X128))

		s.T().Logf("StateViewSuite Test08FeeGrowthInsideMatchesGlobalsWhileInRange Result ([%s, %s])", r.tickLower, r.tickUpper)
		s.T().Logf("\tFeeGrowthInside0X128: %s", inside.FeeGrowthInside0X128)
	}

	outOfRange, err := s.stateView.GetFeeGrowthInside(s.ctx, s.stateViewContract, s.poolID, big.NewInt(600), big.NewInt(1200), blockHex)
	s.Require().NoError(err)
	s.Require().Zero(outOfRange.FeeGrowthInside0X128.Sign(), "a range the price never entered has no fee growth inside")
	s.Require().Zero(outOfRange.FeeGrowthInside1X128.Sign())

	var salt [32]byte
	s.narrowTokenID.FillBytes(salt[:])
	position, err := s.stateView.GetPositionInfo(s.ctx, s.stateViewContract, s.poolID, s.positionManagerContract, s.narrowLower, s.narrowUpper, salt, blockHex)
	s.Require().NoError(err)
	s.Require().Zero(position.FeeGrowthInside0LastX128.Sign(), "feeGrowthInsideLast only updates when the position is modified")
}

func (s *StateViewSuite) Test09UnknownPoolReadsAsZero() {
	unknown := NewPoolKey(s.tokens[0], s.tokens[1], big.NewInt(500), big.NewInt(10), types.ZeroAddress())
	poolID, err := unknown.PoolID()
	s.Require().NoError(err)

	slot0, err := s.stateView.GetSlot0(s.ctx, s.stateViewContract, poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(slot0.SqrtPriceX96.Sign())
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().Zero(slot0.LPFee.Sign())

	liquidity, err := s.stateView.GetLiquidity(s.ctx, s.stateViewContract, poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign())

	globals, err := s.stateView.GetFeeGrowthGlobals(s.ctx, s.stateViewContract, poolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(globals.FeeGrowthGlobal0.Sign())
	s.Require().Zero(globals.FeeGrowthGlobal1.Sign())

	s.T().Logf("StateViewSuite Test09UnknownPoolReadsAsZero Result")
	s.T().Logf("\tPoolID: %s", encoding.Hex.EncodePrefixed(poolID[:]))
}
