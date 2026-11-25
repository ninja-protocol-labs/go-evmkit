package v3

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

type FactorySuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	FactoryHex          string
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Fee                 *big.Int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Constructor         *abi.Function
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	factory   *types.Address
	multicall *types.Address
	nicks     *types.Address
	tokenA    *types.Address
	tokenB    *types.Address
	pool      *types.Address

	v3Factory *Factory
	erc20     *erc20.ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestFactorySuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V3.Factory == "" {
		t.Skip("testenv: no Uniswap V3 factory configured for this network")
	}

	suite.Run(t, &FactorySuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		FactoryHex:          c.Addresses.Protocol.Uniswap.V3.Factory,
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Fee:                 big.NewInt(FeeMedium),
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *FactorySuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Factory = NewFactory(s.client)
	s.erc20 = erc20.New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target UniswapV3Factory %s (fee tier %s)", s.factory, s.Fee)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.nicks, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit test token A", "evmkit test token B"}
	symbols := [2]string{"EVKA", "EVKB"}
	tokens := [2]*types.Address{}

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

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), deploy.NewCreate2Request(s.nicks, deployment))
		s.Require().NoError(err)

		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], deployTx.AddressNonce(), deployTx.MaxFee())

		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d (gas %d)", symbols[i], receipt.BlockNumber, receipt.GasUsed)

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

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, deployTx.AddressNonce(), hash)
	}

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s", s.tokenA, s.tokenB)
}

func (s *FactorySuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, nil, erc20.NewAddressBalancesRequest(s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB}, false))
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

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
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

		s.T().Logf("FactorySuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *FactorySuite) Test01ReadMethods() {
	owner, err := s.v3Factory.Owner(s.ctx, nil, NewOwnerRequest(s.factory))
	s.Require().NoError(err)

	tickSpacingLow, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, nil, NewFeeAmountTickSpacingRequest(s.factory, big.NewInt(FeeLow)))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(10).Cmp(tickSpacingLow))

	tickSpacingMedium, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, nil, NewFeeAmountTickSpacingRequest(s.factory, big.NewInt(FeeMedium)))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(60).Cmp(tickSpacingMedium))

	tickSpacingHigh, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, nil, NewFeeAmountTickSpacingRequest(s.factory, big.NewInt(FeeHigh)))
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(200).Cmp(tickSpacingHigh))

	pool, err := s.v3Factory.GetPool(s.ctx, nil, NewGetPoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().NoError(err)
	s.Require().True(pool.IsZero(), "tokenA/tokenB shouldn't have a pool yet")

	s.T().Logf("UniswapV3Factory %s Test01ReadMethods Result", s.factory)
	s.T().Logf("\tOwner: %s", owner)
	s.T().Logf("\tFeeAmountTickSpacing(%d): %s", FeeLow, tickSpacingLow)
	s.T().Logf("\tFeeAmountTickSpacing(%d): %s", FeeMedium, tickSpacingMedium)
	s.T().Logf("\tFeeAmountTickSpacing(%d): %s", FeeHigh, tickSpacingHigh)
	s.T().Logf("\tGetPool(tokenA, tokenB, %s): %s", s.Fee, pool)
}

