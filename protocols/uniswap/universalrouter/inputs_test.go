package universalrouter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestInputsEncodeMatchesCastGolden(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB, testTokenC}, []*big.Int{big.NewInt(3000), big.NewInt(500)})
	require.NoError(t, err)
	require.Equal(t, "0x1111111111111111111111111111111111111111000bb822222222222222222222222222222222222222220001f43333333333333333333333333333333333333333", encoding.Hex.EncodePrefixed(path))

	v3Golden := "0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000f424000000000000000000000000000000000000000000000000000000000000f1b3000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000421111111111111111111111111111111111111111000bb822222222222222222222222222222222222222220001f43333333333333333333333333333333333333333000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000006"
	v2Golden := "0x000000000000000000000000333333333333333333333333333333333333333300000000000000000000000000000000000000000000000000000000000001f4000000000000000000000000000000000000000000000000000000000000025800000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000030000000000000000000000001111111111111111111111111111111111111111000000000000000000000000222222222222222222222222222222222222222200000000000000000000000033333333333333333333333333333333333333330000000000000000000000000000000000000000000000000000000000000000"
	tokenRecipientGolden := "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000022222222222222222222222222222222222222220000000000000000000000000000000000000000000000000000000000000007"
	recipientAmountGolden := "0x00000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000000"

	burnCalldata, err := v3.EncodeNfpmBurn(big.NewInt(1))
	require.NoError(t, err)
	v3PermitCalldata := append([]byte{0x7a, 0xc2, 0xff, 0x7b}, make([]byte, 6*32)...)
	mintData, err := v4.EncodeActionSettlePair(testTokenA, testTokenB)
	require.NoError(t, err)
	v4PMCalldata, err := v4.EncodeModifyLiquidities([]byte{v4.ActionMintPosition, v4.ActionSettlePair}, [][]byte{{0x01}, mintData}, big.NewInt(1_700_000_000))
	require.NoError(t, err)

	for _, tc := range []struct {
		name  string
		input Input
		want  string
	}{
		{"V3_SWAP_EXACT_IN", NewV3SwapExactInInput(MsgSender(), big.NewInt(1_000_000), big.NewInt(990_000), path, true, []*big.Int{big.NewInt(5), big.NewInt(6)}), v3Golden},
		{"V3_SWAP_EXACT_OUT", NewV3SwapExactOutInput(MsgSender(), big.NewInt(1_000_000), big.NewInt(990_000), path, true, []*big.Int{big.NewInt(5), big.NewInt(6)}), v3Golden},
		{"V2_SWAP_EXACT_IN", NewV2SwapExactInInput(testTokenC, big.NewInt(500), big.NewInt(600), []*types.Address{testTokenA, testTokenB, testTokenC}, false, nil), v2Golden},
		{"V2_SWAP_EXACT_OUT", NewV2SwapExactOutInput(testTokenC, big.NewInt(500), big.NewInt(600), []*types.Address{testTokenA, testTokenB, testTokenC}, false, []*big.Int{}), v2Golden},
		{"PERMIT2_TRANSFER_FROM", NewPermit2TransferFromInput(testTokenA, AddressThis(), big.NewInt(123_456_789)), "0x0000000000000000000000001111111111111111111111111111111111111111000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000075bcd15"},
		{"SWEEP", NewSweepInput(types.ZeroAddress(), testTokenB, big.NewInt(7)), tokenRecipientGolden},
		{"TRANSFER", NewTransferInput(types.ZeroAddress(), testTokenB, big.NewInt(7)), tokenRecipientGolden},
		{"PAY_PORTION", NewPayPortionInput(types.ZeroAddress(), testTokenB, big.NewInt(7)), tokenRecipientGolden},
		{"PAY_PORTION_FULL_PRECISION", NewPayPortionFullPrecisionInput(types.ZeroAddress(), testTokenB, big.NewInt(7)), tokenRecipientGolden},
		{"BALANCE_CHECK_ERC20", NewBalanceCheckERC20Input(types.ZeroAddress(), testTokenB, big.NewInt(7)), tokenRecipientGolden},
		{"PERMIT2_PERMIT", NewPermit2PermitInput(permit2.NewPermitSingle(permit2.NewPermitDetails(testTokenA, big.NewInt(1000), big.NewInt(1_700_000_000), big.NewInt(3)), testTokenB, big.NewInt(1_700_000_600)), []byte{0xaa, 0xbb, 0xcc}), "0x000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000003e8000000000000000000000000000000000000000000000000000000006553f10000000000000000000000000000000000000000000000000000000000000000030000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000006553f35800000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000003aabbcc0000000000000000000000000000000000000000000000000000000000"},
		{"PERMIT2_PERMIT_BATCH", NewPermit2PermitBatchInput(permit2.NewPermitBatch([]permit2.PermitDetails{*permit2.NewPermitDetails(testTokenA, big.NewInt(1000), big.NewInt(1_700_000_000), big.NewInt(3)), *permit2.NewPermitDetails(testTokenC, big.NewInt(2000), big.NewInt(1_700_000_001), big.NewInt(4))}, testTokenB, big.NewInt(1_700_000_600)), []byte{0xde, 0xad, 0xbe, 0xef}), "0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001c000000000000000000000000000000000000000000000000000000000000000600000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000006553f3580000000000000000000000000000000000000000000000000000000000000002000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000003e8000000000000000000000000000000000000000000000000000000006553f1000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000333333333333333333333333333333333333333300000000000000000000000000000000000000000000000000000000000007d0000000000000000000000000000000000000000000000000000000006553f10100000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000004deadbeef00000000000000000000000000000000000000000000000000000000"},
		{"PERMIT2_TRANSFER_FROM_BATCH", NewPermit2TransferFromBatchInput([]permit2.AllowanceTransferDetails{*permit2.NewAllowanceTransferDetails(testTokenA, testTokenB, big.NewInt(10), testTokenC), *permit2.NewAllowanceTransferDetails(testTokenA, testTokenC, big.NewInt(20), testTokenB)}), "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000033333333333333333333333333333333333333330000000000000000000000001111111111111111111111111111111111111111000000000000000000000000333333333333333333333333333333333333333300000000000000000000000000000000000000000000000000000000000000140000000000000000000000002222222222222222222222222222222222222222"},
		{"WRAP_ETH", NewWrapETHInput(AddressThis(), ContractBalance()), recipientAmountGolden},
		{"UNWRAP_WETH", NewUnwrapWETHInput(AddressThis(), ContractBalance()), recipientAmountGolden},
		{"V4_SWAP", NewV4SwapInput([]byte{v4.ActionSwapExactInSingle, v4.ActionSettleAll, v4.ActionTakeAll}, [][]byte{{0x01}, {0x02, 0x03}, {}}), "0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000003060c0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000010100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000202030000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
		{"V4_INITIALIZE_POOL", NewV4InitializePoolInput(v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(3000), big.NewInt(-60), types.ZeroAddress()), new(big.Int).Lsh(big.NewInt(1), 96)), "0x000000000000000000000000111111111111111111111111111111111111111100000000000000000000000022222222222222222222222222222222222222220000000000000000000000000000000000000000000000000000000000000bb8ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000"},
		{"EXECUTE_SUB_PLAN", NewExecuteSubPlanInput(NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())).AddAllowRevert(NewUnwrapWETHInput(MsgSender(), big.NewInt(5)))), "0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000020b8c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000005"},
		{"ACROSS_V4_DEPOSIT_V3", NewAcrossV4DepositV3Input(testTokenA, testTokenB, testTokenC, types.ZeroAddress(), big.NewInt(1000), big.NewInt(990), big.NewInt(10), types.ZeroAddress(), 1_700_000_000, 1_700_003_600, 0, []byte{0xca, 0xfe}, true), "0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000111111111111111111111111111111111111111100000000000000000000000022222222222222222222222222222222222222220000000000000000000000003333333333333333333333333333333333333333000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003e800000000000000000000000000000000000000000000000000000000000003de000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006553f100000000000000000000000000000000000000000000000000000000006553ff10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002cafe000000000000000000000000000000000000000000000000000000000000"},
		{"V3_POSITION_MANAGER_PERMIT", NewV3PositionManagerPermitInput(v3PermitCalldata), encoding.Hex.EncodePrefixed(v3PermitCalldata)},
		{"V3_POSITION_MANAGER_CALL", NewV3PositionManagerCallInput(burnCalldata), encoding.Hex.EncodePrefixed(burnCalldata)},
		{"V4_POSITION_MANAGER_CALL", NewV4PositionManagerCallInput(v4PMCalldata), encoding.Hex.EncodePrefixed(v4PMCalldata)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.name, tc.input.Command().String())

			got, err := tc.input.Encode()
			require.NoError(t, err)
			require.Equal(t, tc.want, encoding.Hex.EncodePrefixed(got))

			golden, err := encoding.Hex.Decode(tc.want)
			require.NoError(t, err)
			decoded, err := DecodeInput(tc.input.Command(), golden)
			require.NoError(t, err)
			require.Equal(t, tc.input.Command(), decoded.Command())
			require.NoError(t, decoded.Validate())

			again, err := decoded.Encode()
			require.NoError(t, err)
			require.Equal(t, tc.want, encoding.Hex.EncodePrefixed(again))
		})
	}
}

