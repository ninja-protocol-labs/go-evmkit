package v4

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// Planner collects v4 actions with their encoded params in order, for V4_SWAP, modifyLiquidities or executeActions.
type Planner struct {
	actions []byte
	params  [][]byte
}

// NewPlanner returns an empty Planner.
func NewPlanner() *Planner {
	return new(Planner)
}

// Append appends action with already encoded params.
func (p *Planner) Append(action byte, params []byte) {
	p.actions = append(p.actions, action)
	p.params = append(p.params, bytes.Clone(params))
}

// AppendSwapExactInSingle appends SWAP_EXACT_IN_SINGLE.
func (p *Planner) AppendSwapExactInSingle(params *ExactInputSingleParams) error {
	data, err := EncodeActionSwapExactInSingle(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SWAP_EXACT_IN_SINGLE: %w", err)
	}

	p.Append(ActionSwapExactInSingle, data)
	return nil
}

// AppendSwapExactIn appends SWAP_EXACT_IN.
func (p *Planner) AppendSwapExactIn(params *ExactInputParams) error {
	data, err := EncodeActionSwapExactIn(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SWAP_EXACT_IN: %w", err)
	}

	p.Append(ActionSwapExactIn, data)
	return nil
}

// AppendSwapExactOutSingle appends SWAP_EXACT_OUT_SINGLE.
func (p *Planner) AppendSwapExactOutSingle(params *ExactOutputSingleParams) error {
	data, err := EncodeActionSwapExactOutSingle(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SWAP_EXACT_OUT_SINGLE: %w", err)
	}

	p.Append(ActionSwapExactOutSingle, data)
	return nil
}

// AppendSwapExactOut appends SWAP_EXACT_OUT.
func (p *Planner) AppendSwapExactOut(params *ExactOutputParams) error {
	data, err := EncodeActionSwapExactOut(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SWAP_EXACT_OUT: %w", err)
	}

	p.Append(ActionSwapExactOut, data)
	return nil
}

// AppendSettle appends SETTLE.
func (p *Planner) AppendSettle(currency Currency, amount *big.Int, payerIsUser bool) error {
	data, err := EncodeActionSettle(currency, amount, payerIsUser)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SETTLE: %w", err)
	}

	p.Append(ActionSettle, data)
	return nil
}

// AppendSettleAll appends SETTLE_ALL.
func (p *Planner) AppendSettleAll(currency Currency, maxAmount *big.Int) error {
	data, err := EncodeActionSettleAll(currency, maxAmount)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SETTLE_ALL: %w", err)
	}

	p.Append(ActionSettleAll, data)
	return nil
}

// AppendSettlePair appends SETTLE_PAIR.
func (p *Planner) AppendSettlePair(currency0, currency1 Currency) error {
	data, err := EncodeActionSettlePair(currency0, currency1)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SETTLE_PAIR: %w", err)
	}

	p.Append(ActionSettlePair, data)
	return nil
}

// AppendTake appends TAKE.
func (p *Planner) AppendTake(currency Currency, recipient *types.Address, amount *big.Int) error {
	data, err := EncodeActionTake(currency, recipient, amount)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append TAKE: %w", err)
	}

	p.Append(ActionTake, data)
	return nil
}

// AppendTakeAll appends TAKE_ALL.
func (p *Planner) AppendTakeAll(currency Currency, minAmount *big.Int) error {
	data, err := EncodeActionTakeAll(currency, minAmount)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append TAKE_ALL: %w", err)
	}

	p.Append(ActionTakeAll, data)
	return nil
}

// AppendTakePortion appends TAKE_PORTION.
func (p *Planner) AppendTakePortion(currency Currency, recipient *types.Address, bips *big.Int) error {
	data, err := EncodeActionTakePortion(currency, recipient, bips)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append TAKE_PORTION: %w", err)
	}

	p.Append(ActionTakePortion, data)
	return nil
}

// AppendTakePair appends TAKE_PAIR.
func (p *Planner) AppendTakePair(currency0, currency1, recipient Currency) error {
	data, err := EncodeActionTakePair(currency0, currency1, recipient)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append TAKE_PAIR: %w", err)
	}

	p.Append(ActionTakePair, data)
	return nil
}

// AppendCloseCurrency appends CLOSE_CURRENCY.
func (p *Planner) AppendCloseCurrency(currency Currency) error {
	data, err := EncodeActionCloseCurrency(currency)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append CLOSE_CURRENCY: %w", err)
	}

	p.Append(ActionCloseCurrency, data)
	return nil
}