func (s *FactorySuite) Test02CreatePool() {
	before, err := s.v3Factory.GetPool(s.ctx, nil, NewGetPoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().NoError(err)
	s.Require().True(before.IsZero(), "tokenA/tokenB shouldn't have a pool yet")

	predicted, err := s.v3Factory.CallCreatePool(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().NoError(err)
	s.T().Logf("Test02CreatePool: CallCreatePool predicted pool %s", predicted)

	createBuilder := tx.NewDynamicFeeBuilder()
	createBuilder.PriceBuffer(150)
	sentTx, hash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, createBuilder), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02CreatePool: sent createPool (nonce %d, maxFee %s), tx %s", sentTx.AddressNonce(), sentTx.MaxFee(), hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createPool transaction reverted")

	created, err := ExtractPoolCreated(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.Require().True(created[0].Contract.Equal(s.factory))
	s.Require().True(created[0].Token0.Equal(s.tokenA))
	s.Require().True(created[0].Token1.Equal(s.tokenB))
	s.Require().Zero(s.Fee.Cmp(created[0].Fee))
	s.Require().Zero(big.NewInt(60).Cmp(created[0].TickSpacing))
	s.Require().True(created[0].Pool.Equal(predicted))
	s.pool = created[0].Pool

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var after *types.Address
	s.Require().Eventually(func() bool {
		var err error
		after, err = s.v3Factory.GetPool(s.ctx, atBlock, NewGetPoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "after not readable at the receipt block")
	s.Require().True(after.Equal(predicted))

	var reversed *types.Address
	s.Require().Eventually(func() bool {
		var err error
		reversed, err = s.v3Factory.GetPool(s.ctx, atBlock, NewGetPoolRequest(s.factory, s.tokenB, s.tokenA, s.Fee))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reversed not readable at the receipt block")
	s.Require().True(reversed.Equal(predicted))

	s.T().Logf("UniswapV3Factory %s Test02CreatePool Result", s.factory)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tToken0: %s", s.tokenA)
	s.T().Logf("\tToken1: %s", s.tokenB)
	s.T().Logf("\tCallCreatePool predicted: %s", predicted)
	s.T().Logf("\tPoolCreated event: token0=%s token1=%s fee=%s tickSpacing=%s pool=%s", created[0].Token0, created[0].Token1, created[0].Fee, created[0].TickSpacing, created[0].Pool)
	s.T().Logf("\tGetPool(A,B)/GetPool(B,A): %s/%s", after, reversed)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *FactorySuite) Test03CreatePoolAlreadyExists() {
	revertTx, revertHash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenB, s.Fee))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test03CreatePoolAlreadyExists Result", s.factory)
	s.T().Logf("\tCreatePool(%s, %s, %s) error: %v", s.tokenA, s.tokenB, s.Fee, err)

	predicted, err := s.v3Factory.CallCreatePool(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreatePoolRequest(s.factory, s.tokenB, s.tokenA, s.Fee))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(predicted)

	s.T().Logf("\tCallCreatePool(%s, %s, %s) error: %v", s.tokenB, s.tokenA, s.Fee, err)
}

func (s *FactorySuite) Test04CreatePoolIdenticalTokens() {
	revertTx, revertHash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenA, s.Fee))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test04CreatePoolIdenticalTokens Result", s.factory)
	s.T().Logf("\tRequest: tokenA=%s tokenB=%s fee=%s", s.tokenA, s.tokenA, s.Fee)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test05CreatePoolZeroAddress() {
	revertTx, revertHash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePoolRequest(s.factory, s.tokenA, types.ZeroAddress(), s.Fee))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test05CreatePoolZeroAddress Result", s.factory)
	s.T().Logf("\tRequest: tokenA=%s tokenB=%s fee=%s", s.tokenA, types.ZeroAddress(), s.Fee)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test06CreatePoolUnsupportedFee() {
	unsupportedFee := big.NewInt(1234)

	tickSpacing, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, nil, NewFeeAmountTickSpacingRequest(s.factory, unsupportedFee))
	s.Require().NoError(err)
	s.Require().Zero(tickSpacing.Sign(), "fee %s shouldn't be an enabled tier", unsupportedFee)

	revertTx, revertHash, err := s.v3Factory.CreatePool(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePoolRequest(s.factory, s.tokenA, s.tokenB, unsupportedFee))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test06CreatePoolUnsupportedFee Result", s.factory)
	s.T().Logf("\tFeeAmountTickSpacing(%s): %s (disabled)", unsupportedFee, tickSpacing)
	s.T().Logf("\tRequest: tokenA=%s tokenB=%s fee=%s", s.tokenA, s.tokenB, unsupportedFee)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test07SetOwnerForbidden() {
	owner, err := s.v3Factory.Owner(s.ctx, nil, NewOwnerRequest(s.factory))
	s.Require().NoError(err)
	s.Require().False(owner.Equal(s.from), "s.from shouldn't already be the factory owner")

	revertTx, revertHash, err := s.v3Factory.SetOwner(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetOwnerRequest(s.factory, s.from))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test07SetOwnerForbidden Result", s.factory)
	s.T().Logf("\tCaller: %s (owner is %s)", s.from, owner)
	s.T().Logf("\tRequest: owner=%s", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test08EnableFeeAmountForbidden() {
	owner, err := s.v3Factory.Owner(s.ctx, nil, NewOwnerRequest(s.factory))
	s.Require().NoError(err)
	s.Require().False(owner.Equal(s.from), "s.from shouldn't already be the factory owner")

	newFee, newTickSpacing := big.NewInt(1234), big.NewInt(1)
	revertTx, revertHash, err := s.v3Factory.EnableFeeAmount(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewEnableFeeAmountRequest(s.factory, newFee, newTickSpacing))
	s.Require().ErrorIs(err, ErrReverted)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV3Factory %s Test08EnableFeeAmountForbidden Result", s.factory)
	s.T().Logf("\tCaller: %s (owner is %s)", s.from, owner)
	s.T().Logf("\tRequest: fee=%s tickSpacing=%s", newFee, newTickSpacing)
	s.T().Logf("\tError: %v", err)
}
