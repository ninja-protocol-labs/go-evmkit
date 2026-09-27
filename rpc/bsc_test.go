package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type bscNetwork struct {
	name      string
	client    *Client
	address   string
	blockHash string
	txHash    string
}

type BSCRPCSuite struct {
	suite.Suite

	Enable bool

	ctx    context.Context
	cancel context.CancelFunc

	mainnet bscNetwork
}

func TestBSCRPCSuite(t *testing.T) {
	suite.Run(t, &BSCRPCSuite{
		Enable: false,
	})
}

func (s *BSCRPCSuite) SetupSuite() {
	if !s.Enable && !GlobalEnable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.mainnet = bscNetwork{
		name:      "mainnet",
		client:    NewClient("https://bsc-dataseed.binance.org", WithTimeout(15*time.Second)),
		blockHash: "0x0d6b5075743c6f66bd7195f5da684a3779893c872faa30431db2b17a074dbf9f",
		txHash:    "0x4c7f8cd8c3a6277ab62982aa1963b1e13fac7e89bb5de1b0ca7e2dcab914f86e",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
}

func (s *BSCRPCSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *BSCRPCSuite) networks() []bscNetwork {
	return []bscNetwork{s.mainnet}
}

func (s *BSCRPCSuite) TestETHChainID() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHChainID(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGasPrice() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGasPrice(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHBlockNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHBlockNumber(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBalance() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBalance(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetCode() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetCode(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHCall() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"to": n.address}
		elems := Elements{ETHCall(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHEstimateGas() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"from": n.address, "to": n.address, "value": "0x0"}
		elems := Elements{ETHEstimateGas(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetTransactionCount() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetTransactionCount(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHMaxPriorityFeePerGas() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHMaxPriorityFeePerGas(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBlockByNumber() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByNumber(uuid.NewString(), "latest", true, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBlockByHash() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByHash(uuid.NewString(), n.blockHash, false, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetTransactionByHash() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionByHash(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetTransactionReceipt() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionReceipt(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetTransactionByBlockHashAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockHashAndIndex(uuid.NewString(), n.blockHash, "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetTransactionByBlockNumberAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockNumberAndIndex(uuid.NewString(), "latest", "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBlockReceipts() {
	for _, n := range s.networks() {
		var result []map[string]any
		elems := Elements{ETHGetBlockReceipts(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBlockTransactionCountByHash() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByHash(uuid.NewString(), n.blockHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetBlockTransactionCountByNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByNumber(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHGetStorageAt() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetStorageAt(uuid.NewString(), n.address, "0x0", "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHFeeHistory() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHFeeHistory(uuid.NewString(), "0x4", "latest", []float64{25, 50, 75}, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestETHSyncing() {
	for _, n := range s.networks() {
		var result any
		elems := Elements{ETHSyncing(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestNetVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{NetVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestWeb3ClientVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{Web3ClientVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BSCRPCSuite) TestBatchEIP1559FeeData() {
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
