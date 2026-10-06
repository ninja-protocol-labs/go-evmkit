package core

import (
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ninja-protocol-labs/go-evmkit/core/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/stretchr/testify/require"
)

const legacyTestKey = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

type legacyVector struct {
	name    string
	tx      LegacyTx
	sigHash string
	raw     string
	hash    string
	sig     string
	recID   byte
}

func legacyTestAddress(t *testing.T) *types.Address {
	t.Helper()
	a, err := types.NewAddressFromHex("0x3535353535353535353535353535353535353535")
	require.NoError(t, err)
	return a
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func legacyVectors(t *testing.T) []legacyVector {
	to := legacyTestAddress(t)
	oneEther, ok := new(big.Int).SetString("1000000000000000000", 10)
	require.True(t, ok)

	return []legacyVector{
		{
			name: "mainnet transfer",
			tx: LegacyTx{
				ChainID:  big.NewInt(1),
				Nonce:    9,
				GasPrice: big.NewInt(20_000_000_000),
				GasLimit: 21000,
				To:       to,
				Value:    oneEther,
			},
			sigHash: "daf5a779ae972f972197303d7b574746c7ef83eadac0f2791ad23db92e4c8e53",
			raw:     "f86c098504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a0499aa1110848b179aa0f228e20faa3ba68b350e1feeab49638c6b8ce40ea56aea0053ec9b43dcea26f8d10b43a44bdfafae5b1b26462367921079005d4d274e06d",
			hash:    "e4e0d6b0c5b43efcf6651888cf149d88384f0762ff684a58a06f8997b6e1f979",
			sig:     "499aa1110848b179aa0f228e20faa3ba68b350e1feeab49638c6b8ce40ea56ae053ec9b43dcea26f8d10b43a44bdfafae5b1b26462367921079005d4d274e06d",
			recID:   0,
		},
		{
			name: "contract creation",
			tx: LegacyTx{
				ChainID:  big.NewInt(1),
				Nonce:    0,
				GasPrice: big.NewInt(1_000_000_000),
				GasLimit: 100000,
				Value:    big.NewInt(0),
				Data:     mustDecodeHex(t, "6080604052"),
			},
			sigHash: "557f33e863ebd94b5e59649fdd39bc5659ccf082f4f031d61445d0bdb60a6d57",
			raw:     "f85580843b9aca00830186a0808085608060405225a09a1f8df4775546749d19062cd1c0e55d4fa788f2a27d79ab7e45163005362b7aa031b0293821cba07de3ea55ffcdbc9a0756609f7c3663b128bf86d1d2c52c58af",
			hash:    "171f6250685d00782e6def94aaa2bf09332fbae7391c96a06d0dd89633084aca",
			sig:     "9a1f8df4775546749d19062cd1c0e55d4fa788f2a27d79ab7e45163005362b7a31b0293821cba07de3ea55ffcdbc9a0756609f7c3663b128bf86d1d2c52c58af",
			recID:   0,
		},
		{
			name: "erc20 call on sepolia",
			tx: LegacyTx{
				ChainID:  big.NewInt(11155111),
				Nonce:    300,
				GasPrice: big.NewInt(7),
				GasLimit: 60000,
				To:       to,
				Value:    big.NewInt(0),
				Data:     mustDecodeHex(t, "a9059cbb00000000000000000000000035353535353535353535353535353535353535350000000000000000000000000000000000000000000000000000000000000001"),
			},
			sigHash: "b36666c54a2d3914c950fcae09258a452b004363e74b8158ca00cb863086b517",
			raw:     "f8aa82012c0782ea6094353535353535353535353535353535353535353580b844a9059cbb000000000000000000000000353535353535353535353535353535353535353500000000000000000000000000000000000000000000000000000000000000018401546d71a0f736fb6ebcf91458e07bafae401adb0257620f50a187e3846f610cc1560f468ea0214547ed38701be171d965c8cbca2e12b849c30ec5400fa46da7050425bb0218",
			hash:    "d9f6d0632f95eb4cbb46cd5db5072a4180be6632f160eeba42d2dd6c2098dbd9",
			sig:     "f736fb6ebcf91458e07bafae401adb0257620f50a187e3846f610cc1560f468e214547ed38701be171d965c8cbca2e12b849c30ec5400fa46da7050425bb0218",
			recID:   0,
		},
		{
			name: "unprotected pre-eip155",
			tx: LegacyTx{
				ChainID:  big.NewInt(0),
				Nonce:    1,
				GasPrice: big.NewInt(2_000_000_000),
				GasLimit: 21000,
				To:       to,
				Value:    big.NewInt(12345),
			},
			sigHash: "eac0eea517e3fa71d2dcd0bcfe320844223ae8b9d1244af88ce3bc5acba69992",
			raw:     "f865018477359400825208943535353535353535353535353535353535353535823039801ba08b658f1eda41f671154751160cfd4dc026e427c48a49d1393a8e700cac8f6b2ca03d4e25002bde45fa449cbf4a3f649a465a3376b956e48757860e8cc5ae5df45f",
			hash:    "707fb1bef82c7d14c1c1618fa37500114b429515ae08ba1efcec082af85a1fdd",
			sig:     "8b658f1eda41f671154751160cfd4dc026e427c48a49d1393a8e700cac8f6b2c3d4e25002bde45fa449cbf4a3f649a465a3376b956e48757860e8cc5ae5df45f",
			recID:   0,
		},
		{
			name: "large chain id and r with leading zero byte",
			tx: LegacyTx{
				ChainID:  big.NewInt(1380012617),
				Nonce:    2,
				GasPrice: big.NewInt(100),
				GasLimit: 21000,
				To:       to,
				Value:    big.NewInt(1),
			},
			sigHash: "97f9bd8a217be152ed7dad2b837108ecfe4fe01c2ef2066b290858a0afa58e9d",
			raw:     "f8620264825208943535353535353535353535353535353535353535018084a482a4b59f5a62fe8971d6215cc5d577a1655804104ef069588e3d0081aec62b6583a9f1a027ffc14adf23c8f177070867662fef6845d08b3049566799207e29f9c927d281",
			hash:    "3b584a777a481052695248abb48dc4231fb39dfb5d3617392e582d6a1906f743",
			sig:     "005a62fe8971d6215cc5d577a1655804104ef069588e3d0081aec62b6583a9f127ffc14adf23c8f177070867662fef6845d08b3049566799207e29f9c927d281",
			recID:   0,
		},
		{
			name: "mainnet transfer with recovery id 1",
			tx: LegacyTx{
				ChainID:  big.NewInt(1),
				Nonce:    1,
				GasPrice: big.NewInt(3_000_000_000),
				GasLimit: 21000,
				To:       to,
				Value:    big.NewInt(777),
			},
			sigHash: "05a52fa256b408bc46990fcecf810fea38a72df03f24f952c2fbb8670ae3085a",
			raw:     "f8650184b2d05e008252089435353535353535353535353535353535353535358203098026a00128e512386b150081bbc17a3310e2e1768d8e05dba30422ab98791b002707fea043ec668b396bca3b75ccb942f5773d1e5f9ae04ed69109665e1c3c377898fc64",
			hash:    "4b89814df66cf84985b44b0b207326bc1cee2f5d0fe08805d3a2406c2b6fb458",
			sig:     "0128e512386b150081bbc17a3310e2e1768d8e05dba30422ab98791b002707fe43ec668b396bca3b75ccb942f5773d1e5f9ae04ed69109665e1c3c377898fc64",
			recID:   1,
		},
		{
			name: "unprotected with recovery id 1",
			tx: LegacyTx{
				ChainID:  big.NewInt(0),
				Nonce:    0,
				GasPrice: big.NewInt(3_000_000_000),
				GasLimit: 21000,
				To:       to,
				Value:    big.NewInt(777),
			},
			sigHash: "d89a70fcf2ae366b01a6c983915169ed209ab2f6b41eb6201a8f8eecc377994e",
			raw:     "f8658084b2d05e00825208943535353535353535353535353535353535353535820309801ca0e8c8ef11b7a5ffefcd975d668a3c8036cee74d41217e028579e6f44e2b4d62a3a0688264114f2e02967a3065e772894b8118b98499af252bc09822d29d70fbfea4",
			hash:    "ece03833ebb141fb71b36c00688f79552953ff962b9d92ea3722a6beafb6aa4b",
			sig:     "e8c8ef11b7a5ffefcd975d668a3c8036cee74d41217e028579e6f44e2b4d62a3688264114f2e02967a3065e772894b8118b98499af252bc09822d29d70fbfea4",
			recID:   1,
		},
	}
}

func signatureWithRecoveryID(t *testing.T, rs string, recID byte) *types.Signature {
	t.Helper()
	sig, err := types.NewSignatureFromBytes(append(mustDecodeHex(t, rs), recID))
	require.NoError(t, err)
	return sig
}

func TestLegacyTxType(t *testing.T) {
	require.Equal(t, LegacyTxType, (&LegacyTx{}).Type())
}

func TestLegacyTxSigningHashMatchesGeth(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			got, err := tx.SigningHash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.sigHash), got.Bytes())
		})
	}
}

func TestLegacyTxSigningHashIgnoresSignature(t *testing.T) {
	v := legacyVectors(t)[0]
	unsigned := v.tx
	signed := v.tx
	signed.Signature = signatureWithRecoveryID(t, v.sig, v.recID)

	a, err := unsigned.SigningHash()
	require.NoError(t, err)
	b, err := signed.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestLegacyTxEncodeRLPMatchesGeth(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), got)
		})
	}
}

func TestLegacyTxHashMatchesGeth(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
			got, err := tx.Hash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.hash), got.Bytes())
		})
	}
}

func TestLegacyTxDecodeRLPMatchesGeth(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			var got LegacyTx
			require.NoError(t, got.DecodeRLP(mustDecodeHex(t, v.raw)))

			require.Equal(t, 0, v.tx.ChainID.Cmp(got.ChainID))
			require.Equal(t, v.tx.Nonce, got.Nonce)
			require.Equal(t, 0, v.tx.GasPrice.Cmp(got.GasPrice))
			require.Equal(t, v.tx.GasLimit, got.GasLimit)
			require.True(t, v.tx.To.Equal(got.To))
			require.Equal(t, 0, v.tx.Value.Cmp(got.Value))
			require.Equal(t, len(v.tx.Data), len(got.Data))
			require.Equal(t, string(v.tx.Data), string(got.Data))
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(got.Signature))
		})
	}
}

func TestLegacyTxDecodeThenEncodeIsIdentity(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			raw := mustDecodeHex(t, v.raw)
			var tx LegacyTx
			require.NoError(t, tx.DecodeRLP(raw))
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, raw, got)
		})
	}
}

func TestLegacyTxSignRecoversSender(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	require.NoError(t, err)
	want := PubkeyToAddress(key.PublicKey())

	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			require.NotNil(t, tx.Signature)

			digest, err := tx.SigningHash()
			require.NoError(t, err)
			ok, err := VerifyAddress(digest, tx.Signature, want)
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestLegacyTxSignThenRoundTrip(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	require.NoError(t, err)

	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)

			var decoded LegacyTx
			require.NoError(t, decoded.DecodeRLP(raw))
			require.True(t, tx.Signature.Equal(decoded.Signature))
			require.Equal(t, 0, tx.ChainID.Cmp(decoded.ChainID))
		})
	}
}

func TestLegacyTxEncodeRLPUnsigned(t *testing.T) {
	tx := legacyVectors(t)[0].tx
	_, err := tx.EncodeRLP()
	require.ErrorIs(t, err, ErrUnsigned)
	_, err = tx.Hash()
	require.ErrorIs(t, err, ErrUnsigned)
}

func TestLegacyTxNilFieldsEncodeAsZero(t *testing.T) {
	var tx LegacyTx
	nilHash, err := tx.SigningHash()
	require.NoError(t, err)

	zero := LegacyTx{ChainID: big.NewInt(0), GasPrice: big.NewInt(0), Value: big.NewInt(0)}
	zeroHash, err := zero.SigningHash()
	require.NoError(t, err)
	require.True(t, nilHash.Equal(zeroHash))
}

func TestLegacyTxNegativeValueRejected(t *testing.T) {
	tx := LegacyTx{ChainID: big.NewInt(1), Value: big.NewInt(-1)}
	_, err := tx.SigningHash()
	require.ErrorIs(t, err, ErrInvalidInteger)
}

func TestLegacyTxEncodeRLPRejectsHighRecoveryID(t *testing.T) {
	v := legacyVectors(t)[0]
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, 2)
	_, err := tx.EncodeRLP()
	require.Error(t, err)
}

func TestLegacyTxDecodeRLPInvalid(t *testing.T) {
	valid := mustDecodeHex(t, legacyVectors(t)[0].raw)

	tests := []struct {
		name string
		raw  []byte
	}{
		{"empty", nil},
		{"trailing byte", append(append([]byte{}, valid...), 0x00)},
		{"truncated", valid[:len(valid)-1]},
		{"typed envelope", append([]byte{0x02}, valid...)},
		{"not a list", []byte{0x83, 0x01, 0x02, 0x03}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tx LegacyTx
			require.Error(t, tx.DecodeRLP(tt.raw))
		})
	}
}

func TestLegacyTxDecodeRLPInvalidV(t *testing.T) {
	v := legacyVectors(t)[0]
	sigRS := mustDecodeHex(t, v.sig)

	for _, vv := range []int64{0, 1, 26, 29, 34} {
		t.Run(big.NewInt(vv).String(), func(t *testing.T) {
			enc, err := rlpEncodeLegacySigned(v.tx, big.NewInt(vv), sigRS)
			require.NoError(t, err)
			var tx LegacyTx
			require.Error(t, tx.DecodeRLP(enc))
		})
	}
}

func rlpEncodeLegacySigned(tx LegacyTx, v *big.Int, rs []byte) ([]byte, error) {
	return rlpEncodeForTest(&rlp.LegacySignedRLP{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		To:       rlp.FromAddress(tx.To),
		Value:    tx.Value,
		Data:     tx.Data,
		V:        v,
		R:        new(big.Int).SetBytes(rs[:32]),
		S:        new(big.Int).SetBytes(rs[32:]),
	})
}

func TestLegacyTxDecodeRLPResetsReceiver(t *testing.T) {
	tx := LegacyTx{Nonce: 99, Data: []byte{1}, To: legacyTestAddress(t)}
	require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, legacyVectors(t)[1].raw)))
	require.Nil(t, tx.To)
	require.Equal(t, uint64(0), tx.Nonce)
}

func rlpEncodeForTest(v any) ([]byte, error) {
	return rlp.Encode(v)
}

func mustDecodeHexF(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

type failingSigner struct {
	err error
}

func (f failingSigner) Sign(*types.Hash) (*types.Signature, error) {
	return nil, f.err
}

func (f failingSigner) PublicKey() *types.PublicKey {
	return nil
}

func legacyTestPrivateKey(t *testing.T) *types.PrivateKey {
	t.Helper()
	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	require.NoError(t, err)
	return key
}

func TestLegacyTxSignPropagatesSignerError(t *testing.T) {
	boom := errors.New("boom")
	tx := legacyVectors(t)[0].tx

	err := tx.Sign(failingSigner{err: boom})
	require.ErrorIs(t, err, boom)
	require.Nil(t, tx.Signature)
}

func TestLegacyTxSignRejectsInvalidPayload(t *testing.T) {
	tx := LegacyTx{ChainID: big.NewInt(1), Value: big.NewInt(-1)}
	require.ErrorIs(t, tx.Sign(legacyTestPrivateKey(t)), ErrInvalidInteger)
	require.Nil(t, tx.Signature)
}

func TestLegacyTxSignMatchesGethSignature(t *testing.T) {
	key := legacyTestPrivateKey(t)
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(tx.Signature))

			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), raw)
		})
	}
}

