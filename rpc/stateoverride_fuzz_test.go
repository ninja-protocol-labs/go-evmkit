package rpc

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzStateOverrideMarshalJSON$' -fuzztime=20s ./rpc
func FuzzStateOverrideMarshalJSON(f *testing.F) {
	f.Add(
		[]byte{0x11},
		false,
		[]byte{0x0d, 0xe0},
		false,
		uint64(0),
		[]byte{},
		true,
		[]byte{0x01},
		[]byte{0x2a},
		false,
		true,
	)
	f.Add(
		make([]byte, 20),
		true,
		make([]byte, 33),
		true,
		^uint64(0),
		[]byte(nil),
		false,
		[]byte{},
		[]byte{},
		true,
		true,
	)
	f.Add(
		[]byte{0xff, 0xee},
		false,
		[]byte{0x01},
		false,
		uint64(7),
		[]byte{0x60, 0x00},
		true,
		make([]byte, 40),
		make([]byte, 32),
		true,
		false,
	)
	f.Add([]byte{}, true, []byte{}, false, uint64(1), []byte{0x00}, false, []byte{0xaa}, []byte{0xbb}, false, false)

	f.Fuzz(func(
		t *testing.T,
		addr []byte,
		nilAccount bool,
		balanceBytes []byte,
		negative bool,
		nonce uint64,
		code []byte,
		setCode bool,
		slotBytes,
		valueBytes []byte,
		useState,
		useDiff bool,
	) {
		var account *types.Address
		if !nilAccount {
			account = types.NewAddressFromBytes(addr)
		}
		balance := new(big.Int).SetBytes(balanceBytes)
		if negative {
			balance.Neg(balance)
		}
		var slot, value [32]byte
		copy(slot[32-min(len(slotBytes), 32):], slotBytes)
		copy(value[32-min(len(valueBytes), 32):], valueBytes)

		override := NewStateOverride().WithBalance(account, balance).WithNonce(account, nonce)
		if setCode {
			override.WithCode(account, code)
		}
		if useState {
			override.WithState(account, slot, value)
		}
		if useDiff {
			override.WithStateDiff(account, slot, value)
		}

		wantInvalid := nilAccount || balance.Sign() < 0 || balance.BitLen() > 256 || (useState && useDiff)
		validateErr := override.Validate()
		out, marshalErr := json.Marshal(override)
		if wantInvalid {
			if validateErr == nil || marshalErr == nil {
				t.Fatalf("invalid override accepted: validate=%v marshal=%v", validateErr, marshalErr)
			}
			if !errors.Is(validateErr, ErrInvalidStateOverride) || !errors.Is(marshalErr, ErrInvalidStateOverride) {
				t.Fatalf("marshal error does not carry sentinel: %v", marshalErr)
			}
			return
		}
		if validateErr != nil || marshalErr != nil {
			t.Fatalf("valid override rejected: validate=%v marshal=%v", validateErr, marshalErr)
		}

		var decoded map[string]map[string]any
		if err := json.Unmarshal(out, &decoded); err != nil {
			t.Fatalf("output is not an account map: %v: %s", err, out)
		}
		key := encoding.Hex.EncodePrefixed(account.Bytes())
		fields, ok := decoded[key]
		if len(decoded) != 1 || !ok {
			t.Fatalf("want only account %s, got %s", key, out)
		}

		gotBalance, ok := new(big.Int).SetString(strings.TrimPrefix(fields["balance"].(string), "0x"), 16)
		if !ok || gotBalance.Cmp(balance) != 0 {
			t.Fatalf("balance %v round-tripped to %v", balance, fields["balance"])
		}
		gotNonce, err := strconv.ParseUint(strings.TrimPrefix(fields["nonce"].(string), "0x"), 16, 64)
		if err != nil || gotNonce != nonce {
			t.Fatalf("nonce %d round-tripped to %v", nonce, fields["nonce"])
		}
		_, hasCode := fields["code"]
		if hasCode != (setCode && code != nil) {
			t.Fatalf("code presence %v, setCode=%v code=%v", hasCode, setCode, code)
		}
		if hasCode && fields["code"] != encoding.Hex.EncodePrefixed(code) {
			t.Fatalf("code %x round-tripped to %v", code, fields["code"])
		}
		wantStorage := map[string]any{encoding.Hex.EncodePrefixed(slot[:]): encoding.Hex.EncodePrefixed(value[:])}
		for name, set := range map[string]bool{"state": useState, "stateDiff": useDiff} {
			got, has := fields[name]
			if has != set {
				t.Fatalf("%s presence %v, want %v", name, has, set)
			}
			if has && len(got.(map[string]any)) != 1 {
				t.Fatalf("%s has %d slots, want 1", name, len(got.(map[string]any)))
			}
			slotKey := encoding.Hex.EncodePrefixed(slot[:])
			if has && got.(map[string]any)[slotKey] != wantStorage[slotKey] {
				t.Fatalf("%s slot round-tripped to %v", name, got)
			}
		}
	})
}
