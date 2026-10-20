// Package eip712 implements EIP-712 typed structured data hashing and signing.
package eip712

import (
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// eip712Prefix is the two-byte version prefix defined by EIP-712.
const eip712Prefix = "\x19\x01"

// Message is an EIP-712 struct instance's field values, keyed by field name -- the "message" (or
// a nested struct-typed field's own value) in the signed typed data.
type Message map[string]any

// Domain describes a signing domain per EIP-712. Only non-zero/non-nil fields are used.
type Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract *types.Address
	Salt              *types.Hash
}

// NewDomain returns a Domain with the given fields; any zero-valued field
// (e.g. a nil chainID) is omitted from the signed domain, per EIP-712.
func NewDomain(name, version string, chainID *big.Int, verifyingContract *types.Address, salt *types.Hash) *Domain {
	return &Domain{
		Name:              name,
		Version:           version,
		ChainID:           chainID,
		VerifyingContract: verifyingContract,
		Salt:              salt,
	}
}

// fieldsAndData returns domain's EIP-712 type fields and value map,
// including only the fields actually set.
func (d *Domain) fieldsAndData() ([]Field, Message) {
	var fields []Field
	data := make(Message)

	if d.Name != "" {
		fields = append(fields, NewField("name", "string"))
		data["name"] = d.Name
	}
	if d.Version != "" {
		fields = append(fields, NewField("version", "string"))
		data["version"] = d.Version
	}
	if d.ChainID != nil {
		fields = append(fields, NewField("chainId", "uint256"))
		data["chainId"] = d.ChainID
	}
	if d.VerifyingContract != nil {
		fields = append(fields, NewField("verifyingContract", "address"))
		data["verifyingContract"] = d.VerifyingContract
	}
	if d.Salt != nil {
		fields = append(fields, NewField("salt", "bytes32"))
		data["salt"] = d.Salt.Bytes()
	}
	return fields, data
}

// DomainSeparator returns hashStruct(eip712Domain) for domain. domain must
// have at least one field set, per EIP-712's requirement that the
// EIP712Domain type contain one or more members.
func DomainSeparator(domain *Domain) (*types.Hash, error) {
	fields, data := domain.fieldsAndData()
	if len(fields) == 0 {
		return nil, fmt.Errorf("eip712: domain has no fields set")
	}
	return NewTypes("EIP712Domain", fields...).HashStruct("EIP712Domain", data)
}

// Field is one named, typed field of an EIP-712 struct type.
type Field struct {
	Name string
	Type string
}

// NewField returns a Field named name with the given EIP-712 type string.
func NewField(name, typ string) Field {
	return Field{
		Name: name,
		Type: typ,
	}
}

// Types maps a custom struct type's name to its ordered fields.
type Types map[string][]Field

// NewTypes returns a Types schema with a single entry, name mapped to fields.
func NewTypes(name string, fields ...Field) Types {
	return Types{
		name: fields,
	}
}

// Append adds another entry to t, name mapped to fields.
func (t Types) Append(name string, fields ...Field) {
	t[name] = fields
}

// EncodeType returns primaryType's EIP-712 type encoding: its own
// definition, then every struct type it references (directly or
// transitively), sorted alphabetically.
func (t Types) EncodeType(primaryType string) string {
	deps := t.collectDependencies(primaryType, nil)
	if len(deps) > 1 {
		sort.Strings(deps[1:])
	}

	var sb strings.Builder
	for _, name := range deps {
		sb.WriteString(name)
		sb.WriteByte('(')
		for i, f := range t[name] {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(f.Type)
			sb.WriteByte(' ')
			sb.WriteString(f.Name)
		}
		sb.WriteByte(')')
	}
	return sb.String()
}

// TypeHash returns keccak256(EncodeType(primaryType)).
func (t Types) TypeHash(primaryType string) *types.Hash {
	return core.Keccak256([]byte(t.EncodeType(primaryType)))
}

// HashStruct returns keccak256(EncodeData(primaryType, data)).
func (t Types) HashStruct(primaryType string, data Message) (*types.Hash, error) {
	encoded, err := t.EncodeData(primaryType, data)
	if err != nil {
		return nil, err
	}
	return core.Keccak256(encoded), nil
}

func (t Types) collectDependencies(typeName string, found []string) []string {
	name := baseTypeName(typeName)
	if slices.Contains(found, name) {
		return found
	}
	if _, ok := t[name]; !ok {
		return found
	}
	found = append(found, name)
	for _, field := range t[name] {
		found = t.collectDependencies(field.Type, found)
	}
	return found
}

// EncodeData returns the EIP-712 encoding of a struct instance:
// typeHash ++ enc(value) for each field in order.
func (t Types) EncodeData(primaryType string, data Message) ([]byte, error) {
	buf := t.TypeHash(primaryType).Bytes()
	fields, err := t.encodeFields(primaryType, data)
	if err != nil {
		return nil, err
	}
	return append(buf, fields...), nil
}

