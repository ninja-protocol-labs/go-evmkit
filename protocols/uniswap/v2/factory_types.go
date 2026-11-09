package v2

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// GetPairParams is GetPair's argument list.
type GetPairParams struct {
	Factory *types.Address
	TokenA  *types.Address
	TokenB  *types.Address
}

func NewGetPairParams(factory, tokenA, tokenB *types.Address) *GetPairParams {
	return &GetPairParams{
		Factory: factory,
		TokenA:  tokenA,
		TokenB:  tokenB,
	}
}

// AllPairsParams is AllPairs's argument list.
type AllPairsParams struct {
	Factory *types.Address
	Index   *big.Int
}

func NewAllPairsParams(factory *types.Address, index *big.Int) *AllPairsParams {
	return &AllPairsParams{
		Factory: factory,
		Index:   index,
	}
}

// CreatePairParams is CreateCreatePair's and CallCreatePair's argument list.
type CreatePairParams struct {
	Factory *types.Address
	Sender  *types.Address
	TokenA  *types.Address
	TokenB  *types.Address
}

func NewCreatePairParams(factory, sender, tokenA, tokenB *types.Address) *CreatePairParams {
	return &CreatePairParams{
		Factory: factory,
		Sender:  sender,
		TokenA:  tokenA,
		TokenB:  tokenB,
	}
}

// SetFeeToParams is CreateSetFeeTo's argument list.
type SetFeeToParams struct {
	Factory *types.Address
	Sender  *types.Address
	FeeTo   *types.Address
}

func NewSetFeeToParams(factory, sender, feeTo *types.Address) *SetFeeToParams {
	return &SetFeeToParams{
		Factory: factory,
		Sender:  sender,
		FeeTo:   feeTo,
	}
}

// SetFeeToSetterParams is CreateSetFeeToSetter's argument list.
type SetFeeToSetterParams struct {
	Factory     *types.Address
	Sender      *types.Address
	FeeToSetter *types.Address
}

func NewSetFeeToSetterParams(factory, sender, feeToSetter *types.Address) *SetFeeToSetterParams {
	return &SetFeeToSetterParams{
		Factory:     factory,
		Sender:      sender,
		FeeToSetter: feeToSetter,
	}
}

// PairCreated mirrors IUniswapV2Factory's PairCreated event: a new Pair was deployed for
// (Token0, Token1) via createPair, now reachable at Pair; AllPairsLength is allPairs.length
// right after the push, i.e. this pair's 1-indexed position.
type PairCreated struct {
	Factory        *types.Address
	Token0         *types.Address
	Token1         *types.Address
	Pair           *types.Address
	AllPairsLength *big.Int
}

func NewPairCreated(factory, token0, token1, pair *types.Address, allPairsLength *big.Int) *PairCreated {
	return &PairCreated{
		Factory:        factory,
		Token0:         token0,
		Token1:         token1,
		Pair:           pair,
		AllPairsLength: allPairsLength,
	}
}