func TestLegacyTxSignOverwritesSignature(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := legacyVectors(t)[0].tx

	require.NoError(t, tx.Sign(key))
	first := tx.Signature

	tx.Nonce++
	require.NoError(t, tx.Sign(key))
	require.False(t, first.Equal(tx.Signature))
}

func TestLegacyTxSigningHashDependsOnChainID(t *testing.T) {
	base := legacyVectors(t)[0].tx

	seen := map[string]int64{}
	for _, chainID := range []int64{0, 1, 5, 11155111} {
		tx := base
		tx.ChainID = big.NewInt(chainID)
		h, err := tx.SigningHash()
		require.NoError(t, err)
		prev, dup := seen[h.String()]
		require.False(t, dup, "chain %d collides with chain %d", chainID, prev)
		seen[h.String()] = chainID
	}
}

func TestLegacyTxSigningHashDependsOnEveryField(t *testing.T) {
	base := legacyVectors(t)[0].tx
	baseHash, err := base.SigningHash()
	require.NoError(t, err)

	otherTo, err := types.NewAddressFromHex("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	mutations := map[string]func(tx *LegacyTx){
		"nonce":    func(tx *LegacyTx) { tx.Nonce++ },
		"gasPrice": func(tx *LegacyTx) { tx.GasPrice = big.NewInt(1) },
		"gasLimit": func(tx *LegacyTx) { tx.GasLimit++ },
		"to":       func(tx *LegacyTx) { tx.To = otherTo },
		"value":    func(tx *LegacyTx) { tx.Value = big.NewInt(1) },
		"data":     func(tx *LegacyTx) { tx.Data = []byte{0x01} },
		"chainID":  func(tx *LegacyTx) { tx.ChainID = big.NewInt(2) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			tx := base
			mutate(&tx)
			h, err := tx.SigningHash()
			require.NoError(t, err)
			require.False(t, baseHash.Equal(h))
		})
	}
}

func TestLegacyTxNilToDiffersFromZeroAddress(t *testing.T) {
	key := legacyTestPrivateKey(t)
	zero, err := types.NewAddressFromHex("0x0000000000000000000000000000000000000000")
	require.NoError(t, err)

	creation := LegacyTx{ChainID: big.NewInt(1), GasLimit: 21000}
	toZero := creation
	toZero.To = zero

	a, err := creation.SigningHash()
	require.NoError(t, err)
	b, err := toZero.SigningHash()
	require.NoError(t, err)
	require.False(t, a.Equal(b))

	for _, tx := range []*LegacyTx{&creation, &toZero} {
		require.NoError(t, tx.Sign(key))
		raw, err := tx.EncodeRLP()
		require.NoError(t, err)

		var decoded LegacyTx
		require.NoError(t, decoded.DecodeRLP(raw))
		if tx.To == nil {
			require.Nil(t, decoded.To)
		} else {
			require.True(t, tx.To.Equal(decoded.To))
		}
	}
}

func TestLegacyTxNilDataEqualsEmptyData(t *testing.T) {
	key := legacyTestPrivateKey(t)
	nilData := legacyVectors(t)[0].tx
	emptyData := nilData
	emptyData.Data = []byte{}

	require.NoError(t, nilData.Sign(key))
	require.NoError(t, emptyData.Sign(key))

	a, err := nilData.EncodeRLP()
	require.NoError(t, err)
	b, err := emptyData.EncodeRLP()
	require.NoError(t, err)
	require.Equal(t, a, b)
}

func TestLegacyTxExtremeValuesRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)
	maxU256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	hugeChainID := new(big.Int).Lsh(big.NewInt(1), 70)

	tx := LegacyTx{
		ChainID:  hugeChainID,
		Nonce:    math.MaxUint64,
		GasPrice: maxU256,
		GasLimit: math.MaxUint64,
		To:       legacyTestAddress(t),
		Value:    maxU256,
		Data:     make([]byte, 4096),
	}
	require.NoError(t, tx.Sign(key))
	raw, err := tx.EncodeRLP()
	require.NoError(t, err)

	var got LegacyTx
	require.NoError(t, got.DecodeRLP(raw))
	require.Equal(t, 0, tx.ChainID.Cmp(got.ChainID))
	require.Equal(t, tx.Nonce, got.Nonce)
	require.Equal(t, 0, tx.GasPrice.Cmp(got.GasPrice))
	require.Equal(t, tx.GasLimit, got.GasLimit)
	require.Equal(t, 0, tx.Value.Cmp(got.Value))
	require.Equal(t, len(tx.Data), len(got.Data))
	require.True(t, tx.Signature.Equal(got.Signature))

	digest, err := got.SigningHash()
	require.NoError(t, err)
	ok, err := VerifyAddress(digest, got.Signature, PubkeyToAddress(key.PublicKey()))
	require.NoError(t, err)
	require.True(t, ok)
}

func TestLegacyTxHashIsKeccakOfEncoding(t *testing.T) {
	for _, v := range legacyVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			h, err := tx.Hash()
			require.NoError(t, err)
			require.True(t, Keccak256(raw).Equal(h))
		})
	}
}

func TestLegacyTxDecodeRLPRejectsProtectedVWithZeroChainID(t *testing.T) {
	v := legacyVectors(t)[0]
	rs := mustDecodeHex(t, v.sig)

	for _, vv := range []int64{35, 36} {
		t.Run(big.NewInt(vv).String(), func(t *testing.T) {
			enc, err := rlpEncodeLegacySigned(v.tx, big.NewInt(vv), rs)
			require.NoError(t, err)
			var tx LegacyTx
			require.Error(t, tx.DecodeRLP(enc))
		})
	}
}

func TestLegacyTxDecodeRLPAcceptsBothLegacyVValues(t *testing.T) {
	v := legacyVectors(t)[3]
	rs := mustDecodeHex(t, v.sig)

	for _, vv := range []int64{27, 28} {
		t.Run(big.NewInt(vv).String(), func(t *testing.T) {
			enc, err := rlpEncodeLegacySigned(v.tx, big.NewInt(vv), rs)
			require.NoError(t, err)
			var tx LegacyTx
			require.NoError(t, tx.DecodeRLP(enc))
			require.Equal(t, 0, tx.ChainID.Sign())
			require.Equal(t, vv-27, tx.Signature.V().Int64())
		})
	}
}

func TestLegacyTxDecodeRLPInvalidSignatureValues(t *testing.T) {
	v := legacyVectors(t)[0]
	rs := mustDecodeHex(t, v.sig)
	r := new(big.Int).SetBytes(rs[:32])
	s := new(big.Int).SetBytes(rs[32:])
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)

	tests := []struct {
		name string
		r, s *big.Int
	}{
		{"zero r", big.NewInt(0), s},
		{"zero s", r, big.NewInt(0)},
		{"r over 256 bits", tooBig, s},
		{"s over 256 bits", r, tooBig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := rlp.Encode(&rlp.LegacySignedRLP{
				Nonce:    v.tx.Nonce,
				GasPrice: v.tx.GasPrice,
				GasLimit: v.tx.GasLimit,
				To:       rlp.FromAddress(v.tx.To),
				Value:    v.tx.Value,
				Data:     v.tx.Data,
				V:        big.NewInt(37),
				R:        tt.r,
				S:        tt.s,
			})
			require.NoError(t, err)
			var tx LegacyTx
			require.Error(t, tx.DecodeRLP(enc))
		})
	}
}

func TestLegacyTxDecodeRLPInvalidShape(t *testing.T) {
	v := legacyVectors(t)[0]
	one := big.NewInt(1)

	tests := []struct {
		name string
		in   any
	}{
		{"short to", []any{uint64(0), one, uint64(21000), make([]byte, 19), one, []byte{}, big.NewInt(37), one, one}},
		{"long to", []any{uint64(0), one, uint64(21000), make([]byte, 21), one, []byte{}, big.NewInt(37), one, one}},
		{"eight fields", []any{uint64(0), one, uint64(21000), make([]byte, 20), one, []byte{}, big.NewInt(37), one}},
		{"ten fields", []any{uint64(0), one, uint64(21000), make([]byte, 20), one, []byte{}, big.NewInt(37), one, one, one}},
		{"nested list as data", []any{uint64(0), one, uint64(21000), make([]byte, 20), one, []any{}, big.NewInt(37), one, one}},
		{"empty list", []any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := rlp.Encode(tt.in)
			require.NoError(t, err)
			var tx LegacyTx
			require.Error(t, tx.DecodeRLP(enc))
		})
	}

	valid := mustDecodeHex(t, v.raw)
	var tx LegacyTx
	require.NoError(t, tx.DecodeRLP(valid))
}

func TestLegacyTxDecodeRLPFailureLeavesReceiverUntouched(t *testing.T) {
	tx := LegacyTx{Nonce: 42, ChainID: big.NewInt(7)}
	require.Error(t, tx.DecodeRLP([]byte{0xc0}))
	require.Equal(t, uint64(42), tx.Nonce)
	require.Equal(t, int64(7), tx.ChainID.Int64())
}

func TestLegacyTxEncodeRLPRejectsNegativeValue(t *testing.T) {
	v := legacyVectors(t)[0]
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
	tx.Value = big.NewInt(-1)

	_, err := tx.EncodeRLP()
	require.ErrorIs(t, err, ErrInvalidInteger)
	_, err = tx.Hash()
	require.ErrorIs(t, err, ErrInvalidInteger)
}

func TestLegacyTxIntegerFieldBounds(t *testing.T) {
	key := legacyTestPrivateKey(t)
	maxU256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	fields := map[string]func(tx *LegacyTx, n *big.Int){
		"chainId":  func(tx *LegacyTx, n *big.Int) { tx.ChainID = n },
		"gasPrice": func(tx *LegacyTx, n *big.Int) { tx.GasPrice = n },
		"value":    func(tx *LegacyTx, n *big.Int) { tx.Value = n },
	}
	cases := []struct {
		name string
		n    *big.Int
		want error
	}{
		{"negative", big.NewInt(-1), ErrInvalidInteger},
		{"large negative", new(big.Int).Neg(overflow), ErrInvalidInteger},
		{"overflow", overflow, ErrInvalidInteger},
		{"far overflow", new(big.Int).Lsh(big.NewInt(1), 1024), ErrInvalidInteger},
		{"max", maxU256, nil},
		{"zero", big.NewInt(0), nil},
	}

	for field, set := range fields {
		for _, tt := range cases {
			t.Run(field+"/"+tt.name, func(t *testing.T) {
				tx := legacyVectors(t)[0].tx
				set(&tx, tt.n)

				_, err := tx.SigningHash()
				if tt.want != nil {
					require.ErrorIs(t, err, tt.want)
					require.ErrorContains(t, err, field)
					require.ErrorIs(t, tx.Sign(key), tt.want)

					tx.Signature = signatureWithRecoveryID(t, legacyVectors(t)[0].sig, 0)
					_, err = tx.EncodeRLP()
					require.ErrorIs(t, err, tt.want)
					_, err = tx.Hash()
					require.ErrorIs(t, err, tt.want)
					return
				}
				require.NoError(t, err)
				require.NoError(t, tx.Sign(key))
				_, err = tx.EncodeRLP()
				require.NoError(t, err)
			})
		}
	}
}

func TestLegacyTxDecodeRLPRejectsOversizedIntegers(t *testing.T) {
	v := legacyVectors(t)[0]
	rs := mustDecodeHex(t, v.sig)
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	encode := func(tx LegacyTx, vv *big.Int) []byte {
		enc, err := rlpEncodeLegacySigned(tx, vv, rs)
		require.NoError(t, err)
		return enc
	}

	t.Run("gasPrice", func(t *testing.T) {
		tx := v.tx
		tx.GasPrice = overflow
		var got LegacyTx
		require.ErrorIs(t, got.DecodeRLP(encode(tx, big.NewInt(37))), ErrInvalidInteger)
	})
	t.Run("value", func(t *testing.T) {
		tx := v.tx
		tx.Value = overflow
		var got LegacyTx
		require.ErrorIs(t, got.DecodeRLP(encode(tx, big.NewInt(37))), ErrInvalidInteger)
	})
	t.Run("chainId from v", func(t *testing.T) {
		huge := new(big.Int).Lsh(big.NewInt(1), 258)
		var got LegacyTx
		require.ErrorIs(t, got.DecodeRLP(encode(v.tx, huge)), ErrInvalidInteger)
	})
	t.Run("failure leaves receiver untouched", func(t *testing.T) {
		tx := v.tx
		tx.Value = overflow
		got := LegacyTx{Nonce: 42}
		require.Error(t, got.DecodeRLP(encode(tx, big.NewInt(37))))
		require.Equal(t, uint64(42), got.Nonce)
	})
}

type dynamicFeeVector struct {
	name    string
	tx      DynamicFeeTx
	sigHash string
	raw     string
	hash    string
	sig     string
	recID   byte
}

func mustBigDecimal(t *testing.T, s string) *big.Int {
	t.Helper()
	n, ok := new(big.Int).SetString(s, 10)
	require.True(t, ok)
	return n
}

func testAddress(t *testing.T, s string) *types.Address {
	t.Helper()
	a, err := types.NewAddressFromHex(s)
	require.NoError(t, err)
	return a
}

func testHash(t *testing.T, s string) *types.Hash {
	t.Helper()
	h, err := types.NewHashFromHex(s)
	require.NoError(t, err)
	return h
}

