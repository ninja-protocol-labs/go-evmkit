package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPoolKey(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	k := NewPoolKey(currency0, currency1, fee, tickSpacing, hooks)
	require.Same(t, currency0, k.Currency0)
	require.Same(t, currency1, k.Currency1)
	require.Same(t, fee, k.Fee)
	require.Same(t, tickSpacing, k.TickSpacing)
	require.Same(t, hooks, k.Hooks)
}

func TestNewBalanceDelta(t *testing.T) {
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(-50)

	d := NewBalanceDelta(amount0, amount1)
	require.Same(t, amount0, d.Amount0)
	require.Same(t, amount1, d.Amount1)
}

func TestNewLiquidityParams(t *testing.T) {
	tickLower := big.NewInt(-100)
	tickUpper := big.NewInt(100)
	liquidityDelta := big.NewInt(1000)
	salt := [32]byte{0x01, 0x02}

	p := NewLiquidityParams(tickLower, tickUpper, liquidityDelta, salt)
	require.Same(t, tickLower, p.TickLower)
	require.Same(t, tickUpper, p.TickUpper)
	require.Same(t, liquidityDelta, p.LiquidityDelta)
	require.Equal(t, salt, p.Salt)
}

func TestNewSwapAmountParams(t *testing.T) {
	amountSpecified := big.NewInt(-500)
	sqrtPriceLimitX96 := big.NewInt(123456789)

	p := NewSwapAmountParams(true, amountSpecified, sqrtPriceLimitX96)
	require.True(t, p.ZeroForOne)
	require.Same(t, amountSpecified, p.AmountSpecified)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewInitializeEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x01}
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	hooks := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	sqrtPriceX96 := big.NewInt(792281625142643375)
	tick := big.NewInt(0)

	init := NewInitializeEvent(poolManager, id, currency0, currency1, fee, tickSpacing, hooks, sqrtPriceX96, tick)
	require.Same(t, poolManager, init.Contract)
	require.Equal(t, id, init.Id)
	require.Same(t, currency0, init.Currency0)
	require.Same(t, currency1, init.Currency1)
	require.Same(t, fee, init.Fee)
	require.Same(t, tickSpacing, init.TickSpacing)
	require.Same(t, hooks, init.Hooks)
	require.Same(t, sqrtPriceX96, init.SqrtPriceX96)
	require.Same(t, tick, init.Tick)
}

func TestNewModifyLiquidityEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x02}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tickLower := big.NewInt(-100)
	tickUpper := big.NewInt(100)
	liquidityDelta := big.NewInt(1000)
	salt := [32]byte{0x03, 0x04}

	m := NewModifyLiquidityEvent(poolManager, id, sender, tickLower, tickUpper, liquidityDelta, salt)
	require.Same(t, poolManager, m.Contract)
	require.Equal(t, id, m.Id)
	require.Same(t, sender, m.Sender)
	require.Same(t, tickLower, m.TickLower)
	require.Same(t, tickUpper, m.TickUpper)
	require.Same(t, liquidityDelta, m.LiquidityDelta)
	require.Equal(t, salt, m.Salt)
}

func TestNewSwapEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x05}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(-50)
	sqrtPriceX96 := big.NewInt(792281625142643375)
	liquidity := big.NewInt(1000)
	tick := big.NewInt(42)
	fee := big.NewInt(3000)

	s := NewSwapEvent(poolManager, id, sender, amount0, amount1, sqrtPriceX96, liquidity, tick, fee)
	require.Same(t, poolManager, s.Contract)
	require.Equal(t, id, s.Id)
	require.Same(t, sender, s.Sender)
	require.Same(t, amount0, s.Amount0)
	require.Same(t, amount1, s.Amount1)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, liquidity, s.Liquidity)
	require.Same(t, tick, s.Tick)
	require.Same(t, fee, s.Fee)
}

func TestNewDonateEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x06}
	sender := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	d := NewDonateEvent(poolManager, id, sender, amount0, amount1)
	require.Same(t, poolManager, d.Contract)
	require.Equal(t, id, d.Id)
	require.Same(t, sender, d.Sender)
	require.Same(t, amount0, d.Amount0)
	require.Same(t, amount1, d.Amount1)
}

func TestNewProtocolFeeControllerUpdatedEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	controller := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	u := NewProtocolFeeControllerUpdatedEvent(poolManager, controller)
	require.Same(t, poolManager, u.Contract)
	require.Same(t, controller, u.ProtocolFeeController)
}

func TestNewProtocolFeeUpdatedEvent(t *testing.T) {
	poolManager := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	id := PoolID{0x07}
	protocolFee := big.NewInt(500)

	u := NewProtocolFeeUpdatedEvent(poolManager, id, protocolFee)
	require.Same(t, poolManager, u.Contract)
	require.Equal(t, id, u.Id)
	require.Same(t, protocolFee, u.ProtocolFee)
}

func TestPoolManagerRequestsValidate(t *testing.T) {
	pm := testPoolManager
	a := testSender
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	liq := NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(1), [32]byte{})
	swap := NewSwapAmountParams(true, big.NewInt(-1), big.NewInt(4295128740))
	one := big.NewInt(1)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"protocolFeesAccrued nil", (*ProtocolFeesAccruedRequest)(nil).Validate(), "invalid request: request is nil"},
		{"protocolFeesAccrued empty", (&ProtocolFeesAccruedRequest{}).Validate(), "invalid request: contract is nil"},
		{"protocolFeesAccrued nil currency", (&ProtocolFeesAccruedRequest{Contract: pm}).Validate(), "invalid request: currency is nil"},
		{"protocolFeesAccrued native", (&ProtocolFeesAccruedRequest{Contract: pm, Currency: types.ZeroAddress()}).Validate(), ""},
		{"protocolFeeController nil", (*ProtocolFeeControllerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"protocolFeeController empty", (&ProtocolFeeControllerRequest{}).Validate(), "invalid request: contract is nil"},
		{"protocolFeeController valid", (&ProtocolFeeControllerRequest{Contract: pm}).Validate(), ""},
		{"extsload nil", (*ExtsloadRequest)(nil).Validate(), "invalid request: request is nil"},
		{"extsload empty", (&ExtsloadRequest{}).Validate(), "invalid request: contract is nil"},
		{"extsload zero slot", (&ExtsloadRequest{Contract: pm}).Validate(), ""},
		{"extsloadRange nil", (*ExtsloadRangeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"extsloadRange empty", (&ExtsloadRangeRequest{}).Validate(), "invalid request: contract is nil"},
		{"extsloadRange nil nSlots", (&ExtsloadRangeRequest{Contract: pm}).Validate(), "invalid request: nSlots is nil"},
		{"extsloadRange zero nSlots", (&ExtsloadRangeRequest{Contract: pm, NSlots: big.NewInt(0)}).Validate(), ""},
		{"extsloadSlots nil", (*ExtsloadSlotsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"extsloadSlots empty", (&ExtsloadSlotsRequest{}).Validate(), "invalid request: contract is nil"},
		{"extsloadSlots nil slots", (&ExtsloadSlotsRequest{Contract: pm}).Validate(), ""},
		{"exttload nil", (*ExttloadRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exttload empty", (&ExttloadRequest{}).Validate(), "invalid request: contract is nil"},
		{"exttload valid", (&ExttloadRequest{Contract: pm}).Validate(), ""},
		{"exttloadSlots nil", (*ExttloadSlotsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exttloadSlots empty", (&ExttloadSlotsRequest{}).Validate(), "invalid request: contract is nil"},
		{"exttloadSlots valid", (&ExttloadSlotsRequest{Contract: pm, Slots: [][32]byte{{}}}).Validate(), ""},
		{"initialize nil", (*InitializeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"initialize empty", (&InitializeRequest{}).Validate(), "invalid request: contract is nil"},
		{"initialize nil key", (&InitializeRequest{Contract: pm, SqrtPriceX96: one}).Validate(), "invalid request: key is nil"},
		{"initialize nil currency0", (&InitializeRequest{Contract: pm, Key: &PoolKey{Currency1: testCurrency1, Fee: one, TickSpacing: one, Hooks: testHooks}, SqrtPriceX96: one}).Validate(), "invalid request: key.currency0 is nil"},
		{"initialize nil currency1", (&InitializeRequest{Contract: pm, Key: &PoolKey{Currency0: testCurrency0, Fee: one, TickSpacing: one, Hooks: testHooks}, SqrtPriceX96: one}).Validate(), "invalid request: key.currency1 is nil"},
		{"initialize nil fee", (&InitializeRequest{Contract: pm, Key: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, TickSpacing: one, Hooks: testHooks}, SqrtPriceX96: one}).Validate(), "invalid request: key.fee is nil"},
		{"initialize nil tickSpacing", (&InitializeRequest{Contract: pm, Key: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, Fee: one, Hooks: testHooks}, SqrtPriceX96: one}).Validate(), "invalid request: key.tickSpacing is nil"},
		{"initialize nil hooks", (&InitializeRequest{Contract: pm, Key: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, Fee: one, TickSpacing: one}, SqrtPriceX96: one}).Validate(), "invalid request: key.hooks is nil"},
		{"initialize nil sqrtPriceX96", (&InitializeRequest{Contract: pm, Key: key}).Validate(), "invalid request: sqrtPriceX96 is nil"},
		{"initialize valid", (&InitializeRequest{Contract: pm, Key: key, SqrtPriceX96: one}).Validate(), ""},
		{"initialize unordered currencies", (&InitializeRequest{Contract: pm, Key: NewPoolKey(testCurrency1, testCurrency0, one, one, types.ZeroAddress()), SqrtPriceX96: one}).Validate(), ""},
		{"modifyLiquidity nil", (*ModifyLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"modifyLiquidity empty", (&ModifyLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"modifyLiquidity nil key", (&ModifyLiquidityRequest{Contract: pm, Params: liq}).Validate(), "invalid request: key is nil"},
		{"modifyLiquidity nil params", (&ModifyLiquidityRequest{Contract: pm, Key: key}).Validate(), "invalid request: params is nil"},
		{"modifyLiquidity nil tickLower", (&ModifyLiquidityRequest{Contract: pm, Key: key, Params: &LiquidityParams{TickUpper: one, LiquidityDelta: one}}).Validate(), "invalid request: params.tickLower is nil"},
		{"modifyLiquidity nil tickUpper", (&ModifyLiquidityRequest{Contract: pm, Key: key, Params: &LiquidityParams{TickLower: one, LiquidityDelta: one}}).Validate(), "invalid request: params.tickUpper is nil"},
		{"modifyLiquidity nil liquidityDelta", (&ModifyLiquidityRequest{Contract: pm, Key: key, Params: &LiquidityParams{TickLower: one, TickUpper: one}}).Validate(), "invalid request: params.liquidityDelta is nil"},
		{"modifyLiquidity valid", (&ModifyLiquidityRequest{Contract: pm, Key: key, Params: liq}).Validate(), ""},
		{"modifyLiquidity negative delta", (&ModifyLiquidityRequest{Contract: pm, Key: key, Params: NewLiquidityParams(one, one, big.NewInt(-1), [32]byte{}), HookData: []byte{0x01}}).Validate(), ""},
		{"swap nil", (*SwapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swap empty", (&SwapRequest{}).Validate(), "invalid request: contract is nil"},
		{"swap nil key", (&SwapRequest{Contract: pm, Params: swap}).Validate(), "invalid request: key is nil"},
		{"swap bad key", (&SwapRequest{Contract: pm, Key: &PoolKey{}, Params: swap}).Validate(), "invalid request: key.currency0 is nil"},
		{"swap nil params", (&SwapRequest{Contract: pm, Key: key}).Validate(), "invalid request: params is nil"},
		{"swap nil amountSpecified", (&SwapRequest{Contract: pm, Key: key, Params: &SwapAmountParams{SqrtPriceLimitX96: one}}).Validate(), "invalid request: params.amountSpecified is nil"},
		{"swap nil sqrtPriceLimitX96", (&SwapRequest{Contract: pm, Key: key, Params: &SwapAmountParams{AmountSpecified: one}}).Validate(), "invalid request: params.sqrtPriceLimitX96 is nil"},
		{"swap valid", (&SwapRequest{Contract: pm, Key: key, Params: swap}).Validate(), ""},
		{"donate nil", (*DonateRequest)(nil).Validate(), "invalid request: request is nil"},
		{"donate empty", (&DonateRequest{}).Validate(), "invalid request: contract is nil"},
		{"donate nil key", (&DonateRequest{Contract: pm, Amount0: one, Amount1: one}).Validate(), "invalid request: key is nil"},
		{"donate nil amount0", (&DonateRequest{Contract: pm, Key: key, Amount1: one}).Validate(), "invalid request: amount0 is nil"},
		{"donate nil amount1", (&DonateRequest{Contract: pm, Key: key, Amount0: one}).Validate(), "invalid request: amount1 is nil"},
		{"donate valid", (&DonateRequest{Contract: pm, Key: key, Amount0: one, Amount1: big.NewInt(0)}).Validate(), ""},
		{"unlock nil", (*UnlockRequest)(nil).Validate(), "invalid request: request is nil"},
		{"unlock empty", (&UnlockRequest{}).Validate(), "invalid request: contract is nil"},
		{"unlock nil data", (&UnlockRequest{Contract: pm}).Validate(), ""},
		{"sync nil", (*SyncRequest)(nil).Validate(), "invalid request: request is nil"},
		{"sync empty", (&SyncRequest{}).Validate(), "invalid request: contract is nil"},
		{"sync nil currency", (&SyncRequest{Contract: pm}).Validate(), "invalid request: currency is nil"},
		{"sync native", (&SyncRequest{Contract: pm, Currency: types.ZeroAddress()}).Validate(), ""},
		{"take nil", (*TakeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"take empty", (&TakeRequest{}).Validate(), "invalid request: contract is nil"},
		{"take nil currency", (&TakeRequest{Contract: pm, To: a, Amount: one}).Validate(), "invalid request: currency is nil"},
		{"take nil to", (&TakeRequest{Contract: pm, Currency: testCurrency0, Amount: one}).Validate(), "invalid request: to is nil"},
		{"take nil amount", (&TakeRequest{Contract: pm, Currency: testCurrency0, To: a}).Validate(), "invalid request: amount is nil"},
		{"take valid", (&TakeRequest{Contract: pm, Currency: testCurrency0, To: a, Amount: one}).Validate(), ""},
		{"settle nil", (*SettleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"settle empty", (&SettleRequest{}).Validate(), "invalid request: contract is nil"},
		{"settle negative nativeAmount", (&SettleRequest{Contract: pm, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"settle nil nativeAmount", (&SettleRequest{Contract: pm}).Validate(), ""},
		{"settle zero nativeAmount", (&SettleRequest{Contract: pm, NativeAmount: big.NewInt(0)}).Validate(), ""},
		{"settle max nativeAmount", (&SettleRequest{Contract: pm, NativeAmount: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))}).Validate(), ""},
		{"settle oversized nativeAmount", (&SettleRequest{Contract: pm, NativeAmount: new(big.Int).Lsh(big.NewInt(1), 256)}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"settleFor oversized nativeAmount", (&SettleForRequest{Contract: pm, Recipient: a, NativeAmount: new(big.Int).Lsh(big.NewInt(1), 256)}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"settleFor nil", (*SettleForRequest)(nil).Validate(), "invalid request: request is nil"},
		{"settleFor empty", (&SettleForRequest{}).Validate(), "invalid request: contract is nil"},
		{"settleFor nil recipient", (&SettleForRequest{Contract: pm}).Validate(), "invalid request: recipient is nil"},
		{"settleFor negative nativeAmount", (&SettleForRequest{Contract: pm, Recipient: a, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"settleFor valid", (&SettleForRequest{Contract: pm, Recipient: a, NativeAmount: one}).Validate(), ""},
		{"clear nil", (*ClearRequest)(nil).Validate(), "invalid request: request is nil"},
		{"clear empty", (&ClearRequest{}).Validate(), "invalid request: contract is nil"},
		{"clear nil currency", (&ClearRequest{Contract: pm, Amount: one}).Validate(), "invalid request: currency is nil"},
		{"clear nil amount", (&ClearRequest{Contract: pm, Currency: testCurrency0}).Validate(), "invalid request: amount is nil"},
		{"clear valid", (&ClearRequest{Contract: pm, Currency: testCurrency0, Amount: one}).Validate(), ""},
		{"mint nil", (*MintRequest)(nil).Validate(), "invalid request: request is nil"},
		{"mint empty", (&MintRequest{}).Validate(), "invalid request: contract is nil"},
		{"mint nil to", (&MintRequest{Contract: pm, ID: one, Amount: one}).Validate(), "invalid request: to is nil"},
		{"mint nil id", (&MintRequest{Contract: pm, To: a, Amount: one}).Validate(), "invalid request: id is nil"},
		{"mint nil amount", (&MintRequest{Contract: pm, To: a, ID: one}).Validate(), "invalid request: amount is nil"},
		{"mint valid", (&MintRequest{Contract: pm, To: a, ID: one, Amount: one}).Validate(), ""},
		{"burn nil", (*BurnRequest)(nil).Validate(), "invalid request: request is nil"},
		{"burn empty", (&BurnRequest{}).Validate(), "invalid request: contract is nil"},
		{"burn nil from", (&BurnRequest{Contract: pm, ID: one, Amount: one}).Validate(), "invalid request: from is nil"},
		{"burn nil id", (&BurnRequest{Contract: pm, From: a, Amount: one}).Validate(), "invalid request: id is nil"},
		{"burn nil amount", (&BurnRequest{Contract: pm, From: a, ID: one}).Validate(), "invalid request: amount is nil"},
		{"burn valid", (&BurnRequest{Contract: pm, From: a, ID: one, Amount: one}).Validate(), ""},
		{"updateDynamicLPFee nil", (*UpdateDynamicLPFeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"updateDynamicLPFee empty", (&UpdateDynamicLPFeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"updateDynamicLPFee nil key", (&UpdateDynamicLPFeeRequest{Contract: pm, NewDynamicLPFee: one}).Validate(), "invalid request: key is nil"},
		{"updateDynamicLPFee nil fee", (&UpdateDynamicLPFeeRequest{Contract: pm, Key: key}).Validate(), "invalid request: newDynamicLPFee is nil"},
		{"updateDynamicLPFee valid", (&UpdateDynamicLPFeeRequest{Contract: pm, Key: key, NewDynamicLPFee: one}).Validate(), ""},
		{"setProtocolFee nil", (*SetProtocolFeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setProtocolFee empty", (&SetProtocolFeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"setProtocolFee nil key", (&SetProtocolFeeRequest{Contract: pm, NewProtocolFee: one}).Validate(), "invalid request: key is nil"},
		{"setProtocolFee nil fee", (&SetProtocolFeeRequest{Contract: pm, Key: key}).Validate(), "invalid request: newProtocolFee is nil"},
		{"setProtocolFee valid", (&SetProtocolFeeRequest{Contract: pm, Key: key, NewProtocolFee: one}).Validate(), ""},
		{"setProtocolFeeController nil", (*SetProtocolFeeControllerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setProtocolFeeController empty", (&SetProtocolFeeControllerRequest{}).Validate(), "invalid request: contract is nil"},
		{"setProtocolFeeController nil controller", (&SetProtocolFeeControllerRequest{Contract: pm}).Validate(), "invalid request: controller is nil"},
		{"setProtocolFeeController zero controller", (&SetProtocolFeeControllerRequest{Contract: pm, Controller: types.ZeroAddress()}).Validate(), ""},
		{"collectProtocolFees nil", (*CollectProtocolFeesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"collectProtocolFees empty", (&CollectProtocolFeesRequest{}).Validate(), "invalid request: contract is nil"},
		{"collectProtocolFees nil recipient", (&CollectProtocolFeesRequest{Contract: pm, Currency: testCurrency0, Amount: one}).Validate(), "invalid request: recipient is nil"},
		{"collectProtocolFees nil currency", (&CollectProtocolFeesRequest{Contract: pm, Recipient: a, Amount: one}).Validate(), "invalid request: currency is nil"},
		{"collectProtocolFees nil amount", (&CollectProtocolFeesRequest{Contract: pm, Recipient: a, Currency: testCurrency0}).Validate(), "invalid request: amount is nil"},
		{"collectProtocolFees valid", (&CollectProtocolFeesRequest{Contract: pm, Recipient: a, Currency: testCurrency0, Amount: big.NewInt(0)}).Validate(), ""},
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

func TestNewPoolManagerRequestsSetFields(t *testing.T) {
	pm := testPoolManager
	a := testSender
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	liq := NewLiquidityParams(big.NewInt(-60), big.NewInt(60), big.NewInt(7), [32]byte{0x01})
	swap := NewSwapAmountParams(false, big.NewInt(9), big.NewInt(11))
	slot := [32]byte{0xaa}
	slots := [][32]byte{{0x01}, {0x02}}
	hookData := []byte{0xde, 0xad}

	require.Equal(t, &ProtocolFeesAccruedRequest{Contract: pm, Currency: testCurrency0}, NewProtocolFeesAccruedRequest(pm, testCurrency0))
	require.Equal(t, &ProtocolFeeControllerRequest{Contract: pm}, NewProtocolFeeControllerRequest(pm))
	require.Equal(t, &ExtsloadRequest{Contract: pm, Slot: slot}, NewExtsloadRequest(pm, slot))
	require.Equal(t, &ExtsloadRangeRequest{Contract: pm, StartSlot: slot, NSlots: big.NewInt(3)}, NewExtsloadRangeRequest(pm, slot, big.NewInt(3)))
	require.Equal(t, &ExtsloadSlotsRequest{Contract: pm, Slots: slots}, NewExtsloadSlotsRequest(pm, slots))
	require.Equal(t, &ExttloadRequest{Contract: pm, Slot: slot}, NewExttloadRequest(pm, slot))
	require.Equal(t, &ExttloadSlotsRequest{Contract: pm, Slots: slots}, NewExttloadSlotsRequest(pm, slots))
	require.Equal(t, &InitializeRequest{Contract: pm, Key: key, SqrtPriceX96: big.NewInt(5)}, NewInitializeRequest(pm, key, big.NewInt(5)))
	require.Equal(t, &ModifyLiquidityRequest{Contract: pm, Key: key, Params: liq, HookData: hookData}, NewModifyLiquidityRequest(pm, key, liq, hookData))
	require.Equal(t, &SwapRequest{Contract: pm, Key: key, Params: swap, HookData: hookData}, NewSwapRequest(pm, key, swap, hookData))
	require.Equal(t, &DonateRequest{Contract: pm, Key: key, Amount0: big.NewInt(1), Amount1: big.NewInt(2), HookData: hookData}, NewDonateRequest(pm, key, big.NewInt(1), big.NewInt(2), hookData))
	require.Equal(t, &UnlockRequest{Contract: pm, Data: hookData}, NewUnlockRequest(pm, hookData))
	require.Equal(t, &SyncRequest{Contract: pm, Currency: testCurrency1}, NewSyncRequest(pm, testCurrency1))
	require.Equal(t, &TakeRequest{Contract: pm, Currency: testCurrency0, To: a, Amount: big.NewInt(4)}, NewTakeRequest(pm, testCurrency0, a, big.NewInt(4)))
	require.Equal(t, &SettleRequest{Contract: pm, NativeAmount: big.NewInt(6)}, NewSettleRequest(pm, big.NewInt(6)))
	require.Equal(t, &SettleForRequest{Contract: pm, Recipient: a, NativeAmount: nil}, NewSettleForRequest(pm, a, nil))
	require.Equal(t, &ClearRequest{Contract: pm, Currency: testCurrency1, Amount: big.NewInt(8)}, NewClearRequest(pm, testCurrency1, big.NewInt(8)))
	require.Equal(t, &MintRequest{Contract: pm, To: a, ID: big.NewInt(10), Amount: big.NewInt(12)}, NewMintRequest(pm, a, big.NewInt(10), big.NewInt(12)))
	require.Equal(t, &BurnRequest{Contract: pm, From: a, ID: big.NewInt(13), Amount: big.NewInt(14)}, NewBurnRequest(pm, a, big.NewInt(13), big.NewInt(14)))
	require.Equal(t, &UpdateDynamicLPFeeRequest{Contract: pm, Key: key, NewDynamicLPFee: big.NewInt(500)}, NewUpdateDynamicLPFeeRequest(pm, key, big.NewInt(500)))
	require.Equal(t, &SetProtocolFeeRequest{Contract: pm, Key: key, NewProtocolFee: big.NewInt(100)}, NewSetProtocolFeeRequest(pm, key, big.NewInt(100)))
	require.Equal(t, &SetProtocolFeeControllerRequest{Contract: pm, Controller: a}, NewSetProtocolFeeControllerRequest(pm, a))
	require.Equal(t, &CollectProtocolFeesRequest{Contract: pm, Recipient: a, Currency: testCurrency0, Amount: big.NewInt(15)}, NewCollectProtocolFeesRequest(pm, a, testCurrency0, big.NewInt(15)))

	req := NewModifyLiquidityRequest(pm, key, liq, hookData)
	require.Same(t, key, req.Key)
	require.Same(t, liq, req.Params)
}
