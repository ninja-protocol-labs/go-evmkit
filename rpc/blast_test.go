package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type blastNetwork struct {
	name      string
	client    *Client
	address   string
	blockHash string
	txHash    string
}

type BlastRPCSuite struct {
	suite.Suite

	Enable bool

	ctx    context.Context
	cancel context.CancelFunc

	mainnet blastNetwork
	testnet blastNetwork
}

func TestBlastRPCSuite(t *testing.T) {
	suite.Run(t, &BlastRPCSuite{
		Enable: false,
	})
}

func (s *BlastRPCSuite) SetupSuite() {
	if !s.Enable && !GlobalEnable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	s.mainnet = blastNetwork{
		name:      "mainnet",
		client:    NewClient("https://rpc.blast.io", WithTimeout(15*time.Second)),
		blockHash: "0x46ac9af92aac154640672b65f4a07cbe7961b638fcdce0444910f539698f6752",
		txHash:    "0x8173a2989a4217ada80306d6766c16be482d4285aa8f072f2bb17d2407ff7ab1",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
	s.testnet = blastNetwork{
		name:      "testnet",
		client:    NewClient("https://sepolia.blast.io", WithTimeout(15*time.Second)),
		blockHash: "0x97c633c6dc82069d37232877c969df2984cd03f8983e583b993d9d6b09f6cbe7",
		txHash:    "0x1312bf840cefed360bd018369a9e9d1807fdcc58ff4edd292292b5e570493b54",
		address:   "0x833e1D0b8Bc979D49d57b65dCF18364694B16D52",
	}
}

func (s *BlastRPCSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *BlastRPCSuite) networks() []blastNetwork {
	return []blastNetwork{s.mainnet, s.testnet}
}

func (s *BlastRPCSuite) TestETHChainID() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHChainID(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGasPrice() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGasPrice(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHBlockNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHBlockNumber(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBalance() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBalance(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetCode() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetCode(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHCall() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"to": n.address}
		elems := Elements{ETHCall(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHEstimateGas() {
	for _, n := range s.networks() {
		var result string
		params := map[string]any{"from": n.address, "to": n.address, "value": "0x0"}
		elems := Elements{ETHEstimateGas(uuid.NewString(), params, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetTransactionCount() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetTransactionCount(uuid.NewString(), n.address, "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHMaxPriorityFeePerGas() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHMaxPriorityFeePerGas(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBlockByNumber() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByNumber(uuid.NewString(), "latest", true, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBlockByHash() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetBlockByHash(uuid.NewString(), n.blockHash, false, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetTransactionByHash() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionByHash(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetTransactionReceipt() {
	for _, n := range s.networks() {
		if n.txHash == "" {
			continue
		}
		var result map[string]any
		elems := Elements{ETHGetTransactionReceipt(uuid.NewString(), n.txHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetTransactionByBlockHashAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockHashAndIndex(uuid.NewString(), n.blockHash, "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetTransactionByBlockNumberAndIndex() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHGetTransactionByBlockNumberAndIndex(uuid.NewString(), "latest", "0x0", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBlockReceipts() {
	for _, n := range s.networks() {
		var result []map[string]any
		elems := Elements{ETHGetBlockReceipts(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBlockTransactionCountByHash() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByHash(uuid.NewString(), n.blockHash, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetBlockTransactionCountByNumber() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetBlockTransactionCountByNumber(uuid.NewString(), "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHGetStorageAt() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{ETHGetStorageAt(uuid.NewString(), n.address, "0x0", "latest", &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHFeeHistory() {
	for _, n := range s.networks() {
		var result map[string]any
		elems := Elements{ETHFeeHistory(uuid.NewString(), "0x4", "latest", []float64{25, 50, 75}, &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestETHSyncing() {
	for _, n := range s.networks() {
		var result any
		elems := Elements{ETHSyncing(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestNetVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{NetVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestWeb3ClientVersion() {
	for _, n := range s.networks() {
		var result string
		elems := Elements{Web3ClientVersion(uuid.NewString(), &result)}
		s.Require().NoError(n.client.Batch(s.ctx, elems), n.name)
	}
}

func (s *BlastRPCSuite) TestBatchEIP1559FeeData() {
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
