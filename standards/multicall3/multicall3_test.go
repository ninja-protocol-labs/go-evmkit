package multicall3

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

type Multicall3Suite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	MulticallHex        string
	NicksFactoryHex     string
	ChainID             int
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ToHex               string
	Value               *big.Int
	ByteCode            string
	Constructor         *abi.Function
	Args                []any
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	to        *types.Address
	multicall *types.Address
	factory   *types.Address
	token     *types.Address

	multicall3 *Multicall3

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestMulticall3Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &Multicall3Suite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		MulticallHex:        c.Multicall,
		NicksFactoryHex:     c.NicksFactory,
		ChainID:             c.ChainID,
		NativeSymbol:        c.NativeSymbol,
		NativeDecimals:      c.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.ReceiptPollInterval,
		ReceiptTimeout:      c.ReceiptTimeout,
		ToHex:               "0x432311F8926294B3C525634d204a7F1145272282",
		Value:               big.NewInt(1),
		ByteCode:            testenv.ERC20ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		Args:                []any{"evmkit test ERC20", "EVK20", uint8(18)},
		Salt:                core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *Multicall3Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.multicall3 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.factory, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

	deployment := deploy.NewDeterministicContractDeployment(
		deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
		salt,
	)

	unsigned, err := s.deployer.CreateDeployment(s.ctx, tx.NewDynamicFeeBuilder(), s.from, s.factory, big.NewInt(0), deployment)
	s.Require().NoError(err)

	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.deployer.Deploy(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")

	token, err := s.deployer.DeployAddress(s.factory, deployment)
	s.Require().NoError(err)
	s.token = token

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 10*time.Second, s.ReceiptPollInterval, "no code visible at deployed token %s", token)

	s.T().Logf("SetupSuite: deployed token %s (nonce %d, tx %s)", token, unsigned.AddressNonce(), hash)
}