// encodeFields encodes primaryType's fields in order, no typeHash prefix.
func (t Types) encodeFields(primaryType string, data Message) ([]byte, error) {
	var buf []byte
	for _, field := range t[primaryType] {
		encoded, err := t.encodeFieldValue(field.Type, data[field.Name])
		if err != nil {
			return nil, fmt.Errorf("eip712: field %q of %q: %w", field.Name, primaryType, err)
		}
		buf = append(buf, encoded...)
	}
	return buf, nil
}

// HashStructWithTypeHash is like HashStruct, but uses typeHash directly instead of deriving it
// from schema -- for a type signature with a caller-supplied opaque fragment (e.g. a "witness").
func (t Types) HashStructWithTypeHash(typeHash *types.Hash, primaryType string, data Message) (*types.Hash, error) {
	fields, err := t.encodeFields(primaryType, data)
	if err != nil {
		return nil, err
	}
	return core.Keccak256(append(typeHash.Bytes(), fields...)), nil
}

// encodeFieldValue encodes one field's value per its EIP-712 type: an
// array (recurses), a reference to another struct type (hashStruct),
// bytes/string (keccak256 of the raw data), or an atomic ABI scalar
// (encoded via the abi package, the same 32-byte word Pack would produce).
func (t Types) encodeFieldValue(typeName string, value any) ([]byte, error) {
	if strings.Contains(typeName, "[") {
		return t.encodeArrayValue(typeName, value)
	}
	if _, ok := t[typeName]; ok {
		var mapValue Message
		switch v := value.(type) {
		case Message:
			mapValue = v
		case map[string]any:
			mapValue = v
		default:
			return nil, fmt.Errorf("expected eip712.Message for %q, got %T", typeName, value)
		}
		encoded, err := t.EncodeData(typeName, mapValue)
		if err != nil {
			return nil, err
		}
		return core.Keccak256(encoded).Bytes(), nil
	}
	if typeName == "bytes" {
		b, ok := value.([]byte)
		if !ok {
			return nil, fmt.Errorf("expected []byte for bytes, got %T", value)
		}
		return core.Keccak256(b).Bytes(), nil
	}
	if typeName == "string" {
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		return core.Keccak256([]byte(s)).Bytes(), nil
	}

	abiType, err := abi.ParseType(typeName)
	if err != nil {
		return nil, err
	}
	return abi.Pack(abi.Types{abiType}, value)
}

// encodeArrayValue encodes an array/slice field as
// keccak256(concat(encodeFieldValue(elem) for each elem)). Nesting depth
// is inferred from the Go value's own shape, not from the type string, so
// a []byte element is treated as a fixed/dynamic bytes leaf rather than a
// nested array. This is ambiguous for "uint8[][]" specifically: Go's byte
// is an alias for uint8, so a []uint8 row is indistinguishable by
// reflection from a bytes-shaped leaf. Use uint256[] (or larger) for
// numeric arrays-of-arrays to avoid it.
func (t Types) encodeArrayValue(typeName string, value any) ([]byte, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected a slice or array for %q, got %T", typeName, value)
	}

	elemType := baseTypeName(typeName)
	var buf []byte
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		itemVal := reflect.ValueOf(item)

		var encoded []byte
		var err error
		if (itemVal.Kind() == reflect.Slice || itemVal.Kind() == reflect.Array) &&
			itemVal.Type().Elem().Kind() != reflect.Uint8 {
			encoded, err = t.encodeArrayValue(elemType, item)
		} else {
			encoded, err = t.encodeFieldValue(elemType, item)
		}
		if err != nil {
			return nil, err
		}
		buf = append(buf, encoded...)
	}
	return core.Keccak256(buf).Bytes(), nil
}

// baseTypeName strips any array suffix, e.g. "Person[3][]" -> "Person".
func baseTypeName(typeName string) string {
	if before, _, ok := strings.Cut(typeName, "["); ok {
		return before
	}
	return typeName
}

// Hash returns the final signable EIP-712 digest:
// keccak256("\x19\x01" ++ domainSeparator ++ hashStruct(message)).
func Hash(domain *Domain, primaryType string, schema Types, message Message) (*types.Hash, error) {
	msgHash, err := schema.HashStruct(primaryType, message)
	if err != nil {
		return nil, err
	}
	return FinalizeHash(domain, msgHash)
}

// FinalizeHash wraps a precomputed struct hash (e.g. from HashStructWithTypeHash) with domain.
func FinalizeHash(domain *Domain, structHash *types.Hash) (*types.Hash, error) {
	domainSep, err := DomainSeparator(domain)
	if err != nil {
		return nil, err
	}

	buf := append([]byte(eip712Prefix), domainSep.Bytes()...)
	buf = append(buf, structHash.Bytes()...)
	return core.Keccak256(buf), nil
}