func dynamicFeeVectors(t *testing.T) []dynamicFeeVector {
	return []dynamicFeeVector{
		{
			name: "simple transfer",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      9,
				GasTipCap:  mustBigDecimal(t, "1500000000"),
				GasFeeCap:  mustBigDecimal(t, "30000000000"),
				GasLimit:   21000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "1000000000000000000"),
				Data:       nil,
				AccessList: nil,
			},
			sigHash: "b6d5fd62e3a62df7b38773c31edc4e889e055257a8fcfb64d40d33adb6ad5273",
			raw:     "02f87301098459682f008506fc23ac00825208943535353535353535353535353535353535353535880de0b6b3a764000080c001a0a2a0cb81cd4d8e240f38279012ee02b696002ebd27203a64d4acd6cc60d2f139a029de4d469feeadf9d8bd087fbd31c50c015928b4ac387409758415b8e301fe57",
			hash:    "8abf3210e020899ef495828939d4eca9ee9c8119c35eac183abbc824c4a627e5",
			sig:     "a2a0cb81cd4d8e240f38279012ee02b696002ebd27203a64d4acd6cc60d2f13929de4d469feeadf9d8bd087fbd31c50c015928b4ac387409758415b8e301fe57",
			recID:   1,
		},
		{
			name: "contract creation",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      0,
				GasTipCap:  mustBigDecimal(t, "1000000000"),
				GasFeeCap:  mustBigDecimal(t, "2000000000"),
				GasLimit:   300000,
				To:         nil,
				Value:      mustBigDecimal(t, "0"),
				Data:       mustDecodeHex(t, "6080604052"),
				AccessList: nil,
			},
			sigHash: "ef12c6909a4ea58ea70c5d2ca274fef4173281a4725eb9b40fa40289b0fc51c7",
			raw:     "02f85c0180843b9aca008477359400830493e08080856080604052c080a0fb01d4a851187ba170117585b16e34daf1669fbf9cfc7e49a7b491ee65d948b0a0085bad6f450392057b227c6af7206f570a6df58f554d78906bfb3e8b12d94b17",
			hash:    "d383a8449089a32a453a6774672e26c635171412b99baa2290434d6327a9cd40",
			sig:     "fb01d4a851187ba170117585b16e34daf1669fbf9cfc7e49a7b491ee65d948b0085bad6f450392057b227c6af7206f570a6df58f554d78906bfb3e8b12d94b17",
			recID:   0,
		},
		{
			name: "erc20 call with access list",
			tx: DynamicFeeTx{
				ChainID:   mustBigDecimal(t, "11155111"),
				Nonce:     300,
				GasTipCap: mustBigDecimal(t, "2"),
				GasFeeCap: mustBigDecimal(t, "7"),
				GasLimit:  60000,
				To:        testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:     mustBigDecimal(t, "0"),
				Data:      mustDecodeHex(t, "a9059cbb00000000000000000000000035353535353535353535353535353535353535350000000000000000000000000000000000000000000000000000000000000001"),
				AccessList: AccessList{
					{Address: testAddress(t, "0x3535353535353535353535353535353535353535"), StorageKeys: []*types.Hash{testHash(t, "0x0000000000000000000000000000000000000000000000000000000000000000"), testHash(t, "0x0000000000000000000000000000000000000000000000000000000000000001")}},
					{Address: testAddress(t, "0x0000000000000000000000000000000000000001"), StorageKeys: []*types.Hash{}},
				},
			},
			sigHash: "65058d1adb3c4661a716d8b03ff85267df9ea98206ff96c6e864c8d32cd62e01",
			raw:     "02f9011f83aa36a782012c020782ea6094353535353535353535353535353535353535353580b844a9059cbb00000000000000000000000035353535353535353535353535353535353535350000000000000000000000000000000000000000000000000000000000000001f872f859943535353535353535353535353535353535353535f842a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000001d6940000000000000000000000000000000000000001c001a077d78238660067a518b7725cf7d64aa65a5cfd1ac0b13f175eb5c0e9b7cceb4ea002ffc76be8ff4bca0096f173a39df34a7afea6529b2d9393d74e4b213c78e70c",
			hash:    "15467eeeb65df915e0e22838db5914889bfdd29bfe77589b0059327cf98adfde",
			sig:     "77d78238660067a518b7725cf7d64aa65a5cfd1ac0b13f175eb5c0e9b7cceb4e02ffc76be8ff4bca0096f173a39df34a7afea6529b2d9393d74e4b213c78e70c",
			recID:   1,
		},
		{
			name: "access list entry without keys",
			tx: DynamicFeeTx{
				ChainID:   mustBigDecimal(t, "1"),
				Nonce:     4,
				GasTipCap: mustBigDecimal(t, "1"),
				GasFeeCap: mustBigDecimal(t, "1"),
				GasLimit:  25000,
				To:        testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:     mustBigDecimal(t, "5"),
				Data:      nil,
				AccessList: AccessList{
					{Address: testAddress(t, "0x0000000000000000000000000000000000000002"), StorageKeys: []*types.Hash{}},
				},
			},
			sigHash: "44807ee475e4c4c768f87e0ce69515103de18bf749e26849d555a4db08485c55",
			raw:     "02f879010401018261a89435353535353535353535353535353535353535350580d7d6940000000000000000000000000000000000000002c080a0789f8ffe24f585b430af85ca7756d91cf70e7fa3ee4609a1ff744071fc314fa3a0458d637340a352ca3b0eafe3755aca6a138172ccd943559e21ccdf49ba330dcd",
			hash:    "30d6c77c7cc31e955724e634c686962b3651920f8bce63045250df3c4b54e4f4",
			sig:     "789f8ffe24f585b430af85ca7756d91cf70e7fa3ee4609a1ff744071fc314fa3458d637340a352ca3b0eafe3755aca6a138172ccd943559e21ccdf49ba330dcd",
			recID:   0,
		},
		{
			name: "zero fees and value",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      0,
				GasTipCap:  mustBigDecimal(t, "0"),
				GasFeeCap:  mustBigDecimal(t, "0"),
				GasLimit:   21000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "0"),
				Data:       nil,
				AccessList: nil,
			},
			sigHash: "fe51751c4b08e01c021d5c24bea0aebbebe360e64bdaa2adb7bc5440c241b1ed",
			raw:     "02f862018080808252089435353535353535353535353535353535353535358080c001a00fe3d44e990382326b128813890cd2afc4c687670af4dc63a732565157b6af1ea0491d73e8afb107b720832670b11efe5a057025d3e4d936f40196c18f02d140ac",
			hash:    "9735509fd5549140c62c201c1498f7f58ea9ed679534a655fcd9173f15fe3153",
			sig:     "0fe3d44e990382326b128813890cd2afc4c687670af4dc63a732565157b6af1e491d73e8afb107b720832670b11efe5a057025d3e4d936f40196c18f02d140ac",
			recID:   1,
		},
		{
			name: "large chain id",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1380012617"),
				Nonce:      2,
				GasTipCap:  mustBigDecimal(t, "100"),
				GasFeeCap:  mustBigDecimal(t, "200"),
				GasLimit:   21000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "1"),
				Data:       nil,
				AccessList: nil,
			},
			sigHash: "c2b069e75db8c9470c3d8e3ed789405a1566717a7714be6bb4583d9a62d4685e",
			raw:     "02f8678452415249026481c88252089435353535353535353535353535353535353535350180c080a01e3fa8beaa7d4b8d723034225755b5469745bd45f40d2013b3259e6cecc9563ba038d1ab678ee9e6c21fa5e532348d602d532593ae17a299098036866d1462bdde",
			hash:    "c8ddb8dab0c3b03f977ff79ce23363a96ed52aba653fc4fdd87fb1291a77864b",
			sig:     "1e3fa8beaa7d4b8d723034225755b5469745bd45f40d2013b3259e6cecc9563b38d1ab678ee9e6c21fa5e532348d602d532593ae17a299098036866d1462bdde",
			recID:   0,
		},
		{
			name: "recovery id 1",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      2,
				GasTipCap:  mustBigDecimal(t, "3"),
				GasFeeCap:  mustBigDecimal(t, "9"),
				GasLimit:   21000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "777"),
				Data:       nil,
				AccessList: nil,
			},
			sigHash: "8c308f2e2dbeb28d7c17f00b146e7f223ef2ede63873aaeb1a7f21222a3dc784",
			raw:     "02f8640102030982520894353535353535353535353535353535353535353582030980c001a035364d5f553a9ee80bccb4a1242d6f078e1d2b778d41c023883ba7d6885a5d33a06f11f2e71982f17b22cf894f312c26466a4ae85c595e680bda1a56c420b3bc98",
			hash:    "a8c0b7e7b2f5c72aa1270d8aff8e49c876ff352e73a129e37da91d0079511306",
			sig:     "35364d5f553a9ee80bccb4a1242d6f078e1d2b778d41c023883ba7d6885a5d336f11f2e71982f17b22cf894f312c26466a4ae85c595e680bda1a56c420b3bc98",
			recID:   1,
		},
		{
			name: "max fee values",
			tx: DynamicFeeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      18446744073709551615,
				GasTipCap:  mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				GasFeeCap:  mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				GasLimit:   18446744073709551615,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				Data:       nil,
				AccessList: nil,
			},
			sigHash: "c3fdc82cab8f8023ed54f5b77b222bd4f892061a5087da6546deeaf6d941ed9d",
			raw:     "02f8d00188ffffffffffffffffa0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff88ffffffffffffffff943535353535353535353535353535353535353535a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c001a0e8b5b5f7bea01a1b9a82b1a990a3fc4ba36507fd072aa5d5fe83cd93684b28afa024c16419d9fc14cd1aadf758bd9468739fc8e2aa4daee328936e5d1516e0eba6",
			hash:    "b841ac56d82eea9957d53f16fc3d7480ef899565153005c78ce04ddb51d601ae",
			sig:     "e8b5b5f7bea01a1b9a82b1a990a3fc4ba36507fd072aa5d5fe83cd93684b28af24c16419d9fc14cd1aadf758bd9468739fc8e2aa4daee328936e5d1516e0eba6",
			recID:   1,
		},
	}
}

func withSignature(t *testing.T, v dynamicFeeVector) DynamicFeeTx {
	t.Helper()
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
	return tx
}

func encodeDynamicFeeSigned(t *testing.T, tx DynamicFeeTx, yParity, r, s *big.Int) []byte {
	t.Helper()
	accessList, err := accessListToRLP(tx.AccessList)
	require.NoError(t, err)
	enc, err := rlp.Encode(&rlp.DynamicFeeSignedRLP{
		ChainID:    tx.ChainID,
		Nonce:      tx.Nonce,
		GasTipCap:  tx.GasTipCap,
		GasFeeCap:  tx.GasFeeCap,
		GasLimit:   tx.GasLimit,
		To:         rlp.FromAddress(tx.To),
		Value:      tx.Value,
		Data:       tx.Data,
		AccessList: accessList,
		V:          yParity,
		R:          r,
		S:          s,
	})
	require.NoError(t, err)
	return append([]byte{0x02}, enc...)
}

func sigRS(t *testing.T, v dynamicFeeVector) (*big.Int, *big.Int) {
	t.Helper()
	b := mustDecodeHex(t, v.sig)
	return new(big.Int).SetBytes(b[:32]), new(big.Int).SetBytes(b[32:])
}

func requireSameDynamicFee(t *testing.T, want, got DynamicFeeTx) {
	t.Helper()
	require.Equal(t, 0, want.ChainID.Cmp(got.ChainID))
	require.Equal(t, want.Nonce, got.Nonce)
	require.Equal(t, 0, want.GasTipCap.Cmp(got.GasTipCap))
	require.Equal(t, 0, want.GasFeeCap.Cmp(got.GasFeeCap))
	require.Equal(t, want.GasLimit, got.GasLimit)
	require.True(t, want.To.Equal(got.To))
	require.Equal(t, 0, want.Value.Cmp(got.Value))
	require.Equal(t, string(want.Data), string(got.Data))
	require.Equal(t, len(want.AccessList), len(got.AccessList))
	for i := range want.AccessList {
		require.True(t, want.AccessList[i].Address.Equal(got.AccessList[i].Address))
		require.Equal(t, len(want.AccessList[i].StorageKeys), len(got.AccessList[i].StorageKeys))
		for j := range want.AccessList[i].StorageKeys {
			require.True(t, want.AccessList[i].StorageKeys[j].Equal(got.AccessList[i].StorageKeys[j]))
		}
	}
}

func TestDynamicFeeTxType(t *testing.T) {
	require.Equal(t, DynamicFeeTxType, (&DynamicFeeTx{}).Type())
}

func TestDynamicFeeTxSigningHashMatchesGeth(t *testing.T) {
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			got, err := tx.SigningHash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.sigHash), got.Bytes())
		})
	}
}

func TestDynamicFeeTxSigningHashIgnoresSignature(t *testing.T) {
	v := dynamicFeeVectors(t)[0]
	unsigned := v.tx
	signed := withSignature(t, v)

	a, err := unsigned.SigningHash()
	require.NoError(t, err)
	b, err := signed.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestDynamicFeeTxEncodeRLPMatchesGeth(t *testing.T) {
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := withSignature(t, v)
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), got)
		})
	}
}

func TestDynamicFeeTxHashMatchesGeth(t *testing.T) {
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := withSignature(t, v)
			got, err := tx.Hash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.hash), got.Bytes())
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.True(t, Keccak256(raw).Equal(got))
		})
	}
}

func TestDynamicFeeTxDecodeRLPMatchesGeth(t *testing.T) {
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			var got DynamicFeeTx
			require.NoError(t, got.DecodeRLP(mustDecodeHex(t, v.raw)))
			requireSameDynamicFee(t, v.tx, got)
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(got.Signature))
		})
	}
}

func TestDynamicFeeTxDecodeThenEncodeIsIdentity(t *testing.T) {
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			raw := mustDecodeHex(t, v.raw)
			var tx DynamicFeeTx
			require.NoError(t, tx.DecodeRLP(raw))
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, raw, got)
		})
	}
}

func TestDynamicFeeTxSignMatchesGeth(t *testing.T) {
	key := legacyTestPrivateKey(t)
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(tx.Signature))

			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), raw)
		})
	}
}

func TestDynamicFeeTxSignRecoversSender(t *testing.T) {
	key := legacyTestPrivateKey(t)
	want := PubkeyToAddress(key.PublicKey())

	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))

			digest, err := tx.SigningHash()
			require.NoError(t, err)
			ok, err := VerifyAddress(digest, tx.Signature, want)
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestDynamicFeeTxSignPropagatesSignerError(t *testing.T) {
	boom := errors.New("boom")
	tx := dynamicFeeVectors(t)[0].tx

	require.ErrorIs(t, tx.Sign(failingSigner{err: boom}), boom)
	require.Nil(t, tx.Signature)
}

func TestDynamicFeeTxSignOverwritesSignature(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := dynamicFeeVectors(t)[0].tx

	require.NoError(t, tx.Sign(key))
	first := tx.Signature

	tx.Nonce++
	require.NoError(t, tx.Sign(key))
	require.False(t, first.Equal(tx.Signature))
}

func TestDynamicFeeTxEncodeRLPUnsigned(t *testing.T) {
	tx := dynamicFeeVectors(t)[0].tx
	_, err := tx.EncodeRLP()
	require.ErrorIs(t, err, ErrUnsigned)
	_, err = tx.Hash()
	require.ErrorIs(t, err, ErrUnsigned)
}

func TestDynamicFeeTxEncodeRLPRejectsHighRecoveryID(t *testing.T) {
	v := dynamicFeeVectors(t)[0]
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, 2)
	_, err := tx.EncodeRLP()
	require.Error(t, err)
}

func TestDynamicFeeTxSigningHashDependsOnEveryField(t *testing.T) {
	base := dynamicFeeVectors(t)[2].tx
	baseHash, err := base.SigningHash()
	require.NoError(t, err)

	otherTo := testAddress(t, "0x0000000000000000000000000000000000000001")
	mutations := map[string]func(tx *DynamicFeeTx){
		"chainID":   func(tx *DynamicFeeTx) { tx.ChainID = big.NewInt(2) },
		"nonce":     func(tx *DynamicFeeTx) { tx.Nonce++ },
		"gasTipCap": func(tx *DynamicFeeTx) { tx.GasTipCap = big.NewInt(99) },
		"gasFeeCap": func(tx *DynamicFeeTx) { tx.GasFeeCap = big.NewInt(99) },
		"gasLimit":  func(tx *DynamicFeeTx) { tx.GasLimit++ },
		"to":        func(tx *DynamicFeeTx) { tx.To = otherTo },
		"value":     func(tx *DynamicFeeTx) { tx.Value = big.NewInt(1) },
		"data":      func(tx *DynamicFeeTx) { tx.Data = []byte{0x01} },
		"access list address": func(tx *DynamicFeeTx) {
			tx.AccessList = AccessList{{Address: otherTo, StorageKeys: tx.AccessList[0].StorageKeys}}
		},
		"access list key": func(tx *DynamicFeeTx) {
			tx.AccessList = AccessList{{Address: tx.AccessList[0].Address, StorageKeys: []*types.Hash{testHash(t, "0x00000000000000000000000000000000000000000000000000000000000000ff")}}}
		},
		"access list removed": func(tx *DynamicFeeTx) { tx.AccessList = nil },
		"access list key order": func(tx *DynamicFeeTx) {
			keys := tx.AccessList[0].StorageKeys
			tx.AccessList = AccessList{{Address: tx.AccessList[0].Address, StorageKeys: []*types.Hash{keys[1], keys[0]}}, tx.AccessList[1]}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			tx := base
			mutate(&tx)
			h, err := tx.SigningHash()
			require.NoError(t, err)
			require.False(t, baseHash.Equal(h))
		})
	}
}

