package deploy

import (
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
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip1014"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestDeploysSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	deployment := NewContractDeployment([]byte{0x60, 0x80, 0x60, 0x40}, abi.NewTypes(abi.String, abi.Uint8), []any{"token", uint8(18)})
	deterministic := NewDeterministicContractDeployment(deployment, [32]byte{0x01, 0x02, 0x03})

	initCode, err := deployment.InitCode()
	require.NoError(t, err)
	nicksData := append(append([]byte{}, deterministic.Salt[:]...), initCode...)
	erc2470Data, err := EncodeERC2470Deploy(deterministic)
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		value    *big.Int
		wantTo   *types.Address
		wantData []byte
		send     func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"create",
			nil,
			nil,
			initCode,
			func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return NewCreateDeployer(c).Deploy(t.Context(), opts, NewCreateRequest(deployment))
			},
		},
		{
			"create with value",
			big.NewInt(42),
			nil,
			initCode,
			func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return NewCreateDeployer(c).Deploy(t.Context(), opts, NewCreateRequest(deployment))
			},
		},
		{
			"nicks",
			nil,
			factory,
			nicksData,
			func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return NewNicksDeployer(c).Deploy(t.Context(), opts, NewCreate2Request(factory, deterministic))
			},
		},
		{
			"nicks with value",
			big.NewInt(7),
			factory,
			nicksData,
			func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return NewNicksDeployer(c).Deploy(t.Context(), opts, NewCreate2Request(factory, deterministic))
			},
		},
		{
			"erc2470",
			nil,
			factory,
			erc2470Data,
			func(c rpc.Client, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return NewERC2470Deployer(c).Deploy(t.Context(), opts, NewCreate2Request(factory, deterministic))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(3_000_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			opts := bind.NewTransactOptions(key, builder).WithValue(tc.value)
			signed, hash, err := tc.send(c, opts)
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			if tc.wantTo == nil {
				require.Nil(t, dyn.To)
			} else {
				require.True(t, dyn.To.Equal(tc.wantTo))
			}
			wantValue := big.NewInt(0)
			if tc.value != nil {
				wantValue = tc.value
			}
			require.Equal(t, wantValue.String(), dyn.Value.String())
			require.Equal(t, tc.wantData, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestDeployAddress(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, abi.NewTypes(abi.Uint8), []any{uint8(1)}), [32]byte{0x09})
	initCode, err := deployment.InitCode()
	require.NoError(t, err)
	want := eip1014.Address(factory, deployment.Salt, initCode)

	nicks, err := NewNicksDeployer(nil).DeployAddress(NewCreate2Request(factory, deployment))
	require.NoError(t, err)
	require.True(t, nicks.Equal(want))

	erc2470, err := NewERC2470Deployer(nil).DeployAddress(NewCreate2Request(factory, deployment))
	require.NoError(t, err)
	require.True(t, erc2470.Equal(want))

	other, err := NewNicksDeployer(nil).DeployAddress(NewCreate2Request(DefaultERC2470SingletonFactory, deployment))
	require.NoError(t, err)
	require.False(t, other.Equal(want))

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(12)
	builder.Limit(100_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	signed, err := builder.Build(sender, nil, big.NewInt(0), initCode).Pack(t.Context(), mocks.NewClient(t))
	require.NoError(t, err)
	require.NoError(t, signed.Sign(key))

	created, err := NewCreateDeployer(nil).DeployAddress(signed)
	require.NoError(t, err)
	wantCreate, err := core.CreateAddress(sender, 12)
	require.NoError(t, err)
	require.True(t, created.Equal(wantCreate))

	_, err = NewCreateDeployer(nil).DeployAddress(nil)
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.EqualError(t, err, "deploy: deployAddress: invalid request: transaction is nil")

	unsigned, err := builder.Build(sender, nil, big.NewInt(0), initCode).Pack(t.Context(), mocks.NewClient(t))
	require.NoError(t, err)
	_, err = NewCreateDeployer(nil).DeployAddress(unsigned)
	require.ErrorContains(t, err, "deploy: deployAddress: ")

	_, err = NewNicksDeployer(nil).DeployAddress(NewCreate2Request(nil, deployment))
	require.EqualError(t, err, "deploy: deployAddress: "+NewCreate2Request(nil, deployment).Validate().Error())
	_, err = NewERC2470Deployer(nil).DeployAddress(&Create2Request{})
	require.EqualError(t, err, "deploy: deployAddress: "+(&Create2Request{}).Validate().Error())
	_, err = NewNicksDeployer(nil).DeployAddress(NewCreate2Request(factory, NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60}, abi.NewTypes(abi.Uint256), nil), [32]byte{})))
	require.ErrorContains(t, err, "deploy: deployAddress: deploy: initCode: ")
}

