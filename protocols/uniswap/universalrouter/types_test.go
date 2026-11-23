package universalrouter

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestExecuteSignedMessageHashMatchesCastGolden(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), ContractBalance()))
	domain := NewUniversalRouterDomain(big.NewInt(1), testTokenC)

	var intent, data, nonce [32]byte
	copy(intent[:], bytes.Repeat([]byte{0x01}, 32))
	copy(data[:], bytes.Repeat([]byte{0x02}, 32))
	copy(nonce[:], bytes.Repeat([]byte{0x03}, 32))

	digest, err := NewExecuteSignedMessage(plan, intent, data, types.ZeroAddress(), nonce, big.NewInt(1_700_000_000)).Hash(domain)
	require.NoError(t, err)
	require.Equal(t, "0xf9040489c7ffc55c176a0e2f7ddc84d9c97ece07b7c67c9261e5bafee7b7b95c", digest.String())

	withSender, err := NewExecuteSignedMessage(plan, intent, data, testTokenA, nonce, big.NewInt(1_700_000_000)).Hash(domain)
	require.NoError(t, err)
	require.Equal(t, "0xba464c41febd77a895c05969dc7bcf18ad89b84126f3cabcd46848d46a30b793", withSender.String())

	require.Equal(t, "0x4b503a2e339bf072a489301f3eec7abaefc520266c462796ed9267bbcc6e407c", core.Keccak256([]byte(executeSignedTypes.EncodeType(executeSignedPrimaryType))).String())
	require.Equal(t, "ExecuteSigned(bytes commands,bytes[] inputs,bytes32 intent,bytes32 data,address sender,bytes32 nonce,uint256 deadline)", executeSignedTypes.EncodeType(executeSignedPrimaryType))
}

func TestExecuteSignedMessageHashSensitiveToEveryField(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	other := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(2)))
	domain := NewUniversalRouterDomain(big.NewInt(1), testRouter)

	base, err := NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(10)).Hash(domain)
	require.NoError(t, err)

	for name, msg := range map[string]*ExecuteSignedMessage{
		"plan":     NewExecuteSignedMessage(other, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(10)),
		"intent":   NewExecuteSignedMessage(plan, [32]byte{0x09}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(10)),
		"data":     NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x09}, testSender, [32]byte{0x03}, big.NewInt(10)),
		"sender":   NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, types.ZeroAddress(), [32]byte{0x03}, big.NewInt(10)),
		"nonce":    NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, UncheckedNonce(), big.NewInt(10)),
		"deadline": NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(11)),
	} {
		got, err := msg.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, base.String(), got.String(), name)
	}

	for name, d := range map[string]*struct {
		chainID  *big.Int
		contract *types.Address
	}{
		"chainID":  {big.NewInt(2), testRouter},
		"contract": {big.NewInt(1), testTokenA},
	} {
		got, err := NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(10)).Hash(NewUniversalRouterDomain(d.chainID, d.contract))
		require.NoError(t, err)
		require.NotEqual(t, base.String(), got.String(), name)
	}

	again, err := NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(10)).Hash(domain)
	require.NoError(t, err)
	require.Equal(t, base.String(), again.String())
}

func TestExecuteSignedMessageSignRecoversSigner(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xa11ce))
	require.NoError(t, err)
	signer := core.PubkeyToAddress(key.PublicKey())
	domain := NewUniversalRouterDomain(big.NewInt(11_155_111), testRouter)
	msg := NewExecuteSignedMessage(NewPlan().Add(NewSweepInput(testTokenA, MsgSender(), big.NewInt(0))), [32]byte{0x0a}, [32]byte{}, types.ZeroAddress(), UncheckedNonce(), big.NewInt(1_800_000_000))

	sig, err := msg.Sign(domain, key)
	require.NoError(t, err)
	require.Len(t, sig.LegacyBytes(), 65)

	digest, err := msg.Hash(domain)
	require.NoError(t, err)
	pub, err := sig.ECRecover(digest)
	require.NoError(t, err)
	require.True(t, core.PubkeyToAddress(pub).Equal(signer))
}

func TestExecuteSignedMessageHashRejectsInvalid(t *testing.T) {
	domain := NewUniversalRouterDomain(big.NewInt(1), testRouter)
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))

	_, err := (*ExecuteSignedMessage)(nil).Hash(domain)
	require.EqualError(t, err, "universalrouter: executeSignedMessage: hash: invalid request: message is nil")
	_, err = NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, nil, [32]byte{}, big.NewInt(1)).Hash(domain)
	require.EqualError(t, err, "universalrouter: executeSignedMessage: hash: invalid request: sender is nil")
	_, err = NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, testSender, [32]byte{}, nil).Hash(domain)
	require.EqualError(t, err, "universalrouter: executeSignedMessage: hash: invalid request: deadline is nil")
	_, err = NewExecuteSignedMessage(nil, [32]byte{}, [32]byte{}, testSender, [32]byte{}, big.NewInt(1)).Hash(domain)
	require.ErrorIs(t, err, bind.ErrInvalidRequest)
	require.ErrorContains(t, err, "plan is nil")
	_, err = NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, testSender, [32]byte{}, big.NewInt(-1)).Hash(domain)
	require.Error(t, err)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(1))
	require.NoError(t, err)
	sig, err := NewExecuteSignedMessage(plan, [32]byte{}, [32]byte{}, nil, [32]byte{}, big.NewInt(1)).Sign(domain, key)
	require.Nil(t, sig)
	require.ErrorContains(t, err, "sender is nil")
}

