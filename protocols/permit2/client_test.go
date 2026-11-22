package permit2

import (
	"context"
	"encoding/json/jsontext"
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
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestReadsSendCallAndDecode(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	opts := bind.NewCallOptions().WithFrom(owner).WithBlock("0x1b4")

	separatorWord, err := abi.Pack(abi.NewTypes(abi.Bytes32), core.Keccak256([]byte("domain")).Bytes())
	require.NoError(t, err)
	allowanceWord, err := abi.Pack(abi.NewTypes(abi.Uint160, abi.Uint48, abi.Uint48), big.NewInt(1000), big.NewInt(1_700_000_000), big.NewInt(7))
	require.NoError(t, err)
	bitmapWord, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(0b1011))
	require.NoError(t, err)

	allowanceData, err := EncodeAllowance(owner, token, spender)
	require.NoError(t, err)
	bitmapData, err := EncodeNonceBitmap(owner, big.NewInt(3))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(p *Permit2) (any, error)
		want     any
	}{
		{
			"domainSeparator",
			EncodeDomainSeparator(),
			separatorWord,
			func(p *Permit2) (any, error) {
				return p.DomainSeparator(t.Context(), opts, NewDomainSeparatorRequest(contract))
			},
			core.Keccak256([]byte("domain")),
		},
		{
			"allowance",
			allowanceData,
			allowanceWord,
			func(p *Permit2) (any, error) {
				return p.Allowance(t.Context(), opts, NewAllowanceRequest(contract, owner, token, spender))
			},
			NewPackedAllowance(big.NewInt(1000), big.NewInt(1_700_000_000), big.NewInt(7)),
		},
		{
			"nonceBitmap",
			bitmapData,
			bitmapWord,
			func(p *Permit2) (any, error) {
				return p.NonceBitmap(t.Context(), opts, NewNonceBitmapRequest(contract, owner, big.NewInt(3)))
			},
			big.NewInt(0b1011),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, contract.String(), call["to"])
				require.Equal(t, owner.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(New(c))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestReadsMapRevertToSentinel(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	expiredArgs, err := abi.Pack(errAllowanceExpired.Inputs, big.NewInt(1))
	require.NoError(t, err)
	expiredSel := errAllowanceExpired.Selector()
	amountArgs, err := abi.Pack(errInvalidAmount.Inputs, big.NewInt(5))
	require.NoError(t, err)
	amountSel := errInvalidAmount.Selector()

	for _, tc := range []struct {
		name   string
		sel    abi.Selector
		args   []byte
		want   error
		wantNo error
	}{
		{"allowance expired", expiredSel, expiredArgs, ErrAllowanceExpired, ErrUnknown},
		{"invalid amount", amountSel, amountArgs, ErrInvalidAmount, ErrUnknown},
		{"insufficient allowance", errInsufficientAllowance.Selector(), nil, ErrInsufficientAllowance, ErrUnknown},
		{"excessive invalidation", errExcessiveInvalidation.Selector(), nil, ErrExcessiveInvalidation, ErrUnknown},
		{"signature expired", errSignatureExpired.Selector(), nil, ErrSignatureExpired, ErrUnknown},
		{"invalid nonce", errInvalidNonce.Selector(), nil, ErrInvalidNonce, ErrUnknown},
		{"invalid signature length", errInvalidSignatureLength.Selector(), nil, ErrInvalidSignatureLength, ErrUnknown},
		{"invalid signature", errInvalidSignature.Selector(), nil, ErrInvalidSignature, ErrUnknown},
		{"invalid signer", errInvalidSigner.Selector(), nil, ErrInvalidSigner, ErrUnknown},
		{"invalid contract signature", errInvalidContractSignature.Selector(), nil, ErrInvalidContractSignature, ErrUnknown},
		{"unsafe cast", errUnsafeCast.Selector(), nil, ErrUnsafeCast, ErrUnknown},
		{"length mismatch", errLengthMismatch.Selector(), nil, ErrLengthMismatch, ErrUnknown},
		{"unknown selector", abi.Selector{0xde, 0xad, 0xbe, 0xef}, nil, ErrUnknown, ErrInvalidNonce},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			revert := append(append([]byte{}, tc.sel[:]...), tc.args...)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Once()

			n, err := New(c).NonceBitmap(t.Context(), nil, NewNonceBitmapRequest(contract, owner, big.NewInt(0)))
			require.Nil(t, n)
			require.ErrorIs(t, err, tc.want)
			require.NotErrorIs(t, err, tc.wantNo)
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)
		})
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted"}).Once()
	n, err := New(c).NonceBitmap(t.Context(), nil, NewNonceBitmapRequest(contract, owner, big.NewInt(0)))
	require.Nil(t, n)
	require.NotErrorIs(t, err, ErrUnknown)
	require.ErrorContains(t, err, "permit2: nonceBitmap: ")
}

