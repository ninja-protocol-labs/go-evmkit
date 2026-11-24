package universalrouter

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/permit2"
	v2 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v2"
	v3 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v3"
	v4 "github.com/ninja-protocol-labs/go-evmkit/protocols/uniswap/v4"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var (
	testRouter = types.MustNewAddressFromHex("0x66a9893cC07D91D95644AEDD05D03f95e1dBA8Af")
	testSender = types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
)

func TestCallExecuteSendsCalldataValueAndBlock(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())).AddAllowRevert(NewSweepInput(testTokenA, MsgSender(), big.NewInt(1)))
	commands, inputs, err := plan.Encode()
	require.NoError(t, err)
	withoutDeadline, err := EncodeExecute(commands, inputs)
	require.NoError(t, err)
	withDeadline, err := EncodeExecuteWithDeadline(commands, inputs, big.NewInt(1_700_000_000))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		deadline *big.Int
		native   *big.Int
		calldata []byte
		value    string
	}{
		{"no deadline no native", nil, nil, withoutDeadline, ""},
		{"deadline no native", big.NewInt(1_700_000_000), nil, withDeadline, ""},
		{"no deadline native", nil, big.NewInt(1_000_000), withoutDeadline, "0xf4240"},
		{"deadline native", big.NewInt(1_700_000_000), big.NewInt(1), withDeadline, "0x1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, opts := range []*bind.CallOptions{nil, bind.NewCallOptions().WithFrom(testSender).WithBlock("0x2a")} {
				c := mocks.NewClient(t)
				c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
					require.Equal(t, rpc.MethodETHCall, e.Method)
					params := e.Params.([]any)
					call := params[0].(map[string]any)
					require.Equal(t, testRouter.String(), call["to"])
					require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
					if tc.value == "" {
						require.NotContains(t, call, "value")
					} else {
						require.Equal(t, tc.value, call["value"])
					}
					if opts == nil {
						require.NotContains(t, call, "from")
						require.Equal(t, rpc.BlockTagLatest, params[1])
					} else {
						require.Equal(t, testSender.String(), call["from"])
						require.Equal(t, "0x2a", params[1])
					}
					*e.Result.(*string) = "0x"
					return nil
				}).Once()

				require.NoError(t, New(c).CallExecute(t.Context(), opts, NewExecuteRequest(testRouter, plan, tc.deadline, tc.native)))
				if opts != nil {
					require.Nil(t, opts.Value)
				}
			}
		})
	}
}

func TestCallExecuteSignedSendsCalldataAndValue(t *testing.T) {
	plan := NewPlan().Add(NewUnwrapWETHInput(MsgSender(), big.NewInt(0)))
	commands, inputs, err := plan.Encode()
	require.NoError(t, err)
	intent := [32]byte{0x01}
	data := [32]byte{0x02}
	nonce := [32]byte{0x03}
	calldata, err := EncodeExecuteSigned(commands, inputs, intent, data, true, nonce, []byte{0xaa, 0xbb}, big.NewInt(99))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		call := e.Params.([]any)[0].(map[string]any)
		require.Equal(t, testRouter.String(), call["to"])
		require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
		require.Equal(t, "0x7", call["value"])
		require.Equal(t, testSender.String(), call["from"])
		*e.Result.(*string) = "0x"
		return nil
	}).Once()

	require.NoError(t, New(c).CallExecuteSigned(t.Context(), bind.NewCallOptions().WithFrom(testSender), NewExecuteSignedRequest(testRouter, plan, intent, data, true, nonce, []byte{0xaa, 0xbb}, big.NewInt(99), big.NewInt(7))))
}

