package universalrouter

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"
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
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v2 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v2"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/weth"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type UniversalRouterSuite struct {
	suite.Suite

	RPCURL                     string
	PrivateKeyHex              string
	ChainID                    *big.Int
	UniversalRouterHex         string
	Permit2Hex                 string
	PoolManagerHex             string
	PoolModifyLiquidityTestHex string
	V2FactoryHex               string
	V3FactoryHex               string
	NFPMHex                    string
	WETHHex                    string
	MulticallHex               string
	NicksFactoryHex            string
	NativeSymbol               string
	NativeDecimals             int
	Timeout                    time.Duration
	ReceiptPollInterval        time.Duration
	ReceiptTimeout             time.Duration
	LimitBuffer                uint64
	ByteCode                   string
	Constructor                *abi.Function
	Salt                       *types.Hash
	MintValue                  *big.Int
	LiquidityDelta             *big.Int
	PermitAmount               *big.Int
	SwapAmountIn               *big.Int
	SwapAmountOut              *big.Int
	WrapAmount                 *big.Int
	PortionBips                *big.Int
	PortionFullPrecision       *big.Int
	NativeLiquidityDelta       *big.Int
	NativeLiquidityValue       *big.Int
	NativeSwapAmount           *big.Int
	V2LiquidityAmount          *big.Int
	V3LiquidityAmount          *big.Int

	client                  rpc.Client
	deployer                *deploy.NicksDeployer
	signer                  core.Signer
	from                    *types.Address
	router                  *types.Address
	permit2Contract         *types.Address
	poolManagerContract     *types.Address
	poolModifyLiquidityTest *types.Address
	wethContract            *types.Address
	multicall               *types.Address
	nicks                   *types.Address
	feeRecipient            *types.Address
	v3Pool                  *types.Address
	v2Pair                  *types.Address

	universalRouter *UniversalRouter
	permit2         *permit2.Permit2
	erc20           *erc20.ERC20

	currency0 *types.Address
	currency1 *types.Address
	key       *v4.PoolKey
	nativeKey *v4.PoolKey

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestUniversalRouterSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	uniswap := c.Addresses.Protocol.Uniswap
	if uniswap.UniversalRouter == "" || uniswap.V4.PoolManager == "" || uniswap.V4.PoolModifyLiquidityTest == "" {
		t.Skip("testenv: no Uniswap UniversalRouter/V4 PoolManager/PoolModifyLiquidityTest configured for this network")
	}

	suite.Run(t, &UniversalRouterSuite{
		RPCURL:                     c.RPCURL,
		PrivateKeyHex:              c.ActorPrivateKeyHex,
		ChainID:                    big.NewInt(int64(c.Constant.ChainID)),
		UniversalRouterHex:         uniswap.UniversalRouter,
		Permit2Hex:                 c.Addresses.Standard.Permit2,
		PoolManagerHex:             uniswap.V4.PoolManager,
		PoolModifyLiquidityTestHex: uniswap.V4.PoolModifyLiquidityTest,
		V2FactoryHex:               uniswap.V2.Factory,
		V3FactoryHex:               uniswap.V3.Factory,
		NFPMHex:                    uniswap.V3.NFPM,
		WETHHex:                    c.Addresses.Standard.WETH,
		MulticallHex:               c.Addresses.Standard.Multicall,
		NicksFactoryHex:            c.Addresses.Standard.NicksFactory,
		NativeSymbol:               c.Constant.NativeSymbol,
		NativeDecimals:             c.Constant.NativeDecimals,
		Timeout:                    10 * time.Minute,
		ReceiptPollInterval:        c.Constant.ReceiptPollInterval,
		ReceiptTimeout:             c.Constant.ReceiptTimeout,
		LimitBuffer:                130,
		ByteCode:                   testenv.ERC20ByteCodeHex,
		Constructor:                abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                       core.Keccak256([]byte("go-evmkit-universal-router")),
		MintValue:                  new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
		LiquidityDelta:             new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		PermitAmount:               new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		SwapAmountIn:               new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
		SwapAmountOut:              new(big.Int).Exp(big.NewInt(10), big.NewInt(14), nil),
		WrapAmount:                 big.NewInt(1_000_000_000),
		PortionBips:                big.NewInt(25),
		PortionFullPrecision:       new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil),
		NativeLiquidityDelta:       new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
		NativeLiquidityValue:       new(big.Int).Exp(big.NewInt(10), big.NewInt(14), nil),
		NativeSwapAmount:           new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil),
		V2LiquidityAmount:          new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		V3LiquidityAmount:          new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
	})
}