func TestDynamicFeeTxSigningHashDiffersFromLegacy(t *testing.T) {
	typed := DynamicFeeTx{ChainID: big.NewInt(1), GasTipCap: big.NewInt(1), GasFeeCap: big.NewInt(1)}
	legacy := LegacyTx{ChainID: big.NewInt(1), GasPrice: big.NewInt(1)}

	a, err := typed.SigningHash()
	require.NoError(t, err)
	b, err := legacy.SigningHash()
	require.NoError(t, err)
	require.False(t, a.Equal(b))
}

func TestDynamicFeeTxNilToDiffersFromZeroAddress(t *testing.T) {
	key := legacyTestPrivateKey(t)
	zero := testAddress(t, "0x0000000000000000000000000000000000000000")

	creation := DynamicFeeTx{ChainID: big.NewInt(1), GasLimit: 21000}
	toZero := creation
	toZero.To = zero

	a, err := creation.SigningHash()
	require.NoError(t, err)
	b, err := toZero.SigningHash()
	require.NoError(t, err)
	require.False(t, a.Equal(b))

	for _, tx := range []*DynamicFeeTx{&creation, &toZero} {
		require.NoError(t, tx.Sign(key))
		raw, err := tx.EncodeRLP()
		require.NoError(t, err)

		var decoded DynamicFeeTx
		require.NoError(t, decoded.DecodeRLP(raw))
		if tx.To == nil {
			require.Nil(t, decoded.To)
		} else {
			require.True(t, tx.To.Equal(decoded.To))
		}
	}
}

func TestDynamicFeeTxNilEqualsEmptyDataAndAccessList(t *testing.T) {
	key := legacyTestPrivateKey(t)
	nilFields := dynamicFeeVectors(t)[0].tx
	emptyFields := nilFields
	emptyFields.Data = []byte{}
	emptyFields.AccessList = AccessList{}

	require.NoError(t, nilFields.Sign(key))
	require.NoError(t, emptyFields.Sign(key))

	a, err := nilFields.EncodeRLP()
	require.NoError(t, err)
	b, err := emptyFields.EncodeRLP()
	require.NoError(t, err)
	require.Equal(t, a, b)
}

func TestDynamicFeeTxNilIntegersEncodeAsZero(t *testing.T) {
	var tx DynamicFeeTx
	nilHash, err := tx.SigningHash()
	require.NoError(t, err)

	zero := DynamicFeeTx{ChainID: big.NewInt(0), GasTipCap: big.NewInt(0), GasFeeCap: big.NewInt(0), Value: big.NewInt(0)}
	zeroHash, err := zero.SigningHash()
	require.NoError(t, err)
	require.True(t, nilHash.Equal(zeroHash))
}

func TestDynamicFeeTxIntegerFieldBounds(t *testing.T) {
	key := legacyTestPrivateKey(t)
	maxU256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	fields := map[string]func(tx *DynamicFeeTx, n *big.Int){
		"chainId":   func(tx *DynamicFeeTx, n *big.Int) { tx.ChainID = n },
		"gasTipCap": func(tx *DynamicFeeTx, n *big.Int) { tx.GasTipCap = n },
		"gasFeeCap": func(tx *DynamicFeeTx, n *big.Int) { tx.GasFeeCap = n },
		"value":     func(tx *DynamicFeeTx, n *big.Int) { tx.Value = n },
	}
	cases := []struct {
		name string
		n    *big.Int
		want error
	}{
		{"negative", big.NewInt(-1), ErrInvalidInteger},
		{"large negative", new(big.Int).Neg(overflow), ErrInvalidInteger},
		{"overflow", overflow, ErrInvalidInteger},
		{"far overflow", new(big.Int).Lsh(big.NewInt(1), 1024), ErrInvalidInteger},
		{"max", maxU256, nil},
		{"zero", big.NewInt(0), nil},
	}

	for field, set := range fields {
		for _, tt := range cases {
			t.Run(field+"/"+tt.name, func(t *testing.T) {
				v := dynamicFeeVectors(t)[0]
				tx := v.tx
				set(&tx, tt.n)

				_, err := tx.SigningHash()
				if tt.want != nil {
					require.ErrorIs(t, err, tt.want)
					require.ErrorContains(t, err, field)
					require.ErrorIs(t, tx.Sign(key), tt.want)

					tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
					_, err = tx.EncodeRLP()
					require.ErrorIs(t, err, tt.want)
					_, err = tx.Hash()
					require.ErrorIs(t, err, tt.want)
					return
				}
				require.NoError(t, err)
				require.NoError(t, tx.Sign(key))
				_, err = tx.EncodeRLP()
				require.NoError(t, err)
			})
		}
	}
}

func TestDynamicFeeTxInvalidAccessList(t *testing.T) {
	key := legacyTestPrivateKey(t)
	v := dynamicFeeVectors(t)[0]
	address := testAddress(t, "0x0000000000000000000000000000000000000001")

	tests := map[string]AccessList{
		"nil address":     {{Address: nil}},
		"nil storage key": {{Address: address, StorageKeys: []*types.Hash{nil}}},
		"nil key after valid entry": {
			{Address: address, StorageKeys: []*types.Hash{testHash(t, "0x0000000000000000000000000000000000000000000000000000000000000001")}},
			{Address: address, StorageKeys: []*types.Hash{nil}},
		},
	}
	for name, al := range tests {
		t.Run(name, func(t *testing.T) {
			tx := v.tx
			tx.AccessList = al

			_, err := tx.SigningHash()
			require.ErrorIs(t, err, ErrInvalidTransaction)
			require.ErrorContains(t, err, "access list")
			require.ErrorIs(t, tx.Sign(key), ErrInvalidTransaction)

			tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
			_, err = tx.EncodeRLP()
			require.ErrorIs(t, err, ErrInvalidTransaction)
		})
	}
}

func TestDynamicFeeTxExtremeAccessListRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := dynamicFeeVectors(t)[0].tx
	for i := 0; i < 50; i++ {
		entry := AccessTuple{Address: types.NewAddressFromBytes([]byte{byte(i + 1)})}
		for j := 0; j < 20; j++ {
			entry.StorageKeys = append(entry.StorageKeys, types.NewHashFromBytes([]byte{byte(i), byte(j)}))
		}
		tx.AccessList = append(tx.AccessList, entry)
	}
	require.NoError(t, tx.Sign(key))

	raw, err := tx.EncodeRLP()
	require.NoError(t, err)

	var got DynamicFeeTx
	require.NoError(t, got.DecodeRLP(raw))
	requireSameDynamicFee(t, tx, got)
	require.True(t, tx.Signature.Equal(got.Signature))
}

func TestDynamicFeeTxDecodeRLPEmptyAccessListIsNil(t *testing.T) {
	var tx DynamicFeeTx
	require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)))
	require.Nil(t, tx.AccessList)
}

func TestDynamicFeeTxDecodeRLPTypeMismatch(t *testing.T) {
	valid := mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)
	legacy := mustDecodeHex(t, legacyVectors(t)[0].raw)

	tests := map[string][]byte{
		"empty":         nil,
		"type 1":        append([]byte{0x01}, valid[1:]...),
		"type 3":        append([]byte{0x03}, valid[1:]...),
		"type 4":        append([]byte{0x04}, valid[1:]...),
		"legacy":        legacy,
		"missing byte":  valid[1:],
		"double prefix": append([]byte{0x02}, valid...),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var tx DynamicFeeTx
			err := tx.DecodeRLP(raw)
			require.Error(t, err)
			if name != "double prefix" {
				require.ErrorIs(t, err, ErrInvalidTxType)
			}
		})
	}
}

func TestDynamicFeeTxDecodeRLPInvalidEncoding(t *testing.T) {
	valid := mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)
	one := big.NewInt(1)

	list := func(items ...any) []byte {
		enc, err := rlp.Encode(items)
		require.NoError(t, err)
		return append([]byte{0x02}, enc...)
	}
	address := make([]byte, 20)

	tests := map[string][]byte{
		"trailing byte": append(append([]byte{}, valid...), 0x00),
		"truncated":     valid[:len(valid)-1],
		"only type":     {0x02},
		"not a list":    {0x02, 0x83, 0x01, 0x02, 0x03},
		"empty list":    list(),
		"eleven fields": list(one, uint64(0), one, one, uint64(21000), address, one, []byte{}, []any{}, one, one),
		"thirteen fields": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{}, []any{}, one, one, one, one),
		"short to": list(one, uint64(0), one, one, uint64(21000), make([]byte, 19), one, []byte{}, []any{}, one, one, one),
		"long to":  list(one, uint64(0), one, one, uint64(21000), make([]byte, 21), one, []byte{}, []any{}, one, one, one),
		"short access list address": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{},
			[]any{[]any{make([]byte, 19), []any{}}}, one, one, one),
		"short storage key": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{},
			[]any{[]any{address, []any{make([]byte, 31)}}}, one, one, one),
		"long storage key": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{},
			[]any{[]any{address, []any{make([]byte, 33)}}}, one, one, one),
		"access list entry with one item": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{},
			[]any{[]any{address}}, one, one, one),
		"access list not a list": list(
			one, uint64(0), one, one, uint64(21000), address, one, []byte{}, []byte{0x01}, one, one, one),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var tx DynamicFeeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}
}

func TestDynamicFeeTxDecodeRLPInvalidYParity(t *testing.T) {
	v := dynamicFeeVectors(t)[0]
	r, s := sigRS(t, v)

	for _, yParity := range []int64{2, 3, 27, 28, 37} {
		t.Run(big.NewInt(yParity).String(), func(t *testing.T) {
			raw := encodeDynamicFeeSigned(t, v.tx, big.NewInt(yParity), r, s)
			var tx DynamicFeeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}

	huge := new(big.Int).Lsh(big.NewInt(1), 70)
	var tx DynamicFeeTx
	require.Error(t, tx.DecodeRLP(encodeDynamicFeeSigned(t, v.tx, huge, r, s)))

	for _, yParity := range []int64{0, 1} {
		raw := encodeDynamicFeeSigned(t, v.tx, big.NewInt(yParity), r, s)
		require.NoError(t, tx.DecodeRLP(raw))
		require.Equal(t, yParity, tx.Signature.V().Int64())
	}
}

func TestDynamicFeeTxDecodeRLPInvalidSignatureValues(t *testing.T) {
	v := dynamicFeeVectors(t)[0]
	r, s := sigRS(t, v)
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)

	tests := []struct {
		name string
		r, s *big.Int
	}{
		{"zero r", big.NewInt(0), s},
		{"zero s", r, big.NewInt(0)},
		{"r over 256 bits", tooBig, s},
		{"s over 256 bits", r, tooBig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := encodeDynamicFeeSigned(t, v.tx, big.NewInt(0), tt.r, tt.s)
			var tx DynamicFeeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}
}

func TestDynamicFeeTxDecodeRLPRejectsOversizedIntegers(t *testing.T) {
	v := dynamicFeeVectors(t)[0]
	r, s := sigRS(t, v)
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	mutations := map[string]func(tx *DynamicFeeTx){
		"chainId":   func(tx *DynamicFeeTx) { tx.ChainID = overflow },
		"gasTipCap": func(tx *DynamicFeeTx) { tx.GasTipCap = overflow },
		"gasFeeCap": func(tx *DynamicFeeTx) { tx.GasFeeCap = overflow },
		"value":     func(tx *DynamicFeeTx) { tx.Value = overflow },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			tx := v.tx
			mutate(&tx)
			var got DynamicFeeTx
			err := got.DecodeRLP(encodeDynamicFeeSigned(t, tx, big.NewInt(0), r, s))
			require.ErrorIs(t, err, ErrInvalidInteger)
		})
	}
}

func TestDynamicFeeTxDecodeRLPFailureLeavesReceiverUntouched(t *testing.T) {
	tx := DynamicFeeTx{Nonce: 42, ChainID: big.NewInt(7)}
	require.Error(t, tx.DecodeRLP([]byte{0x02, 0xc0}))
	require.Equal(t, uint64(42), tx.Nonce)
	require.Equal(t, int64(7), tx.ChainID.Int64())

	v := dynamicFeeVectors(t)[0]
	r, s := sigRS(t, v)
	bad := v.tx
	bad.Value = new(big.Int).Lsh(big.NewInt(1), 256)
	require.Error(t, tx.DecodeRLP(encodeDynamicFeeSigned(t, bad, big.NewInt(0), r, s)))
	require.Equal(t, uint64(42), tx.Nonce)
}

func TestDynamicFeeTxDecodeRLPResetsReceiver(t *testing.T) {
	vectors := dynamicFeeVectors(t)
	tx := withSignature(t, vectors[2])

	require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, vectors[1].raw)))
	require.Nil(t, tx.To)
	require.Nil(t, tx.AccessList)
	require.Equal(t, uint64(0), tx.Nonce)
	requireSameDynamicFee(t, vectors[1].tx, tx)
}

func TestDynamicFeeTxDecodeRLPDoesNotAliasInput(t *testing.T) {
	raw := mustDecodeHex(t, dynamicFeeVectors(t)[2].raw)
	var tx DynamicFeeTx
	require.NoError(t, tx.DecodeRLP(raw))

	want := append([]byte{}, tx.Data...)
	for i := range raw {
		raw[i] = 0
	}
	require.Equal(t, want, tx.Data)
}

func TestDynamicFeeTxSignThenRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)
	for _, v := range dynamicFeeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)

			var decoded DynamicFeeTx
			require.NoError(t, decoded.DecodeRLP(raw))
			requireSameDynamicFee(t, tx, decoded)
			require.True(t, tx.Signature.Equal(decoded.Signature))
		})
	}
}

const secondTestKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"

type authorizationVector struct {
	key       int
	sigHash   string
	sig       string
	recID     byte
	authority string
}

type setCodeVector struct {
	name    string
	tx      SetCodeTx
	auths   []authorizationVector
	sigHash string
	raw     string
	hash    string
	sig     string
	recID   byte
}

