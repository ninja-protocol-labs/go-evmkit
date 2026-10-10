package tx

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

type SetCodeTxSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	FromHex       string
	ToHex         string
	PrivateKeyHex string
	TokenHex      string

	ctx    context.Context
	cancel context.CancelFunc
	client rpc.Client

	from  *types.Address
	to    *types.Address
	token *types.Address
	key   *types.PrivateKey
}

func TestSetCodeTxSuite(t *testing.T) {
	suite.Run(t, &SetCodeTxSuite{
		Enable:        false,
		RPCURL:        "https://ethereum-sepolia-rpc.publicnode.com",
		FromHex:       "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
		ToHex:         "0xc10B4374F9654187DeB5eE2d2715c935f5C1Cb02",
		PrivateKeyHex: "",
		TokenHex:      "0x779877A7B0D9E8603169DdbD7836e478b4624789",
	})
}

func (s *SetCodeTxSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 120*time.Second)
	s.client = rpc.NewClient(s.RPCURL, rpc.WithTimeout(15*time.Second))

	var err error
	s.from, err = types.NewAddressFromHex(s.FromHex)
	s.Require().NoError(err)
	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)
	s.token, err = types.NewAddressFromHex(s.TokenHex)
	s.Require().NoError(err)
	s.key, err = types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
}

func (s *SetCodeTxSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *SetCodeTxSuite) authorization() core.Authorization {
	authorityKey, err := types.GeneratePrivateKey()
	s.Require().NoError(err)

	auth := core.Authorization{
		ChainID: big.NewInt(1),
		Address: s.token,
		Nonce:   0,
	}
	s.Require().NoError(auth.Sign(authorityKey))
	return auth
}

func (s *SetCodeTxSuite) TestSubmitETHTransfer() {
	weiValue := new(big.Int).Mul(big.NewInt(1), big.NewInt(1_000_000_000_000)) // 0.000001 ETH
	cfg := NewSetCodeTxConfig(s.from, s.to, weiValue, nil, []core.Authorization{s.authorization()})

	hash, err := Submit(s.ctx, s.client, cfg, s.key)
	s.Require().NoError(err)
	s.T().Logf("tx hash: %s", hash)
}

func (s *SetCodeTxSuite) TestPackSignBroadcastTokenTransfer() {
	fn, err := abi.ParseFunction("transfer(address,uint256) returns (bool)", nil)
	s.Require().NoError(err)

	amount := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1 token (18 decimals)
	data, err := fn.EncodeCall(s.to, amount)
	s.Require().NoError(err)

	cfg := NewSetCodeTxConfig(s.from, s.token, big.NewInt(0), data, []core.Authorization{s.authorization()})

	txn, err := cfg.Pack(s.ctx, s.client)
	s.Require().NoError(err)

	s.Require().NoError(txn.Sign(s.key))

	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)

	hash, err := Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("tx hash: %s", hash)

	receipt, err := WaitForReceipt(s.ctx, s.client, hash, time.Second, 60*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status)
	s.Require().True(hash.Equal(receipt.TxHash))
	s.Require().Len(receipt.Logs, 1)

	log := receipt.Logs[0]
	s.Require().True(s.token.Equal(log.Address))
	s.Require().Len(log.Topics, 3)

	transferTopic := core.Keccak256([]byte("Transfer(address,address,uint256)"))
	s.Require().True(transferTopic.Equal(log.Topics[0]))

	wantFrom := types.NewHashFromBytes(append(make([]byte, 12), s.from.Bytes()...))
	s.Require().True(wantFrom.Equal(log.Topics[1]))

	wantTo := types.NewHashFromBytes(append(make([]byte, 12), s.to.Bytes()...))
	s.Require().True(wantTo.Equal(log.Topics[2]))

	wantData := make([]byte, 32)
	amount.FillBytes(wantData)
	s.Require().Equal(wantData, log.Data)
}
