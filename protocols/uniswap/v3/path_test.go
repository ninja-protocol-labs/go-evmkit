package v3

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEncodePath(t *testing.T) {
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	path, err := EncodePath([]*types.Address{a, b, c}, []*big.Int{big.NewInt(500), big.NewInt(10000)})
	require.NoError(t, err)

	want := append(append(append(append(append([]byte{}, a.Bytes()...), 0x00, 0x01, 0xf4), b.Bytes()...), 0x00, 0x27, 0x10), c.Bytes()...)
	require.Equal(t, want, path)
	require.Len(t, path, 20+23*2)

	single, err := EncodePath([]*types.Address{a, b}, []*big.Int{big.NewInt(0xffffff)})
	require.NoError(t, err)
	require.Equal(t, append(append(append([]byte{}, a.Bytes()...), 0xff, 0xff, 0xff), b.Bytes()...), single)
}

func TestEncodePathRejectsInvalidInput(t *testing.T) {
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name   string
		tokens []*types.Address
		fees   []*big.Int
		target error
		want   string
	}{
		{"no tokens", nil, nil, ErrInvalidPath, "uniswapv3: encodePath: invalid path: need at least 2 tokens, got 0"},
		{"one token", []*types.Address{a}, nil, ErrInvalidPath, "uniswapv3: encodePath: invalid path: need at least 2 tokens, got 1"},
		{"fees too few", []*types.Address{a, b}, nil, ErrInvalidPath, "uniswapv3: encodePath: invalid path: fees length 0 does not match 1 hops"},
		{"fees too many", []*types.Address{a, b}, []*big.Int{big.NewInt(1), big.NewInt(2)}, ErrInvalidPath, "uniswapv3: encodePath: invalid path: fees length 2 does not match 1 hops"},
		{"nil first token", []*types.Address{nil, b}, []*big.Int{big.NewInt(1)}, ErrInvalidPath, "uniswapv3: encodePath: invalid path: tokens[0] is nil"},
		{"nil last token", []*types.Address{a, nil}, []*big.Int{big.NewInt(1)}, ErrInvalidPath, "uniswapv3: encodePath: invalid path: tokens[1] is nil"},
		{"nil fee", []*types.Address{a, b}, []*big.Int{nil}, ErrInvalidPath, "uniswapv3: encodePath: invalid path: fees[0] is nil"},
		{"fee too large", []*types.Address{a, b}, []*big.Int{big.NewInt(1 << 24)}, abi.ErrIntegerOutOfRange, "uniswapv3: encodePath: fees[0]: " + abi.ErrIntegerOutOfRange.Error() + ": 16777216 does not fit uint24"},
		{"fee negative", []*types.Address{a, b}, []*big.Int{big.NewInt(-1)}, abi.ErrIntegerOutOfRange, "uniswapv3: encodePath: fees[0]: " + abi.ErrIntegerOutOfRange.Error() + ": -1 does not fit uint24"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, err := EncodePath(tc.tokens, tc.fees)
			require.Nil(t, path)
			require.ErrorIs(t, err, tc.target)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestDecodePath(t *testing.T) {
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	c := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := append(append(append(append(append([]byte{}, a.Bytes()...), 0x00, 0x01, 0xf4), b.Bytes()...), 0x00, 0x0b, 0xb8), c.Bytes()...)

	tokens, fees, err := DecodePath(path)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprint([]*types.Address{a, b, c}), fmt.Sprint(tokens))
	require.Equal(t, fmt.Sprint([]*big.Int{big.NewInt(500), big.NewInt(3000)}), fmt.Sprint(fees))
}

func TestDecodePathRejectsMalformedLength(t *testing.T) {
	for _, n := range []int{0, 1, 20, 22, 42, 44, 65, 67} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			tokens, fees, err := DecodePath(make([]byte, n))
			require.Nil(t, tokens)
			require.Nil(t, fees)
			require.ErrorIs(t, err, ErrInvalidPath)
			require.EqualError(t, err, fmt.Sprintf("uniswapv3: decodePath: invalid path: length %d is not 20 + 23*n with n >= 1", n))
		})
	}
}

// go test -run '^$' -fuzz '^FuzzPathRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzPathRoundTrip(f *testing.F) {
	f.Add([]byte{}, uint8(1), uint32(0))
	f.Add([]byte{0x01, 0x02}, uint8(3), uint32(3000))
	f.Add(make([]byte, 200), uint8(9), uint32(1<<24-1))

	f.Fuzz(func(t *testing.T, seed []byte, hopsRaw uint8, feeRaw uint32) {
		hops := int(hopsRaw%8) + 1
		tokens := make([]*types.Address, hops+1)
		fees := make([]*big.Int, hops)
		for i := range tokens {
			raw := make([]byte, 20)
			if len(seed) > 0 {
				raw[i%20] = seed[i%len(seed)]
			}
			raw[19] = byte(i)
			tokens[i] = types.NewAddressFromBytes(raw)
		}
		for i := range fees {
			fees[i] = big.NewInt(int64((feeRaw + uint32(i)) % (1 << 24)))
		}

		path, err := EncodePath(tokens, fees)
		require.NoError(t, err)
		require.Len(t, path, 20+23*hops)

		gotTokens, gotFees, err := DecodePath(path)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprint(tokens), fmt.Sprint(gotTokens))
		require.Equal(t, fmt.Sprint(fees), fmt.Sprint(gotFees))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodePathDoesNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodePathDoesNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 43))
	f.Add(make([]byte, 44))

	f.Fuzz(func(t *testing.T, path []byte) {
		tokens, fees, err := DecodePath(path)
		if err != nil {
			require.Nil(t, tokens)
			require.Nil(t, fees)
			return
		}
		require.Len(t, tokens, len(fees)+1)
	})
}