func setCodeVectors(t *testing.T) []setCodeVector {
	return []setCodeVector{
		{
			name: "single authorization",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      9,
				GasTipCap:  mustBigDecimal(t, "1500000000"),
				GasFeeCap:  mustBigDecimal(t, "30000000000"),
				GasLimit:   100000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "0"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "1"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000aa"),
						Nonce:     0,
						Signature: signatureWithRecoveryID(t, "042b827292e0c095b5b0e5a9f9253145fc04ee3c68b2236bda2ad646a963b5ae64a02c24a0fd6805523ff5487f38f172ac68865c4ffcd526906a477d7bca6f01", 0),
					},
				},
			},
			auths: []authorizationVector{
				{key: 1, sigHash: "8d9b12a1365c2b41ad8b2a8d2a7857b0a3c0543a3ec961ae4ed84d654fd289e8", sig: "042b827292e0c095b5b0e5a9f9253145fc04ee3c68b2236bda2ad646a963b5ae64a02c24a0fd6805523ff5487f38f172ac68865c4ffcd526906a477d7bca6f01", recID: 0, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "3942e38034d0d22221148cb5a25945ba2fe5b7b87c33cb1b95aed62640d576a9",
			raw:     "04f8ca01098459682f008506fc23ac00830186a09435353535353535353535353535353535353535358080c0f85cf85a019400000000000000000000000000000000000000aa8080a0042b827292e0c095b5b0e5a9f9253145fc04ee3c68b2236bda2ad646a963b5aea064a02c24a0fd6805523ff5487f38f172ac68865c4ffcd526906a477d7bca6f0180a0335bc722c1669ba60019f08c914067da968205259e8e1697eba5fa390693dac2a00490ef675b056b6920d7a326de3ebf39dc74bccc0e0ddb79e768e076f2fc6052",
			hash:    "dc8c30172e85e4babc71e145d3bbd58108d033e21ce19028fc206bc9440583fa",
			sig:     "335bc722c1669ba60019f08c914067da968205259e8e1697eba5fa390693dac20490ef675b056b6920d7a326de3ebf39dc74bccc0e0ddb79e768e076f2fc6052",
			recID:   0,
		},
		{
			name: "two authorizations any chain with access list",
			tx: SetCodeTx{
				ChainID:   mustBigDecimal(t, "1"),
				Nonce:     3,
				GasTipCap: mustBigDecimal(t, "2"),
				GasFeeCap: mustBigDecimal(t, "7"),
				GasLimit:  200000,
				To:        testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:     mustBigDecimal(t, "1000"),
				Data:      mustDecodeHex(t, "deadbeef"),
				AccessList: AccessList{
					{Address: testAddress(t, "0x3535353535353535353535353535353535353535"), StorageKeys: []*types.Hash{testHash(t, "0x0000000000000000000000000000000000000000000000000000000000000001")}},
				},
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "0"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000bb"),
						Nonce:     5,
						Signature: signatureWithRecoveryID(t, "1fcc43d62d241a9d1f7613a6df05c297d181e2e892cea3f24aaecc17b954ac2c33741c06a948496d549a8a92e4146270d29df0f9324710adfb1f8e3fb138ecb2", 1),
					},
					{
						ChainID:   mustBigDecimal(t, "0"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000cc"),
						Nonce:     6,
						Signature: signatureWithRecoveryID(t, "feff899cb9e37f4c02d812efbdd18e73e526e15d54bf44f3271044ff57ac478e78cdfd8952054fe026820a1b17285659e4e29b7d3a50ea1c9370ba63ded74fa9", 0),
					},
				},
			},
			auths: []authorizationVector{
				{key: 0, sigHash: "ecad45f0d05b806986293d77f66768b19dcfb0064233e7cb94f4d74fb0ee2023", sig: "1fcc43d62d241a9d1f7613a6df05c297d181e2e892cea3f24aaecc17b954ac2c33741c06a948496d549a8a92e4146270d29df0f9324710adfb1f8e3fb138ecb2", recID: 1, authority: "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"},
				{key: 1, sigHash: "1da62b4937c9f1a685b93994bd34bd59c6dc8bd0d6c8efc235f42025d0acc6a3", sig: "feff899cb9e37f4c02d812efbdd18e73e526e15d54bf44f3271044ff57ac478e78cdfd8952054fe026820a1b17285659e4e29b7d3a50ea1c9370ba63ded74fa9", recID: 0, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "73a52a897f592c3a4c2c1e3d5da4e3afe3b77925b10a78e54260258b179f182c",
			raw:     "04f9015c0103020783030d409435353535353535353535353535353535353535358203e884deadbeeff838f7943535353535353535353535353535353535353535e1a00000000000000000000000000000000000000000000000000000000000000001f8b8f85a809400000000000000000000000000000000000000bb0501a01fcc43d62d241a9d1f7613a6df05c297d181e2e892cea3f24aaecc17b954ac2ca033741c06a948496d549a8a92e4146270d29df0f9324710adfb1f8e3fb138ecb2f85a809400000000000000000000000000000000000000cc0680a0feff899cb9e37f4c02d812efbdd18e73e526e15d54bf44f3271044ff57ac478ea078cdfd8952054fe026820a1b17285659e4e29b7d3a50ea1c9370ba63ded74fa901a097d39a855b0d4f8927641c91db37d0f2c61757adebe8fbe433290dbd78c70852a05b1919260f78788897b417bef00a46ade8ddb62b0592d158062ed3497c40db9d",
			hash:    "fe4936b9181d6c5a2b823bd0ba0963e32bc4cb2d0df8a248968c0de1badc2f7f",
			sig:     "97d39a855b0d4f8927641c91db37d0f2c61757adebe8fbe433290dbd78c708525b1919260f78788897b417bef00a46ade8ddb62b0592d158062ed3497c40db9d",
			recID:   1,
		},
		{
			name: "delegation reset to zero address",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "11155111"),
				Nonce:      1,
				GasTipCap:  mustBigDecimal(t, "1"),
				GasFeeCap:  mustBigDecimal(t, "1"),
				GasLimit:   60000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "0"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "11155111"),
						Address:   testAddress(t, "0x0000000000000000000000000000000000000000"),
						Nonce:     18446744073709551614,
						Signature: signatureWithRecoveryID(t, "16bf793051be0e6d2b11af7f9f7d290c655f41f258161bdce78f67b26883d6841c23983457a5e6e01d20a298f7e993b5b45a93e1cfc3915657aef546489408c5", 1),
					},
				},
			},
			auths: []authorizationVector{
				{key: 1, sigHash: "e8a0f88c726b0c026f9f0446e0787dcb23f4c6f5f51a8ed40c2606bf34e3c4cd", sig: "16bf793051be0e6d2b11af7f9f7d290c655f41f258161bdce78f67b26883d6841c23983457a5e6e01d20a298f7e993b5b45a93e1cfc3915657aef546489408c5", recID: 1, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "aaf7170eed7410926da293942131e3de6d6b62fc536d3c088d7603b57a5a1b1a",
			raw:     "04f8ce83aa36a701010182ea609435353535353535353535353535353535353535358080c0f867f86583aa36a794000000000000000000000000000000000000000088fffffffffffffffe01a016bf793051be0e6d2b11af7f9f7d290c655f41f258161bdce78f67b26883d684a01c23983457a5e6e01d20a298f7e993b5b45a93e1cfc3915657aef546489408c580a0947bc8d76c122fe524b575a562e738e75462835cf6b1ebb20c405e513c0faacaa073193038ca623154657e244b5365e30c1a14346d4da3fa01ff51b352d130af98",
			hash:    "0e2e34239ce20c75a8d4e1c750d02bfe5a23826822e745e3ca9488908ea4ecd1",
			sig:     "947bc8d76c122fe524b575a562e738e75462835cf6b1ebb20c405e513c0faaca73193038ca623154657e244b5365e30c1a14346d4da3fa01ff51b352d130af98",
			recID:   0,
		},
		{
			name: "authorization with recovery id 1",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      2,
				GasTipCap:  mustBigDecimal(t, "3"),
				GasFeeCap:  mustBigDecimal(t, "9"),
				GasLimit:   90000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "7"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "1"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000dd"),
						Nonce:     0,
						Signature: signatureWithRecoveryID(t, "fbe14001c569b08593aa80ddae05b1d1c29188b5757c66d55dbf675b672d39f5792ca337f636c63db2bb66c6147d1da2736ff1f8e4139f15b456429c17e3224e", 1),
					},
				},
			},
			auths: []authorizationVector{
				{key: 1, sigHash: "35a77f4cced8760906170f4c37a9c4b01f46fe9dfcad361446c35fdb8aca2594", sig: "fbe14001c569b08593aa80ddae05b1d1c29188b5757c66d55dbf675b672d39f5792ca337f636c63db2bb66c6147d1da2736ff1f8e4139f15b456429c17e3224e", recID: 1, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "8c99ff3956b261348aa989eddb81113322a1f887e61d46947be1516c1d95c95d",
			raw:     "04f8c10102030983015f909435353535353535353535353535353535353535350780c0f85cf85a019400000000000000000000000000000000000000dd8001a0fbe14001c569b08593aa80ddae05b1d1c29188b5757c66d55dbf675b672d39f5a0792ca337f636c63db2bb66c6147d1da2736ff1f8e4139f15b456429c17e3224e80a0dde81cebf506c3c64350e77e7e91c1bb48e7643e170ca3e63048a0035b0b3ff9a053a9d7b04e9034a6748d6febabea25dac9e1bd80d8f9fbb84cb6ffe6abd5f833",
			hash:    "629f4c930e8292efa6f63d69c3c3fa299ca4ae51a1b41df4563e30dd5df8b75e",
			sig:     "dde81cebf506c3c64350e77e7e91c1bb48e7643e170ca3e63048a0035b0b3ff953a9d7b04e9034a6748d6febabea25dac9e1bd80d8f9fbb84cb6ffe6abd5f833",
			recID:   0,
		},
		{
			name: "transaction recovery id 1",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      0,
				GasTipCap:  mustBigDecimal(t, "3"),
				GasFeeCap:  mustBigDecimal(t, "9"),
				GasLimit:   90000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "7"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "1"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000ee"),
						Nonce:     1,
						Signature: signatureWithRecoveryID(t, "a7a0d06fb9d8472d6b98bd6c8f94b4d67c98b02656a079cc4744386ef77df2417a156d58a695e131f725e38af9193d8657dde498ec180518b0beb7cd3d2bda0d", 0),
					},
				},
			},
			auths: []authorizationVector{
				{key: 1, sigHash: "38cafcea802000d81b83d5a31e48ed92be0a3223c75403da360cd6b2207e525e", sig: "a7a0d06fb9d8472d6b98bd6c8f94b4d67c98b02656a079cc4744386ef77df2417a156d58a695e131f725e38af9193d8657dde498ec180518b0beb7cd3d2bda0d", recID: 0, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "723c003cc500e5b23e73fa50673f24ede3d02ee2d962e407f9a632ecf0ac4b62",
			raw:     "04f8c10180030983015f909435353535353535353535353535353535353535350780c0f85cf85a019400000000000000000000000000000000000000ee0180a0a7a0d06fb9d8472d6b98bd6c8f94b4d67c98b02656a079cc4744386ef77df241a07a156d58a695e131f725e38af9193d8657dde498ec180518b0beb7cd3d2bda0d01a0a4ec32f708c57c0065716e48ea6111a2774615785f416ba0938a06c9be6bc0fba07cfa33f23e56b6808935956e52367c93ed6d23985a2fd5c7111adf0a95ff438c",
			hash:    "869aff429c88d9dee284d8fdd81ea06301f75ee14148744ca26afc0c8624b3dd",
			sig:     "a4ec32f708c57c0065716e48ea6111a2774615785f416ba0938a06c9be6bc0fb7cfa33f23e56b6808935956e52367c93ed6d23985a2fd5c7111adf0a95ff438c",
			recID:   1,
		},
		{
			name: "large chain id and zero fees",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "1380012617"),
				Nonce:      2,
				GasTipCap:  mustBigDecimal(t, "0"),
				GasFeeCap:  mustBigDecimal(t, "0"),
				GasLimit:   21000,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "0"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "1380012617"),
						Address:   testAddress(t, "0x00000000000000000000000000000000000000ff"),
						Nonce:     0,
						Signature: signatureWithRecoveryID(t, "721d2f1765e93b285ba7be1561ab7ee7e961e88a2550471f21ce9e258dee60005a36ad287d5c120d86730e94bf34c55f6d3743be1c6b41d450fb3fc2b350b158", 1),
					},
				},
			},
			auths: []authorizationVector{
				{key: 0, sigHash: "c5445a668a94a102c3e70d4062eeae0120cfae336038bf93e9df4ff6a2089a66", sig: "721d2f1765e93b285ba7be1561ab7ee7e961e88a2550471f21ce9e258dee60005a36ad287d5c120d86730e94bf34c55f6d3743be1c6b41d450fb3fc2b350b158", recID: 1, authority: "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"},
			},
			sigHash: "55bc8423e42cacb3fb29b52d8b252ea94dbebf0d3ff5f56a4f42cc7ec77789ee",
			raw:     "04f8c884524152490280808252089435353535353535353535353535353535353535358080c0f860f85e84524152499400000000000000000000000000000000000000ff8001a0721d2f1765e93b285ba7be1561ab7ee7e961e88a2550471f21ce9e258dee6000a05a36ad287d5c120d86730e94bf34c55f6d3743be1c6b41d450fb3fc2b350b15880a0d8cc478b750b6fa0801533341d22291c94d4a64c92e836ca47f8eba41682f022a03e6f5a57142f2a0b16b8e3a47d783fc1bceafeb933dcb74f4876f2ca4c948d3b",
			hash:    "fa0037f0d91f0be29edc22c5bbaa487380e87a94afec0050698f8b3d2375d253",
			sig:     "d8cc478b750b6fa0801533341d22291c94d4a64c92e836ca47f8eba41682f0223e6f5a57142f2a0b16b8e3a47d783fc1bceafeb933dcb74f4876f2ca4c948d3b",
			recID:   0,
		},
		{
			name: "max values",
			tx: SetCodeTx{
				ChainID:    mustBigDecimal(t, "1"),
				Nonce:      18446744073709551615,
				GasTipCap:  mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				GasFeeCap:  mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				GasLimit:   18446744073709551615,
				To:         testAddress(t, "0x3535353535353535353535353535353535353535"),
				Value:      mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
				Data:       nil,
				AccessList: nil,
				AuthList: []Authorization{
					{
						ChainID:   mustBigDecimal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639935"),
						Address:   testAddress(t, "0x3535353535353535353535353535353535353535"),
						Nonce:     18446744073709551615,
						Signature: signatureWithRecoveryID(t, "b9ab36e114f12bb453a2772044bd7fb29210909d6d5386d1e91cbb4ac3bc509409d341c4fbcb28f544e5b5d5242225355c5338cfec776cd2d3b6c90c9bfdd649", 0),
					},
				},
			},
			auths: []authorizationVector{
				{key: 1, sigHash: "b71fd43b861934a82e5dbc5d3823e95bcf6f83b4ee1f24be9abd5e588bad56b2", sig: "b9ab36e114f12bb453a2772044bd7fb29210909d6d5386d1e91cbb4ac3bc509409d341c4fbcb28f544e5b5d5242225355c5338cfec776cd2d3b6c90c9bfdd649", recID: 0, authority: "0x71562b71999873db5b286df957af199ec94617f7"},
			},
			sigHash: "0d5c80eafd076fa654859bfc47872b0404a12ec42c749981147e0248caca1378",
			raw:     "04f901560188ffffffffffffffffa0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff88ffffffffffffffff943535353535353535353535353535353535353535a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c0f884f882a0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff94353535353535353535353535353535353535353588ffffffffffffffff80a0b9ab36e114f12bb453a2772044bd7fb29210909d6d5386d1e91cbb4ac3bc5094a009d341c4fbcb28f544e5b5d5242225355c5338cfec776cd2d3b6c90c9bfdd64901a09f02b96299ee094c6333da3ae76ed341d2efd6d7434a9b4f8895e18206ff5d63a01844b12963f69e9a8cf3839fe90588b697997bfa6778ac4f348279abda448ce0",
			hash:    "4e5a1fe55a186956fe0b03792386c30f35826a7ba2b7ebd2d94c8467206a745b",
			sig:     "9f02b96299ee094c6333da3ae76ed341d2efd6d7434a9b4f8895e18206ff5d631844b12963f69e9a8cf3839fe90588b697997bfa6778ac4f348279abda448ce0",
			recID:   1,
		},
	}
}

func setCodeTestKey(t *testing.T, index int) *types.PrivateKey {
	t.Helper()
	hexKey := legacyTestKey
	if index == 1 {
		hexKey = secondTestKey
	}
	key, err := types.NewPrivateKeyFromHex(hexKey)
	require.NoError(t, err)
	return key
}

func setCodeWithSignature(t *testing.T, v setCodeVector) SetCodeTx {
	t.Helper()
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
	return tx
}

func setCodeSigRS(t *testing.T, v setCodeVector) (*big.Int, *big.Int) {
	t.Helper()
	b := mustDecodeHex(t, v.sig)
	return new(big.Int).SetBytes(b[:32]), new(big.Int).SetBytes(b[32:])
}

func authListRLPForTest(t *testing.T, list []Authorization) []rlp.AuthorizationRLP {
	t.Helper()
	out := make([]rlp.AuthorizationRLP, len(list))
	for i, auth := range list {
		yParity, err := rlp.RecoveryID(auth.Signature)
		require.NoError(t, err)
		out[i] = rlp.AuthorizationRLP{
			ChainID: auth.ChainID,
			Address: *rlp.FromAddress(auth.Address),
			Nonce:   auth.Nonce,
			V:       yParity,
			R:       auth.Signature.R(),
			S:       auth.Signature.S(),
		}
	}
	return out
}

