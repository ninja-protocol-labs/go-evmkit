package multicall3

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestReadsSendCallAndDecode(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	opts := bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4")
	calls := []Call{NewCall(target, []byte{0x70, 0xa0, 0x82, 0x31}), NewCall(contract, EncodeGetChainID())}
	calls3 := []Call3{NewCall3(target, true, []byte{0x70, 0xa0, 0x82, 0x31}), NewCall3(contract, false, EncodeGetChainID())}

	uintWord, err := encoding.Hex.Decode(retUint256)
	require.NoError(t, err)
	hashWord, err := encoding.Hex.Decode(retBytes32)
	require.NoError(t, err)
	addressWord, err := encoding.Hex.Decode(retAddress)
	require.NoError(t, err)
	aggregateRet, err := encoding.Hex.Decode(retAggregate)
	require.NoError(t, err)
	blockAndAggregateRet, err := encoding.Hex.Decode(retBlockAndAggregate)
	require.NoError(t, err)
	resultsRet, err := encoding.Hex.Decode(resultsTwo)
	require.NoError(t, err)

	aggregateData, err := EncodeAggregate(calls)
	require.NoError(t, err)
	tryAggregateData, err := EncodeTryAggregate(true, calls)
	require.NoError(t, err)
	blockAndAggregateData, err := EncodeBlockAndAggregate(calls)
	require.NoError(t, err)
	tryBlockAndAggregateData, err := EncodeTryBlockAndAggregate(false, calls)
	require.NoError(t, err)
	aggregate3Data, err := EncodeAggregate3(calls3)
	require.NoError(t, err)
	getBlockHashData, err := EncodeGetBlockHash(big.NewInt(12345))
	require.NoError(t, err)
	getEthBalanceData, err := EncodeGetEthBalance(target)
	require.NoError(t, err)

	wantAggregate, err := DecodeAggregate(aggregateRet)
	require.NoError(t, err)
	wantBlockResults, err := DecodeBlockAndAggregate(blockAndAggregateRet)
	require.NoError(t, err)
	wantResults, err := DecodeResults(resultsRet)
	require.NoError(t, err)
	wantHash, err := DecodeGetBlockHash(hashWord)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(m *Multicall3) (any, error)
		want     any
	}{
		{
			"callAggregate",
			aggregateData,
			aggregateRet,
			func(m *Multicall3) (any, error) {
				return m.CallAggregate(t.Context(), opts, NewAggregateRequest(contract, calls))
			},
			wantAggregate,
		},
		{
			"callTryAggregate",
			tryAggregateData,
			resultsRet,
			func(m *Multicall3) (any, error) {
				return m.CallTryAggregate(t.Context(), opts, NewTryAggregateRequest(contract, true, calls))
			},
			wantResults,
		},
		{
			"callBlockAndAggregate",
			blockAndAggregateData,
			blockAndAggregateRet,
			func(m *Multicall3) (any, error) {
				return m.CallBlockAndAggregate(t.Context(), opts, NewBlockAndAggregateRequest(contract, calls))
			},
			wantBlockResults,
		},
		{
			"callTryBlockAndAggregate",
			tryBlockAndAggregateData,
			blockAndAggregateRet,
			func(m *Multicall3) (any, error) {
				return m.CallTryBlockAndAggregate(t.Context(), opts, NewTryBlockAndAggregateRequest(contract, false, calls))
			},
			wantBlockResults,
		},
		{
			"callAggregate3",
			aggregate3Data,
			resultsRet,
			func(m *Multicall3) (any, error) {
				return m.CallAggregate3(t.Context(), opts, NewAggregate3Request(contract, calls3))
			},
			wantResults,
		},
		{
			"getBasefee",
			EncodeGetBasefee(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetBasefee(t.Context(), opts, NewGetBasefeeRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getBlockHash",
			getBlockHashData,
			hashWord,
			func(m *Multicall3) (any, error) {
				return m.GetBlockHash(t.Context(), opts, NewGetBlockHashRequest(contract, big.NewInt(12345)))
			},
			wantHash,
		},
		{
			"getBlockNumber",
			EncodeGetBlockNumber(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetBlockNumber(t.Context(), opts, NewGetBlockNumberRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getChainId",
			EncodeGetChainID(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetChainID(t.Context(), opts, NewGetChainIDRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getCurrentBlockCoinbase",
			EncodeGetCurrentBlockCoinbase(),
			addressWord,
			func(m *Multicall3) (any, error) {
				return m.GetCurrentBlockCoinbase(t.Context(), opts, NewGetCurrentBlockCoinbaseRequest(contract))
			},
			types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"),
		},
		{
			"getCurrentBlockDifficulty",
			EncodeGetCurrentBlockDifficulty(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetCurrentBlockDifficulty(t.Context(), opts, NewGetCurrentBlockDifficultyRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getCurrentBlockGasLimit",
			EncodeGetCurrentBlockGasLimit(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetCurrentBlockGasLimit(t.Context(), opts, NewGetCurrentBlockGasLimitRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getCurrentBlockTimestamp",
			EncodeGetCurrentBlockTimestamp(),
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetCurrentBlockTimestamp(t.Context(), opts, NewGetCurrentBlockTimestampRequest(contract))
			},
			big.NewInt(12345),
		},
		{
			"getEthBalance",
			getEthBalanceData,
			uintWord,
			func(m *Multicall3) (any, error) {
				return m.GetEthBalance(t.Context(), opts, NewGetEthBalanceRequest(contract, target))
			},
			big.NewInt(12345),
		},
		{
			"getLastBlockHash",
			EncodeGetLastBlockHash(),
			hashWord,
			func(m *Multicall3) (any, error) {
				return m.GetLastBlockHash(t.Context(), opts, NewGetLastBlockHashRequest(contract))
			},
			wantHash,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, contract.String(), call["to"])
				require.Equal(t, from.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestCallAggregate3ValueSendsTotalValue(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	resultsRet, err := encoding.Hex.Decode(resultsTwo)
	require.NoError(t, err)
	wantResults, err := DecodeResults(resultsRet)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		opts      *bind.CallOptions
		calls     []Call3Value
		wantValue string
		wantFrom  any
		wantBlock string
	}{
		{
			"nil options",
			nil,
			[]Call3Value{NewCall3Value(target, true, big.NewInt(1_000_000_000_000_000_000), nil), NewCall3Value(target, false, big.NewInt(255), []byte{1})},
			"0xde0b6b3a76400ff",
			nil,
			rpc.BlockTagLatest,
		},
		{
			"from and block kept",
			bind.NewCallOptions().WithFrom(from).WithBlock("0x10"),
			[]Call3Value{NewCall3Value(target, true, big.NewInt(16), nil)},
			"0x10",
			from.String(),
			"0x10",
		},
		{
			"zero total",
			bind.NewCallOptions(),
			[]Call3Value{NewCall3Value(target, true, big.NewInt(0), nil), NewCall3Value(target, true, big.NewInt(0), nil)},
			"0x0",
			nil,
			rpc.BlockTagLatest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calldata, err := EncodeAggregate3Value(tc.calls)
			require.NoError(t, err)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				params := e.Params.([]any)
				call := params[0].(map[string]any)
				require.Equal(t, contract.String(), call["to"])
				require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
				require.Equal(t, tc.wantValue, call["value"])
				require.Equal(t, tc.wantFrom, call["from"])
				require.Equal(t, tc.wantBlock, params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(resultsRet)
				return nil
			}).Once()

			got, err := New(c).CallAggregate3Value(t.Context(), tc.opts, NewAggregate3ValueRequest(contract, tc.calls))
			require.NoError(t, err)
			require.Equal(t, wantResults, got)
			if tc.opts != nil {
				require.Nil(t, tc.opts.Value)
			}
		})
	}
}

func TestAggregate3ValueRejectsOptionsValue(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	req := NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, big.NewInt(1), nil)})
	m := New(mocks.NewClient(t))

	results, err := m.CallAggregate3Value(t.Context(), bind.NewCallOptions().WithValue(big.NewInt(1)), req)
	require.Nil(t, results)
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "multicall3: callAggregate3Value: invalid options: value must be nil")

	results, err = m.CallAggregate3Value(t.Context(), bind.NewCallOptions().WithValue(big.NewInt(0)), req)
	require.Nil(t, results)
	require.ErrorIs(t, err, bind.ErrInvalidOptions)

	signed, hash, err := m.Aggregate3Value(t.Context(), bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), req)
	require.Nil(t, signed)
	require.Nil(t, hash)
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "multicall3: aggregate3Value: invalid options: value must be nil")

	signed, hash, err = m.Aggregate3Value(t.Context(), nil, req)
	require.Nil(t, signed)
	require.Nil(t, hash)
	require.EqualError(t, err, "multicall3: aggregate3Value: bind: transact: invalid options: options is nil")
}

func TestReadsMapRevertToSentinel(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	callFailed, err := abi.Pack(abi.NewTypes(abi.String), ErrCallFailedString)
	require.NoError(t, err)
	valueMismatch, err := abi.Pack(abi.NewTypes(abi.String), ErrValueMismatchString)
	require.NoError(t, err)
	other, err := abi.Pack(abi.NewTypes(abi.String), "Multicall3: something else")
	require.NoError(t, err)
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}

	for _, tc := range []struct {
		name   string
		data   jsontext.Value
		want   []error
		wantNo []error
	}{
		{
			"call failed",
			jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(append([]byte{}, errorSel...), callFailed...)) + `"`),
			[]error{ErrCallFailed},
			[]error{ErrValueMismatch, ErrUnknown},
		},
		{
			"value mismatch",
			jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(append([]byte{}, errorSel...), valueMismatch...)) + `"`),
			[]error{ErrValueMismatch},
			[]error{ErrCallFailed, ErrUnknown},
		},
		{
			"other revert string",
			jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(append([]byte{}, errorSel...), other...)) + `"`),
			[]error{ErrUnknown},
			[]error{ErrCallFailed, ErrValueMismatch},
		},
		{
			"custom error selector",
			jsontext.Value(`"0xdeadbeef"`),
			[]error{ErrUnknown},
			[]error{ErrCallFailed, ErrValueMismatch},
		},
		{
			"no revert data",
			nil,
			nil,
			[]error{ErrCallFailed, ErrValueMismatch, ErrUnknown},
		},
		{
			"empty revert data",
			jsontext.Value(`"0x"`),
			nil,
			[]error{ErrCallFailed, ErrValueMismatch, ErrUnknown},
		},
		{
			"non-string revert data",
			jsontext.Value(`{"x":1}`),
			nil,
			[]error{ErrCallFailed, ErrValueMismatch, ErrUnknown},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: tc.data}).Once()

			results, err := New(c).CallAggregate3(t.Context(), nil, NewAggregate3Request(contract, []Call3{NewCall3(target, false, nil)}))
			require.Nil(t, results)
			require.ErrorContains(t, err, "multicall3: callAggregate3: ")
			for _, want := range tc.want {
				require.ErrorIs(t, err, want)
			}
			for _, wantNo := range tc.wantNo {
				require.NotErrorIs(t, err, wantNo)
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}
}