func TestReadsSendCallAndDecode(t *testing.T) {
	nonce := [32]byte{0x03}
	noncesCalldata, err := EncodeNoncesUsed(testTokenA, nonce)
	require.NoError(t, err)
	require.Equal(t, "0x1a27b2010000000000000000000000001111111111111111111111111111111111111111"+"0300000000000000000000000000000000000000000000000000000000000000", encoding.Hex.EncodePrefixed(noncesCalldata))
	contextCalldata, err := EncodeSignedRouteContext()
	require.NoError(t, err)
	require.Equal(t, "0xd04d79a2", encoding.Hex.EncodePrefixed(contextCalldata))

	trueWord, err := abi.Pack(abi.NewTypes(abi.Bool), true)
	require.NoError(t, err)
	contextWord, err := abi.Pack(abi.NewTypes(abi.Address, abi.Bytes32, abi.Bytes32), testTokenB, bytes.Repeat([]byte{0x0a}, 32), bytes.Repeat([]byte{0x0b}, 32))
	require.NoError(t, err)

	opts := bind.NewCallOptions().WithFrom(testSender).WithBlock("0x10")
	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(u *UniversalRouter) (any, error)
		want     string
	}{
		{"noncesUsed", noncesCalldata, trueWord, func(u *UniversalRouter) (any, error) {
			return u.NoncesUsed(t.Context(), opts, NewNoncesUsedRequest(testRouter, testTokenA, nonce))
		}, "true"},
		{"signedRouteContext", contextCalldata, contextWord, func(u *UniversalRouter) (any, error) {
			ctx, err := u.SignedRouteContext(t.Context(), opts, NewSignedRouteContextRequest(testRouter))
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("%s %x %x", ctx.Signer, ctx.Intent, ctx.Data), nil
		}, fmt.Sprintf("%s %x %x", testTokenB, bytes.Repeat([]byte{0x0a}, 32), bytes.Repeat([]byte{0x0b}, 32))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				params := e.Params.([]any)
				call := params[0].(map[string]any)
				require.Equal(t, testRouter.String(), call["to"])
				require.Equal(t, testSender.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x10", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, tc.want, fmt.Sprint(got))
		})
	}
}

func TestReadsRejectMalformedResult(t *testing.T) {
	for _, result := range []string{"0x", "0xzz", "0x01", "not hex", "0x" + fmt.Sprintf("%064x", 2)} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Twice()
			u := New(c)

			used, err := u.NoncesUsed(t.Context(), nil, NewNoncesUsedRequest(testRouter, testSender, [32]byte{}))
			require.False(t, used)
			require.ErrorContains(t, err, "universalrouter: noncesUsed: ")

			routeCtx, err := u.SignedRouteContext(t.Context(), nil, NewSignedRouteContextRequest(testRouter))
			require.Nil(t, routeCtx)
			require.ErrorContains(t, err, "universalrouter: signedRouteContext: ")
		})
	}
}