func TestUniversalRouterDomainAndUncheckedNonce(t *testing.T) {
	domain := NewUniversalRouterDomain(big.NewInt(8453), testRouter)
	require.Equal(t, "UniversalRouter", domain.Name)
	require.Equal(t, "2", domain.Version)
	require.Zero(t, big.NewInt(8453).Cmp(domain.ChainID))
	require.True(t, domain.VerifyingContract.Equal(testRouter))

	nonce := UncheckedNonce()
	require.Equal(t, bytes.Repeat([]byte{0xff}, 32), nonce[:])
	nonce[0] = 0x00
	again := UncheckedNonce()
	require.Equal(t, byte(0xff), again[0])
	require.Zero(t, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).Cmp(new(big.Int).SetBytes(again[:])))
}

func TestExecutionFailedErrorBehaviour(t *testing.T) {
	cause := errors.Join(ErrBalanceTooLow, errors.New("inner"))
	withCause := NewExecutionFailedError(big.NewInt(3), []byte{0xa3, 0x28, 0x16, 0x72}, cause)
	require.Equal(t, "execution failed: command 3: balance too low\ninner", withCause.Error())
	require.ErrorIs(t, withCause, ErrExecutionFailed)
	require.ErrorIs(t, withCause, ErrBalanceTooLow)
	require.NotErrorIs(t, withCause, ErrInsufficientToken)
	require.Same(t, ErrExecutionFailed, withCause.Unwrap())

	noCause := NewExecutionFailedError(big.NewInt(0), []byte{0x01, 0x02}, nil)
	require.Equal(t, "execution failed: command 0: 0x0102", noCause.Error())
	require.ErrorIs(t, noCause, ErrExecutionFailed)
	require.NotErrorIs(t, noCause, ErrBalanceTooLow)
	require.False(t, noCause.Is(ErrExecutionFailed))

	empty := NewExecutionFailedError(big.NewInt(7), nil, nil)
	require.Equal(t, "execution failed: command 7: 0x", empty.Error())

	wrapped := errors.Join(errors.New("outer"), withCause)
	got, ok := errors.AsType[*ExecutionFailedError](wrapped)
	require.True(t, ok)
	require.Same(t, withCause, got)
}

