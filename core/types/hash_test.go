package types

import (
	"math/big"
	"testing"
)

func TestNewHashFromBytes(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want [HashLength]byte
	}{
		{"exact length", make([]byte, HashLength), [HashLength]byte{}},
		{
			"shorter than length is left-padded",
			[]byte{0x01, 0x02},
			func() [HashLength]byte {
				var b [HashLength]byte
				b[HashLength-2] = 0x01
				b[HashLength-1] = 0x02
				return b
			}(),
		},
		{
			"longer than length keeps trailing bytes",
			append([]byte{0xFF}, make([]byte, HashLength)...),
			[HashLength]byte{},
		},
		{"empty", nil, [HashLength]byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHashFromBytes(tt.in)
			if h.bytes != tt.want {
				t.Errorf("got %x want %x", h.bytes, tt.want)
			}
		})
	}
}

func TestNewHashFromHex(t *testing.T) {
	valid := "5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed5aAeb6053F3E94C9b9A09f33"

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
			h, err := NewHashFromHex(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := "0x" + toLowerHex(valid)
			if h.String() != want {
				t.Errorf("got %s want %s", h.String(), want)
			}
		})
	}
}

func toLowerHex(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'F' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func TestNewHashFromBig(t *testing.T) {
	n := big.NewInt(0x1234)
	h := NewHashFromBig(n)
	want := NewHashFromBytes([]byte{0x12, 0x34})
	if !h.Equal(want) {
		t.Errorf("got %s want %s", h.String(), want.String())
	}
}

func TestIsHexHash(t *testing.T) {
	valid := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed5aAeb6053F3E94C9b9A09f33"

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
			if got := IsHexHash(tt.in); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestHashBytes(t *testing.T) {
	in := make([]byte, HashLength)
	in[0] = 0xAB
	h := NewHashFromBytes(in)

	got := h.Bytes()
	if len(got) != HashLength {
		t.Fatalf("length = %d, want %d", len(got), HashLength)
	}
	got[0] = 0xFF
	if h.bytes[0] != 0xAB {
		t.Errorf("Bytes() must return a copy, mutation leaked into Hash")
	}
}

func TestHashIsZero(t *testing.T) {
	var nilHash *Hash
	if !nilHash.IsZero() {
		t.Error("nil hash should be zero")
	}

	zero := NewHashFromBytes(nil)
	if !zero.IsZero() {
		t.Error("all-zero hash should be zero")
	}

	nonZero := NewHashFromBytes([]byte{0x01})
	if nonZero.IsZero() {
		t.Error("non-zero hash reported as zero")
	}
}

func TestHashEqual(t *testing.T) {
	h1 := NewHashFromBytes([]byte{0x01, 0x02})
	h2 := NewHashFromBytes([]byte{0x01, 0x02})
	h3 := NewHashFromBytes([]byte{0x03, 0x04})
	var nilHash *Hash

	if !h1.Equal(h2) {
		t.Error("equal hashes reported as different")
	}
	if h1.Equal(h3) {
		t.Error("different hashes reported as equal")
	}
	if h1.Equal(nilHash) {
		t.Error("non-nil hash equal to nil hash")
	}
	if !nilHash.Equal(nil) {
		t.Error("nil hash should equal nil")
	}
}

func TestHashBig(t *testing.T) {
	h := NewHashFromBytes([]byte{0x01, 0x02})
	want := new(big.Int).SetBytes([]byte{0x01, 0x02})
	if h.Big().Cmp(want) != 0 {
		t.Errorf("got %s want %s", h.Big().String(), want.String())
	}
}

func TestHashStringCached(t *testing.T) {
	h := NewHashFromBytes([]byte{0x01})
	first := h.String()
	second := h.String()
	if first != second {
		t.Errorf("cached String() mismatch: %s vs %s", first, second)
	}
}
