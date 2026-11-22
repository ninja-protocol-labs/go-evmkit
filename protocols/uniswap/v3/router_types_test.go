package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewExactInputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	fee := big.NewInt(1)
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountIn := big.NewInt(3)
	amountOutMinimum := big.NewInt(4)
	sqrtPriceLimitX96 := big.NewInt(5)

	p := NewExactInputSingleParams(tokenIn, tokenOut, fee, recipient, amountIn, amountOutMinimum, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, fee, p.Fee)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewExactInputParams(t *testing.T) {
	path := []byte{0x01, 0x02}
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountIn := big.NewInt(2)
	amountOutMinimum := big.NewInt(3)

	p := NewExactInputParams(path, recipient, amountIn, amountOutMinimum)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
	require.Equal(t, path, p.Path)
}

func TestNewExactOutputSingleParams(t *testing.T) {
	tokenIn := types.MustNewAddressFromHex("0x1111111111111111111111111111111111111111")
	tokenOut := types.MustNewAddressFromHex("0x2222222222222222222222222222222222222222")
	fee := big.NewInt(1)
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountOut := big.NewInt(3)
	amountInMaximum := big.NewInt(4)
	sqrtPriceLimitX96 := big.NewInt(5)

	p := NewExactOutputSingleParams(tokenIn, tokenOut, fee, recipient, amountOut, amountInMaximum, sqrtPriceLimitX96)
	require.Same(t, tokenIn, p.TokenIn)
	require.Same(t, tokenOut, p.TokenOut)
	require.Same(t, fee, p.Fee)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
	require.Same(t, sqrtPriceLimitX96, p.SqrtPriceLimitX96)
}

func TestNewExactOutputParams(t *testing.T) {
	path := []byte{0x03, 0x04}
	recipient := types.MustNewAddressFromHex("0x3333333333333333333333333333333333333333")
	amountOut := big.NewInt(2)
	amountInMaximum := big.NewInt(3)

	p := NewExactOutputParams(path, recipient, amountOut, amountInMaximum)
	require.Same(t, recipient, p.Recipient)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
	require.Equal(t, path, p.Path)
}

