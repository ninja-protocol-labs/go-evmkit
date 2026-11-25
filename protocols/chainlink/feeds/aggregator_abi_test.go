package feeds

import (
	"encoding/json/jsontext"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestEncodeDecimals(t *testing.T) {
	require.Equal(t, []byte{0x31, 0x3c, 0xe5, 0x67}, EncodeDecimals())
}

func TestDecodeDecimals(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint8}, uint8(8))
	require.NoError(t, err)

	got, err := DecodeDecimals(data)
	require.NoError(t, err)
	require.Equal(t, uint8(8), got)
}

func TestDecodeDecimalsRejectsGarbage(t *testing.T) {
	_, err := DecodeDecimals([]byte{0x01, 0x02})
	require.ErrorContains(t, err, "chainlinkfeeds: decodeDecimals: ")
}

func TestEncodeDescription(t *testing.T) {
	require.Equal(t, []byte{0x72, 0x84, 0xe4, 0x16}, EncodeDescription())
}

func TestDecodeDescription(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.String}, "ETH / USD")
	require.NoError(t, err)

	got, err := DecodeDescription(data)
	require.NoError(t, err)
	require.Equal(t, "ETH / USD", got)
}

func TestDecodeDescriptionRejectsGarbage(t *testing.T) {
	_, err := DecodeDescription([]byte{0x01, 0x02})
	require.ErrorContains(t, err, "chainlinkfeeds: decodeDescription: ")
}

func TestEncodeVersion(t *testing.T) {
	require.Equal(t, []byte{0x54, 0xfd, 0x4d, 0x50}, EncodeVersion())
}

func TestDecodeVersion(t *testing.T) {
	data, err := abi.Pack(abi.Types{abi.Uint256}, big.NewInt(6))
	require.NoError(t, err)

	got, err := DecodeVersion(data)
	require.NoError(t, err)
	require.Zero(t, big.NewInt(6).Cmp(got))
}

func TestDecodeVersionRejectsGarbage(t *testing.T) {
	_, err := DecodeVersion([]byte{0x01, 0x02})
	require.ErrorContains(t, err, "chainlinkfeeds: decodeVersion: ")
}

func TestEncodeGetRoundData(t *testing.T) {
	roundID := new(big.Int).Or(new(big.Int).Lsh(big.NewInt(7), 64), big.NewInt(36348))
	data, err := EncodeGetRoundData(roundID)
	require.NoError(t, err)
	require.Equal(t, []byte{0x9a, 0x6f, 0xc8, 0xf5}, data[:4])
	require.Len(t, data, 4+32)

	vals, err := getRoundDataFn.DecodeCall(data)
	require.NoError(t, err)
	require.Zero(t, roundID.Cmp(vals[0].(*big.Int)))
}

func TestEncodeGetRoundDataRejectsOutOfRange(t *testing.T) {
	_, err := EncodeGetRoundData(new(big.Int).Lsh(big.NewInt(1), 80))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "chainlinkfeeds: encodeGetRoundData: ")

	_, err = EncodeGetRoundData(big.NewInt(-1))
	require.Error(t, err)
}

func TestEncodeLatestRoundData(t *testing.T) {
	require.Equal(t, []byte{0xfe, 0xaf, 0x96, 0x8c}, EncodeLatestRoundData())
}

func TestDecodeRoundData(t *testing.T) {
	roundID := new(big.Int).Or(new(big.Int).Lsh(big.NewInt(7), 64), big.NewInt(36348))
	answer := big.NewInt(165_327_740_000)
	startedAt := big.NewInt(1_795_586_656)
	updatedAt := big.NewInt(1_795_586_675)
	data, err := latestRoundDataFn.EncodeReturn(roundID, answer, startedAt, updatedAt, roundID)
	require.NoError(t, err)
	want := NewRoundData(roundID, answer, startedAt, updatedAt, roundID)

	latest, err := DecodeLatestRoundData(data)
	require.NoError(t, err)
	require.Equal(t, want, latest)

	round, err := DecodeGetRoundData(data)
	require.NoError(t, err)
	require.Equal(t, want, round)
}

func TestDecodeRoundDataKeepsNegativeAnswer(t *testing.T) {
	answer := big.NewInt(-42)
	data, err := latestRoundDataFn.EncodeReturn(big.NewInt(1), answer, big.NewInt(2), big.NewInt(3), big.NewInt(1))
	require.NoError(t, err)

	got, err := DecodeLatestRoundData(data)
	require.NoError(t, err)
	require.Zero(t, answer.Cmp(got.Answer))
}

func TestDecodeRoundDataRejectsGarbage(t *testing.T) {
	for _, data := range [][]byte{nil, {0x01}, make([]byte, 4*32)} {
		_, err := DecodeLatestRoundData(data)
		require.ErrorContains(t, err, "chainlinkfeeds: decodeLatestRoundData: ")
		_, err = DecodeGetRoundData(data)
		require.ErrorContains(t, err, "chainlinkfeeds: decodeGetRoundData: ")
	}
}

func TestMapRevertError(t *testing.T) {
	noData, err := abi.Pack(abi.NewTypes(abi.String), reasonNoDataPresent)
	require.NoError(t, err)
	noDataRevert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, noData...)
	other, err := abi.Pack(abi.NewTypes(abi.String), "other reason")
	require.NoError(t, err)
	otherRevert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, other...)

	for _, tc := range []struct {
		name string
		data jsontext.Value
		want error
	}{
		{"no data present", jsontext.Value(`"` + encoding.Hex.EncodePrefixed(noDataRevert) + `"`), ErrNoDataPresent},
		{"other reason", jsontext.Value(`"` + encoding.Hex.EncodePrefixed(otherRevert) + `"`), ErrReverted},
		{"empty revert", jsontext.Value(`"0x"`), ErrReverted},
		{"no data", nil, ErrReverted},
		{"not a string", jsontext.Value(`{"x":1}`), ErrReverted},
		{"not hex", jsontext.Value(`"0xzz"`), ErrReverted},
		{"custom error selector", jsontext.Value(`"0xdeadbeef"`), ErrReverted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: tc.data}
			err := MapRevertError(rpcErr)
			require.ErrorIs(t, err, tc.want)
			require.ErrorIs(t, err, rpcErr)
		})
	}

	plain := errors.New("dial tcp: connection refused")
	require.Equal(t, plain, MapRevertError(plain))
	require.NoError(t, MapRevertError(nil))
	require.NoError(t, MapRevertError((*rpc.ResponseError)(nil)))
}
