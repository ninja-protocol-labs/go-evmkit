package v4

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc721"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type PositionManagerSuite struct {
	suite.Suite

	Enable                 bool
	RPCURL                 string
	PrivateKeyHex          string
	PoolManagerHex         string
	PositionManagerHex     string
	Permit2Hex             string
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
	MintAmount0Max         *big.Int
	MintAmount1Max         *big.Int
	IncreaseLiquidity      *big.Int
	IncreaseAmount0Max     *big.Int
	IncreaseAmount1Max     *big.Int
	DecreaseLiquidity      *big.Int

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	deployer *deploy.NicksDeployer
	signer   core.Signer
	from     *types.Address
	chainID  *big.Int

	positionManager         *PositionManager
	erc20                   *erc20.ERC20
	permit2                 *permit2.Permit2
	poolManagerContract     *types.Address
	positionManagerContract *types.Address
	permit2Contract         *types.Address

	currency0 *types.Address
	currency1 *types.Address
	key       *PoolKey
	tokenID   *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPositionManagerSuite(t *testing.T) {
	suite.Run(t, &PositionManagerSuite{
		Enable:                 false,
		RPCURL:                 "",
		PrivateKeyHex:          "",
		PoolManagerHex:         "",
		PositionManagerHex:     "",
		Permit2Hex:             "",
		NativeSymbol:           "",
		Interval:               1500,
		Timeout:                60,
		ByteCode:               erc20ByteCodeHex,
		Constructor:            abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                   core.Keccak256([]byte("go-evmkit-v4-position")),
		MintValue:              new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		Permit2ApproveAmount:   new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)),
		Permit2ApproveDuration: big.NewInt(30 * 24 * 60 * 60),
		MintLiquidity:          big.NewInt(1_000_000),
		MintAmount0Max:         big.NewInt(1_000_000_000),
		MintAmount1Max:         big.NewInt(1_000_000_000),
		IncreaseLiquidity:      big.NewInt(200_000),
		IncreaseAmount0Max:     big.NewInt(1_000_000_000),
		IncreaseAmount1Max:     big.NewInt(1_000_000_000),
		DecreaseLiquidity:      big.NewInt(300_000),
	})
}

func (s *PositionManagerSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.positionManager = NewPositionManager(s.client)
	s.erc20 = erc20.New(s.client)
	s.permit2 = permit2.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID
	s.T().Logf("SetupSuite: chainID = %s", s.chainID)

	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.positionManagerContract, err = types.NewAddressFromHex(s.PositionManagerHex)
	s.Require().NoError(err)
	s.permit2Contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 position test token A", "evmkit v4 position test token B"}
	symbols := [2]string{"EVPA", "EVPB"}
	tokens := [2]*types.Address{}

	for i := range 2 {
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
		tokens[i] = token

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

		approveBuilder := tx.NewDynamicFeeBuilder()
		approveTx, approveHash, err := s.erc20.Approve(s.ctx, bind.NewTransactOptions(s.signer, approveBuilder), erc20.NewApproveRequest(token, s.permit2Contract, s.MintValue))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())
		s.mu.Unlock()

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "erc20 approve transaction reverted")

		expiration := big.NewInt(time.Now().Add(time.Duration(s.Permit2ApproveDuration.Int64()) * time.Second).Unix())
		permit2Params := permit2.NewApproveParams(s.permit2Contract, s.from, token, s.positionManagerContract, s.Permit2ApproveAmount, expiration)

		permit2Builder := tx.NewDynamicFeeBuilder()
		unsignedPermit2Approve, err := s.permit2.CreateApprove(s.ctx, permit2Builder, permit2Params)
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

	s.currency0, s.currency1 = tokens[0], tokens[1]
	if s.currency0.Big().Cmp(s.currency1.Big()) > 0 {
		s.currency0, s.currency1 = s.currency1, s.currency0
	}
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s", s.currency0, s.currency1)

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	initUnsigned, err := s.positionManager.CreateInitializePool(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, s.key, sqrtPriceX96, big.NewInt(0))
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
	s.T().Logf("SetupSuite: initialized pool currency0=%s currency1=%s via PositionManager.initializePool (tx %s)", s.currency0, s.currency1, initHash)
}

