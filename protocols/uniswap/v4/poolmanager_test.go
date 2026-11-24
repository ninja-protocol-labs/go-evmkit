package v4

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc6909"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type PoolManagerSuite struct {
	suite.Suite

	RPCURL                     string
	PrivateKeyHex              string
	PoolManagerHex             string
	PoolSwapTestHex            string
	PoolModifyLiquidityTestHex string
	MulticallHex               string
	NicksFactoryHex            string
	NativeSymbol               string
	NativeDecimals             int
	Timeout                    time.Duration
	ReceiptPollInterval        time.Duration
	ReceiptTimeout             time.Duration
	ByteCode                   string
	Constructor                *abi.Function
	Salt                       *types.Hash
	MintValue                  *big.Int
	LiquidityDelta             *big.Int
	SwapAmountIn               *big.Int

	client                  rpc.Client
	signer                  core.Signer
	deployer                *deploy.NicksDeployer
	from                    *types.Address
	spender                 *types.Address
	contract                *types.Address
	multicall               *types.Address
	nicks                   *types.Address
	poolSwapTest            *types.Address
	poolModifyLiquidityTest *types.Address
	currency0               *types.Address
	currency1               *types.Address

	poolManager *PoolManager
	erc20       *erc20.ERC20

	key           *PoolKey
	poolID        PoolID
	stateSlot     [32]byte
	liquiditySlot [32]byte
	sqrtPriceX96  *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestPoolManagerSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V4.PoolManager == "" {
		t.Skip("testenv: no Uniswap V4 PoolManager configured for this network")
	}

	suite.Run(t, &PoolManagerSuite{
		RPCURL:                     c.RPCURL,
		PrivateKeyHex:              c.ActorPrivateKeyHex,
		PoolManagerHex:             c.Addresses.Protocol.Uniswap.V4.PoolManager,
		PoolSwapTestHex:            c.Addresses.Protocol.Uniswap.V4.PoolSwapTest,
		PoolModifyLiquidityTestHex: c.Addresses.Protocol.Uniswap.V4.PoolModifyLiquidityTest,
		MulticallHex:               c.Addresses.Standard.Multicall,
		NicksFactoryHex:            c.Addresses.Standard.NicksFactory,
		NativeSymbol:               c.Constant.NativeSymbol,
		NativeDecimals:             c.Constant.NativeDecimals,
		Timeout:                    10 * time.Minute,
		ReceiptPollInterval:        c.Constant.ReceiptPollInterval,
		ReceiptTimeout:             c.Constant.ReceiptTimeout,
		ByteCode:                   testenv.ERC20ByteCodeHex,
		Constructor:                abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                       core.Keccak256([]byte("go-evmkit-v4-poolmanager")),
		MintValue:                  new(big.Int).Mul(big.NewInt(1_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		LiquidityDelta:             new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		SwapAmountIn:               new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil),
	})
}

func (s *PoolManagerSuite) SetupSuite() {
	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.poolManager = New(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.spender = types.NewAddressFromBytes(core.Keccak256(fmt.Appendf(nil, "%s-spender", s.Salt)).Bytes()[12:])
	s.T().Logf("SetupSuite: signing as %s, spender %s", s.from, s.spender)

	s.contract, err = types.NewAddressFromHex(s.PoolManagerHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target PoolManager %s", s.contract)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	if s.PoolSwapTestHex != "" && s.PoolModifyLiquidityTestHex != "" {
		s.poolSwapTest, err = types.NewAddressFromHex(s.PoolSwapTestHex)
		s.Require().NoError(err)
		s.poolModifyLiquidityTest, err = types.NewAddressFromHex(s.PoolModifyLiquidityTestHex)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: PoolSwapTest %s, PoolModifyLiquidityTest %s", s.poolSwapTest, s.poolModifyLiquidityTest)
	}

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit v4 test token A", "evmkit v4 test token B"}
	symbols := [2]string{"EVKA", "EVKB"}
	tokens := [2]*types.Address{}
	tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)

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

		if s.poolSwapTest == nil {
			continue
		}

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

		for _, spender := range []*types.Address{s.poolModifyLiquidityTest, s.poolSwapTest} {
			approveTx, approveHash, err := s.erc20.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc20.NewApproveRequest(token, spender, s.MintValue))
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

	s.key = NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	s.sqrtPriceX96 = new(big.Int).Lsh(big.NewInt(1), 96)

	s.poolID, err = s.key.PoolID()
	s.Require().NoError(err)

	var slotKey [64]byte
	copy(slotKey[:32], s.poolID[:])
	slotKey[63] = 6
	copy(s.stateSlot[:], core.Keccak256(slotKey[:]).Bytes())
	new(big.Int).Add(new(big.Int).SetBytes(s.stateSlot[:]), big.NewInt(3)).FillBytes(s.liquiditySlot[:])
	s.T().Logf("SetupSuite: poolId %x, state slot %x, liquidity slot %x", s.poolID, s.stateSlot, s.liquiditySlot)
}

