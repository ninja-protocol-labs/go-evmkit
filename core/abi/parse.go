package abi

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseType parses a canonical ABI type string, e.g. "uint256", "address[]",
// or "(address,uint256)[]" — the inverse of Type.String().
func ParseType(s string) (Type, error) {
	s = strings.TrimSpace(s)
	t, pos, err := parseTypeExpr(s, 0)
	if err != nil {
		return Type{}, err
	}
	pos = skipSpace(s, pos)
	if pos != len(s) {
		return Type{}, fmt.Errorf("%w: unexpected trailing characters in %q", ErrInvalidTypeString, s)
	}
	return t, nil
}

// ParseFunction parses a Solidity function signature into a Function, with
// or without per-parameter names (e.g. "transfer(address,uint256)" or
// "transfer(address to, uint256 amount)" both work; names are accepted but
// not kept, since they don't affect encoding).
//
// Outputs can be given two ways: passed directly via the outputs
// parameter, or embedded in the signature itself after the input list,
// with or without a "returns" keyword and with or without parentheses
// (e.g. "transfer(address,uint256) bool", "transfer(address,uint256)
// returns (bool)"). Outputs embedded in the signature take precedence
// over the outputs parameter.
//
// A leading "function" keyword and visibility/mutability modifiers
// (external, public, internal, private, pure, view, constant, payable,
// virtual, override) between the input list and the outputs are accepted
// and discarded, so a full Solidity declaration can be pasted as-is.
func ParseFunction(signature string, outputs Types) (*Function, error) {
	sig := strings.TrimSpace(signature)
	if rest, ok := stripPrefix(sig, "function"); ok && (rest == "" || !isIdentChar(rest[0])) {
		sig = strings.TrimSpace(rest)
	}

	idx := strings.IndexByte(sig, '(')
	if idx < 0 {
		return nil, fmt.Errorf("%w: missing '(' in %q", ErrInvalidTypeString, signature)
	}
	name := strings.TrimSpace(sig[:idx])
	if name == "" {
		return nil, fmt.Errorf("%w: missing function name in %q", ErrInvalidTypeString, signature)
	}

	inputs, pos, err := parseParamList(sig, idx)
	if err != nil {
		return nil, err
	}
	pos = skipModifiers(sig, pos)

	if pos < len(sig) {
		outputs, pos, err = parseOutputs(sig, pos)
		if err != nil {
			return nil, err
		}
		pos = skipSpace(sig, pos)
	}
	if pos != len(sig) {
		return nil, fmt.Errorf("%w: unexpected trailing characters in %q", ErrInvalidTypeString, signature)
	}

	return NewFunction(name, inputs, outputs), nil
}

// parseTypeExpr parses one type expression (scalar or tuple, plus any
// trailing array suffixes) starting at pos, returning the position just
// past it.
func parseTypeExpr(s string, pos int) (Type, int, error) {
	pos = skipSpace(s, pos)
	if pos >= len(s) {
		return Type{}, pos, fmt.Errorf("%w: unexpected end of type", ErrInvalidTypeString)
	}

	var t Type
	if s[pos] == '(' {
		components, newPos, err := parseParamList(s, pos)
		if err != nil {
			return Type{}, newPos, err
		}
		pos = newPos
		t = Tuple(components...)
	} else {
		start := pos
		for pos < len(s) && isIdentChar(s[pos]) {
			pos++
		}
		if pos == start {
			return Type{}, pos, fmt.Errorf("%w: expected a type at %d in %q", ErrInvalidTypeString, pos, s)
		}
		var err error
		t, err = baseType(s[start:pos])
		if err != nil {
			return Type{}, pos, err
		}
	}

	for pos < len(s) && s[pos] == '[' {
		pos++
		start := pos
		for pos < len(s) && s[pos] >= '0' && s[pos] <= '9' {
			pos++
		}
		if pos >= len(s) || s[pos] != ']' {
			return Type{}, pos, fmt.Errorf("%w: unterminated array suffix in %q", ErrInvalidTypeString, s)
		}
		digits := s[start:pos]
		pos++

		if digits == "" {
			t = Slice(t)
			continue
		}
		size, convErr := strconv.Atoi(digits)
		if convErr != nil {
			return Type{}, pos, fmt.Errorf("%w: invalid array size %q", ErrInvalidTypeString, digits)
		}
		t, _ = Array(t, size) // size >= 0: digits is all-decimal, so Atoi can't return negative
	}

	return t, pos, nil
}

// skipModifiers skips a whitespace-separated run of words between a
// function's input list and its outputs: visibility/mutability keywords
// (external, view, payable, ...) and arbitrary custom modifiers (e.g.
// onlyOwner) alike. There's no fixed list of custom modifier names to
// check against, so instead each word is tried against baseType: if it
// parses as a real type, it must be the start of a bare output (stop
// here); if it's "returns", hand off to parseOutputs; otherwise it's some
// modifier (built-in or custom) and gets skipped.
func skipModifiers(s string, pos int) int {
	for {
		p := skipSpace(s, pos)
		if p >= len(s) || s[p] == '(' {
			return p
		}

		start := p
		for p < len(s) && isIdentChar(s[p]) {
			p++
		}
		if start == p {
			return p
		}

		word := s[start:p]
		if word == "returns" {
			return start
		}
		if _, err := baseType(word); err == nil {
			return start
		}
		pos = p
	}
}

