package types

import (
	"math/big"
	"testing"
)

func TestNewAddressFromBytes(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want [AddressLength]byte
	}{
		{"exact length", make([]byte, AddressLength), [AddressLength]byte{}},
		{
			"shorter than length is left-padded",
			[]byte{0x01, 0x02},
			[AddressLength]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01, 0x02},
		},
		{
			"longer than length keeps trailing bytes",
			append([]byte{0xFF}, make([]byte, AddressLength)...),
			[AddressLength]byte{},
		},
		{"empty", nil, [AddressLength]byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAddressFromBytes(tt.in)
			if a.bytes != tt.want {
				t.Errorf("got %x want %x", a.bytes, tt.want)
			}
		})
	}
}

func TestNewAddressFromHex(t *testing.T) {
	valid := "5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"with 0x prefix", "0x" + valid, false},
		{"with 0X prefix", "0X" + valid, false},
		{"without prefix", valid, false},
		{"too short", "0x1234", true},
		{"too long", "0x" + valid + "00", true},
		{"invalid hex chars", "0x" + "zz" + valid[2:], true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAddressFromHex(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.String() != "0x"+valid {
				t.Errorf("got %s want %s", a.String(), "0x"+valid)
			}
		})
	}
}

func TestNewAddressFromBig(t *testing.T) {
	n := big.NewInt(0x1234)
	a := NewAddressFromBig(n)
	want := NewAddressFromBytes([]byte{0x12, 0x34})
	if !a.Equal(want) {
		t.Errorf("got %s want %s", a.String(), want.String())
	}
}

func TestIsHexAddress(t *testing.T) {
	valid := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid with prefix", valid, true},
		{"valid without prefix", valid[2:], true},
		{"too short", "0x1234", false},
		{"invalid chars", "0xzz" + valid[4:], false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHexAddress(tt.in); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestAddressBytes(t *testing.T) {
	in := make([]byte, AddressLength)
	in[0] = 0xAB
	a := NewAddressFromBytes(in)

	got := a.Bytes()
	if len(got) != AddressLength {
		t.Fatalf("length = %d, want %d", len(got), AddressLength)
	}
	got[0] = 0xFF
	if a.bytes[0] != 0xAB {
		t.Errorf("Bytes() must return a copy, mutation leaked into Address")
	}
}

func TestAddressIsZero(t *testing.T) {
	var nilAddr *Address
	if !nilAddr.IsZero() {
		t.Error("nil address should be zero")
	}

	zero := NewAddressFromBytes(nil)
	if !zero.IsZero() {
		t.Error("all-zero address should be zero")
	}

	nonZero := NewAddressFromBytes([]byte{0x01})
	if nonZero.IsZero() {
		t.Error("non-zero address reported as zero")
	}
}

func TestAddressEqual(t *testing.T) {
	a1 := NewAddressFromBytes([]byte{0x01, 0x02})
	a2 := NewAddressFromBytes([]byte{0x01, 0x02})
	a3 := NewAddressFromBytes([]byte{0x03, 0x04})
	var nilAddr *Address

	if !a1.Equal(a2) {
		t.Error("equal addresses reported as different")
	}
	if a1.Equal(a3) {
		t.Error("different addresses reported as equal")
	}
	if a1.Equal(nilAddr) {
		t.Error("non-nil address equal to nil address")
	}
	if !nilAddr.Equal(nil) {
		t.Error("nil address should equal nil")
	}
}

func TestAddressBig(t *testing.T) {
	a := NewAddressFromBytes([]byte{0x01, 0x02})
	want := new(big.Int).SetBytes([]byte{0x01, 0x02})
	if a.Big().Cmp(want) != 0 {
		t.Errorf("got %s want %s", a.Big().String(), want.String())
	}
}

func TestAddressStringChecksum(t *testing.T) {
	vectors := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
		"0x52908400098527886E0F7030069857D2E4169EE7",
		"0x8617E340B3D01FA5F11F306F4090FD50E238070D",
		"0xde709f2102306220921060314715629080e2fb77",
		"0x27b1fdb04752bbc536007a920d24acb045561c26",
	}

	for _, v := range vectors {
		a, err := NewAddressFromHex(v)
		if err != nil {
			t.Fatalf("%s: %v", v, err)
		}
		if got := a.String(); got != v {
			t.Errorf("got %s want %s", got, v)
		}
	}
}

func TestAddressStringCached(t *testing.T) {
	a := NewAddressFromBytes([]byte{0x01})
	first := a.String()
	second := a.String()
	if first != second {
		t.Errorf("cached String() mismatch: %s vs %s", first, second)
	}
}