func encodeSetCodeSigned(t *testing.T, tx SetCodeTx, authList []rlp.AuthorizationRLP, yParity, r, s *big.Int) []byte {
	t.Helper()
	accessList, err := accessListToRLP(tx.AccessList)
	require.NoError(t, err)
	enc, err := rlp.Encode(&rlp.SetCodeSignedRLP{
		ChainID:    tx.ChainID,
		Nonce:      tx.Nonce,
		GasTipCap:  tx.GasTipCap,
		GasFeeCap:  tx.GasFeeCap,
		GasLimit:   tx.GasLimit,
		To:         *rlp.FromAddress(tx.To),
		Value:      tx.Value,
		Data:       tx.Data,
		AccessList: accessList,
		AuthList:   authList,
		V:          yParity,
		R:          r,
		S:          s,
	})
	require.NoError(t, err)
	return append([]byte{0x04}, enc...)
}

func requireSameSetCode(t *testing.T, want, got SetCodeTx) {
	t.Helper()
	require.Equal(t, 0, want.ChainID.Cmp(got.ChainID))
	require.Equal(t, want.Nonce, got.Nonce)
	require.Equal(t, 0, want.GasTipCap.Cmp(got.GasTipCap))
	require.Equal(t, 0, want.GasFeeCap.Cmp(got.GasFeeCap))
	require.Equal(t, want.GasLimit, got.GasLimit)
	require.True(t, want.To.Equal(got.To))
	require.Equal(t, 0, want.Value.Cmp(got.Value))
	require.Equal(t, string(want.Data), string(got.Data))
	require.Equal(t, len(want.AccessList), len(got.AccessList))
	for i := range want.AccessList {
		require.True(t, want.AccessList[i].Address.Equal(got.AccessList[i].Address))
		require.Equal(t, len(want.AccessList[i].StorageKeys), len(got.AccessList[i].StorageKeys))
		for j := range want.AccessList[i].StorageKeys {
			require.True(t, want.AccessList[i].StorageKeys[j].Equal(got.AccessList[i].StorageKeys[j]))
		}
	}
	require.Equal(t, len(want.AuthList), len(got.AuthList))
	for i := range want.AuthList {
		require.Equal(t, 0, want.AuthList[i].ChainID.Cmp(got.AuthList[i].ChainID))
		require.True(t, want.AuthList[i].Address.Equal(got.AuthList[i].Address))
		require.Equal(t, want.AuthList[i].Nonce, got.AuthList[i].Nonce)
		require.True(t, want.AuthList[i].Signature.Equal(got.AuthList[i].Signature))
	}
}

func TestAuthorizationSigningHashMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		for i, av := range v.auths {
			t.Run(v.name+"/"+big.NewInt(int64(i)).String(), func(t *testing.T) {
				auth := v.tx.AuthList[i]
				got, err := auth.SigningHash()
				require.NoError(t, err)
				require.Equal(t, mustDecodeHex(t, av.sigHash), got.Bytes())
			})
		}
	}
}

func TestAuthorizationSignMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		for i, av := range v.auths {
			t.Run(v.name+"/"+big.NewInt(int64(i)).String(), func(t *testing.T) {
				auth := v.tx.AuthList[i]
				auth.Signature = nil
				require.NoError(t, auth.Sign(setCodeTestKey(t, av.key)))
				require.True(t, signatureWithRecoveryID(t, av.sig, av.recID).Equal(auth.Signature))
			})
		}
	}
}

func TestAuthorizationAuthorityMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		for i, av := range v.auths {
			t.Run(v.name+"/"+big.NewInt(int64(i)).String(), func(t *testing.T) {
				auth := v.tx.AuthList[i]
				got, err := auth.Authority()
				require.NoError(t, err)
				require.True(t, testAddress(t, av.authority).Equal(got))

				signer := PubkeyToAddress(setCodeTestKey(t, av.key).PublicKey())
				require.True(t, signer.Equal(got))
			})
		}
	}
}

func TestAuthorizationSigningHashIgnoresSignature(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	unsigned := auth
	unsigned.Signature = nil

	a, err := auth.SigningHash()
	require.NoError(t, err)
	b, err := unsigned.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestAuthorizationSigningHashDependsOnEveryField(t *testing.T) {
	base := setCodeVectors(t)[0].tx.AuthList[0]
	baseHash, err := base.SigningHash()
	require.NoError(t, err)

	mutations := map[string]func(a *Authorization){
		"chainID": func(a *Authorization) { a.ChainID = big.NewInt(2) },
		"address": func(a *Authorization) { a.Address = testAddress(t, "0x0000000000000000000000000000000000000001") },
		"nonce":   func(a *Authorization) { a.Nonce++ },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			auth := base
			mutate(&auth)
			h, err := auth.SigningHash()
			require.NoError(t, err)
			require.False(t, baseHash.Equal(h))
		})
	}
}

func TestAuthorizationSigningHashUsesMagicPrefix(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	got, err := auth.SigningHash()
	require.NoError(t, err)

	enc, err := rlp.Encode(&rlp.AuthorizationSigRLP{
		ChainID: auth.ChainID,
		Address: *rlp.FromAddress(auth.Address),
		Nonce:   auth.Nonce,
	})
	require.NoError(t, err)
	require.False(t, Keccak256(enc).Equal(got))
	require.True(t, Keccak256(append([]byte{0x05}, enc...)).Equal(got))
}

func TestAuthorizationNilChainIDIsZero(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	nilChain := auth
	nilChain.ChainID = nil
	zeroChain := auth
	zeroChain.ChainID = big.NewInt(0)

	a, err := nilChain.SigningHash()
	require.NoError(t, err)
	b, err := zeroChain.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestAuthorizationChainIDBounds(t *testing.T) {
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)
	auth := setCodeVectors(t)[0].tx.AuthList[0]

	tests := []struct {
		name string
		n    *big.Int
		want error
	}{
		{"negative", big.NewInt(-1), ErrInvalidInteger},
		{"overflow", overflow, ErrInvalidInteger},
		{"max", new(big.Int).Sub(overflow, big.NewInt(1)), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := auth
			a.ChainID = tt.n
			_, err := a.SigningHash()
			if tt.want == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
			require.ErrorIs(t, a.Sign(legacyTestPrivateKey(t)), tt.want)
		})
	}
}

func TestAuthorizationNilAddress(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	auth.Address = nil

	_, err := auth.SigningHash()
	require.ErrorIs(t, err, ErrInvalidTransaction)
	require.ErrorContains(t, err, "address")
	require.ErrorIs(t, auth.Sign(legacyTestPrivateKey(t)), ErrInvalidTransaction)
	_, err = auth.Authority()
	require.ErrorIs(t, err, ErrInvalidTransaction)
}

func TestAuthorizationSignPropagatesSignerError(t *testing.T) {
	boom := errors.New("boom")
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	auth.Signature = nil

	require.ErrorIs(t, auth.Sign(failingSigner{err: boom}), boom)
	require.Nil(t, auth.Signature)
}

func TestAuthorizationAuthorityUnsigned(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	auth.Signature = nil
	_, err := auth.Authority()
	require.ErrorIs(t, err, ErrUnsigned)
}

func TestAuthorizationAuthorityRejectsHighS(t *testing.T) {
	v := setCodeVectors(t)[0]
	auth := v.tx.AuthList[0]
	r := auth.Signature.R()
	s := auth.Signature.S()

	high, err := rlp.SignatureFromRS(r, new(big.Int).Sub(secp256k1Order(t), s), byte(auth.Signature.V().Uint64())^1)
	require.NoError(t, err)
	auth.Signature = high

	_, err = auth.Authority()
	require.Error(t, err)
}

func TestAuthorizationAuthorityRejectsHighRecoveryID(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	high, err := rlp.SignatureFromRS(auth.Signature.R(), auth.Signature.S(), 2)
	require.NoError(t, err)
	auth.Signature = high

	_, err = auth.Authority()
	require.Error(t, err)
}

func TestAuthorizationAuthorityDiffersWhenPayloadChanges(t *testing.T) {
	auth := setCodeVectors(t)[0].tx.AuthList[0]
	want, err := auth.Authority()
	require.NoError(t, err)

	auth.Nonce++
	got, err := auth.Authority()
	if err == nil {
		require.False(t, want.Equal(got))
	}
}

func TestSetCodeTxType(t *testing.T) {
	require.Equal(t, SetCodeTxType, (&SetCodeTx{}).Type())
}

func TestSetCodeTxSigningHashMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			got, err := tx.SigningHash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.sigHash), got.Bytes())
		})
	}
}

func TestSetCodeTxSigningHashIgnoresSignature(t *testing.T) {
	v := setCodeVectors(t)[0]
	unsigned := v.tx
	signed := setCodeWithSignature(t, v)

	a, err := unsigned.SigningHash()
	require.NoError(t, err)
	b, err := signed.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestSetCodeTxEncodeRLPMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := setCodeWithSignature(t, v)
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), got)
		})
	}
}

func TestSetCodeTxHashMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := setCodeWithSignature(t, v)
			got, err := tx.Hash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.hash), got.Bytes())
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.True(t, Keccak256(raw).Equal(got))
		})
	}
}

func TestSetCodeTxDecodeRLPMatchesGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			var got SetCodeTx
			require.NoError(t, got.DecodeRLP(mustDecodeHex(t, v.raw)))
			requireSameSetCode(t, v.tx, got)
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(got.Signature))
		})
	}
}

func TestSetCodeTxDecodeThenEncodeIsIdentity(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			raw := mustDecodeHex(t, v.raw)
			var tx SetCodeTx
			require.NoError(t, tx.DecodeRLP(raw))
			got, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, raw, got)
		})
	}
}

func TestSetCodeTxDecodedAuthoritiesMatchGeth(t *testing.T) {
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			var tx SetCodeTx
			require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, v.raw)))
			for i, av := range v.auths {
				got, err := tx.AuthList[i].Authority()
				require.NoError(t, err)
				require.True(t, testAddress(t, av.authority).Equal(got))
			}
		})
	}
}

func TestSetCodeTxSignMatchesGeth(t *testing.T) {
	key := legacyTestPrivateKey(t)
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(tx.Signature))

			raw, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.raw), raw)
		})
	}
}

func TestSetCodeTxSignRecoversSender(t *testing.T) {
	key := legacyTestPrivateKey(t)
	want := PubkeyToAddress(key.PublicKey())

	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))

			digest, err := tx.SigningHash()
			require.NoError(t, err)
			ok, err := VerifyAddress(digest, tx.Signature, want)
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestSetCodeTxSignPropagatesSignerError(t *testing.T) {
	boom := errors.New("boom")
	tx := setCodeVectors(t)[0].tx

	require.ErrorIs(t, tx.Sign(failingSigner{err: boom}), boom)
	require.Nil(t, tx.Signature)
}

func TestSetCodeTxSignOverwritesSignature(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := setCodeVectors(t)[0].tx

	require.NoError(t, tx.Sign(key))
	first := tx.Signature

	tx.Nonce++
	require.NoError(t, tx.Sign(key))
	require.False(t, first.Equal(tx.Signature))
}

func TestSetCodeTxSignThenRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)
	for _, v := range setCodeVectors(t) {
		t.Run(v.name, func(t *testing.T) {
			tx := v.tx
			require.NoError(t, tx.Sign(key))
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)

			var decoded SetCodeTx
			require.NoError(t, decoded.DecodeRLP(raw))
			requireSameSetCode(t, tx, decoded)
			require.True(t, tx.Signature.Equal(decoded.Signature))
		})
	}
}

func TestSetCodeTxSignedAuthorizationsThenSign(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := SetCodeTx{
		ChainID:   big.NewInt(1),
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(2),
		GasLimit:  100000,
		To:        legacyTestAddress(t),
		Value:     big.NewInt(0),
		AuthList: []Authorization{
			{ChainID: big.NewInt(1), Address: legacyTestAddress(t), Nonce: 1},
			{ChainID: big.NewInt(0), Address: legacyTestAddress(t), Nonce: 2},
		},
	}
	for i := range tx.AuthList {
		require.NoError(t, tx.AuthList[i].Sign(setCodeTestKey(t, i)))
	}
	require.NoError(t, tx.Sign(key))

	raw, err := tx.EncodeRLP()
	require.NoError(t, err)

	var got SetCodeTx
	require.NoError(t, got.DecodeRLP(raw))
	for i := range got.AuthList {
		authority, err := got.AuthList[i].Authority()
		require.NoError(t, err)
		require.True(t, PubkeyToAddress(setCodeTestKey(t, i).PublicKey()).Equal(authority))
	}
}

func TestSetCodeTxEncodeRLPUnsigned(t *testing.T) {
	tx := setCodeVectors(t)[0].tx
	_, err := tx.EncodeRLP()
	require.ErrorIs(t, err, ErrUnsigned)
	_, err = tx.Hash()
	require.ErrorIs(t, err, ErrUnsigned)
}

func TestSetCodeTxEncodeRLPRejectsHighRecoveryID(t *testing.T) {
	v := setCodeVectors(t)[0]
	tx := v.tx
	tx.Signature = signatureWithRecoveryID(t, v.sig, 2)
	_, err := tx.EncodeRLP()
	require.Error(t, err)
}

func TestSetCodeTxSigningHashDependsOnEveryField(t *testing.T) {
	base := setCodeVectors(t)[1].tx
	baseHash, err := base.SigningHash()
	require.NoError(t, err)

	otherAddress := testAddress(t, "0x0000000000000000000000000000000000000001")
	cloneAuths := func(tx *SetCodeTx) []Authorization {
		return append([]Authorization{}, tx.AuthList...)
	}

	mutations := map[string]func(tx *SetCodeTx){
		"chainID":   func(tx *SetCodeTx) { tx.ChainID = big.NewInt(2) },
		"nonce":     func(tx *SetCodeTx) { tx.Nonce++ },
		"gasTipCap": func(tx *SetCodeTx) { tx.GasTipCap = big.NewInt(99) },
		"gasFeeCap": func(tx *SetCodeTx) { tx.GasFeeCap = big.NewInt(99) },
		"gasLimit":  func(tx *SetCodeTx) { tx.GasLimit++ },
		"to":        func(tx *SetCodeTx) { tx.To = otherAddress },
		"value":     func(tx *SetCodeTx) { tx.Value = big.NewInt(1) },
		"data":      func(tx *SetCodeTx) { tx.Data = []byte{0x01} },
		"access list removed": func(tx *SetCodeTx) {
			tx.AccessList = nil
		},
		"authorization chainID": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)
			tx.AuthList[0].ChainID = big.NewInt(5)
		},
		"authorization address": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)
			tx.AuthList[0].Address = otherAddress
		},
		"authorization nonce": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)
			tx.AuthList[0].Nonce++
		},
		"authorization signature": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)
			tx.AuthList[0].Signature = tx.AuthList[1].Signature
		},
		"authorization order": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)
			tx.AuthList[0], tx.AuthList[1] = tx.AuthList[1], tx.AuthList[0]
		},
		"authorization removed": func(tx *SetCodeTx) {
			tx.AuthList = cloneAuths(tx)[:1]
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			tx := base
			mutate(&tx)
			h, err := tx.SigningHash()
			require.NoError(t, err)
			require.False(t, baseHash.Equal(h))
		})
	}
}