func (s *Multicall3Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.token != nil {
		balanceOfFn := abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		holders := []*types.Address{s.from, s.to}
		reads := make([]Call3, len(holders))
		for i, holder := range holders {
			calldata, err := balanceOfFn.EncodeCall(holder)
			s.Require().NoError(err)
			reads[i] = NewCall3(s.token, false, calldata)
		}

		results, err := s.multicall3.CallAggregate3(ctx, nil, NewAggregate3Request(s.multicall, reads))
		s.Require().NoError(err)

		var burns []Call3
		for i, holder := range holders {
			balance, err := balanceOfFn.DecodeSingleReturn[*big.Int](results[i].ReturnData)
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			calldata, err := burnFn.EncodeCall(holder, balance)
			s.Require().NoError(err)
			burns = append(burns, NewCall3(s.token, false, calldata))
		}

		if len(burns) > 0 {
			sentTx, hash, err := s.multicall3.Aggregate3(ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewAggregate3Request(s.multicall, burns))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, sentTx.MaxFee())
			s.mu.Unlock()

			receipt, err := tx.WaitForReceipt(ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(receipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %d balances of %s (tx %s)", len(burns), s.token, hash)
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("Multicall3Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Multicall3Suite) Test01Aggregate3MintsInOneTransaction() {
	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	balanceOfFn := abi.NewFunction("balanceOf", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	transferTopic := core.Keccak256([]byte("Transfer(address,address,uint256)"))

	mintTo, err := mintFn.EncodeCall(s.to, big.NewInt(1))
	s.Require().NoError(err)
	mintFrom, err := mintFn.EncodeCall(s.from, big.NewInt(2))
	s.Require().NoError(err)

	sentTx, hash, err := s.multicall3.Aggregate3(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.token, false, mintTo),
		NewCall3(s.token, false, mintFrom),
	}))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	var recipients []*types.Address
	var amounts []*big.Int
	for _, log := range receipt.Logs {
		if !log.Address.Equal(s.token) || len(log.Topics) != 3 || !log.Topics[0].Equal(transferTopic) {
			continue
		}
		s.Require().True(log.TxHash.Equal(hash))
		s.Require().True(types.NewAddressFromBytes(log.Topics[1].Bytes()[12:]).Equal(types.ZeroAddress()))

		vals, err := abi.Unpack(abi.NewTypes(abi.Uint256), log.Data)
		s.Require().NoError(err)
		recipients = append(recipients, types.NewAddressFromBytes(log.Topics[2].Bytes()[12:]))
		amounts = append(amounts, vals[0].(*big.Int))
	}
	s.Require().Len(recipients, 2)
	s.Require().True(recipients[0].Equal(s.to))
	s.Require().Zero(big.NewInt(1).Cmp(amounts[0]))
	s.Require().True(recipients[1].Equal(s.from))
	s.Require().Zero(big.NewInt(2).Cmp(amounts[1]))

	balanceOfTo, err := balanceOfFn.EncodeCall(s.to)
	s.Require().NoError(err)
	balanceOfFrom, err := balanceOfFn.EncodeCall(s.from)
	s.Require().NoError(err)
	reads := NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.token, false, balanceOfTo),
		NewCall3(s.token, false, balanceOfFrom),
	})

	before, err := s.multicall3.CallAggregate3(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), reads)
	s.Require().NoError(err)
	after, err := s.multicall3.CallAggregate3(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), reads)
	s.Require().NoError(err)

	toBefore, err := balanceOfFn.DecodeSingleReturn[*big.Int](before[0].ReturnData)
	s.Require().NoError(err)
	toAfter, err := balanceOfFn.DecodeSingleReturn[*big.Int](after[0].ReturnData)
	s.Require().NoError(err)
	fromBefore, err := balanceOfFn.DecodeSingleReturn[*big.Int](before[1].ReturnData)
	s.Require().NoError(err)
	fromAfter, err := balanceOfFn.DecodeSingleReturn[*big.Int](after[1].ReturnData)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(1).Cmp(new(big.Int).Sub(toAfter, toBefore)))
	s.Require().Zero(big.NewInt(2).Cmp(new(big.Int).Sub(fromAfter, fromBefore)))

	s.T().Logf("Multicall3 %s Test01Aggregate3MintsInOneTransaction Result", s.multicall)
	s.T().Logf("\tTx: %s (block %d, gas used %d)", hash, receipt.BlockNumber, receipt.GasUsed)
	s.T().Logf("\tTo(%s) balance: %s -> %s", s.to, toBefore, toAfter)
	s.T().Logf("\tFrom(%s) balance: %s -> %s", s.from, fromBefore, fromAfter)
}

func (s *Multicall3Suite) Test02Aggregate3RevertsWithCallFailed() {
	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)

	mintTo, err := mintFn.EncodeCall(s.to, big.NewInt(1))
	s.Require().NoError(err)

	sentTx, hash, err := s.multicall3.Aggregate3(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.token, false, mintTo),
		NewCall3(s.token, false, []byte{0xde, 0xad, 0xbe, 0xef}),
	}))
	s.Require().ErrorIs(err, ErrCallFailed)
	s.Require().Nil(sentTx)
	s.Require().Nil(hash)

	s.T().Logf("Multicall3 %s Test02Aggregate3RevertsWithCallFailed Result", s.multicall)
	s.T().Logf("\tError: %v", err)
}

func (s *Multicall3Suite) Test03Aggregate3ToleratesAllowedFailure() {
	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	transferTopic := core.Keccak256([]byte("Transfer(address,address,uint256)"))

	mintTo, err := mintFn.EncodeCall(s.to, big.NewInt(3))
	s.Require().NoError(err)

	sentTx, hash, err := s.multicall3.Aggregate3(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.token, true, []byte{0xde, 0xad, 0xbe, 0xef}),
		NewCall3(s.token, false, mintTo),
	}))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	var amounts []*big.Int
	for _, log := range receipt.Logs {
		if !log.Address.Equal(s.token) || len(log.Topics) != 3 || !log.Topics[0].Equal(transferTopic) {
			continue
		}
		s.Require().True(types.NewAddressFromBytes(log.Topics[2].Bytes()[12:]).Equal(s.to))

		vals, err := abi.Unpack(abi.NewTypes(abi.Uint256), log.Data)
		s.Require().NoError(err)
		amounts = append(amounts, vals[0].(*big.Int))
	}
	s.Require().Len(amounts, 1)
	s.Require().Zero(big.NewInt(3).Cmp(amounts[0]))

	s.T().Logf("Multicall3 %s Test03Aggregate3ToleratesAllowedFailure Result", s.multicall)
	s.T().Logf("\tTx: %s (block %d, %d logs)", hash, receipt.BlockNumber, len(receipt.Logs))
}

