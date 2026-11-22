package multicall3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

const (
	callAggregate            = "252dba4200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000002470a08231000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
	callBlockAndAggregate    = "c3077fa900000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000002470a08231000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
	callTryAggregate         = "bce38bd7000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000002470a08231000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
	callTryBlockAndAggregate = "399542e9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000011111111111111111111111111111111111111110000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000002470a08231000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
	callGetBlockHash         = "ee82ac5e0000000000000000000000000000000000000000000000000000000000003039"
	callGetEthBalance        = "4d2301cc0000000000000000000000001111111111111111111111111111111111111111"
	retAggregate             = "00000000000000000000000000000000000000000000000000000000000030390000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000000"
	retBlockAndAggregate     = "0000000000000000000000000000000000000000000000000000000000003039aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000016400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
	retUint256               = "0000000000000000000000000000000000000000000000000000000000003039"
	retBytes32               = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	retAddress               = "0000000000000000000000001111111111111111111111111111111111111111"

	aggregate3TwoCalls = "82ad56cb0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000100000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000002470a082310000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000000000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000"
	aggregate3Empty    = "82ad56cb00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000"
	aggregate3Value1   = "174dea71000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000004deadbeef00000000000000000000000000000000000000000000000000000000"
	resultsTwo         = "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000408c379a000000000000000000000000000000000000000000000000000000000"
	resultsEmpty       = "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000"
)