// parseOutputs parses the trailing output spec after a function's input
// list: an optional "returns" keyword, then either a single bare type
// ("bool") or a parenthesized list ("(bool,uint256)").
func parseOutputs(s string, pos int) (Types, int, error) {
	const kw = "returns"
	if strings.HasPrefix(s[pos:], kw) {
		after := pos + len(kw)
		if after >= len(s) || !isIdentChar(s[after]) {
			pos = skipSpace(s, after)
		}
	}

	if pos >= len(s) {
		return nil, pos, fmt.Errorf("%w: expected output type(s) in %q", ErrInvalidTypeString, s)
	}
	if s[pos] == '(' {
		return parseParamList(s, pos)
	}

	t, newPos, err := parseTypeExpr(s, pos)
	if err != nil {
		return nil, newPos, err
	}
	return Types{t}, newPos, nil
}

// parseParamList parses a parenthesized, comma-separated parameter list
// starting at pos (where s[pos] == '('), returning the component types and
// the position just past the closing ')'.
func parseParamList(s string, pos int) (Types, int, error) {
	if pos >= len(s) || s[pos] != '(' {
		return nil, pos, fmt.Errorf("%w: expected '(' at %d in %q", ErrInvalidTypeString, pos, s)
	}
	pos = skipSpace(s, pos+1)

	var components Types
	if pos < len(s) && s[pos] == ')' {
		return components, pos + 1, nil
	}

	for {
		t, newPos, err := parseParam(s, pos)
		if err != nil {
			return nil, newPos, err
		}
		pos = newPos
		components = append(components, t)

		pos = skipSpace(s, pos)
		if pos >= len(s) {
			return nil, pos, fmt.Errorf("%w: unterminated parameter list in %q", ErrInvalidTypeString, s)
		}
		switch s[pos] {
		case ',':
			pos = skipSpace(s, pos+1)
		case ')':
			return components, pos + 1, nil
		default:
			return nil, pos, fmt.Errorf("%w: expected ',' or ')' at %d in %q", ErrInvalidTypeString, pos, s)
		}
	}
}

// parseParam parses a single parameter: a type, optionally followed by up
// to two more identifiers — a data location ("memory", "storage",
// "calldata") and/or a name, in that order, e.g. "uint256 amount",
// "address[] memory path", or "address[] memory" (location, no name).
// Which keywords Solidity recognizes as data locations doesn't matter
// here: everything after the type is documentation only and gets
// discarded regardless, so capping at two trailing words (Solidity's
// grammar never allows more) is enough without naming them.
func parseParam(s string, pos int) (Type, int, error) {
	t, pos, err := parseTypeExpr(s, pos)
	if err != nil {
		return Type{}, pos, err
	}

	for range 2 {
		newPos, word := skipIdentWord(s, skipSpace(s, pos))
		if word == "" {
			break
		}
		pos = newPos
	}
	return t, pos, nil
}

// baseType maps a scalar type word (e.g. "uint256", "bytes32") to its Type.
func baseType(word string) (Type, error) {
	switch word {
	case "bool":
		return Bool, nil
	case "address":
		return Address, nil
	case "string":
		return String, nil
	case "bytes":
		return Bytes, nil
	case "function":
		return FunctionType, nil
	}

	// The remainder after stripping "uint"/"int"/"bytes" must be all digits
	// (or empty) to count as a match: a word like "internal" starts with
	// "int" too, but "ernal" isn't a size suffix, so it must fall through
	// to the unknown-type error below rather than a misleading size error.
	if rest, ok := stripPrefix(word, "uint"); ok && (rest == "" || isAllDigits(rest)) {
		size, err := parseBitSize(rest)
		if err != nil {
			return Type{}, fmt.Errorf("%w: %s", err, word)
		}
		return Type{Kind: KindUint, Size: size}, nil
	}
	if rest, ok := stripPrefix(word, "int"); ok && (rest == "" || isAllDigits(rest)) {
		size, err := parseBitSize(rest)
		if err != nil {
			return Type{}, fmt.Errorf("%w: %s", err, word)
		}
		return Type{Kind: KindInt, Size: size}, nil
	}
	if rest, ok := stripPrefix(word, "bytes"); ok && rest != "" && isAllDigits(rest) {
		n, err := strconv.Atoi(rest)
		if err != nil || n < 1 || n > 32 {
			return Type{}, fmt.Errorf("%w: invalid bytes size in %q", ErrInvalidTypeString, word)
		}
		return Type{Kind: KindFixedBytes, Size: n}, nil
	}

	return Type{}, fmt.Errorf("%w: unknown type %q", ErrInvalidTypeString, word)
}

// parseBitSize parses the numeric suffix of a uintN/intN type word; a bare
// "uint"/"int" (no suffix) defaults to 256, matching Solidity's alias.
func parseBitSize(rest string) (int, error) {
	if rest == "" {
		return 256, nil
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 8 || n > 256 || n%8 != 0 {
		return 0, fmt.Errorf("%w: invalid bit size %q", ErrInvalidTypeString, rest)
	}
	return n, nil
}

func stripPrefix(word, prefix string) (string, bool) {
	if !strings.HasPrefix(word, prefix) {
		return "", false
	}
	return word[len(prefix):], true
}

func skipSpace(s string, pos int) int {
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}
	return pos
}

func isIdentChar(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// skipIdentWord skips a run of identifier characters starting at pos
// (which must already be past any leading whitespace), returning the new
// position and the word skipped ("" if pos wasn't at an identifier).
func skipIdentWord(s string, pos int) (int, string) {
	start := pos
	for pos < len(s) && isIdentChar(s[pos]) {
		pos++
	}
	return pos, s[start:pos]
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}
