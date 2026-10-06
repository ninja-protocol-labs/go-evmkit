package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type monadNetwork struct {
	name      string
	client    Client
	address   string
	blockHash string
	txHash    string
}

type MonadRPCSuite struct {
	suite.Suite

	Enable bool

	ctx    context.Context
	cancel context.CancelFunc

	mainnet monadNetwork
}

func TestMonadRPCSuite(t *testing.T) {
	suite.Run(t, &MonadRPCSuite{
		Enable: false,
	})
}

func (s *MonadRPCSuite) SetupSuite() {
	if !s.Enable && !GlobalEnable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.mainnet = monadNetwork{
		name:      "mainnet",
		client:    NewClient("https://rpc.monad.xyz", WithTimeout(15*time.Second)),
		blockHash: "0x84113057242ce22de3619269ba7c758f32e9c2176ee3082fe7655fe08684073d",
		txHash:    "0x59600ba436519a2fcbd8947eec5521a9875a0f644f16b68740304e2cb5d2d87b",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
}

func (s *MonadRPCSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *MonadRPCSuite) networks() []monadNetwork {
	return []monadNetwork{s.mainnet}
}

func (s *MonadRPCSuite) TestETHChainID() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHChainID(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGasPrice() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGasPrice(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHBlockNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHBlockNumber(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBalance() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBalance(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetCode() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetCode(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHCall() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"to": n.address}
		elems := Elements{ETHCall(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHEstimateGas() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"from": n.address, "to": n.address, "value": "0x0"}
		elems := Elements{ETHEstimateGas(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetTransactionCount() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetTransactionCount(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHMaxPriorityFeePerGas() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHMaxPriorityFeePerGas(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBlockByNumber() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByNumber(uuid.NewString(), "latest", true, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBlockByHash() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByHash(uuid.NewString(), n.blockHash, false, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetTransactionByHash() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionByHash(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetTransactionReceipt() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionReceipt(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetTransactionByBlockHashAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockHashAndIndex(uuid.NewString(), n.blockHash, "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetTransactionByBlockNumberAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockNumberAndIndex(uuid.NewString(), "latest", "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBlockReceipts() {
	for _, n := range s.networks() {
		var result []map[string]any
		elems := Elements{ETHGetBlockReceipts(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBlockTransactionCountByHash() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByHash(uuid.NewString(), n.blockHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetBlockTransactionCountByNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByNumber(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHGetStorageAt() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetStorageAt(uuid.NewString(), n.address, "0x0", "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHFeeHistory() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHFeeHistory(uuid.NewString(), "0x4", "latest", []float64{25, 50, 75}, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestETHSyncing() {
	for _, n := range s.networks() {
		var result any
		elems := Elements{ETHSyncing(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestNetVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{NetVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestWeb3ClientVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{Web3ClientVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *MonadRPCSuite) TestBatchEIP1559FeeData() {
	for _, n := range s.networks() {
		var nonce, maxPriorityFeePerGas, gasPrice, chainID string
		elems := Elements{
			ETHGetTransactionCount(uuid.NewString(), n.address, "latest", &nonce),
			ETHMaxPriorityFeePerGas(uuid.NewString(), &maxPriorityFeePerGas),
			ETHGasPrice(uuid.NewString(), &gasPrice),
			ETHChainID(uuid.NewString(), &chainID),
		}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
		s.Require().NotEmpty(nonce, n.name)
		s.Require().NotEmpty(maxPriorityFeePerGas, n.name)
		s.Require().NotEmpty(gasPrice, n.name)
		s.Require().NotEmpty(chainID, n.name)
	}
}