func TestMethodsRejectInvalidRequestAndOptionsWithoutRPC(t *testing.T) {
	u := New(mocks.NewClient(t))
	ctx := t.Context()
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	txOpts := bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder())

	err := u.CallExecute(ctx, nil, &ExecuteRequest{Plan: plan})
	require.EqualError(t, err, "universalrouter: callExecute: invalid request: contract is nil")
	err = u.CallExecute(ctx, nil, NewExecuteRequest(testRouter, nil, nil, nil))
	require.EqualError(t, err, "universalrouter: callExecute: plan: invalid request: plan is nil")
	err = u.CallExecute(ctx, nil, NewExecuteRequest(testRouter, NewPlan().Add(&SweepInput{}), nil, nil))
	require.EqualError(t, err, "universalrouter: callExecute: plan: steps[0] SWEEP: invalid request: token is nil")
	err = u.CallExecute(ctx, nil, NewExecuteRequest(testRouter, plan, nil, big.NewInt(-1)))
	require.EqualError(t, err, "universalrouter: callExecute: invalid request: nativeAmount is negative")
	err = u.CallExecute(ctx, bind.NewCallOptions().WithValue(big.NewInt(1)), NewExecuteRequest(testRouter, plan, nil, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "universalrouter: callExecute: invalid options: value must be nil")
	err = u.CallExecute(ctx, nil, NewExecuteRequest(testRouter, plan, big.NewInt(-1), nil))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	err = u.CallExecuteSigned(ctx, nil, NewExecuteSignedRequest(testRouter, plan, [32]byte{}, [32]byte{}, false, [32]byte{}, nil, big.NewInt(1), nil))
	require.EqualError(t, err, "universalrouter: callExecuteSigned: invalid request: signature is empty")
	err = u.CallExecuteSigned(ctx, nil, NewExecuteSignedRequest(testRouter, plan, [32]byte{}, [32]byte{}, false, [32]byte{}, []byte{0x01}, nil, nil))
	require.EqualError(t, err, "universalrouter: callExecuteSigned: invalid request: deadline is nil")
	err = u.CallExecuteSigned(ctx, bind.NewCallOptions().WithValue(big.NewInt(0)), NewExecuteSignedRequest(testRouter, plan, [32]byte{}, [32]byte{}, false, [32]byte{}, []byte{0x01}, big.NewInt(1), nil))
	require.EqualError(t, err, "universalrouter: callExecuteSigned: invalid options: value must be nil")

	_, err = u.NoncesUsed(ctx, nil, NewNoncesUsedRequest(testRouter, nil, [32]byte{}))
	require.EqualError(t, err, "universalrouter: noncesUsed: invalid request: user is nil")
	_, err = u.SignedRouteContext(ctx, nil, &SignedRouteContextRequest{})
	require.EqualError(t, err, "universalrouter: signedRouteContext: invalid request: contract is nil")

	_, _, err = u.Execute(ctx, txOpts, NewExecuteRequest(testRouter, NewPlan(), nil, nil))
	require.EqualError(t, err, "universalrouter: execute: plan: invalid request: plan has no steps")
	_, _, err = u.Execute(ctx, bind.NewTransactOptions(nil, tx.NewDynamicFeeBuilder()).WithValue(big.NewInt(1)), NewExecuteRequest(testRouter, plan, nil, big.NewInt(1)))
	require.EqualError(t, err, "universalrouter: execute: invalid options: value must be nil")
	_, _, err = u.Execute(ctx, nil, NewExecuteRequest(testRouter, plan, nil, nil))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.ErrorContains(t, err, "universalrouter: execute: ")

	_, _, err = u.ExecuteSigned(ctx, txOpts, NewExecuteSignedRequest(testRouter, plan, [32]byte{}, [32]byte{}, false, [32]byte{}, []byte{0x01}, big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 256)))
	require.EqualError(t, err, "universalrouter: executeSigned: invalid request: nativeAmount exceeds uint256")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())).Add(NewSweepInput(types.ZeroAddress(), MsgSender(), big.NewInt(0)))
	commands, inputs, err := plan.Encode()
	require.NoError(t, err)
	withoutDeadline, err := EncodeExecute(commands, inputs)
	require.NoError(t, err)
	withDeadline, err := EncodeExecuteWithDeadline(commands, inputs, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	signedCalldata, err := EncodeExecuteSigned(commands, inputs, [32]byte{0x01}, [32]byte{0x02}, false, UncheckedNonce(), []byte{0x05}, big.NewInt(1_700_000_000))
	require.NoError(t, err)
	native := big.NewInt(1_000_000_007)

	for _, tc := range []struct {
		name     string
		calldata []byte
		value    *big.Int
		send     func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{"execute", withoutDeadline, nil, func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.Execute(t.Context(), opts, NewExecuteRequest(testRouter, plan, nil, nil))
		}},
		{"execute native", withoutDeadline, native, func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.Execute(t.Context(), opts, NewExecuteRequest(testRouter, plan, nil, native))
		}},
		{"execute deadline native", withDeadline, native, func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.Execute(t.Context(), opts, NewExecuteRequest(testRouter, plan, big.NewInt(1_700_000_000), native))
		}},
		{"executeSigned", signedCalldata, nil, func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.ExecuteSigned(t.Context(), opts, NewExecuteSignedRequest(testRouter, plan, [32]byte{0x01}, [32]byte{0x02}, false, UncheckedNonce(), []byte{0x05}, big.NewInt(1_700_000_000), nil))
		}},
		{"executeSigned native", signedCalldata, native, func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.ExecuteSigned(t.Context(), opts, NewExecuteSignedRequest(testRouter, plan, [32]byte{0x01}, [32]byte{0x02}, false, UncheckedNonce(), []byte{0x05}, big.NewInt(1_700_000_000), native))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(6_000_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				var err error
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder)
			signed, hash, err := tc.send(New(c), opts)
			require.NoError(t, err)
			require.Nil(t, opts.Value)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(testRouter))
			if tc.value == nil {
				require.Zero(t, dyn.Value.Sign())
			} else {
				require.Zero(t, tc.value.Cmp(dyn.Value))
			}
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))

			call, err := DecodeExecuteCall(dyn.Data)
			require.NoError(t, err)
			require.Len(t, call.Plan.Steps, 2)
		})
	}
}