func (s *UniversalRouterSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.universalRouter = New(s.client)
	s.permit2 = permit2.New(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.router, err = types.NewAddressFromHex(s.UniversalRouterHex)
	s.Require().NoError(err)
	s.permit2Contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)
	s.poolManagerContract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.poolModifyLiquidityTest, err = types.NewAddressFromHex(s.PoolModifyLiquidityTestHex)
	s.Require().NoError(err)
	s.wethContract, err = types.NewAddressFromHex(s.WETHHex)
	s.Require().NoError(err)
	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)
	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: UniversalRouter %s, Permit2 %s, PoolManager %s, PoolModifyLiquidityTest %s, WETH %s", s.router, s.permit2Contract, s.poolManagerContract, s.poolModifyLiquidityTest, s.wethContract)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.feeRecipient = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-fee-recipient", runtime)).Bytes()[12:])
	s.T().Logf("SetupSuite: fee recipient %s", s.feeRecipient)

	names := [2]string{"evmkit universal router test token A", "evmkit universal router test token B"}
	symbols := [2]string{"EVURA", "EVURB"}
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
		builder.LimitBuffer(s.LimitBuffer)

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
		tokens[i] = token

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
		mintBuilder.LimitBuffer(s.LimitBuffer)
		mintTx, mintHash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, mintBuilder), token, calldata)
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintTx.MaxFee())
		s.mu.Unlock()

		mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s %s to %s (tx %s)", s.MintValue, symbols[i], s.from, mintHash)

		for _, spender := range []*types.Address{s.poolModifyLiquidityTest, s.permit2Contract} {
			approveBuilder := tx.NewDynamicFeeBuilder()
			approveBuilder.LimitBuffer(s.LimitBuffer)
			approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, approveBuilder), erc20.NewApproveRequest(token, spender, s.MintValue))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, approveTx.MaxFee())
			s.mu.Unlock()

			approveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "approve transaction reverted")
			s.T().Logf("SetupSuite: approved %s for %s (tx %s)", spender, symbols[i], approveHash)
		}
	}

	s.currency0, s.currency1 = tokens[0], tokens[1]
	if s.currency0.Big().Cmp(s.currency1.Big()) > 0 {
		s.currency0, s.currency1 = s.currency1, s.currency0
	}
	s.T().Logf("SetupSuite: sorted currency0=%s currency1=%s", s.currency0, s.currency1)

	s.key = v4.NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	initBuilder := tx.NewDynamicFeeBuilder()
	initBuilder.LimitBuffer(s.LimitBuffer)
	initTx, initHash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, initBuilder), NewExecuteRequest(s.router, NewPlan().Add(NewV4InitializePoolInput(s.key, sqrtPriceX96)), big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, initTx.MaxFee())
	s.mu.Unlock()

	initReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, initHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(initReceipt.Status, "V4_INITIALIZE_POOL transaction reverted")

	inits, err := v4.ExtractInitialize(initReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(inits, 1)
	s.Require().True(inits[0].Contract.Equal(s.poolManagerContract))
	s.Require().True(inits[0].Currency0.Equal(s.currency0))
	s.Require().True(inits[0].Currency1.Equal(s.currency1))
	s.Require().Zero(sqrtPriceX96.Cmp(inits[0].SqrtPriceX96))
	s.T().Logf("SetupSuite: initialized pool via UniversalRouter V4_INITIALIZE_POOL (tx %s, block %d)", initHash, initReceipt.BlockNumber)

	modifyLiquidityTestFn := abi.NewFunction("modifyLiquidity", abi.NewTypes(poolKeyTuple, abi.Tuple(abi.Int24, abi.Int24, abi.Int256, abi.Bytes32), abi.Bytes, abi.Bool, abi.Bool), abi.NewTypes(abi.Bytes32))

	var liquiditySalt [32]byte
	calldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.LiquidityDelta, liquiditySalt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	var liquidityTx core.Transaction
	var liquidityHash *types.Hash
	s.Require().Eventually(func() bool {
		liquidityBuilder := tx.NewDynamicFeeBuilder()
		liquidityBuilder.LimitBuffer(s.LimitBuffer)
		liquidityTx, liquidityHash, err = bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, liquidityBuilder), s.poolModifyLiquidityTest, calldata)
		return err == nil || liquidityTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "pool still not initialized on the estimating node")
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, liquidityTx.MaxFee())
	s.mu.Unlock()

	liquidityReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, liquidityHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(liquidityReceipt.Status, "modifyLiquidity transaction reverted")
	s.T().Logf("SetupSuite: added liquidity %s in [-600, 600] (tx %s, block %d)", s.LiquidityDelta, liquidityHash, liquidityReceipt.BlockNumber)

	s.nativeKey = v4.NewPoolKey(types.ZeroAddress(), s.currency0, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())

	nativeInitBuilder := tx.NewDynamicFeeBuilder()
	nativeInitBuilder.LimitBuffer(s.LimitBuffer)
	nativeInitTx, nativeInitHash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, nativeInitBuilder), NewExecuteRequest(s.router, NewPlan().Add(NewV4InitializePoolInput(s.nativeKey, sqrtPriceX96)), big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, nativeInitTx.MaxFee())
	s.mu.Unlock()

	nativeInitReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, nativeInitHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(nativeInitReceipt.Status, "native V4_INITIALIZE_POOL transaction reverted")
	s.T().Logf("SetupSuite: initialized native pool ETH/%s via UniversalRouter (tx %s, block %d)", s.currency0, nativeInitHash, nativeInitReceipt.BlockNumber)

	nativeCalldata, err := modifyLiquidityTestFn.EncodeCall(
		[]any{s.nativeKey.Currency0, s.nativeKey.Currency1, s.nativeKey.Fee, s.nativeKey.TickSpacing, s.nativeKey.Hooks},
		[]any{big.NewInt(-600), big.NewInt(600), s.NativeLiquidityDelta, liquiditySalt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	var nativeLiquidityTx core.Transaction
	var nativeLiquidityHash *types.Hash
	s.Require().Eventually(func() bool {
		nativeLiquidityBuilder := tx.NewDynamicFeeBuilder()
		nativeLiquidityBuilder.LimitBuffer(s.LimitBuffer)
		nativeLiquidityTx, nativeLiquidityHash, err = bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, nativeLiquidityBuilder).WithValue(s.NativeLiquidityValue), s.poolModifyLiquidityTest, nativeCalldata)
		return err == nil || nativeLiquidityTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "native pool still not initialized on the estimating node")
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, nativeLiquidityTx.MaxFee())
	s.mu.Unlock()

	nativeLiquidityReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, nativeLiquidityHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(nativeLiquidityReceipt.Status, "native modifyLiquidity transaction reverted")
	s.T().Logf("SetupSuite: added native liquidity %s with %s wei (tx %s, block %d)", s.NativeLiquidityDelta, s.NativeLiquidityValue, nativeLiquidityHash, nativeLiquidityReceipt.BlockNumber)

	if s.V3FactoryHex != "" && s.NFPMHex != "" {
		v3FactoryContract, err := types.NewAddressFromHex(s.V3FactoryHex)
		s.Require().NoError(err)
		nfpmContract, err := types.NewAddressFromHex(s.NFPMHex)
		s.Require().NoError(err)

		for _, token := range []*types.Address{s.currency0, s.currency1} {
			approveBuilder := tx.NewDynamicFeeBuilder()
			approveBuilder.LimitBuffer(s.LimitBuffer)
			approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, approveBuilder), erc20.NewApproveRequest(token, nfpmContract, s.MintValue))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, approveTx.MaxFee())
			s.mu.Unlock()

			approveReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, approveHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(approveReceipt.Status, "NFPM approve transaction reverted")
		}

		nfpm := v3.NewNFPM(s.client)
		createBuilder := tx.NewDynamicFeeBuilder()
		createBuilder.LimitBuffer(s.LimitBuffer)
		createTx, createHash, err := nfpm.CreateAndInitializePoolIfNecessary(s.T().Context(), bind.NewTransactOptions(s.signer, createBuilder), v3.NewCreateAndInitializePoolIfNecessaryRequest(nfpmContract, s.currency0, s.currency1, big.NewInt(3000), sqrtPriceX96))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, createTx.MaxFee())
		s.mu.Unlock()

		createReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, createHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(createReceipt.Status, "createAndInitializePoolIfNecessary transaction reverted")

		mintParams := v3.NewMintParams(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(-887220), big.NewInt(887220), s.V3LiquidityAmount, s.V3LiquidityAmount, big.NewInt(0), big.NewInt(0), s.from, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
		var mintTx core.Transaction
		var mintHash *types.Hash
		s.Require().Eventually(func() bool {
			mintBuilder := tx.NewDynamicFeeBuilder()
			mintBuilder.LimitBuffer(s.LimitBuffer)
			mintTx, mintHash, err = nfpm.Mint(s.T().Context(), bind.NewTransactOptions(s.signer, mintBuilder), v3.NewNFPMMintRequest(nfpmContract, mintParams, nil))
			return err == nil || mintTx != nil
		}, 20*time.Second, s.ReceiptPollInterval, "V3 pool still not initialized on the estimating node")
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, mintTx.MaxFee())
		s.mu.Unlock()

		mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "NFPM mint transaction reverted")

		s.Require().Eventually(func() bool {
			s.v3Pool, err = v3.NewFactory(s.client).GetPool(s.T().Context(), nil, v3.NewGetPoolRequest(v3FactoryContract, s.currency0, s.currency1, big.NewInt(3000)))
			return err == nil && !s.v3Pool.Equal(types.ZeroAddress())
		}, 20*time.Second, s.ReceiptPollInterval, "V3 pool not visible on the factory")
		s.T().Logf("SetupSuite: V3 pool %s full-range liquidity %s/%s (tx %s, block %d)", s.v3Pool, s.V3LiquidityAmount, s.V3LiquidityAmount, mintHash, mintReceipt.BlockNumber)
	}

	if s.V2FactoryHex != "" {
		v2FactoryContract, err := types.NewAddressFromHex(s.V2FactoryHex)
		s.Require().NoError(err)

		var routerCodeHex string
		s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetCode("code", s.router.String(), rpc.BlockTagLatest, &routerCodeHex)))
		if !strings.Contains(strings.ToLower(routerCodeHex), strings.ToLower(strings.TrimPrefix(v2FactoryContract.String(), "0x"))) {
			s.T().Logf("SetupSuite: V2 factory %s is not the one UniversalRouter %s is bound to; V2 tests will skip", v2FactoryContract, s.router)
		} else {
			createBuilder := tx.NewDynamicFeeBuilder()
			createBuilder.LimitBuffer(s.LimitBuffer)
			createTx, createHash, err := v2.NewFactory(s.client).CreatePair(s.T().Context(), bind.NewTransactOptions(s.signer, createBuilder), v2.NewCreatePairRequest(v2FactoryContract, s.currency0, s.currency1))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, createTx.MaxFee())
			s.mu.Unlock()

			createReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, createHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(createReceipt.Status, "createPair transaction reverted")

			created, err := v2.ExtractPairCreated(createReceipt.Logs)
			s.Require().NoError(err)
			s.Require().Len(created, 1)
			s.v2Pair = created[0].Pair

			for _, token := range []*types.Address{s.currency0, s.currency1} {
				transferBuilder := tx.NewDynamicFeeBuilder()
				transferBuilder.LimitBuffer(s.LimitBuffer)
				transferTx, transferHash, err := s.erc20.Transfer(s.T().Context(), bind.NewTransactOptions(s.signer, transferBuilder), erc20.NewTransferRequest(token, s.v2Pair, s.V2LiquidityAmount))
				s.Require().NoError(err)
				s.mu.Lock()
				s.totalFee.Add(s.totalFee, transferTx.MaxFee())
				s.mu.Unlock()

				transferReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
				s.Require().NoError(err)
				s.Require().True(transferReceipt.Status, "transfer to pair reverted")
			}

			var mintTx core.Transaction
			var mintHash *types.Hash
			s.Require().Eventually(func() bool {
				mintBuilder := tx.NewDynamicFeeBuilder()
				mintBuilder.LimitBuffer(s.LimitBuffer)
				mintTx, mintHash, err = v2.NewPair(s.client).Mint(s.T().Context(), bind.NewTransactOptions(s.signer, mintBuilder), v2.NewMintRequest(s.v2Pair, s.from))
				return err == nil || mintTx != nil
			}, 20*time.Second, s.ReceiptPollInterval, "pair transfers still not visible on the estimating node")
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, mintTx.MaxFee())
			s.mu.Unlock()

			mintReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(mintReceipt.Status, "pair mint transaction reverted")
			s.T().Logf("SetupSuite: V2 pair %s reserves %s/%s (tx %s, block %d)", s.v2Pair, s.V2LiquidityAmount, s.V2LiquidityAmount, mintHash, mintReceipt.BlockNumber)
		}
	}
}

func (s *UniversalRouterSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	if s.currency0 != nil && s.currency1 != nil {
		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, holder := range []*types.Address{s.from, s.feeRecipient} {
			balances, err := s.erc20.AddressBalances(ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, holder, []*types.Address{s.currency0, s.currency1}, false))
			s.Require().NoError(err)

			var lastBlock uint64
			for _, tb := range balances.Tokens {
				if tb.Amount == nil || tb.Amount.Sign() == 0 {
					continue
				}

				calldata, err := tokenBurnFn.EncodeCall(holder, tb.Amount)
				s.Require().NoError(err)

				burnBuilder := tx.NewDynamicFeeBuilder()
				burnBuilder.LimitBuffer(s.LimitBuffer)
				burnTx, burnHash, err := bind.Transact(ctx, s.client, bind.NewTransactOptions(s.signer, burnBuilder), tb.Token, calldata)
				s.Require().NoError(err)
				s.mu.Lock()
				s.totalFee.Add(s.totalFee, burnTx.MaxFee())
				s.mu.Unlock()

				burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
				s.Require().NoError(err)
				s.Require().True(burnReceipt.Status, "burn transaction reverted")
				lastBlock = max(lastBlock, burnReceipt.BlockNumber)

				s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s, block %d, gas %d)", tb.Amount, tb.Token, holder, burnHash, burnReceipt.BlockNumber, burnReceipt.GasUsed)
			}

			if lastBlock == 0 {
				continue
			}

			var remaining *erc20.AddressBalances
			s.Require().Eventually(func() bool {
				remaining, err = s.erc20.AddressBalances(ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", lastBlock)), erc20.NewAddressBalancesRequest(s.multicall, holder, []*types.Address{s.currency0, s.currency1}, false))
				return err == nil
			}, 20*time.Second, s.ReceiptPollInterval, "balances not readable after burning")
			for _, tb := range remaining.Tokens {
				s.Require().Zero(tb.Amount.Sign(), "%s still holds %s of %s after burning", holder, tb.Amount, tb.Token)
			}
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

		s.T().Logf("UniversalRouterSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *UniversalRouterSuite) Test01InitializePoolAgainMapsExecutionFailed() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)

	err := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().Add(NewV4InitializePoolInput(s.key, sqrtPriceX96)), nil, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrExecutionFailed)
	s.Require().ErrorIs(err, v4.ErrPoolAlreadyInitialized)
	s.Require().NotErrorIs(err, ErrUnknown)

	failed, ok := errors.AsType[*ExecutionFailedError](err)
	s.Require().True(ok)
	s.Require().Zero(failed.CommandIndex.Sign())

	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	s.Require().True(ok)
	s.Require().NotEmpty(rpcErr.Data)

	allowRevertErr := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().AddAllowRevert(NewV4InitializePoolInput(s.key, sqrtPriceX96)), nil, nil))
	s.Require().NoError(allowRevertErr)

	s.T().Logf("UniversalRouter Test01InitializePoolAgainMapsExecutionFailed Result")
	s.T().Logf("\tError: %s", err)
	s.T().Logf("\tCommandIndex: %s, Message: 0x%x", failed.CommandIndex, failed.Message)
}

