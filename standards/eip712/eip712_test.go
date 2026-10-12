package eip712

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEIP712EncodeTypeMail(t *testing.T) {
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)
	got := ts.EncodeType("Mail")
	require.Equal(t, "Mail(Person from,Person to,string contents)Person(string name,address wallet)", got)
}

func TestEIP712TypeHashMail(t *testing.T) {
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)
	got := ts.TypeHash("Mail")
	want, err := hex.DecodeString("a0cedeb2dc280ba39b857546d74f5549c3a1d7bdc2dd96bf881f76108e23dac2")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712EncodeTypeTransaction(t *testing.T) {
	ts := NewTypes("Transaction",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("tx", "Asset"),
	)
	ts.Append("Asset",
		NewField("token", abi.Address.String()),
		NewField("amount", abi.Uint256.String()),
	)
	ts.Append("Person",
		NewField("wallet", abi.Address.String()),
		NewField("name", abi.String.String()),
	)
	got := ts.EncodeType("Transaction")
	require.Equal(t, "Transaction(Person from,Person to,Asset tx)Asset(address token,uint256 amount)Person(address wallet,string name)", got)
}

func TestEIP712TypeHashTransaction(t *testing.T) {
	ts := NewTypes("Transaction",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("tx", "Asset"),
	)
	ts.Append("Asset",
		NewField("token", abi.Address.String()),
		NewField("amount", abi.Uint256.String()),
	)
	ts.Append("Person",
		NewField("wallet", abi.Address.String()),
		NewField("name", abi.String.String()),
	)
	got := ts.TypeHash("Transaction")
	want, err := hex.DecodeString("358262ad2b1b6af9edb8b4f81ee9a13ec2ed2473132bcfe1721ac7a2e191791e")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712EncodeTypeNoReferences(t *testing.T) {
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	require.Equal(t, "Person(string name,address wallet)", ts.EncodeType("Person"))
}

func TestEIP712EncodeTypeArrayReference(t *testing.T) {
	ts := NewTypes("Group", NewField("members", "Person[]"))
	ts.Append("Person", NewField("name", abi.String.String()))
	got := ts.EncodeType("Group")
	require.Equal(t, "Group(Person[] members)Person(string name)", got)
}

func TestEIP712EncodeTypeUnknownType(t *testing.T) {
	require.Equal(t, "", Types{}.EncodeType("Nonexistent"))
}

func TestEIP712BaseTypeName(t *testing.T) {
	require.Equal(t, "Person", baseTypeName("Person"))
	require.Equal(t, "Person", baseTypeName("Person[]"))
	require.Equal(t, "Person", baseTypeName("Person[3]"))
	require.Equal(t, "Person", baseTypeName("Person[3][]"))
}

func TestEIP712EncodeTypeMutualCycle(t *testing.T) {
	ts := NewTypes("A", NewField("b", "B"))
	ts.Append("B", NewField("a", "A"))
	got := ts.EncodeType("A")
	require.Equal(t, "A(B b)B(A a)", got)
}

func TestEIP712EncodeTypeSelfReference(t *testing.T) {
	ts := NewTypes("Node",
		NewField("value", abi.Uint256.String()),
		NewField("children", "Node[]"),
	)
	got := ts.EncodeType("Node")
	require.Equal(t, "Node(uint256 value,Node[] children)", got)
}

func TestEIP712EncodeTypeDiamond(t *testing.T) {
	ts := NewTypes("Top",
		NewField("l", "Left"),
		NewField("r", "Right"),
	)
	ts.Append("Left", NewField("shared", "Shared"))
	ts.Append("Right", NewField("shared", "Shared"))
	ts.Append("Shared", NewField("value", abi.Uint256.String()))
	got := ts.EncodeType("Top")
	require.Equal(t, "Top(Left l,Right r)Left(Shared shared)Right(Shared shared)Shared(uint256 value)", got)
}

func TestEIP712HashStructPerson(t *testing.T) {
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)
	wallet, err := types.NewAddressFromHex("0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826")
	require.NoError(t, err)
	cow := map[string]any{
		"name":   "Cow",
		"wallet": wallet,
	}
	got, err := ts.HashStruct("Person", cow)
	require.NoError(t, err)
	want, err := hex.DecodeString("fc71e5fa27ff56c350aa531bc129ebdf613b772b6604664f5d8dbe21b85eb0c8")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712HashStructMail(t *testing.T) {
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)
	cowWallet, err := types.NewAddressFromHex("0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826")
	require.NoError(t, err)
	bobWallet, err := types.NewAddressFromHex("0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB")
	require.NoError(t, err)
	mail := map[string]any{
		"from": map[string]any{
			"name":   "Cow",
			"wallet": cowWallet,
		},
		"to": map[string]any{
			"name":   "Bob",
			"wallet": bobWallet,
		},
		"contents": "Hello, Bob!",
	}
	got, err := ts.HashStruct("Mail", mail)
	require.NoError(t, err)
	want, err := hex.DecodeString("c52c0ee5d84264471806290a3f2c4cecfc5490626bf912d01f240d7a274b371e")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712EncodeDataFieldCountMismatchIgnored(t *testing.T) {
	ts := NewTypes("Simple", NewField("a", abi.Uint256.String()))
	_, err := ts.EncodeData("Simple", map[string]any{"a": big.NewInt(1), "extra": "ignored"})
	require.NoError(t, err)
}

func TestEIP712EncodeDataMissingFieldErrors(t *testing.T) {
	ts := NewTypes("Simple", NewField("a", abi.Address.String()))
	_, err := ts.EncodeData("Simple", map[string]any{})
	require.Error(t, err)
}

func TestEIP712EncodeDataArrayOfStructs(t *testing.T) {
	ts := NewTypes("Group", NewField("members", "Person[]"))
	ts.Append("Person", NewField("name", abi.String.String()))
	group := map[string]any{
		"members": []any{
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		},
	}
	got, err := ts.HashStruct("Group", group)
	require.NoError(t, err)
	require.Len(t, got.Bytes(), 32)
}

func TestEIP712EncodeDataArrayOfBytes32(t *testing.T) {
	ts := NewTypes("Batch", NewField("ids", abi.Slice(abi.Bytes32).String()))
	id1 := make([]byte, 32)
	id1[0] = 0x01
	id2 := make([]byte, 32)
	id2[0] = 0x02

	got, err := ts.EncodeData("Batch", map[string]any{"ids": [][]byte{id1, id2}})
	require.NoError(t, err)
	require.Len(t, got, 64)
}

func TestEIP712EncodeDataUnknownFieldTypeErrors(t *testing.T) {
	ts := NewTypes("Bad", NewField("x", "notarealtype"))
	_, err := ts.EncodeData("Bad", map[string]any{"x": "whatever"})
	require.Error(t, err)
}

func TestEIP712EncodeDataArrayWrongGoTypeErrors(t *testing.T) {
	ts := NewTypes("Bad", NewField("x", abi.Slice(abi.Uint256).String()))
	_, err := ts.EncodeData("Bad", map[string]any{"x": "not a slice"})
	require.Error(t, err)
}

func TestEIP712DomainSeparator(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	domain := NewDomain("Ether Mail", "1", big.NewInt(1), contract, nil)

	got, err := DomainSeparator(domain)
	require.NoError(t, err)
	want, err := hex.DecodeString("f2cee375fa42b42143804025fc449deafd50cc031ca257e0b194a650a912090f")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712DomainSeparatorOmitsUnsetFields(t *testing.T) {
	domain := NewDomain("Only Name", "", nil, nil, nil)
	fields, data := domain.fieldsAndData()
	require.Len(t, fields, 1)
	require.Equal(t, "name", fields[0].Name)
	require.Len(t, data, 1)
}

func TestEIP712DomainSeparatorRejectsEmptyDomain(t *testing.T) {
	_, err := DomainSeparator(NewDomain("", "", nil, nil, nil))
	require.Error(t, err)
}

func TestEIP712Hash(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	domain := NewDomain("Ether Mail", "1", big.NewInt(1), contract, nil)
	cowWallet, err := types.NewAddressFromHex("0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826")
	require.NoError(t, err)
	bobWallet, err := types.NewAddressFromHex("0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB")
	require.NoError(t, err)
	mail := map[string]any{
		"from": map[string]any{
			"name":   "Cow",
			"wallet": cowWallet,
		},
		"to": map[string]any{
			"name":   "Bob",
			"wallet": bobWallet,
		},
		"contents": "Hello, Bob!",
	}
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)

	got, err := Hash(domain, "Mail", ts, mail)
	require.NoError(t, err)
	want, err := hex.DecodeString("be609aee343fb3c4b28e1df9e632fca64fcfaede20f02e86244efddf30957bd2")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712HashErrorsOnBadMessage(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	domain := NewDomain("Ether Mail", "1", big.NewInt(1), contract, nil)
	ts := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	ts.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
	)
	_, err = Hash(domain, "Mail", ts, map[string]any{})
	require.Error(t, err)
}

func TestEIP712DomainSeparatorWithSalt(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	salt := types.NewHashFromBytes([]byte("some salt value"))

	domain := NewDomain("Ether Mail", "1", big.NewInt(1), contract, salt)
	fields, data := domain.fieldsAndData()
	require.Len(t, fields, 5)
	require.Equal(t, "salt", fields[4].Name)
	require.Equal(t, "bytes32", fields[4].Type)
	require.Equal(t, salt.Bytes(), data["salt"])

	got, err := DomainSeparator(domain)
	require.NoError(t, err)
	require.Len(t, got.Bytes(), 32)
}

func TestEIP712EncodeDataBoolField(t *testing.T) {
	ts := NewTypes("Simple", NewField("ok", abi.Bool.String()))

	got, err := ts.EncodeData("Simple", map[string]any{"ok": true})
	require.NoError(t, err)
	require.Len(t, got, 64)

	gotFalse, err := ts.EncodeData("Simple", map[string]any{"ok": false})
	require.NoError(t, err)
	require.NotEqual(t, got, gotFalse)
}

func TestEIP712EncodeDataNegativeInt256Field(t *testing.T) {
	ts := NewTypes("Simple", NewField("delta", abi.Int256.String()))

	got, err := ts.EncodeData("Simple", map[string]any{"delta": big.NewInt(-42)})
	require.NoError(t, err)
	require.Len(t, got, 64)
	require.Equal(t, byte(0xff), got[32])
}

func TestEIP712EncodeDataSelfReferencingTree(t *testing.T) {
	ts := NewTypes("Node",
		NewField("value", abi.Uint256.String()),
		NewField("children", "Node[]"),
	)
	leaf := map[string]any{"value": big.NewInt(2), "children": []any{}}
	root := map[string]any{"value": big.NewInt(1), "children": []any{leaf}}

	got, err := ts.HashStruct("Node", root)
	require.NoError(t, err)
	require.Len(t, got.Bytes(), 32)

	leafHash, err := ts.HashStruct("Node", leaf)
	require.NoError(t, err)
	require.NotEqual(t, got.Bytes(), leafHash.Bytes())
}

func TestEIP712EncodeDataNestedArray(t *testing.T) {
	ts := NewTypes("Matrix", NewField("rows", abi.Slice(abi.Slice(abi.Uint256)).String()))

	got, err := ts.EncodeData("Matrix", map[string]any{
		"rows": [][]*big.Int{
			{big.NewInt(1), big.NewInt(2)},
			{big.NewInt(3), big.NewInt(4), big.NewInt(5)},
		},
	})
	require.NoError(t, err)
	require.Len(t, got, 64)
}

func TestEIP712EncodeDataArrayOfUint8IsAmbiguousWithBytes(t *testing.T) {
	ts := NewTypes("Matrix", NewField("rows", abi.Slice(abi.Slice(abi.Uint8)).String()))
	_, err := ts.EncodeData("Matrix", map[string]any{
		"rows": [][]uint8{{1, 2}, {3, 4, 5}},
	})
	require.Error(t, err)
}

func TestNewDomain(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	chainID := big.NewInt(1)
	salt := types.NewHashFromBytes([]byte("salt"))

	d := NewDomain("Ether Mail", "1", chainID, contract, salt)
	require.Equal(t, "Ether Mail", d.Name)
	require.Equal(t, "1", d.Version)
	require.Same(t, chainID, d.ChainID)
	require.Same(t, contract, d.VerifyingContract)
	require.Same(t, salt, d.Salt)
}

func TestNewField(t *testing.T) {
	f := NewField("wallet", "address")
	require.Equal(t, Field{Name: "wallet", Type: "address"}, f)
}
