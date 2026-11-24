package tx

import (
	"fmt"
	"math/big"
	"slices"
)

const (
	// defaultTipPercentile is the per-block tip percentile Pack reads when none is set.
	defaultTipPercentile = 50
	// defaultTipBlocks is how many recent blocks Pack reads tips from when none is set.
	defaultTipBlocks = 10
	// maxTipBlocks is the most blocks geth serves in one eth_feeHistory call.
	maxTipBlocks = 1024
)

// resolveTipPercentile applies the defaults to zero values and checks percentile is in [0, 100] and blocks at most 1024.
func resolveTipPercentile(percentile float64, blocks uint64) (float64, uint64, error) {
	if percentile < 0 || percentile > 100 {
		return 0, 0, fmt.Errorf("tx: tipPercentile: percentile %v is outside [0, 100]", percentile)
	}
	if blocks > maxTipBlocks {
		return 0, 0, fmt.Errorf("tx: tipPercentile: blocks %d exceeds %d", blocks, maxTipBlocks)
	}

	if percentile == 0 {
		percentile = defaultTipPercentile
	}
	if blocks == 0 {
		blocks = defaultTipBlocks
	}
	return percentile, blocks, nil
}

// parseFeeHistory returns the median of the reported per-block tips (nil when the node reports none) and the next block's base fee.
func parseFeeHistory(raw map[string]any) (*big.Int, *big.Int, error) {
	baseFees, ok := raw["baseFeePerGas"].([]any)
	if !ok || len(baseFees) == 0 {
		return nil, nil, fmt.Errorf("tx: feeHistory: chain does not report baseFeePerGas")
	}
	nextBaseFeeHex, ok := baseFees[len(baseFees)-1].(string)
	if !ok {
		return nil, nil, fmt.Errorf("tx: feeHistory: baseFeePerGas[%d]: unexpected type %T", len(baseFees)-1, baseFees[len(baseFees)-1])
	}
	nextBaseFee, err := parseQuantity(nextBaseFeeHex)
	if err != nil {
		return nil, nil, fmt.Errorf("tx: feeHistory: baseFeePerGas[%d]: %w", len(baseFees)-1, err)
	}

	rewards, _ := raw["reward"].([]any)
	if len(rewards) == 0 {
		return nil, nextBaseFee, nil
	}
	tips := make([]*big.Int, len(rewards))
	for i, reward := range rewards {
		percentiles, ok := reward.([]any)
		if !ok || len(percentiles) != 1 {
			return nil, nil, fmt.Errorf("tx: feeHistory: reward[%d]: expected one percentile, got %v", i, reward)
		}
		tipHex, ok := percentiles[0].(string)
		if !ok {
			return nil, nil, fmt.Errorf("tx: feeHistory: reward[%d][0]: unexpected type %T", i, percentiles[0])
		}
		tips[i], err = parseQuantity(tipHex)
		if err != nil {
			return nil, nil, fmt.Errorf("tx: feeHistory: reward[%d][0]: %w", i, err)
		}
	}

	slices.SortFunc(tips, func(a, b *big.Int) int { return a.Cmp(b) })
	mid := len(tips) / 2
	tip := new(big.Int).Set(tips[mid])
	if len(tips)%2 == 0 {
		tip.Add(tips[mid-1], tips[mid]).Rsh(tip, 1)
	}
	return tip, nextBaseFee, nil
}