func (s *PoolManagerSuite) TearDownSuite() {
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

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(ctx, s.client, raw)
			s.Require().NoError(err)

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

		s.T().Logf("PoolManagerSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}
}

func (s *PoolManagerSuite) Test01ReadMethods() {
	controller, err := s.poolManager.ProtocolFeeController(s.T().Context(), nil, NewProtocolFeeControllerRequest(s.contract))
	s.Require().NoError(err)
	s.Require().NotNil(controller)

	accruedNative, err := s.poolManager.ProtocolFeesAccrued(s.T().Context(), nil, NewProtocolFeesAccruedRequest(s.contract, types.ZeroAddress()))
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(accruedNative.Sign(), 0)

	accrued0, err := s.poolManager.ProtocolFeesAccrued(s.T().Context(), nil, NewProtocolFeesAccruedRequest(s.contract, s.currency0))
	s.Require().NoError(err)
	s.Require().Zero(accrued0.Sign(), "a fresh token can't have accrued protocol fees")

	slot0, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.contract, s.stateSlot))
	s.Require().NoError(err)
	s.Require().Equal([32]byte{}, slot0, "the pool shouldn't be initialized yet")

	transient, err := s.poolManager.Exttload(s.T().Context(), nil, NewExttloadRequest(s.contract, s.stateSlot))
	s.Require().NoError(err)
	s.Require().Equal([32]byte{}, transient)

	empty, err := s.poolManager.ExtsloadSlots(s.T().Context(), nil, NewExtsloadSlotsRequest(s.contract, nil))
	s.Require().NoError(err)
	s.Require().Empty(empty)

	claim, err := s.poolManager.BalanceOf(s.T().Context(), nil, erc6909.NewBalanceOfRequest(s.contract, s.from, s.currency0.Big()))
	s.Require().NoError(err)
	s.Require().Zero(claim.Sign())

	operator, err := s.poolManager.IsOperator(s.T().Context(), nil, erc6909.NewIsOperatorRequest(s.contract, s.from, s.spender))
	s.Require().NoError(err)
	s.Require().False(operator)

	s.T().Logf("PoolManager %s Test01ReadMethods Result", s.contract)
	s.T().Logf("\tProtocolFeeController: %s", controller)
	s.T().Logf("\tProtocolFeesAccrued(native): %s", accruedNative)
	s.T().Logf("\tProtocolFeesAccrued(currency0): %s", accrued0)
	s.T().Logf("\tExtsload(stateSlot): %x", slot0)
	s.T().Logf("\tExttload(stateSlot): %x", transient)
	s.T().Logf("\tBalanceOf(from, currency0): %s", claim)
	s.T().Logf("\tIsOperator(from, spender): %v", operator)
}

func (s *PoolManagerSuite) Test02CallInitializePredictsTick() {
	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().NoError(err)
	s.Require().Zero(tick.Sign())

	otherKey := NewPoolKey(s.currency0, s.currency1, big.NewInt(500), big.NewInt(10), types.ZeroAddress())
	doublePrice := new(big.Int).Lsh(big.NewInt(1), 97)
	otherTick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, otherKey, doublePrice))
	s.Require().NoError(err)
	want := int64(math.Floor(math.Log(4) / math.Log(1.0001)))
	s.Require().Equal(want, otherTick.Int64())

	slot0, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.contract, s.stateSlot))
	s.Require().NoError(err)
	s.Require().Equal([32]byte{}, slot0, "eth_call must not persist the initialization")

	s.T().Logf("PoolManager %s Test02CallInitializePredictsTick Result", s.contract)
	s.T().Logf("\tCallInitialize(fee 3000, sqrtPrice 2^96): tick %s", tick)
	s.T().Logf("\tCallInitialize(fee 500, sqrtPrice 2^97): tick %s (want %d)", otherTick, want)
}

func (s *PoolManagerSuite) Test03Initialize() {
	sentTx, hash, err := s.poolManager.Initialize(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03Initialize: sent initialize (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "initialize transaction reverted")

	inits, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(inits, 1)
	s.Require().True(inits[0].Contract.Equal(s.contract))
	s.Require().Equal(s.poolID, inits[0].Id)
	s.Require().True(inits[0].Currency0.Equal(s.key.Currency0))
	s.Require().True(inits[0].Currency1.Equal(s.key.Currency1))
	s.Require().Zero(s.key.Fee.Cmp(inits[0].Fee))
	s.Require().Zero(s.key.TickSpacing.Cmp(inits[0].TickSpacing))
	s.Require().True(inits[0].Hooks.Equal(s.key.Hooks))
	s.Require().Zero(s.sqrtPriceX96.Cmp(inits[0].SqrtPriceX96))
	s.Require().Zero(inits[0].Tick.Sign())

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var slot0 [32]byte
	s.Require().Eventually(func() bool {
		slot0, err = s.poolManager.Extsload(s.T().Context(), atBlock, NewExtsloadRequest(s.contract, s.stateSlot))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "slot0 not readable at the receipt block")

	word := new(big.Int).SetBytes(slot0[:])
	sqrtPriceX96 := new(big.Int).And(word, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)))
	tick := new(big.Int).And(new(big.Int).Rsh(word, 160), big.NewInt(0xffffff))
	if tick.Bit(23) == 1 {
		tick.Sub(tick, big.NewInt(1<<24))
	}
	protocolFee := new(big.Int).And(new(big.Int).Rsh(word, 184), big.NewInt(0xffffff))
	lpFee := new(big.Int).And(new(big.Int).Rsh(word, 208), big.NewInt(0xffffff))
	s.Require().Zero(s.sqrtPriceX96.Cmp(sqrtPriceX96))
	s.Require().Zero(tick.Sign())
	s.Require().Zero(protocolFee.Sign())
	s.Require().Zero(s.key.Fee.Cmp(lpFee))

	liquidity, err := s.poolManager.Extsload(s.T().Context(), atBlock, NewExtsloadRequest(s.contract, s.liquiditySlot))
	s.Require().NoError(err)
	s.Require().Equal([32]byte{}, liquidity)

	s.T().Logf("PoolManager %s Test03Initialize Result", s.contract)
	s.T().Logf("\tInitialize event: id=%x currency0=%s currency1=%s fee=%s tickSpacing=%s hooks=%s sqrtPriceX96=%s tick=%s", inits[0].Id, inits[0].Currency0, inits[0].Currency1, inits[0].Fee, inits[0].TickSpacing, inits[0].Hooks, inits[0].SqrtPriceX96, inits[0].Tick)
	s.T().Logf("\tSlot0 at receipt block: sqrtPriceX96=%s tick=%s protocolFee=%s lpFee=%s", sqrtPriceX96, tick, protocolFee, lpFee)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolManagerSuite) Test04InitializeAlreadyInitialized() {
	revertTx, revertHash, err := s.poolManager.Initialize(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrPoolAlreadyInitialized)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	tick, callErr := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, s.key, s.sqrtPriceX96))
	s.Require().ErrorIs(callErr, ErrPoolAlreadyInitialized)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test04InitializeAlreadyInitialized Result", s.contract)
	s.T().Logf("\tInitialize error: %v", err)
	s.T().Logf("\tCallInitialize error: %v", callErr)
}