func TestReadsRejectMalformedResult(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(3)
			p := New(c)

			h, err := p.DomainSeparator(t.Context(), nil, NewDomainSeparatorRequest(contract))
			require.Nil(t, h)
			require.ErrorContains(t, err, "permit2: domainSeparator: ")

			a, err := p.Allowance(t.Context(), nil, NewAllowanceRequest(contract, owner, owner, owner))
			require.Nil(t, a)
			require.ErrorContains(t, err, "permit2: allowance: ")

			n, err := p.NonceBitmap(t.Context(), nil, NewNonceBitmapRequest(contract, owner, big.NewInt(0)))
			require.Nil(t, n)
			require.ErrorContains(t, err, "permit2: nonceBitmap: ")
		})
	}
}

func TestWritesSignAndSend(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	spender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	witness := core.Keccak256([]byte("witness"))
	witnessTypeString := "Witness witness)TokenPermissions(address token,uint256 amount)Witness(uint256 x)"
	domain := NewDomain(big.NewInt(1), contract)

	permitSingle := NewPermitSingle(NewPermitDetails(token, big.NewInt(1000), big.NewInt(1_700_000_000), big.NewInt(0)), spender, big.NewInt(1_800_000_000))
	singleSig, err := permitSingle.Sign(domain, key)
	require.NoError(t, err)
	permitBatch := NewPermitBatch([]PermitDetails{*NewPermitDetails(token, big.NewInt(1), big.NewInt(2), big.NewInt(3)), *NewPermitDetails(owner, big.NewInt(4), big.NewInt(5), big.NewInt(6))}, spender, big.NewInt(1_800_000_000))
	batchSig, err := permitBatch.Sign(domain, key)
	require.NoError(t, err)
	permit := NewPermitTransferFrom(NewTokenPermissions(token, big.NewInt(50)), big.NewInt(11), big.NewInt(1_800_000_000))
	permitSig, err := permit.Sign(domain, sender, key)
	require.NoError(t, err)
	batchPermit := NewPermitBatchTransferFrom([]TokenPermissions{*NewTokenPermissions(token, big.NewInt(50)), *NewTokenPermissions(owner, big.NewInt(60))}, big.NewInt(12), big.NewInt(1_800_000_000))
	batchPermitSig, err := batchPermit.Sign(domain, sender, key)
	require.NoError(t, err)
	single := NewSignatureTransferDetails(spender, big.NewInt(50))
	batch := []SignatureTransferDetails{*NewSignatureTransferDetails(spender, big.NewInt(50)), *NewSignatureTransferDetails(owner, big.NewInt(60))}
	transferDetails := []AllowanceTransferDetails{*NewAllowanceTransferDetails(owner, spender, big.NewInt(1), token), *NewAllowanceTransferDetails(owner, owner, big.NewInt(2), token)}
	approvals := []TokenSpenderPair{*NewTokenSpenderPair(token, spender), *NewTokenSpenderPair(owner, spender)}

	approveData, err := EncodeApprove(token, spender, big.NewInt(1000), big.NewInt(1_700_000_000))
	require.NoError(t, err)
	permitSingleData, err := EncodePermitSingle(owner, permitSingle, singleSig)
	require.NoError(t, err)
	permitBatchData, err := EncodePermitBatch(owner, permitBatch, batchSig)
	require.NoError(t, err)
	transferFromData, err := EncodeTransferFrom(owner, spender, big.NewInt(7), token)
	require.NoError(t, err)
	transferFromBatchData, err := EncodeTransferFromBatch(transferDetails)
	require.NoError(t, err)
	lockdownData, err := EncodeLockdown(approvals)
	require.NoError(t, err)
	invalidateNoncesData, err := EncodeInvalidateNonces(token, spender, big.NewInt(9))
	require.NoError(t, err)
	permitTransferFromData, err := EncodePermitTransferFrom(permit, single, owner, permitSig)
	require.NoError(t, err)
	permitWitnessTransferFromData, err := EncodePermitWitnessTransferFrom(permit, single, owner, witness, witnessTypeString, permitSig)
	require.NoError(t, err)
	permitTransferFromBatchData, err := EncodePermitTransferFromBatch(batchPermit, batch, owner, batchPermitSig)
	require.NoError(t, err)
	permitWitnessTransferFromBatchData, err := EncodePermitWitnessTransferFromBatch(batchPermit, batch, owner, witness, witnessTypeString, batchPermitSig)
	require.NoError(t, err)
	invalidateUnorderedNoncesData, err := EncodeInvalidateUnorderedNonces(big.NewInt(2), big.NewInt(0xff))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"approve",
			approveData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Approve(t.Context(), opts, NewApproveRequest(contract, token, spender, big.NewInt(1000), big.NewInt(1_700_000_000)))
			},
		},
		{
			"permitSingle",
			permitSingleData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitSingle(t.Context(), opts, NewPermitSingleRequest(contract, owner, permitSingle, singleSig))
			},
		},
		{
			"permitBatch",
			permitBatchData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitBatch(t.Context(), opts, NewPermitBatchRequest(contract, owner, permitBatch, batchSig))
			},
		},
		{
			"transferFrom",
			transferFromData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.TransferFrom(t.Context(), opts, NewTransferFromRequest(contract, owner, spender, big.NewInt(7), token))
			},
		},
		{
			"transferFromBatch",
			transferFromBatchData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.TransferFromBatch(t.Context(), opts, NewTransferFromBatchRequest(contract, transferDetails))
			},
		},
		{
			"lockdown",
			lockdownData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Lockdown(t.Context(), opts, NewLockdownRequest(contract, approvals))
			},
		},
		{
			"invalidateNonces",
			invalidateNoncesData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.InvalidateNonces(t.Context(), opts, NewInvalidateNoncesRequest(contract, token, spender, big.NewInt(9)))
			},
		},
		{
			"permitTransferFrom",
			permitTransferFromData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitTransferFrom(t.Context(), opts, NewPermitTransferFromRequest(contract, permit, single, owner, permitSig))
			},
		},
		{
			"permitWitnessTransferFrom",
			permitWitnessTransferFromData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitWitnessTransferFrom(t.Context(), opts, NewPermitWitnessTransferFromRequest(contract, permit, single, owner, witness, witnessTypeString, permitSig))
			},
		},
		{
			"permitTransferFromBatch",
			permitTransferFromBatchData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitTransferFromBatch(t.Context(), opts, NewPermitTransferFromBatchRequest(contract, batchPermit, batch, owner, batchPermitSig))
			},
		},
		{
			"permitWitnessTransferFromBatch",
			permitWitnessTransferFromBatchData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.PermitWitnessTransferFromBatch(t.Context(), opts, NewPermitWitnessTransferFromBatchRequest(contract, batchPermit, batch, owner, witness, witnessTypeString, batchPermitSig))
			},
		},
		{
			"invalidateUnorderedNonces",
			invalidateUnorderedNoncesData,
			func(p *Permit2, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.InvalidateUnorderedNonces(t.Context(), opts, NewInvalidateUnorderedNoncesRequest(contract, big.NewInt(2), big.NewInt(0xff)))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(300_000)
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

			signed, hash, err := tc.send(New(c), bind.NewTransactOptions(key, builder))
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(contract))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestClientMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	sig := &types.Signature{}
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	p := New(mocks.NewClient(t))
	ctx := t.Context()

	_, err = p.DomainSeparator(ctx, nil, &DomainSeparatorRequest{})
	require.EqualError(t, err, "permit2: domainSeparator: "+(&DomainSeparatorRequest{}).Validate().Error())
	_, err = p.Allowance(ctx, nil, NewAllowanceRequest(contract, owner, nil, owner))
	require.EqualError(t, err, "permit2: allowance: "+NewAllowanceRequest(contract, owner, nil, owner).Validate().Error())
	_, err = p.NonceBitmap(ctx, nil, NewNonceBitmapRequest(contract, owner, nil))
	require.EqualError(t, err, "permit2: nonceBitmap: "+NewNonceBitmapRequest(contract, owner, nil).Validate().Error())

	_, _, err = p.Approve(ctx, opts, NewApproveRequest(contract, token, owner, big.NewInt(1), nil))
	require.EqualError(t, err, "permit2: approve: "+NewApproveRequest(contract, token, owner, big.NewInt(1), nil).Validate().Error())
	_, _, err = p.PermitSingle(ctx, opts, NewPermitSingleRequest(contract, owner, NewPermitSingle(nil, owner, big.NewInt(1)), sig))
	require.EqualError(t, err, "permit2: permitSingle: invalid request: permitSingle.details is nil")
	_, _, err = p.PermitBatch(ctx, opts, NewPermitBatchRequest(contract, owner, NewPermitBatch(nil, owner, big.NewInt(1)), sig))
	require.EqualError(t, err, "permit2: permitBatch: invalid request: permitBatch.details is empty")
	_, _, err = p.TransferFrom(ctx, opts, NewTransferFromRequest(contract, owner, nil, big.NewInt(1), token))
	require.EqualError(t, err, "permit2: transferFrom: invalid request: to is nil")
	_, _, err = p.TransferFromBatch(ctx, opts, NewTransferFromBatchRequest(contract, nil))
	require.EqualError(t, err, "permit2: transferFromBatch: invalid request: transferDetails is empty")
	_, _, err = p.Lockdown(ctx, opts, NewLockdownRequest(contract, []TokenSpenderPair{*NewTokenSpenderPair(token, nil)}))
	require.EqualError(t, err, "permit2: lockdown: invalid request: approvals[0].spender is nil")
	_, _, err = p.InvalidateNonces(ctx, opts, &InvalidateNoncesRequest{})
	require.EqualError(t, err, "permit2: invalidateNonces: invalid request: contract is nil")
	_, _, err = p.PermitTransferFrom(ctx, opts, NewPermitTransferFromRequest(contract, NewPermitTransferFrom(nil, big.NewInt(1), big.NewInt(1)), NewSignatureTransferDetails(owner, big.NewInt(1)), owner, sig))
	require.EqualError(t, err, "permit2: permitTransferFrom: invalid request: permit.permitted is nil")
	_, _, err = p.PermitWitnessTransferFrom(ctx, opts, NewPermitWitnessTransferFromRequest(contract, NewPermitTransferFrom(NewTokenPermissions(token, big.NewInt(1)), big.NewInt(1), big.NewInt(1)), NewSignatureTransferDetails(owner, big.NewInt(1)), owner, core.Keccak256([]byte("w")), "", sig))
	require.EqualError(t, err, "permit2: permitWitnessTransferFrom: invalid request: witnessTypeString is empty")
	_, _, err = p.PermitTransferFromBatch(ctx, opts, NewPermitTransferFromBatchRequest(contract, NewPermitBatchTransferFrom([]TokenPermissions{*NewTokenPermissions(token, big.NewInt(1))}, big.NewInt(1), big.NewInt(1)), nil, owner, sig))
	require.EqualError(t, err, "permit2: permitTransferFromBatch: invalid request: transferDetails length 0 does not match permit.permitted length 1")
	_, _, err = p.PermitWitnessTransferFromBatch(ctx, opts, nil)
	require.EqualError(t, err, "permit2: permitWitnessTransferFromBatch: invalid request: request is nil")
	_, _, err = p.InvalidateUnorderedNonces(ctx, opts, NewInvalidateUnorderedNoncesRequest(contract, big.NewInt(0), nil))
	require.EqualError(t, err, "permit2: invalidateUnorderedNonces: invalid request: mask is nil")
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = p.Approve(ctx, opts, NewApproveRequest(contract, token, owner, new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(0)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)
	require.ErrorContains(t, err, "permit2: approve: permit2: encodeApprove: ")

	_, _, err = p.Lockdown(ctx, nil, NewLockdownRequest(contract, []TokenSpenderPair{*NewTokenSpenderPair(token, owner)}))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "permit2: lockdown: bind: transact: invalid options: options is nil")
}

