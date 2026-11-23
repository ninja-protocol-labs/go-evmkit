package universalrouter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestExecuteSelectorsMatchUniversalRouter(t *testing.T) {
	for _, tc := range []struct {
		fn  *abi.Function
		sig string
		sel string
	}{
		{executeFn, "execute(bytes,bytes[])", "0x24856bc3"},
		{executeWithDeadlineFn, "execute(bytes,bytes[],uint256)", "0x3593564c"},
		{executeSignedFn, "executeSigned(bytes,bytes[],bytes32,bytes32,bool,bytes32,bytes,uint256)", "0x44684dc3"},
		{v4ModifyLiquiditiesFn, "modifyLiquidities(bytes,uint256)", "0xdd46508f"},
	} {
		require.Equal(t, tc.sig, tc.fn.Signature())
		sel := tc.fn.Selector()
		require.Equal(t, tc.sel, encoding.Hex.EncodePrefixed(sel[:]))
	}

	require.Equal(t, "0x7ac2ff7b", encoding.Hex.EncodePrefixed(v3PermitSelector[:]))
	require.Equal(t, "0x0c49ccbe", encoding.Hex.EncodePrefixed(v3DecreaseLiquiditySelector[:]))
	require.Equal(t, "0xfc6f7865", encoding.Hex.EncodePrefixed(v3CollectSelector[:]))
	require.Equal(t, "0x42966c68", encoding.Hex.EncodePrefixed(v3BurnSelector[:]))
}

func TestEncodeExecuteMatchesCastGolden(t *testing.T) {
	commands := []byte{0x0b, 0x04}
	inputs := [][]byte{{0x01}, {0x02}}

	data, err := EncodeExecute(commands, inputs)
	require.NoError(t, err)
	require.Equal(t, "0x24856bc30000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000020b040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010200000000000000000000000000000000000000000000000000000000000000", encoding.Hex.EncodePrefixed(data))

	data, err = EncodeExecuteWithDeadline(commands, inputs, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	require.Equal(t, "0x3593564c000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000006553f10000000000000000000000000000000000000000000000000000000000000000020b040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010200000000000000000000000000000000000000000000000000000000000000", encoding.Hex.EncodePrefixed(data))

	var intent, word, nonce [32]byte
	for i := range 32 {
		intent[i], word[i], nonce[i] = 0xaa, 0xbb, 0xcc
	}
	data, err = EncodeExecuteSigned([]byte{0x0b}, [][]byte{{0x01}}, intent, word, true, nonce, []byte{0x12, 0x34}, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	require.Equal(t, "0x44684dc300000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000140aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0000000000000000000000000000000000000000000000000000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc00000000000000000000000000000000000000000000000000000000000001c0000000000000000000000000000000000000000000000000000000006553f10000000000000000000000000000000000000000000000000000000000000000010b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000021234000000000000000000000000000000000000000000000000000000000000", encoding.Hex.EncodePrefixed(data))
}

func TestEncodeExecuteHandlesEmptyAndNilInputs(t *testing.T) {
	withNil, err := EncodeExecute([]byte{0x04}, [][]byte{nil})
	require.NoError(t, err)
	withEmpty, err := EncodeExecute([]byte{0x04}, [][]byte{{}})
	require.NoError(t, err)
	require.Equal(t, withEmpty, withNil)

	none, err := EncodeExecute(nil, nil)
	require.NoError(t, err)
	vals, err := executeFn.DecodeCall(none)
	require.NoError(t, err)
	require.Empty(t, vals[0])
	require.Empty(t, vals[1])

	_, err = EncodeExecuteWithDeadline([]byte{0x04}, [][]byte{{0x01}}, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "universalrouter: encodeExecuteWithDeadline: ")

	_, err = EncodeExecuteWithDeadline([]byte{0x04}, [][]byte{{0x01}}, big.NewInt(-1))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, err = EncodeExecuteSigned([]byte{0x04}, [][]byte{{0x01}}, [32]byte{}, [32]byte{}, false, [32]byte{}, nil, new(big.Int).Lsh(big.NewInt(1), 256))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "universalrouter: encodeExecuteSigned: ")
}

func TestDecodeExecuteCallRoundTripsEverySelector(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())).AddAllowRevert(NewSweepInput(testTokenA, MsgSender(), big.NewInt(1)))
	commands, inputs, err := plan.Encode()
	require.NoError(t, err)

	plain, err := EncodeExecute(commands, inputs)
	require.NoError(t, err)
	call, err := DecodeExecuteCall(plain)
	require.NoError(t, err)
	require.Nil(t, call.Deadline)
	require.Nil(t, call.Signed)
	require.Len(t, call.Plan.Steps, 2)
	require.Equal(t, CommandWrapETH, call.Plan.Steps[0].Input.Command())
	require.True(t, call.Plan.Steps[1].AllowRevert)

	withDeadline, err := EncodeExecuteWithDeadline(commands, inputs, big.NewInt(1_800_000_000))
	require.NoError(t, err)
	call, err = DecodeExecuteCall(withDeadline)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(1_800_000_000).Cmp(call.Deadline))
	require.Nil(t, call.Signed)
	againCommands, againInputs, err := call.Plan.Encode()
	require.NoError(t, err)
	require.Equal(t, commands, againCommands)
	require.Equal(t, inputs, againInputs)

	intent := [32]byte{0x01, 0x02}
	data := [32]byte{0x03}
	nonce := [32]byte{31: 0x09}
	signed, err := EncodeExecuteSigned(commands, inputs, intent, data, false, nonce, []byte{0xde, 0xad}, big.NewInt(7))
	require.NoError(t, err)
	call, err = DecodeExecuteCall(signed)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(7).Cmp(call.Deadline))
	require.NotNil(t, call.Signed)
	require.Equal(t, intent, call.Signed.Intent)
	require.Equal(t, data, call.Signed.Data)
	require.Equal(t, nonce, call.Signed.Nonce)
	require.False(t, call.Signed.VerifySender)
	require.Equal(t, []byte{0xde, 0xad}, call.Signed.Signature)
	require.Len(t, call.Plan.Steps, 2)
}

