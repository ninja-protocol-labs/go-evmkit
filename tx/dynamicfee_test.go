package tx

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

type DynamicFeeTxLiveSuite struct {
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

func TestDynamicFeeTxLiveSuite(t *testing.T) {
	suite.Run(t, &DynamicFeeTxLiveSuite{
		Enable:        false,
		RPCURL:        "https://sepolia.base.org",
		FromHex:       "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
		ToHex:         "0xc10B4374F9654187DeB5eE2d2715c935f5C1Cb02",
		PrivateKeyHex: "PRIVATE_KEY",
		TokenHex:      "0xE4aB69C077896252FAFBD49EFD26B5D171A32410",
	})
}

func (s *DynamicFeeTxLiveSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 30*time.Second)
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

func (s *DynamicFeeTxLiveSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

// TestSubmitETHTransfer packs, signs and broadcasts a plain 0.000001 ETH
// transfer in one call.
func (s *DynamicFeeTxLiveSuite) TestSubmitETHTransfer() {
	weiValue := new(big.Int).Mul(big.NewInt(1), big.NewInt(1_000_000_000_000)) // 0.000001 ETH
	cfg := NewDynamicFeeTxConfig(s.from, s.to, weiValue, nil)

	hash, err := Submit(s.ctx, s.client, cfg, s.key)
	s.Require().NoError(err)
	s.T().Logf("tx hash: %s", hash)
}

// TestPackSignBroadcastTokenTransfer packs an ERC-20 transfer of 1 token
// (18 decimals), then signs and broadcasts it as separate steps instead of
// via Submit.
func (s *DynamicFeeTxLiveSuite) TestPackSignBroadcastTokenTransfer() {
	fn, err := abi.ParseFunction("transfer(address,uint256) returns (bool)", nil)
	s.Require().NoError(err)

	amount := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1 token (18 decimals)
	data, err := fn.EncodeCall(s.to, amount)
	s.Require().NoError(err)

	cfg := NewDynamicFeeTxConfig(s.from, s.token, big.NewInt(0), data)

	txn, err := cfg.Pack(s.ctx, s.client)
	s.Require().NoError(err)

	s.Require().NoError(txn.Sign(s.key))

	raw, err := txn.EncodeRLP()
	s.Require().NoError(err)

	hash, err := Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)
	s.T().Logf("tx hash: %s", hash)
}