func (s *UniversalRouterSuite) Test02WrapUnwrapRoundTrip() {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), s.WrapAmount)).Add(NewUnwrapWETHInput(MsgSender(), s.WrapAmount))

	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, s.WrapAmount)))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), s.WrapAmount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02WrapUnwrapRoundTrip: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1)
	s.Require().True(deposits[0].Contract.Equal(s.wethContract))
	s.Require().True(deposits[0].Dst.Equal(s.router))
	s.Require().Zero(s.WrapAmount.Cmp(deposits[0].Wad))

	withdrawals, err := weth.ExtractWithdrawals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(withdrawals, 1)
	s.Require().True(withdrawals[0].Contract.Equal(s.wethContract))
	s.Require().True(withdrawals[0].Src.Equal(s.router))
	s.Require().GreaterOrEqual(withdrawals[0].Wad.Cmp(s.WrapAmount), 0)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	var routerWETH *big.Int
	s.Require().Eventually(func() bool {
		routerWETH, err = s.erc20.BalanceOf(s.T().Context(), atBlock, erc20.NewBalanceOfRequest(s.wethContract, s.router))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "WETH balance not readable at the receipt block")
	s.Require().Zero(routerWETH.Sign(), "router kept WETH after the round trip")

	s.T().Logf("UniversalRouter Test02WrapUnwrapRoundTrip Result")
	s.T().Logf("\tWrapped/Unwrapped: %s / %s", deposits[0].Wad, withdrawals[0].Wad)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test03Permit2PermitAndV4SwapExactInSingle() {
	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, expiration, allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.key, true, s.SwapAmountIn, big.NewInt(1), big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency1, big.NewInt(1))
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams}))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03Permit2PermitAndV4SwapExactInSingle: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v4.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolManagerContract))
	s.Require().True(swaps[0].Sender.Equal(s.router))
	s.Require().Zero(new(big.Int).Neg(s.SwapAmountIn).Cmp(swaps[0].Amount0), "amount0 should equal the exact input specified")
	s.Require().Positive(swaps[0].Amount1.Sign())

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))

	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.currency0, s.currency1}, false))
	s.Require().NoError(err)

	s.Require().Zero(new(big.Int).Add(before.Tokens[0].Amount, swaps[0].Amount0).Cmp(after.Tokens[0].Amount), "currency0 balance didn't move by exactly amount0")
	s.Require().Zero(new(big.Int).Add(before.Tokens[1].Amount, swaps[0].Amount1).Cmp(after.Tokens[1].Amount), "currency1 balance didn't move by exactly amount1")

	allowanceAfter, err := s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(s.PermitAmount, s.SwapAmountIn).Cmp(allowanceAfter.Amount))
	s.Require().Zero(expiration.Cmp(allowanceAfter.Expiration))
	s.Require().Zero(new(big.Int).Add(allowance.Nonce, big.NewInt(1)).Cmp(allowanceAfter.Nonce))

	var minedTx map[string]any
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetTransactionByHash("tx", hash.String(), &minedTx)) == nil && minedTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "mined transaction not readable")
	minedInput, err := encoding.Hex.Decode(minedTx["input"].(string))
	s.Require().NoError(err)
	decoded, err := DecodeExecuteCall(minedInput)
	s.Require().NoError(err)
	s.Require().Nil(decoded.Signed)
	s.Require().NotNil(decoded.Deadline)
	wantCommands, wantInputs, err := plan.Encode()
	s.Require().NoError(err)
	gotCommands, gotInputs, err := decoded.Plan.Encode()
	s.Require().NoError(err)
	s.Require().Equal(wantCommands, gotCommands)
	s.Require().Equal(wantInputs, gotInputs)
	s.Require().Len(decoded.Plan.Steps, 2)
	s.Require().Equal(CommandPermit2Permit, decoded.Plan.Steps[0].Input.Command())
	s.Require().Equal(CommandV4Swap, decoded.Plan.Steps[1].Input.Command())
	decodedPermit, ok := decoded.Plan.Steps[0].Input.(*Permit2PermitInput)
	s.Require().True(ok)
	s.Require().Equal(sig.LegacyBytes(), decodedPermit.Signature)

	s.T().Logf("UniversalRouter Test03Permit2PermitAndV4SwapExactInSingle Result")
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tPermit2 allowance: amount %s, nonce %s -> %s", allowanceAfter.Amount, allowance.Nonce, allowanceAfter.Nonce)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test04Permit2PermitBatchAndV4SwapExactOutSingle() {
	allowance0, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	allowance1, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	permitBatch := permit2.NewPermitBatch([]permit2.PermitDetails{
		*permit2.NewPermitDetails(s.currency0, s.PermitAmount, expiration, allowance0.Nonce),
		*permit2.NewPermitDetails(s.currency1, s.PermitAmount, expiration, allowance1.Nonce),
	}, s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitBatch.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	swapParams, err := v4.EncodeActionSwapExactOutSingle(v4.NewExactOutputSingleParams(s.key, false, s.SwapAmountOut, s.SwapAmountIn, big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency1, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency0, s.SwapAmountOut)
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitBatchInput(permitBatch, sig.LegacyBytes())).
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactOutSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams}))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04Permit2PermitBatchAndV4SwapExactOutSingle: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v4.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Sender.Equal(s.router))
	s.Require().Zero(s.SwapAmountOut.Cmp(swaps[0].Amount0), "amount0 should equal the exact output requested")
	s.Require().Negative(swaps[0].Amount1.Sign())
	s.Require().LessOrEqual(new(big.Int).Neg(swaps[0].Amount1).Cmp(s.SwapAmountIn), 0, "paid more than amountInMaximum")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	var allowanceAfter1 *permit2.PackedAllowance
	s.Require().Eventually(func() bool {
		allowanceAfter1, err = s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "allowance not readable at the receipt block")
	s.Require().Zero(new(big.Int).Add(s.PermitAmount, swaps[0].Amount1).Cmp(allowanceAfter1.Amount))
	s.Require().Zero(new(big.Int).Add(allowance1.Nonce, big.NewInt(1)).Cmp(allowanceAfter1.Nonce))

	allowanceAfter0, err := s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	s.Require().Zero(s.PermitAmount.Cmp(allowanceAfter0.Amount))
	s.Require().Zero(new(big.Int).Add(allowance0.Nonce, big.NewInt(1)).Cmp(allowanceAfter0.Nonce))

	s.T().Logf("UniversalRouter Test04Permit2PermitBatchAndV4SwapExactOutSingle Result")
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tPermit2 allowance currency1: %s (nonce %s)", allowanceAfter1.Amount, allowanceAfter1.Nonce)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test05SwapToRouterPayPortionAndSweep() {
	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.key, true, s.SwapAmountIn, big.NewInt(1), big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTake(s.currency1, v4.AddressThis(), v4.OpenDelta())
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTake}, [][]byte{swapParams, settleParams, takeParams})).
		Add(NewPayPortionInput(s.currency1, s.feeRecipient, s.PortionBips)).
		Add(NewSweepInput(s.currency1, MsgSender(), big.NewInt(0)))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05SwapToRouterPayPortionAndSweep: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v4.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	out := swaps[0].Amount1
	s.Require().Positive(out.Sign())
	fee := new(big.Int).Div(new(big.Int).Mul(out, s.PortionBips), big.NewInt(10_000))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))

	var recipientAfter *big.Int
	s.Require().Eventually(func() bool {
		recipientAfter, err = s.erc20.BalanceOf(s.T().Context(), atBlock, erc20.NewBalanceOfRequest(s.currency1, s.feeRecipient))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	recipientBefore, err := s.erc20.BalanceOf(s.T().Context(), beforeAt, erc20.NewBalanceOfRequest(s.currency1, s.feeRecipient))
	s.Require().NoError(err)
	userAfter, err := s.erc20.BalanceOf(s.T().Context(), atBlock, erc20.NewBalanceOfRequest(s.currency1, s.from))
	s.Require().NoError(err)
	userBefore, err := s.erc20.BalanceOf(s.T().Context(), beforeAt, erc20.NewBalanceOfRequest(s.currency1, s.from))
	s.Require().NoError(err)
	routerAfter, err := s.erc20.BalanceOf(s.T().Context(), atBlock, erc20.NewBalanceOfRequest(s.currency1, s.router))
	s.Require().NoError(err)

	s.Require().Zero(recipientBefore.Sign(), "fresh fee recipient held currency1 before the swap")
	s.Require().Zero(fee.Cmp(recipientAfter), "fee recipient didn't receive exactly the portion")
	s.Require().Zero(new(big.Int).Sub(out, fee).Cmp(new(big.Int).Sub(userAfter, userBefore)), "user didn't receive the swept remainder")
	s.Require().Zero(routerAfter.Sign(), "router kept currency1 after the sweep")

	s.T().Logf("UniversalRouter Test05SwapToRouterPayPortionAndSweep Result")
	s.T().Logf("\tOut: %s, portion (%s bips): %s, swept: %s", out, s.PortionBips, fee, new(big.Int).Sub(userAfter, userBefore))
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test06SubPlanFailureNestsExecutionFailed() {
	huge := new(big.Int).Lsh(big.NewInt(1), 255)
	subPlan := NewPlan().
		Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0))).
		Add(NewBalanceCheckERC20Input(s.from, s.currency1, huge))
	plan := NewPlan().
		Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0))).
		Add(NewExecuteSubPlanInput(subPlan))

	err := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrExecutionFailed)
	s.Require().ErrorIs(err, ErrBalanceTooLow)
	s.Require().NotErrorIs(err, ErrUnknown)
	s.Require().Contains(err.Error(), "execution failed: command 1: execution failed: command 1: balance too low")

	outer, ok := errors.AsType[*ExecutionFailedError](err)
	s.Require().True(ok)
	s.Require().Zero(big.NewInt(1).Cmp(outer.CommandIndex))

	allowRevertPlan := NewPlan().
		Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0))).
		AddAllowRevert(NewExecuteSubPlanInput(subPlan))
	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, allowRevertPlan, nil, nil)))

	s.T().Logf("UniversalRouter Test06SubPlanFailureNestsExecutionFailed Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test07V4SlippageMapsThroughRouter() {
	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.key, true, s.SwapAmountIn, s.SwapAmountIn, big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency1, big.NewInt(0))
	s.Require().NoError(err)
	plan := NewPlan().Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams}))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, v4.ErrV4TooLittleReceived)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().ErrorIs(err, v4.ErrV4TooLittleReceived)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	s.T().Logf("UniversalRouter Test07V4SlippageMapsThroughRouter Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test08DeadlinePassedMaps() {
	plan := NewPlan().Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0)))

	err := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, big.NewInt(1), nil))
	s.Require().ErrorIs(err, ErrTransactionDeadlinePassed)
	s.Require().NotErrorIs(err, ErrUnknown)

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(1), nil))
	s.Require().ErrorIs(err, ErrTransactionDeadlinePassed)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil)))

	s.T().Logf("UniversalRouter Test08DeadlinePassedMaps Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test09ExecuteSignedConsumesNonce() {
	var nonce, intent, data [32]byte
	copy(nonce[:], core.Keccak256(fmt.Appendf(nil, "%s-nonce-%d", s.Salt, time.Now().UnixNano())).Bytes())
	copy(intent[:], core.Keccak256([]byte("go-evmkit-universal-router-intent")).Bytes())
	copy(data[:], core.Keccak256([]byte("go-evmkit-universal-router-data")).Bytes())
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), s.WrapAmount)).Add(NewUnwrapWETHInput(MsgSender(), s.WrapAmount))

	sig, err := NewExecuteSignedMessage(plan, intent, data, s.from, nonce, deadline).Sign(NewUniversalRouterDomain(s.ChainID, s.router), s.signer)
	s.Require().NoError(err)
	req := NewExecuteSignedRequest(s.router, plan, intent, data, true, nonce, sig.LegacyBytes(), deadline, s.WrapAmount)

	used, err := s.universalRouter.NoncesUsed(s.T().Context(), nil, NewNoncesUsedRequest(s.router, s.from, nonce))
	s.Require().NoError(err)
	s.Require().False(used)

	s.Require().NoError(s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), req))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.ExecuteSigned(s.T().Context(), bind.NewTransactOptions(s.signer, builder), req)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09ExecuteSignedConsumesNonce: sent executeSigned (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "executeSigned transaction reverted")

	deposits, err := weth.ExtractDeposits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(deposits, 1)
	s.Require().Zero(s.WrapAmount.Cmp(deposits[0].Wad))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	s.Require().Eventually(func() bool {
		used, err = s.universalRouter.NoncesUsed(s.T().Context(), atBlock, NewNoncesUsedRequest(s.router, s.from, nonce))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "noncesUsed not readable at the receipt block")
	s.Require().True(used)
	usedBefore, err := s.universalRouter.NoncesUsed(s.T().Context(), beforeAt, NewNoncesUsedRequest(s.router, s.from, nonce))
	s.Require().NoError(err)
	s.Require().False(usedBefore)

	replayErr := s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from).WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), req)
	s.Require().ErrorIs(replayErr, ErrNonceAlreadyUsed)
	s.Require().NotErrorIs(replayErr, ErrUnknown)

	routeCtx, err := s.universalRouter.SignedRouteContext(s.T().Context(), atBlock, NewSignedRouteContextRequest(s.router))
	s.Require().NoError(err)
	s.Require().True(routeCtx.Signer.Equal(types.ZeroAddress()))
	s.Require().Equal([32]byte{}, routeCtx.Intent)
	s.Require().Equal([32]byte{}, routeCtx.Data)

	var minedTx map[string]any
	s.Require().Eventually(func() bool {
		return s.client.Call(s.T().Context(), rpc.ETHGetTransactionByHash("tx", hash.String(), &minedTx)) == nil && minedTx != nil
	}, 20*time.Second, s.ReceiptPollInterval, "mined transaction not readable")
	minedInput, err := encoding.Hex.Decode(minedTx["input"].(string))
	s.Require().NoError(err)
	minedValue, ok := new(big.Int).SetString(strings.TrimPrefix(minedTx["value"].(string), "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(s.WrapAmount.Cmp(minedValue))
	decoded, err := DecodeExecuteCall(minedInput)
	s.Require().NoError(err)
	s.Require().NotNil(decoded.Signed)
	s.Require().Zero(deadline.Cmp(decoded.Deadline))
	s.Require().Equal(intent, decoded.Signed.Intent)
	s.Require().Equal(data, decoded.Signed.Data)
	s.Require().True(decoded.Signed.VerifySender)
	s.Require().Equal(nonce, decoded.Signed.Nonce)
	s.Require().Equal(sig.LegacyBytes(), decoded.Signed.Signature)
	recomputed, err := NewExecuteSignedMessage(decoded.Plan, decoded.Signed.Intent, decoded.Signed.Data, s.from, decoded.Signed.Nonce, decoded.Deadline).Hash(NewUniversalRouterDomain(s.ChainID, s.router))
	s.Require().NoError(err)
	original, err := NewExecuteSignedMessage(plan, intent, data, s.from, nonce, deadline).Hash(NewUniversalRouterDomain(s.ChainID, s.router))
	s.Require().NoError(err)
	s.Require().Equal(original.String(), recomputed.String())

	s.T().Logf("UniversalRouter Test09ExecuteSignedConsumesNonce Result")
	s.T().Logf("\tNonce 0x%x used: %t -> %t", nonce, usedBefore, used)
	s.T().Logf("\tReplay error: %s", replayErr)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test10ExecuteSignedUncheckedNonceReplays() {
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	plan := NewPlan().Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0)))

	sig, err := NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, types.ZeroAddress(), UncheckedNonce(), deadline).Sign(NewUniversalRouterDomain(s.ChainID, s.router), s.signer)
	s.Require().NoError(err)
	req := NewExecuteSignedRequest(s.router, plan, [32]byte{}, [32]byte{}, false, UncheckedNonce(), sig.LegacyBytes(), deadline, nil)

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.ExecuteSigned(s.T().Context(), bind.NewTransactOptions(s.signer, builder), req)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test10ExecuteSignedUncheckedNonceReplays: sent executeSigned (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "executeSigned transaction reverted")

	atBlock := fmt.Sprintf("0x%x", receipt.BlockNumber)
	s.Require().Eventually(func() bool {
		return s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from).WithBlock(atBlock), req) == nil
	}, 20*time.Second, s.ReceiptPollInterval, "unchecked-nonce route not replayable at the receipt block")

	used, err := s.universalRouter.NoncesUsed(s.T().Context(), bind.NewCallOptions().WithBlock(atBlock), NewNoncesUsedRequest(s.router, s.from, UncheckedNonce()))
	s.Require().NoError(err)
	s.Require().False(used)

	relayer, err := types.NewPrivateKeyFromBig(big.NewInt(0x5ee))
	s.Require().NoError(err)
	s.Require().NoError(s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(core.PubkeyToAddress(relayer.PublicKey())).WithBlock(atBlock), req))

	s.T().Logf("UniversalRouter Test10ExecuteSignedUncheckedNonceReplays Result")
	s.T().Logf("\tNoncesUsed(unchecked): %t", used)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test11ExecuteSignedRejectsShortSignature() {
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	plan := NewPlan().Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0)))

	err := s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteSignedRequest(s.router, plan, [32]byte{}, [32]byte{}, false, UncheckedNonce(), make([]byte, 64), deadline, nil))
	s.Require().ErrorIs(err, ErrECDSAInvalidSignatureLength)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test11ExecuteSignedRejectsShortSignature Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test12ExecuteSignedRejectsZeroSignature() {
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	plan := NewPlan().Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0)))

	err := s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteSignedRequest(s.router, plan, [32]byte{}, [32]byte{}, false, UncheckedNonce(), make([]byte, 65), deadline, nil))
	s.Require().ErrorIs(err, ErrECDSAInvalidSignature)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test12ExecuteSignedRejectsZeroSignature Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test13ExecuteSignedRejectsExpiredDeadline() {
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	plan := NewPlan().Add(NewBalanceCheckERC20Input(s.from, s.currency0, big.NewInt(0)))

	sig, err := NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, types.ZeroAddress(), UncheckedNonce(), deadline).Sign(NewUniversalRouterDomain(s.ChainID, s.router), s.signer)
	s.Require().NoError(err)

	err = s.universalRouter.CallExecuteSigned(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteSignedRequest(s.router, plan, [32]byte{}, [32]byte{}, false, UncheckedNonce(), sig.LegacyBytes(), big.NewInt(1), nil))
	s.Require().ErrorIs(err, ErrTransactionDeadlinePassed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test13ExecuteSignedRejectsExpiredDeadline Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test14Permit2PermitReplayMapsInvalidNonce() {
	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)

	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency1, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)
	plan := NewPlan().Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes()))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14Permit2PermitReplayMapsInvalidNonce: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	atBlock := bind.NewCallOptions().WithFrom(s.from).WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	var replayErr error
	s.Require().Eventually(func() bool {
		replayErr = s.universalRouter.CallExecute(s.T().Context(), atBlock, NewExecuteRequest(s.router, plan, nil, nil))
		return replayErr != nil && !errors.Is(replayErr, ErrUnknown) && errors.Is(replayErr, ErrExecutionFailed)
	}, 20*time.Second, s.ReceiptPollInterval, "replay not evaluated at the receipt block: %v", replayErr)
	s.Require().ErrorIs(replayErr, permit2.ErrInvalidNonce)

	failed, ok := errors.AsType[*ExecutionFailedError](replayErr)
	s.Require().True(ok)
	s.Require().Zero(failed.CommandIndex.Sign())

	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), atBlock, NewExecuteRequest(s.router, NewPlan().AddAllowRevert(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())), nil, nil)))

	s.T().Logf("UniversalRouter Test14Permit2PermitReplayMapsInvalidNonce Result")
	s.T().Logf("\tReplay error: %s", replayErr)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test15SweepTokenMinimumMapsInsufficientToken() {
	huge := new(big.Int).Lsh(big.NewInt(1), 159)

	err := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().Add(NewSweepInput(s.currency1, MsgSender(), huge)), nil, nil))
	s.Require().ErrorIs(err, ErrInsufficientToken)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	allowRevertErr := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().AddAllowRevert(NewSweepInput(s.currency1, MsgSender(), huge)), nil, nil))
	s.Require().ErrorIs(allowRevertErr, ErrInsufficientToken, "SWEEP reverts inside dispatch, so the allow-revert flag cannot absorb it")
	s.Require().NotErrorIs(allowRevertErr, ErrExecutionFailed)

	overflowErr := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().Add(NewSweepInput(s.currency1, MsgSender(), new(big.Int).Lsh(big.NewInt(1), 200))), nil, nil))
	s.Require().ErrorIs(overflowErr, bind.ErrInvalidRequest)
	s.Require().ErrorContains(overflowErr, "exceeds uint160")

	s.T().Logf("UniversalRouter Test15SweepTokenMinimumMapsInsufficientToken Result")
	s.T().Logf("\tError: %s", err)
	s.T().Logf("\tAllow-revert error: %s", allowRevertErr)
	s.T().Logf("\tOverflow error: %s", overflowErr)
}

