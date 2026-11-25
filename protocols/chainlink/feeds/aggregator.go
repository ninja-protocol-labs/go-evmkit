package feeds

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
)

// IAggregatorV3 mirrors AggregatorV3Interface.
type IAggregatorV3 interface {
	Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error)
	Description(ctx context.Context, opts *bind.CallOptions, req *DescriptionRequest) (string, error)
	Version(ctx context.Context, opts *bind.CallOptions, req *VersionRequest) (*big.Int, error)
	GetRoundData(ctx context.Context, opts *bind.CallOptions, req *GetRoundDataRequest) (*RoundData, error)
	LatestRoundData(ctx context.Context, opts *bind.CallOptions, req *LatestRoundDataRequest) (*RoundData, error)
}

var _ IAggregatorV3 = (*Aggregator)(nil)

// Aggregator calls any AggregatorV3Interface contract over a Client; it holds no contract address, so one instance serves every feed.
type Aggregator struct {
	c rpc.Client
}

// NewAggregator returns an Aggregator calling feeds via cli.
func NewAggregator(cli rpc.Client) *Aggregator {
	return &Aggregator{
		c: cli,
	}
}

// Decimals returns the number of decimals req.Contract's answers are scaled by.
func (a *Aggregator) Decimals(ctx context.Context, opts *bind.CallOptions, req *DecimalsRequest) (uint8, error) {
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("chainlinkfeeds: decimals: %w", err)
	}

	data, err := bind.Call(ctx, a.c, opts, req.Contract, EncodeDecimals())
	if err != nil {
		return 0, fmt.Errorf("chainlinkfeeds: decimals: %w", MapRevertError(err))
	}

	out, err := DecodeDecimals(data)
	if err != nil {
		return 0, fmt.Errorf("chainlinkfeeds: decimals: %w", err)
	}

	return out, nil
}

// Description returns req.Contract's description, usually its pair such as "ETH / USD".
func (a *Aggregator) Description(ctx context.Context, opts *bind.CallOptions, req *DescriptionRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("chainlinkfeeds: description: %w", err)
	}

	data, err := bind.Call(ctx, a.c, opts, req.Contract, EncodeDescription())
	if err != nil {
		return "", fmt.Errorf("chainlinkfeeds: description: %w", MapRevertError(err))
	}

	out, err := DecodeDescription(data)
	if err != nil {
		return "", fmt.Errorf("chainlinkfeeds: description: %w", err)
	}

	return out, nil
}

// Version returns the version of the aggregator behind req.Contract.
func (a *Aggregator) Version(ctx context.Context, opts *bind.CallOptions, req *VersionRequest) (*big.Int, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: version: %w", err)
	}

	data, err := bind.Call(ctx, a.c, opts, req.Contract, EncodeVersion())
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: version: %w", MapRevertError(err))
	}

	out, err := DecodeVersion(data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: version: %w", err)
	}

	return out, nil
}

// GetRoundData returns round req.RoundID of req.Contract, or ErrNoDataPresent when that round holds no data.
func (a *Aggregator) GetRoundData(ctx context.Context, opts *bind.CallOptions, req *GetRoundDataRequest) (*RoundData, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: getRoundData: %w", err)
	}

	calldata, err := EncodeGetRoundData(req.RoundID)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: getRoundData: %w", err)
	}

	data, err := bind.Call(ctx, a.c, opts, req.Contract, calldata)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: getRoundData: %w", MapRevertError(err))
	}

	out, err := DecodeGetRoundData(data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: getRoundData: %w", err)
	}

	if out.UpdatedAt.Sign() == 0 {
		return nil, fmt.Errorf("chainlinkfeeds: getRoundData: %w: round %s has zero updatedAt", ErrNoDataPresent, req.RoundID)
	}

	return out, nil
}

// LatestRoundData returns req.Contract's latest round, or ErrNoDataPresent when the feed holds no data yet; staleness is left to the caller via UpdatedAt.
func (a *Aggregator) LatestRoundData(ctx context.Context, opts *bind.CallOptions, req *LatestRoundDataRequest) (*RoundData, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: latestRoundData: %w", err)
	}

	data, err := bind.Call(ctx, a.c, opts, req.Contract, EncodeLatestRoundData())
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: latestRoundData: %w", MapRevertError(err))
	}

	out, err := DecodeLatestRoundData(data)
	if err != nil {
		return nil, fmt.Errorf("chainlinkfeeds: latestRoundData: %w", err)
	}

	if out.UpdatedAt.Sign() == 0 {
		return nil, fmt.Errorf("chainlinkfeeds: latestRoundData: %w: latest round has zero updatedAt", ErrNoDataPresent)
	}

	return out, nil
}