func TestWritesMapEstimateRevertToSentinel(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token := types.MustNewAddressFromHex("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")

	args, err := abi.Pack(errInsufficientAllowance.Inputs, big.NewInt(0))
	require.NoError(t, err)
	sel := errInsufficientAllowance.Selector()
	revert := append(append([]byte{}, sel[:]...), args...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := New(c).TransferFrom(t.Context(), bind.NewTransactOptions(key, builder), NewTransferFromRequest(contract, owner, owner, big.NewInt(1), token))
	require.ErrorIs(t, err, ErrInsufficientAllowance)
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(100_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: -32000, Message: "nonce too low"}).Once()

	signed, hash, err := New(c).InvalidateUnorderedNonces(t.Context(), bind.NewTransactOptions(key, builder), NewInvalidateUnorderedNoncesRequest(contract, big.NewInt(0), big.NewInt(1)))
	require.ErrorContains(t, err, "permit2: invalidateUnorderedNonces: ")
	require.ErrorContains(t, err, "nonce too low")
	require.NotErrorIs(t, err, ErrUnknown)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "permit2")
	contract := types.MustNewAddressFromHex("0x000000000022D473030F116dDEE9F6B43aC78BA3")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "permit2", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = New(c).NonceBitmap(ctx, nil, NewNonceBitmapRequest(contract, owner, big.NewInt(0)))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(100_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = New(c).InvalidateUnorderedNonces(ctx, bind.NewTransactOptions(key, builder), NewInvalidateUnorderedNoncesRequest(contract, big.NewInt(0), big.NewInt(1)))
	require.NoError(t, err)
}