// AppendClearOrTake appends CLEAR_OR_TAKE.
func (p *Planner) AppendClearOrTake(currency Currency, amountMax *big.Int) error {
	data, err := EncodeActionClearOrTake(currency, amountMax)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append CLEAR_OR_TAKE: %w", err)
	}

	p.Append(ActionClearOrTake, data)
	return nil
}

// AppendSweep appends SWEEP.
func (p *Planner) AppendSweep(currency Currency, to *types.Address) error {
	data, err := EncodeActionSweep(currency, to)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append SWEEP: %w", err)
	}

	p.Append(ActionSweep, data)
	return nil
}

// AppendWrap appends WRAP.
func (p *Planner) AppendWrap(amount *big.Int) error {
	data, err := EncodeActionWrap(amount)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append WRAP: %w", err)
	}

	p.Append(ActionWrap, data)
	return nil
}

// AppendUnwrap appends UNWRAP.
func (p *Planner) AppendUnwrap(amount *big.Int) error {
	data, err := EncodeActionUnwrap(amount)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append UNWRAP: %w", err)
	}

	p.Append(ActionUnwrap, data)
	return nil
}

// AppendMintPosition appends MINT_POSITION.
func (p *Planner) AppendMintPosition(params *MintPositionParams) error {
	data, err := EncodeActionMintPosition(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append MINT_POSITION: %w", err)
	}

	p.Append(ActionMintPosition, data)
	return nil
}

// AppendMintPositionFromDeltas appends MINT_POSITION_FROM_DELTAS.
func (p *Planner) AppendMintPositionFromDeltas(params *MintPositionFromDeltasParams) error {
	data, err := EncodeActionMintPositionFromDeltas(params)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append MINT_POSITION_FROM_DELTAS: %w", err)
	}

	p.Append(ActionMintPositionFromDeltas, data)
	return nil
}

// AppendIncreaseLiquidity appends INCREASE_LIQUIDITY.
func (p *Planner) AppendIncreaseLiquidity(tokenID, liquidity, amount0Max, amount1Max *big.Int, hookData []byte) error {
	data, err := EncodeActionIncreaseLiquidity(tokenID, liquidity, amount0Max, amount1Max, hookData)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append INCREASE_LIQUIDITY: %w", err)
	}

	p.Append(ActionIncreaseLiquidity, data)
	return nil
}

// AppendIncreaseLiquidityFromDeltas appends INCREASE_LIQUIDITY_FROM_DELTAS.
func (p *Planner) AppendIncreaseLiquidityFromDeltas(tokenID, amount0Max, amount1Max *big.Int, hookData []byte) error {
	data, err := EncodeActionIncreaseLiquidityFromDeltas(tokenID, amount0Max, amount1Max, hookData)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append INCREASE_LIQUIDITY_FROM_DELTAS: %w", err)
	}

	p.Append(ActionIncreaseLiquidityFromDeltas, data)
	return nil
}

// AppendDecreaseLiquidity appends DECREASE_LIQUIDITY.
func (p *Planner) AppendDecreaseLiquidity(tokenID, liquidity, amount0Min, amount1Min *big.Int, hookData []byte) error {
	data, err := EncodeActionDecreaseLiquidity(tokenID, liquidity, amount0Min, amount1Min, hookData)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append DECREASE_LIQUIDITY: %w", err)
	}

	p.Append(ActionDecreaseLiquidity, data)
	return nil
}

// AppendBurnPosition appends BURN_POSITION.
func (p *Planner) AppendBurnPosition(tokenID, amount0Min, amount1Min *big.Int, hookData []byte) error {
	data, err := EncodeActionBurnPosition(tokenID, amount0Min, amount1Min, hookData)
	if err != nil {
		return fmt.Errorf("uniswapv4: planner: append BURN_POSITION: %w", err)
	}

	p.Append(ActionBurnPosition, data)
	return nil
}

// Len returns how many actions were appended.
func (p *Planner) Len() int {
	return len(p.actions)
}

// Actions returns a copy of the appended action bytes.
func (p *Planner) Actions() []byte {
	return bytes.Clone(p.actions)
}

// Params returns a copy of the appended params, one per action.
func (p *Planner) Params() [][]byte {
	params := make([][]byte, len(p.params))
	for i, param := range p.params {
		params[i] = bytes.Clone(param)
	}
	return params
}