func TestSetCodeTxSigningHashDiffersFromDynamicFee(t *testing.T) {
	v := setCodeVectors(t)[0]
	setCode := v.tx
	dynamic := DynamicFeeTx{
		ChainID:   setCode.ChainID,
		Nonce:     setCode.Nonce,
		GasTipCap: setCode.GasTipCap,
		GasFeeCap: setCode.GasFeeCap,
		GasLimit:  setCode.GasLimit,
		To:        setCode.To,
		Value:     setCode.Value,
		Data:      setCode.Data,
	}

	a, err := setCode.SigningHash()
	require.NoError(t, err)
	b, err := dynamic.SigningHash()
	require.NoError(t, err)
	require.False(t, a.Equal(b))
}

func TestSetCodeTxNilIntegersEncodeAsZero(t *testing.T) {
	base := setCodeVectors(t)[0].tx
	nilInts := base
	nilInts.ChainID, nilInts.GasTipCap, nilInts.GasFeeCap, nilInts.Value = nil, nil, nil, nil
	zeroInts := base
	zeroInts.ChainID, zeroInts.GasTipCap, zeroInts.GasFeeCap, zeroInts.Value = big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)

	a, err := nilInts.SigningHash()
	require.NoError(t, err)
	b, err := zeroInts.SigningHash()
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

func TestSetCodeTxNilEqualsEmptyDataAndAccessList(t *testing.T) {
	key := legacyTestPrivateKey(t)
	nilFields := setCodeVectors(t)[0].tx
	emptyFields := nilFields
	emptyFields.Data = []byte{}
	emptyFields.AccessList = AccessList{}

	require.NoError(t, nilFields.Sign(key))
	require.NoError(t, emptyFields.Sign(key))

	a, err := nilFields.EncodeRLP()
	require.NoError(t, err)
	b, err := emptyFields.EncodeRLP()
	require.NoError(t, err)
	require.Equal(t, a, b)
}

func TestSetCodeTxRequiresDestination(t *testing.T) {
	key := legacyTestPrivateKey(t)
	v := setCodeVectors(t)[0]
	tx := setCodeWithSignature(t, v)
	tx.To = nil

	_, err := tx.SigningHash()
	require.ErrorIs(t, err, ErrInvalidTransaction)
	require.ErrorContains(t, err, "destination")
	require.ErrorIs(t, tx.Sign(key), ErrInvalidTransaction)
	_, err = tx.EncodeRLP()
	require.ErrorIs(t, err, ErrInvalidTransaction)
	_, err = tx.Hash()
	require.ErrorIs(t, err, ErrInvalidTransaction)
}

func TestSetCodeTxRequiresAuthorizations(t *testing.T) {
	key := legacyTestPrivateKey(t)
	v := setCodeVectors(t)[0]

	for name, list := range map[string][]Authorization{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			tx := setCodeWithSignature(t, v)
			tx.AuthList = list

			_, err := tx.SigningHash()
			require.ErrorIs(t, err, ErrInvalidTransaction)
			require.ErrorContains(t, err, "authorization")
			require.ErrorIs(t, tx.Sign(key), ErrInvalidTransaction)
			_, err = tx.EncodeRLP()
			require.ErrorIs(t, err, ErrInvalidTransaction)
		})
	}
}

func TestSetCodeTxInvalidAuthorizations(t *testing.T) {
	key := legacyTestPrivateKey(t)
	v := setCodeVectors(t)[1]
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	high, err := rlp.SignatureFromRS(v.tx.AuthList[0].Signature.R(), v.tx.AuthList[0].Signature.S(), 2)
	require.NoError(t, err)

	tests := []struct {
		name   string
		mutate func(a *Authorization)
		want   error
	}{
		{"nil address", func(a *Authorization) { a.Address = nil }, ErrInvalidTransaction},
		{"unsigned", func(a *Authorization) { a.Signature = nil }, ErrUnsigned},
		{"negative chain id", func(a *Authorization) { a.ChainID = big.NewInt(-1) }, ErrInvalidInteger},
		{"chain id overflow", func(a *Authorization) { a.ChainID = overflow }, ErrInvalidInteger},
		{"high recovery id", func(a *Authorization) { a.Signature = high }, nil},
	}
	for _, tt := range tests {
		for index := 0; index < 2; index++ {
			t.Run(tt.name+"/"+big.NewInt(int64(index)).String(), func(t *testing.T) {
				tx := setCodeWithSignature(t, v)
				tx.AuthList = append([]Authorization{}, tx.AuthList...)
				tt.mutate(&tx.AuthList[index])

				_, err := tx.SigningHash()
				require.Error(t, err)
				require.ErrorContains(t, err, "authorization "+big.NewInt(int64(index)).String())
				if tt.want != nil {
					require.ErrorIs(t, err, tt.want)
					require.ErrorIs(t, tx.Sign(key), tt.want)
					_, err = tx.EncodeRLP()
					require.ErrorIs(t, err, tt.want)
				}
			})
		}
	}
}

func TestSetCodeTxIntegerFieldBounds(t *testing.T) {
	key := legacyTestPrivateKey(t)
	maxU256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	fields := map[string]func(tx *SetCodeTx, n *big.Int){
		"chainId":   func(tx *SetCodeTx, n *big.Int) { tx.ChainID = n },
		"gasTipCap": func(tx *SetCodeTx, n *big.Int) { tx.GasTipCap = n },
		"gasFeeCap": func(tx *SetCodeTx, n *big.Int) { tx.GasFeeCap = n },
		"value":     func(tx *SetCodeTx, n *big.Int) { tx.Value = n },
	}
	cases := []struct {
		name string
		n    *big.Int
		want error
	}{
		{"negative", big.NewInt(-1), ErrInvalidInteger},
		{"large negative", new(big.Int).Neg(overflow), ErrInvalidInteger},
		{"overflow", overflow, ErrInvalidInteger},
		{"far overflow", new(big.Int).Lsh(big.NewInt(1), 1024), ErrInvalidInteger},
		{"max", maxU256, nil},
		{"zero", big.NewInt(0), nil},
	}

	for field, set := range fields {
		for _, tt := range cases {
			t.Run(field+"/"+tt.name, func(t *testing.T) {
				v := setCodeVectors(t)[0]
				tx := v.tx
				set(&tx, tt.n)

				_, err := tx.SigningHash()
				if tt.want != nil {
					require.ErrorIs(t, err, tt.want)
					require.ErrorContains(t, err, field)
					require.ErrorIs(t, tx.Sign(key), tt.want)

					tx.Signature = signatureWithRecoveryID(t, v.sig, v.recID)
					_, err = tx.EncodeRLP()
					require.ErrorIs(t, err, tt.want)
					_, err = tx.Hash()
					require.ErrorIs(t, err, tt.want)
					return
				}
				require.NoError(t, err)
				require.NoError(t, tx.Sign(key))
				_, err = tx.EncodeRLP()
				require.NoError(t, err)
			})
		}
	}
}

func TestSetCodeTxInvalidAccessList(t *testing.T) {
	key := legacyTestPrivateKey(t)
	v := setCodeVectors(t)[0]
	address := testAddress(t, "0x0000000000000000000000000000000000000001")

	tests := map[string]AccessList{
		"nil address":     {{Address: nil}},
		"nil storage key": {{Address: address, StorageKeys: []*types.Hash{nil}}},
	}
	for name, al := range tests {
		t.Run(name, func(t *testing.T) {
			tx := setCodeWithSignature(t, v)
			tx.AccessList = al

			_, err := tx.SigningHash()
			require.ErrorIs(t, err, ErrInvalidTransaction)
			require.ErrorContains(t, err, "access list")
			require.ErrorIs(t, tx.Sign(key), ErrInvalidTransaction)
			_, err = tx.EncodeRLP()
			require.ErrorIs(t, err, ErrInvalidTransaction)
		})
	}
}

func TestSetCodeTxManyAuthorizationsRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)
	tx := setCodeVectors(t)[0].tx
	tx.AuthList = nil
	for i := 0; i < 40; i++ {
		auth := Authorization{
			ChainID: big.NewInt(int64(i % 3)),
			Address: types.NewAddressFromBytes([]byte{byte(i + 1)}),
			Nonce:   uint64(i),
		}
		require.NoError(t, auth.Sign(setCodeTestKey(t, i%2)))
		tx.AuthList = append(tx.AuthList, auth)
	}
	require.NoError(t, tx.Sign(key))

	raw, err := tx.EncodeRLP()
	require.NoError(t, err)

	var got SetCodeTx
	require.NoError(t, got.DecodeRLP(raw))
	requireSameSetCode(t, tx, got)
	require.True(t, tx.Signature.Equal(got.Signature))
}

func TestSetCodeTxDecodeRLPEmptyAccessListIsNil(t *testing.T) {
	var tx SetCodeTx
	require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, setCodeVectors(t)[0].raw)))
	require.Nil(t, tx.AccessList)
}

func TestSetCodeTxDecodeRLPTypeMismatch(t *testing.T) {
	valid := mustDecodeHex(t, setCodeVectors(t)[0].raw)
	dynamic := mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)
	legacy := mustDecodeHex(t, legacyVectors(t)[0].raw)

	tests := map[string][]byte{
		"empty":        nil,
		"type 1":       append([]byte{0x01}, valid[1:]...),
		"type 2":       append([]byte{0x02}, valid[1:]...),
		"type 3":       append([]byte{0x03}, valid[1:]...),
		"dynamic fee":  dynamic,
		"legacy":       legacy,
		"missing byte": valid[1:],
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var tx SetCodeTx
			require.ErrorIs(t, tx.DecodeRLP(raw), ErrInvalidTxType)
		})
	}

	var tx SetCodeTx
	require.Error(t, tx.DecodeRLP(append([]byte{0x04}, valid...)))
}

func TestOtherTypesRejectSetCodeEncoding(t *testing.T) {
	raw := mustDecodeHex(t, setCodeVectors(t)[0].raw)

	var dynamic DynamicFeeTx
	require.ErrorIs(t, dynamic.DecodeRLP(raw), ErrInvalidTxType)

	var legacy LegacyTx
	require.Error(t, legacy.DecodeRLP(raw))
}

func TestSetCodeTxDecodeRLPInvalidEncoding(t *testing.T) {
	v := setCodeVectors(t)[0]
	valid := mustDecodeHex(t, v.raw)
	one := big.NewInt(1)
	address := make([]byte, 20)
	auth := []any{one, address, uint64(0), uint8(0), one, one}

	list := func(items ...any) []byte {
		enc, err := rlp.Encode(items)
		require.NoError(t, err)
		return append([]byte{0x04}, enc...)
	}
	head := func(to []byte, authList any, tail ...any) []byte {
		items := []any{one, uint64(0), one, one, uint64(21000), to, one, []byte{}, []any{}, authList}
		return list(append(items, tail...)...)
	}

	tests := map[string][]byte{
		"trailing byte":           append(append([]byte{}, valid...), 0x00),
		"truncated":               valid[:len(valid)-1],
		"only type":               {0x04},
		"not a list":              {0x04, 0x83, 0x01, 0x02, 0x03},
		"empty list":              list(),
		"twelve fields":           list(one, uint64(0), one, one, uint64(21000), address, one, []byte{}, []any{}, []any{auth}, one, one),
		"fourteen fields":         head(address, []any{auth}, one, one, one, one),
		"empty destination":       head([]byte{}, []any{auth}, one, one, one),
		"short destination":       head(make([]byte, 19), []any{auth}, one, one, one),
		"long destination":        head(make([]byte, 21), []any{auth}, one, one, one),
		"authorization list type": head(address, []byte{0x01}, one, one, one),
		"authorization not list":  head(address, []any{[]byte{0x01}}, one, one, one),
		"authorization 5 fields":  head(address, []any{[]any{one, address, uint64(0), uint8(0), one}}, one, one, one),
		"authorization 7 fields":  head(address, []any{[]any{one, address, uint64(0), uint8(0), one, one, one}}, one, one, one),
		"authorization short address": head(address, []any{
			[]any{one, make([]byte, 19), uint64(0), uint8(0), one, one}}, one, one, one),
		"authorization long address": head(address, []any{
			[]any{one, make([]byte, 21), uint64(0), uint8(0), one, one}}, one, one, one),
		"authorization empty address": head(address, []any{
			[]any{one, []byte{}, uint64(0), uint8(0), one, one}}, one, one, one),
		"authorization yParity 256": head(address, []any{
			[]any{one, address, uint64(0), uint16(256), one, one}}, one, one, one),
		"authorization nonce too big": head(address, []any{
			[]any{one, address, new(big.Int).Lsh(big.NewInt(1), 64), uint8(0), one, one}}, one, one, one),
		"authorization yParity list": head(address, []any{
			[]any{one, address, uint64(0), []any{}, one, one}}, one, one, one),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var tx SetCodeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}
}

func TestSetCodeTxDecodeRLPEmptyAuthorizationList(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	raw := encodeSetCodeSigned(t, v.tx, []rlp.AuthorizationRLP{}, big.NewInt(0), r, s)

	var tx SetCodeTx
	err := tx.DecodeRLP(raw)
	require.ErrorIs(t, err, ErrInvalidTransaction)
	require.ErrorContains(t, err, "authorization")
}

func TestSetCodeTxDecodeRLPInvalidAuthorizationYParity(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)

	for _, yParity := range []uint8{2, 3, 27, 28, 255} {
		t.Run(big.NewInt(int64(yParity)).String(), func(t *testing.T) {
			authList := authListRLPForTest(t, v.tx.AuthList)
			authList[0].V = yParity
			raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(0), r, s)

			var tx SetCodeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}

	for _, yParity := range []uint8{0, 1} {
		authList := authListRLPForTest(t, v.tx.AuthList)
		authList[0].V = yParity
		raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(0), r, s)

		var tx SetCodeTx
		require.NoError(t, tx.DecodeRLP(raw))
		require.Equal(t, int64(yParity), tx.AuthList[0].Signature.V().Int64())
	}
}

func TestSetCodeTxDecodeRLPInvalidAuthorizationSignature(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)

	tests := map[string]func(a *rlp.AuthorizationRLP){
		"zero r":          func(a *rlp.AuthorizationRLP) { a.R = big.NewInt(0) },
		"zero s":          func(a *rlp.AuthorizationRLP) { a.S = big.NewInt(0) },
		"r over 256 bits": func(a *rlp.AuthorizationRLP) { a.R = tooBig },
		"s over 256 bits": func(a *rlp.AuthorizationRLP) { a.S = tooBig },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			authList := authListRLPForTest(t, v.tx.AuthList)
			mutate(&authList[0])
			raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(0), r, s)

			var tx SetCodeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}
}

func TestSetCodeTxDecodeRLPAuthorizationChainIDOverflow(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)

	authList := authListRLPForTest(t, v.tx.AuthList)
	authList[0].ChainID = new(big.Int).Lsh(big.NewInt(1), 256)
	raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(0), r, s)

	var tx SetCodeTx
	require.ErrorIs(t, tx.DecodeRLP(raw), ErrInvalidInteger)
}

func TestSetCodeTxDecodeRLPInvalidYParity(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	authList := authListRLPForTest(t, v.tx.AuthList)

	for _, yParity := range []int64{2, 3, 27, 28, 37} {
		t.Run(big.NewInt(yParity).String(), func(t *testing.T) {
			raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(yParity), r, s)
			var tx SetCodeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}

	var tx SetCodeTx
	huge := new(big.Int).Lsh(big.NewInt(1), 70)
	require.Error(t, tx.DecodeRLP(encodeSetCodeSigned(t, v.tx, authList, huge, r, s)))

	for _, yParity := range []int64{0, 1} {
		raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(yParity), r, s)
		require.NoError(t, tx.DecodeRLP(raw))
		require.Equal(t, yParity, tx.Signature.V().Int64())
	}
}