func TestCallDeployReturnsSimulatedAddress(t *testing.T) {
	factory := types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, nil, nil), [32]byte{0x05})
	calldata, err := EncodeERC2470Deploy(deployment)
	require.NoError(t, err)
	deployed := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	deployedWord, err := abi.Pack(abi.NewTypes(abi.Address), deployed)
	require.NoError(t, err)
	zeroWord, err := abi.Pack(abi.NewTypes(abi.Address), types.ZeroAddress())
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		require.Equal(t, rpc.MethodETHCall, e.Method)
		params := e.Params.([]any)
		call := params[0].(map[string]any)
		require.Equal(t, factory.String(), call["to"])
		require.Equal(t, from.String(), call["from"])
		require.Equal(t, encoding.Hex.EncodePrefixed(calldata), call["data"])
		require.Equal(t, "0x1b4", params[1])
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(deployedWord)
		return nil
	}).Once()

	got, err := NewERC2470Deployer(c).CallDeploy(t.Context(), bind.NewCallOptions().WithFrom(from).WithBlock("0x1b4"), NewCreate2Request(factory, deployment))
	require.NoError(t, err)
	require.True(t, got.Equal(deployed))

	c = mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
		*e.Result.(*string) = encoding.Hex.EncodePrefixed(zeroWord)
		return nil
	}).Once()

	got, err = NewERC2470Deployer(c).CallDeploy(t.Context(), nil, NewCreate2Request(factory, deployment))
	require.Nil(t, got)
	require.ErrorIs(t, err, ErrDeploymentFailed)
	require.EqualError(t, err, "deploy: callERC2470: deployment failed")
}