func TestCallExecuteMapsRevertToSentinel(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))

	for _, tc := range []struct {
		name string
		err  *abi.Error
		args []any
		raw  jsontext.Value
		want error
	}{
		{"ETHNotAccepted", errETHNotAccepted, nil, nil, ErrETHNotAccepted},
		{"TransactionDeadlinePassed", errTransactionDeadlinePassed, nil, nil, ErrTransactionDeadlinePassed},
		{"LengthMismatch", errLengthMismatch, nil, nil, ErrLengthMismatch},
		{"InvalidEthSender", errInvalidEthSender, nil, nil, ErrInvalidEthSender},
		{"ContractLocked", errContractLocked, nil, nil, ErrContractLocked},
		{"NonceAlreadyUsed", errNonceAlreadyUsed, nil, nil, ErrNonceAlreadyUsed},
		{"InvalidCommandType", errInvalidCommandType, []any{big.NewInt(0x0f)}, nil, ErrInvalidCommandType},
		{"BalanceTooLow", errBalanceTooLow, nil, nil, ErrBalanceTooLow},
		{"InsufficientToken", errInsufficientToken, nil, nil, ErrInsufficientToken},
		{"InsufficientETH", errInsufficientETH, nil, nil, ErrInsufficientETH},
		{"InvalidPortion", errInvalidPortion, nil, nil, ErrInvalidPortion},
		{"InvalidBips", errInvalidBips, nil, nil, ErrInvalidBips},
		{"FromAddressIsNotOwner", errFromAddressIsNotOwner, nil, nil, ErrFromAddressIsNotOwner},
		{"InvalidAction", errInvalidAction, []any{[]byte{0xde, 0xad, 0xbe, 0xef}}, nil, ErrInvalidAction},
		{"OnlyMintAllowed", errOnlyMintAllowed, nil, nil, ErrOnlyMintAllowed},
		{"NotAuthorizedForToken", errNotAuthorizedForToken, []any{big.NewInt(9)}, nil, ErrNotAuthorizedForToken},
		{"SafeERC20FailedOperation", errSafeERC20FailedOperation, []any{testTokenA}, nil, ErrSafeERC20FailedOperation},
		{"AddressEmptyCode", errAddressEmptyCode, []any{testTokenA}, nil, ErrAddressEmptyCode},
		{"FailedInnerCall", errFailedInnerCall, nil, nil, ErrFailedInnerCall},
		{"InvalidReserves", errInvalidReserves, nil, nil, ErrInvalidReserves},
		{"InvalidPath", errInvalidPath, nil, nil, ErrInvalidPath},
		{"SliceOutOfBounds", errSliceOutOfBounds, nil, nil, ErrSliceOutOfBounds},
		{"V2TooLittleReceived", errV2TooLittleReceived, nil, nil, ErrV2TooLittleReceived},
		{"V2TooMuchRequested", errV2TooMuchRequested, nil, nil, ErrV2TooMuchRequested},
		{"V2InvalidPath", errV2InvalidPath, nil, nil, ErrV2InvalidPath},
		{"V2TooLittleReceivedPerHop", errV2TooLittleReceivedPerHop, []any{big.NewInt(1), big.NewInt(2), big.NewInt(1)}, nil, ErrV2TooLittleReceivedPerHop},
		{"V2InvalidHopPriceLength", errV2InvalidHopPriceLength, nil, nil, ErrV2InvalidHopPriceLength},
		{"V3InvalidSwap", errV3InvalidSwap, nil, nil, ErrV3InvalidSwap},
		{"V3TooLittleReceived", errV3TooLittleReceived, nil, nil, ErrV3TooLittleReceived},
		{"V3TooMuchRequested", errV3TooMuchRequested, nil, nil, ErrV3TooMuchRequested},
		{"V3InvalidAmountOut", errV3InvalidAmountOut, nil, nil, ErrV3InvalidAmountOut},
		{"V3InvalidCaller", errV3InvalidCaller, nil, nil, ErrV3InvalidCaller},
		{"V3TooLittleReceivedPerHop", errV3TooLittleReceivedPerHop, []any{big.NewInt(0), big.NewInt(2), big.NewInt(1)}, nil, ErrV3TooLittleReceivedPerHop},
		{"V3TooMuchRequestedPerHop", errV3TooMuchRequestedPerHop, []any{big.NewInt(1), big.NewInt(2), big.NewInt(3)}, nil, ErrV3TooMuchRequestedPerHop},
		{"V3HopPriceAndPathLengthMismatch", errV3HopPriceAndPathLengthMismatch, nil, nil, ErrV3HopPriceAndPathLengthMismatch},
		{"ECDSAInvalidSignature", errECDSAInvalidSignature, nil, nil, ErrECDSAInvalidSignature},
		{"ECDSAInvalidSignatureLength", errECDSAInvalidSignatureLength, []any{big.NewInt(64)}, nil, ErrECDSAInvalidSignatureLength},
		{"ECDSAInvalidSignatureS", errECDSAInvalidSignatureS, []any{bytes.Repeat([]byte{0xff}, 32)}, nil, ErrECDSAInvalidSignatureS},
		{"Permit2 AllowanceExpired", abi.NewError("AllowanceExpired", abi.NewTypes(abi.Uint256)), []any{big.NewInt(1)}, nil, permit2.ErrAllowanceExpired},
		{"Permit2 InvalidNonce", abi.NewError("InvalidNonce", abi.NewTypes()), nil, nil, permit2.ErrInvalidNonce},
		{"v4 V4TooLittleReceived", abi.NewError("V4TooLittleReceived", abi.NewTypes(abi.Uint256, abi.Uint256)), []any{big.NewInt(2), big.NewInt(1)}, nil, v4.ErrV4TooLittleReceived},
		{"v4 PoolNotInitialized", abi.NewError("PoolNotInitialized", abi.NewTypes()), nil, nil, v4.ErrPoolNotInitialized},
		{"v4 NotApproved", abi.NewError("NotApproved", abi.NewTypes(abi.Address)), []any{testSender}, nil, v4.ErrNotApproved},
		{"Error(string)", nil, nil, jsontext.Value(`"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000c554e415554484f52495a45440000000000000000000000000000000000000000"`), ErrUnknown},
		{"unknown selector", nil, nil, jsontext.Value(`"0xdeadbeef"`), ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.raw
			if tc.err != nil {
				sel := tc.err.Selector()
				packed, err := abi.Pack(tc.err.Inputs, tc.args...)
				require.NoError(t, err)
				data = jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(sel[:], packed...)) + `"`)
			}
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(rpcErr).Once()

			err := New(c).CallExecute(t.Context(), nil, NewExecuteRequest(testRouter, plan, nil, nil))
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "universalrouter: callExecute: ")
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
			if !errors.Is(tc.want, ErrUnknown) {
				require.NotErrorIs(t, err, ErrUnknown)
			}
			require.NotErrorIs(t, err, ErrExecutionFailed)
		})
	}
}

