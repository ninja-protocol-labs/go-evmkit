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

type PositionManagerSuite struct {
	suite.Suite

	RPCURL                 string
	PrivateKeyHex          string
	PoolManagerHex         string
	PositionManagerHex     string
	Permit2Hex             string
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
	MintAmount0Max         *big.Int
	MintAmount1Max         *big.Int
	IncreaseLiquidity      *big.Int
	IncreaseAmount0Max     *big.Int
	IncreaseAmount1Max     *big.Int
	DecreaseLiquidity      *big.Int
	NativeMintLiquidity    *big.Int
	NativeAmount           *big.Int

	client          rpc.Client
	signer          core.Signer
	deployer        *deploy.NicksDeployer
	from            *types.Address
	spender         *types.Address
	operator        *types.Address
	chainID         *big.Int
	poolManager     *types.Address
	contract        *types.Address
	permit2Contract *types.Address
	multicall       *types.Address
	nicks           *types.Address
	currency0       *types.Address
	currency1       *types.Address

	positionManager *PositionManager
	erc20           *erc20.ERC20
	permit2         *permit2.Permit2

	key          *PoolKey
	nativeKey    *PoolKey
	sqrtPriceX96 *big.Int
	tickLower    *big.Int
	tickUpper    *big.Int
	tokenID      *big.Int
	nonceWord    *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPositionManagerSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V4.PoolManager == "" || c.Addresses.Protocol.Uniswap.V4.PositionManager == "" {
		t.Skip("testenv: no Uniswap V4 PoolManager/PositionManager configured for this network")
	}
	if c.Addresses.Standard.Permit2 == "" {
		t.Skip("testenv: no Permit2 configured for this network")
	}

	suite.Run(t, &PositionManagerSuite{
		RPCURL:                 c.RPCURL,
		PrivateKeyHex:          c.ActorPrivateKeyHex,
		PoolManagerHex:         c.Addresses.Protocol.Uniswap.V4.PoolManager,
		PositionManagerHex:     c.Addresses.Protocol.Uniswap.V4.PositionManager,
		Permit2Hex:             c.Addresses.Standard.Permit2,
		MulticallHex:           c.Addresses.Standard.Multicall,
		NicksFactoryHex:        c.Addresses.Standard.NicksFactory,
		NativeSymbol:           c.Constant.NativeSymbol,
		NativeDecimals:         c.Constant.NativeDecimals,
		Timeout:                10 * time.Minute,
		ReceiptPollInterval:    c.Constant.ReceiptPollInterval,
		ReceiptTimeout:         c.Constant.ReceiptTimeout,
		ByteCode:               testenv.ERC20ByteCodeHex,
		Constructor:            abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                   core.Keccak256([]byte("go-evmkit-v4-positionmanager")),
		MintValue:              new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		Permit2ApproveAmount:   new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)),
		Permit2ApproveDuration: 30 * 24 * time.Hour,
		MintLiquidity:          big.NewInt(1_000_000),
		MintAmount0Max:         big.NewInt(1_000_000_000),
		MintAmount1Max:         big.NewInt(1_000_000_000),
		IncreaseLiquidity:      big.NewInt(200_000),
		IncreaseAmount0Max:     big.NewInt(1_000_000_000),
		IncreaseAmount1Max:     big.NewInt(1_000_000_000),
		DecreaseLiquidity:      big.NewInt(300_000),
		NativeMintLiquidity:    big.NewInt(1_000_000),
		NativeAmount:           big.NewInt(1_000_000_000),
	})
}

func (s *PositionManagerSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.positionManager = NewPositionManager(s.client)
	s.erc20 = erc20.New(s.client)
	s.permit2 = permit2.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.spender = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-spender", s.Salt)).Bytes()[12:])
	s.operator = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-operator", s.Salt)).Bytes()[12:])
	s.T().Logf("SetupSuite: signing as %s, spender %s, operator %s", s.from, s.spender, s.operator)

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID
	s.T().Logf("SetupSuite: chainID %s", s.chainID)

	s.poolManager, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.contract, err = types.NewAddressFromHex(s.PositionManagerHex)
	s.Require().NoError(err)
	s.permit2Contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: PoolManager %s, PositionManager %s, Permit2 %s", s.poolManager, s.contract, s.permit2Contract)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 position test token A", "evmkit v4 position test token B"}
	symbols := [2]string{"EVPA", "EVPB"}
	tokens := [2]*types.Address{}
	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	expiration := big.NewInt(time.Now().Add(s.Permit2ApproveDuration).Unix())

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		predicted, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

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
		s.T().Logf("SetupSuite: %s mined in block %d (gas %d)", symbols[i], receipt.BlockNumber, receipt.GasUsed)

		token, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.T().Context(), rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, deployTx.AddressNonce(), hash)

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

		approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(token, s.permit2Contract, s.MintValue))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, approveTx.MaxFee())
		s.mu.Unlock()

		approveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "erc20 approve transaction reverted")
		s.T().Logf("SetupSuite: approved Permit2 for %s %s (tx %s)", s.MintValue, symbols[i], approveHash)

		permit2ApproveTx, permit2ApproveHash, err := s.permit2.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), permit2.NewApproveRequest(s.permit2Contract, token, s.contract, s.Permit2ApproveAmount, expiration))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permit2ApproveTx.MaxFee())
		s.mu.Unlock()

		permit2ApproveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, permit2ApproveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(permit2ApproveReceipt.Status, "permit2 approve transaction reverted")
		s.T().Logf("SetupSuite: Permit2 allowance for PositionManager on %s set to %s until %s (tx %s)", symbols[i], s.Permit2ApproveAmount, expiration, permit2ApproveHash)
	}

	s.currency0, s.currency1 = tokens[0], tokens[1]
	if s.currency0.Big().Cmp(s.currency1.Big()) > 0 {
		s.currency0, s.currency1 = s.currency1, s.currency0
	}
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s", s.currency0, s.currency1)

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.nativeKey = NewPoolKey(types.ZeroAddress(), s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.sqrtPriceX96 = new(big.Int).Lsh(big.NewInt(1), 96)
	s.tickLower = big.NewInt(-600)
	s.tickUpper = big.NewInt(600)
	s.nonceWord = big.NewInt(time.Now().UnixNano())
	s.T().Logf("SetupSuite: pool key fee %s tickSpacing %s, native key currency1 %s, unordered nonce word %s", s.key.Fee, s.key.TickSpacing, s.nativeKey.Currency1, s.nonceWord)
}

