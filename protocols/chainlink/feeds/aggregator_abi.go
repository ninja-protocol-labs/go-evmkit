package feeds

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// Function signatures from AggregatorV3Interface.
var (
	decimalsFn        = abi.NewFunction("decimals", nil, abi.NewTypes(abi.Uint8))
	descriptionFn     = abi.NewFunction("description", nil, abi.NewTypes(abi.String))
	versionFn         = abi.NewFunction("version", nil, abi.NewTypes(abi.Uint256))
	getRoundDataFn    = abi.NewFunction("getRoundData", abi.NewTypes(abi.Uint80), abi.NewTypes(abi.Uint80, abi.Int256, abi.Uint256, abi.Uint256, abi.Uint80))
	latestRoundDataFn = abi.NewFunction("latestRoundData", nil, abi.NewTypes(abi.Uint80, abi.Int256, abi.Uint256, abi.Uint256, abi.Uint80))
)

// reasonNoDataPresent is the revert reason OCR1 aggregators give for a round they never stored.
const reasonNoDataPresent = "No data present"

// ErrNoDataPresent marks a round with no data: OCR1 aggregators revert with "No data present", OCR2 aggregators return a zero updatedAt.
var ErrNoDataPresent = errors.New("no data present")

// ErrReverted catches every other aggregator or proxy revert, including the empty revert a proxy gives for an unknown phase.
var ErrReverted = errors.New("reverted")

// EncodeDecimals returns the calldata for decimals().
func EncodeDecimals() []byte {
	return decimalsFn.SelectorBytes()
}

// DecodeDecimals decodes the return data of decimals.
func DecodeDecimals(data []byte) (uint8, error) {
	n, err := decimalsFn.DecodeSingleReturn[uint8](data)
	if err != nil {
		return 0, fmt.Errorf("chainlinkfeeds: decodeDecimals: %w", err)
	}
	return n, nil
}

// EncodeDescription returns the calldata for description().
func EncodeDescription() []byte {
	return descriptionFn.SelectorBytes()
}

// DecodeDescription decodes the return data of description.
func DecodeDescription(data []byte) (string, error) {
	s, err := descriptionFn.DecodeSingleReturn[string](data)
	if err != nil {
		return "", fmt.Errorf("chainlinkfeeds: decodeDescription: %w", err)
	}
	return s, nil
}

// EncodeVersion returns the calldata for version().
func EncodeVersion() []byte {
	return versionFn.SelectorBytes()
}

// DecodeVersion decodes the return data of version.
func DecodeVersion(data []byte) (*big.Int, error) {
	n, err := versionFn.DecodeSingleReturn[*big.Int](data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: decodeVersion: %w", err)
	}
	return n, nil
}

// EncodeGetRoundData returns the calldata for getRoundData(roundId).
func EncodeGetRoundData(roundID *big.Int) ([]byte, error) {
	data, err := getRoundDataFn.EncodeCall(roundID)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: encodeGetRoundData: %w", err)
	}
	return data, nil
}

// DecodeGetRoundData decodes the return data of getRoundData.
func DecodeGetRoundData(data []byte) (*RoundData, error) {
	vals, err := getRoundDataFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: %w", err)
	}
	if len(vals) != 5 {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: expected 5 return values, got %d", len(vals))
	}
	roundID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: unexpected decoded type for roundId: %T", vals[0])
	}
	answer, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: unexpected decoded type for answer: %T", vals[1])
	}
	startedAt, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: unexpected decoded type for startedAt: %T", vals[2])
	}
	updatedAt, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: unexpected decoded type for updatedAt: %T", vals[3])
	}
	answeredInRound, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeGetRoundData: unexpected decoded type for answeredInRound: %T", vals[4])
	}
	return NewRoundData(roundID, answer, startedAt, updatedAt, answeredInRound), nil
}

// EncodeLatestRoundData returns the calldata for latestRoundData().
func EncodeLatestRoundData() []byte {
	return latestRoundDataFn.SelectorBytes()
}

// DecodeLatestRoundData decodes the return data of latestRoundData.
func DecodeLatestRoundData(data []byte) (*RoundData, error) {
	vals, err := latestRoundDataFn.DecodeReturn(data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: %w", err)
	}
	if len(vals) != 5 {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: expected 5 return values, got %d", len(vals))
	}
	roundID, ok := vals[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: unexpected decoded type for roundId: %T", vals[0])
	}
	answer, ok := vals[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: unexpected decoded type for answer: %T", vals[1])
	}
	startedAt, ok := vals[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: unexpected decoded type for startedAt: %T", vals[2])
	}
	updatedAt, ok := vals[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: unexpected decoded type for updatedAt: %T", vals[3])
	}
	answeredInRound, ok := vals[4].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("chainlinkfeeds: decodeLatestRoundData: unexpected decoded type for answeredInRound: %T", vals[4])
	}
	return NewRoundData(roundID, answer, startedAt, updatedAt, answeredInRound), nil
}

// MapRevertError maps an aggregator revert in err to ErrNoDataPresent or ErrReverted; non-RPC errors pass through unchanged.
func MapRevertError(err error) error {
	if rpcErr, ok := errors.AsType[*rpc.ResponseError](err); ok && rpcErr == nil {
		return nil
	}

	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) {
		return err
	}
	if len(rpcErr.Data) == 0 {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	reason, decErr := abi.DecodeRevert(data)
	if decErr != nil {
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}

	switch reason {
	case reasonNoDataPresent:
		return fmt.Errorf("%w: %w", ErrNoDataPresent, err)
	default:
		return fmt.Errorf("%w: %w", ErrReverted, err)
	}
}