func TestExecutionFailedUnwrapsInnerRevert(t *testing.T) {
	failedSel := errExecutionFailed.Selector()
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))

	invalidNonce := abi.NewError("InvalidNonce", abi.NewTypes()).Selector()
	poolAlreadyInitialized := abi.NewError("PoolAlreadyInitialized", abi.NewTypes()).Selector()
	balanceTooLow := errBalanceTooLow.Selector()
	insufficientToken := errInsufficientToken.Selector()
	innerFailed, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(0), insufficientToken[:])
	require.NoError(t, err)

	for _, tc := range []struct {
		name    string
		index   int64
		message []byte
		want    []error
		notWant []error
	}{
		{"permit2 InvalidNonce", 0, invalidNonce[:], []error{ErrExecutionFailed, permit2.ErrInvalidNonce}, []error{ErrUnknown}},
		{"v4 PoolAlreadyInitialized", 3, poolAlreadyInitialized[:], []error{ErrExecutionFailed, v4.ErrPoolAlreadyInitialized}, []error{ErrUnknown}},
		{"BalanceTooLow", 1, balanceTooLow[:], []error{ErrExecutionFailed, ErrBalanceTooLow}, []error{ErrUnknown}},
		{"nested ExecutionFailed", 2, append(failedSel[:], innerFailed...), []error{ErrExecutionFailed, ErrInsufficientToken}, []error{ErrUnknown}},
		{"unknown inner", 4, []byte{0xde, 0xad, 0xbe, 0xef}, []error{ErrExecutionFailed, ErrUnknown}, nil},
		{"empty message", 5, []byte{}, []error{ErrExecutionFailed}, []error{ErrUnknown, ErrBalanceTooLow}},
		{"short message", 6, []byte{0x01, 0x02}, []error{ErrExecutionFailed}, []error{ErrUnknown}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(tc.index), tc.message)
			require.NoError(t, err)
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(failedSel[:], packed...)) + `"`)}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(rpcErr).Once()

			err = New(c).CallExecute(t.Context(), nil, NewExecuteRequest(testRouter, plan, nil, nil))
			for _, w := range tc.want {
				require.ErrorIs(t, err, w)
			}
			for _, nw := range tc.notWant {
				require.NotErrorIs(t, err, nw)
			}

			failed, ok := errors.AsType[*ExecutionFailedError](err)
			require.True(t, ok)
			require.Zero(t, big.NewInt(tc.index).Cmp(failed.CommandIndex))
			require.Equal(t, tc.message, failed.Message)
			require.ErrorContains(t, err, fmt.Sprintf("universalrouter: callExecute: execution failed: command %d: ", tc.index))

			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
		})
	}

	nestedPacked, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(0), insufficientToken[:])
	require.NoError(t, err)
	outerPacked, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(2), append(failedSel[:], nestedPacked...))
	require.NoError(t, err)
	wrapped := MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(failedSel[:], outerPacked...)) + `"`)})
	outer, ok := errors.AsType[*ExecutionFailedError](wrapped)
	require.True(t, ok)
	require.Zero(t, big.NewInt(2).Cmp(outer.CommandIndex))
	require.Contains(t, wrapped.Error(), "execution failed: command 2: execution failed: command 0: insufficient token")

	for _, data := range [][]byte{append(failedSel[:], 0x01), failedSel[:]} {
		bad := MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(data) + `"`)})
		require.ErrorIs(t, bad, ErrExecutionFailed)
		_, ok := errors.AsType[*ExecutionFailedError](bad)
		require.False(t, ok)
	}
}

