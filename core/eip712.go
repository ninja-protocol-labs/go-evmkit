package core

import (
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// EIP712Domain describes a signing domain per EIP-712. Only non-zero/non-nil fields are used.
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract *types.Address
	Salt              *types.Hash
}

// fieldsAndData returns domain's EIP-712 type fields and value map,
// including only the fields actually set.
func (d EIP712Domain) fieldsAndData() ([]EIP712Field, map[string]any) {
	var fields []EIP712Field
	data := make(map[string]any)

	if d.Name != "" {
		fields = append(fields, EIP712Field{Name: "name", Type: "string"})
		data["name"] = d.Name
	}
	if d.Version != "" {
		fields = append(fields, EIP712Field{Name: "version", Type: "string"})
		data["version"] = d.Version
	}
	if d.ChainID != nil {
		fields = append(fields, EIP712Field{Name: "chainId", Type: "uint256"})
		data["chainId"] = d.ChainID
	}
	if d.VerifyingContract != nil {
		fields = append(fields, EIP712Field{Name: "verifyingContract", Type: "address"})
		data["verifyingContract"] = d.VerifyingContract
	}
	if d.Salt != nil {
		fields = append(fields, EIP712Field{Name: "salt", Type: "bytes32"})
		data["salt"] = d.Salt.Bytes()
	}
	return fields, data
}

// DomainSeparator returns hashStruct(eip712Domain) for domain.
func DomainSeparator(domain EIP712Domain) (*types.Hash, error) {
	fields, data := domain.fieldsAndData()
	return EIP712Types{"EIP712Domain": fields}.HashStruct("EIP712Domain", data)
}

// EIP712Field is one named, typed field of an EIP-712 struct type.
type EIP712Field struct {
	Name string
	Type string
}

// EIP712Types maps a custom struct type's name to its ordered fields.
type EIP712Types map[string][]EIP712Field

// EncodeType returns primaryType's EIP-712 type encoding: its own
// definition, then every struct type it references (directly or
// transitively), sorted alphabetically.
func (t EIP712Types) EncodeType(primaryType string) string {
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
func (t EIP712Types) TypeHash(primaryType string) *types.Hash {
	return Keccak256([]byte(t.EncodeType(primaryType)))
}

// HashStruct returns keccak256(EncodeData(primaryType, data)).
func (t EIP712Types) HashStruct(primaryType string, data map[string]any) (*types.Hash, error) {
	encoded, err := t.EncodeData(primaryType, data)
	if err != nil {
		return nil, err
	}
	return Keccak256(encoded), nil
}

func (t EIP712Types) collectDependencies(typeName string, found []string) []string {
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

// baseTypeName strips any array suffix, e.g. "Person[3][]" -> "Person".
func baseTypeName(typeName string) string {
	if before, _, ok := strings.Cut(typeName, "["); ok {
		return before
	}
	return typeName
}

// EncodeData returns the EIP-712 encoding of a struct instance:
// typeHash ++ enc(value) for each field in order.
func (t EIP712Types) EncodeData(primaryType string, data map[string]any) ([]byte, error) {
	buf := t.TypeHash(primaryType).Bytes()
	for _, field := range t[primaryType] {
		encoded, err := t.encodeFieldValue(field.Type, data[field.Name])
		if err != nil {
			return nil, fmt.Errorf("eip712: field %q of %q: %w", field.Name, primaryType, err)
		}
		buf = append(buf, encoded...)
	}
	return buf, nil
}

// encodeFieldValue encodes one field's value per its EIP-712 type: an
// array (recurses), a reference to another struct type (hashStruct),
// bytes/string (keccak256 of the raw data), or an atomic ABI scalar
// (encoded via the abi package, the same 32-byte word Pack would produce).
func (t EIP712Types) encodeFieldValue(typeName string, value any) ([]byte, error) {
	if strings.Contains(typeName, "[") {
		return t.encodeArrayValue(typeName, value)
	}
	if _, ok := t[typeName]; ok {
		mapValue, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map[string]any for %q, got %T", typeName, value)
		}
		encoded, err := t.EncodeData(typeName, mapValue)
		if err != nil {
			return nil, err
		}
		return Keccak256(encoded).Bytes(), nil
	}
	if typeName == "bytes" {
		b, ok := value.([]byte)
		if !ok {
			return nil, fmt.Errorf("expected []byte for bytes, got %T", value)
		}
		return Keccak256(b).Bytes(), nil
	}
	if typeName == "string" {
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		return Keccak256([]byte(s)).Bytes(), nil
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
func (t EIP712Types) encodeArrayValue(typeName string, value any) ([]byte, error) {
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
	return Keccak256(buf).Bytes(), nil
}
