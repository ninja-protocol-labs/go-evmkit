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

type CreateDeployerSuite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ByteCode            string
	Constructor         *abi.Function
	Args                []any

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	signer   core.Signer
	from     *types.Address
	deployer *CreateDeployer
	req      *CreateRequest

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestCreateDeployerSuite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &CreateDeployerSuite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		Args:                []any{"evmkit test ERC20", "EVK20", uint8(18)},
	})
}

func (s *CreateDeployerSuite) SetupSuite() {
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

	s.deployer = NewCreateDeployer(s.client)
	s.req = NewCreateRequest(NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args))
}

func (s *CreateDeployerSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("%s total fee spent: %s %s", "CreateDeployerSuite", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *CreateDeployerSuite) Test01Deploy() {
	deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), s.req)
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, deployTx.MaxFee())
	s.mu.Unlock()

	predicted, err := s.deployer.DeployAddress(deployTx)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")
	s.Require().NotNil(receipt.ContractAddress)
	s.Require().True(receipt.ContractAddress.Equal(predicted))

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", predicted.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 10*time.Second, s.ReceiptPollInterval, "no code visible at deployed contract %s", predicted)

	s.T().Logf("CreateDeployer Test01Deploy Result")
	s.T().Logf("\tAddress: %s", predicted)
	s.T().Logf("\tTx: %s (nonce %d, gas used %d)", hash, deployTx.AddressNonce(), receipt.GasUsed)
}