func TestWrapRevertErrorMapsV3AndV2Reasons(t *testing.T) {
	failedSel := errExecutionFailed.Selector()
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))

	for _, tc := range []struct {
		reason string
		want   error
	}{
		{"LOK", v3.ErrLocked},
		{"SPL", v3.ErrInvalidSqrtPriceLimit},
		{"UniswapV2: K", v2.ErrK},
		{"UniswapV2: LOCKED", v2.ErrLocked},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			reason := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

			direct := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(reason) + `"`)}
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(direct).Once()
			err = New(c).CallExecute(t.Context(), nil, NewExecuteRequest(testRouter, plan, nil, nil))
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, ErrUnknown)
			require.NotErrorIs(t, err, ErrExecutionFailed)
			got, ok := errors.AsType[*rpc.ResponseError](err)
			require.True(t, ok)
			require.Same(t, direct, got)

			failed, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(1), reason)
			require.NoError(t, err)
			wrapped := MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(append(failedSel[:], failed...)) + `"`)})
			require.ErrorIs(t, wrapped, ErrExecutionFailed)
			require.ErrorIs(t, wrapped, tc.want)
			require.NotErrorIs(t, wrapped, ErrUnknown)
			execFailed, ok := errors.AsType[*ExecutionFailedError](wrapped)
			require.True(t, ok)
			require.Equal(t, reason, execFailed.Message)
		})
	}

	for _, data := range []string{
		"0x08c379a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000046e6f706500000000000000000000000000000000000000000000000000000000",
		"0x4e487b710000000000000000000000000000000000000000000000000000000000000011",
	} {
		mapped := MapRevertError(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + data + `"`)})
		require.ErrorIs(t, mapped, ErrUnknown)
		require.NotErrorIs(t, mapped, v3.ErrReverted)
		require.NotErrorIs(t, mapped, v2.ErrUnknown)
	}
}