func TestEncodeAggregate3(t *testing.T) {
	balanceOf, err := encoding.Hex.Decode("70a082310000000000000000000000002222222222222222222222222222222222222222")
	require.NoError(t, err)
	want, err := encoding.Hex.Decode(aggregate3TwoCalls)
	require.NoError(t, err)

	calls := []Call3{
		{
			Target:       types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"),
			AllowFailure: true,
			CallData:     balanceOf,
		},
		{Target: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")},
	}

	got, err := EncodeAggregate3(calls)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeAggregate3Empty(t *testing.T) {
	want, err := encoding.Hex.Decode(aggregate3Empty)
	require.NoError(t, err)

	got, err := EncodeAggregate3(nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeAggregate3NilTarget(t *testing.T) {
	_, err := EncodeAggregate3([]Call3{{}})
	require.ErrorIs(t, err, abi.ErrNilAddress)
}

func TestEncodeAggregate3Value(t *testing.T) {
	callData, err := encoding.Hex.Decode("deadbeef")
	require.NoError(t, err)
	want, err := encoding.Hex.Decode(aggregate3Value1)
	require.NoError(t, err)

	calls := []Call3Value{{
		Target:       types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"),
		AllowFailure: true,
		Value:        big.NewInt(1_000_000_000_000_000_000),
		CallData:     callData,
	}}

	got, err := EncodeAggregate3Value(calls)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeAggregate3ValueNilValue(t *testing.T) {
	_, err := EncodeAggregate3Value([]Call3Value{{Target: types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")}})
	require.ErrorIs(t, err, abi.ErrInvalidGoType)
}

func TestDecodeResults(t *testing.T) {
	data, err := encoding.Hex.Decode(resultsTwo)
	require.NoError(t, err)
	word, err := encoding.Hex.Decode("0000000000000000000000000000000000000000000000000000000000000064")
	require.NoError(t, err)

	got, err := DecodeResults(data)
	require.NoError(t, err)
	require.Equal(t, []Result{
		{Success: true, ReturnData: word},
		{Success: false, ReturnData: []byte{0x08, 0xc3, 0x79, 0xa0}},
	}, got)
}

func TestDecodeResultsEmpty(t *testing.T) {
	data, err := encoding.Hex.Decode(resultsEmpty)
	require.NoError(t, err)

	got, err := DecodeResults(data)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestDecodeResultsTruncated(t *testing.T) {
	full, err := encoding.Hex.Decode(resultsTwo)
	require.NoError(t, err)

	const unpaddedEnd = 356
	for n := range unpaddedEnd {
		_, err := DecodeResults(full[:n])
		require.Error(t, err, "prefix length %d", n)
	}
}

func TestDecodeResultsRejectsHugeOffsets(t *testing.T) {
	data, err := encoding.Hex.Decode(resultsTwo)
	require.NoError(t, err)

	for i := 32; i < 64; i++ {
		data[i] = 0xff
	}
	_, err = DecodeResults(data)
	require.Error(t, err)
}

func TestAggregate3RoundTripThroughDecodeCall(t *testing.T) {
	calls := []Call3{
		{Target: types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"), AllowFailure: true, CallData: []byte{1, 2, 3}},
		{Target: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222"), CallData: nil},
	}

	data, err := EncodeAggregate3(calls)
	require.NoError(t, err)

	vals, err := aggregate3.DecodeCall(data)
	require.NoError(t, err)

	elems := vals[0].([]any)
	require.Len(t, elems, len(calls))
	for i, e := range elems {
		fields := e.([]any)
		require.True(t, calls[i].Target.Equal(fields[0].(*types.Address)))
		require.Equal(t, calls[i].AllowFailure, fields[1])
		require.Equal(t, append([]byte{}, calls[i].CallData...), fields[2])
	}
}

func TestEncodeCallsFunctions(t *testing.T) {
	balanceOf, err := encoding.Hex.Decode("70a082310000000000000000000000002222222222222222222222222222222222222222")
	require.NoError(t, err)
	calls := []Call{
		{Target: types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"), CallData: balanceOf},
		{Target: types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")},
	}

	aggregateData, err := EncodeAggregate(calls)
	require.NoError(t, err)
	blockAndAggregateData, err := EncodeBlockAndAggregate(calls)
	require.NoError(t, err)
	tryAggregateData, err := EncodeTryAggregate(true, calls)
	require.NoError(t, err)
	tryBlockAndAggregateData, err := EncodeTryBlockAndAggregate(false, calls)
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		got  []byte
		want string
	}{
		{"aggregate", aggregateData, callAggregate},
		{"blockAndAggregate", blockAndAggregateData, callBlockAndAggregate},
		{"tryAggregate", tryAggregateData, callTryAggregate},
		{"tryBlockAndAggregate", tryBlockAndAggregateData, callTryBlockAndAggregate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := encoding.Hex.Decode(tc.want)
			require.NoError(t, err)
			require.Equal(t, want, tc.got)
		})
	}
}

func TestEncodeCallsNilTarget(t *testing.T) {
	_, err := EncodeAggregate([]Call{{}})
	require.ErrorIs(t, err, abi.ErrNilAddress)
	_, err = EncodeTryAggregate(true, []Call{{}})
	require.ErrorIs(t, err, abi.ErrNilAddress)
	_, err = EncodeBlockAndAggregate([]Call{{}})
	require.ErrorIs(t, err, abi.ErrNilAddress)
	_, err = EncodeTryBlockAndAggregate(false, []Call{{}})
	require.ErrorIs(t, err, abi.ErrNilAddress)
}

func TestDecodeAggregate(t *testing.T) {
	data, err := encoding.Hex.Decode(retAggregate)
	require.NoError(t, err)
	word, err := encoding.Hex.Decode("0000000000000000000000000000000000000000000000000000000000000064")
	require.NoError(t, err)

	got, err := DecodeAggregate(data)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(12345), got.BlockNumber)
	require.Equal(t, [][]byte{word, {}}, got.ReturnData)
}

func TestDecodeBlockAndAggregate(t *testing.T) {
	data, err := encoding.Hex.Decode(retBlockAndAggregate)
	require.NoError(t, err)
	hash, err := encoding.Hex.Decode("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	got, err := DecodeBlockAndAggregate(data)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(12345), got.BlockNumber)
	require.Equal(t, hash, got.BlockHash.Bytes())
	require.Equal(t, []Result{{Success: true, ReturnData: []byte{0x64}}, {Success: false, ReturnData: []byte{}}}, got.Results)
}

func TestDecodeAggregateTruncated(t *testing.T) {
	full, err := encoding.Hex.Decode(retAggregate)
	require.NoError(t, err)

	_, err = DecodeAggregate(full[:len(full)/2])
	require.Error(t, err)
}

func TestDecodeBlockAndAggregateTruncated(t *testing.T) {
	full, err := encoding.Hex.Decode(retBlockAndAggregate)
	require.NoError(t, err)

	_, err = DecodeBlockAndAggregate(full[:len(full)/2])
	require.Error(t, err)
}

func TestEncodeGettersWithoutArguments(t *testing.T) {
	for _, tc := range []struct {
		name     string
		got      []byte
		selector string
	}{
		{"getBasefee", EncodeGetBasefee(), "3e64a696"},
		{"getBlockNumber", EncodeGetBlockNumber(), "42cbb15c"},
		{"getChainId", EncodeGetChainID(), "3408e470"},
		{"getCurrentBlockCoinbase", EncodeGetCurrentBlockCoinbase(), "a8b0574e"},
		{"getCurrentBlockDifficulty", EncodeGetCurrentBlockDifficulty(), "72425d9d"},
		{"getCurrentBlockGasLimit", EncodeGetCurrentBlockGasLimit(), "86d516e8"},
		{"getCurrentBlockTimestamp", EncodeGetCurrentBlockTimestamp(), "0f28c97d"},
		{"getLastBlockHash", EncodeGetLastBlockHash(), "27e86d6e"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.selector, encoding.Hex.Encode(tc.got))
		})
	}
}

func TestEncodeGetBlockHash(t *testing.T) {
	want, err := encoding.Hex.Decode(callGetBlockHash)
	require.NoError(t, err)

	got, err := EncodeGetBlockHash(big.NewInt(12345))
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeGetBlockHashNil(t *testing.T) {
	_, err := EncodeGetBlockHash(nil)
	require.ErrorIs(t, err, abi.ErrInvalidGoType)
}

func TestEncodeGetEthBalance(t *testing.T) {
	want, err := encoding.Hex.Decode(callGetEthBalance)
	require.NoError(t, err)

	got, err := EncodeGetEthBalance(types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"))
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestEncodeGetEthBalanceNil(t *testing.T) {
	_, err := EncodeGetEthBalance(nil)
	require.ErrorIs(t, err, abi.ErrNilAddress)
}

func TestDecodeUint256Getters(t *testing.T) {
	data, err := encoding.Hex.Decode(retUint256)
	require.NoError(t, err)

	for name, decode := range map[string]func([]byte) (*big.Int, error){
		"getBasefee":                DecodeGetBasefee,
		"getBlockNumber":            DecodeGetBlockNumber,
		"getChainId":                DecodeGetChainID,
		"getCurrentBlockDifficulty": DecodeGetCurrentBlockDifficulty,
		"getCurrentBlockGasLimit":   DecodeGetCurrentBlockGasLimit,
		"getCurrentBlockTimestamp":  DecodeGetCurrentBlockTimestamp,
		"getEthBalance":             DecodeGetEthBalance,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := decode(data)
			require.NoError(t, err)
			require.Equal(t, big.NewInt(12345), got)

			_, err = decode(data[:31])
			require.Error(t, err)
		})
	}
}

func TestDecodeHashGetters(t *testing.T) {
	data, err := encoding.Hex.Decode(retBytes32)
	require.NoError(t, err)
	want, err := encoding.Hex.Decode("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	blockHash, err := DecodeGetBlockHash(data)
	require.NoError(t, err)
	require.Equal(t, want, blockHash.Bytes())

	lastHash, err := DecodeGetLastBlockHash(data)
	require.NoError(t, err)
	require.Equal(t, want, lastHash.Bytes())

	_, err = DecodeGetBlockHash(data[:31])
	require.Error(t, err)
	_, err = DecodeGetLastBlockHash(data[:31])
	require.Error(t, err)
}

func TestDecodeGetCurrentBlockCoinbase(t *testing.T) {
	data, err := encoding.Hex.Decode(retAddress)
	require.NoError(t, err)

	got, err := DecodeGetCurrentBlockCoinbase(data)
	require.NoError(t, err)
	require.True(t, got.Equal(types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")))

	_, err = DecodeGetCurrentBlockCoinbase(data[:31])
	require.Error(t, err)
}