func (s *UniversalRouterSuite) Test16SweepNativeMinimumMapsInsufficientETH() {
	huge := new(big.Int).Lsh(big.NewInt(1), 159)

	err := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().Add(NewSweepInput(types.ZeroAddress(), MsgSender(), huge)), nil, nil))
	s.Require().ErrorIs(err, ErrInsufficientETH)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test16SweepNativeMinimumMapsInsufficientETH Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test17ExpiredPermit2AllowanceMapsThroughV4() {
	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)

	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency1, s.PermitAmount, big.NewInt(1), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.key, false, s.SwapAmountIn, big.NewInt(1), big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency1, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency0, big.NewInt(1))
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams}))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().ErrorIs(err, permit2.ErrAllowanceExpired)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	allowanceAfter, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)
	s.Require().Zero(allowance.Nonce.Cmp(allowanceAfter.Nonce), "eth_call must not consume the permit nonce")

	s.T().Logf("UniversalRouter Test17ExpiredPermit2AllowanceMapsThroughV4 Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test18Permit2TransfersAndPayments() {
	allowance0, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	allowance1, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	permitBatch := permit2.NewPermitBatch([]permit2.PermitDetails{
		*permit2.NewPermitDetails(s.currency0, s.PermitAmount, expiration, allowance0.Nonce),
		*permit2.NewPermitDetails(s.currency1, s.PermitAmount, expiration, allowance1.Nonce),
	}, s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitBatch.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	pulled0 := big.NewInt(1_000_000)
	transferred0 := big.NewInt(100_000)
	pulled1ToRouter := big.NewInt(300_000)
	pulled1ToRecipient := big.NewInt(70_000)
	portion0 := new(big.Int).Div(new(big.Int).Mul(new(big.Int).Sub(pulled0, transferred0), s.PortionFullPrecision), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	plan := NewPlan().
		Add(NewPermit2PermitBatchInput(permitBatch, sig.LegacyBytes())).
		Add(NewPermit2TransferFromInput(s.currency0, AddressThis(), pulled0)).
		Add(NewPermit2TransferFromBatchInput([]permit2.AllowanceTransferDetails{
			*permit2.NewAllowanceTransferDetails(s.from, s.router, pulled1ToRouter, s.currency1),
			*permit2.NewAllowanceTransferDetails(s.from, s.feeRecipient, pulled1ToRecipient, s.currency1),
		})).
		Add(NewTransferInput(s.currency0, s.feeRecipient, transferred0)).
		Add(NewPayPortionFullPrecisionInput(s.currency0, s.feeRecipient, s.PortionFullPrecision)).
		Add(NewSweepInput(s.currency0, MsgSender(), big.NewInt(0))).
		Add(NewSweepInput(s.currency1, MsgSender(), pulled1ToRouter))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test18Permit2TransfersAndPayments: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	tokens := []*types.Address{s.currency0, s.currency1}

	var userAfter *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		userAfter, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	userBefore, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	recipientAfter, err := s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.feeRecipient, tokens, false))
	s.Require().NoError(err)
	recipientBefore, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.feeRecipient, tokens, false))
	s.Require().NoError(err)
	routerAfter, err := s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.router, tokens, false))
	s.Require().NoError(err)

	paid0 := new(big.Int).Add(transferred0, portion0)
	s.Require().Zero(paid0.Cmp(new(big.Int).Sub(recipientAfter.Tokens[0].Amount, recipientBefore.Tokens[0].Amount)), "recipient currency0 != transfer + full-precision portion")
	s.Require().Zero(pulled1ToRecipient.Cmp(new(big.Int).Sub(recipientAfter.Tokens[1].Amount, recipientBefore.Tokens[1].Amount)), "recipient currency1 != batch transfer amount")
	s.Require().Zero(paid0.Cmp(new(big.Int).Sub(userBefore.Tokens[0].Amount, userAfter.Tokens[0].Amount)), "user currency0 didn't drop by exactly what was paid out")
	s.Require().Zero(pulled1ToRecipient.Cmp(new(big.Int).Sub(userBefore.Tokens[1].Amount, userAfter.Tokens[1].Amount)), "user currency1 didn't drop by exactly the recipient's share")
	s.Require().Zero(routerAfter.Tokens[0].Amount.Sign(), "router kept currency0")
	s.Require().Zero(routerAfter.Tokens[1].Amount.Sign(), "router kept currency1")

	allowanceAfter0, err := s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(s.PermitAmount, pulled0).Cmp(allowanceAfter0.Amount))
	allowanceAfter1, err := s.permit2.Allowance(s.T().Context(), atBlock, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency1, s.router))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(new(big.Int).Sub(s.PermitAmount, pulled1ToRouter), pulled1ToRecipient).Cmp(allowanceAfter1.Amount))

	s.T().Logf("UniversalRouter Test18Permit2TransfersAndPayments Result")
	s.T().Logf("\tRecipient currency0 +%s (transfer %s + portion %s), currency1 +%s", paid0, transferred0, portion0, pulled1ToRecipient)
	s.T().Logf("\tPermit2 allowance currency0 %s, currency1 %s", allowanceAfter0.Amount, allowanceAfter1.Amount)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test19NativeETHForTokenSwap() {
	sent := new(big.Int).Mul(s.NativeSwapAmount, big.NewInt(3))

	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.nativeKey, true, s.NativeSwapAmount, big.NewInt(1), big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(types.ZeroAddress(), s.NativeSwapAmount)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency0, big.NewInt(1))
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams})).
		Add(NewSweepInput(types.ZeroAddress(), MsgSender(), big.NewInt(0)))

	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, sent)))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), sent))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test19NativeETHForTokenSwap: sent execute with %s wei (nonce %d, maxFee %s), tx %s", sent, sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v4.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Sender.Equal(s.router))
	s.Require().Zero(new(big.Int).Neg(s.NativeSwapAmount).Cmp(swaps[0].Amount0), "amount0 should equal the native input")
	s.Require().Positive(swaps[0].Amount1.Sign())

	blockHex := fmt.Sprintf("0x%x", receipt.BlockNumber)
	var userAfter *big.Int
	s.Require().Eventually(func() bool {
		userAfter, err = s.erc20.BalanceOf(s.T().Context(), bind.NewCallOptions().WithBlock(blockHex), erc20.NewBalanceOfRequest(s.currency0, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	userBefore, err := s.erc20.BalanceOf(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), erc20.NewBalanceOfRequest(s.currency0, s.from))
	s.Require().NoError(err)
	s.Require().Zero(swaps[0].Amount1.Cmp(new(big.Int).Sub(userAfter, userBefore)), "user token balance didn't move by exactly amount1")

	var routerETHHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("balance", s.router.String(), blockHex, &routerETHHex)))
	routerETH, ok := new(big.Int).SetString(strings.TrimPrefix(routerETHHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(routerETH.Sign(), "router kept native ETH after the sweep")

	s.T().Logf("UniversalRouter Test19NativeETHForTokenSwap Result")
	s.T().Logf("\tSent %s wei, swapped %s wei, received %s tokens", sent, s.NativeSwapAmount, swaps[0].Amount1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test20TokenForNativeETHSwap() {
	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)

	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	swapParams, err := v4.EncodeActionSwapExactInSingle(v4.NewExactInputSingleParams(s.nativeKey, false, s.NativeSwapAmount, big.NewInt(1), big.NewInt(0), []byte{}))
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.NativeSwapAmount)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(types.ZeroAddress(), big.NewInt(1))
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{swapParams, settleParams, takeParams}))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test20TokenForNativeETHSwap: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v4.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Sender.Equal(s.router))
	s.Require().Zero(new(big.Int).Neg(s.NativeSwapAmount).Cmp(swaps[0].Amount1), "amount1 should equal the token input")
	s.Require().Positive(swaps[0].Amount0.Sign())

	blockHex := fmt.Sprintf("0x%x", receipt.BlockNumber)
	var userAfter *big.Int
	s.Require().Eventually(func() bool {
		userAfter, err = s.erc20.BalanceOf(s.T().Context(), bind.NewCallOptions().WithBlock(blockHex), erc20.NewBalanceOfRequest(s.currency0, s.from))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	userBefore, err := s.erc20.BalanceOf(s.T().Context(), bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), erc20.NewBalanceOfRequest(s.currency0, s.from))
	s.Require().NoError(err)
	s.Require().Zero(s.NativeSwapAmount.Cmp(new(big.Int).Sub(userBefore, userAfter)), "user token balance didn't drop by exactly the input")

	var routerETHHex, routerETHBeforeHex string
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("balance", s.router.String(), blockHex, &routerETHHex)))
	s.Require().NoError(s.client.Call(s.T().Context(), rpc.ETHGetBalance("balance", s.router.String(), fmt.Sprintf("0x%x", receipt.BlockNumber-1), &routerETHBeforeHex)))
	routerETH, ok := new(big.Int).SetString(strings.TrimPrefix(routerETHHex, "0x"), 16)
	s.Require().True(ok)
	routerETHBefore, ok := new(big.Int).SetString(strings.TrimPrefix(routerETHBeforeHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(routerETHBefore.Cmp(routerETH), "TAKE_ALL to the user moved native ETH through the router")

	s.T().Logf("UniversalRouter Test20TokenForNativeETHSwap Result")
	s.T().Logf("\tSwapped %s tokens for %s wei", s.NativeSwapAmount, swaps[0].Amount0)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test21V3SwapExactIn() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency0, s.currency1}, []*big.Int{big.NewInt(3000)})
	s.Require().NoError(err)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactInInput(MsgSender(), s.SwapAmountIn, big.NewInt(1), path, true, nil))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test21V3SwapExactIn: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v3.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.v3Pool))
	s.Require().True(swaps[0].Sender.Equal(s.router))
	s.Require().True(swaps[0].Recipient.Equal(s.from))
	s.Require().Zero(s.SwapAmountIn.Cmp(swaps[0].Amount0), "pool should receive exactly amountIn of currency0")
	s.Require().Negative(swaps[0].Amount1.Sign())

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	tokens := []*types.Address{s.currency0, s.currency1}
	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(before.Tokens[0].Amount, swaps[0].Amount0).Cmp(after.Tokens[0].Amount), "currency0 didn't drop by exactly amount0")
	s.Require().Zero(new(big.Int).Sub(before.Tokens[1].Amount, swaps[0].Amount1).Cmp(after.Tokens[1].Amount), "currency1 didn't rise by exactly -amount1")

	s.T().Logf("UniversalRouter Test21V3SwapExactIn Result")
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test22V3SwapExactInMultiHop() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency0, s.currency1, s.currency0}, []*big.Int{big.NewInt(3000), big.NewInt(3000)})
	s.Require().NoError(err)
	s.Require().Len(path, 20+23*2)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactInInput(MsgSender(), s.SwapAmountIn, big.NewInt(1), path, true, []*big.Int{big.NewInt(0), big.NewInt(0)}))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test22V3SwapExactInMultiHop: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v3.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2)
	s.Require().Zero(s.SwapAmountIn.Cmp(swaps[0].Amount0), "first hop should take exactly amountIn of currency0")
	s.Require().True(swaps[0].Recipient.Equal(s.router), "first hop output should stay in the router")
	s.Require().Zero(new(big.Int).Neg(swaps[0].Amount1).Cmp(swaps[1].Amount1), "second hop should spend exactly the first hop's output")
	s.Require().True(swaps[1].Recipient.Equal(s.from))
	s.Require().Negative(swaps[1].Amount0.Sign())

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	tokens := []*types.Address{s.currency0, s.currency1}
	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	net0 := new(big.Int).Add(swaps[0].Amount0, swaps[1].Amount0)
	s.Require().Zero(new(big.Int).Sub(before.Tokens[0].Amount, net0).Cmp(after.Tokens[0].Amount), "currency0 didn't move by exactly the round-trip net")
	s.Require().Zero(before.Tokens[1].Amount.Cmp(after.Tokens[1].Amount), "intermediate currency1 leaked to the user")
	routerAfter, err := s.erc20.BalanceOf(s.T().Context(), atBlock, erc20.NewBalanceOfRequest(s.currency1, s.router))
	s.Require().NoError(err)
	s.Require().Zero(routerAfter.Sign(), "router kept intermediate currency1")

	s.T().Logf("UniversalRouter Test22V3SwapExactInMultiHop Result")
	s.T().Logf("\tHop0 %s/%s, Hop1 %s/%s, net currency0 %s", swaps[0].Amount0, swaps[0].Amount1, swaps[1].Amount0, swaps[1].Amount1, net0)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test23V3SwapExactOut() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency1, s.currency0}, []*big.Int{big.NewInt(3000)})
	s.Require().NoError(err)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactOutInput(MsgSender(), s.SwapAmountOut, s.SwapAmountIn, path, true, nil))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test23V3SwapExactOut: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	swaps, err := v3.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.v3Pool))
	s.Require().Zero(new(big.Int).Neg(s.SwapAmountOut).Cmp(swaps[0].Amount1), "pool should pay exactly amountOut of currency1")
	s.Require().Positive(swaps[0].Amount0.Sign())
	s.Require().LessOrEqual(swaps[0].Amount0.Cmp(s.SwapAmountIn), 0, "paid more than amountInMaximum")

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	tokens := []*types.Address{s.currency0, s.currency1}
	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(before.Tokens[0].Amount, swaps[0].Amount0).Cmp(after.Tokens[0].Amount), "currency0 didn't drop by exactly amount0")
	s.Require().Zero(new(big.Int).Add(before.Tokens[1].Amount, s.SwapAmountOut).Cmp(after.Tokens[1].Amount), "currency1 didn't rise by exactly amountOut")

	s.T().Logf("UniversalRouter Test23V3SwapExactOut Result")
	s.T().Logf("\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test24V3ExactInSlippageMapsDirectly() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency0, s.currency1}, []*big.Int{big.NewInt(3000)})
	s.Require().NoError(err)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactInInput(MsgSender(), s.SwapAmountIn, s.SwapAmountIn, path, true, nil))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().ErrorIs(err, ErrV3TooLittleReceived)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test24V3ExactInSlippageMapsDirectly Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test25V3ExactOutSlippageMapsDirectly() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency1, s.currency0}, []*big.Int{big.NewInt(3000)})
	s.Require().NoError(err)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactOutInput(MsgSender(), s.SwapAmountOut, big.NewInt(1), path, true, nil))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().ErrorIs(err, ErrV3TooMuchRequested)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test25V3ExactOutSlippageMapsDirectly Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test26V2SwapExactIn() {
	if s.v2Pair == nil {
		s.T().Skip("no Uniswap V2 factory bound to this UniversalRouter")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV2SwapExactInInput(MsgSender(), s.SwapAmountIn, big.NewInt(1), []*types.Address{s.currency0, s.currency1}, true, nil))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test26V2SwapExactIn: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	var reserves *v2.Reserves
	s.Require().Eventually(func() bool {
		reserves, err = v2.NewPair(s.client).GetReserves(s.T().Context(), beforeAt, v2.NewGetReservesRequest(s.v2Pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reserves not readable before the receipt block")
	amountInWithFee := new(big.Int).Mul(s.SwapAmountIn, big.NewInt(997))
	wantOut := new(big.Int).Div(new(big.Int).Mul(amountInWithFee, reserves.Reserve1), new(big.Int).Add(new(big.Int).Mul(reserves.Reserve0, big.NewInt(1000)), amountInWithFee))

	swaps, err := v2.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.v2Pair))
	s.Require().True(swaps[0].To.Equal(s.from))
	s.Require().Zero(s.SwapAmountIn.Cmp(swaps[0].Amount0In))
	s.Require().Zero(wantOut.Cmp(swaps[0].Amount1Out), "amount1Out != getAmountOut from the pre-swap reserves")

	tokens := []*types.Address{s.currency0, s.currency1}
	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(before.Tokens[0].Amount, s.SwapAmountIn).Cmp(after.Tokens[0].Amount))
	s.Require().Zero(new(big.Int).Add(before.Tokens[1].Amount, wantOut).Cmp(after.Tokens[1].Amount))

	s.T().Logf("UniversalRouter Test26V2SwapExactIn Result")
	s.T().Logf("\tReserves before %s/%s, in %s, out %s", reserves.Reserve0, reserves.Reserve1, s.SwapAmountIn, wantOut)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test27V2SwapExactOut() {
	if s.v2Pair == nil {
		s.T().Skip("no Uniswap V2 factory bound to this UniversalRouter")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV2SwapExactOutInput(MsgSender(), s.SwapAmountOut, s.SwapAmountIn, []*types.Address{s.currency0, s.currency1}, true, nil))

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(s.LimitBuffer)
	sentTx, hash, err := s.universalRouter.Execute(s.T().Context(), bind.NewTransactOptions(s.signer, builder), NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test27V2SwapExactOut: sent execute (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "execute transaction reverted")

	beforeAt := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1))
	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))
	var reserves *v2.Reserves
	s.Require().Eventually(func() bool {
		reserves, err = v2.NewPair(s.client).GetReserves(s.T().Context(), beforeAt, v2.NewGetReservesRequest(s.v2Pair))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reserves not readable before the receipt block")
	wantIn := new(big.Int).Add(new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(reserves.Reserve0, s.SwapAmountOut), big.NewInt(1000)), new(big.Int).Mul(new(big.Int).Sub(reserves.Reserve1, s.SwapAmountOut), big.NewInt(997))), big.NewInt(1))

	swaps, err := v2.ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(wantIn.Cmp(swaps[0].Amount0In), "amount0In != getAmountIn from the pre-swap reserves")
	s.Require().Zero(s.SwapAmountOut.Cmp(swaps[0].Amount1Out))

	tokens := []*types.Address{s.currency0, s.currency1}
	var after *erc20.AddressBalances
	s.Require().Eventually(func() bool {
		after, err = s.erc20.AddressBalances(s.T().Context(), atBlock, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "balances not readable at the receipt block")
	before, err := s.erc20.AddressBalances(s.T().Context(), beforeAt, erc20.NewAddressBalancesRequest(s.multicall, s.from, tokens, false))
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Sub(before.Tokens[0].Amount, wantIn).Cmp(after.Tokens[0].Amount))
	s.Require().Zero(new(big.Int).Add(before.Tokens[1].Amount, s.SwapAmountOut).Cmp(after.Tokens[1].Amount))

	s.T().Logf("UniversalRouter Test27V2SwapExactOut Result")
	s.T().Logf("\tReserves before %s/%s, in %s, out %s", reserves.Reserve0, reserves.Reserve1, wantIn, s.SwapAmountOut)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *UniversalRouterSuite) Test28V2ExactInSlippageMapsDirectly() {
	if s.v2Pair == nil {
		s.T().Skip("no Uniswap V2 factory bound to this UniversalRouter")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV2SwapExactInInput(MsgSender(), s.SwapAmountIn, s.SwapAmountIn, []*types.Address{s.currency0, s.currency1}, true, nil))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().ErrorIs(err, ErrV2TooLittleReceived)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test28V2ExactInSlippageMapsDirectly Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test29V2ExactOutSlippageMapsDirectly() {
	if s.v2Pair == nil {
		s.T().Skip("no Uniswap V2 factory bound to this UniversalRouter")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV2SwapExactOutInput(MsgSender(), s.SwapAmountOut, big.NewInt(1), []*types.Address{s.currency0, s.currency1}, true, nil))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().ErrorIs(err, ErrV2TooMuchRequested)
	s.Require().NotErrorIs(err, ErrExecutionFailed)
	s.Require().NotErrorIs(err, ErrUnknown)

	s.T().Logf("UniversalRouter Test29V2ExactOutSlippageMapsDirectly Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test30V3ExactOutMultiHopSamePoolRevertsLOK() {
	if s.v3Pool == nil {
		s.T().Skip("no Uniswap V3 factory/NFPM configured for this network")
	}

	allowance, err := s.permit2.Allowance(s.T().Context(), nil, permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router))
	s.Require().NoError(err)
	permitSingle := permit2.NewPermitSingle(permit2.NewPermitDetails(s.currency0, s.PermitAmount, big.NewInt(time.Now().Add(time.Hour).Unix()), allowance.Nonce), s.router, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	sig, err := permitSingle.Sign(permit2.NewDomain(s.ChainID, s.permit2Contract), s.signer)
	s.Require().NoError(err)

	path, err := v3.EncodePath([]*types.Address{s.currency0, s.currency1, s.currency0}, []*big.Int{big.NewInt(3000), big.NewInt(3000)})
	s.Require().NoError(err)
	plan := NewPlan().
		Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).
		Add(NewV3SwapExactOutInput(MsgSender(), s.SwapAmountOut, s.SwapAmountIn, path, true, nil))

	err = s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, plan, nil, nil))
	s.Require().Error(err)
	s.Require().ErrorIs(err, v3.ErrLocked)
	s.Require().NotErrorIs(err, ErrUnknown)
	s.Require().NotErrorIs(err, ErrExecutionFailed)

	rpcErr, ok := errors.AsType[*rpc.ResponseError](err)
	s.Require().True(ok)
	var revertHex string
	s.Require().NoError(json.Unmarshal(rpcErr.Data, &revertHex))
	reason, err := abi.Pack(abi.NewTypes(abi.String), "LOK")
	s.Require().NoError(err)
	s.Require().Equal("0x08c379a0"+encoding.Hex.Encode(reason), strings.ToLower(revertHex), "expected the V3 pool's reentrancy guard Error(\"LOK\")")

	singleHopPath, err := v3.EncodePath([]*types.Address{s.currency1, s.currency0}, []*big.Int{big.NewInt(3000)})
	s.Require().NoError(err)
	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewExecuteRequest(s.router, NewPlan().Add(NewPermit2PermitInput(permitSingle, sig.LegacyBytes())).Add(NewV3SwapExactOutInput(MsgSender(), s.SwapAmountOut, s.SwapAmountIn, singleHopPath, true, nil)), nil, nil)))

	s.T().Logf("UniversalRouter Test30V3ExactOutMultiHopSamePoolRevertsLOK Result")
	s.T().Logf("\tError: %s", err)
}

