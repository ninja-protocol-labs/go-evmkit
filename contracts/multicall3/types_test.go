package multicall3

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
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