func TestDecodeExecuteCallRejectsBadCalldata(t *testing.T) {
	_, err := DecodeExecuteCall(nil)
	require.EqualError(t, err, "universalrouter: decodeExecuteCall: calldata is 0 bytes, shorter than a selector")

	_, err = DecodeExecuteCall([]byte{0x24, 0x85, 0x6b})
	require.ErrorContains(t, err, "shorter than a selector")

	_, err = DecodeExecuteCall([]byte{0xde, 0xad, 0xbe, 0xef, 0x00})
	require.ErrorIs(t, err, ErrUnknownSelector)
	require.EqualError(t, err, "universalrouter: decodeExecuteCall: unknown execute selector: deadbeef")

	for _, sel := range [][]byte{{0x24, 0x85, 0x6b, 0xc3}, {0x35, 0x93, 0x56, 0x4c}, {0x44, 0x68, 0x4d, 0xc3}} {
		_, err = DecodeExecuteCall(append(sel, 0x01))
		require.ErrorContains(t, err, "universalrouter: decodeExecuteCall: ")
		require.NotErrorIs(t, err, ErrUnknownSelector)
	}

	unknownCommand, err := EncodeExecute([]byte{0x0f}, [][]byte{{}})
	require.NoError(t, err)
	_, err = DecodeExecuteCall(unknownCommand)
	require.ErrorIs(t, err, ErrUnknownCommand)

	mismatch, err := EncodeExecute([]byte{0x04, 0x04}, [][]byte{{}})
	require.NoError(t, err)
	_, err = DecodeExecuteCall(mismatch)
	require.ErrorContains(t, err, "2 commands but 1 inputs")
}

func TestInputLayoutsMatchDispatcherComments(t *testing.T) {
	for _, tc := range []struct {
		types abi.Types
		want  string
	}{
		{v3SwapTypes, "(address,uint256,uint256,bytes,bool,uint256[])"},
		{v2SwapTypes, "(address,uint256,uint256,address[],bool,uint256[])"},
		{tokenRecipientAmountTypes, "(address,address,uint256)"},
		{permit2TransferFromTypes, "(address,address,uint160)"},
		{permit2PermitTypes, "(((address,uint160,uint48,uint48),address,uint256),bytes)"},
		{permit2PermitBatchTypes, "(((address,uint160,uint48,uint48)[],address,uint256),bytes)"},
		{permit2TransferFromBatchTypes, "((address,address,uint160,address)[])"},
		{recipientAmountTypes, "(address,uint256)"},
		{ownerTokenAmountTypes, "(address,address,uint256)"},
		{v4InitializePoolTypes, "((address,address,uint24,int24,address),uint160)"},
		{subPlanTypes, "(bytes,bytes[])"},
		{acrossV4DepositV3Types, "((address,address,address,address,uint256,uint256,uint256,address,uint32,uint32,uint32,bytes,bool))"},
	} {
		require.Equal(t, tc.want, abi.NewFunction("", tc.types, nil).Signature())
	}
}