func (s *UniversalRouterSuite) Test31CallExecuteWithoutApprovalsNeedsOverrides() {
	ghost := types.NewAddressFromBytes(core.Keccak256([]byte("go-evmkit-universal-router-ghost")).Bytes())
	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())

	balanceKey, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), ghost, big.NewInt(0))
	s.Require().NoError(err)
	var balanceSlot [32]byte
	copy(balanceSlot[:], core.Keccak256(balanceKey).Bytes())

	tokenOwnerKey, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), ghost, big.NewInt(1))
	s.Require().NoError(err)
	tokenAllowanceKey, err := abi.Pack(
		abi.NewTypes(abi.Address, abi.Uint256),
		s.permit2Contract,
		new(big.Int).SetBytes(core.Keccak256(tokenOwnerKey).Bytes()),
	)
	s.Require().NoError(err)
	var tokenAllowanceSlot [32]byte
	copy(tokenAllowanceSlot[:], core.Keccak256(tokenAllowanceKey).Bytes())

	permitOwnerKey, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), ghost, big.NewInt(1))
	s.Require().NoError(err)
	permitTokenKey, err := abi.Pack(
		abi.NewTypes(abi.Address, abi.Uint256),
		s.currency0,
		new(big.Int).SetBytes(core.Keccak256(permitOwnerKey).Bytes()),
	)
	s.Require().NoError(err)
	permitSpenderKey, err := abi.Pack(
		abi.NewTypes(abi.Address, abi.Uint256),
		s.router,
		new(big.Int).SetBytes(core.Keccak256(permitTokenKey).Bytes()),
	)
	s.Require().NoError(err)
	var permitAllowanceSlot [32]byte
	copy(permitAllowanceSlot[:], core.Keccak256(permitSpenderKey).Bytes())

	var amountWord, permitWord [32]byte
	s.SwapAmountIn.FillBytes(amountWord[:])
	new(big.Int).Or(new(big.Int).Lsh(expiration, 160), s.SwapAmountIn).FillBytes(permitWord[:])

	swapInput := v4.NewExactInputSingleParams(s.key, true, s.SwapAmountIn, big.NewInt(1), big.NewInt(0), []byte{})
	swapParams, err := v4.EncodeActionSwapExactInSingle(swapInput)
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency1, big.NewInt(1))
	s.Require().NoError(err)
	plan := NewPlan().Add(NewV4SwapInput(
		[]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll},
		[][]byte{swapParams, settleParams, takeParams},
	))
	req := NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil)

	plainErr := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(ghost), req)
	s.Require().ErrorIs(plainErr, permit2.ErrAllowanceExpired)

	permitOnly := rpc.NewStateOverride().WithStateDiff(s.permit2Contract, permitAllowanceSlot, permitWord)
	opts := bind.NewCallOptions().WithFrom(ghost).WithStateOverride(permitOnly)
	permitOnlyErr := s.universalRouter.CallExecute(s.T().Context(), opts, req)
	s.Require().Error(permitOnlyErr)
	s.Require().NotErrorIs(permitOnlyErr, permit2.ErrAllowanceExpired)

	noBalance := rpc.NewStateOverride().
		WithStateDiff(s.permit2Contract, permitAllowanceSlot, permitWord).
		WithStateDiff(s.currency0, tokenAllowanceSlot, amountWord)
	opts = bind.NewCallOptions().WithFrom(ghost).WithStateOverride(noBalance)
	noBalanceErr := s.universalRouter.CallExecute(s.T().Context(), opts, req)
	s.Require().Error(noBalanceErr)

	full := rpc.NewStateOverride().
		WithStateDiff(s.permit2Contract, permitAllowanceSlot, permitWord).
		WithStateDiff(s.currency0, tokenAllowanceSlot, amountWord).
		WithStateDiff(s.currency0, balanceSlot, amountWord)
	opts = bind.NewCallOptions().WithFrom(ghost).WithStateOverride(full)
	s.Require().NoError(s.universalRouter.CallExecute(s.T().Context(), opts, req))

	permitAllowance, err := s.permit2.Allowance(
		s.T().Context(),
		bind.NewCallOptions().WithStateOverride(full),
		permit2.NewAllowanceRequest(s.permit2Contract, ghost, s.currency0, s.router),
	)
	s.Require().NoError(err)
	s.Require().Zero(s.SwapAmountIn.Cmp(permitAllowance.Amount))
	s.Require().Zero(expiration.Cmp(permitAllowance.Expiration))
	s.Require().Zero(permitAllowance.Nonce.Sign())

	commands, inputs, err := plan.Encode()
	s.Require().NoError(err)
	calldata, err := EncodeExecuteWithDeadline(commands, inputs, req.Deadline)
	s.Require().NoError(err)
	_, packErr := tx.NewDynamicFeeBuilder().
		Build(ghost, s.router, big.NewInt(0), calldata).
		Pack(s.T().Context(), s.client)
	s.Require().Error(packErr)
	builder := tx.NewDynamicFeeBuilder()
	builder.StateOverride(full)
	packed, err := builder.Build(ghost, s.router, big.NewInt(0), calldata).Pack(s.T().Context(), s.client)
	s.Require().NoError(err)
	s.Require().Greater(packed.Gas(), uint64(21_000))

	s.T().Logf("UniversalRouter Test31CallExecuteWithoutApprovalsNeedsOverrides Result")
	s.T().Logf("\tghost %s swaps %s of %s with no balance, approval or permit", ghost, s.SwapAmountIn, s.currency0)
	s.T().Logf("\ttoken balance slot %x", balanceSlot)
	s.T().Logf("\ttoken allowance(ghost, Permit2) slot %x", tokenAllowanceSlot)
	s.T().Logf("\tPermit2 allowance(ghost, token, router) slot %x", permitAllowanceSlot)
	s.T().Logf("\tPermit2 packed word %x", permitWord)
	s.T().Logf("\tno override: %v", plainErr)
	s.T().Logf("\tPermit2 allowance only: %v", permitOnlyErr)
	s.T().Logf("\tPermit2 + token allowance, no balance: %v", noBalanceErr)
	s.T().Logf("\tall three overrides: success")
	s.T().Logf("\tPermit2.allowance under override: amount %s", permitAllowance.Amount)
	s.T().Logf("\t\texpiration %s, nonce %s", permitAllowance.Expiration, permitAllowance.Nonce)
	s.T().Logf("\tPack without override: %v", packErr)
	s.T().Logf("\tPack with override: %d gas", packed.Gas())
}

