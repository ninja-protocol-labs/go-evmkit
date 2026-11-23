package universalrouter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	testTokenA = types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	testTokenB = types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	testTokenC = types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
)

func TestCommandValuesMatchCommandsSol(t *testing.T) {
	for _, tc := range []struct {
		cmd  Command
		want byte
		name string
	}{
		{CommandV3SwapExactIn, 0x00, "V3_SWAP_EXACT_IN"},
		{CommandV3SwapExactOut, 0x01, "V3_SWAP_EXACT_OUT"},
		{CommandPermit2TransferFrom, 0x02, "PERMIT2_TRANSFER_FROM"},
		{CommandPermit2PermitBatch, 0x03, "PERMIT2_PERMIT_BATCH"},
		{CommandSweep, 0x04, "SWEEP"},
		{CommandTransfer, 0x05, "TRANSFER"},
		{CommandPayPortion, 0x06, "PAY_PORTION"},
		{CommandPayPortionFullPrecision, 0x07, "PAY_PORTION_FULL_PRECISION"},
		{CommandV2SwapExactIn, 0x08, "V2_SWAP_EXACT_IN"},
		{CommandV2SwapExactOut, 0x09, "V2_SWAP_EXACT_OUT"},
		{CommandPermit2Permit, 0x0a, "PERMIT2_PERMIT"},
		{CommandWrapETH, 0x0b, "WRAP_ETH"},
		{CommandUnwrapWETH, 0x0c, "UNWRAP_WETH"},
		{CommandPermit2TransferFromBatch, 0x0d, "PERMIT2_TRANSFER_FROM_BATCH"},
		{CommandBalanceCheckERC20, 0x0e, "BALANCE_CHECK_ERC20"},
		{CommandV4Swap, 0x10, "V4_SWAP"},
		{CommandV3PositionManagerPermit, 0x11, "V3_POSITION_MANAGER_PERMIT"},
		{CommandV3PositionManagerCall, 0x12, "V3_POSITION_MANAGER_CALL"},
		{CommandV4InitializePool, 0x13, "V4_INITIALIZE_POOL"},
		{CommandV4PositionManagerCall, 0x14, "V4_POSITION_MANAGER_CALL"},
		{CommandExecuteSubPlan, 0x21, "EXECUTE_SUB_PLAN"},
		{CommandAcrossV4DepositV3, 0x40, "ACROSS_V4_DEPOSIT_V3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, byte(tc.cmd))
			require.Equal(t, tc.name, tc.cmd.String())
			require.True(t, tc.cmd.Valid())
			require.Equal(t, tc.want, tc.cmd.Byte(false))
			require.Equal(t, tc.want|0x80, tc.cmd.Byte(true))

			cmd, allowRevert := ParseCommandByte(tc.cmd.Byte(true))
			require.Equal(t, tc.cmd, cmd)
			require.True(t, allowRevert)

			cmd, allowRevert = ParseCommandByte(tc.cmd.Byte(false))
			require.Equal(t, tc.cmd, cmd)
			require.False(t, allowRevert)
		})
	}
}

func TestCommandPlaceholdersAreInvalid(t *testing.T) {
	var invalid []byte
	for b := 0; b <= 0x7f; b++ {
		if !Command(b).Valid() {
			invalid = append(invalid, byte(b))
		}
	}

	require.Len(t, invalid, 128-22)
	require.Contains(t, invalid, byte(0x0f))
	require.Contains(t, invalid, byte(0x15))
	require.Contains(t, invalid, byte(0x20))
	require.Contains(t, invalid, byte(0x22))
	require.Contains(t, invalid, byte(0x3f))
	require.Contains(t, invalid, byte(0x41))
	require.Contains(t, invalid, byte(0x7f))
	require.Equal(t, "0x0f", Command(0x0f).String())
	require.Equal(t, "0x7f", Command(0x7f).String())
}

func TestParseCommandByteMasksOnlyTheRevertFlag(t *testing.T) {
	cmd, allowRevert := ParseCommandByte(0xc0)
	require.Equal(t, CommandAcrossV4DepositV3, cmd)
	require.True(t, allowRevert)

	cmd, allowRevert = ParseCommandByte(0xa1)
	require.Equal(t, CommandExecuteSubPlan, cmd)
	require.True(t, allowRevert)

	cmd, allowRevert = ParseCommandByte(0xff)
	require.Equal(t, Command(0x7f), cmd)
	require.False(t, cmd.Valid())
	require.True(t, allowRevert)

	require.Equal(t, byte(0x80), FlagAllowRevert)
	require.Equal(t, byte(0x7f), CommandTypeMask)
}