func TestRouterRequestsValidate(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), a.Bytes()...)
	inSingle := NewExactInputSingleParams(a, a, big.NewInt(3000), a, big.NewInt(1000), big.NewInt(0), big.NewInt(0))
	inMulti := NewExactInputParams(path, a, big.NewInt(1000), big.NewInt(0))
	outSingle := NewExactOutputSingleParams(a, a, big.NewInt(3000), a, big.NewInt(1000), big.NewInt(2000), big.NewInt(0))
	outMulti := NewExactOutputParams(path, a, big.NewInt(1000), big.NewInt(2000))

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"exactInputSingle nil", (*ExactInputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactInputSingle empty", (&ExactInputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactInputSingle nil contract", (&ExactInputSingleRequest{Params: inSingle, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"exactInputSingle nil params", (&ExactInputSingleRequest{Contract: router, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"exactInputSingle nil params.tokenIn", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenOut: inSingle.TokenOut, Fee: inSingle.Fee, Recipient: inSingle.Recipient, AmountIn: inSingle.AmountIn, AmountOutMinimum: inSingle.AmountOutMinimum, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tokenIn is nil"},
		{"exactInputSingle nil params.tokenOut", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, Fee: inSingle.Fee, Recipient: inSingle.Recipient, AmountIn: inSingle.AmountIn, AmountOutMinimum: inSingle.AmountOutMinimum, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tokenOut is nil"},
		{"exactInputSingle nil params.fee", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, TokenOut: inSingle.TokenOut, Recipient: inSingle.Recipient, AmountIn: inSingle.AmountIn, AmountOutMinimum: inSingle.AmountOutMinimum, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.fee is nil"},
		{"exactInputSingle nil params.recipient", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, TokenOut: inSingle.TokenOut, Fee: inSingle.Fee, AmountIn: inSingle.AmountIn, AmountOutMinimum: inSingle.AmountOutMinimum, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.recipient is nil"},
		{"exactInputSingle nil params.amountIn", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, TokenOut: inSingle.TokenOut, Fee: inSingle.Fee, Recipient: inSingle.Recipient, AmountOutMinimum: inSingle.AmountOutMinimum, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountIn is nil"},
		{"exactInputSingle nil params.amountOutMinimum", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, TokenOut: inSingle.TokenOut, Fee: inSingle.Fee, Recipient: inSingle.Recipient, AmountIn: inSingle.AmountIn, SqrtPriceLimitX96: inSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountOutMinimum is nil"},
		{"exactInputSingle nil params.sqrtPriceLimitX96", (&ExactInputSingleRequest{Contract: router, Params: &ExactInputSingleParams{TokenIn: inSingle.TokenIn, TokenOut: inSingle.TokenOut, Fee: inSingle.Fee, Recipient: inSingle.Recipient, AmountIn: inSingle.AmountIn, AmountOutMinimum: inSingle.AmountOutMinimum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.sqrtPriceLimitX96 is nil"},
		{"exactInputSingle valid", (&ExactInputSingleRequest{Contract: router, Params: inSingle, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"exactInputSingle nil nativeAmount", (&ExactInputSingleRequest{Contract: router, Params: inSingle}).Validate(), ""},
		{"exactInput nil", (*ExactInputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactInput empty", (&ExactInputRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactInput nil contract", (&ExactInputRequest{Params: inMulti, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"exactInput nil params", (&ExactInputRequest{Contract: router, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"exactInput empty params.path", (&ExactInputRequest{Contract: router, Params: &ExactInputParams{Recipient: inMulti.Recipient, AmountIn: inMulti.AmountIn, AmountOutMinimum: inMulti.AmountOutMinimum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.path is empty"},
		{"exactInput nil params.recipient", (&ExactInputRequest{Contract: router, Params: &ExactInputParams{Path: inMulti.Path, AmountIn: inMulti.AmountIn, AmountOutMinimum: inMulti.AmountOutMinimum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.recipient is nil"},
		{"exactInput nil params.amountIn", (&ExactInputRequest{Contract: router, Params: &ExactInputParams{Path: inMulti.Path, Recipient: inMulti.Recipient, AmountOutMinimum: inMulti.AmountOutMinimum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountIn is nil"},
		{"exactInput nil params.amountOutMinimum", (&ExactInputRequest{Contract: router, Params: &ExactInputParams{Path: inMulti.Path, Recipient: inMulti.Recipient, AmountIn: inMulti.AmountIn}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountOutMinimum is nil"},
		{"exactInput valid", (&ExactInputRequest{Contract: router, Params: inMulti, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"exactInput nil nativeAmount", (&ExactInputRequest{Contract: router, Params: inMulti}).Validate(), ""},
		{"exactOutputSingle nil", (*ExactOutputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactOutputSingle empty", (&ExactOutputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactOutputSingle nil contract", (&ExactOutputSingleRequest{Params: outSingle, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"exactOutputSingle nil params", (&ExactOutputSingleRequest{Contract: router, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"exactOutputSingle nil params.tokenIn", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenOut: outSingle.TokenOut, Fee: outSingle.Fee, Recipient: outSingle.Recipient, AmountOut: outSingle.AmountOut, AmountInMaximum: outSingle.AmountInMaximum, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tokenIn is nil"},
		{"exactOutputSingle nil params.tokenOut", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, Fee: outSingle.Fee, Recipient: outSingle.Recipient, AmountOut: outSingle.AmountOut, AmountInMaximum: outSingle.AmountInMaximum, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.tokenOut is nil"},
		{"exactOutputSingle nil params.fee", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, TokenOut: outSingle.TokenOut, Recipient: outSingle.Recipient, AmountOut: outSingle.AmountOut, AmountInMaximum: outSingle.AmountInMaximum, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.fee is nil"},
		{"exactOutputSingle nil params.recipient", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, TokenOut: outSingle.TokenOut, Fee: outSingle.Fee, AmountOut: outSingle.AmountOut, AmountInMaximum: outSingle.AmountInMaximum, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.recipient is nil"},
		{"exactOutputSingle nil params.amountOut", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, TokenOut: outSingle.TokenOut, Fee: outSingle.Fee, Recipient: outSingle.Recipient, AmountInMaximum: outSingle.AmountInMaximum, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountOut is nil"},
		{"exactOutputSingle nil params.amountInMaximum", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, TokenOut: outSingle.TokenOut, Fee: outSingle.Fee, Recipient: outSingle.Recipient, AmountOut: outSingle.AmountOut, SqrtPriceLimitX96: outSingle.SqrtPriceLimitX96}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountInMaximum is nil"},
		{"exactOutputSingle nil params.sqrtPriceLimitX96", (&ExactOutputSingleRequest{Contract: router, Params: &ExactOutputSingleParams{TokenIn: outSingle.TokenIn, TokenOut: outSingle.TokenOut, Fee: outSingle.Fee, Recipient: outSingle.Recipient, AmountOut: outSingle.AmountOut, AmountInMaximum: outSingle.AmountInMaximum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.sqrtPriceLimitX96 is nil"},
		{"exactOutputSingle valid", (&ExactOutputSingleRequest{Contract: router, Params: outSingle, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"exactOutputSingle nil nativeAmount", (&ExactOutputSingleRequest{Contract: router, Params: outSingle}).Validate(), ""},
		{"exactOutput nil", (*ExactOutputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactOutput empty", (&ExactOutputRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactOutput nil contract", (&ExactOutputRequest{Params: outMulti, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: contract is nil"},
		{"exactOutput nil params", (&ExactOutputRequest{Contract: router, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params is nil"},
		{"exactOutput empty params.path", (&ExactOutputRequest{Contract: router, Params: &ExactOutputParams{Recipient: outMulti.Recipient, AmountOut: outMulti.AmountOut, AmountInMaximum: outMulti.AmountInMaximum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.path is empty"},
		{"exactOutput nil params.recipient", (&ExactOutputRequest{Contract: router, Params: &ExactOutputParams{Path: outMulti.Path, AmountOut: outMulti.AmountOut, AmountInMaximum: outMulti.AmountInMaximum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.recipient is nil"},
		{"exactOutput nil params.amountOut", (&ExactOutputRequest{Contract: router, Params: &ExactOutputParams{Path: outMulti.Path, Recipient: outMulti.Recipient, AmountInMaximum: outMulti.AmountInMaximum}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountOut is nil"},
		{"exactOutput nil params.amountInMaximum", (&ExactOutputRequest{Contract: router, Params: &ExactOutputParams{Path: outMulti.Path, Recipient: outMulti.Recipient, AmountOut: outMulti.AmountOut}, NativeAmount: big.NewInt(21)}).Validate(), "invalid request: params.amountInMaximum is nil"},
		{"exactOutput valid", (&ExactOutputRequest{Contract: router, Params: outMulti, NativeAmount: big.NewInt(21)}).Validate(), ""},
		{"exactOutput nil nativeAmount", (&ExactOutputRequest{Contract: router, Params: outMulti}).Validate(), ""},
		{"routerUnwrapWETH9 nil", (*RouterUnwrapWETH9Request)(nil).Validate(), "invalid request: request is nil"},
		{"routerUnwrapWETH9 empty", (&RouterUnwrapWETH9Request{}).Validate(), "invalid request: contract is nil"},
		{"routerUnwrapWETH9 nil contract", (&RouterUnwrapWETH9Request{AmountMinimum: big.NewInt(11), Recipient: a}).Validate(), "invalid request: contract is nil"},
		{"routerUnwrapWETH9 nil amountMinimum", (&RouterUnwrapWETH9Request{Contract: router, Recipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"routerUnwrapWETH9 nil recipient", (&RouterUnwrapWETH9Request{Contract: router, AmountMinimum: big.NewInt(11)}).Validate(), "invalid request: recipient is nil"},
		{"routerUnwrapWETH9 valid", (&RouterUnwrapWETH9Request{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a}).Validate(), ""},
		{"routerUnwrapWETH9WithFee nil", (*RouterUnwrapWETH9WithFeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"routerUnwrapWETH9WithFee empty", (&RouterUnwrapWETH9WithFeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerUnwrapWETH9WithFee nil contract", (&RouterUnwrapWETH9WithFeeRequest{AmountMinimum: big.NewInt(11), Recipient: a, FeeBips: big.NewInt(31), FeeRecipient: a}).Validate(), "invalid request: contract is nil"},
		{"routerUnwrapWETH9WithFee nil amountMinimum", (&RouterUnwrapWETH9WithFeeRequest{Contract: router, Recipient: a, FeeBips: big.NewInt(31), FeeRecipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"routerUnwrapWETH9WithFee nil recipient", (&RouterUnwrapWETH9WithFeeRequest{Contract: router, AmountMinimum: big.NewInt(11), FeeBips: big.NewInt(31), FeeRecipient: a}).Validate(), "invalid request: recipient is nil"},
		{"routerUnwrapWETH9WithFee nil feeBips", (&RouterUnwrapWETH9WithFeeRequest{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a, FeeRecipient: a}).Validate(), "invalid request: feeBips is nil"},
		{"routerUnwrapWETH9WithFee nil feeRecipient", (&RouterUnwrapWETH9WithFeeRequest{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a, FeeBips: big.NewInt(31)}).Validate(), "invalid request: feeRecipient is nil"},
		{"routerUnwrapWETH9WithFee valid", (&RouterUnwrapWETH9WithFeeRequest{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a, FeeBips: big.NewInt(31), FeeRecipient: a}).Validate(), ""},
		{"routerSweepToken nil", (*RouterSweepTokenRequest)(nil).Validate(), "invalid request: request is nil"},
		{"routerSweepToken empty", (&RouterSweepTokenRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerSweepToken nil contract", (&RouterSweepTokenRequest{Token: a, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), "invalid request: contract is nil"},
		{"routerSweepToken nil token", (&RouterSweepTokenRequest{Contract: router, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), "invalid request: token is nil"},
		{"routerSweepToken nil amountMinimum", (&RouterSweepTokenRequest{Contract: router, Token: a, Recipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"routerSweepToken nil recipient", (&RouterSweepTokenRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21)}).Validate(), "invalid request: recipient is nil"},
		{"routerSweepToken valid", (&RouterSweepTokenRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a}).Validate(), ""},
		{"routerSweepTokenWithFee nil", (*RouterSweepTokenWithFeeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"routerSweepTokenWithFee empty", (&RouterSweepTokenWithFeeRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerSweepTokenWithFee nil contract", (&RouterSweepTokenWithFeeRequest{Token: a, AmountMinimum: big.NewInt(21), Recipient: a, FeeBips: big.NewInt(41), FeeRecipient: a}).Validate(), "invalid request: contract is nil"},
		{"routerSweepTokenWithFee nil token", (&RouterSweepTokenWithFeeRequest{Contract: router, AmountMinimum: big.NewInt(21), Recipient: a, FeeBips: big.NewInt(41), FeeRecipient: a}).Validate(), "invalid request: token is nil"},
		{"routerSweepTokenWithFee nil amountMinimum", (&RouterSweepTokenWithFeeRequest{Contract: router, Token: a, Recipient: a, FeeBips: big.NewInt(41), FeeRecipient: a}).Validate(), "invalid request: amountMinimum is nil"},
		{"routerSweepTokenWithFee nil recipient", (&RouterSweepTokenWithFeeRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), FeeBips: big.NewInt(41), FeeRecipient: a}).Validate(), "invalid request: recipient is nil"},
		{"routerSweepTokenWithFee nil feeBips", (&RouterSweepTokenWithFeeRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a, FeeRecipient: a}).Validate(), "invalid request: feeBips is nil"},
		{"routerSweepTokenWithFee nil feeRecipient", (&RouterSweepTokenWithFeeRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a, FeeBips: big.NewInt(41)}).Validate(), "invalid request: feeRecipient is nil"},
		{"routerSweepTokenWithFee valid", (&RouterSweepTokenWithFeeRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a, FeeBips: big.NewInt(41), FeeRecipient: a}).Validate(), ""},
		{"routerRefundETH nil", (*RouterRefundETHRequest)(nil).Validate(), "invalid request: request is nil"},
		{"routerRefundETH empty", (&RouterRefundETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerRefundETH nil contract", (&RouterRefundETHRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerRefundETH valid", (&RouterRefundETHRequest{Contract: router}).Validate(), ""},
		{"routerMulticall nil", (*RouterMulticallRequest)(nil).Validate(), "invalid request: request is nil"},
		{"routerMulticall empty", (&RouterMulticallRequest{}).Validate(), "invalid request: contract is nil"},
		{"routerMulticall nil contract", (&RouterMulticallRequest{Data: [][]byte{{0x01}}, Deadline: big.NewInt(21), NativeAmount: big.NewInt(31)}).Validate(), "invalid request: contract is nil"},
		{"routerMulticall empty data", (&RouterMulticallRequest{Contract: router, Data: [][]byte{}, Deadline: big.NewInt(21), NativeAmount: big.NewInt(31)}).Validate(), "invalid request: data is empty"},
		{"routerMulticall valid", (&RouterMulticallRequest{Contract: router, Data: [][]byte{{0x01}}, Deadline: big.NewInt(21), NativeAmount: big.NewInt(31)}).Validate(), ""},
		{"routerMulticall nil nativeAmount", (&RouterMulticallRequest{Contract: router, Data: [][]byte{{0x01}}, Deadline: big.NewInt(21)}).Validate(), ""},
		{"routerMulticall nil deadline", (&RouterMulticallRequest{Contract: router, Data: [][]byte{{0x01}}, NativeAmount: big.NewInt(31)}).Validate(), ""},
		{"selfPermit nil", (*SelfPermitRequest)(nil).Validate(), "invalid request: request is nil"},
		{"selfPermit empty", (&SelfPermitRequest{}).Validate(), "invalid request: contract is nil"},
		{"selfPermit nil contract", (&SelfPermitRequest{Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"selfPermit nil token", (&SelfPermitRequest{Contract: router, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: token is nil"},
		{"selfPermit nil amount", (&SelfPermitRequest{Contract: router, Token: a, Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: amount is nil"},
		{"selfPermit nil deadline", (&SelfPermitRequest{Contract: router, Token: a, Amount: big.NewInt(21), Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"selfPermit nil sig", (&SelfPermitRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31)}).Validate(), "invalid request: sig is nil"},
		{"selfPermit valid", (&SelfPermitRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), ""},
		{"selfPermitIfNecessary nil", (*SelfPermitIfNecessaryRequest)(nil).Validate(), "invalid request: request is nil"},
		{"selfPermitIfNecessary empty", (&SelfPermitIfNecessaryRequest{}).Validate(), "invalid request: contract is nil"},
		{"selfPermitIfNecessary nil contract", (&SelfPermitIfNecessaryRequest{Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"selfPermitIfNecessary nil token", (&SelfPermitIfNecessaryRequest{Contract: router, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: token is nil"},
		{"selfPermitIfNecessary nil amount", (&SelfPermitIfNecessaryRequest{Contract: router, Token: a, Deadline: big.NewInt(31), Sig: sig}).Validate(), "invalid request: amount is nil"},
		{"selfPermitIfNecessary nil deadline", (&SelfPermitIfNecessaryRequest{Contract: router, Token: a, Amount: big.NewInt(21), Sig: sig}).Validate(), "invalid request: deadline is nil"},
		{"selfPermitIfNecessary nil sig", (&SelfPermitIfNecessaryRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31)}).Validate(), "invalid request: sig is nil"},
		{"selfPermitIfNecessary valid", (&SelfPermitIfNecessaryRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}).Validate(), ""},
		{"selfPermitAllowed nil", (*SelfPermitAllowedRequest)(nil).Validate(), "invalid request: request is nil"},
		{"selfPermitAllowed empty", (&SelfPermitAllowedRequest{}).Validate(), "invalid request: contract is nil"},
		{"selfPermitAllowed nil contract", (&SelfPermitAllowedRequest{Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"selfPermitAllowed nil token", (&SelfPermitAllowedRequest{Contract: router, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: token is nil"},
		{"selfPermitAllowed nil nonce", (&SelfPermitAllowedRequest{Contract: router, Token: a, Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: nonce is nil"},
		{"selfPermitAllowed nil expiry", (&SelfPermitAllowedRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Sig: sig}).Validate(), "invalid request: expiry is nil"},
		{"selfPermitAllowed nil sig", (&SelfPermitAllowedRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31)}).Validate(), "invalid request: sig is nil"},
		{"selfPermitAllowed valid", (&SelfPermitAllowedRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), ""},
		{"selfPermitAllowedIfNecessary nil", (*SelfPermitAllowedIfNecessaryRequest)(nil).Validate(), "invalid request: request is nil"},
		{"selfPermitAllowedIfNecessary empty", (&SelfPermitAllowedIfNecessaryRequest{}).Validate(), "invalid request: contract is nil"},
		{"selfPermitAllowedIfNecessary nil contract", (&SelfPermitAllowedIfNecessaryRequest{Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: contract is nil"},
		{"selfPermitAllowedIfNecessary nil token", (&SelfPermitAllowedIfNecessaryRequest{Contract: router, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: token is nil"},
		{"selfPermitAllowedIfNecessary nil nonce", (&SelfPermitAllowedIfNecessaryRequest{Contract: router, Token: a, Expiry: big.NewInt(31), Sig: sig}).Validate(), "invalid request: nonce is nil"},
		{"selfPermitAllowedIfNecessary nil expiry", (&SelfPermitAllowedIfNecessaryRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Sig: sig}).Validate(), "invalid request: expiry is nil"},
		{"selfPermitAllowedIfNecessary nil sig", (&SelfPermitAllowedIfNecessaryRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31)}).Validate(), "invalid request: sig is nil"},
		{"selfPermitAllowedIfNecessary valid", (&SelfPermitAllowedIfNecessaryRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}).Validate(), ""},
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

func TestNewRouterRequestsSetFields(t *testing.T) {
	router := types.MustNewAddressFromHex("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sig, err := key.Sign(core.Keccak256([]byte("permit")))
	require.NoError(t, err)
	path := append(append(append([]byte{}, a.Bytes()...), 0x00, 0x0b, 0xb8), a.Bytes()...)
	inSingle := NewExactInputSingleParams(a, a, big.NewInt(3000), a, big.NewInt(1000), big.NewInt(0), big.NewInt(0))
	inMulti := NewExactInputParams(path, a, big.NewInt(1000), big.NewInt(0))
	outSingle := NewExactOutputSingleParams(a, a, big.NewInt(3000), a, big.NewInt(1000), big.NewInt(2000), big.NewInt(0))
	outMulti := NewExactOutputParams(path, a, big.NewInt(1000), big.NewInt(2000))

	require.Equal(t, &ExactInputSingleRequest{Contract: router, Params: inSingle, NativeAmount: big.NewInt(21)}, NewExactInputSingleRequest(router, inSingle, big.NewInt(21)))
	require.Equal(t, &ExactInputRequest{Contract: router, Params: inMulti, NativeAmount: big.NewInt(21)}, NewExactInputRequest(router, inMulti, big.NewInt(21)))
	require.Equal(t, &ExactOutputSingleRequest{Contract: router, Params: outSingle, NativeAmount: big.NewInt(21)}, NewExactOutputSingleRequest(router, outSingle, big.NewInt(21)))
	require.Equal(t, &ExactOutputRequest{Contract: router, Params: outMulti, NativeAmount: big.NewInt(21)}, NewExactOutputRequest(router, outMulti, big.NewInt(21)))
	require.Equal(t, &RouterUnwrapWETH9Request{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a}, NewRouterUnwrapWETH9Request(router, big.NewInt(11), a))
	require.Equal(t, &RouterUnwrapWETH9WithFeeRequest{Contract: router, AmountMinimum: big.NewInt(11), Recipient: a, FeeBips: big.NewInt(31), FeeRecipient: a}, NewRouterUnwrapWETH9WithFeeRequest(router, big.NewInt(11), a, big.NewInt(31), a))
	require.Equal(t, &RouterSweepTokenRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a}, NewRouterSweepTokenRequest(router, a, big.NewInt(21), a))
	require.Equal(t, &RouterSweepTokenWithFeeRequest{Contract: router, Token: a, AmountMinimum: big.NewInt(21), Recipient: a, FeeBips: big.NewInt(41), FeeRecipient: a}, NewRouterSweepTokenWithFeeRequest(router, a, big.NewInt(21), a, big.NewInt(41), a))
	require.Equal(t, &RouterRefundETHRequest{Contract: router}, NewRouterRefundETHRequest(router))
	require.Equal(t, &RouterMulticallRequest{Contract: router, Data: [][]byte{{0x01}}, Deadline: big.NewInt(21), NativeAmount: big.NewInt(31)}, NewRouterMulticallRequest(router, [][]byte{{0x01}}, big.NewInt(21), big.NewInt(31)))
	require.Equal(t, &SelfPermitRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}, NewSelfPermitRequest(router, a, big.NewInt(21), big.NewInt(31), sig))
	require.Equal(t, &SelfPermitIfNecessaryRequest{Contract: router, Token: a, Amount: big.NewInt(21), Deadline: big.NewInt(31), Sig: sig}, NewSelfPermitIfNecessaryRequest(router, a, big.NewInt(21), big.NewInt(31), sig))
	require.Equal(t, &SelfPermitAllowedRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}, NewSelfPermitAllowedRequest(router, a, big.NewInt(21), big.NewInt(31), sig))
	require.Equal(t, &SelfPermitAllowedIfNecessaryRequest{Contract: router, Token: a, Nonce: big.NewInt(21), Expiry: big.NewInt(31), Sig: sig}, NewSelfPermitAllowedIfNecessaryRequest(router, a, big.NewInt(21), big.NewInt(31), sig))
}
