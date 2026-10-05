package core

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestEIP712EncodeTypeMail(t *testing.T) {
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}
	got := ts.EncodeType("Mail")
	require.Equal(t, "Mail(Person from,Person to,string contents)Person(string name,address wallet)", got)
}

func TestEIP712TypeHashMail(t *testing.T) {
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}
	got := ts.TypeHash("Mail")
	want, err := hex.DecodeString("a0cedeb2dc280ba39b857546d74f5549c3a1d7bdc2dd96bf881f76108e23dac2")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712EncodeTypeTransaction(t *testing.T) {
	ts := EIP712Types{
		"Transaction": {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "tx", Type: "Asset"}},
		"Asset":       {{Name: "token", Type: "address"}, {Name: "amount", Type: "uint256"}},
		"Person":      {{Name: "wallet", Type: "address"}, {Name: "name", Type: "string"}},
	}
	got := ts.EncodeType("Transaction")
	require.Equal(t, "Transaction(Person from,Person to,Asset tx)Asset(address token,uint256 amount)Person(address wallet,string name)", got)
}

func TestEIP712TypeHashTransaction(t *testing.T) {
	ts := EIP712Types{
		"Transaction": {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "tx", Type: "Asset"}},
		"Asset":       {{Name: "token", Type: "address"}, {Name: "amount", Type: "uint256"}},
		"Person":      {{Name: "wallet", Type: "address"}, {Name: "name", Type: "string"}},
	}
	got := ts.TypeHash("Transaction")
	want, err := hex.DecodeString("358262ad2b1b6af9edb8b4f81ee9a13ec2ed2473132bcfe1721ac7a2e191791e")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712EncodeTypeNoReferences(t *testing.T) {
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
	}
	require.Equal(t, "Person(string name,address wallet)", ts.EncodeType("Person"))
}

func TestEIP712EncodeTypeArrayReference(t *testing.T) {
	ts := EIP712Types{
		"Group":  {{Name: "members", Type: "Person[]"}},
		"Person": {{Name: "name", Type: "string"}},
	}
	got := ts.EncodeType("Group")
	require.Equal(t, "Group(Person[] members)Person(string name)", got)
}

func TestEIP712EncodeTypeUnknownType(t *testing.T) {
	ts := EIP712Types{}
	require.Equal(t, "", ts.EncodeType("Nonexistent"))
}

func TestEIP712BaseTypeName(t *testing.T) {
	require.Equal(t, "Person", baseTypeName("Person"))
	require.Equal(t, "Person", baseTypeName("Person[]"))
	require.Equal(t, "Person", baseTypeName("Person[3]"))
	require.Equal(t, "Person", baseTypeName("Person[3][]"))
}

func TestEIP712EncodeTypeMutualCycle(t *testing.T) {
	ts := EIP712Types{
		"A": {{Name: "b", Type: "B"}},
		"B": {{Name: "a", Type: "A"}},
	}
	got := ts.EncodeType("A")
	require.Equal(t, "A(B b)B(A a)", got)
}

func TestEIP712EncodeTypeSelfReference(t *testing.T) {
	ts := EIP712Types{
		"Node": {{Name: "value", Type: "uint256"}, {Name: "children", Type: "Node[]"}},
	}
	got := ts.EncodeType("Node")
	require.Equal(t, "Node(uint256 value,Node[] children)", got)
}

func TestEIP712EncodeTypeDiamond(t *testing.T) {
	ts := EIP712Types{
		"Top":    {{Name: "l", Type: "Left"}, {Name: "r", Type: "Right"}},
		"Left":   {{Name: "shared", Type: "Shared"}},
		"Right":  {{Name: "shared", Type: "Shared"}},
		"Shared": {{Name: "value", Type: "uint256"}},
	}
	got := ts.EncodeType("Top")
	require.Equal(t, "Top(Left l,Right r)Left(Shared shared)Right(Shared shared)Shared(uint256 value)", got)
}

func TestEIP712HashStructPerson(t *testing.T) {
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}
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
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}
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
	ts := EIP712Types{"Simple": {{Name: "a", Type: "uint256"}}}
	_, err := ts.EncodeData("Simple", map[string]any{"a": big.NewInt(1), "extra": "ignored"})
	require.NoError(t, err)
}

func TestEIP712EncodeDataMissingFieldErrors(t *testing.T) {
	ts := EIP712Types{"Simple": {{Name: "a", Type: "address"}}}
	_, err := ts.EncodeData("Simple", map[string]any{})
	require.Error(t, err)
}