func (s *UniversalRouterSuite) Test32Permit2ApproveThenSwapBackToBack() {
	expiration := big.NewInt(time.Now().Add(time.Hour).Unix())
	resetBuilder := tx.NewDynamicFeeBuilder()
	resetBuilder.LimitBuffer(s.LimitBuffer)
	resetTx, resetHash, err := s.permit2.Approve(
		s.T().Context(),
		bind.NewTransactOptions(s.signer, resetBuilder),
		permit2.NewApproveRequest(s.permit2Contract, s.currency0, s.router, big.NewInt(0), expiration),
	)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, resetTx.MaxFee())
	s.mu.Unlock()
	resetReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		resetHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(resetReceipt.Status, "Permit2 reset approve reverted")

	allowance, err := s.permit2.Allowance(
		s.T().Context(),
		bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", resetReceipt.BlockNumber)),
		permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router),
	)
	s.Require().NoError(err)
	s.Require().Zero(allowance.Amount.Sign())
	s.Require().Zero(expiration.Cmp(allowance.Expiration))

	ownerKey, err := abi.Pack(abi.NewTypes(abi.Address, abi.Uint256), s.from, big.NewInt(1))
	s.Require().NoError(err)
	tokenKey, err := abi.Pack(
		abi.NewTypes(abi.Address, abi.Uint256),
		s.currency0,
		new(big.Int).SetBytes(core.Keccak256(ownerKey).Bytes()),
	)
	s.Require().NoError(err)
	spenderKey, err := abi.Pack(
		abi.NewTypes(abi.Address, abi.Uint256),
		s.router,
		new(big.Int).SetBytes(core.Keccak256(tokenKey).Bytes()),
	)
	s.Require().NoError(err)
	var permitAllowanceSlot, permitWord [32]byte
	copy(permitAllowanceSlot[:], core.Keccak256(spenderKey).Bytes())
	packedWord := new(big.Int).Lsh(allowance.Nonce, 208)
	packedWord.Or(packedWord, new(big.Int).Lsh(expiration, 160))
	packedWord.Or(packedWord, s.SwapAmountIn)
	packedWord.FillBytes(permitWord[:])
	override := rpc.NewStateOverride().WithStateDiff(s.permit2Contract, permitAllowanceSlot, permitWord)

	swapInput := v4.NewExactInputSingleParams(s.key, true, s.SwapAmountIn, big.NewInt(1), big.NewInt(0), []byte{})
	swapParams, err := v4.EncodeActionSwapExactInSingle(swapInput)
	s.Require().NoError(err)
	settleParams, err := v4.EncodeActionSettleAll(s.currency0, s.SwapAmountIn)
	s.Require().NoError(err)
	takeParams, err := v4.EncodeActionTakeAll(s.currency1, big.NewInt(1))
	s.Require().NoError(err)
	plan := NewPlan().Add(NewV4SwapInput(
		[]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll},
		[][]byte{swapParams, settleParams, takeParams},
	))
	req := NewExecuteRequest(s.router, plan, big.NewInt(time.Now().Add(10*time.Minute).Unix()), nil)

	unapprovedErr := s.universalRouter.CallExecute(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), req)
	s.Require().ErrorIs(unapprovedErr, permit2.ErrInsufficientAllowance)

	var nonceHex string
	e := rpc.ETHGetTransactionCount("nonce", s.from.String(), rpc.BlockTagPending, &nonceHex)
	s.Require().NoError(s.client.Call(s.T().Context(), e))
	nonce, err := strconv.ParseUint(strings.TrimPrefix(nonceHex, "0x"), 16, 64)
	s.Require().NoError(err)

	approveBuilder := tx.NewDynamicFeeBuilder()
	approveBuilder.Nonce(nonce)
	approveTx, approveHash, err := s.permit2.Approve(
		s.T().Context(),
		bind.NewTransactOptions(s.signer, approveBuilder),
		permit2.NewApproveRequest(s.permit2Contract, s.currency0, s.router, s.SwapAmountIn, expiration),
	)
	s.Require().NoError(err)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.Nonce(nonce + 1)
	swapBuilder.StateOverride(override)
	swapOpts := bind.NewTransactOptions(s.signer, swapBuilder)
	swapTx, swapHash, err := s.universalRouter.Execute(s.T().Context(), swapOpts, req)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, approveTx.MaxFee())
	s.totalFee.Add(s.totalFee, swapTx.MaxFee())
	s.mu.Unlock()

	approveReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		approveHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "Permit2 approve reverted")
	swapReceipt, err := tx.WaitForReceipt(
		s.T().Context(),
		s.client,
		swapHash,
		s.ReceiptPollInterval,
		s.ReceiptTimeout,
	)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "execute reverted with its override-based gas limit")
	s.Require().LessOrEqual(approveReceipt.GasUsed, approveTx.Gas())
	s.Require().LessOrEqual(swapReceipt.GasUsed, swapTx.Gas())

	swaps, err := v4.ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().Zero(new(big.Int).Neg(s.SwapAmountIn).Cmp(swaps[0].Amount0))

	allowanceAfter, err := s.permit2.Allowance(
		s.T().Context(),
		bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", swapReceipt.BlockNumber)),
		permit2.NewAllowanceRequest(s.permit2Contract, s.from, s.currency0, s.router),
	)
	s.Require().NoError(err)
	s.Require().Zero(allowanceAfter.Amount.Sign())
	s.Require().Zero(expiration.Cmp(allowanceAfter.Expiration))
	s.Require().Zero(allowance.Nonce.Cmp(allowanceAfter.Nonce))

	approvePriceHex, ok := approveReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	approvePrice, ok := new(big.Int).SetString(strings.TrimPrefix(approvePriceHex, "0x"), 16)
	s.Require().True(ok)
	swapPriceHex, ok := swapReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	swapPrice, ok := new(big.Int).SetString(strings.TrimPrefix(swapPriceHex, "0x"), 16)
	s.Require().True(ok)
	approveFee := new(big.Int).Mul(new(big.Int).SetUint64(approveReceipt.GasUsed), approvePrice)
	swapFee := new(big.Int).Mul(new(big.Int).SetUint64(swapReceipt.GasUsed), swapPrice)

	overrideJSON, err := override.MarshalJSON()
	s.Require().NoError(err)
	s.T().Logf("UniversalRouter Test32Permit2ApproveThenSwapBackToBack Result")
	s.T().Logf("\treset Permit2 allowance to 0 until %s", expiration)
	s.T().Logf("\t\ttx %s, block %d", resetHash, resetReceipt.BlockNumber)
	s.T().Logf("\twithout approval: %v", unapprovedErr)
	s.T().Logf("\toverride for swap estimate: %s", overrideJSON)
	s.T().Logf("\tstep 1 Permit2.approve(%s, %s, %s, %s)", s.currency0, s.router, s.SwapAmountIn, expiration)
	s.T().Logf("\t\ttx %s, nonce %d, block %d", approveHash, approveTx.AddressNonce(), approveReceipt.BlockNumber)
	s.T().Logf("\t\testimated limit %d, gasUsed %d", approveTx.Gas(), approveReceipt.GasUsed)
	s.T().Logf("\t\tused %d%% of the estimate", approveReceipt.GasUsed*100/approveTx.Gas())
	s.T().Logf("\t\teffectiveGasPrice %s wei, fee %s wei", approvePrice, approveFee)
	s.T().Logf("\tstep 2 execute(V4_SWAP %s %s -> %s)", s.SwapAmountIn, s.currency0, s.currency1)
	s.T().Logf("\t\ttx %s, nonce %d, block %d", swapHash, swapTx.AddressNonce(), swapReceipt.BlockNumber)
	s.T().Logf("\t\testimated limit %d, gasUsed %d", swapTx.Gas(), swapReceipt.GasUsed)
	s.T().Logf("\t\tused %d%% of the estimate", swapReceipt.GasUsed*100/swapTx.Gas())
	s.T().Logf("\t\teffectiveGasPrice %s wei, fee %s wei", swapPrice, swapFee)
	s.T().Logf("\t\tAmount0/Amount1: %s / %s", swaps[0].Amount0, swaps[0].Amount1)
	s.T().Logf("\ttotal")
	s.T().Logf("\t\testimated %d gas", approveTx.Gas()+swapTx.Gas())
	s.T().Logf("\t\tused %d gas", approveReceipt.GasUsed+swapReceipt.GasUsed)
	s.T().Logf("\t\tfee %s wei %s", new(big.Int).Add(approveFee, swapFee), s.NativeSymbol)
	s.T().Logf("\tPermit2 allowance after: amount %s", allowanceAfter.Amount)
	s.T().Logf("\t\texpiration %s, nonce %s", allowanceAfter.Expiration, allowanceAfter.Nonce)
}
