package feeds

import (
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// DecimalsRequest is decimals()'s input on Contract.
type DecimalsRequest struct {
	Contract *types.Address
}

func NewDecimalsRequest(contract *types.Address) *DecimalsRequest {
	return &DecimalsRequest{
		Contract: contract,
	}
}

func (r *DecimalsRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// DescriptionRequest is description()'s input on Contract.
type DescriptionRequest struct {
	Contract *types.Address
}

func NewDescriptionRequest(contract *types.Address) *DescriptionRequest {
	return &DescriptionRequest{
		Contract: contract,
	}
}

func (r *DescriptionRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// VersionRequest is version()'s input on Contract.
type VersionRequest struct {
	Contract *types.Address
}

func NewVersionRequest(contract *types.Address) *VersionRequest {
	return &VersionRequest{
		Contract: contract,
	}
}

func (r *VersionRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// GetRoundDataRequest is getRoundData(roundId)'s input on Contract; on a proxy RoundID is phaseId<<64 | aggregatorRoundId.
type GetRoundDataRequest struct {
	Contract *types.Address
	RoundID  *big.Int
}

func NewGetRoundDataRequest(contract *types.Address, roundID *big.Int) *GetRoundDataRequest {
	return &GetRoundDataRequest{
		Contract: contract,
		RoundID:  roundID,
	}
}

func (r *GetRoundDataRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	if r.RoundID == nil {
		return fmt.Errorf("%w: roundId is nil", bind.ErrInvalidRequest)
	}

	if r.RoundID.Sign() < 0 {
		return fmt.Errorf("%w: roundId is negative", bind.ErrInvalidRequest)
	}

	if r.RoundID.BitLen() > 80 {
		return fmt.Errorf("%w: roundId exceeds uint80", bind.ErrInvalidRequest)
	}

	return nil
}

// LatestRoundDataRequest is latestRoundData()'s input on Contract.
type LatestRoundDataRequest struct {
	Contract *types.Address
}

func NewLatestRoundDataRequest(contract *types.Address) *LatestRoundDataRequest {
	return &LatestRoundDataRequest{
		Contract: contract,
	}
}

func (r *LatestRoundDataRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("%w: request is nil", bind.ErrInvalidRequest)
	}

	if r.Contract == nil {
		return fmt.Errorf("%w: contract is nil", bind.ErrInvalidRequest)
	}

	return nil
}

// RoundData mirrors getRoundData/latestRoundData's returns; Answer is scaled by decimals(), and AnsweredInRound is deprecated upstream.
type RoundData struct {
	RoundID         *big.Int
	Answer          *big.Int
	StartedAt       *big.Int
	UpdatedAt       *big.Int
	AnsweredInRound *big.Int
}

func NewRoundData(roundID, answer, startedAt, updatedAt, answeredInRound *big.Int) *RoundData {
	return &RoundData{
		RoundID:         roundID,
		Answer:          answer,
		StartedAt:       startedAt,
		UpdatedAt:       updatedAt,
		AnsweredInRound: answeredInRound,
	}
}
