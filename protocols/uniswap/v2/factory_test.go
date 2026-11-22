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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type FactorySuite struct {
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
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	signer   core.Signer
	deployer *deploy.NicksDeployer
	from     *types.Address
	factory  *types.Address
	nicks    *types.Address
	tokenA   *types.Address
	tokenB   *types.Address

	v2Factory *Factory

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestFactorySuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)
	if c.Addresses.Protocol.Uniswap.V2.Factory == "" {
		t.Skip("testenv: no Uniswap V2 factory configured for this network")
	}

	suite.Run(t, &FactorySuite{
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
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *FactorySuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v2Factory = NewFactory(s.client)

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
	}

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}

	s.T().Logf("SetupSuite: factory %s, from %s, sorted tokens (%s, %s)", s.factory, s.from, s.tokenA, s.tokenB)
}

func (s *FactorySuite) TearDownSuite() {
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

func (s *FactorySuite) Test01CreatePair() {
	before, err := s.v2Factory.GetPair(s.ctx, nil, NewGetPairRequest(s.factory, s.tokenA, s.tokenB))
	s.Require().NoError(err)
	s.Require().True(before.IsZero(), "tokenA/tokenB shouldn't have a pair yet")

	allPairsLengthBefore, err := s.v2Factory.AllPairsLength(s.ctx, nil, NewAllPairsLengthRequest(s.factory))
	s.Require().NoError(err)

	predicted, err := s.v2Factory.CallCreatePair(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreatePairRequest(s.factory, s.tokenA, s.tokenB))
	s.Require().NoError(err)

	sentTx, hash, err := s.v2Factory.CreatePair(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePairRequest(s.factory, s.tokenA, s.tokenB))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createPair transaction reverted")

	created, err := ExtractPairCreated(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.Require().True(created[0].Contract.Equal(s.factory))
	s.Require().True(created[0].Token0.Equal(s.tokenA))
	s.Require().True(created[0].Token1.Equal(s.tokenB))
	s.Require().True(created[0].Pair.Equal(predicted))
	s.Require().Zero(new(big.Int).Add(allPairsLengthBefore, big.NewInt(1)).Cmp(created[0].AllPairsLength))

	atBlock := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber))

	var after *types.Address
	s.Require().Eventually(func() bool {
		var err error
		after, err = s.v2Factory.GetPair(s.ctx, atBlock, NewGetPairRequest(s.factory, s.tokenA, s.tokenB))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "after not readable at the receipt block")
	s.Require().True(after.Equal(predicted))

	var reversed *types.Address
	s.Require().Eventually(func() bool {
		var err error
		reversed, err = s.v2Factory.GetPair(s.ctx, atBlock, NewGetPairRequest(s.factory, s.tokenB, s.tokenA))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "reversed not readable at the receipt block")
	s.Require().True(reversed.Equal(predicted))

	var allPairsLengthAfter *big.Int
	s.Require().Eventually(func() bool {
		var err error
		allPairsLengthAfter, err = s.v2Factory.AllPairsLength(s.ctx, atBlock, NewAllPairsLengthRequest(s.factory))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "allPairsLengthAfter not readable at the receipt block")
	s.Require().Zero(created[0].AllPairsLength.Cmp(allPairsLengthAfter))

	var last *types.Address
	s.Require().Eventually(func() bool {
		var err error
		last, err = s.v2Factory.AllPairs(s.ctx, atBlock, NewAllPairsRequest(s.factory, new(big.Int).Sub(allPairsLengthAfter, big.NewInt(1))))
		return err == nil
	}, 20*time.Second, s.ReceiptPollInterval, "last not readable at the receipt block")
	s.Require().True(last.Equal(predicted))

	s.T().Logf("UniswapV2Factory %s Test01CreatePair Result", s.factory)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tToken0: %s", s.tokenA)
	s.T().Logf("\tToken1: %s", s.tokenB)
	s.T().Logf("\tCallCreatePair predicted: %s", predicted)
	s.T().Logf("\tPairCreated event: token0=%s token1=%s pair=%s allPairsLength=%s", created[0].Token0, created[0].Token1, created[0].Pair, created[0].AllPairsLength)
	s.T().Logf("\tGetPair(A,B)/GetPair(B,A): %s/%s", after, reversed)
	s.T().Logf("\tAllPairs(%s): %s", new(big.Int).Sub(allPairsLengthAfter, big.NewInt(1)), last)
	s.T().Logf("\tAllPairsLength: %s -> %s", allPairsLengthBefore, allPairsLengthAfter)
	s.T().Logf("\tTx: %s (nonce %d, block %d, gas %d)", hash, sentTx.AddressNonce(), receipt.BlockNumber, receipt.GasUsed)
}

func (s *FactorySuite) Test02CreatePairExists() {
	revertTx, revertHash, err := s.v2Factory.CreatePair(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePairRequest(s.factory, s.tokenA, s.tokenB))
	s.Require().ErrorIs(err, ErrPairExists)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Factory %s Test02CreatePairExists Result", s.factory)
	s.T().Logf("\tCreatePair(%s, %s) error: %v", s.tokenA, s.tokenB, err)

	predicted, err := s.v2Factory.CallCreatePair(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewCreatePairRequest(s.factory, s.tokenB, s.tokenA))
	s.Require().ErrorIs(err, ErrPairExists)
	s.Require().Nil(predicted)

	s.T().Logf("\tCallCreatePair(%s, %s) error: %v", s.tokenB, s.tokenA, err)
}

func (s *FactorySuite) Test03CreatePairIdenticalAddresses() {
	revertTx, revertHash, err := s.v2Factory.CreatePair(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePairRequest(s.factory, s.tokenA, s.tokenA))
	s.Require().ErrorIs(err, ErrIdenticalAddresses)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Factory %s Test03CreatePairIdenticalAddresses Result", s.factory)
	s.T().Logf("\tRequest: tokenA=%s tokenB=%s", s.tokenA, s.tokenA)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test04CreatePairZeroAddress() {
	revertTx, revertHash, err := s.v2Factory.CreatePair(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewCreatePairRequest(s.factory, s.tokenA, types.ZeroAddress()))
	s.Require().ErrorIs(err, ErrZeroAddress)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Factory %s Test04CreatePairZeroAddress Result", s.factory)
	s.T().Logf("\tRequest: tokenA=%s tokenB=%s", s.tokenA, types.ZeroAddress())
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test05ReadMethods() {
	feeTo, err := s.v2Factory.FeeTo(s.ctx, nil, NewFeeToRequest(s.factory))
	s.Require().NoError(err)

	feeToSetter, err := s.v2Factory.FeeToSetter(s.ctx, nil, NewFeeToSetterRequest(s.factory))
	s.Require().NoError(err)

	allPairsLength, err := s.v2Factory.AllPairsLength(s.ctx, nil, NewAllPairsLengthRequest(s.factory))
	s.Require().NoError(err)
	s.Require().Positive(allPairsLength.Sign())

	first, err := s.v2Factory.AllPairs(s.ctx, nil, NewAllPairsRequest(s.factory, big.NewInt(0)))
	s.Require().NoError(err)
	s.Require().False(first.IsZero())

	last, err := s.v2Factory.AllPairs(s.ctx, nil, NewAllPairsRequest(s.factory, new(big.Int).Sub(allPairsLength, big.NewInt(1))))
	s.Require().NoError(err)
	s.Require().False(last.IsZero())

	s.T().Logf("UniswapV2Factory %s Test05ReadMethods Result", s.factory)
	s.T().Logf("\tFeeTo: %s", feeTo)
	s.T().Logf("\tFeeToSetter: %s", feeToSetter)
	s.T().Logf("\tAllPairsLength: %s", allPairsLength)
	s.T().Logf("\tAllPairs(0): %s", first)
	s.T().Logf("\tAllPairs(%s): %s", new(big.Int).Sub(allPairsLength, big.NewInt(1)), last)
}

func (s *FactorySuite) Test06SetFeeToForbidden() {
	feeToSetter, err := s.v2Factory.FeeToSetter(s.ctx, nil, NewFeeToSetterRequest(s.factory))
	s.Require().NoError(err)

	revertTx, revertHash, err := s.v2Factory.SetFeeTo(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetFeeToRequest(s.factory, s.from))
	s.Require().ErrorIs(err, ErrForbidden)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Factory %s Test06SetFeeToForbidden Result", s.factory)
	s.T().Logf("\tCaller: %s (feeToSetter is %s)", s.from, feeToSetter)
	s.T().Logf("\tRequest: feeTo=%s", s.from)
	s.T().Logf("\tError: %v", err)
}

func (s *FactorySuite) Test07SetFeeToSetterForbidden() {
	feeToSetter, err := s.v2Factory.FeeToSetter(s.ctx, nil, NewFeeToSetterRequest(s.factory))
	s.Require().NoError(err)

	revertTx, revertHash, err := s.v2Factory.SetFeeToSetter(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetFeeToSetterRequest(s.factory, s.from))
	s.Require().ErrorIs(err, ErrForbidden)
	s.Require().Nil(revertTx)
	s.Require().Nil(revertHash)

	s.T().Logf("UniswapV2Factory %s Test07SetFeeToSetterForbidden Result", s.factory)
	s.T().Logf("\tCaller: %s (feeToSetter is %s)", s.from, feeToSetter)
	s.T().Logf("\tRequest: feeToSetter=%s", s.from)
	s.T().Logf("\tError: %v", err)
}