func (s *Multicall3Suite) Test04Aggregate3ValueSendsValue() {
	sentTx, hash, err := s.multicall3.Aggregate3Value(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewAggregate3ValueRequest(s.multicall, []Call3Value{
		NewCall3Value(s.to, false, s.Value, nil),
		NewCall3Value(s.multicall, false, big.NewInt(0), EncodeGetBlockNumber()),
	}))
	s.Require().NoError(err)
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, sentTx.MaxFee())
	s.totalFee.Add(s.totalFee, s.Value)
	s.mu.Unlock()

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	before, err := s.multicall3.GetEthBalance(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber-1)), NewGetEthBalanceRequest(s.multicall, s.to))
	s.Require().NoError(err)
	after, err := s.multicall3.GetEthBalance(s.ctx, bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", receipt.BlockNumber)), NewGetEthBalanceRequest(s.multicall, s.to))
	s.Require().NoError(err)
	s.Require().Zero(s.Value.Cmp(new(big.Int).Sub(after, before)))

	s.T().Logf("Multicall3 %s Test04Aggregate3ValueSendsValue Result", s.multicall)
	s.T().Logf("\tTx: %s (block %d)", hash, receipt.BlockNumber)
	s.T().Logf("\tTo(%s) balance: %s -> %s", s.to, before, after)
}

func (s *Multicall3Suite) Test05CallAggregate() {
	result, err := s.multicall3.CallAggregate(s.ctx, nil, NewAggregateRequest(s.multicall, []Call{
		NewCall(s.multicall, EncodeGetChainID()),
		NewCall(s.multicall, EncodeGetBlockNumber()),
	}))
	s.Require().NoError(err)
	s.Require().Len(result.ReturnData, 2)

	chainID, err := DecodeGetChainID(result.ReturnData[0])
	s.Require().NoError(err)
	s.Require().Equal(int64(s.ChainID), chainID.Int64())

	blockNumber, err := DecodeGetBlockNumber(result.ReturnData[1])
	s.Require().NoError(err)
	s.Require().Zero(result.BlockNumber.Cmp(blockNumber))

	s.T().Logf("Multicall3 %s Test05CallAggregate Result", s.multicall)
	s.T().Logf("\tBlockNumber: %s", result.BlockNumber)
	s.T().Logf("\tChainID: %s", chainID)
}

func (s *Multicall3Suite) Test06CallAggregate3() {
	ethBalanceData, err := EncodeGetEthBalance(s.from)
	s.Require().NoError(err)

	results, err := s.multicall3.CallAggregate3(s.ctx, nil, NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.multicall, false, EncodeGetChainID()),
		NewCall3(s.multicall, true, []byte{0xde, 0xad, 0xbe, 0xef}),
		NewCall3(s.multicall, false, ethBalanceData),
	}))
	s.Require().NoError(err)
	s.Require().Len(results, 3)
	s.Require().True(results[0].Success)
	s.Require().False(results[1].Success)
	s.Require().True(results[2].Success)

	chainID, err := DecodeGetChainID(results[0].ReturnData)
	s.Require().NoError(err)
	s.Require().Equal(int64(s.ChainID), chainID.Int64())

	balance, err := DecodeGetEthBalance(results[2].ReturnData)
	s.Require().NoError(err)

	s.T().Logf("Multicall3 %s Test06CallAggregate3 Result", s.multicall)
	s.T().Logf("\tChainID: %s", chainID)
	s.T().Logf("\tFailing call: success=%v", results[1].Success)
	s.T().Logf("\tFrom(%s) balance: %s", s.from, balance)
}

