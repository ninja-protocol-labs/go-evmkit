package deploy

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip1014"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzCreate2RequestValidate$' -fuzztime=10s ./standards/deploy
func FuzzCreate2RequestValidate(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(0x0f))
	f.Add(uint8(0x10))
	f.Add(uint8(0x07))

	factory := types.MustNewAddressFromHex("0x4e59b44847b379578588920cA78FbF26c0B4956C")

	f.Fuzz(func(t *testing.T, mask uint8) {
		var req *Create2Request
		if mask&0x10 == 0 {
			req = &Create2Request{}
			if mask&0x01 != 0 {
				req.Factory = factory
			}
			if mask&0x02 != 0 {
				req.Deployment = &DeterministicContractDeployment{}
				if mask&0x04 != 0 {
					req.Deployment.ContractDeployment = &ContractDeployment{}
					if mask&0x08 != 0 {
						req.Deployment.Bytecode = []byte{0x60}
					}
				}
			}
		}

		want := ""
		switch {
		case req == nil:
			want = "invalid request: request is nil"
		case req.Factory == nil:
			want = "invalid request: factory is nil"
		case req.Deployment == nil || req.Deployment.ContractDeployment == nil:
			want = "invalid request: deployment is nil"
		case len(req.Deployment.Bytecode) == 0:
			want = "invalid request: deployment bytecode is empty"
		}

		err := req.Validate()
		if want == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorIs(t, err, bind.ErrInvalidRequest)
		require.EqualError(t, err, want)
	})
}

// go test -run '^$' -fuzz '^FuzzNicksDeployThroughClient$' -fuzztime=10s ./standards/deploy
func FuzzNicksDeployThroughClient(f *testing.F) {
	f.Add([]byte{0x60}, []byte{}, []byte{}, uint64(0), true)
	f.Add([]byte{0x60, 0x80, 0x60, 0x40}, []byte{0x01}, make([]byte, 20), uint64(42), false)
	f.Add(make([]byte, 300), make([]byte, 32), []byte{0xff}, ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, bytecode, saltRaw, factoryRaw []byte, nonce uint64, dynamic bool) {
		if len(bytecode) == 0 {
			bytecode = []byte{0x00}
		}
		var salt [32]byte
		copy(salt[:], saltRaw)
		factory := types.NewAddressFromBytes(factoryRaw)
		req := NewCreate2Request(factory, NewDeterministicContractDeployment(NewContractDeployment(bytecode, nil, nil), salt))

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(3_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		d := NewNicksDeployer(c)
		signed, hash, err := d.Deploy(t.Context(), bind.NewTransactOptions(key, builder), req)
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotData = v.To, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotData = v.To, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(factory))
		require.Equal(t, salt[:], gotData[:32])
		require.Equal(t, bytecode, gotData[32:])

		addr, err := d.DeployAddress(req)
		require.NoError(t, err)
		require.True(t, addr.Equal(eip1014.Address(factory, salt, bytecode)))
	})
}

// go test -run '^$' -fuzz '^FuzzCreateDeployAddress$' -fuzztime=10s ./standards/deploy
func FuzzCreateDeployAddress(f *testing.F) {
	f.Add(int64(0xb1d), uint64(0), []byte{0x60})
	f.Add(int64(1), uint64(1), []byte{0x60, 0x80})
	f.Add(int64(0x7fffffff), ^uint64(0)-1, make([]byte, 64))

	f.Fuzz(func(t *testing.T, keySeed int64, nonce uint64, bytecode []byte) {
		if keySeed <= 0 {
			keySeed = -keySeed + 1
		}
		if len(bytecode) == 0 {
			bytecode = []byte{0x00}
		}
		key, err := types.NewPrivateKeyFromBig(big.NewInt(keySeed))
		require.NoError(t, err)
		sender := core.PubkeyToAddress(key.PublicKey())

		builder := tx.NewLegacyBuilder()
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(3_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err := encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		d := NewCreateDeployer(c)
		signed, _, err := d.Deploy(t.Context(), bind.NewTransactOptions(key, builder), NewCreateRequest(NewContractDeployment(bytecode, nil, nil)))
		require.NoError(t, err)

		addr, err := d.DeployAddress(signed)
		require.NoError(t, err)
		want, err := core.CreateAddress(sender, nonce)
		require.NoError(t, err)
		require.True(t, addr.Equal(want))
	})
}

// go test -run '^$' -fuzz '^FuzzCallDeploySurvivesArbitraryResult$' -fuzztime=10s ./standards/deploy
func FuzzCallDeploySurvivesArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(append(make([]byte, 12), make([]byte, 20)...))
	f.Add(append(make([]byte, 31), 0x01))

	factory := types.MustNewAddressFromHex("0xce0042B868300000d44A59004Da54A005ffdcf9f")
	req := NewCreate2Request(factory, NewDeterministicContractDeployment(NewContractDeployment([]byte{0x60, 0x80}, nil, nil), [32]byte{}))

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Once()

		addr, err := NewERC2470Deployer(c).CallDeploy(t.Context(), nil, req)
		require.True(t, (err == nil) != (addr == nil))
		if addr != nil {
			require.False(t, addr.Equal(types.ZeroAddress()))
		}
	})
}