func TestDecodeInputRestoresFields(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB}, []*big.Int{big.NewInt(500)})
	require.NoError(t, err)

	v3Data, err := NewV3SwapExactOutInput(testTokenC, big.NewInt(77), big.NewInt(88), path, false, []*big.Int{big.NewInt(9)}).Encode()
	require.NoError(t, err)
	v3Out, err := DecodeV3SwapExactOutInput(v3Data)
	require.NoError(t, err)
	require.True(t, v3Out.Recipient.Equal(testTokenC))
	require.Zero(t, big.NewInt(77).Cmp(v3Out.AmountOut))
	require.Zero(t, big.NewInt(88).Cmp(v3Out.AmountInMax))
	require.Equal(t, path, v3Out.Path)
	require.False(t, v3Out.PayerIsUser)
	require.Len(t, v3Out.MinHopPriceX36, 1)
	require.Zero(t, big.NewInt(9).Cmp(v3Out.MinHopPriceX36[0]))

	v2Data, err := NewV2SwapExactInInput(MsgSender(), big.NewInt(1), big.NewInt(2), []*types.Address{testTokenB, testTokenA}, true, nil).Encode()
	require.NoError(t, err)
	v2In, err := DecodeV2SwapExactInInput(v2Data)
	require.NoError(t, err)
	require.True(t, v2In.Recipient.Equal(MsgSender()))
	require.Len(t, v2In.Path, 2)
	require.True(t, v2In.Path[0].Equal(testTokenB))
	require.True(t, v2In.Path[1].Equal(testTokenA))
	require.True(t, v2In.PayerIsUser)
	require.Empty(t, v2In.MinHopPriceX36)

	for _, cmd := range []Command{CommandSweep, CommandTransfer, CommandPayPortion, CommandPayPortionFullPrecision, CommandBalanceCheckERC20} {
		data, err := abi.Pack(tokenRecipientAmountTypes, testTokenA, testTokenB, big.NewInt(42))
		require.NoError(t, err)
		decoded, err := DecodeInput(cmd, data)
		require.NoError(t, err)
		require.Equal(t, cmd, decoded.Command())
	}

	wrapData, err := NewWrapETHInput(testTokenA, big.NewInt(5)).Encode()
	require.NoError(t, err)
	unwrap, err := DecodeUnwrapWETHInput(wrapData)
	require.NoError(t, err)
	require.True(t, unwrap.Recipient.Equal(testTokenA))
	require.Zero(t, big.NewInt(5).Cmp(unwrap.AmountMin))

	initData, err := NewV4InitializePoolInput(v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(500), big.NewInt(-887272), testTokenC), big.NewInt(4295128739)).Encode()
	require.NoError(t, err)
	initIn, err := DecodeV4InitializePoolInput(initData)
	require.NoError(t, err)
	require.True(t, initIn.Key.Currency0.Equal(testTokenA))
	require.True(t, initIn.Key.Currency1.Equal(testTokenB))
	require.Zero(t, big.NewInt(500).Cmp(initIn.Key.Fee))
	require.Zero(t, big.NewInt(-887272).Cmp(initIn.Key.TickSpacing))
	require.True(t, initIn.Key.Hooks.Equal(testTokenC))
	require.Zero(t, big.NewInt(4295128739).Cmp(initIn.SqrtPriceX96))

	acrossData, err := NewAcrossV4DepositV3Input(testTokenA, testTokenB, testTokenC, testTokenA, ContractBalance(), big.NewInt(1), big.NewInt(8453), testTokenB, 1, 2, 3, nil, false).Encode()
	require.NoError(t, err)
	across, err := DecodeAcrossV4DepositV3Input(acrossData)
	require.NoError(t, err)
	require.True(t, across.Depositor.Equal(testTokenA))
	require.True(t, across.ExclusiveRelayer.Equal(testTokenB))
	require.Zero(t, ContractBalance().Cmp(across.InputAmount))
	require.Zero(t, big.NewInt(8453).Cmp(across.DestinationChainID))
	require.Equal(t, uint32(1), across.QuoteTimestamp)
	require.Equal(t, uint32(2), across.FillDeadline)
	require.Equal(t, uint32(3), across.ExclusivityDeadline)
	require.Empty(t, across.Message)
	require.False(t, across.UseNative)

	permitData, err := NewPermit2PermitInput(permit2.NewPermitSingle(permit2.NewPermitDetails(testTokenC, big.NewInt(11), big.NewInt(12), big.NewInt(13)), testTokenA, big.NewInt(14)), []byte{0x01}).Encode()
	require.NoError(t, err)
	permitIn, err := DecodePermit2PermitInput(permitData)
	require.NoError(t, err)
	require.True(t, permitIn.PermitSingle.Details.Token.Equal(testTokenC))
	require.Zero(t, big.NewInt(11).Cmp(permitIn.PermitSingle.Details.Amount))
	require.Zero(t, big.NewInt(12).Cmp(permitIn.PermitSingle.Details.Expiration))
	require.Zero(t, big.NewInt(13).Cmp(permitIn.PermitSingle.Details.Nonce))
	require.True(t, permitIn.PermitSingle.Spender.Equal(testTokenA))
	require.Zero(t, big.NewInt(14).Cmp(permitIn.PermitSingle.SigDeadline))
	require.Equal(t, []byte{0x01}, permitIn.Signature)

	subData, err := NewExecuteSubPlanInput(NewPlan().AddAllowRevert(NewSweepInput(testTokenA, MsgSender(), big.NewInt(0)))).Encode()
	require.NoError(t, err)
	sub, err := DecodeExecuteSubPlanInput(subData)
	require.NoError(t, err)
	require.Len(t, sub.Plan.Steps, 1)
	require.True(t, sub.Plan.Steps[0].AllowRevert)
	require.Equal(t, CommandSweep, sub.Plan.Steps[0].Input.Command())
}