func (s *PositionManagerSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	if s.currency0 != nil && s.currency1 != nil {
		balances, err := s.erc20.AddressBalances(ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
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

		s.T().Logf("PositionManagerSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *PositionManagerSuite) Test01ImmutableGetters() {
	poolManager, err := s.positionManager.PoolManager(s.T().Context(), nil, NewPoolManagerRequest(s.contract))
	s.Require().NoError(err)
	s.Require().True(poolManager.Equal(s.poolManager))

	permit2Addr, err := s.positionManager.Permit2(s.T().Context(), nil, NewPermit2Request(s.contract))
	s.Require().NoError(err)
	s.Require().True(permit2Addr.Equal(s.permit2Contract))

	weth9, err := s.positionManager.WETH9(s.T().Context(), nil, NewWETH9Request(s.contract))
	s.Require().NoError(err)
	s.Require().False(weth9.Equal(types.ZeroAddress()))

	gasLimit, err := s.positionManager.UnsubscribeGasLimit(s.T().Context(), nil, NewUnsubscribeGasLimitRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Positive(gasLimit.Sign())

	nextTokenID, err := s.positionManager.NextTokenID(s.T().Context(), nil, NewNextTokenIDRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Positive(nextTokenID.Sign())

	name, err := s.positionManager.Name(s.T().Context(), nil, erc721.NewNameRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Equal("Uniswap v4 Positions NFT", name)

	s.T().Logf("PositionManager %s Test01ImmutableGetters Result", s.contract)
	s.T().Logf("\tPoolManager: %s", poolManager)
	s.T().Logf("\tPermit2: %s", permit2Addr)
	s.T().Logf("\tWETH9: %s", weth9)
	s.T().Logf("\tUnsubscribeGasLimit: %s", gasLimit)
	s.T().Logf("\tNextTokenID: %s", nextTokenID)
	s.T().Logf("\tName: %s", name)
}

func (s *PositionManagerSuite) Test02CallInitializePoolPredictsTick() {
	tick, err := s.positionManager.CallInitializePool(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializePoolRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().Zero(tick.Sign())

	s.T().Logf("PositionManager %s Test02CallInitializePoolPredictsTick Result", s.contract)
	s.T().Logf("\tCallInitializePool(sqrtPrice 2^96): tick %s", tick)
}

func (s *PositionManagerSuite) Test03InitializePool() {
	sentTx, hash, err := s.positionManager.InitializePool(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializePoolRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03InitializePool: sent initializePool (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "initializePool transaction reverted")

	inits, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(inits, 1)
	s.Require().True(inits[0].Contract.Equal(s.poolManager))
	s.Require().True(inits[0].Currency0.Equal(s.currency0))
	s.Require().True(inits[0].Currency1.Equal(s.currency1))
	s.Require().Zero(inits[0].Tick.Sign())

	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var tick *big.Int
	s.Require().Eventually(func() bool {
		tick, err = s.positionManager.CallInitializePool(s.T().Context(), atBlock, NewInitializePoolRequest(s.contract, s.key, s.sqrtPriceX96))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "initializePool not callable at the receipt block")
	s.Require().Zero(big.NewInt(1<<23-1).Cmp(tick), "an initialized pool must report type(int24).max")

	s.T().Logf("PositionManager %s Test03InitializePool Result", s.contract)
	s.T().Logf("\tInitialize event: id=%x currency0=%s currency1=%s fee=%s tickSpacing=%s sqrtPriceX96=%s tick=%s", inits[0].Id, inits[0].Currency0, inits[0].Currency1, inits[0].Fee, inits[0].TickSpacing, inits[0].SqrtPriceX96, inits[0].Tick)
	s.T().Logf("\tCallInitializePool at receipt block: tick %s", tick)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test04MintPosition() {
	mintData, err := EncodeActionMintPosition(NewMintPositionParams(s.key, s.tickLower, s.tickUpper, s.MintLiquidity, s.MintAmount0Max, s.MintAmount1Max, s.from, nil))
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	var sentTx core.Transaction
	var hash *types.Hash
	s.Require().Eventually(func() bool {
		sentTx, hash, err = s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionMintPosition, ActionSettlePair}, [][]byte{mintData, settleData}, deadline, nil))
		return !errors.Is(err, ErrPoolNotInitialized)
	}, 20*time.Second, s.ReceiptPollInterval, "pool still not initialized on the estimating node")
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04MintPosition: sent modifyLiquidities [MINT_POSITION, SETTLE_PAIR] (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidities (mint) transaction reverted")

	var positionManagerLogs, poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.contract) {
			positionManagerLogs = append(positionManagerLogs, log)
		}
		if log.Address.Equal(s.poolManager) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}

	transfers, err := erc721.ExtractTransfers(positionManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].From.Equal(types.ZeroAddress()))
	s.Require().True(transfers[0].To.Equal(s.from))
	s.tokenID = transfers[0].TokenId

	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().True(mods[0].Sender.Equal(s.contract))
	s.Require().Zero(s.tickLower.Cmp(mods[0].TickLower))
	s.Require().Zero(s.tickUpper.Cmp(mods[0].TickUpper))
	s.Require().Zero(s.MintLiquidity.Cmp(mods[0].LiquidityDelta))
	var wantSalt [32]byte
	s.tokenID.FillBytes(wantSalt[:])
	s.Require().Equal(wantSalt, mods[0].Salt)

	positions, err := ExtractModifyPosition(positionManagerLogs)
	s.Require().NoError(err)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.positionManager.GetPositionLiquidity(s.T().Context(), atBlock, NewGetPositionLiquidityRequest(s.contract, s.tokenID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "position liquidity not readable at the receipt block")
	s.Require().Zero(s.MintLiquidity.Cmp(liquidity))

	owner, err := s.positionManager.OwnerOf(s.T().Context(), atBlock, erc721.NewOwnerOfRequest(s.contract, s.tokenID))
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))

	s.T().Logf("PositionManager %s Test04MintPosition Result", s.contract)
	s.T().Logf("\tTokenID: %s (owner %s)", s.tokenID, owner)
	s.T().Logf("\tModifyLiquidity event: id=%x sender=%s ticks=[%s, %s] delta=%s salt=%x", mods[0].Id, mods[0].Sender, mods[0].TickLower, mods[0].TickUpper, mods[0].LiquidityDelta, mods[0].Salt)
	s.T().Logf("\tModifyPosition events emitted by PositionManager: %d", len(positions))
	s.T().Logf("\tLiquidity at receipt block: %s", liquidity)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test05IncreaseLiquidity() {
	increaseData, err := EncodeActionIncreaseLiquidity(s.tokenID, s.IncreaseLiquidity, s.IncreaseAmount0Max, s.IncreaseAmount1Max, nil)
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	sentTx, hash, err := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionIncreaseLiquidity, ActionSettlePair}, [][]byte{increaseData, settleData}, deadline, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05IncreaseLiquidity: sent modifyLiquidities [INCREASE_LIQUIDITY, SETTLE_PAIR] (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidities (increase) transaction reverted")

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManager) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().Zero(s.IncreaseLiquidity.Cmp(mods[0].LiquidityDelta))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.positionManager.GetPositionLiquidity(s.T().Context(), atBlock, NewGetPositionLiquidityRequest(s.contract, s.tokenID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "position liquidity not readable at the receipt block")
	want := new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity)
	s.Require().Zero(want.Cmp(liquidity))

	s.T().Logf("PositionManager %s Test05IncreaseLiquidity Result", s.contract)
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tModifyLiquidity delta: %s", mods[0].LiquidityDelta)
	s.T().Logf("\tLiquidity at receipt block: %s (want %s)", liquidity, want)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test06DecreaseLiquidity() {
	decreaseData, err := EncodeActionDecreaseLiquidity(s.tokenID, s.DecreaseLiquidity, big.NewInt(0), big.NewInt(0), nil)
	s.Require().NoError(err)
	takeData, err := EncodeActionTakePair(s.key.Currency0, s.key.Currency1, s.from)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	before, err := s.erc20.AddressBalances(s.T().Context(), nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
	s.Require().NoError(err)

	sentTx, hash, err := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionDecreaseLiquidity, ActionTakePair}, [][]byte{decreaseData, takeData}, deadline, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test06DecreaseLiquidity: sent modifyLiquidities [DECREASE_LIQUIDITY, TAKE_PAIR] (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidities (decrease) transaction reverted")

	var poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.poolManager) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}
	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().Zero(new(big.Int).Neg(s.DecreaseLiquidity).Cmp(mods[0].LiquidityDelta))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.positionManager.GetPositionLiquidity(s.T().Context(), atBlock, NewGetPositionLiquidityRequest(s.contract, s.tokenID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "position liquidity not readable at the receipt block")
	want := new(big.Int).Sub(new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity), s.DecreaseLiquidity)
	s.Require().Zero(want.Cmp(liquidity))

	after, err := s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
	s.Require().NoError(err)
	s.Require().Len(after.Tokens, 2)
	gain0 := new(big.Int).Sub(after.Tokens[0].Amount, before.Tokens[0].Amount)
	gain1 := new(big.Int).Sub(after.Tokens[1].Amount, before.Tokens[1].Amount)
	s.Require().Positive(gain0.Sign())
	s.Require().Positive(gain1.Sign())

	s.T().Logf("PositionManager %s Test06DecreaseLiquidity Result", s.contract)
	s.T().Logf("\tTokenID: %s", s.tokenID)
	s.T().Logf("\tModifyLiquidity delta: %s", mods[0].LiquidityDelta)
	s.T().Logf("\tLiquidity at receipt block: %s (want %s)", liquidity, want)
	s.T().Logf("\tTaken: currency0 +%s, currency1 +%s", gain0, gain1)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test07ViewFunctions() {
	nextTokenID, err := s.positionManager.NextTokenID(s.T().Context(), nil, NewNextTokenIDRequest(s.contract))
	s.Require().NoError(err)
	s.Require().Positive(nextTokenID.Cmp(s.tokenID))

	poolKey, info, err := s.positionManager.GetPoolAndPositionInfo(s.T().Context(), nil, NewGetPoolAndPositionInfoRequest(s.contract, s.tokenID))
	s.Require().NoError(err)
	s.Require().True(poolKey.Currency0.Equal(s.key.Currency0))
	s.Require().True(poolKey.Currency1.Equal(s.key.Currency1))
	s.Require().Zero(s.key.Fee.Cmp(poolKey.Fee))
	s.Require().Zero(s.key.TickSpacing.Cmp(poolKey.TickSpacing))
	s.Require().True(poolKey.Hooks.Equal(s.key.Hooks))
	s.Require().Zero(s.tickLower.Cmp(info.TickLower))
	s.Require().Zero(s.tickUpper.Cmp(info.TickUpper))
	s.Require().False(info.HasSubscriber)

	poolID, err := s.key.PoolID()
	s.Require().NoError(err)
	s.Require().Equal(poolID[:25], info.PoolID[:])

	standalone, err := s.positionManager.PositionInfo(s.T().Context(), nil, NewPositionInfoRequest(s.contract, s.tokenID))
	s.Require().NoError(err)
	s.Require().Equal(info.PoolID, standalone.PoolID)
	s.Require().Zero(info.TickLower.Cmp(standalone.TickLower))
	s.Require().Zero(info.TickUpper.Cmp(standalone.TickUpper))
	s.Require().Equal(info.HasSubscriber, standalone.HasSubscriber)

	byID, err := s.positionManager.PoolKeys(s.T().Context(), nil, NewPoolKeysRequest(s.contract, info.PoolID))
	s.Require().NoError(err)
	s.Require().True(byID.Currency0.Equal(s.key.Currency0))
	s.Require().True(byID.Currency1.Equal(s.key.Currency1))
	s.Require().Zero(s.key.Fee.Cmp(byID.Fee))

	subscriber, err := s.positionManager.Subscriber(s.T().Context(), nil, NewSubscriberRequest(s.contract, s.tokenID))
	s.Require().NoError(err)
	s.Require().True(subscriber.Equal(types.ZeroAddress()))

	s.T().Logf("PositionManager %s Test07ViewFunctions Result", s.contract)
	s.T().Logf("\tNextTokenID: %s", nextTokenID)
	s.T().Logf("\tPoolKey: currency0=%s currency1=%s fee=%s tickSpacing=%s hooks=%s", poolKey.Currency0, poolKey.Currency1, poolKey.Fee, poolKey.TickSpacing, poolKey.Hooks)
	s.T().Logf("\tPositionInfo: poolId=%x ticks=[%s, %s] hasSubscriber=%v", info.PoolID, info.TickLower, info.TickUpper, info.HasSubscriber)
	s.T().Logf("\tSubscriber: %s", subscriber)
}

func (s *PositionManagerSuite) Test08ModifyLiquiditiesRejectsPastDeadline() {
	mintData, err := EncodeActionMintPosition(NewMintPositionParams(s.key, s.tickLower, s.tickUpper, s.MintLiquidity, s.MintAmount0Max, s.MintAmount1Max, s.from, nil))
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.key.Currency0, s.key.Currency1)
	s.Require().NoError(err)
	pastDeadline := big.NewInt(time.Now().Add(-time.Hour).Unix())

	sentTx, hash, err := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionMintPosition, ActionSettlePair}, [][]byte{mintData, settleData}, pastDeadline, nil))
	s.Require().ErrorIs(err, ErrDeadlinePassed)
	s.Require().ErrorContains(err, "uniswapv4: modifyLiquidities: ")
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	s.T().Logf("PositionManager %s Test08ModifyLiquiditiesRejectsPastDeadline Result", s.contract)
	s.T().Logf("\tDeadline: %s", pastDeadline)
	s.T().Logf("\tError: %s", err)
}

func (s *PositionManagerSuite) Test09UnsettledMintMapsPoolManagerRevert() {
	mintData, err := EncodeActionMintPosition(NewMintPositionParams(s.key, s.tickLower, s.tickUpper, s.MintLiquidity, s.MintAmount0Max, s.MintAmount1Max, s.from, nil))
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	sentTx, hash, err := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionMintPosition}, [][]byte{mintData}, deadline, nil))
	s.Require().ErrorIs(err, ErrCurrencyNotSettled)
	s.Require().NotErrorIs(err, ErrUnknown)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	s.T().Logf("PositionManager %s Test09UnsettledMintMapsPoolManagerRevert Result", s.contract)
	s.T().Logf("\tError: %s", err)
}

