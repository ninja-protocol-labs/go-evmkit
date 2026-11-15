package v4

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// IRouter mirrors a V4Router-based deployment's own entry point: a concrete inheriting contract's
// executeActions(bytes), which forwards to BaseActionsRouter._executeActions and unlocks
// PoolManager to dispatch the encoded actions.
type IRouter interface {
	ExecuteActions(ctx context.Context, signedTx core.Transaction) (*types.Hash, error)
}

var _ IRouter = (*Router)(nil)

// Router calls any V4Router-based deployment (e.g. our own Kit router fixture) over a Client; it
// holds no contract address, so one instance serves every deployment.
type Router struct {
	c rpc.Client
}

// NewRouter returns a Router calling deployments via cli.
func NewRouter(cli rpc.Client) *Router {
	return &Router{c: cli}
}

// CreateExecuteActions builds an unsigned executeActions(abi.encode(actions, params)) transaction
// via b, sent by sender with value wei attached (for native-ETH-funded swaps).
func (r *Router) CreateExecuteActions(ctx context.Context, b tx.Builder, contract, sender *types.Address, actions []byte, params [][]byte, value *big.Int) (core.Transaction, error) {
	calldata, err := EncodeExecuteActions(actions, params)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createExecuteActions: %w", err)
	}
	unsigned, err := b.Build(sender, contract, value, calldata).Pack(ctx, r.c)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: createExecuteActions: %w", r.wrapRevertError(err))
	}
	return unsigned, nil
}

// ExecuteActions broadcasts an already-signed executeActions transaction.
func (r *Router) ExecuteActions(ctx context.Context, signedTx core.Transaction) (*types.Hash, error) {
	raw, err := signedTx.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: executeActions: %w", err)
	}
	hash, err := tx.Broadcast(ctx, r.c, raw)
	if err != nil {
		return nil, fmt.Errorf("uniswapv4: executeActions: %w", err)
	}
	return hash, nil
}

// wrapRevertError maps a known IV4Router/BaseActionsRouter custom error in err to its sentinel. A
// revert that bubbles up from the underlying PoolManager call (e.g. PoolNotInitialized) isn't
// mapped here; it falls through to ErrUnknown, since that selector space belongs to poolmanager.go.
func (_ *Router) wrapRevertError(err error) error {
	var rpcErr *rpc.ResponseError
	if !errors.As(err, &rpcErr) || len(rpcErr.Data) == 0 {
		return err
	}

	var hexData string
	if jsonErr := json.Unmarshal(rpcErr.Data, &hexData); jsonErr != nil {
		return err
	}

	data, decErr := encoding.Hex.Decode(hexData)
	if decErr != nil || len(data) < 4 {
		return err
	}

	var sel abi.Selector
	copy(sel[:], data[:4])

	switch sel {
	case errV4TooLittleReceived.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceived, err)
	case errV4TooMuchRequested.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequested, err)
	case errV4TooLittleReceivedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceivedPerHop, err)
	case errV4TooMuchRequestedPerHop.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequestedPerHop, err)
	case errV4TooLittleReceivedPerHopSingle.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooLittleReceivedPerHopSingle, err)
	case errV4TooMuchRequestedPerHopSingle.Selector():
		return fmt.Errorf("%w: %w", ErrV4TooMuchRequestedPerHopSingle, err)
	case errInvalidHopPriceLength.Selector():
		return fmt.Errorf("%w: %w", ErrInvalidHopPriceLength, err)
	case errV4ExactOutputUnfilled.Selector():
		return fmt.Errorf("%w: %w", ErrV4ExactOutputUnfilled, err)
	case errInputLengthMismatch.Selector():
		return fmt.Errorf("%w: %w", ErrInputLengthMismatch, err)
	case errUnsupportedAction.Selector():
		return fmt.Errorf("%w: %w", ErrUnsupportedAction, err)
	default:
		return fmt.Errorf("%w: %w", ErrUnknown, err)
	}
}