func TestSetCodeTxDecodeRLPInvalidSignatureValues(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	authList := authListRLPForTest(t, v.tx.AuthList)
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)

	tests := []struct {
		name string
		r, s *big.Int
	}{
		{"zero r", big.NewInt(0), s},
		{"zero s", r, big.NewInt(0)},
		{"r over 256 bits", tooBig, s},
		{"s over 256 bits", r, tooBig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := encodeSetCodeSigned(t, v.tx, authList, big.NewInt(0), tt.r, tt.s)
			var tx SetCodeTx
			require.Error(t, tx.DecodeRLP(raw))
		})
	}
}

func TestSetCodeTxDecodeRLPRejectsOversizedIntegers(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	authList := authListRLPForTest(t, v.tx.AuthList)
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	mutations := map[string]func(tx *SetCodeTx){
		"chainId":   func(tx *SetCodeTx) { tx.ChainID = overflow },
		"gasTipCap": func(tx *SetCodeTx) { tx.GasTipCap = overflow },
		"gasFeeCap": func(tx *SetCodeTx) { tx.GasFeeCap = overflow },
		"value":     func(tx *SetCodeTx) { tx.Value = overflow },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			tx := v.tx
			mutate(&tx)
			var got SetCodeTx
			err := got.DecodeRLP(encodeSetCodeSigned(t, tx, authList, big.NewInt(0), r, s))
			require.ErrorIs(t, err, ErrInvalidInteger)
		})
	}
}

func TestSetCodeTxDecodeRLPFailureLeavesReceiverUntouched(t *testing.T) {
	tx := SetCodeTx{Nonce: 42, ChainID: big.NewInt(7)}
	require.Error(t, tx.DecodeRLP([]byte{0x04, 0xc0}))
	require.Equal(t, uint64(42), tx.Nonce)
	require.Equal(t, int64(7), tx.ChainID.Int64())

	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)
	bad := v.tx
	bad.Value = new(big.Int).Lsh(big.NewInt(1), 256)
	raw := encodeSetCodeSigned(t, bad, authListRLPForTest(t, v.tx.AuthList), big.NewInt(0), r, s)
	require.Error(t, tx.DecodeRLP(raw))
	require.Equal(t, uint64(42), tx.Nonce)
	require.Len(t, tx.AuthList, 0)
}

func TestSetCodeTxDecodeRLPResetsReceiver(t *testing.T) {
	vectors := setCodeVectors(t)
	tx := setCodeWithSignature(t, vectors[1])
	require.NotNil(t, tx.AccessList)

	require.NoError(t, tx.DecodeRLP(mustDecodeHex(t, vectors[0].raw)))
	require.Nil(t, tx.AccessList)
	require.Empty(t, tx.Data)
	require.Len(t, tx.AuthList, 1)
	requireSameSetCode(t, vectors[0].tx, tx)
}

func TestSetCodeTxDecodeRLPDoesNotAliasInput(t *testing.T) {
	raw := mustDecodeHex(t, setCodeVectors(t)[1].raw)
	var tx SetCodeTx
	require.NoError(t, tx.DecodeRLP(raw))

	want := append([]byte{}, tx.Data...)
	for i := range raw {
		raw[i] = 0
	}
	require.Equal(t, want, tx.Data)
	require.True(t, tx.To != nil && tx.AuthList[0].Address != nil)
}

func TestDecodeTransactionDispatchesByType(t *testing.T) {
	legacy := legacyVectors(t)
	dynamic := dynamicFeeVectors(t)
	setCode := setCodeVectors(t)

	t.Run("legacy", func(t *testing.T) {
		for _, v := range legacy {
			tx, err := DecodeTransaction(mustDecodeHex(t, v.raw))
			require.NoError(t, err)
			require.IsType(t, &LegacyTx{}, tx)
			require.Equal(t, LegacyTxType, tx.Type())
			require.True(t, signatureWithRecoveryID(t, v.sig, v.recID).Equal(tx.(*LegacyTx).Signature))
		}
	})
	t.Run("dynamic fee", func(t *testing.T) {
		for _, v := range dynamic {
			tx, err := DecodeTransaction(mustDecodeHex(t, v.raw))
			require.NoError(t, err)
			require.IsType(t, &DynamicFeeTx{}, tx)
			require.Equal(t, DynamicFeeTxType, tx.Type())
			requireSameDynamicFee(t, v.tx, *tx.(*DynamicFeeTx))
		}
	})
	t.Run("set code", func(t *testing.T) {
		for _, v := range setCode {
			tx, err := DecodeTransaction(mustDecodeHex(t, v.raw))
			require.NoError(t, err)
			require.IsType(t, &SetCodeTx{}, tx)
			require.Equal(t, SetCodeTxType, tx.Type())
			requireSameSetCode(t, v.tx, *tx.(*SetCodeTx))
		}
	})
}

func TestDecodeTransactionHashAndEncodingMatchGeth(t *testing.T) {
	type vector struct{ name, raw, hash string }
	var vectors []vector
	for _, v := range legacyVectors(t) {
		vectors = append(vectors, vector{"legacy/" + v.name, v.raw, v.hash})
	}
	for _, v := range dynamicFeeVectors(t) {
		vectors = append(vectors, vector{"dynamic fee/" + v.name, v.raw, v.hash})
	}
	for _, v := range setCodeVectors(t) {
		vectors = append(vectors, vector{"set code/" + v.name, v.raw, v.hash})
	}

	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			raw := mustDecodeHex(t, v.raw)
			tx, err := DecodeTransaction(raw)
			require.NoError(t, err)

			again, err := tx.EncodeRLP()
			require.NoError(t, err)
			require.Equal(t, raw, again)

			hash, err := tx.Hash()
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, v.hash), hash.Bytes())
		})
	}
}

func TestDecodeTransactionEmpty(t *testing.T) {
	for _, raw := range [][]byte{nil, {}} {
		tx, err := DecodeTransaction(raw)
		require.ErrorIs(t, err, ErrInvalidTxType)
		require.ErrorContains(t, err, "empty")
		require.Nil(t, tx)
	}
}

func TestDecodeTransactionUnsupportedTypes(t *testing.T) {
	body := mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)[1:]

	for _, first := range []byte{0x00, 0x01, 0x03, 0x05, 0x06, 0x10, 0x7f, 0x80, 0x81, 0xb7, 0xbf} {
		t.Run(hex.EncodeToString([]byte{first}), func(t *testing.T) {
			tx, err := DecodeTransaction(append([]byte{first}, body...))
			require.ErrorIs(t, err, ErrInvalidTxType)
			require.ErrorContains(t, err, "0x"+hex.EncodeToString([]byte{first}))
			require.Nil(t, tx)
		})
	}
}

func TestDecodeTransactionPropagatesDecodeErrors(t *testing.T) {
	legacy := mustDecodeHex(t, legacyVectors(t)[0].raw)
	dynamic := mustDecodeHex(t, dynamicFeeVectors(t)[0].raw)
	setCode := mustDecodeHex(t, setCodeVectors(t)[0].raw)

	tests := map[string][]byte{
		"legacy truncated":      legacy[:len(legacy)-1],
		"legacy trailing":       append(append([]byte{}, legacy...), 0x00),
		"legacy empty list":     {0xc0},
		"dynamic fee truncated": dynamic[:len(dynamic)-1],
		"dynamic fee only type": {0x02},
		"dynamic fee trailing":  append(append([]byte{}, dynamic...), 0x00),
		"set code truncated":    setCode[:len(setCode)-1],
		"set code only type":    {0x04},
		"set code trailing":     append(append([]byte{}, setCode...), 0x00),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			tx, err := DecodeTransaction(raw)
			require.Error(t, err)
			require.Nil(t, tx)
		})
	}
}

func TestDecodeTransactionValidationErrors(t *testing.T) {
	v := setCodeVectors(t)[0]
	r, s := setCodeSigRS(t, v)

	raw := encodeSetCodeSigned(t, v.tx, []rlp.AuthorizationRLP{}, big.NewInt(0), r, s)
	_, err := DecodeTransaction(raw)
	require.ErrorIs(t, err, ErrInvalidTransaction)
	require.ErrorContains(t, err, "authorization")

	overflow := v.tx
	overflow.Value = new(big.Int).Lsh(big.NewInt(1), 256)
	raw = encodeSetCodeSigned(t, overflow, authListRLPForTest(t, v.tx.AuthList), big.NewInt(0), r, s)
	_, err = DecodeTransaction(raw)
	require.ErrorIs(t, err, ErrInvalidInteger)
}

func TestDecodeTransactionReturnsIndependentValues(t *testing.T) {
	raw := mustDecodeHex(t, legacyVectors(t)[0].raw)

	a, err := DecodeTransaction(raw)
	require.NoError(t, err)
	b, err := DecodeTransaction(raw)
	require.NoError(t, err)
	require.NotSame(t, a, b)

	a.(*LegacyTx).Nonce++
	require.NotEqual(t, a.(*LegacyTx).Nonce, b.(*LegacyTx).Nonce)
}

func TestDecodeTransactionSignThenRoundTrip(t *testing.T) {
	key := legacyTestPrivateKey(t)

	txs := []Transaction{
		func() Transaction { tx := legacyVectors(t)[0].tx; return &tx }(),
		func() Transaction { tx := dynamicFeeVectors(t)[2].tx; return &tx }(),
		func() Transaction { tx := setCodeVectors(t)[1].tx; return &tx }(),
	}
	for _, tx := range txs {
		t.Run("type "+hex.EncodeToString([]byte{byte(tx.Type())}), func(t *testing.T) {
			require.NoError(t, tx.Sign(key))
			raw, err := tx.EncodeRLP()
			require.NoError(t, err)

			decoded, err := DecodeTransaction(raw)
			require.NoError(t, err)
			require.Equal(t, tx.Type(), decoded.Type())

			a, err := tx.Hash()
			require.NoError(t, err)
			b, err := decoded.Hash()
			require.NoError(t, err)
			require.True(t, a.Equal(b))
		})
	}
}

const senderTestAddress = "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"

type senderCase struct {
	name string
	tx   Transaction
	sig  string
	rec  byte
}

func senderCases(t *testing.T) []senderCase {
	t.Helper()
	var cases []senderCase
	for _, v := range legacyVectors(t) {
		tx := v.tx
		cases = append(cases, senderCase{"legacy/" + v.name, &tx, v.sig, v.recID})
	}
	for _, v := range dynamicFeeVectors(t) {
		tx := v.tx
		cases = append(cases, senderCase{"dynamic fee/" + v.name, &tx, v.sig, v.recID})
	}
	for _, v := range setCodeVectors(t) {
		tx := v.tx
		cases = append(cases, senderCase{"set code/" + v.name, &tx, v.sig, v.recID})
	}
	return cases
}

func setTransactionSignature(tx Transaction, sig *types.Signature) {
	switch tx := tx.(type) {
	case *LegacyTx:
		tx.Signature = sig
	case *DynamicFeeTx:
		tx.Signature = sig
	case *SetCodeTx:
		tx.Signature = sig
	}
}

func secp256k1Order(t *testing.T) *big.Int {
	t.Helper()
	n, ok := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	require.True(t, ok)
	return n
}

func highSSignature(t *testing.T, rs string, recID byte) *types.Signature {
	t.Helper()
	b := mustDecodeHex(t, rs)
	sig, err := rlp.SignatureFromRS(new(big.Int).SetBytes(b[:32]), new(big.Int).Sub(secp256k1Order(t), new(big.Int).SetBytes(b[32:])), recID^1)
	require.NoError(t, err)
	return sig
}

func TestSenderMatchesGeth(t *testing.T) {
	want := testAddress(t, senderTestAddress)
	for _, c := range senderCases(t) {
		t.Run(c.name, func(t *testing.T) {
			setTransactionSignature(c.tx, signatureWithRecoveryID(t, c.sig, c.rec))
			got, err := c.tx.Sender()
			require.NoError(t, err)
			require.True(t, want.Equal(got))
		})
	}
}

func TestSenderAfterDecodeTransaction(t *testing.T) {
	want := testAddress(t, senderTestAddress)
	var raws []string
	for _, v := range legacyVectors(t) {
		raws = append(raws, v.raw)
	}
	for _, v := range dynamicFeeVectors(t) {
		raws = append(raws, v.raw)
	}
	for _, v := range setCodeVectors(t) {
		raws = append(raws, v.raw)
	}

	for _, raw := range raws {
		tx, err := DecodeTransaction(mustDecodeHex(t, raw))
		require.NoError(t, err)
		got, err := tx.Sender()
		require.NoError(t, err)
		require.True(t, want.Equal(got))
	}
}

func TestSenderMatchesSigningKey(t *testing.T) {
	for index := 0; index < 2; index++ {
		key := setCodeTestKey(t, index)
		want := PubkeyToAddress(key.PublicKey())
		for _, c := range senderCases(t) {
			t.Run(c.name+"/"+big.NewInt(int64(index)).String(), func(t *testing.T) {
				require.NoError(t, c.tx.Sign(key))
				got, err := c.tx.Sender()
				require.NoError(t, err)
				require.True(t, want.Equal(got))
			})
		}
	}
}

func TestSenderUnsigned(t *testing.T) {
	for _, c := range senderCases(t) {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.tx.Sender()
			require.ErrorIs(t, err, ErrUnsigned)
			require.Nil(t, got)
		})
	}
}

func TestSenderRejectsHighS(t *testing.T) {
	for _, c := range senderCases(t) {
		t.Run(c.name, func(t *testing.T) {
			setTransactionSignature(c.tx, highSSignature(t, c.sig, c.rec))
			got, err := c.tx.Sender()
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestSenderRejectsHighRecoveryID(t *testing.T) {
	for _, c := range senderCases(t) {
		t.Run(c.name, func(t *testing.T) {
			b := mustDecodeHex(t, c.sig)
			sig, err := rlp.SignatureFromRS(new(big.Int).SetBytes(b[:32]), new(big.Int).SetBytes(b[32:]), 2)
			require.NoError(t, err)
			setTransactionSignature(c.tx, sig)

			got, err := c.tx.Sender()
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestSenderChangesWhenTransactionIsTampered(t *testing.T) {
	want := testAddress(t, senderTestAddress)
	for _, c := range senderCases(t) {
		t.Run(c.name, func(t *testing.T) {
			setTransactionSignature(c.tx, signatureWithRecoveryID(t, c.sig, c.rec))
			switch tx := c.tx.(type) {
			case *LegacyTx:
				tx.Nonce++
			case *DynamicFeeTx:
				tx.Nonce++
			case *SetCodeTx:
				tx.Nonce++
			}

			got, err := c.tx.Sender()
			if err == nil {
				require.False(t, want.Equal(got))
			}
		})
	}
}

func TestSenderPropagatesInvalidFieldErrors(t *testing.T) {
	legacy := legacyVectors(t)[0]
	legacyTx := legacy.tx
	legacyTx.Signature = signatureWithRecoveryID(t, legacy.sig, legacy.recID)
	legacyTx.Value = big.NewInt(-1)
	_, err := legacyTx.Sender()
	require.ErrorIs(t, err, ErrInvalidInteger)

	dynamic := dynamicFeeVectors(t)[0]
	dynamicTx := withSignature(t, dynamic)
	dynamicTx.GasFeeCap = new(big.Int).Lsh(big.NewInt(1), 256)
	_, err = dynamicTx.Sender()
	require.ErrorIs(t, err, ErrInvalidInteger)

	setCode := setCodeVectors(t)[0]
	setCodeTx := setCodeWithSignature(t, setCode)
	setCodeTx.To = nil
	_, err = setCodeTx.Sender()
	require.ErrorIs(t, err, ErrInvalidTransaction)

	setCodeTx = setCodeWithSignature(t, setCode)
	setCodeTx.AuthList = append([]Authorization{}, setCodeTx.AuthList...)
	setCodeTx.AuthList[0].Signature = nil
	_, err = setCodeTx.Sender()
	require.ErrorIs(t, err, ErrUnsigned)
}