func TestMapRevertErrorCoversRouterPermit2V4V3AndV2(t *testing.T) {
	failedSel := errExecutionFailed.Selector()
	invalidNonce := abi.NewError("InvalidNonce", abi.NewTypes()).Selector()
	poolAlreadyInitialized := abi.NewError("PoolAlreadyInitialized", abi.NewTypes()).Selector()
	deadlinePassed := errTransactionDeadlinePassed.Selector()
	lok, err := abi.Pack(abi.NewTypes(abi.String), "LOK")
	require.NoError(t, err)
	k, err := abi.Pack(abi.NewTypes(abi.String), "UniswapV2: K")
	require.NoError(t, err)
	wrappedNonce, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(2), invalidNonce[:])
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		data []byte
		want []error
	}{
		{"router TransactionDeadlinePassed", deadlinePassed[:], []error{ErrTransactionDeadlinePassed}},
		{"permit2 InvalidNonce", invalidNonce[:], []error{permit2.ErrInvalidNonce}},
		{"v4 PoolAlreadyInitialized", poolAlreadyInitialized[:], []error{v4.ErrPoolAlreadyInitialized}},
		{"v3 LOK", append([]byte{0x08, 0xc3, 0x79, 0xa0}, lok...), []error{v3.ErrLocked}},
		{"v2 K", append([]byte{0x08, 0xc3, 0x79, 0xa0}, k...), []error{v2.ErrK}},
		{"ExecutionFailed wrapping InvalidNonce", append(failedSel[:], wrappedNonce...), []error{ErrExecutionFailed, permit2.ErrInvalidNonce}},
		{"unknown", []byte{0xde, 0xad, 0xbe, 0xef}, []error{ErrUnknown}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(tc.data) + `"`)}

			mapped := MapRevertError(rpcErr)
			for _, w := range tc.want {
				require.ErrorIs(t, mapped, w)
			}
			got, ok := errors.AsType[*rpc.ResponseError](mapped)
			require.True(t, ok)
			require.Same(t, rpcErr, got)
		})
	}

	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
	empty := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	require.Same(t, empty, MapRevertError(empty))
	require.NoError(t, MapRevertError(nil))
	var success *rpc.ResponseError
	require.NoError(t, MapRevertError(success))
}

func TestWrapRevertErrorPassesThroughUndecodable(t *testing.T) {
	plain := errors.New("dial tcp: connection refused")
	require.Same(t, plain, MapRevertError(plain))
	require.Nil(t, MapRevertError(nil))

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`"0x123456"`), jsontext.Value(`"0xzz"`), jsontext.Value(`{"x":1}`), jsontext.Value(`42`)} {
		t.Run(string(data), func(t *testing.T) {
			rpcErr := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}
			require.Same(t, rpcErr, MapRevertError(rpcErr))

			wrapped := fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, rpcErr)
			require.Same(t, wrapped, MapRevertError(wrapped))
		})
	}
}