func (s *Multicall3Suite) Test07CallAggregate3RevertsWithCallFailed() {
	results, err := s.multicall3.CallAggregate3(s.ctx, nil, NewAggregate3Request(s.multicall, []Call3{
		NewCall3(s.multicall, false, EncodeGetChainID()),
		NewCall3(s.multicall, false, []byte{0xde, 0xad, 0xbe, 0xef}),
	}))
	s.Require().ErrorIs(err, ErrCallFailed)
	s.Require().Nil(results)

	s.T().Logf("Multicall3 %s Test07CallAggregate3RevertsWithCallFailed Result", s.multicall)
	s.T().Logf("\tError: %v", err)
}

func (s *Multicall3Suite) Test08CallAggregate3Value() {
	results, err := s.multicall3.CallAggregate3Value(s.ctx, bind.NewCallOptions().WithFrom(s.from), NewAggregate3ValueRequest(s.multicall, []Call3Value{
		NewCall3Value(s.to, false, s.Value, nil),
		NewCall3Value(s.multicall, true, big.NewInt(0), EncodeGetChainID()),
	}))
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	s.Require().True(results[0].Success)
	s.Require().True(results[1].Success)

	chainID, err := DecodeGetChainID(results[1].ReturnData)
	s.Require().NoError(err)
	s.Require().Equal(int64(s.ChainID), chainID.Int64())

	s.T().Logf("Multicall3 %s Test08CallAggregate3Value Result", s.multicall)
	s.T().Logf("\tValue to %s: %s", s.to, s.Value)
	s.T().Logf("\tChainID: %s", chainID)
}

func (s *Multicall3Suite) Test09CallBlockAndAggregate() {
	result, err := s.multicall3.CallBlockAndAggregate(s.ctx, nil, NewBlockAndAggregateRequest(s.multicall, []Call{
		NewCall(s.multicall, EncodeGetBlockNumber()),
		NewCall(s.multicall, EncodeGetLastBlockHash()),
	}))
	s.Require().NoError(err)
	s.Require().Len(result.Results, 2)
	s.Require().True(result.Results[0].Success)
	s.Require().True(result.Results[1].Success)

	blockNumber, err := DecodeGetBlockNumber(result.Results[0].ReturnData)
	s.Require().NoError(err)
	s.Require().Zero(result.BlockNumber.Cmp(blockNumber))

	s.T().Logf("Multicall3 %s Test09CallBlockAndAggregate Result", s.multicall)
	s.T().Logf("\tBlockNumber: %s", result.BlockNumber)
	s.T().Logf("\tBlockHash: %s", result.BlockHash)
}

func (s *Multicall3Suite) Test10CallTryAggregate() {
	results, err := s.multicall3.CallTryAggregate(s.ctx, nil, NewTryAggregateRequest(s.multicall, false, []Call{
		NewCall(s.multicall, EncodeGetChainID()),
		NewCall(s.multicall, []byte{0xde, 0xad, 0xbe, 0xef}),
	}))
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	s.Require().True(results[0].Success)
	s.Require().False(results[1].Success)

	results, err = s.multicall3.CallTryAggregate(s.ctx, nil, NewTryAggregateRequest(s.multicall, true, []Call{
		NewCall(s.multicall, EncodeGetChainID()),
		NewCall(s.multicall, []byte{0xde, 0xad, 0xbe, 0xef}),
	}))
	s.Require().ErrorIs(err, ErrCallFailed)
	s.Require().Nil(results)

	s.T().Logf("Multicall3 %s Test10CallTryAggregate Result", s.multicall)
	s.T().Logf("\tRequireSuccess true: %v", err)
}