func (s *PositionManagerSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("PositionManagerSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PositionManagerSuite) Test01MintPosition() {
	mintParams := NewMintPositionParams(s.key, big.NewInt(-600), big.NewInt(600), s.MintLiquidity, s.MintAmount0Max, s.MintAmount1Max, s.from, nil)
	mintData, err := EncodeActionMintPosition(mintParams)
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)

	actions := []byte{ActionMintPosition, ActionSettlePair}
	params := [][]byte{mintData, settleData}
	deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	unsigned, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, actions, params, deadline, big.NewInt(0))
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
	s.Require().True(receipt.Status, "modifyLiquidities (mint) transaction reverted")

	var positionManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.positionManagerContract) {
			positionManagerLogs = append(positionManagerLogs, log)
		}
	}

	transfers, err := erc721.ExtractTransfers(positionManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].From.Equal(types.ZeroAddress()))
	s.Require().True(transfers[0].To.Equal(s.from))
	s.tokenID = transfers[0].TokenId

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().Zero(big.NewInt(-600).Cmp(mods[0].TickLower))
	s.Require().Zero(big.NewInt(600).Cmp(mods[0].TickUpper))
	s.Require().Zero(s.MintLiquidity.Cmp(mods[0].LiquidityDelta))
	var wantSalt [32]byte
	s.tokenID.FillBytes(wantSalt[:])
	s.Require().Equal(wantSalt, mods[0].Salt)

	liquidity, err := s.positionManager.GetPositionLiquidity(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.MintLiquidity.Cmp(liquidity))

	s.T().Logf("PositionManagerSuite Test01MintPosition Result")
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tLiquidity: %s", liquidity)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PositionManagerSuite) Test02IncreaseLiquidity() {
	increaseData, err := EncodeActionIncreaseLiquidity(s.tokenID, s.IncreaseLiquidity, s.IncreaseAmount0Max, s.IncreaseAmount1Max, nil)
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)

	actions := []byte{ActionIncreaseLiquidity, ActionSettlePair}
	params := [][]byte{increaseData, settleData}
	deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	unsigned, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, actions, params, deadline, big.NewInt(0))
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
	s.Require().True(receipt.Status, "modifyLiquidities (increase) transaction reverted")

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().Zero(s.IncreaseLiquidity.Cmp(mods[0].LiquidityDelta))
	var wantSalt [32]byte
	s.tokenID.FillBytes(wantSalt[:])
	s.Require().Equal(wantSalt, mods[0].Salt)

	liquidity, err := s.positionManager.GetPositionLiquidity(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	wantLiquidity := new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity)
	s.Require().Zero(wantLiquidity.Cmp(liquidity))

	s.T().Logf("PositionManagerSuite Test02IncreaseLiquidity Result")
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tLiquidity after: %s", liquidity)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PositionManagerSuite) Test03DecreaseLiquidity() {
	decreaseData, err := EncodeActionDecreaseLiquidity(s.tokenID, s.DecreaseLiquidity, big.NewInt(0), big.NewInt(0), nil)
	s.Require().NoError(err)
	takeData, err := EncodeActionTakePair(s.key.Currency0, s.key.Currency1, s.from)
	s.Require().NoError(err)

	actions := []byte{ActionDecreaseLiquidity, ActionTakePair}
	params := [][]byte{decreaseData, takeData}
	deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	unsigned, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, actions, params, deadline, big.NewInt(0))
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
	s.Require().True(receipt.Status, "modifyLiquidities (decrease) transaction reverted")

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	wantDelta := new(big.Int).Neg(s.DecreaseLiquidity)
	s.Require().Zero(wantDelta.Cmp(mods[0].LiquidityDelta))
	var wantSalt [32]byte
	s.tokenID.FillBytes(wantSalt[:])
	s.Require().Equal(wantSalt, mods[0].Salt)

	liquidity, err := s.positionManager.GetPositionLiquidity(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	wantLiquidity := new(big.Int).Sub(new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity), s.DecreaseLiquidity)
	s.Require().Zero(wantLiquidity.Cmp(liquidity))

	s.T().Logf("PositionManagerSuite Test03DecreaseLiquidity Result")
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tLiquidity after: %s", liquidity)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PositionManagerSuite) Test04ViewFunctions() {
	nextTokenID, err := s.positionManager.NextTokenID(s.ctx, s.positionManagerContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Add(s.tokenID, big.NewInt(1)).Cmp(nextTokenID))

	poolKey, info, err := s.positionManager.GetPoolAndPositionInfo(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolKey.Currency0.Equal(s.key.Currency0))
	s.Require().True(poolKey.Currency1.Equal(s.key.Currency1))
	s.Require().Zero(s.key.Fee.Cmp(poolKey.Fee))
	s.Require().Zero(s.key.TickSpacing.Cmp(poolKey.TickSpacing))
	s.Require().Zero(big.NewInt(-600).Cmp(info.TickLower))
	s.Require().Zero(big.NewInt(600).Cmp(info.TickUpper))
	s.Require().False(info.HasSubscriber)

	standaloneInfo, err := s.positionManager.PositionInfo(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal(info.PoolID, standaloneInfo.PoolID)
	s.Require().Zero(info.TickLower.Cmp(standaloneInfo.TickLower))
	s.Require().Zero(info.TickUpper.Cmp(standaloneInfo.TickUpper))

	poolKeyByID, err := s.positionManager.PoolKeys(s.ctx, s.positionManagerContract, info.PoolID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolKeyByID.Currency0.Equal(s.key.Currency0))
	s.Require().True(poolKeyByID.Currency1.Equal(s.key.Currency1))

	s.T().Logf("PositionManagerSuite Test04ViewFunctions Result")
	s.T().Logf("\tNextTokenID: %s", nextTokenID)
	s.T().Logf("\tTickLower/Upper: %s/%s", info.TickLower, info.TickUpper)
}