func (s *PoolManagerSuite) Test05InitializeCurrenciesReversed() {
	key := NewPoolKey(s.currency1, s.currency0, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())

	tick, callErr := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(callErr, ErrCurrenciesOutOfOrderOrEqual)
	s.Require().Nil(tick)

	revertTx, revertHash, err := s.poolManager.Initialize(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrCurrenciesOutOfOrderOrEqual)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test05InitializeCurrenciesReversed Result", s.contract)
	s.T().Logf("\tKey: currency0=%s currency1=%s", key.Currency0, key.Currency1)
	s.T().Logf("\tCallInitialize error: %v", callErr)
	s.T().Logf("\tInitialize error: %v", err)
}

func (s *PoolManagerSuite) Test06InitializeCurrenciesEqual() {
	key := NewPoolKey(s.currency0, s.currency0, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrCurrenciesOutOfOrderOrEqual)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test06InitializeCurrenciesEqual Result", s.contract)
	s.T().Logf("\tKey: currency0=%s currency1=%s", key.Currency0, key.Currency1)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test07InitializeTickSpacingZero() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(0), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrTickSpacingTooSmall)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test07InitializeTickSpacingZero Result", s.contract)
	s.T().Logf("\tKey: tickSpacing=%s", key.TickSpacing)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test08InitializeTickSpacingNegative() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(-60), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrTickSpacingTooSmall)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test08InitializeTickSpacingNegative Result", s.contract)
	s.T().Logf("\tKey: tickSpacing=%s", key.TickSpacing)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test09InitializeTickSpacingTooLarge() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(32768), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrTickSpacingTooLarge)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test09InitializeTickSpacingTooLarge Result", s.contract)
	s.T().Logf("\tKey: tickSpacing=%s", key.TickSpacing)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test10InitializeLPFeeTooLarge() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(1_000_001), big.NewInt(60), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrLPFeeTooLarge)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test10InitializeLPFeeTooLarge Result", s.contract)
	s.T().Logf("\tKey: fee=%s", key.Fee)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test11InitializeDynamicFeeWithoutHooks() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(0x800000), big.NewInt(60), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrHookAddressNotValid)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test11InitializeDynamicFeeWithoutHooks Result", s.contract)
	s.T().Logf("\tKey: fee=0x%x hooks=%s", key.Fee, key.Hooks)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test12InitializeHookWithoutFlags() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(3000), big.NewInt(60), types.MustNewAddressFromHex("0x0000000000000000000000000000000000004000"))

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, s.sqrtPriceX96))
	s.Require().ErrorIs(err, ErrHookAddressNotValid)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test12InitializeHookWithoutFlags Result", s.contract)
	s.T().Logf("\tKey: hooks=%s", key.Hooks)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test13InitializeSqrtPriceBelowMin() {
	key := NewPoolKey(s.currency0, s.currency1, big.NewInt(500), big.NewInt(10), types.ZeroAddress())

	tick, err := s.poolManager.CallInitialize(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewInitializeRequest(s.contract, key, big.NewInt(1)))
	s.Require().ErrorIs(err, ErrInvalidSqrtPrice)
	s.Require().Nil(tick)

	s.T().Logf("PoolManager %s Test13InitializeSqrtPriceBelowMin Result", s.contract)
	s.T().Logf("\tSqrtPriceX96: 1")
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test14CallModifyLiquidityLocked() {
	callerDelta, feesAccrued, err := s.poolManager.CallModifyLiquidity(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewModifyLiquidityRequest(s.contract, s.key, NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1), [32]byte{}), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(callerDelta)
	s.Require().Nil(feesAccrued)

	s.T().Logf("PoolManager %s Test14CallModifyLiquidityLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test15CallSwapLocked() {
	delta, err := s.poolManager.CallSwap(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewSwapRequest(s.contract, s.key, NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(4295128740)), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(delta)

	s.T().Logf("PoolManager %s Test15CallSwapLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test16CallDonateLocked() {
	delta, err := s.poolManager.CallDonate(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewDonateRequest(s.contract, s.key, big.NewInt(1), big.NewInt(1), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(delta)

	s.T().Logf("PoolManager %s Test16CallDonateLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test17CallSettleLocked() {
	paid, err := s.poolManager.CallSettle(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewSettleRequest(s.contract, nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(paid)

	s.T().Logf("PoolManager %s Test17CallSettleLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test18CallSettleForLocked() {
	paid, err := s.poolManager.CallSettleFor(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewSettleForRequest(s.contract, s.spender, nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(paid)

	s.T().Logf("PoolManager %s Test18CallSettleForLocked Result", s.contract)
	s.T().Logf("\tRecipient: %s", s.spender)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test19ModifyLiquidityLocked() {
	revertTx, revertHash, err := s.poolManager.ModifyLiquidity(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewModifyLiquidityRequest(s.contract, s.key, NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1), [32]byte{}), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test19ModifyLiquidityLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test20SwapLocked() {
	revertTx, revertHash, err := s.poolManager.Swap(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSwapRequest(s.contract, s.key, NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(4295128740)), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test20SwapLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test21DonateLocked() {
	revertTx, revertHash, err := s.poolManager.Donate(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewDonateRequest(s.contract, s.key, big.NewInt(1), big.NewInt(1), nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test21DonateLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test22TakeLocked() {
	revertTx, revertHash, err := s.poolManager.Take(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewTakeRequest(s.contract, s.currency0, s.from, big.NewInt(1)))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test22TakeLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test23SettleLocked() {
	revertTx, revertHash, err := s.poolManager.Settle(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSettleRequest(s.contract, nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test23SettleLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test24SettleForLocked() {
	revertTx, revertHash, err := s.poolManager.SettleFor(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSettleForRequest(s.contract, s.from, nil))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test24SettleForLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test25ClearLocked() {
	revertTx, revertHash, err := s.poolManager.Clear(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewClearRequest(s.contract, s.currency0, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test25ClearLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test26MintLocked() {
	revertTx, revertHash, err := s.poolManager.Mint(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewMintRequest(s.contract, s.from, s.currency0.Big(), big.NewInt(1)))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test26MintLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test27BurnLocked() {
	revertTx, revertHash, err := s.poolManager.Burn(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewBurnRequest(s.contract, s.from, s.currency0.Big(), big.NewInt(1)))
	s.Require().ErrorIs(err, ErrManagerLocked)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test27BurnLocked Result", s.contract)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test28UnlockFromEOAReverts() {
	out, callErr := s.poolManager.CallUnlock(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewUnlockRequest(s.contract, []byte{0x01}))
	s.Require().Error(callErr)
	s.Require().Nil(out)
	s.Require().NotErrorIs(callErr, ErrManagerLocked)
	_, isRPC := errors.AsType[*rpc.ResponseError](callErr)
	s.Require().True(isRPC)

	sent, hash, err := s.poolManager.Unlock(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewUnlockRequest(s.contract, nil))
	s.Require().Error(err)
	s.Require().Nil(sent)
	s.Require().Nil(hash)

	s.T().Logf("PoolManager %s Test28UnlockFromEOAReverts Result", s.contract)
	s.T().Logf("\tCallUnlock error: %v", callErr)
	s.T().Logf("\tUnlock error: %v", err)
}

func (s *PoolManagerSuite) Test29Sync() {
	sentTx, hash, err := s.poolManager.Sync(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSyncRequest(s.contract, s.currency0))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "sync transaction reverted")
	s.Require().Empty(receipt.Logs)

	s.T().Logf("PoolManager %s Test29Sync Result", s.contract)
	s.T().Logf("\tCurrency: %s", s.currency0)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolManagerSuite) Test30SetProtocolFeeInvalidCaller() {
	controller, err := s.poolManager.ProtocolFeeController(s.T().Context(), nil, NewProtocolFeeControllerRequest(s.contract))
	s.Require().NoError(err)
	s.Require().False(controller.Equal(s.from), "s.from shouldn't be the protocol fee controller")

	revertTx, revertHash, err := s.poolManager.SetProtocolFee(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetProtocolFeeRequest(s.contract, s.key, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrInvalidCaller)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test30SetProtocolFeeInvalidCaller Result", s.contract)
	s.T().Logf("\tCaller: %s (controller is %s)", s.from, controller)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test31CollectProtocolFeesInvalidCaller() {
	collected, callErr := s.poolManager.CallCollectProtocolFees(s.T().Context(), bind.NewCallOptions().WithFrom(s.from), NewCollectProtocolFeesRequest(s.contract, s.from, s.currency0, big.NewInt(0)))
	s.Require().ErrorIs(callErr, ErrInvalidCaller)
	s.Require().Nil(collected)

	revertTx, revertHash, err := s.poolManager.CollectProtocolFees(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCollectProtocolFeesRequest(s.contract, s.from, s.currency0, big.NewInt(0)))
	s.Require().ErrorIs(err, ErrInvalidCaller)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test31CollectProtocolFeesInvalidCaller Result", s.contract)
	s.T().Logf("\tCallCollectProtocolFees error: %v", callErr)
	s.T().Logf("\tCollectProtocolFees error: %v", err)
}

func (s *PoolManagerSuite) Test32SetProtocolFeeControllerNotOwner() {
	revertTx, revertHash, err := s.poolManager.SetProtocolFeeController(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetProtocolFeeControllerRequest(s.contract, s.from))
	s.Require().ErrorIs(err, ErrUnknown)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test32SetProtocolFeeControllerNotOwner Result", s.contract)
	s.T().Logf("\tRequest: controller=%s", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test33UpdateDynamicLPFeeUnauthorized() {
	revertTx, revertHash, err := s.poolManager.UpdateDynamicLPFee(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewUpdateDynamicLPFeeRequest(s.contract, s.key, big.NewInt(500)))
	s.Require().ErrorIs(err, ErrUnauthorizedDynamicLPFeeUpdate)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test33UpdateDynamicLPFeeUnauthorized Result", s.contract)
	s.T().Logf("\tKey: fee=%s hooks=%s", s.key.Fee, s.key.Hooks)
	s.T().Logf("\tError: %v", err)
}

func (s *PoolManagerSuite) Test34ModifyLiquidityThroughTestRouter() {
	if s.poolSwapTest == nil || s.poolModifyLiquidityTest == nil {
		s.T().Skip("testenv: no PoolSwapTest/PoolModifyLiquidityTest configured for this network")
	}

	var salt [32]byte
	copy(salt[:], core.Keccak256([]byte("go-evmkit-v4-poolmanager-position")).Bytes())
	tickLower, tickUpper := big.NewInt(-600), big.NewInt(600)

	calldata, err := abi.NewFunction(
		"modifyLiquidity",
		abi.NewTypes(poolKeyTuple, liquidityParamsTuple, abi.Bytes, abi.Bool, abi.Bool),
		abi.NewTypes(abi.Bytes32),
	).EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{tickLower, tickUpper, s.LiquidityDelta, salt[:]},
		[]byte{}, false, false,
	)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, builder), s.poolModifyLiquidityTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "modifyLiquidity transaction reverted")

	mods, err := ExtractModifyLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(mods, 1)
	s.Require().True(mods[0].Contract.Equal(s.contract))
	s.Require().Equal(s.poolID, mods[0].Id)
	s.Require().True(mods[0].Sender.Equal(s.poolModifyLiquidityTest))
	s.Require().Zero(tickLower.Cmp(mods[0].TickLower))
	s.Require().Zero(tickUpper.Cmp(mods[0].TickUpper))
	s.Require().Zero(s.LiquidityDelta.Cmp(mods[0].LiquidityDelta))
	s.Require().Equal(salt, mods[0].Salt)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var liquidityWord [32]byte
	s.Require().Eventually(func() bool {
		liquidityWord, err = s.poolManager.Extsload(s.T().Context(), atBlock, NewExtsloadRequest(s.contract, s.liquiditySlot))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "liquidity not readable at the receipt block")
	s.Require().Zero(s.LiquidityDelta.Cmp(new(big.Int).SetBytes(liquidityWord[16:])))

	rangeWords, err := s.poolManager.ExtsloadRange(s.T().Context(), atBlock, NewExtsloadRangeRequest(s.contract, s.stateSlot, big.NewInt(4)))
	s.Require().NoError(err)
	s.Require().Len(rangeWords, 4)

	var feeGrowth0Slot, feeGrowth1Slot [32]byte
	new(big.Int).Add(new(big.Int).SetBytes(s.stateSlot[:]), big.NewInt(1)).FillBytes(feeGrowth0Slot[:])
	new(big.Int).Add(new(big.Int).SetBytes(s.stateSlot[:]), big.NewInt(2)).FillBytes(feeGrowth1Slot[:])

	slotWords, err := s.poolManager.ExtsloadSlots(s.T().Context(), atBlock, NewExtsloadSlotsRequest(s.contract, [][32]byte{s.stateSlot, feeGrowth0Slot, feeGrowth1Slot, s.liquiditySlot}))
	s.Require().NoError(err)
	s.Require().Equal(rangeWords, slotWords)
	s.Require().Equal(liquidityWord, rangeWords[3])

	transient, err := s.poolManager.ExttloadSlots(s.T().Context(), atBlock, NewExttloadSlotsRequest(s.contract, [][32]byte{s.stateSlot, s.liquiditySlot}))
	s.Require().NoError(err)
	s.Require().Equal([][32]byte{{}, {}}, transient)

	s.T().Logf("PoolManager %s Test34ModifyLiquidityThroughTestRouter Result", s.contract)
	s.T().Logf("\tPoolModifyLiquidityTest: %s", s.poolModifyLiquidityTest)
	s.T().Logf("\tModifyLiquidity event: id=%x sender=%s tickLower=%s tickUpper=%s liquidityDelta=%s salt=%x", mods[0].Id, mods[0].Sender, mods[0].TickLower, mods[0].TickUpper, mods[0].LiquidityDelta, mods[0].Salt)
	s.T().Logf("\tPool liquidity at receipt block: %s", new(big.Int).SetBytes(liquidityWord[16:]))
	s.T().Logf("\tExtsloadRange(stateSlot, 4): %x", rangeWords)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolManagerSuite) Test35SwapThroughTestRouter() {
	if s.poolSwapTest == nil || s.poolModifyLiquidityTest == nil {
		s.T().Skip("testenv: no PoolSwapTest/PoolModifyLiquidityTest configured for this network")
	}

	beforeSlot0, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.contract, s.stateSlot))
	s.Require().NoError(err)
	beforeWord := new(big.Int).SetBytes(beforeSlot0[:])
	beforePrice := new(big.Int).And(beforeWord, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)))
	beforeTick := new(big.Int).And(new(big.Int).Rsh(beforeWord, 160), big.NewInt(0xffffff))
	if beforeTick.Bit(23) == 1 {
		beforeTick.Sub(beforeTick, big.NewInt(1<<24))
	}

	calldata, err := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	).EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), big.NewInt(4295128740)},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	sentTx, hash, err := bind.Transact(s.T().Context(), s.client, bind.NewTransactOptions(s.signer, builder), s.poolSwapTest, calldata)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "swap transaction reverted")

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.contract))
	s.Require().Equal(s.poolID, swaps[0].Id)
	s.Require().True(swaps[0].Sender.Equal(s.poolSwapTest))
	s.Require().NotZero(swaps[0].Amount0.Sign())
	s.Require().NotZero(swaps[0].Amount1.Sign())
	s.Require().NotEqual(swaps[0].Amount0.Sign(), swaps[0].Amount1.Sign())
	s.Require().Zero(new(big.Int).Abs(swaps[0].Amount0).Cmp(s.SwapAmountIn))
	s.Require().Zero(s.key.Fee.Cmp(swaps[0].Fee))
	s.Require().Zero(s.LiquidityDelta.Cmp(swaps[0].Liquidity))
	s.Require().Negative(swaps[0].SqrtPriceX96.Cmp(beforePrice))
	s.Require().LessOrEqual(swaps[0].Tick.Cmp(beforeTick), 0)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var afterSlot0 [32]byte
	s.Require().Eventually(func() bool {
		afterSlot0, err = s.poolManager.Extsload(s.T().Context(), atBlock, NewExtsloadRequest(s.contract, s.stateSlot))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "slot0 not readable at the receipt block")

	afterWord := new(big.Int).SetBytes(afterSlot0[:])
	afterPrice := new(big.Int).And(afterWord, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)))
	afterTick := new(big.Int).And(new(big.Int).Rsh(afterWord, 160), big.NewInt(0xffffff))
	if afterTick.Bit(23) == 1 {
		afterTick.Sub(afterTick, big.NewInt(1<<24))
	}
	afterLPFee := new(big.Int).And(new(big.Int).Rsh(afterWord, 208), big.NewInt(0xffffff))
	s.Require().Zero(swaps[0].SqrtPriceX96.Cmp(afterPrice))
	s.Require().Zero(swaps[0].Tick.Cmp(afterTick))
	s.Require().Zero(s.key.Fee.Cmp(afterLPFee))

	accrued, err := s.poolManager.ProtocolFeesAccrued(s.T().Context(), atBlock, NewProtocolFeesAccruedRequest(s.contract, s.currency0))
	s.Require().NoError(err)
	s.Require().Zero(accrued.Sign(), "protocol fee is zero, so nothing should accrue")

	s.T().Logf("PoolManager %s Test35SwapThroughTestRouter Result", s.contract)
	s.T().Logf("\tPoolSwapTest: %s", s.poolSwapTest)
	s.T().Logf("\tSwap event: id=%x sender=%s amount0=%s amount1=%s sqrtPriceX96=%s liquidity=%s tick=%s fee=%s", swaps[0].Id, swaps[0].Sender, swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Liquidity, swaps[0].Tick, swaps[0].Fee)
	s.T().Logf("\tSlot0 before: sqrtPriceX96=%s tick=%s", beforePrice, beforeTick)
	s.T().Logf("\tSlot0 after: sqrtPriceX96=%s tick=%s lpFee=%s", afterPrice, afterTick, afterLPFee)
	s.T().Logf("\tProtocolFeesAccrued(currency0): %s", accrued)
	s.T().Logf("\tTx: %s (block %d, gas %d)", hash, receipt.BlockNumber, receipt.GasUsed)
}

func (s *PoolManagerSuite) Test36SwapRejectsBadPriceLimit() {
	if s.poolSwapTest == nil || s.poolModifyLiquidityTest == nil {
		s.T().Skip("testenv: no PoolSwapTest/PoolModifyLiquidityTest configured for this network")
	}

	slot0, err := s.poolManager.Extsload(s.T().Context(), nil, NewExtsloadRequest(s.contract, s.stateSlot))
	s.Require().NoError(err)
	price := new(big.Int).And(new(big.Int).SetBytes(slot0[:]), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)))
	limit := new(big.Int).Add(price, big.NewInt(1))

	calldata, err := abi.NewFunction(
		"swap",
		abi.NewTypes(poolKeyTuple, swapAmountParamsTuple, abi.Tuple(abi.Bool, abi.Bool), abi.Bytes),
		abi.NewTypes(abi.Bytes32),
	).EncodeCall(
		[]any{s.key.Currency0, s.key.Currency1, s.key.Fee, s.key.TickSpacing, s.key.Hooks},
		[]any{true, new(big.Int).Neg(s.SwapAmountIn), limit},
		[]any{false, false},
		[]byte{},
	)
	s.Require().NoError(err)

	_, err = tx.NewDynamicFeeBuilder().Build(s.from, s.poolSwapTest, big.NewInt(0), calldata).Pack(s.T().Context(), s.client)
	s.Require().Error(err)
	wrapped := mapPoolManagerRevertError(err)
	s.Require().ErrorIs(wrapped, ErrPriceLimitAlreadyExceeded)

	s.T().Logf("PoolManager %s Test36SwapRejectsBadPriceLimit Result", s.contract)
	s.T().Logf("\tCurrent sqrtPriceX96: %s, limit %s", price, limit)
	s.T().Logf("\tError: %v", wrapped)
}

func (s *PoolManagerSuite) Test37SetOperator() {
	sentTx, hash, err := s.poolManager.SetOperator(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewSetOperatorRequest(s.contract, s.spender, true))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "setOperator transaction reverted")

	operatorSets, err := erc6909.ExtractOperatorSets(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(operatorSets, 1)
	s.Require().True(operatorSets[0].Contract.Equal(s.contract))
	s.Require().True(operatorSets[0].Owner.Equal(s.from))
	s.Require().True(operatorSets[0].Operator.Equal(s.spender))
	s.Require().True(operatorSets[0].Approved)

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var after bool
	s.Require().Eventually(func() bool {
		after, err = s.poolManager.IsOperator(s.T().Context(), atBlock, erc6909.NewIsOperatorRequest(s.contract, s.from, s.spender))
		return err == nil && after
	}, 20*time.Second, s.ReceiptPollInterval, "operator not visible at the receipt block")

	revokeTx, revokeHash, err := s.poolManager.SetOperator(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewSetOperatorRequest(s.contract, s.spender, false))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, revokeTx.MaxFee())
	s.mu.Unlock()

	revokeReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, revokeHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(revokeReceipt.Status, "revoke setOperator transaction reverted")

	revokeAtBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", revokeReceipt.BlockNumber))

	var revoked bool
	s.Require().Eventually(func() bool {
		revoked, err = s.poolManager.IsOperator(s.T().Context(), revokeAtBlock, erc6909.NewIsOperatorRequest(s.contract, s.from, s.spender))
		return err == nil && !revoked
	}, 20*time.Second, s.ReceiptPollInterval, "operator not revoked at the receipt block")

	s.T().Logf("PoolManager %s Test37SetOperator Result", s.contract)
	s.T().Logf("\tOwner: %s, Operator: %s", s.from, s.spender)
	s.T().Logf("\tSet tx: %s (block %d), IsOperator after: %v", hash, receipt.BlockNumber, after)
	s.T().Logf("\tRevoke tx: %s (block %d), IsOperator after: %v", revokeHash, revokeReceipt.BlockNumber, revoked)
}

func (s *PoolManagerSuite) Test38ApproveAndTransferZero() {
	id := s.currency0.Big()
	amount := big.NewInt(100)

	sentTx, hash, err := s.poolManager.Approve(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewApproveRequest(s.contract, s.spender, id, amount))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	approvals, err := erc6909.ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Spender.Equal(s.spender))
	s.Require().Zero(id.Cmp(approvals[0].Id))
	s.Require().Zero(amount.Cmp(approvals[0].Amount))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.poolManager.Allowance(s.T().Context(), atBlock, erc6909.NewAllowanceRequest(s.contract, s.from, s.spender, id))
		return err == nil && amount.Cmp(allowance) == 0
	}, 20*time.Second, s.ReceiptPollInterval, "allowance not visible at the receipt block")

	transferTx, transferHash, err := s.poolManager.Transfer(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewTransferRequest(s.contract, s.spender, id, big.NewInt(0)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, transferTx.MaxFee())
	s.mu.Unlock()

	transferReceipt, err := tx.WaitForReceipt(s.T().Context(), s.client, transferHash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")

	transfers, err := erc6909.ExtractTransfers(transferReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.from))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.spender))
	s.Require().Zero(id.Cmp(transfers[0].Id))
	s.Require().Zero(transfers[0].Amount.Sign())

	overTx, overHash, overErr := s.poolManager.Transfer(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewTransferRequest(s.contract, s.spender, id, big.NewInt(1)))
	s.Require().Error(overErr)
	s.Require().Nil(overTx)
	s.Require().Nil(overHash)

	s.T().Logf("PoolManager %s Test38ApproveAndTransferZero Result", s.contract)
	s.T().Logf("\tApprove tx: %s, Allowance(from, spender, %s): %s", hash, id, allowance)
	s.T().Logf("\tTransfer(0) tx: %s (block %d)", transferHash, transferReceipt.BlockNumber)
	s.T().Logf("\tTransfer(1) with zero balance error: %v", overErr)
}

func (s *PoolManagerSuite) Test39SupportsInterface() {
	erc165Supported, err := s.poolManager.SupportsInterface(s.T().Context(), nil, erc165.NewSupportsInterfaceRequest(s.contract, [4]byte{0x01, 0xff, 0xc9, 0xa7}))
	s.Require().NoError(err)
	s.Require().True(erc165Supported)

	erc6909Supported, err := s.poolManager.SupportsInterface(s.T().Context(), nil, erc165.NewSupportsInterfaceRequest(s.contract, [4]byte{0x0f, 0x63, 0x2f, 0xb3}))
	s.Require().NoError(err)
	s.Require().True(erc6909Supported)

	erc721Supported, err := s.poolManager.SupportsInterface(s.T().Context(), nil, erc165.NewSupportsInterfaceRequest(s.contract, [4]byte{0x80, 0xac, 0x58, 0xcd}))
	s.Require().NoError(err)
	s.Require().False(erc721Supported)

	invalidSupported, err := s.poolManager.SupportsInterface(s.T().Context(), nil, erc165.NewSupportsInterfaceRequest(s.contract, [4]byte{0xff, 0xff, 0xff, 0xff}))
	s.Require().NoError(err)
	s.Require().False(invalidSupported)

	s.T().Logf("PoolManager %s Test39SupportsInterface Result", s.contract)
	s.T().Logf("\tERC165 (0x01ffc9a7): %v", erc165Supported)
	s.T().Logf("\tERC6909 (0x0f632fb3): %v", erc6909Supported)
	s.T().Logf("\tERC721 (0x80ac58cd): %v", erc721Supported)
	s.T().Logf("\t0xffffffff: %v", invalidSupported)
}

func (s *PoolManagerSuite) Test40ClaimCallSimulations() {
	id := s.currency1.Big()
	opts := bind.NewCallOptions().WithFrom(s.from)

	approved, err := s.poolManager.CallApprove(s.T().Context(), opts, erc6909.NewApproveRequest(s.contract, s.spender, id, big.NewInt(5)))
	s.Require().NoError(err)
	s.Require().True(approved)

	operatorSet, err := s.poolManager.CallSetOperator(s.T().Context(), opts, erc6909.NewSetOperatorRequest(s.contract, s.spender, true))
	s.Require().NoError(err)
	s.Require().True(operatorSet)

	transferred, err := s.poolManager.CallTransfer(s.T().Context(), opts, erc6909.NewTransferRequest(s.contract, s.spender, id, big.NewInt(0)))
	s.Require().NoError(err)
	s.Require().True(transferred)

	transferredFrom, err := s.poolManager.CallTransferFrom(s.T().Context(), opts, erc6909.NewTransferFromRequest(s.contract, s.from, s.spender, id, big.NewInt(0)))
	s.Require().NoError(err)
	s.Require().True(transferredFrom)

	overTransfer, overErr := s.poolManager.CallTransfer(s.T().Context(), opts, erc6909.NewTransferRequest(s.contract, s.spender, id, big.NewInt(1)))
	s.Require().Error(overErr)
	s.Require().False(overTransfer)

	noAllowance, noAllowanceErr := s.poolManager.CallTransferFrom(s.T().Context(), bind.NewCallOptions().WithFrom(s.spender), erc6909.NewTransferFromRequest(s.contract, s.from, s.spender, id, big.NewInt(1)))
	s.Require().Error(noAllowanceErr)
	s.Require().False(noAllowance)

	allowance, err := s.poolManager.Allowance(s.T().Context(), nil, erc6909.NewAllowanceRequest(s.contract, s.from, s.spender, id))
	s.Require().NoError(err)
	s.Require().Zero(allowance.Sign(), "eth_call must not persist the approval")

	s.T().Logf("PoolManager %s Test40ClaimCallSimulations Result", s.contract)
	s.T().Logf("\tCallApprove: %v, CallSetOperator: %v", approved, operatorSet)
	s.T().Logf("\tCallTransfer(0): %v, CallTransferFrom(self, 0): %v", transferred, transferredFrom)
	s.T().Logf("\tCallTransfer(1) error: %v", overErr)
	s.T().Logf("\tCallTransferFrom(spender, 1) error: %v", noAllowanceErr)
}

func (s *PoolManagerSuite) Test41TransferFromSelfZero() {
	id := s.currency1.Big()

	sentTx, hash, err := s.poolManager.TransferFrom(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewTransferFromRequest(s.contract, s.from, s.spender, id, big.NewInt(0)))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.T().Context(), s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom transaction reverted")

	transfers, err := erc6909.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.from))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.spender))
	s.Require().Zero(id.Cmp(transfers[0].Id))
	s.Require().Zero(transfers[0].Amount.Sign())

	revertTx, revertHash, revertErr := s.poolManager.TransferFrom(s.T().Context(), bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), erc6909.NewTransferFromRequest(s.contract, s.from, s.spender, id, big.NewInt(1)))
	s.Require().Error(revertErr)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("PoolManager %s Test41TransferFromSelfZero Result", s.contract)
	s.T().Logf("\tTransferFrom(self, spender, %s, 0) tx: %s (block %d, gas %d)", id, hash, receipt.BlockNumber, receipt.GasUsed)
	s.T().Logf("\tTransferFrom(self, spender, %s, 1) error: %v", id, revertErr)
}