func TestGetBasefeeMapsResponseErrorToNotImplemented(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "invalid opcode: BASEFEE"}).Once()
	n, err := New(c).GetBasefee(t.Context(), nil, NewGetBasefeeRequest(contract))
	require.Nil(t, n)
	require.ErrorIs(t, err, ErrBasefeeNotImplemented)
	require.EqualError(t, err, "multicall3: getBasefee: basefee opcode not implemented: rpc error -32000: invalid opcode: BASEFEE")

	transport := errors.New("dial tcp: connection refused")
	c = mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(transport).Once()
	n, err = New(c).GetBasefee(t.Context(), nil, NewGetBasefeeRequest(contract))
	require.Nil(t, n)
	require.ErrorIs(t, err, transport)
	require.NotErrorIs(t, err, ErrBasefeeNotImplemented)
	require.EqualError(t, err, "multicall3: getBasefee: dial tcp: connection refused")
}

func TestReadsRejectMalformedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(5)
			m := New(c)

			agg, err := m.CallAggregate(t.Context(), nil, NewAggregateRequest(contract, []Call{NewCall(target, nil)}))
			require.Nil(t, agg)
			require.ErrorContains(t, err, "multicall3: callAggregate: ")

			block, err := m.CallBlockAndAggregate(t.Context(), nil, NewBlockAndAggregateRequest(contract, []Call{NewCall(target, nil)}))
			require.Nil(t, block)
			require.ErrorContains(t, err, "multicall3: callBlockAndAggregate: ")

			results, err := m.CallAggregate3(t.Context(), nil, NewAggregate3Request(contract, []Call3{NewCall3(target, true, nil)}))
			require.Nil(t, results)
			require.ErrorContains(t, err, "multicall3: callAggregate3: ")

			n, err := m.GetBlockNumber(t.Context(), nil, NewGetBlockNumberRequest(contract))
			require.Nil(t, n)
			require.ErrorContains(t, err, "multicall3: getBlockNumber: ")

			hash, err := m.GetLastBlockHash(t.Context(), nil, NewGetLastBlockHashRequest(contract))
			require.Nil(t, hash)
			require.ErrorContains(t, err, "multicall3: getLastBlockHash: ")
		})
	}
}

func TestClientMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	m := New(mocks.NewClient(t))
	ctx := t.Context()

	_, err := m.CallAggregate(ctx, nil, NewAggregateRequest(contract, nil))
	require.EqualError(t, err, "multicall3: callAggregate: "+NewAggregateRequest(contract, nil).Validate().Error())
	_, err = m.CallTryAggregate(ctx, nil, NewTryAggregateRequest(nil, true, []Call{NewCall(target, nil)}))
	require.EqualError(t, err, "multicall3: callTryAggregate: "+NewTryAggregateRequest(nil, true, []Call{NewCall(target, nil)}).Validate().Error())
	_, err = m.CallBlockAndAggregate(ctx, nil, NewBlockAndAggregateRequest(contract, []Call{{}}))
	require.EqualError(t, err, "multicall3: callBlockAndAggregate: "+NewBlockAndAggregateRequest(contract, []Call{{}}).Validate().Error())
	_, err = m.CallTryBlockAndAggregate(ctx, nil, &TryBlockAndAggregateRequest{})
	require.EqualError(t, err, "multicall3: callTryBlockAndAggregate: "+(&TryBlockAndAggregateRequest{}).Validate().Error())
	_, err = m.CallAggregate3(ctx, nil, NewAggregate3Request(contract, []Call3{NewCall3(target, true, nil), {}}))
	require.EqualError(t, err, "multicall3: callAggregate3: "+NewAggregate3Request(contract, []Call3{NewCall3(target, true, nil), {}}).Validate().Error())
	_, err = m.CallAggregate3Value(ctx, nil, NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, nil, nil)}))
	require.EqualError(t, err, "multicall3: callAggregate3Value: "+NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, nil, nil)}).Validate().Error())
	_, err = m.CallAggregate3Value(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), &Aggregate3ValueRequest{})
	require.EqualError(t, err, "multicall3: callAggregate3Value: "+(&Aggregate3ValueRequest{}).Validate().Error())

	_, _, err = m.Aggregate(ctx, nil, &AggregateRequest{})
	require.EqualError(t, err, "multicall3: aggregate: "+(&AggregateRequest{}).Validate().Error())
	_, _, err = m.TryAggregate(ctx, nil, NewTryAggregateRequest(contract, false, nil))
	require.EqualError(t, err, "multicall3: tryAggregate: "+NewTryAggregateRequest(contract, false, nil).Validate().Error())
	_, _, err = m.BlockAndAggregate(ctx, nil, &BlockAndAggregateRequest{})
	require.EqualError(t, err, "multicall3: blockAndAggregate: "+(&BlockAndAggregateRequest{}).Validate().Error())
	_, _, err = m.TryBlockAndAggregate(ctx, nil, NewTryBlockAndAggregateRequest(contract, true, []Call{NewCall(target, nil), {}}))
	require.EqualError(t, err, "multicall3: tryBlockAndAggregate: "+NewTryBlockAndAggregateRequest(contract, true, []Call{NewCall(target, nil), {}}).Validate().Error())
	_, _, err = m.Aggregate3(ctx, nil, NewAggregate3Request(nil, []Call3{NewCall3(target, true, nil)}))
	require.EqualError(t, err, "multicall3: aggregate3: "+NewAggregate3Request(nil, []Call3{NewCall3(target, true, nil)}).Validate().Error())
	_, _, err = m.Aggregate3Value(ctx, nil, NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, big.NewInt(-1), nil)}))
	require.EqualError(t, err, "multicall3: aggregate3Value: "+NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, big.NewInt(-1), nil)}).Validate().Error())

	_, err = m.GetBasefee(ctx, nil, &GetBasefeeRequest{})
	require.EqualError(t, err, "multicall3: getBasefee: "+(&GetBasefeeRequest{}).Validate().Error())
	_, err = m.GetBlockHash(ctx, nil, NewGetBlockHashRequest(contract, nil))
	require.EqualError(t, err, "multicall3: getBlockHash: "+NewGetBlockHashRequest(contract, nil).Validate().Error())
	_, err = m.GetBlockNumber(ctx, nil, NewGetBlockNumberRequest(nil))
	require.EqualError(t, err, "multicall3: getBlockNumber: "+NewGetBlockNumberRequest(nil).Validate().Error())
	_, err = m.GetChainID(ctx, nil, &GetChainIDRequest{})
	require.EqualError(t, err, "multicall3: getChainId: "+(&GetChainIDRequest{}).Validate().Error())
	_, err = m.GetCurrentBlockCoinbase(ctx, nil, &GetCurrentBlockCoinbaseRequest{})
	require.EqualError(t, err, "multicall3: getCurrentBlockCoinbase: "+(&GetCurrentBlockCoinbaseRequest{}).Validate().Error())
	_, err = m.GetCurrentBlockDifficulty(ctx, nil, NewGetCurrentBlockDifficultyRequest(nil))
	require.EqualError(t, err, "multicall3: getCurrentBlockDifficulty: "+NewGetCurrentBlockDifficultyRequest(nil).Validate().Error())
	_, err = m.GetCurrentBlockGasLimit(ctx, nil, NewGetCurrentBlockGasLimitRequest(nil))
	require.EqualError(t, err, "multicall3: getCurrentBlockGasLimit: "+NewGetCurrentBlockGasLimitRequest(nil).Validate().Error())
	_, err = m.GetCurrentBlockTimestamp(ctx, nil, NewGetCurrentBlockTimestampRequest(nil))
	require.EqualError(t, err, "multicall3: getCurrentBlockTimestamp: "+NewGetCurrentBlockTimestampRequest(nil).Validate().Error())
	_, err = m.GetEthBalance(ctx, nil, NewGetEthBalanceRequest(contract, nil))
	require.EqualError(t, err, "multicall3: getEthBalance: "+NewGetEthBalanceRequest(contract, nil).Validate().Error())
	_, err = m.GetLastBlockHash(ctx, nil, NewGetLastBlockHashRequest(nil))
	require.EqualError(t, err, "multicall3: getLastBlockHash: "+NewGetLastBlockHashRequest(nil).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = m.Aggregate3(ctx, nil, NewAggregate3Request(contract, []Call3{NewCall3(target, true, nil)}))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "multicall3: aggregate3: bind: transact: invalid options: options is nil")
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	calls := []Call{NewCall(target, []byte{0x70, 0xa0, 0x82, 0x31}), NewCall(contract, EncodeGetChainID())}
	calls3 := []Call3{NewCall3(target, true, []byte{0x70, 0xa0, 0x82, 0x31}), NewCall3(contract, false, nil)}
	calls3Value := []Call3Value{NewCall3Value(target, true, big.NewInt(1_000), nil), NewCall3Value(target, false, big.NewInt(234), []byte{1})}

	aggregateData, err := EncodeAggregate(calls)
	require.NoError(t, err)
	tryAggregateData, err := EncodeTryAggregate(false, calls)
	require.NoError(t, err)
	blockAndAggregateData, err := EncodeBlockAndAggregate(calls)
	require.NoError(t, err)
	tryBlockAndAggregateData, err := EncodeTryBlockAndAggregate(true, calls)
	require.NoError(t, err)
	aggregate3Data, err := EncodeAggregate3(calls3)
	require.NoError(t, err)
	aggregate3ValueData, err := EncodeAggregate3Value(calls3Value)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		calldata  []byte
		optsValue *big.Int
		wantValue int64
		send      func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"aggregate",
			aggregateData,
			nil,
			0,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.Aggregate(t.Context(), opts, NewAggregateRequest(contract, calls))
			},
		},
		{
			"tryAggregate",
			tryAggregateData,
			nil,
			0,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.TryAggregate(t.Context(), opts, NewTryAggregateRequest(contract, false, calls))
			},
		},
		{
			"blockAndAggregate",
			blockAndAggregateData,
			nil,
			0,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.BlockAndAggregate(t.Context(), opts, NewBlockAndAggregateRequest(contract, calls))
			},
		},
		{
			"tryBlockAndAggregate",
			tryBlockAndAggregateData,
			nil,
			0,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.TryBlockAndAggregate(t.Context(), opts, NewTryBlockAndAggregateRequest(contract, true, calls))
			},
		},
		{
			"aggregate3",
			aggregate3Data,
			nil,
			0,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.Aggregate3(t.Context(), opts, NewAggregate3Request(contract, calls3))
			},
		},
		{
			"aggregate3 with options value",
			aggregate3Data,
			big.NewInt(5),
			5,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.Aggregate3(t.Context(), opts, NewAggregate3Request(contract, calls3))
			},
		},
		{
			"aggregate3Value",
			aggregate3ValueData,
			nil,
			1_234,
			func(m *Multicall3, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return m.Aggregate3Value(t.Context(), opts, NewAggregate3ValueRequest(contract, calls3Value))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(200_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder)
			opts.Value = tc.optsValue

			signed, hash, err := tc.send(New(c), opts)
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())
			require.Equal(t, tc.optsValue, opts.Value)

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(contract))
			require.Equal(t, big.NewInt(tc.wantValue).String(), dyn.Value.String())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	reason, err := abi.Pack(abi.NewTypes(abi.String), ErrCallFailedString)
	require.NoError(t, err)
	revert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, reason...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := New(c).Aggregate3(t.Context(), bind.NewTransactOptions(key, builder), NewAggregate3Request(contract, []Call3{NewCall3(target, false, []byte{0xde, 0xad})}))
	require.ErrorIs(t, err, ErrCallFailed)
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(200_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "insufficient funds for gas * price + value"}).Twice()

	signed, hash, err := New(c).Aggregate3Value(t.Context(), bind.NewTransactOptions(key, builder), NewAggregate3ValueRequest(contract, []Call3Value{NewCall3Value(target, true, big.NewInt(77), nil)}))
	require.ErrorContains(t, err, "multicall3: aggregate3Value: ")
	require.ErrorContains(t, err, "insufficient funds")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	dyn, ok := signed.(*core.DynamicFeeTx)
	require.True(t, ok)
	require.Equal(t, "77", dyn.Value.String())
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "multicall3")
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	target := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	uintWord, err := encoding.Hex.Decode(retUint256)
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "multicall3", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(uintWord)
		}
		return nil
	}).Twice()

	_, err = New(c).GetBlockNumber(ctx, nil, NewGetBlockNumberRequest(contract))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(200_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).Aggregate3(ctx, bind.NewTransactOptions(key, builder), NewAggregate3Request(contract, []Call3{NewCall3(target, true, nil)}))
	require.NoError(t, err)
}

func TestMapRevertErrorPassesThroughNilAndUndecodable(t *testing.T) {
	require.NoError(t, MapRevertError(nil))
	var success *rpc.ResponseError
	require.NoError(t, MapRevertError(success))

	plain := fmt.Errorf("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
	empty := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	require.Same(t, empty, MapRevertError(empty))
}
