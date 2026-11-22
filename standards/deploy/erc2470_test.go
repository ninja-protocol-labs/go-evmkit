package deploy

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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

type ERC2470DeployerSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	FactoryHex          string
	ByteCode            string
	Constructor         *abi.Function
	Args                []any
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	signer   core.Signer
	from     *types.Address
	deployer *ERC2470Deployer
	req      *Create2Request

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC2470DeployerSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &ERC2470DeployerSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		FactoryHex:          DefaultERC2470SingletonFactory.String(),
		ByteCode:            testenv.ERC721ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String), nil),
		Args:                []any{"evmkit test NFT", "EVKNFT"},
		Salt:                core.Keccak256([]byte("go-evmkit-erc2470")),
	})
}

func (s *ERC2470DeployerSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	factory, err := types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	var factoryCode string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetCode("code", factory.String(), rpc.BlockTagLatest, &factoryCode)))
	if factoryCode == "0x" || factoryCode == "" {
		s.T().Skipf("no ERC-2470 singleton factory at %s on this network", factory)
	}

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

	s.deployer = NewERC2470Deployer(s.client)
	s.req = NewCreate2Request(factory, NewDeterministicContractDeployment(NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args), salt))
}

func (s *ERC2470DeployerSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("%s total fee spent: %s %s", "ERC2470DeployerSuite", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC2470DeployerSuite) Test01CallDeploy() {
	predicted, err := s.deployer.DeployAddress(s.req)
	s.Require().NoError(err)

	simulated, err := s.deployer.CallDeploy(s.ctx, bind.NewCallOptions().WithFrom(s.from), s.req)
	s.Require().NoError(err)
	s.Require().True(simulated.Equal(predicted))

	s.T().Logf("ERC2470Deployer Test01CallDeploy Result")
	s.T().Logf("\tAddress: %s", simulated)
}

func (s *ERC2470DeployerSuite) Test02Deploy() {
	predicted, err := s.deployer.DeployAddress(s.req)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.LimitBuffer(150)
	deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, builder), s.req)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, deployTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", predicted.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 10*time.Second, s.ReceiptPollInterval, "no code visible at deployed contract %s", predicted)

	s.T().Logf("ERC2470Deployer Test02Deploy Result")
	s.T().Logf("\tAddress: %s", predicted)
	s.T().Logf("\tTx: %s (nonce %d, gas used %d)", hash, deployTx.AddressNonce(), receipt.GasUsed)
}

func (s *ERC2470DeployerSuite) Test03CallDeployAgainFails() {
	addr, err := s.deployer.CallDeploy(s.ctx, bind.NewCallOptions().WithFrom(s.from), s.req)
	s.Require().ErrorIs(err, ErrDeploymentFailed)
	s.Require().Nil(addr)

	s.T().Logf("ERC2470Deployer Test03CallDeployAgainFails Result")
	s.T().Logf("\tError: %v", err)
}
