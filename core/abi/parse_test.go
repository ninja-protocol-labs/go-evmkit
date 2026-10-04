package abi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTypeScalars(t *testing.T) {
	tests := []struct {
		in   string
		want Type
	}{
		{"bool", Bool},
		{"address", Address},
		{"string", String},
		{"bytes", Bytes},
		{"function", FunctionType},
		{"uint8", Uint8},
		{"uint256", Uint256},
		{"uint", Uint256},
		{"int8", Int8},
		{"int256", Int256},
		{"int", Int256},
		{"bytes1", Bytes1},
		{"bytes32", Bytes32},
		{"uint72", Type{Kind: KindUint, Size: 72}},
		{"int128", Int128},
	}
	for _, tt := range tests {
		got, err := ParseType(tt.in)
		require.NoError(t, err, tt.in)
		require.Equal(t, tt.want, got, tt.in)
	}
}

func TestParseTypeWhitespaceTolerant(t *testing.T) {
	got, err := ParseType("  uint256  ")
	require.NoError(t, err)
	require.Equal(t, Uint256, got)
}

func TestParseTypeSlice(t *testing.T) {
	got, err := ParseType("uint256[]")
	require.NoError(t, err)
	require.Equal(t, Slice(Uint256), got)
}

func TestParseTypeArray(t *testing.T) {
	got, err := ParseType("address[3]")
	require.NoError(t, err)
	want, err := Array(Address, 3)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestParseTypeArrayOfArray(t *testing.T) {
	got, err := ParseType("uint8[2][3]")
	require.NoError(t, err)

	inner, err := Array(Uint8, 2)
	require.NoError(t, err)
	want, err := Array(inner, 3)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestParseTypeTuple(t *testing.T) {
	got, err := ParseType("(address,uint256)")
	require.NoError(t, err)
	require.Equal(t, Tuple(Address, Uint256), got)
}

func TestParseTypeNestedSliceOfTuple(t *testing.T) {
	got, err := ParseType("(address,uint256[])[]")
	require.NoError(t, err)

	inner := Tuple(Address, Slice(Uint256))
	require.Equal(t, Slice(inner), got)
}

func TestParseTypeRoundTripsWithString(t *testing.T) {
	types := []Type{
		Bool, Address, String, Bytes, FunctionType,
		Uint8, Uint256, Int8, Int256, Bytes1, Bytes32,
		Slice(Uint256), Tuple(Address, Uint256),
		Slice(Tuple(Address, Slice(Uint256))),
	}
	for _, typ := range types {
		s := typ.String()
		got, err := ParseType(s)
		require.NoError(t, err, s)
		require.Equal(t, typ, got, s)
	}
}

func TestParseTypeInvalid(t *testing.T) {
	tests := []string{
		"",
		"uint257",
		"uint7",
		"bytes33",
		"bytes0",
		"nonsense",
		"(address,uint256",
		"address]",
		"uint256[abc]",
	}
	for _, in := range tests {
		_, err := ParseType(in)
		require.ErrorIs(t, err, ErrInvalidTypeString, in)
	}
}

func TestParseTypeTrailingGarbageRejected(t *testing.T) {
	_, err := ParseType("uint256 extra")
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseFunctionNoNames(t *testing.T) {
	fn, err := ParseFunction("transfer(address,uint256)", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, "transfer", fn.Name)
	require.Equal(t, Types{Address, Uint256}, fn.Inputs)
	require.Equal(t, Selector{0xa9, 0x05, 0x9c, 0xbb}, fn.Selector())
}

func TestParseFunctionWithNamesNoSpaceAfterComma(t *testing.T) {
	fn, err := ParseFunction("transfer(address to,uint256 amount)", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, Types{Address, Uint256}, fn.Inputs)
	require.Equal(t, Selector{0xa9, 0x05, 0x9c, 0xbb}, fn.Selector())
}

func TestParseFunctionWithNamesSpaceAfterComma(t *testing.T) {
	fn, err := ParseFunction("transfer(address to, uint256 amount)", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, Types{Address, Uint256}, fn.Inputs)
	require.Equal(t, Selector{0xa9, 0x05, 0x9c, 0xbb}, fn.Selector())
}

func TestParseFunctionBothFormsMatch(t *testing.T) {
	a, err := ParseFunction("transfer(address to, uint256 amount)", nil)
	require.NoError(t, err)
	b, err := ParseFunction("transfer(address to,uint256 amount)", nil)
	require.NoError(t, err)
	require.Equal(t, a.Signature(), b.Signature())
	require.Equal(t, a.Selector(), b.Selector())
}

func TestParseFunctionNoArgs(t *testing.T) {
	fn, err := ParseFunction("totalSupply()", Types{Uint256})
	require.NoError(t, err)
	require.Empty(t, fn.Inputs)
	require.Equal(t, Selector{0x18, 0x16, 0x0d, 0xdd}, fn.Selector())
}

func TestParseFunctionNestedTupleWithNames(t *testing.T) {
	fn, err := ParseFunction("swap((address tokenIn, address tokenOut, uint256 amountIn) params)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Tuple(Address, Address, Uint256)}, fn.Inputs)
}

func TestParseFunctionNestedTupleAndOtherParamWithNames(t *testing.T) {
	fn, err := ParseFunction("sync((address tokenIn, address tokenOut, uint256 amountIn) params, uint256 tick)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Tuple(Address, Address, Uint256), Uint256}, fn.Inputs)
}

func TestParseFunctionMissingParen(t *testing.T) {
	_, err := ParseFunction("transfer address,uint256", nil)
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseFunctionMissingName(t *testing.T) {
	_, err := ParseFunction("(address,uint256)", nil)
	require.ErrorIs(t, err, ErrInvalidTypeString)
}

func TestParseFunctionBalanceOfSelector(t *testing.T) {
	fn, err := ParseFunction("balanceOf(address account)", Types{Uint256})
	require.NoError(t, err)
	require.Equal(t, Selector{0x70, 0xa0, 0x82, 0x31}, fn.Selector())
}

func TestParseFunctionBareOutputType(t *testing.T) {
	fn, err := ParseFunction("transfer(address to, uint256 amount) bool", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Address, Uint256}, fn.Inputs)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionReturnsKeywordBareType(t *testing.T) {
	fn, err := ParseFunction("transfer(address to, uint256 amount) returns (bool)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionReturnsKeywordBareTypeNoParens(t *testing.T) {
	fn, err := ParseFunction("transfer(address to, uint256 amount) returns bool", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionMultipleOutputs(t *testing.T) {
	fn, err := ParseFunction("getReserves() returns (uint112 reserve0, uint112 reserve1)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Type{Kind: KindUint, Size: 112}, Type{Kind: KindUint, Size: 112}}, fn.Outputs)
}

func TestParseFunctionNoOutputsFallsBackToParam(t *testing.T) {
	fn, err := ParseFunction("transfer(address,uint256)", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionEmbeddedOutputsOverrideParam(t *testing.T) {
	fn, err := ParseFunction("transfer(address,uint256) bool", Types{Uint256})
	require.NoError(t, err)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionOutputsNoArgsInput(t *testing.T) {
	fn, err := ParseFunction("totalSupply() returns (uint256)", nil)
	require.NoError(t, err)
	require.Empty(t, fn.Inputs)
	require.Equal(t, Types{Uint256}, fn.Outputs)
}

func TestParseFunctionFullSolidityDeclaration(t *testing.T) {
	fn, err := ParseFunction(
		"function mul(uint256 a, uint256 b) internal constant returns (uint256, uint256)",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "mul", fn.Name)
	require.Equal(t, Types{Uint256, Uint256}, fn.Inputs)
	require.Equal(t, Types{Uint256, Uint256}, fn.Outputs)
}

func TestParseFunctionModifiersNoOutputs(t *testing.T) {
	fn, err := ParseFunction("function transfer(address to, uint256 amount) external", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, "transfer", fn.Name)
	require.Equal(t, Types{Address, Uint256}, fn.Inputs)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionKeywordPrefixOnFunctionNamedFunction(t *testing.T) {
	_, err := ParseFunction("functionCall(uint256)", nil)
	require.NoError(t, err)
}

func TestParseFunctionMultipleModifiersAndReturns(t *testing.T) {
	fn, err := ParseFunction("balanceOf(address account) public view virtual override returns (uint256)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Uint256}, fn.Outputs)
	require.Equal(t, Selector{0x70, 0xa0, 0x82, 0x31}, fn.Selector())
}

func TestParseFunctionCustomModifier(t *testing.T) {
	fn, err := ParseFunction("function withdraw(uint256 amount) public onlyOwner returns (bool)", nil)
	require.NoError(t, err)
	require.Equal(t, "withdraw", fn.Name)
	require.Equal(t, Types{Uint256}, fn.Inputs)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionCustomModifierNoOutputs(t *testing.T) {
	fn, err := ParseFunction("function withdraw(uint256 amount) external nonReentrant", Types{Bool})
	require.NoError(t, err)
	require.Equal(t, Types{Uint256}, fn.Inputs)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionCustomModifierBareOutputNoReturnsKeyword(t *testing.T) {
	fn, err := ParseFunction("function withdraw(uint256 amount) onlyOwner bool", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Bool}, fn.Outputs)
}

func TestParseFunctionSingleTupleOutput(t *testing.T) {
	fn, err := ParseFunction("getUser() returns ((address,uint256))", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Tuple(Address, Uint256)}, fn.Outputs)
}

func TestParseFunctionMultipleSeparateOutputsNotMergedIntoTuple(t *testing.T) {
	fn, err := ParseFunction("swap() returns (address,uint256)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Address, Uint256}, fn.Outputs)
}

func TestParseFunctionDataLocationMemory(t *testing.T) {
	fn, err := ParseFunction("swap(address[] memory path, uint256 amountIn, uint256 amountOutMin)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Slice(Address), Uint256, Uint256}, fn.Inputs)
}

func TestParseFunctionDataLocationCalldataNoName(t *testing.T) {
	fn, err := ParseFunction("swap(bytes calldata data)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Bytes}, fn.Inputs)
}

func TestParseFunctionDataLocationStorage(t *testing.T) {
	fn, err := ParseFunction("swap(uint256[] storage amounts)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Slice(Uint256)}, fn.Inputs)
}

func TestParseFunctionNestedTupleWithDataLocationAndModifiersAndReturns(t *testing.T) {
	fn, err := ParseFunction(
		"swap((address tokenIn, address tokenOut, uint256 amountIn, uint256 amountOutMin, address[] path, uint256 deadline) calldata params) external returns (uint256[] memory amounts)",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "swap", fn.Name)
	require.Equal(t, Types{Tuple(Address, Address, Uint256, Uint256, Slice(Address), Uint256)}, fn.Inputs)
	require.Equal(t, Types{Slice(Uint256)}, fn.Outputs)
}

func TestParseFunctionUnknownFutureDataLocationKeyword(t *testing.T) {
	fn, err := ParseFunction("swap(address[] transient path)", nil)
	require.NoError(t, err)
	require.Equal(t, Types{Slice(Address)}, fn.Inputs)
}

func TestParseFunctionThreeTrailingWordsRejected(t *testing.T) {
	_, err := ParseFunction("swap(uint256 foo bar baz)", nil)
	require.ErrorIs(t, err, ErrInvalidTypeString)
}