func TestEIP712EncodeDataArrayOfStructs(t *testing.T) {
	ts := EIP712Types{
		"Group":  {{Name: "members", Type: "Person[]"}},
		"Person": {{Name: "name", Type: "string"}},
	}
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
	ts := EIP712Types{"Batch": {{Name: "ids", Type: "bytes32[]"}}}
	id1 := make([]byte, 32)
	id1[0] = 0x01
	id2 := make([]byte, 32)
	id2[0] = 0x02

	got, err := ts.EncodeData("Batch", map[string]any{"ids": [][]byte{id1, id2}})
	require.NoError(t, err)
	require.Len(t, got, 64)
}

func TestEIP712EncodeDataUnknownFieldTypeErrors(t *testing.T) {
	ts := EIP712Types{"Bad": {{Name: "x", Type: "notarealtype"}}}
	_, err := ts.EncodeData("Bad", map[string]any{"x": "whatever"})
	require.Error(t, err)
}

func TestEIP712EncodeDataArrayWrongGoTypeErrors(t *testing.T) {
	ts := EIP712Types{"Bad": {{Name: "x", Type: "uint256[]"}}}
	_, err := ts.EncodeData("Bad", map[string]any{"x": "not a slice"})
	require.Error(t, err)
}

func etherMailDomain(t *testing.T) EIP712Domain {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	return EIP712Domain{
		Name:              "Ether Mail",
		Version:           "1",
		ChainID:           big.NewInt(1),
		VerifyingContract: contract,
	}
}

func TestEIP712DomainSeparator(t *testing.T) {
	got, err := DomainSeparator(etherMailDomain(t))
	require.NoError(t, err)
	want, err := hex.DecodeString("f2cee375fa42b42143804025fc449deafd50cc031ca257e0b194a650a912090f")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712DomainSeparatorOmitsUnsetFields(t *testing.T) {
	domain := EIP712Domain{Name: "Only Name"}
	fields, data := domain.fieldsAndData()
	require.Len(t, fields, 1)
	require.Equal(t, "name", fields[0].Name)
	require.Len(t, data, 1)
}

func TestEIP712Hash(t *testing.T) {
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
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}

	got, err := EIP712Hash(etherMailDomain(t), "Mail", ts, mail)
	require.NoError(t, err)
	want, err := hex.DecodeString("be609aee343fb3c4b28e1df9e632fca64fcfaede20f02e86244efddf30957bd2")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP712HashErrorsOnBadMessage(t *testing.T) {
	ts := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail":   {{Name: "from", Type: "Person"}, {Name: "to", Type: "Person"}, {Name: "contents", Type: "string"}},
	}
	_, err := EIP712Hash(etherMailDomain(t), "Mail", ts, map[string]any{})
	require.Error(t, err)
}

func TestEIP712DomainSeparatorWithSalt(t *testing.T) {
	contract, err := types.NewAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	require.NoError(t, err)
	salt := types.NewHashFromBytes([]byte("some salt value"))

	domain := EIP712Domain{
		Name:              "Ether Mail",
		Version:           "1",
		ChainID:           big.NewInt(1),
		VerifyingContract: contract,
		Salt:              salt,
	}
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
	ts := EIP712Types{"Simple": {{Name: "ok", Type: "bool"}}}

	got, err := ts.EncodeData("Simple", map[string]any{"ok": true})
	require.NoError(t, err)
	require.Len(t, got, 64)

	gotFalse, err := ts.EncodeData("Simple", map[string]any{"ok": false})
	require.NoError(t, err)
	require.NotEqual(t, got, gotFalse)
}

func TestEIP712EncodeDataNegativeInt256Field(t *testing.T) {
	ts := EIP712Types{"Simple": {{Name: "delta", Type: "int256"}}}

	got, err := ts.EncodeData("Simple", map[string]any{"delta": big.NewInt(-42)})
	require.NoError(t, err)
	require.Len(t, got, 64)
	require.Equal(t, byte(0xff), got[32])
}

func TestEIP712EncodeDataSelfReferencingTree(t *testing.T) {
	ts := EIP712Types{
		"Node": {{Name: "value", Type: "uint256"}, {Name: "children", Type: "Node[]"}},
	}
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
	ts := EIP712Types{"Matrix": {{Name: "rows", Type: "uint256[][]"}}}

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
	ts := EIP712Types{"Matrix": {{Name: "rows", Type: "uint8[][]"}}}
	_, err := ts.EncodeData("Matrix", map[string]any{
		"rows": [][]uint8{{1, 2}, {3, 4, 5}},
	})
	require.Error(t, err)
}