func TestCallDeployRejectsMalformedResult(t *testing.T) {
	factory := types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, nil, nil), [32]byte{0x05})

	for _, result := range []string{"0x", "0xzz", "0x01", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Once()

			got, err := NewERC2470Deployer(c).CallDeploy(t.Context(), nil, NewCreate2Request(factory, deployment))
			require.Nil(t, got)
			require.ErrorContains(t, err, "deploy: callERC2470: ")
			require.NotErrorIs(t, err, ErrDeploymentFailed)
		})
	}

	revert := &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"0x"`)}
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(revert).Once()
	got, err := NewERC2470Deployer(c).CallDeploy(t.Context(), nil, NewCreate2Request(factory, deployment))
	require.Nil(t, got)
	require.ErrorIs(t, err, revert)
}

func TestDeployersRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	badArgs := NewContractDeployment([]byte{0x60}, abi.NewTypes(abi.Uint256), nil)
	ctx := t.Context()
	c := mocks.NewClient(t)

	_, _, err = NewCreateDeployer(c).Deploy(ctx, opts, &CreateRequest{})
	require.EqualError(t, err, "deploy: create: "+(&CreateRequest{}).Validate().Error())
	_, _, err = NewCreateDeployer(c).Deploy(ctx, opts, NewCreateRequest(NewContractDeployment(nil, nil, nil)))
	require.EqualError(t, err, "deploy: create: "+NewCreateRequest(NewContractDeployment(nil, nil, nil)).Validate().Error())
	_, _, err = NewCreateDeployer(c).Deploy(ctx, opts, NewCreateRequest(badArgs))
	require.ErrorContains(t, err, "deploy: create: deploy: initCode: ")

	_, _, err = NewNicksDeployer(c).Deploy(ctx, opts, NewCreate2Request(nil, NewDeterministicContractDeployment(badArgs, [32]byte{})))
	require.EqualError(t, err, "deploy: nicks: "+NewCreate2Request(nil, NewDeterministicContractDeployment(badArgs, [32]byte{})).Validate().Error())
	_, _, err = NewNicksDeployer(c).Deploy(ctx, opts, NewCreate2Request(factory, NewDeterministicContractDeployment(badArgs, [32]byte{})))
	require.ErrorContains(t, err, "deploy: nicks: deploy: initCode: ")

	_, _, err = NewERC2470Deployer(c).Deploy(ctx, opts, NewCreate2Request(factory, nil))
	require.EqualError(t, err, "deploy: erc2470: "+NewCreate2Request(factory, nil).Validate().Error())
	_, _, err = NewERC2470Deployer(c).Deploy(ctx, opts, NewCreate2Request(factory, NewDeterministicContractDeployment(badArgs, [32]byte{})))
	require.ErrorContains(t, err, "deploy: erc2470: deploy: encodeERC2470Deploy: ")

	_, err = NewERC2470Deployer(c).CallDeploy(ctx, nil, &Create2Request{Factory: factory})
	require.EqualError(t, err, "deploy: callERC2470: "+(&Create2Request{Factory: factory}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60}, nil, nil), [32]byte{})
	_, _, err = NewNicksDeployer(c).Deploy(ctx, nil, NewCreate2Request(factory, deployment))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "deploy: nicks: bind: transact: invalid options: options is nil")
	_, _, err = NewCreateDeployer(c).Deploy(ctx, &bind.TransactOptions{Builder: tx.NewDynamicFeeBuilder()}, NewCreateRequest(deployment.ContractDeployment))
	require.EqualError(t, err, "deploy: create: bind: transact: invalid options: signer is nil")
}

func TestDeployMapsEstimateRevert(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, nil, nil), [32]byte{0x01})

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	revert := &rpc.ResponseError{Code: 3, Message: "execution reverted"}
	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, revert)
	}).Once()

	signed, hash, err := NewNicksDeployer(c).Deploy(t.Context(), bind.NewTransactOptions(key, builder), NewCreate2Request(factory, deployment))
	require.ErrorIs(t, err, revert)
	require.ErrorContains(t, err, "deploy: nicks: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestDeployReturnsSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(500_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	broadcast := errors.New("connection reset")
	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(broadcast).Once()

	d := NewCreateDeployer(c)
	signed, hash, err := d.Deploy(t.Context(), bind.NewTransactOptions(key, builder), NewCreateRequest(NewContractDeployment([]byte{0x60, 0x80}, nil, nil)))
	require.ErrorIs(t, err, broadcast)
	require.ErrorContains(t, err, "deploy: create: ")
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())

	addr, err := d.DeployAddress(signed)
	require.NoError(t, err)
	want, err := core.CreateAddress(sender, 9)
	require.NoError(t, err)
	require.True(t, addr.Equal(want))
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "deploy")
	factory := types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")
	deployment := NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, nil, nil), [32]byte{0x05})
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	deployedWord, err := abi.Pack(abi.NewTypes(abi.Address), types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111"))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "deploy", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(deployedWord)
		}
		return nil
	}).Times(4)

	_, err = NewERC2470Deployer(c).CallDeploy(ctx, nil, NewCreate2Request(factory, deployment))
	require.NoError(t, err)

	for i, send := range []func(opts *bind.TransactOptions) error{
		func(opts *bind.TransactOptions) error {
			_, _, err := NewCreateDeployer(c).Deploy(ctx, opts, NewCreateRequest(deployment.ContractDeployment))
			return err
		},
		func(opts *bind.TransactOptions) error {
			_, _, err := NewNicksDeployer(c).Deploy(ctx, opts, NewCreate2Request(factory, deployment))
			return err
		},
		func(opts *bind.TransactOptions) error {
			_, _, err := NewERC2470Deployer(c).Deploy(ctx, opts, NewCreate2Request(factory, deployment))
			return err
		},
	} {
		builder := tx.NewDynamicFeeBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(uint64(i))
		builder.Limit(500_000)
		builder.Price(big.NewInt(2), big.NewInt(1))
		require.NoError(t, send(bind.NewTransactOptions(key, builder)))
	}
}
