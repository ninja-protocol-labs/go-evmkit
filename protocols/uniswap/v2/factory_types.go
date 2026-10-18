package v2

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// PairCreated mirrors IUniswapV2Factory's PairCreated event: a new Pair was deployed for
// (Token0, Token1) via createPair, now reachable at Pair; AllPairsLength is allPairs.length
// right after the push, i.e. this pair's 1-indexed position.
type PairCreated struct {
	Contract       *types.Address
	Token0         *types.Address
	Token1         *types.Address
	Pair           *types.Address
	AllPairsLength *big.Int
}

func NewPairCreated(contract, token0, token1, pair *types.Address, allPairsLength *big.Int) *PairCreated {
	return &PairCreated{
		Contract:       contract,
		Token0:         token0,
		Token1:         token1,
		Pair:           pair,
		AllPairsLength: allPairsLength,
	}
}
