package rpc

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/suite"
)

type robinhoodNetwork struct {
	name      string
	client    Client
	address   string
	blockHash string
	txHash    string
}

type RobinhoodRPCSuite struct {
	suite.Suite

	Enable bool

	ctx    context.Context
	cancel context.CancelFunc

	mainnet robinhoodNetwork
	testnet robinhoodNetwork
}

func TestRobinhoodRPCSuite(t *testing.T) {
	suite.Run(t, &RobinhoodRPCSuite{
		Enable: false,
	})
}

func (s *RobinhoodRPCSuite) SetupSuite() {
	if !s.Enable && !GlobalEnable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.mainnet = robinhoodNetwork{
		name:      "mainnet",
		client:    NewClient("https://rpc.mainnet.chain.robinhood.com", WithTimeout(15*time.Second)),
		blockHash: "0x20292dc32f1a6efc0e579f1e735084ffa5139c742ea608a2ac45c4ecaee35af0",
		txHash:    "0xc185fd028c7238e214dcdd58fcf330a581523e9e60b632ee5bb0fc9cf2d8e115",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
	s.testnet = robinhoodNetwork{
		name:      "testnet",
		client:    NewClient("https://rpc.testnet.chain.robinhood.com", WithTimeout(15*time.Second)),
		blockHash: "0x5af25250b8d032993973c747763e3210ff32429bfddbea1312e6663889bbc147",
		txHash:    "0xa81b330f5524a2c9cab21fe8a22da0ad03d18f8dd12800cc5f01046d1850e2b1",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
}

func (s *RobinhoodRPCSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RobinhoodRPCSuite) networks() []robinhoodNetwork {
	return []robinhoodNetwork{s.mainnet, s.testnet}
}

func (s *RobinhoodRPCSuite) TestETHChainID() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHChainID(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGasPrice() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGasPrice(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHBlockNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHBlockNumber(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBalance() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBalance(uuid.New().String(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetCode() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetCode(uuid.New().String(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHCall() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"to": n.address}
		elems := Elements{ETHCall(uuid.New().String(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHEstimateGas() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"from": n.address, "to": n.address, "value": "0x0"}
		elems := Elements{ETHEstimateGas(uuid.New().String(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetTransactionCount() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetTransactionCount(uuid.New().String(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHMaxPriorityFeePerGas() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHMaxPriorityFeePerGas(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBlockByNumber() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByNumber(uuid.New().String(), "latest", true, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBlockByHash() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByHash(uuid.New().String(), n.blockHash, false, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetTransactionByHash() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionByHash(uuid.New().String(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetTransactionReceipt() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionReceipt(uuid.New().String(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetTransactionByBlockHashAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockHashAndIndex(uuid.New().String(), n.blockHash, "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetTransactionByBlockNumberAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockNumberAndIndex(uuid.New().String(), "latest", "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBlockReceipts() {
	for _, n := range s.networks() {
		var result []map[string]any
		elems := Elements{ETHGetBlockReceipts(uuid.New().String(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBlockTransactionCountByHash() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByHash(uuid.New().String(), n.blockHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetBlockTransactionCountByNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByNumber(uuid.New().String(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHGetStorageAt() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetStorageAt(uuid.New().String(), n.address, "0x0", "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHFeeHistory() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHFeeHistory(uuid.New().String(), "0x4", "latest", []float64{25, 50, 75}, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestETHSyncing() {
	s.T().Skip("robinhood RPC does not support syncing")

	for _, n := range s.networks() {
		var result any
		elems := Elements{ETHSyncing(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestNetVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{NetVersion(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestWeb3ClientVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{Web3ClientVersion(uuid.New().String(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *RobinhoodRPCSuite) TestBatchEIP1559FeeData() {
	for _, n := range s.networks() {
		var nonce, maxPriorityFeePerGas, gasPrice, chainID string
		elems := Elements{
			ETHGetTransactionCount(uuid.New().String(), n.address, "latest", &nonce),
			ETHMaxPriorityFeePerGas(uuid.New().String(), &maxPriorityFeePerGas),
			ETHGasPrice(uuid.New().String(), &gasPrice),
			ETHChainID(uuid.New().String(), &chainID),
		}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
		s.Require().NotEmpty(nonce, n.name)
		s.Require().NotEmpty(maxPriorityFeePerGas, n.name)
		s.Require().NotEmpty(gasPrice, n.name)
		s.Require().NotEmpty(chainID, n.name)
	}
}
