package multicall3

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestFunctionSelectors(t *testing.T) {
	tests := []struct {
		fn        *abi.Function
		signature string
		selector  string
	}{
		{aggregate, "aggregate((address,bytes)[])", "252dba42"},
		{aggregate3, "aggregate3((address,bool,bytes)[])", "82ad56cb"},
		{aggregate3Value, "aggregate3Value((address,bool,uint256,bytes)[])", "174dea71"},
		{blockAndAggregate, "blockAndAggregate((address,bytes)[])", "c3077fa9"},
		{tryAggregate, "tryAggregate(bool,(address,bytes)[])", "bce38bd7"},
		{tryBlockAndAggregate, "tryBlockAndAggregate(bool,(address,bytes)[])", "399542e9"},
		{getBasefee, "getBasefee()", "3e64a696"},
		{getBlockHash, "getBlockHash(uint256)", "ee82ac5e"},
		{getBlockNumber, "getBlockNumber()", "42cbb15c"},
		{getChainId, "getChainId()", "3408e470"},
		{getCurrentBlockCoinbase, "getCurrentBlockCoinbase()", "a8b0574e"},
		{getCurrentBlockDiff, "getCurrentBlockDifficulty()", "72425d9d"},
		{getCurrentBlockGasLimit, "getCurrentBlockGasLimit()", "86d516e8"},
		{getCurrentBlockTime, "getCurrentBlockTimestamp()", "0f28c97d"},
		{getEthBalance, "getEthBalance(address)", "4d2301cc"},
		{getLastBlockHash, "getLastBlockHash()", "27e86d6e"},
	}

	for _, tt := range tests {
		t.Run(tt.signature, func(t *testing.T) {
			require.Equal(t, tt.signature, tt.fn.Signature())
			sel := tt.fn.Selector()
			require.Equal(t, tt.selector, hex.EncodeToString(sel[:]))
		})
	}
}

func TestFunctionsMatchSolidityDeclarations(t *testing.T) {
	const (
		call3  = "(address target, bool allowFailure, bytes callData)"
		result = "(bool success, bytes returnData)"
	)

	parsed, err := abi.ParseFunction("function aggregate3("+call3+"[] calldata calls) external payable returns ("+result+"[] memory returnData)", nil)
	require.NoError(t, err)
	require.Equal(t, aggregate3.Signature(), parsed.Signature())
	require.Equal(t, aggregate3.Outputs, parsed.Outputs)
}