func TestRevertSelectorsAreUniqueAndMatchSolidity(t *testing.T) {
	seen := map[abi.Selector]string{}
	for _, e := range []*abi.Error{
		errExecutionFailed, errETHNotAccepted, errTransactionDeadlinePassed, errLengthMismatch, errInvalidEthSender,
		errContractLocked, errNonceAlreadyUsed, errInvalidCommandType, errBalanceTooLow, errInsufficientToken,
		errInsufficientETH, errInvalidPortion, errInvalidBips, errFromAddressIsNotOwner, errInvalidAction,
		errOnlyMintAllowed, errNotAuthorizedForToken, errSafeERC20FailedOperation, errAddressEmptyCode, errFailedInnerCall, errInvalidReserves, errInvalidPath,
		errSliceOutOfBounds, errV2TooLittleReceived, errV2TooMuchRequested, errV2InvalidPath, errV2TooLittleReceivedPerHop,
		errV2InvalidHopPriceLength, errV3InvalidSwap, errV3TooLittleReceived, errV3TooMuchRequested, errV3InvalidAmountOut,
		errV3InvalidCaller, errV3TooLittleReceivedPerHop, errV3TooMuchRequestedPerHop, errV3HopPriceAndPathLengthMismatch,
		errECDSAInvalidSignature, errECDSAInvalidSignatureLength, errECDSAInvalidSignatureS,
	} {
		prev, dup := seen[e.Selector()]
		require.False(t, dup, "%s collides with %s", e.Signature(), prev)
		seen[e.Selector()] = e.Signature()
	}

	for _, tc := range []struct {
		err *abi.Error
		sig string
		sel string
	}{
		{errExecutionFailed, "ExecutionFailed(uint256,bytes)", "0x2c4029e9"},
		{errTransactionDeadlinePassed, "TransactionDeadlinePassed()", "0x5bf6f916"},
		{errV3TooLittleReceived, "V3TooLittleReceived()", "0x39d35496"},
		{errV2TooLittleReceivedPerHop, "V2TooLittleReceivedPerHop(uint256,uint256,uint256)", "0x65d564a5"},
		{errInvalidAction, "InvalidAction(bytes4)", "0xf801e525"},
		{errECDSAInvalidSignatureLength, "ECDSAInvalidSignatureLength(uint256)", "0xfce698f7"},
		{errBalanceTooLow, "BalanceTooLow()", "0xa3281672"},
		{errInvalidCommandType, "InvalidCommandType(uint256)", "0xd76a1e9e"},
		{errAddressEmptyCode, "AddressEmptyCode(address)", "0x9996b315"},
		{errFailedInnerCall, "FailedInnerCall()", "0x1425ea42"},
	} {
		require.Equal(t, tc.sig, tc.err.Signature())
		sel := tc.err.Selector()
		require.Equal(t, tc.sel, encoding.Hex.EncodePrefixed(sel[:]))
	}

	require.Equal(t, "noncesUsed(address,bytes32)", noncesUsedFn.Signature())
	require.Equal(t, "signedRouteContext()", signedRouteContextFn.Signature())
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	failedSel := errExecutionFailed.Selector()
	balanceTooLow := errBalanceTooLow.Selector()
	failedArgs, err := abi.Pack(errExecutionFailed.Inputs, big.NewInt(1), balanceTooLow[:])
	require.NoError(t, err)
	deadlineSel := errTransactionDeadlinePassed.Selector()
	nonceSel := errNonceAlreadyUsed.Selector()

	for _, tc := range []struct {
		name string
		data []byte
		send func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
		want error
	}{
		{"execute deadline", deadlineSel[:], func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.Execute(t.Context(), opts, NewExecuteRequest(testRouter, plan, big.NewInt(1), nil))
		}, ErrTransactionDeadlinePassed},
		{"execute command failed", append(failedSel[:], failedArgs...), func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.Execute(t.Context(), opts, NewExecuteRequest(testRouter, plan, nil, big.NewInt(1)))
		}, ErrBalanceTooLow},
		{"executeSigned nonce used", nonceSel[:], func(u *UniversalRouter, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
			return u.ExecuteSigned(t.Context(), opts, NewExecuteSignedRequest(testRouter, plan, [32]byte{}, [32]byte{}, true, [32]byte{0x01}, []byte{0x01}, big.NewInt(1), nil))
		}, ErrNonceAlreadyUsed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			c := mocks.NewClient(t)
			c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
				require.Equal(t, 1, elems.Len())
				require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
				return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(tc.data) + `"`)})
			}).Once()

			signed, hash, err := tc.send(New(c), bind.NewTransactOptions(key, builder))
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, ErrUnknown)
			require.Nil(t, signed)
			require.Nil(t, hash)
		})
	}
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(6_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := New(c).Execute(t.Context(), bind.NewTransactOptions(key, builder), NewExecuteRequest(testRouter, NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance())), big.NewInt(1), big.NewInt(5)))
	require.ErrorContains(t, err, "universalrouter: execute: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
	require.Zero(t, big.NewInt(5).Cmp(signed.(*core.DynamicFeeTx).Value))
}

func TestReadsPassThroughTransportError(t *testing.T) {
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	used, err := New(c).NoncesUsed(t.Context(), nil, NewNoncesUsedRequest(testRouter, testSender, [32]byte{}))
	require.False(t, used)
	require.EqualError(t, err, "universalrouter: noncesUsed: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrUnknown)
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "universalrouter")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "universalrouter", got.Value(marker{}))
		*e.Result.(*string) = "0x"
		return nil
	}).Twice()

	var u IUniversalRouter = New(c)
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	require.NoError(t, u.CallExecute(ctx, nil, NewExecuteRequest(testRouter, plan, nil, nil)))

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = u.Execute(ctx, bind.NewTransactOptions(key, builder), NewExecuteRequest(testRouter, plan, big.NewInt(1), nil))
	require.NoError(t, err)
}