func TestInputsValidate(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB, testTokenC}, []*big.Int{big.NewInt(3000), big.NewInt(500)})
	require.NoError(t, err)
	one := big.NewInt(1)
	details := permit2.NewPermitDetails(testTokenA, one, one, one)
	key := v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(3000), big.NewInt(60), types.ZeroAddress())
	increaseData, err := v4.EncodeModifyLiquidities([]byte{v4.ActionSettlePair, v4.ActionIncreaseLiquidity}, [][]byte{{0x01}, {0x02}}, one)
	require.NoError(t, err)
	burnData, err := v4.EncodeModifyLiquidities([]byte{v4.ActionBurnPosition}, [][]byte{{0x01}}, one)
	require.NoError(t, err)
	decreaseCalldata, err := v3.EncodeDecreaseLiquidity(one, one, one, one, one)
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"v3 in nil", (*V3SwapExactInInput)(nil).Validate(), "invalid request: input is nil"},
		{"v3 in recipient", (&V3SwapExactInInput{AmountIn: one, AmountOutMin: one, Path: path}).Validate(), "invalid request: recipient is nil"},
		{"v3 in amountIn", (&V3SwapExactInInput{Recipient: testTokenA, AmountOutMin: one, Path: path}).Validate(), "invalid request: amountIn is nil"},
		{"v3 in amountOutMin", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, Path: path}).Validate(), "invalid request: amountOutMin is nil"},
		{"v3 in empty path", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one}).Validate(), "invalid request: path length 0 is not 20 + 23*n with n >= 1"},
		{"v3 in short path", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: path[:42]}).Validate(), "invalid request: path length 42 is not 20 + 23*n with n >= 1"},
		{"v3 in ragged path", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: path[:65]}).Validate(), "invalid request: path length 65 is not 20 + 23*n with n >= 1"},
		{"v3 in hop mismatch", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: path, MinHopPriceX36: []*big.Int{one}}).Validate(), "invalid request: minHopPriceX36 length 1 does not match 2 hops"},
		{"v3 in nil hop", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: path, MinHopPriceX36: []*big.Int{one, nil}}).Validate(), "invalid request: minHopPriceX36[1] is nil"},
		{"v3 in valid", (&V3SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: path[:43]}).Validate(), ""},
		{"v3 out nil", (*V3SwapExactOutInput)(nil).Validate(), "invalid request: input is nil"},
		{"v3 out amountOut", (&V3SwapExactOutInput{Recipient: testTokenA, AmountInMax: one, Path: path}).Validate(), "invalid request: amountOut is nil"},
		{"v3 out amountInMax", (&V3SwapExactOutInput{Recipient: testTokenA, AmountOut: one, Path: path}).Validate(), "invalid request: amountInMax is nil"},
		{"v3 out valid", (&V3SwapExactOutInput{Recipient: testTokenA, AmountOut: one, AmountInMax: one, Path: path, MinHopPriceX36: []*big.Int{one, one}}).Validate(), ""},
		{"v2 in nil", (*V2SwapExactInInput)(nil).Validate(), "invalid request: input is nil"},
		{"v2 in recipient", (&V2SwapExactInInput{AmountIn: one, AmountOutMin: one, Path: []*types.Address{testTokenA, testTokenB}}).Validate(), "invalid request: recipient is nil"},
		{"v2 in one token", (&V2SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: []*types.Address{testTokenA}}).Validate(), "invalid request: path needs at least 2 tokens, got 1"},
		{"v2 in nil token", (&V2SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: []*types.Address{testTokenA, nil}}).Validate(), "invalid request: path[1] is nil"},
		{"v2 in hop mismatch", (&V2SwapExactInInput{Recipient: testTokenA, AmountIn: one, AmountOutMin: one, Path: []*types.Address{testTokenA, testTokenB}, MinHopPriceX36: []*big.Int{one, one}}).Validate(), "invalid request: minHopPriceX36 length 2 does not match 1 hops"},
		{"v2 out amountInMax", (&V2SwapExactOutInput{Recipient: testTokenA, AmountOut: one, Path: []*types.Address{testTokenA, testTokenB}}).Validate(), "invalid request: amountInMax is nil"},
		{"v2 out nil hop", (&V2SwapExactOutInput{Recipient: testTokenA, AmountOut: one, AmountInMax: one, Path: []*types.Address{testTokenA, testTokenB}, MinHopPriceX36: []*big.Int{nil}}).Validate(), "invalid request: minHopPriceX36[0] is nil"},
		{"v2 out valid", (&V2SwapExactOutInput{Recipient: testTokenA, AmountOut: one, AmountInMax: one, Path: []*types.Address{testTokenA, testTokenB, testTokenC}}).Validate(), ""},
		{"permit2 transferFrom token", (&Permit2TransferFromInput{Recipient: testTokenA, Amount: one}).Validate(), "invalid request: token is nil"},
		{"permit2 transferFrom recipient", (&Permit2TransferFromInput{Token: testTokenA, Amount: one}).Validate(), "invalid request: recipient is nil"},
		{"permit2 transferFrom amount", (&Permit2TransferFromInput{Token: testTokenA, Recipient: testTokenB}).Validate(), "invalid request: amount is nil"},
		{"permit2 permitBatch nil batch", (&Permit2PermitBatchInput{Signature: []byte{0x01}}).Validate(), "invalid request: permitBatch is nil"},
		{"permit2 permitBatch empty details", (&Permit2PermitBatchInput{PermitBatch: permit2.NewPermitBatch(nil, testTokenB, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitBatch.details is empty"},
		{"permit2 permitBatch details token", (&Permit2PermitBatchInput{PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details, {Amount: one, Expiration: one, Nonce: one}}, testTokenB, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitBatch.details[1].token is nil"},
		{"permit2 permitBatch spender", (&Permit2PermitBatchInput{PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details}, nil, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitBatch.spender is nil"},
		{"permit2 permitBatch sigDeadline", (&Permit2PermitBatchInput{PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details}, testTokenB, nil), Signature: []byte{0x01}}).Validate(), "invalid request: permitBatch.sigDeadline is nil"},
		{"permit2 permitBatch signature", (&Permit2PermitBatchInput{PermitBatch: permit2.NewPermitBatch([]permit2.PermitDetails{*details}, testTokenB, one)}).Validate(), "invalid request: signature is empty"},
		{"permit2 permit nil single", (&Permit2PermitInput{Signature: []byte{0x01}}).Validate(), "invalid request: permitSingle is nil"},
		{"permit2 permit nil details", (&Permit2PermitInput{PermitSingle: permit2.NewPermitSingle(nil, testTokenB, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitSingle.details is nil"},
		{"permit2 permit details nonce", (&Permit2PermitInput{PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(testTokenA, one, one, nil), testTokenB, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitSingle.details.nonce is nil"},
		{"permit2 permit details expiration", (&Permit2PermitInput{PermitSingle: permit2.NewPermitSingle(permit2.NewPermitDetails(testTokenA, one, nil, one), testTokenB, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitSingle.details.expiration is nil"},
		{"permit2 permit spender", (&Permit2PermitInput{PermitSingle: permit2.NewPermitSingle(details, nil, one), Signature: []byte{0x01}}).Validate(), "invalid request: permitSingle.spender is nil"},
		{"permit2 permit signature", (&Permit2PermitInput{PermitSingle: permit2.NewPermitSingle(details, testTokenB, one), Signature: []byte{}}).Validate(), "invalid request: signature is empty"},
		{"sweep token", (&SweepInput{Recipient: testTokenA, AmountMin: one}).Validate(), "invalid request: token is nil"},
		{"sweep amountMin", (&SweepInput{Token: types.ZeroAddress(), Recipient: testTokenA}).Validate(), "invalid request: amountMin is nil"},
		{"transfer value", (&TransferInput{Token: testTokenA, Recipient: testTokenB}).Validate(), "invalid request: value is nil"},
		{"payPortion bips", (&PayPortionInput{Token: testTokenA, Recipient: testTokenB}).Validate(), "invalid request: bips is nil"},
		{"payPortion bips too large", (&PayPortionInput{Token: testTokenA, Recipient: testTokenB, Bips: big.NewInt(10_001)}).Validate(), "invalid request: bips 10001 exceeds 10000"},
		{"payPortion bips max", (&PayPortionInput{Token: testTokenA, Recipient: testTokenB, Bips: big.NewInt(10_000)}).Validate(), ""},
		{"payPortionFull portion too large", (&PayPortionFullPrecisionInput{Token: testTokenA, Recipient: testTokenB, Portion: big.NewInt(1_000_000_000_000_000_001)}).Validate(), "invalid request: portion 1000000000000000001 exceeds 1000000000000000000"},
		{"payPortionFull portion max", (&PayPortionFullPrecisionInput{Token: testTokenA, Recipient: testTokenB, Portion: big.NewInt(1_000_000_000_000_000_000)}).Validate(), ""},
		{"payPortionFull recipient", (&PayPortionFullPrecisionInput{Token: testTokenA, Portion: one}).Validate(), "invalid request: recipient is nil"},
		{"wrap recipient", (&WrapETHInput{Amount: one}).Validate(), "invalid request: recipient is nil"},
		{"wrap amount", (&WrapETHInput{Recipient: testTokenA}).Validate(), "invalid request: amount is nil"},
		{"unwrap amountMin", (&UnwrapWETHInput{Recipient: testTokenA}).Validate(), "invalid request: amountMin is nil"},
		{"transferFromBatch empty", (&Permit2TransferFromBatchInput{}).Validate(), "invalid request: details is empty"},
		{"transferFromBatch from", (&Permit2TransferFromBatchInput{Details: []permit2.AllowanceTransferDetails{{To: testTokenA, Amount: one, Token: testTokenB}}}).Validate(), "invalid request: details[0].from is nil"},
		{"transferFromBatch token", (&Permit2TransferFromBatchInput{Details: []permit2.AllowanceTransferDetails{{From: testTokenA, To: testTokenA, Amount: one, Token: testTokenB}, {From: testTokenA, To: testTokenA, Amount: one}}}).Validate(), "invalid request: details[1].token is nil"},
		{"balanceCheck owner", (&BalanceCheckERC20Input{Token: testTokenA, MinBalance: one}).Validate(), "invalid request: owner is nil"},
		{"balanceCheck minBalance", (&BalanceCheckERC20Input{Owner: testTokenA, Token: testTokenB}).Validate(), "invalid request: minBalance is nil"},
		{"v4Swap empty", (&V4SwapInput{}).Validate(), "invalid request: actions is empty"},
		{"v4Swap mismatch", (&V4SwapInput{Actions: []byte{v4.ActionSettleAll}, Params: [][]byte{{0x01}, {0x02}}}).Validate(), "invalid request: params length 2 does not match actions length 1"},
		{"v3 pm permit short", (&V3PositionManagerPermitInput{Calldata: []byte{0x7a, 0xc2}}).Validate(), "invalid request: calldata is shorter than a selector"},
		{"v3 pm permit selector", (&V3PositionManagerPermitInput{Calldata: decreaseCalldata}).Validate(), "invalid request: calldata selector 0c49ccbe is not permit"},
		{"v3 pm call short", (&V3PositionManagerCallInput{Calldata: decreaseCalldata[:35]}).Validate(), "invalid request: calldata is shorter than a selector and tokenId"},
		{"v3 pm call selector", (&V3PositionManagerCallInput{Calldata: append([]byte{0x7a, 0xc2, 0xff, 0x7b}, make([]byte, 32)...)}).Validate(), "invalid request: calldata selector 7ac2ff7b is not decreaseLiquidity, collect or burn"},
		{"v3 pm call decrease", (&V3PositionManagerCallInput{Calldata: decreaseCalldata}).Validate(), ""},
		{"v4 init key", (&V4InitializePoolInput{SqrtPriceX96: one}).Validate(), "invalid request: key is nil"},
		{"v4 init hooks", (&V4InitializePoolInput{Key: v4.NewPoolKey(testTokenA, testTokenB, one, one, nil), SqrtPriceX96: one}).Validate(), "invalid request: key.hooks is nil"},
		{"v4 init sqrtPrice", (&V4InitializePoolInput{Key: key}).Validate(), "invalid request: sqrtPriceX96 is nil"},
		{"v4 pm call short", (&V4PositionManagerCallInput{Calldata: []byte{0xdd}}).Validate(), "invalid request: calldata is shorter than a selector"},
		{"v4 pm call selector", (&V4PositionManagerCallInput{Calldata: decreaseCalldata}).Validate(), "invalid request: calldata selector 0c49ccbe is not modifyLiquidities"},
		{"v4 pm call increase", (&V4PositionManagerCallInput{Calldata: increaseData}).Validate(), "invalid request: calldata actions[1] 0x00 is not allowed, only mint"},
		{"v4 pm call burn", (&V4PositionManagerCallInput{Calldata: burnData}).Validate(), "invalid request: calldata actions[0] 0x03 is not allowed, only mint"},
		{"subPlan nil plan", (&ExecuteSubPlanInput{}).Validate(), "invalid request: plan is nil"},
		{"subPlan empty plan", (&ExecuteSubPlanInput{Plan: NewPlan()}).Validate(), "plan: invalid request: plan has no steps"},
		{"subPlan invalid step", (&ExecuteSubPlanInput{Plan: NewPlan().Add(&SweepInput{})}).Validate(), "plan: steps[0] SWEEP: invalid request: token is nil"},
		{"across depositor", (&AcrossV4DepositV3Input{}).Validate(), "invalid request: depositor is nil"},
		{"across outputToken", (&AcrossV4DepositV3Input{Depositor: testTokenA, Recipient: testTokenA, InputToken: testTokenA}).Validate(), "invalid request: outputToken is nil"},
		{"across exclusiveRelayer", (&AcrossV4DepositV3Input{Depositor: testTokenA, Recipient: testTokenA, InputToken: testTokenA, OutputToken: testTokenA, InputAmount: one, OutputAmount: one, DestinationChainID: one}).Validate(), "invalid request: exclusiveRelayer is nil"},
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

func TestInputsEncodeRejectInvalidAndOutOfRange(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB}, []*big.Int{big.NewInt(500)})
	require.NoError(t, err)
	uint256Overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	_, err = (&SweepInput{}).Encode()
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "universalrouter: encode SWEEP: invalid request: token is nil")

	_, err = NewV3SwapExactInInput(testTokenA, big.NewInt(-1), big.NewInt(0), path, true, nil).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "universalrouter: encode V3_SWAP_EXACT_IN: ")

	_, err = NewPermit2TransferFromInput(testTokenA, testTokenB, new(big.Int).Lsh(big.NewInt(1), 160)).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, err = NewWrapETHInput(testTokenA, uint256Overflow).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, err = NewV4InitializePoolInput(v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(1<<24), big.NewInt(60), testTokenC), big.NewInt(1)).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, err = NewPermit2PermitInput(permit2.NewPermitSingle(permit2.NewPermitDetails(testTokenA, big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 48), big.NewInt(1)), testTokenB, big.NewInt(1)), []byte{0x01}).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, err = NewExecuteSubPlanInput(NewPlan().Add(NewWrapETHInput(testTokenA, uint256Overflow))).Encode()
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "universalrouter: encode EXECUTE_SUB_PLAN: universalrouter: encodePlan: steps[0]: ")
}

func TestDecodeInputRejectsUnknownAndMalformed(t *testing.T) {
	for _, b := range []byte{0x0f, 0x15, 0x20, 0x22, 0x3f, 0x41, 0x7f} {
		_, err := DecodeInput(Command(b), nil)
		require.ErrorIs(t, err, ErrUnknownCommand)
	}

	for _, cmd := range []Command{
		CommandV3SwapExactIn, CommandV3SwapExactOut, CommandPermit2TransferFrom, CommandPermit2PermitBatch, CommandSweep,
		CommandTransfer, CommandPayPortion, CommandPayPortionFullPrecision, CommandV2SwapExactIn, CommandV2SwapExactOut,
		CommandPermit2Permit, CommandWrapETH, CommandUnwrapWETH, CommandPermit2TransferFromBatch, CommandBalanceCheckERC20,
		CommandV4Swap, CommandV4InitializePool, CommandExecuteSubPlan, CommandAcrossV4DepositV3,
	} {
		for _, data := range [][]byte{nil, {0x01}, make([]byte, 31)} {
			input, err := DecodeInput(cmd, data)
			require.Nil(t, input, "%s", cmd)
			require.ErrorContains(t, err, "universalrouter: decode "+cmd.String()+": ")
		}
	}

	for _, cmd := range []Command{CommandV3PositionManagerPermit, CommandV3PositionManagerCall, CommandV4PositionManagerCall} {
		input, err := DecodeInput(cmd, []byte{0x01})
		require.NoError(t, err)
		require.Equal(t, cmd, input.Command())
		require.Error(t, input.Validate())
	}
}

func TestRecipientAndAmountConstantsMatchActionConstants(t *testing.T) {
	require.Equal(t, "0x0000000000000000000000000000000000000001", MsgSender().String())
	require.Equal(t, "0x0000000000000000000000000000000000000002", AddressThis().String())
	require.Zero(t, new(big.Int).Lsh(big.NewInt(1), 255).Cmp(ContractBalance()))
	require.NotSame(t, MsgSender(), MsgSender())
	require.NotSame(t, ContractBalance(), ContractBalance())
}

func TestTypedDecodersMatchDecodeInput(t *testing.T) {
	path, err := v3.EncodePath([]*types.Address{testTokenA, testTokenB}, []*big.Int{big.NewInt(500)})
	require.NoError(t, err)
	burnCalldata, err := v3.EncodeNfpmBurn(big.NewInt(1))
	require.NoError(t, err)
	v4PMCalldata, err := v4.EncodeModifyLiquidities([]byte{v4.ActionMintPosition}, [][]byte{{0x01}}, big.NewInt(1))
	require.NoError(t, err)
	details := permit2.NewPermitDetails(testTokenA, big.NewInt(1), big.NewInt(2), big.NewInt(3))

	for _, tc := range []struct {
		input  Input
		decode func([]byte) (Input, error)
	}{
		{NewV3SwapExactInInput(testTokenA, big.NewInt(1), big.NewInt(2), path, true, nil), func(b []byte) (Input, error) { return DecodeV3SwapExactInInput(b) }},
		{NewV3SwapExactOutInput(testTokenA, big.NewInt(1), big.NewInt(2), path, false, []*big.Int{big.NewInt(3)}), func(b []byte) (Input, error) { return DecodeV3SwapExactOutInput(b) }},
		{NewPermit2TransferFromInput(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodePermit2TransferFromInput(b) }},
		{NewPermit2PermitBatchInput(permit2.NewPermitBatch([]permit2.PermitDetails{*details}, testTokenB, big.NewInt(4)), []byte{0x01}), func(b []byte) (Input, error) { return DecodePermit2PermitBatchInput(b) }},
		{NewSweepInput(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodeSweepInput(b) }},
		{NewTransferInput(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodeTransferInput(b) }},
		{NewPayPortionInput(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodePayPortionInput(b) }},
		{NewPayPortionFullPrecisionInput(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodePayPortionFullPrecisionInput(b) }},
		{NewV2SwapExactInInput(testTokenA, big.NewInt(1), big.NewInt(2), []*types.Address{testTokenA, testTokenB}, true, nil), func(b []byte) (Input, error) { return DecodeV2SwapExactInInput(b) }},
		{NewV2SwapExactOutInput(testTokenA, big.NewInt(1), big.NewInt(2), []*types.Address{testTokenA, testTokenB}, true, []*big.Int{big.NewInt(9)}), func(b []byte) (Input, error) { return DecodeV2SwapExactOutInput(b) }},
		{NewPermit2PermitInput(permit2.NewPermitSingle(details, testTokenB, big.NewInt(4)), []byte{0x01}), func(b []byte) (Input, error) { return DecodePermit2PermitInput(b) }},
		{NewWrapETHInput(testTokenA, big.NewInt(1)), func(b []byte) (Input, error) { return DecodeWrapETHInput(b) }},
		{NewUnwrapWETHInput(testTokenA, big.NewInt(1)), func(b []byte) (Input, error) { return DecodeUnwrapWETHInput(b) }},
		{NewPermit2TransferFromBatchInput([]permit2.AllowanceTransferDetails{*permit2.NewAllowanceTransferDetails(testTokenA, testTokenB, big.NewInt(1), testTokenC)}), func(b []byte) (Input, error) { return DecodePermit2TransferFromBatchInput(b) }},
		{NewBalanceCheckERC20Input(testTokenA, testTokenB, big.NewInt(1)), func(b []byte) (Input, error) { return DecodeBalanceCheckERC20Input(b) }},
		{NewV4SwapInput([]byte{v4.ActionSettleAll}, [][]byte{{0x01}}), func(b []byte) (Input, error) { return DecodeV4SwapInput(b) }},
		{NewV3PositionManagerPermitInput(append([]byte{0x7a, 0xc2, 0xff, 0x7b}, make([]byte, 32)...)), func(b []byte) (Input, error) { return DecodeV3PositionManagerPermitInput(b) }},
		{NewV3PositionManagerCallInput(burnCalldata), func(b []byte) (Input, error) { return DecodeV3PositionManagerCallInput(b) }},
		{NewV4InitializePoolInput(v4.NewPoolKey(testTokenA, testTokenB, big.NewInt(500), big.NewInt(10), testTokenC), big.NewInt(1)), func(b []byte) (Input, error) { return DecodeV4InitializePoolInput(b) }},
		{NewV4PositionManagerCallInput(v4PMCalldata), func(b []byte) (Input, error) { return DecodeV4PositionManagerCallInput(b) }},
		{NewExecuteSubPlanInput(NewPlan().Add(NewWrapETHInput(testTokenA, big.NewInt(1)))), func(b []byte) (Input, error) { return DecodeExecuteSubPlanInput(b) }},
		{NewAcrossV4DepositV3Input(testTokenA, testTokenB, testTokenC, testTokenA, big.NewInt(1), big.NewInt(1), big.NewInt(1), testTokenB, 1, 2, 3, nil, true), func(b []byte) (Input, error) { return DecodeAcrossV4DepositV3Input(b) }},
	} {
		cmd := tc.input.Command()
		t.Run(cmd.String(), func(t *testing.T) {
			data, err := tc.input.Encode()
			require.NoError(t, err)

			typed, err := tc.decode(data)
			require.NoError(t, err)
			require.Equal(t, cmd, typed.Command())
			typedAgain, err := typed.Encode()
			require.NoError(t, err)
			require.Equal(t, data, typedAgain)

			generic, err := DecodeInput(cmd, data)
			require.NoError(t, err)
			require.IsType(t, typed, generic)
			genericAgain, err := generic.Encode()
			require.NoError(t, err)
			require.Equal(t, data, genericAgain)

			raw := cmd == CommandV3PositionManagerPermit || cmd == CommandV3PositionManagerCall || cmd == CommandV4PositionManagerCall
			_, typedErr := tc.decode([]byte{0x01})
			input, genericErr := DecodeInput(cmd, []byte{0x01})
			if raw {
				require.NoError(t, typedErr)
				require.NoError(t, genericErr)
				return
			}
			require.ErrorContains(t, typedErr, "universalrouter: decode "+cmd.String()+": ")
			require.EqualError(t, genericErr, typedErr.Error())
			require.True(t, input == nil)
		})
	}

	input, err := DecodeInput(Command(0x0f), nil)
	require.True(t, input == nil)
	require.ErrorIs(t, err, ErrUnknownCommand)
}