func (s *PositionManagerSuite) Test05ERC721Permit() {
	spenderKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	spender := core.PubkeyToAddress(spenderKey.PublicKey())

	nonce := big.NewInt(1)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	domain := NewPositionManagerDomain(s.chainID, s.positionManagerContract)
	permit := NewERC721Permit(spender, s.tokenID, nonce, deadline)
	sig, err := permit.Sign(domain, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	unsigned, err := s.positionManager.CreateERC721Permit(s.ctx, builder, s.positionManagerContract, s.from, spender, s.tokenID, deadline, nonce, sig.LegacyBytes(), big.NewInt(0))
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
	s.Require().True(receipt.Status, "permit transaction reverted")

	approved, err := s.positionManager.GetApproved(s.ctx, nil, erc721.NewGetApprovedRequest(s.positionManagerContract, s.tokenID))
	s.Require().NoError(err)
	s.Require().True(approved.Equal(spender))

	s.T().Logf("PositionManagerSuite Test05ERC721Permit Result")
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tApproved: %s", approved)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PositionManagerSuite) Test06BurnPosition() {
	burnData, err := EncodeActionBurnPosition(s.tokenID, big.NewInt(0), big.NewInt(0), nil)
	s.Require().NoError(err)
	takeData, err := EncodeActionTakePair(s.key.Currency0, s.key.Currency1, s.from)
	s.Require().NoError(err)

	actions := []byte{ActionBurnPosition, ActionTakePair}
	params := [][]byte{burnData, takeData}
	deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	unsigned, err := s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, actions, params, deadline, big.NewInt(0))
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
	s.Require().True(receipt.Status, "modifyLiquidities (burn) transaction reverted")

	var positionManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.positionManagerContract) {
			positionManagerLogs = append(positionManagerLogs, log)
		}
	}

	transfers, err := erc721.ExtractTransfers(positionManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(types.ZeroAddress()))
	s.Require().Zero(s.tokenID.Cmp(transfers[0].TokenId))

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManagerContract) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	wantDelta := new(big.Int).Neg(new(big.Int).Sub(new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity), s.DecreaseLiquidity))
	s.Require().Zero(wantDelta.Cmp(mods[0].LiquidityDelta))
	var wantSalt [32]byte
	s.tokenID.FillBytes(wantSalt[:])
	s.Require().Equal(wantSalt, mods[0].Salt)

	liquidity, err := s.positionManager.GetPositionLiquidity(s.ctx, s.positionManagerContract, s.tokenID, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(liquidity))

	s.T().Logf("PositionManagerSuite Test06BurnPosition Result")
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tLiquidity after: %s", liquidity)
	s.T().Logf("\tTx: %s", hash)
}