func TestClientRequestsValidate(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"execute nil", (*ExecuteRequest)(nil).Validate(), "invalid request: request is nil"},
		{"execute contract", (&ExecuteRequest{Plan: plan}).Validate(), "invalid request: contract is nil"},
		{"execute nil plan", (&ExecuteRequest{Contract: testRouter}).Validate(), "plan: invalid request: plan is nil"},
		{"execute empty plan", (&ExecuteRequest{Contract: testRouter, Plan: NewPlan()}).Validate(), "plan: invalid request: plan has no steps"},
		{"execute bad step", (&ExecuteRequest{Contract: testRouter, Plan: NewPlan().Add(&WrapETHInput{})}).Validate(), "plan: steps[0] WRAP_ETH: invalid request: recipient is nil"},
		{"execute negative native", (&ExecuteRequest{Contract: testRouter, Plan: plan, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"execute overflow native", (&ExecuteRequest{Contract: testRouter, Plan: plan, NativeAmount: overflow}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"execute valid no deadline", (&ExecuteRequest{Contract: testRouter, Plan: plan}).Validate(), ""},
		{"execute valid deadline native", (&ExecuteRequest{Contract: testRouter, Plan: plan, Deadline: big.NewInt(1), NativeAmount: new(big.Int).Sub(overflow, big.NewInt(1))}).Validate(), ""},
		{"executeSigned nil", (*ExecuteSignedRequest)(nil).Validate(), "invalid request: request is nil"},
		{"executeSigned contract", (&ExecuteSignedRequest{Plan: plan, Signature: []byte{0x01}, Deadline: big.NewInt(1)}).Validate(), "invalid request: contract is nil"},
		{"executeSigned plan", (&ExecuteSignedRequest{Contract: testRouter, Signature: []byte{0x01}, Deadline: big.NewInt(1)}).Validate(), "plan: invalid request: plan is nil"},
		{"executeSigned signature", (&ExecuteSignedRequest{Contract: testRouter, Plan: plan, Deadline: big.NewInt(1)}).Validate(), "invalid request: signature is empty"},
		{"executeSigned deadline", (&ExecuteSignedRequest{Contract: testRouter, Plan: plan, Signature: []byte{0x01}}).Validate(), "invalid request: deadline is nil"},
		{"executeSigned negative native", (&ExecuteSignedRequest{Contract: testRouter, Plan: plan, Signature: []byte{0x01}, Deadline: big.NewInt(1), NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"executeSigned valid", (&ExecuteSignedRequest{Contract: testRouter, Plan: plan, Signature: []byte{0x01}, Deadline: big.NewInt(1), Nonce: UncheckedNonce(), VerifySender: true}).Validate(), ""},
		{"noncesUsed nil", (*NoncesUsedRequest)(nil).Validate(), "invalid request: request is nil"},
		{"noncesUsed contract", (&NoncesUsedRequest{User: testSender}).Validate(), "invalid request: contract is nil"},
		{"noncesUsed user", (&NoncesUsedRequest{Contract: testRouter}).Validate(), "invalid request: user is nil"},
		{"noncesUsed valid", (&NoncesUsedRequest{Contract: testRouter, User: testSender}).Validate(), ""},
		{"signedRouteContext nil", (*SignedRouteContextRequest)(nil).Validate(), "invalid request: request is nil"},
		{"signedRouteContext contract", (&SignedRouteContextRequest{}).Validate(), "invalid request: contract is nil"},
		{"signedRouteContext valid", (&SignedRouteContextRequest{Contract: testRouter}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewClientTypesSetFields(t *testing.T) {
	plan := NewPlan().Add(NewWrapETHInput(AddressThis(), big.NewInt(1)))

	require.Equal(t, &ExecuteRequest{Contract: testRouter, Plan: plan, Deadline: big.NewInt(1), NativeAmount: big.NewInt(2)}, NewExecuteRequest(testRouter, plan, big.NewInt(1), big.NewInt(2)))
	require.Equal(t, &ExecuteSignedRequest{Contract: testRouter, Plan: plan, Intent: [32]byte{0x01}, Data: [32]byte{0x02}, VerifySender: true, Nonce: [32]byte{0x03}, Signature: []byte{0x04}, Deadline: big.NewInt(5), NativeAmount: big.NewInt(6)}, NewExecuteSignedRequest(testRouter, plan, [32]byte{0x01}, [32]byte{0x02}, true, [32]byte{0x03}, []byte{0x04}, big.NewInt(5), big.NewInt(6)))
	require.Equal(t, &NoncesUsedRequest{Contract: testRouter, User: testSender, Nonce: [32]byte{0x07}}, NewNoncesUsedRequest(testRouter, testSender, [32]byte{0x07}))
	require.Equal(t, &SignedRouteContextRequest{Contract: testRouter}, NewSignedRouteContextRequest(testRouter))
	require.Equal(t, &RouteContext{Signer: testSender, Intent: [32]byte{0x08}, Data: [32]byte{0x09}}, NewRouteContext(testSender, [32]byte{0x08}, [32]byte{0x09}))
	require.Equal(t, &ExecuteSignedMessage{Plan: plan, Intent: [32]byte{0x01}, Data: [32]byte{0x02}, Sender: testSender, Nonce: [32]byte{0x03}, Deadline: big.NewInt(4)}, NewExecuteSignedMessage(plan, [32]byte{0x01}, [32]byte{0x02}, testSender, [32]byte{0x03}, big.NewInt(4)))

	req := NewExecuteRequest(testRouter, plan, nil, nil)
	require.Same(t, plan, req.Plan)
	require.Nil(t, req.Deadline)
	require.Nil(t, req.NativeAmount)
}

func TestReadCodecsRoundTrip(t *testing.T) {
	trueWord, err := noncesUsedFn.EncodeReturn(true)
	require.NoError(t, err)
	used, err := DecodeNoncesUsed(trueWord)
	require.NoError(t, err)
	require.True(t, used)

	falseWord, err := noncesUsedFn.EncodeReturn(false)
	require.NoError(t, err)
	used, err = DecodeNoncesUsed(falseWord)
	require.NoError(t, err)
	require.False(t, used)

	_, err = DecodeNoncesUsed([]byte{0x01})
	require.ErrorContains(t, err, "universalrouter: decodeNoncesUsed: ")

	contextWord, err := signedRouteContextFn.EncodeReturn(testSender, bytes.Repeat([]byte{0xab}, 32), make([]byte, 32))
	require.NoError(t, err)
	routeCtx, err := DecodeSignedRouteContext(contextWord)
	require.NoError(t, err)
	require.True(t, routeCtx.Signer.Equal(testSender))
	require.Equal(t, bytes.Repeat([]byte{0xab}, 32), routeCtx.Intent[:])
	require.Equal(t, [32]byte{}, routeCtx.Data)

	zeroWord, err := signedRouteContextFn.EncodeReturn(types.ZeroAddress(), make([]byte, 32), make([]byte, 32))
	require.NoError(t, err)
	zero, err := DecodeSignedRouteContext(zeroWord)
	require.NoError(t, err)
	require.True(t, zero.Signer.Equal(types.ZeroAddress()))

	_, err = DecodeSignedRouteContext(make([]byte, 64))
	require.ErrorContains(t, err, "universalrouter: decodeSignedRouteContext: ")

	_, err = EncodeNoncesUsed(nil, [32]byte{})
	require.ErrorContains(t, err, "universalrouter: encodeNoncesUsed: ")
}