func TestRequestsValidate(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"aggregate nil", (*AggregateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"aggregate empty", (&AggregateRequest{}).Validate(), "invalid request: contract is nil"},
		{"aggregate nil contract", (&AggregateRequest{Calls: []Call{NewCall(a, nil)}}).Validate(), "invalid request: contract is nil"},
		{"aggregate nil calls", (&AggregateRequest{Contract: contract}).Validate(), "invalid request: calls is empty"},
		{"aggregate empty calls", (&AggregateRequest{Contract: contract, Calls: []Call{}}).Validate(), "invalid request: calls is empty"},
		{"aggregate nil first target", (&AggregateRequest{Contract: contract, Calls: []Call{{}, {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"aggregate nil second target", (&AggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(nil, []byte{1})}}).Validate(), "invalid request: calls[1].target is nil"},
		{"aggregate valid", (&AggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(b, []byte{1})}}).Validate(), ""},
		{"blockAndAggregate nil", (*BlockAndAggregateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"blockAndAggregate empty", (&BlockAndAggregateRequest{}).Validate(), "invalid request: contract is nil"},
		{"blockAndAggregate nil contract", (&BlockAndAggregateRequest{Calls: []Call{NewCall(a, nil)}}).Validate(), "invalid request: contract is nil"},
		{"blockAndAggregate nil calls", (&BlockAndAggregateRequest{Contract: contract}).Validate(), "invalid request: calls is empty"},
		{"blockAndAggregate empty calls", (&BlockAndAggregateRequest{Contract: contract, Calls: []Call{}}).Validate(), "invalid request: calls is empty"},
		{"blockAndAggregate nil first target", (&BlockAndAggregateRequest{Contract: contract, Calls: []Call{{}, {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"blockAndAggregate nil second target", (&BlockAndAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(nil, []byte{1})}}).Validate(), "invalid request: calls[1].target is nil"},
		{"blockAndAggregate valid", (&BlockAndAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(b, []byte{1})}}).Validate(), ""},
		{"tryAggregate nil", (*TryAggregateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tryAggregate empty", (&TryAggregateRequest{}).Validate(), "invalid request: contract is nil"},
		{"tryAggregate nil contract", (&TryAggregateRequest{RequireSuccess: true, Calls: []Call{NewCall(a, nil)}}).Validate(), "invalid request: contract is nil"},
		{"tryAggregate nil calls", (&TryAggregateRequest{Contract: contract, RequireSuccess: true}).Validate(), "invalid request: calls is empty"},
		{"tryAggregate nil first target", (&TryAggregateRequest{Contract: contract, Calls: []Call{{}, {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"tryAggregate nil second target", (&TryAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(nil, nil)}}).Validate(), "invalid request: calls[1].target is nil"},
		{"tryAggregate valid", (&TryAggregateRequest{Contract: contract, RequireSuccess: true, Calls: []Call{NewCall(a, nil)}}).Validate(), ""},
		{"tryAggregate valid without requireSuccess", (&TryAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil)}}).Validate(), ""},
		{"tryBlockAndAggregate nil", (*TryBlockAndAggregateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tryBlockAndAggregate empty", (&TryBlockAndAggregateRequest{}).Validate(), "invalid request: contract is nil"},
		{"tryBlockAndAggregate nil contract", (&TryBlockAndAggregateRequest{RequireSuccess: true, Calls: []Call{NewCall(a, nil)}}).Validate(), "invalid request: contract is nil"},
		{"tryBlockAndAggregate nil calls", (&TryBlockAndAggregateRequest{Contract: contract, RequireSuccess: true}).Validate(), "invalid request: calls is empty"},
		{"tryBlockAndAggregate nil first target", (&TryBlockAndAggregateRequest{Contract: contract, Calls: []Call{{}, {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"tryBlockAndAggregate nil second target", (&TryBlockAndAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil), NewCall(nil, nil)}}).Validate(), "invalid request: calls[1].target is nil"},
		{"tryBlockAndAggregate valid", (&TryBlockAndAggregateRequest{Contract: contract, RequireSuccess: true, Calls: []Call{NewCall(a, nil)}}).Validate(), ""},
		{"tryBlockAndAggregate valid without requireSuccess", (&TryBlockAndAggregateRequest{Contract: contract, Calls: []Call{NewCall(a, nil)}}).Validate(), ""},
		{"aggregate3 nil", (*Aggregate3Request)(nil).Validate(), "invalid request: request is nil"},
		{"aggregate3 empty", (&Aggregate3Request{}).Validate(), "invalid request: contract is nil"},
		{"aggregate3 nil contract", (&Aggregate3Request{Calls: []Call3{NewCall3(a, true, nil)}}).Validate(), "invalid request: contract is nil"},
		{"aggregate3 nil calls", (&Aggregate3Request{Contract: contract}).Validate(), "invalid request: calls is empty"},
		{"aggregate3 nil first target", (&Aggregate3Request{Contract: contract, Calls: []Call3{NewCall3(nil, true, nil), {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"aggregate3 nil second target", (&Aggregate3Request{Contract: contract, Calls: []Call3{NewCall3(a, false, nil), NewCall3(nil, true, []byte{1})}}).Validate(), "invalid request: calls[1].target is nil"},
		{"aggregate3 valid", (&Aggregate3Request{Contract: contract, Calls: []Call3{NewCall3(a, false, nil), NewCall3(b, true, []byte{1})}}).Validate(), ""},
		{"aggregate3Value nil", (*Aggregate3ValueRequest)(nil).Validate(), "invalid request: request is nil"},
		{"aggregate3Value empty", (&Aggregate3ValueRequest{}).Validate(), "invalid request: contract is nil"},
		{"aggregate3Value nil contract", (&Aggregate3ValueRequest{Calls: []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil)}}).Validate(), "invalid request: contract is nil"},
		{"aggregate3Value nil calls", (&Aggregate3ValueRequest{Contract: contract}).Validate(), "invalid request: calls is empty"},
		{"aggregate3Value nil first target", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(nil, true, big.NewInt(1), nil), {}}}).Validate(), "invalid request: calls[0].target is nil"},
		{"aggregate3Value nil first value", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(a, true, nil, nil), {}}}).Validate(), "invalid request: calls[0].value is nil"},
		{"aggregate3Value negative first value", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(a, true, big.NewInt(-1), nil), {}}}).Validate(), "invalid request: calls[0].value is negative"},
		{"aggregate3Value nil second target", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil), NewCall3Value(nil, true, nil, nil)}}).Validate(), "invalid request: calls[1].target is nil"},
		{"aggregate3Value nil second value", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil), NewCall3Value(b, true, nil, nil)}}).Validate(), "invalid request: calls[1].value is nil"},
		{"aggregate3Value valid", (&Aggregate3ValueRequest{Contract: contract, Calls: []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil), NewCall3Value(b, false, big.NewInt(0), []byte{1})}}).Validate(), ""},
		{"getBasefee nil", (*GetBasefeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getBasefee empty", (&GetBasefeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"getBasefee valid", (&GetBasefeeRequest{Contract: contract}).Validate(), ""},
		{"getBlockNumber nil", (*GetBlockNumberRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getBlockNumber empty", (&GetBlockNumberRequest{}).Validate(), "invalid request: contract is nil"},
		{"getBlockNumber valid", (&GetBlockNumberRequest{Contract: contract}).Validate(), ""},
		{"getChainId nil", (*GetChainIDRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getChainId empty", (&GetChainIDRequest{}).Validate(), "invalid request: contract is nil"},
		{"getChainId valid", (&GetChainIDRequest{Contract: contract}).Validate(), ""},
		{"getCurrentBlockCoinbase nil", (*GetCurrentBlockCoinbaseRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getCurrentBlockCoinbase empty", (&GetCurrentBlockCoinbaseRequest{}).Validate(), "invalid request: contract is nil"},
		{"getCurrentBlockCoinbase valid", (&GetCurrentBlockCoinbaseRequest{Contract: contract}).Validate(), ""},
		{"getCurrentBlockDifficulty nil", (*GetCurrentBlockDifficultyRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getCurrentBlockDifficulty empty", (&GetCurrentBlockDifficultyRequest{}).Validate(), "invalid request: contract is nil"},
		{"getCurrentBlockDifficulty valid", (&GetCurrentBlockDifficultyRequest{Contract: contract}).Validate(), ""},
		{"getCurrentBlockGasLimit nil", (*GetCurrentBlockGasLimitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getCurrentBlockGasLimit empty", (&GetCurrentBlockGasLimitRequest{}).Validate(), "invalid request: contract is nil"},
		{"getCurrentBlockGasLimit valid", (&GetCurrentBlockGasLimitRequest{Contract: contract}).Validate(), ""},
		{"getCurrentBlockTimestamp nil", (*GetCurrentBlockTimestampRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getCurrentBlockTimestamp empty", (&GetCurrentBlockTimestampRequest{}).Validate(), "invalid request: contract is nil"},
		{"getCurrentBlockTimestamp valid", (&GetCurrentBlockTimestampRequest{Contract: contract}).Validate(), ""},
		{"getLastBlockHash nil", (*GetLastBlockHashRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getLastBlockHash empty", (&GetLastBlockHashRequest{}).Validate(), "invalid request: contract is nil"},
		{"getLastBlockHash valid", (&GetLastBlockHashRequest{Contract: contract}).Validate(), ""},
		{"getBlockHash nil", (*GetBlockHashRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getBlockHash empty", (&GetBlockHashRequest{}).Validate(), "invalid request: contract is nil"},
		{"getBlockHash nil contract", (&GetBlockHashRequest{BlockNumber: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"getBlockHash nil block number", (&GetBlockHashRequest{Contract: contract}).Validate(), "invalid request: block number is nil"},
		{"getBlockHash valid", (&GetBlockHashRequest{Contract: contract, BlockNumber: big.NewInt(1)}).Validate(), ""},
		{"getBlockHash zero block number", (&GetBlockHashRequest{Contract: contract, BlockNumber: big.NewInt(0)}).Validate(), ""},
		{"getEthBalance nil", (*GetEthBalanceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getEthBalance empty", (&GetEthBalanceRequest{}).Validate(), "invalid request: contract is nil"},
		{"getEthBalance nil contract", (&GetEthBalanceRequest{Addr: a}).Validate(), "invalid request: contract is nil"},
		{"getEthBalance nil addr", (&GetEthBalanceRequest{Contract: contract}).Validate(), "invalid request: addr is nil"},
		{"getEthBalance valid", (&GetEthBalanceRequest{Contract: contract, Addr: a}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewRequestsSetFields(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	calls := []Call{NewCall(a, []byte{1})}
	calls3 := []Call3{NewCall3(a, true, []byte{1})}
	calls3Value := []Call3Value{NewCall3Value(a, true, big.NewInt(5), []byte{1})}

	require.Equal(t, &AggregateRequest{Contract: contract, Calls: calls}, NewAggregateRequest(contract, calls))
	require.Equal(t, &TryAggregateRequest{Contract: contract, RequireSuccess: true, Calls: calls}, NewTryAggregateRequest(contract, true, calls))
	require.Equal(t, &BlockAndAggregateRequest{Contract: contract, Calls: calls}, NewBlockAndAggregateRequest(contract, calls))
	require.Equal(t, &TryBlockAndAggregateRequest{Contract: contract, RequireSuccess: true, Calls: calls}, NewTryBlockAndAggregateRequest(contract, true, calls))
	require.Equal(t, &Aggregate3Request{Contract: contract, Calls: calls3}, NewAggregate3Request(contract, calls3))
	require.Equal(t, &Aggregate3ValueRequest{Contract: contract, Calls: calls3Value}, NewAggregate3ValueRequest(contract, calls3Value))
	require.Equal(t, &GetBasefeeRequest{Contract: contract}, NewGetBasefeeRequest(contract))
	require.Equal(t, &GetBlockHashRequest{Contract: contract, BlockNumber: big.NewInt(9)}, NewGetBlockHashRequest(contract, big.NewInt(9)))
	require.Equal(t, &GetBlockNumberRequest{Contract: contract}, NewGetBlockNumberRequest(contract))
	require.Equal(t, &GetChainIDRequest{Contract: contract}, NewGetChainIDRequest(contract))
	require.Equal(t, &GetCurrentBlockCoinbaseRequest{Contract: contract}, NewGetCurrentBlockCoinbaseRequest(contract))
	require.Equal(t, &GetCurrentBlockDifficultyRequest{Contract: contract}, NewGetCurrentBlockDifficultyRequest(contract))
	require.Equal(t, &GetCurrentBlockGasLimitRequest{Contract: contract}, NewGetCurrentBlockGasLimitRequest(contract))
	require.Equal(t, &GetCurrentBlockTimestampRequest{Contract: contract}, NewGetCurrentBlockTimestampRequest(contract))
	require.Equal(t, &GetEthBalanceRequest{Contract: contract, Addr: a}, NewGetEthBalanceRequest(contract, a))
	require.Equal(t, &GetLastBlockHashRequest{Contract: contract}, NewGetLastBlockHashRequest(contract))
}

func TestNewCallsSetFields(t *testing.T) {
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	require.Equal(t, Call{Target: a, CallData: []byte{1}}, NewCall(a, []byte{1}))
	require.Equal(t, Call3{Target: a, CallData: []byte{1}, AllowFailure: true}, NewCall3(a, true, []byte{1}))
	require.Equal(t, Call3Value{Target: a, CallData: []byte{1}, AllowFailure: true, Value: big.NewInt(5)}, NewCall3Value(a, true, big.NewInt(5), []byte{1}))
	require.Equal(t, Result{Success: true, ReturnData: []byte{2}}, NewResult(true, []byte{2}))
}

func TestAggregate3ValueRequestTotalValue(t *testing.T) {
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	huge, ok := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	require.True(t, ok)

	for _, tc := range []struct {
		name  string
		calls []Call3Value
		want  string
	}{
		{"no calls", nil, "0"},
		{"one call", []Call3Value{NewCall3Value(a, true, big.NewInt(7), nil)}, "7"},
		{"many calls", []Call3Value{NewCall3Value(a, true, big.NewInt(1), nil), NewCall3Value(a, false, big.NewInt(0), nil), NewCall3Value(a, true, big.NewInt(1_000_000_000_000_000_000), nil)}, "1000000000000000001"},
		{"nil value skipped", []Call3Value{NewCall3Value(a, true, nil, nil), NewCall3Value(a, true, big.NewInt(3), nil)}, "3"},
		{"beyond uint256", []Call3Value{NewCall3Value(a, true, huge, nil), NewCall3Value(a, true, big.NewInt(1), nil)}, "115792089237316195423570985008687907853269984665640564039457584007913129639936"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, NewAggregate3ValueRequest(a, tc.calls).TotalValue().String())
		})
	}

	first := big.NewInt(7)
	req := NewAggregate3ValueRequest(a, []Call3Value{NewCall3Value(a, true, first, nil), NewCall3Value(a, true, big.NewInt(3), nil)})
	require.Equal(t, "10", req.TotalValue().String())
	require.Equal(t, "10", req.TotalValue().String())
	require.Equal(t, big.NewInt(7), first)
}
