package universalrouter

import "fmt"

// Command is one Universal Router command type, the low 7 bits of a commands byte (Commands.sol).
type Command byte

const (
	CommandV3SwapExactIn            Command = 0x00
	CommandV3SwapExactOut           Command = 0x01
	CommandPermit2TransferFrom      Command = 0x02
	CommandPermit2PermitBatch       Command = 0x03
	CommandSweep                    Command = 0x04
	CommandTransfer                 Command = 0x05
	CommandPayPortion               Command = 0x06
	CommandPayPortionFullPrecision  Command = 0x07
	CommandV2SwapExactIn            Command = 0x08
	CommandV2SwapExactOut           Command = 0x09
	CommandPermit2Permit            Command = 0x0a
	CommandWrapETH                  Command = 0x0b
	CommandUnwrapWETH               Command = 0x0c
	CommandPermit2TransferFromBatch Command = 0x0d
	CommandBalanceCheckERC20        Command = 0x0e
	CommandV4Swap                   Command = 0x10
	CommandV3PositionManagerPermit  Command = 0x11
	CommandV3PositionManagerCall    Command = 0x12
	CommandV4InitializePool         Command = 0x13
	CommandV4PositionManagerCall    Command = 0x14
	CommandExecuteSubPlan           Command = 0x21
	CommandAcrossV4DepositV3        Command = 0x40
)

const (
	FlagAllowRevert byte = 0x80
	CommandTypeMask byte = 0x7f
)

const (
	BipsDenominator                 = 10_000
	FullPrecisionPortionDenominator = 1_000_000_000_000_000_000
)

// Valid reports whether c is a command the 2.1 dispatcher accepts; any other type reverts with InvalidCommandType.
func (c Command) Valid() bool {
	switch c {
	case CommandV3SwapExactIn,
		CommandV3SwapExactOut,
		CommandPermit2TransferFrom,
		CommandPermit2PermitBatch,
		CommandSweep,
		CommandTransfer,
		CommandPayPortion,
		CommandPayPortionFullPrecision,
		CommandV2SwapExactIn,
		CommandV2SwapExactOut,
		CommandPermit2Permit,
		CommandWrapETH,
		CommandUnwrapWETH,
		CommandPermit2TransferFromBatch,
		CommandBalanceCheckERC20,
		CommandV4Swap,
		CommandV3PositionManagerPermit,
		CommandV3PositionManagerCall,
		CommandV4InitializePool,
		CommandV4PositionManagerCall,
		CommandExecuteSubPlan,
		CommandAcrossV4DepositV3:
		return true
	default:
		return false
	}
}

// String returns c's Commands.sol name, or its hex value if unknown.
func (c Command) String() string {
	switch c {
	case CommandV3SwapExactIn:
		return "V3_SWAP_EXACT_IN"
	case CommandV3SwapExactOut:
		return "V3_SWAP_EXACT_OUT"
	case CommandPermit2TransferFrom:
		return "PERMIT2_TRANSFER_FROM"
	case CommandPermit2PermitBatch:
		return "PERMIT2_PERMIT_BATCH"
	case CommandSweep:
		return "SWEEP"
	case CommandTransfer:
		return "TRANSFER"
	case CommandPayPortion:
		return "PAY_PORTION"
	case CommandPayPortionFullPrecision:
		return "PAY_PORTION_FULL_PRECISION"
	case CommandV2SwapExactIn:
		return "V2_SWAP_EXACT_IN"
	case CommandV2SwapExactOut:
		return "V2_SWAP_EXACT_OUT"
	case CommandPermit2Permit:
		return "PERMIT2_PERMIT"
	case CommandWrapETH:
		return "WRAP_ETH"
	case CommandUnwrapWETH:
		return "UNWRAP_WETH"
	case CommandPermit2TransferFromBatch:
		return "PERMIT2_TRANSFER_FROM_BATCH"
	case CommandBalanceCheckERC20:
		return "BALANCE_CHECK_ERC20"
	case CommandV4Swap:
		return "V4_SWAP"
	case CommandV3PositionManagerPermit:
		return "V3_POSITION_MANAGER_PERMIT"
	case CommandV3PositionManagerCall:
		return "V3_POSITION_MANAGER_CALL"
	case CommandV4InitializePool:
		return "V4_INITIALIZE_POOL"
	case CommandV4PositionManagerCall:
		return "V4_POSITION_MANAGER_CALL"
	case CommandExecuteSubPlan:
		return "EXECUTE_SUB_PLAN"
	case CommandAcrossV4DepositV3:
		return "ACROSS_V4_DEPOSIT_V3"
	default:
		return fmt.Sprintf("0x%02x", byte(c))
	}
}

// Byte returns c's commands byte, with FlagAllowRevert set when allowRevert is true.
func (c Command) Byte(allowRevert bool) byte {
	if allowRevert {
		return byte(c) | FlagAllowRevert
	}
	return byte(c)
}

// ParseCommandByte splits a commands byte into its command type and allow-revert flag.
func ParseCommandByte(b byte) (Command, bool) {
	return Command(b & CommandTypeMask), b&FlagAllowRevert != 0
}
