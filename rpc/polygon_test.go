package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type polygonNetwork struct {
	name      string
	client    Client
	address   string
	blockHash string
	txHash    string
}

type PolygonRPCSuite struct {
	suite.Suite

	Enable bool

	ctx    context.Context
	cancel context.CancelFunc

	mainnet polygonNetwork
	testnet polygonNetwork
}

func TestPolygonRPCSuite(t *testing.T) {
	suite.Run(t, &PolygonRPCSuite{
		Enable: false,
	})
}

func (s *PolygonRPCSuite) SetupSuite() {
	if !s.Enable && !GlobalEnable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.mainnet = polygonNetwork{
		name:      "mainnet",
		client:    NewClient("https://polygon-bor-rpc.publicnode.com", WithTimeout(15*time.Second)),
		blockHash: "0x990b8baf08bf3b7d129a1dd0785ad6d6de58da4a8c750a1a4c1be40b7ab25421",
		txHash:    "0x847177caf9c18d3c9d8461b0347c21ef1287894bf9497e007b036216f8dd0ab4",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
	s.testnet = polygonNetwork{
		name:      "testnet",
		client:    NewClient("https://polygon-amoy-bor-rpc.publicnode.com", WithTimeout(15*time.Second)),
		blockHash: "0xc7613fea247bf3c240ab3c70c1b5c031628e5f34d40452af69ad7e4f37034517",
		txHash:    "0x3b6fa86f2b0b0d4fbff2eb9fc31730dd7fc644c5f301840af8156b6b83beae00",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
}

func (s *PolygonRPCSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PolygonRPCSuite) networks() []polygonNetwork {
	return []polygonNetwork{s.mainnet, s.testnet}
}

func (s *PolygonRPCSuite) TestETHChainID() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHChainID(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGasPrice() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGasPrice(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHBlockNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHBlockNumber(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBalance() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBalance(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetCode() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetCode(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHCall() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"to": n.address}
		elems := Elements{ETHCall(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHEstimateGas() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"from": n.address, "to": n.address, "value": "0x0"}
		elems := Elements{ETHEstimateGas(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetTransactionCount() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetTransactionCount(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHMaxPriorityFeePerGas() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHMaxPriorityFeePerGas(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBlockByNumber() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByNumber(uuid.NewString(), "latest", true, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBlockByHash() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByHash(uuid.NewString(), n.blockHash, false, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetTransactionByHash() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionByHash(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetTransactionReceipt() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionReceipt(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetTransactionByBlockHashAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockHashAndIndex(uuid.NewString(), n.blockHash, "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetTransactionByBlockNumberAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockNumberAndIndex(uuid.NewString(), "latest", "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBlockReceipts() {
	for _, n := range s.networks() {
		var result []map[string]any
		elems := Elements{ETHGetBlockReceipts(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBlockTransactionCountByHash() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByHash(uuid.NewString(), n.blockHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetBlockTransactionCountByNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByNumber(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHGetStorageAt() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetStorageAt(uuid.NewString(), n.address, "0x0", "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHFeeHistory() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHFeeHistory(uuid.NewString(), "0x4", "latest", []float64{25, 50, 75}, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestETHSyncing() {
	for _, n := range s.networks() {
		var result any
		elems := Elements{ETHSyncing(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestNetVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{NetVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestWeb3ClientVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{Web3ClientVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *PolygonRPCSuite) TestBatchEIP1559FeeData() {
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
