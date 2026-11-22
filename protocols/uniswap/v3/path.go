package v3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

const (
	pathAddressLen = 20
	pathFeeLen     = 3
	pathHopLen     = pathAddressLen + pathFeeLen
)

// ErrInvalidPath reports a v3 path that isn't token | (fee | token)+.
var ErrInvalidPath = errors.New("invalid path")

// EncodePath packs a v3 swap path as tokens[0] | fees[0] | tokens[1] | ... | tokens[n]; fees needs one entry per hop.
func EncodePath(tokens []*types.Address, fees []*big.Int) ([]byte, error) {
	if len(tokens) < 2 {
		return nil, fmt.Errorf("uniswapv3: encodePath: %w: need at least 2 tokens, got %d", ErrInvalidPath, len(tokens))
	}

	if len(fees) != len(tokens)-1 {
		return nil, fmt.Errorf("uniswapv3: encodePath: %w: fees length %d does not match %d hops", ErrInvalidPath, len(fees), len(tokens)-1)
	}

	path := make([]byte, 0, pathAddressLen+len(fees)*pathHopLen)
	for i, token := range tokens {
		if token == nil {
			return nil, fmt.Errorf("uniswapv3: encodePath: %w: tokens[%d] is nil", ErrInvalidPath, i)
		}

		if i > 0 {
			fee := fees[i-1]
			if fee == nil {
				return nil, fmt.Errorf("uniswapv3: encodePath: %w: fees[%d] is nil", ErrInvalidPath, i-1)
			}
			if fee.Sign() < 0 || fee.BitLen() > 24 {
				return nil, fmt.Errorf("uniswapv3: encodePath: fees[%d]: %w: %s does not fit uint24", i-1, abi.ErrIntegerOutOfRange, fee)
			}
			var buf [pathFeeLen]byte
			fee.FillBytes(buf[:])
			path = append(path, buf[:]...)
		}

		path = append(path, token.Bytes()...)
	}

	return path, nil
}

// DecodePath splits a v3 swap path back into its tokens and per-hop fees.
func DecodePath(path []byte) (tokens []*types.Address, fees []*big.Int, err error) {
	if len(path) < pathAddressLen+pathHopLen || (len(path)-pathAddressLen)%pathHopLen != 0 {
		return nil, nil, fmt.Errorf("uniswapv3: decodePath: %w: length %d is not 20 + 23*n with n >= 1", ErrInvalidPath, len(path))
	}

	hops := (len(path) - pathAddressLen) / pathHopLen
	tokens = make([]*types.Address, 0, hops+1)
	fees = make([]*big.Int, 0, hops)

	tokens = append(tokens, types.NewAddressFromBytes(path[:pathAddressLen]))
	for i := range hops {
		off := pathAddressLen + i*pathHopLen
		fees = append(fees, new(big.Int).SetBytes(path[off:off+pathFeeLen]))
		tokens = append(tokens, types.NewAddressFromBytes(path[off+pathFeeLen:off+pathHopLen]))
	}

	return tokens, fees, nil
}