func (s *PositionManagerSuite) Test10SubscribeAndUnsubscribeRevertsMapToSentinels() {
	subscribeTx, subscribeHash, err := s.positionManager.Subscribe(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSubscribeRequest(s.contract, s.tokenID, s.spender, nil))
	s.Require().ErrorIs(err, ErrNoCodeSubscriber)
	s.Require().ErrorContains(err, "uniswapv4: subscribe: ")
	s.Require().Nil(subscribeTx)
	s.Require().Nil(subscribeHash)

	unsubscribeTx, unsubscribeHash, err2 := s.positionManager.Unsubscribe(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewUnsubscribeRequest(s.contract, s.tokenID))
	s.Require().ErrorIs(err2, ErrNotSubscribed)
	s.Require().ErrorContains(err2, "uniswapv4: unsubscribe: ")
	s.Require().Nil(unsubscribeTx)
	s.Require().Nil(unsubscribeHash)

	s.T().Logf("PositionManager %s Test10SubscribeAndUnsubscribeRevertsMapToSentinels Result", s.contract)
	s.T().Logf("\tSubscribe(codeless %s): %s", s.spender, err)
	s.T().Logf("\tUnsubscribe(no subscriber): %s", err2)
}

func (s *PositionManagerSuite) Test11ERC721Permit() {
	nonce := new(big.Int).Lsh(s.nonceWord, 8)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	sig, err := NewERC721Permit(s.spender, s.tokenID, nonce, deadline).Sign(NewPositionManagerDomain(s.chainID, s.contract), s.signer)
	s.Require().NoError(err)

	sentTx, hash, err := s.positionManager.ERC721Permit(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewERC721PermitRequest(s.contract, s.spender, s.tokenID, deadline, nonce, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11ERC721Permit: sent permit (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")

	approvals, err := erc721.ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Approved.Equal(s.spender))
	s.Require().Zero(s.tokenID.Cmp(approvals[0].TokenId))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var approved *types.Address
	s.Require().Eventually(func() bool {
		approved, err = s.positionManager.GetApproved(s.T().Context(), atBlock, erc721.NewGetApprovedRequest(s.contract, s.tokenID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "getApproved not readable at the receipt block")
	s.Require().True(approved.Equal(s.spender))

	bitmap, err := s.positionManager.Nonces(s.T().Context(), atBlock, NewNoncesRequest(s.contract, s.from, s.nonceWord))
	s.Require().NoError(err)
	s.Require().Equal(uint(1), bitmap.Bit(0))

	replayTx, replayHash, err := s.positionManager.ERC721Permit(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewERC721PermitRequest(s.contract, s.spender, s.tokenID, deadline, nonce, sig.LegacyBytes()))
	s.Require().ErrorIs(err, ErrNonceAlreadyUsed)
	s.Require().Nil(replayTx)
	s.Require().Nil(replayHash)

	s.T().Logf("PositionManager %s Test11ERC721Permit Result", s.contract)
	s.T().Logf("\tApproval event: owner=%s approved=%s tokenId=%s", approvals[0].Owner, approvals[0].Approved, approvals[0].TokenId)
	s.T().Logf("\tGetApproved at receipt block: %s", approved)
	s.T().Logf("\tNonce word %s bitmap: %s", s.nonceWord, bitmap)
	s.T().Logf("\tReplay: %s", err)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test12ERC721PermitForAll() {
	nonce := new(big.Int).Add(new(big.Int).Lsh(s.nonceWord, 8), big.NewInt(1))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	sig, err := NewERC721PermitForAll(s.operator, true, nonce, deadline).Sign(NewPositionManagerDomain(s.chainID, s.contract), s.signer)
	s.Require().NoError(err)

	sentTx, hash, err := s.positionManager.ERC721PermitForAll(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewERC721PermitForAllRequest(s.contract, s.from, s.operator, true, deadline, nonce, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test12ERC721PermitForAll: sent permitForAll (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitForAll transaction reverted")

	approvals, err := erc721.ExtractApprovalForAll(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Operator.Equal(s.operator))
	s.Require().True(approvals[0].Approved)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var approvedForAll bool
	s.Require().Eventually(func() bool {
		approvedForAll, err = s.positionManager.IsApprovedForAll(s.T().Context(), atBlock, erc721.NewIsApprovedForAllRequest(s.contract, s.from, s.operator))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "isApprovedForAll not readable at the receipt block")
	s.Require().True(approvedForAll)

	expiredDeadline := big.NewInt(time.Now().Add(-time.Minute).Unix())
	expiredNonce := new(big.Int).Add(nonce, big.NewInt(1))
	expiredSig, err := NewERC721PermitForAll(s.operator, false, expiredNonce, expiredDeadline).Sign(NewPositionManagerDomain(s.chainID, s.contract), s.signer)
	s.Require().NoError(err)
	expiredTx, expiredHash, err := s.positionManager.ERC721PermitForAll(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewERC721PermitForAllRequest(s.contract, s.from, s.operator, false, expiredDeadline, expiredNonce, expiredSig.LegacyBytes()))
	s.Require().ErrorIs(err, ErrSignatureDeadlineExpired)
	s.Require().Nil(expiredTx)
	s.Require().Nil(expiredHash)

	s.T().Logf("PositionManager %s Test12ERC721PermitForAll Result", s.contract)
	s.T().Logf("\tApprovalForAll event: owner=%s operator=%s approved=%v", approvals[0].Owner, approvals[0].Operator, approvals[0].Approved)
	s.T().Logf("\tIsApprovedForAll at receipt block: %v", approvedForAll)
	s.T().Logf("\tExpired permit: %s", err)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test13RevokeNonce() {
	nonce := new(big.Int).Add(new(big.Int).Lsh(s.nonceWord, 8), big.NewInt(5))

	before, err := s.positionManager.Nonces(s.T().Context(), nil, NewNoncesRequest(s.contract, s.from, s.nonceWord))
	s.Require().NoError(err)
	s.Require().Zero(before.Bit(5))

	sentTx, hash, err := s.positionManager.RevokeNonce(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewRevokeNonceRequest(s.contract, nonce))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test13RevokeNonce: sent revokeNonce(%s) (nonce %d, maxFee %s), tx %s", nonce, sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "revokeNonce transaction reverted")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.positionManager.Nonces(s.T().Context(), atBlock, NewNoncesRequest(s.contract, s.from, s.nonceWord))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "nonces not readable at the receipt block")
	s.Require().Equal(uint(1), after.Bit(5))

	replayTx, replayHash, err := s.positionManager.RevokeNonce(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewRevokeNonceRequest(s.contract, nonce))
	s.Require().ErrorIs(err, ErrNonceAlreadyUsed)
	s.Require().Nil(replayTx)
	s.Require().Nil(replayHash)

	s.T().Logf("PositionManager %s Test13RevokeNonce Result", s.contract)
	s.T().Logf("\tNonce word %s bitmap: %s -> %s", s.nonceWord, before, after)
	s.T().Logf("\tReplay: %s", err)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test14Multicall() {
	poolManagerCalldata, err := EncodePoolManager()
	s.Require().NoError(err)
	permit2Calldata, err := EncodePermit2Contract()
	s.Require().NoError(err)

	results, err := s.positionManager.CallMulticall(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewMulticallRequest(s.contract, [][]byte{poolManagerCalldata, permit2Calldata}, nil))
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	poolManager, err := DecodePoolManager(results[0])
	s.Require().NoError(err)
	s.Require().True(poolManager.Equal(s.poolManager))
	permit2Addr, err := DecodePermit2Contract(results[1])
	s.Require().NoError(err)
	s.Require().True(permit2Addr.Equal(s.permit2Contract))

	revoke7, err := EncodeRevokeNonce(new(big.Int).Add(new(big.Int).Lsh(s.nonceWord, 8), big.NewInt(7)))
	s.Require().NoError(err)
	revoke8, err := EncodeRevokeNonce(new(big.Int).Add(new(big.Int).Lsh(s.nonceWord, 8), big.NewInt(8)))
	s.Require().NoError(err)

	sentTx, hash, err := s.positionManager.Multicall(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewMulticallRequest(s.contract, [][]byte{revoke7, revoke8}, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14Multicall: sent multicall [revokeNonce, revokeNonce] (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "multicall transaction reverted")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var bitmap *big.Int
	s.Require().Eventually(func() bool {
		bitmap, err = s.positionManager.Nonces(s.T().Context(), atBlock, NewNoncesRequest(s.contract, s.from, s.nonceWord))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "nonces not readable at the receipt block")
	s.Require().Equal(uint(1), bitmap.Bit(7))
	s.Require().Equal(uint(1), bitmap.Bit(8))

	s.T().Logf("PositionManager %s Test14Multicall Result", s.contract)
	s.T().Logf("\tCallMulticall: poolManager=%s permit2=%s", poolManager, permit2Addr)
	s.T().Logf("\tNonce word %s bitmap after multicall: %s", s.nonceWord, bitmap)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test15Permit2Permit() {
	current, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.spender))
	s.Require().NoError(err)

	amount := big.NewInt(123_456_789)
	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, amount, expiration, current.Nonce), s.spender, sigDeadline)
	sig, err := permitSingle.Sign(permit2.NewDomain(s.chainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	errData, err := s.positionManager.CallPermit2Permit(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewPermit2PermitRequest(s.contract, s.from, permitSingle, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.Require().Empty(errData)

	sentTx, hash, err := s.positionManager.Permit2Permit(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewPermit2PermitRequest(s.contract, s.from, permitSingle, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test15Permit2Permit: sent Permit2Forwarder.permit (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "Permit2Forwarder.permit transaction reverted")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var allowance *permit2.PackedAllowance
	s.Require().Eventually(func() bool {
		allowance, err = s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.spender))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "permit2 allowance not readable at the receipt block")
	s.Require().Zero(amount.Cmp(allowance.Amount))
	s.Require().Zero(expiration.Cmp(allowance.Expiration))
	s.Require().Zero(new(big.Int).Add(current.Nonce, big.NewInt(1)).Cmp(allowance.Nonce))

	replayData, err := s.positionManager.CallPermit2Permit(s.T().Context(), bind.NewCallOptions().WithFrom(s.from).WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewPermit2PermitRequest(s.contract, s.from, permitSingle, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.Require().NotEmpty(replayData, "a replayed permit must return the inner permit2 revert data")

	s.T().Logf("PositionManager %s Test15Permit2Permit Result", s.contract)
	s.T().Logf("\tAllowance(from, currency0, spender): amount %s -> %s, expiration %s, nonce %s -> %s", current.Amount, allowance.Amount, allowance.Expiration, current.Nonce, allowance.Nonce)
	s.T().Logf("\tReplay inner revert data: %x", replayData)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test16Permit2PermitBatch() {
	current0, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.operator))
	s.Require().NoError(err)
	current1, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.operator))
	s.Require().NoError(err)

	amount0 := big.NewInt(111)
	amount1 := big.NewInt(222)
	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	permitBatch := permit2.NewPermitBatch([]permit2.PermitDetails{
		*permit2.NewPermitDetails(s.currency0, amount0, expiration, current0.Nonce),
		*permit2.NewPermitDetails(s.currency1, amount1, expiration, current1.Nonce),
	}, s.operator, sigDeadline)
	sig, err := permitBatch.Sign(permit2.NewDomain(s.chainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	errData, err := s.positionManager.CallPermit2PermitBatch(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewPermit2PermitBatchRequest(s.contract, s.from, permitBatch, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.Require().Empty(errData)

	sentTx, hash, err := s.positionManager.Permit2PermitBatch(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewPermit2PermitBatchRequest(s.contract, s.from, permitBatch, sig.LegacyBytes()))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test16Permit2PermitBatch: sent Permit2Forwarder.permitBatch (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "Permit2Forwarder.permitBatch transaction reverted")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var allowance0 *permit2.PackedAllowance
	s.Require().Eventually(func() bool {
		allowance0, err = s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.operator))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "permit2 allowance not readable at the receipt block")
	s.Require().Zero(amount0.Cmp(allowance0.Amount))

	allowance1, err := s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.operator))
	s.Require().NoError(err)
	s.Require().Zero(amount1.Cmp(allowance1.Amount))
	s.Require().Zero(expiration.Cmp(allowance1.Expiration))

	s.T().Logf("PositionManager %s Test16Permit2PermitBatch Result", s.contract)
	s.T().Logf("\tAllowance(from, currency0, operator): %s (nonce %s -> %s)", allowance0.Amount, current0.Nonce, allowance0.Nonce)
	s.T().Logf("\tAllowance(from, currency1, operator): %s (nonce %s -> %s)", allowance1.Amount, current1.Nonce, allowance1.Nonce)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test17BurnPosition() {
	burnData, err := EncodeActionBurnPosition(s.tokenID, big.NewInt(0), big.NewInt(0), nil)
	s.Require().NoError(err)
	takeData, err := EncodeActionTakePair(s.key.Currency0, s.key.Currency1, s.from)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	sentTx, hash, err := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionBurnPosition, ActionTakePair}, [][]byte{burnData, takeData}, deadline, nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test17BurnPosition: sent modifyLiquidities [BURN_POSITION, TAKE_PAIR] (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidities (burn) transaction reverted")

	var positionManagerLogs, poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.contract) {
			positionManagerLogs = append(positionManagerLogs, log)
		}
		if log.Address.Equal(s.poolManager) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}

	transfers, err := erc721.ExtractTransfers(positionManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(types.ZeroAddress()))
	s.Require().Zero(s.tokenID.Cmp(transfers[0].TokenId))

	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	remaining := new(big.Int).Sub(new(big.Int).Add(s.MintLiquidity, s.IncreaseLiquidity), s.DecreaseLiquidity)
	s.Require().Zero(new(big.Int).Neg(remaining).Cmp(mods[0].LiquidityDelta))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.positionManager.GetPositionLiquidity(s.T().Context(), atBlock, NewGetPositionLiquidityRequest(s.contract, s.tokenID))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "position liquidity not readable at the receipt block")
	s.Require().Zero(liquidity.Sign())

	_, err = s.positionManager.OwnerOf(s.T().Context(), atBlock, erc721.NewOwnerOfRequest(s.contract, s.tokenID))
	s.Require().Error(err, "a burned position must have no owner")

	s.T().Logf("PositionManager %s Test17BurnPosition Result", s.contract)
	s.T().Logf("\tTokenID: %s burned", s.tokenID)
	s.T().Logf("\tModifyLiquidity delta: %s (want -%s)", mods[0].LiquidityDelta, remaining)
	s.T().Logf("\tLiquidity at receipt block: %s", liquidity)
	s.T().Logf("\tOwnerOf after burn: %s", err)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PositionManagerSuite) Test18NativePosition() {
	initTx, initHash, err := s.positionManager.InitializePool(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializePoolRequest(s.contract, s.nativeKey, s.sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, initTx.MaxFee())
	s.mu.Unlock()

	initReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, initHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(initReceipt.Status, "initializePool (native) transaction reverted")
	s.T().Logf("Test18NativePosition: initialized native pool (tx %s, block %d)", initHash, initReceipt.BlockNumber)

	mintData, err := EncodeActionMintPosition(NewMintPositionParams(s.nativeKey, s.tickLower, s.tickUpper, s.NativeMintLiquidity, s.NativeAmount, s.MintAmount1Max, s.from, nil))
	s.Require().NoError(err)
	settleData, err := EncodeActionSettlePair(s.nativeKey.Currency0, s.nativeKey.Currency1)
	s.Require().NoError(err)
	sweepData, err := EncodeActionSweep(s.nativeKey.Currency0, s.from)
	s.Require().NoError(err)
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())

	rejectedTx, rejectedHash, rejectedErr := s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()).WithValue(s.NativeAmount), NewModifyLiquiditiesRequest(s.contract, []byte{ActionMintPosition, ActionSettlePair, ActionSweep}, [][]byte{mintData, settleData, sweepData}, deadline, s.NativeAmount))
	s.Require().ErrorIs(rejectedErr, bind.ErrInvalidOptions)
	s.Require().Nil(rejectedTx)
	s.Require().Nil(rejectedHash)

	var sentTx core.Transaction
	var hash *types.Hash
	s.Require().Eventually(func() bool {
		sentTx, hash, err = s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionMintPosition, ActionSettlePair, ActionSweep}, [][]byte{mintData, settleData, sweepData}, deadline, s.NativeAmount))
		return !errors.Is(err, ErrPoolNotInitialized)
	}, 20*time.Second, s.ReceiptPollInterval, "native pool still not initialized on the estimating node")
	s.Require().NoError(err)
	s.Require().Zero(s.NativeAmount.Cmp(sentTx.(*core.DynamicFeeTx).Value))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test18NativePosition: sent modifyLiquidities [MINT_POSITION, SETTLE_PAIR, SWEEP] with %s wei (nonce %d, maxFee %s), tx %s", s.NativeAmount, sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidities (native mint) transaction reverted")

	var positionManagerLogs, poolManagerLogs []tx.Log
	for _, log := range receipt.Logs {
		if log.Address.Equal(s.contract) {
			positionManagerLogs = append(positionManagerLogs, log)
		}
		if log.Address.Equal(s.poolManager) {
			poolManagerLogs = append(poolManagerLogs, log)
		}
	}

	transfers, err := erc721.ExtractTransfers(positionManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	nativeTokenID := transfers[0].TokenId

	mods, err := ExtractModifyLiquidity(poolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().Zero(s.NativeMintLiquidity.Cmp(mods[0].LiquidityDelta))

	atBlock := fmt.Sprintf("0x%x", receipt.BlockNumber)

	var balanceHex string
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetBalance("balance", s.contract.String(), atBlock, &balanceHex)) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "PositionManager balance not readable at the receipt block")
	balance, ok := new(big.Int).SetString(strings.TrimPrefix(balanceHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(balance.Sign(), "SWEEP must refund every unused wei")

	burnData, err := EncodeActionBurnPosition(nativeTokenID, big.NewInt(0), big.NewInt(0), nil)
	s.Require().NoError(err)
	takeData, err := EncodeActionTakePair(s.nativeKey.Currency0, s.nativeKey.Currency1, s.from)
	s.Require().NoError(err)

	var burnTx core.Transaction
	var burnHash *types.Hash
	s.Require().Eventually(func() bool {
		burnTx, burnHash, err = s.positionManager.ModifyLiquidities(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquiditiesRequest(s.contract, []byte{ActionBurnPosition, ActionTakePair}, [][]byte{burnData, takeData}, deadline, nil))
		return err == nil || burnTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "native position still not visible on the estimating node")
	s.Require().NoError(err)
	s.Require().Zero(burnTx.(*core.DynamicFeeTx).Value.Sign())
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, burnTx.MaxFee())
	s.mu.Unlock()

	burnReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(burnReceipt.Status, "modifyLiquidities (native burn) transaction reverted")

	var burnPoolManagerLogs []tx.Log
	for _, log := range burnReceipt.Logs {
		if log.Address.Equal(s.poolManager) {
			burnPoolManagerLogs = append(burnPoolManagerLogs, log)
		}
	}
	burnMods, err := ExtractModifyLiquidity(burnPoolManagerLogs)
	s.Require().NoError(err)
	s.Require().Len(burnMods, 1)
	s.Require().Zero(new(big.Int).Neg(s.NativeMintLiquidity).Cmp(burnMods[0].LiquidityDelta))

	s.T().Logf("PositionManager %s Test18NativePosition Result", s.contract)
	s.T().Logf("\tNative pool: currency0=%s currency1=%s", s.nativeKey.Currency0, s.nativeKey.Currency1)
	s.T().Logf("\topts.Value rejected: %s", rejectedErr)
	s.T().Logf("\tTokenID: %s, minted liquidity %s with msg.value %s", nativeTokenID, mods[0].LiquidityDelta, sentTx.(*core.DynamicFeeTx).Value)
	s.T().Logf("\tPositionManager native balance after SWEEP: %s", balance)
	s.T().Logf("\tBurned: delta %s (tx %s, block %d, gas %d)", burnMods[0].LiquidityDelta, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
	s.T().Logf("\tMint tx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}