func (s *PositionManagerSuite) Test07InitializePoolAlreadyInitializedReturnsMaxTick() {
	tick, err := s.positionManager.CallInitializePool(s.ctx, s.positionManagerContract, s.from, s.key, new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(0), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(8388607).Cmp(tick), "already-initialized pool should report type(int24).max")

	s.T().Logf("PositionManagerSuite Test07InitializePoolAlreadyInitializedReturnsMaxTick Result")
	s.T().Logf("\tTick: %s", tick)
}

func (s *PositionManagerSuite) Test08DeadlinePassedError() {
	mintParams := NewMintPositionParams(s.key, big.NewInt(-600), big.NewInt(600), s.MintLiquidity, s.MintAmount0Max, s.MintAmount1Max, s.from, nil)
	mintData, err := EncodeActionMintPosition(mintParams)
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)

	actions := []byte{ActionMintPosition, ActionSettlePair}
	params := [][]byte{mintData, settleData}
	pastDeadline := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())

	_, err = s.positionManager.CreateModifyLiquidities(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, actions, params, pastDeadline, big.NewInt(0))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrDeadlinePassed)

	s.T().Logf("PositionManagerSuite Test08DeadlinePassedError Result")
	s.T().Logf("\tError: %s", err)
}

func (s *PositionManagerSuite) Test09ImmutableGetters() {
	poolManagerAddr, err := s.positionManager.PoolManager(s.ctx, s.positionManagerContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolManagerAddr.Equal(s.poolManagerContract))

	permit2Addr, err := s.positionManager.Permit2(s.ctx, s.positionManagerContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(permit2Addr.Equal(s.permit2Contract))

	weth9Addr, err := s.positionManager.WETH9(s.ctx, s.positionManagerContract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(weth9Addr.Equal(types.ZeroAddress()))

	s.T().Logf("PositionManagerSuite Test09ImmutableGetters Result")
	s.T().Logf("\tPoolManager: %s", poolManagerAddr)
	s.T().Logf("\tPermit2: %s", permit2Addr)
	s.T().Logf("\tWETH9: %s", weth9Addr)
}

func (s *PositionManagerSuite) Test10Multicall() {
	poolManagerCalldata, err := EncodePoolManager()
	s.Require().NoError(err)
	weth9Calldata, err := EncodeWETH9()
	s.Require().NoError(err)

	results, err := s.positionManager.CallMulticall(s.ctx, s.positionManagerContract, s.from, [][]byte{poolManagerCalldata, weth9Calldata}, big.NewInt(0), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(results, 2)

	poolManagerAddr, err := DecodePoolManager(results[0])
	s.Require().NoError(err)
	s.Require().True(poolManagerAddr.Equal(s.poolManagerContract))

	weth9Addr, err := DecodeWETH9(results[1])
	s.Require().NoError(err)
	s.Require().False(weth9Addr.Equal(types.ZeroAddress()))

	s.T().Logf("PositionManagerSuite Test10Multicall Result")
	s.T().Logf("\tPoolManager: %s", poolManagerAddr)
	s.T().Logf("\tWETH9: %s", weth9Addr)
}

func (s *PositionManagerSuite) Test11NoncesAndRevokeNonce() {
	nonce := big.NewInt(5)

	before, err := s.positionManager.Nonces(s.ctx, s.positionManagerContract, s.from, big.NewInt(0), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(before.Bit(5) == 1, "nonce 5 should not be used yet")

	unsigned, err := s.positionManager.CreateRevokeNonce(s.ctx, tx.NewDynamicFeeBuilder(), s.positionManagerContract, s.from, nonce, big.NewInt(0))
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
	s.Require().True(receipt.Status, "revokeNonce transaction reverted")

	after, err := s.positionManager.Nonces(s.ctx, s.positionManagerContract, s.from, big.NewInt(0), rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(after.Bit(5) == 1, "nonce 5 should now be used")

	s.T().Logf("PositionManagerSuite Test11NoncesAndRevokeNonce Result")
	s.T().Logf("\tBitmap before/after: %s -> %s", before, after)
	s.T().Logf("\tTx: %s", hash)
}