func (s *Multicall3Suite) Test11CallTryBlockAndAggregate() {
	blockNumber, err := s.multicall3.GetBlockNumber(s.ctx, nil, NewGetBlockNumberRequest(s.multicall))
	s.Require().NoError(err)

	opts := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", blockNumber.Uint64()))

	result, err := s.multicall3.CallTryBlockAndAggregate(s.ctx, opts, NewTryBlockAndAggregateRequest(s.multicall, false, []Call{
		NewCall(s.multicall, EncodeGetBlockNumber()),
		NewCall(s.multicall, []byte{0xde, 0xad, 0xbe, 0xef}),
	}))
	s.Require().NoError(err)
	s.Require().Zero(blockNumber.Cmp(result.BlockNumber))
	s.Require().Len(result.Results, 2)
	s.Require().True(result.Results[0].Success)
	s.Require().False(result.Results[1].Success)

	hash, err := s.multicall3.GetBlockHash(s.ctx, opts, NewGetBlockHashRequest(s.multicall, blockNumber))
	s.Require().NoError(err)
	s.Require().True(hash.Equal(result.BlockHash))

	s.T().Logf("Multicall3 %s Test11CallTryBlockAndAggregate Result", s.multicall)
	s.T().Logf("\tBlockNumber: %s", result.BlockNumber)
	s.T().Logf("\tBlockHash: %s", result.BlockHash)
}

func (s *Multicall3Suite) Test12Getters() {
	blockNumber, err := s.multicall3.GetBlockNumber(s.ctx, nil, NewGetBlockNumberRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Positive(blockNumber.Sign())

	opts := bind.NewCallOptions().WithBlock(fmt.Sprintf("0x%x", blockNumber.Uint64()))

	pinned, err := s.multicall3.GetBlockNumber(s.ctx, opts, NewGetBlockNumberRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Zero(blockNumber.Cmp(pinned))

	chainID, err := s.multicall3.GetChainID(s.ctx, opts, NewGetChainIDRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Equal(int64(s.ChainID), chainID.Int64())

	lastBlockHash, err := s.multicall3.GetLastBlockHash(s.ctx, opts, NewGetLastBlockHashRequest(s.multicall))
	s.Require().NoError(err)

	blockHash, err := s.multicall3.GetBlockHash(s.ctx, opts, NewGetBlockHashRequest(s.multicall, new(big.Int).Sub(blockNumber, big.NewInt(1))))
	s.Require().NoError(err)
	s.Require().True(blockHash.Equal(lastBlockHash))

	coinbase, err := s.multicall3.GetCurrentBlockCoinbase(s.ctx, opts, NewGetCurrentBlockCoinbaseRequest(s.multicall))
	s.Require().NoError(err)

	difficulty, err := s.multicall3.GetCurrentBlockDifficulty(s.ctx, opts, NewGetCurrentBlockDifficultyRequest(s.multicall))
	s.Require().NoError(err)

	gasLimit, err := s.multicall3.GetCurrentBlockGasLimit(s.ctx, opts, NewGetCurrentBlockGasLimitRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Positive(gasLimit.Sign())

	timestamp, err := s.multicall3.GetCurrentBlockTimestamp(s.ctx, opts, NewGetCurrentBlockTimestampRequest(s.multicall))
	s.Require().NoError(err)
	s.Require().Positive(timestamp.Sign())

	ethBalance, err := s.multicall3.GetEthBalance(s.ctx, opts, NewGetEthBalanceRequest(s.multicall, s.from))
	s.Require().NoError(err)

	basefee, err := s.multicall3.GetBasefee(s.ctx, opts, NewGetBasefeeRequest(s.multicall))
	if err != nil {
		s.Require().ErrorIs(err, ErrBasefeeNotImplemented)
	}

	s.T().Logf("Multicall3 %s Test12Getters Result", s.multicall)
	s.T().Logf("\tBlockNumber: %s", blockNumber)
	s.T().Logf("\tChainID: %s", chainID)
	s.T().Logf("\tLastBlockHash: %s", lastBlockHash)
	s.T().Logf("\tCoinbase: %s", coinbase)
	s.T().Logf("\tDifficulty: %s", difficulty)
	s.T().Logf("\tGasLimit: %s", gasLimit)
	s.T().Logf("\tTimestamp: %s", timestamp)
	s.T().Logf("\tFrom(%s) balance: %s", s.from, ethBalance)
	s.T().Logf("\tBasefee: %v (err %v)", basefee, err)
}
