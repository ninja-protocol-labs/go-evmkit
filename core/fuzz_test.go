package core

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzKeccak256$' -fuzztime=10s ./core
func FuzzKeccak256(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("abc"))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, data []byte) {
		got1 := Keccak256(data)
		got2 := Keccak256(data)
		require.Len(t, got1.Bytes(), 32)
		require.Equal(t, got1.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzEIP191Hash$' -fuzztime=10s ./core
func FuzzEIP191Hash(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, msg []byte) {
		got := EIP191Hash(msg)
		require.Len(t, got.Bytes(), 32)
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeArrayOfUint8$' -fuzztime=10s ./core
func FuzzEIP712EncodeArrayOfUint8(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})
	f.Add(make([]byte, 256))

	schema := EIP712Types{"Simple": {{Name: "a", Type: "uint8[]"}}}

	f.Fuzz(func(t *testing.T, raw []byte) {
		_, err := schema.EncodeData("Simple", map[string]any{"a": raw})
		require.NoError(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataAddressField$' -fuzztime=10s ./core
func FuzzEIP712EncodeDataAddressField(f *testing.F) {
	f.Add([]byte{}, false)
	f.Add(make([]byte, 20), false)
	f.Add(make([]byte, 20), true)
	f.Add(make([]byte, 5), true)

	schema := EIP712Types{"Simple": {{Name: "a", Type: "address"}}}

	f.Fuzz(func(t *testing.T, raw []byte, wrongType bool) {
		var value any
		if wrongType {
			value = raw
		} else {
			value = types.NewAddressFromBytes(raw)
		}
		_, _ = schema.EncodeData("Simple", map[string]any{"a": value})
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataCrashSafety$' -fuzztime=10s ./core
func FuzzEIP712EncodeDataCrashSafety(f *testing.F) {
	schema := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail": {
			{Name: "from", Type: "Person"},
			{Name: "to", Type: "Person"},
			{Name: "contents", Type: "string"},
			{Name: "nums", Type: "uint256[]"},
		},
	}

	f.Add("hello", []byte{1, 2, 3}, int8(5))
	f.Add("", []byte(nil), int8(0))

	f.Fuzz(func(t *testing.T, s string, b []byte, n int8) {
		values := []any{
			s, b, n, nil,
			map[string]any{"name": s, "wallet": b},
			[]any{s, b, n},
		}
		for _, v := range values {
			_, _ = schema.EncodeData("Mail", map[string]any{
				"from":     v,
				"to":       v,
				"contents": v,
				"nums":     v,
			})
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEIP1014Address$' -fuzztime=10s ./core
func FuzzEIP1014Address(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32), []byte{})
	f.Add(make([]byte, 20), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, deployerBytes []byte, saltBytes []byte, initCode []byte) {
		deployer := types.NewAddressFromBytes(deployerBytes)

		var salt [32]byte
		copy(salt[:], saltBytes)

		got1 := EIP1014Address(deployer, salt, initCode)
		got2 := EIP1014Address(deployer, salt, initCode)
		require.Equal(t, got1, got2)
	})
}

// go test -run '^$' -fuzz '^FuzzPubkeyToAddressFromRawKey$' -fuzztime=10s ./core
func FuzzPubkeyToAddressFromRawKey(f *testing.F) {
	f.Add(make([]byte, 32))
	one := make([]byte, 32)
	one[31] = 1
	f.Add(one)

	f.Fuzz(func(t *testing.T, raw []byte) {
		key, err := types.NewPrivateKeyFromBytes(raw)
		if err != nil {
			return
		}

		addr := PubkeyToAddress(key.PublicKey())
		require.Len(t, addr.Bytes(), 20)

		digest := Keccak256([]byte("hello"))
		sig, err := key.Sign(digest)
		require.NoError(t, err)

		ok, err := VerifyAddress(digest, sig, addr)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

// go test -run '^$' -fuzz '^FuzzVerifyAddressRoundTrip$' -fuzztime=10s ./core
func FuzzVerifyAddressRoundTrip(f *testing.F) {
	key, err := types.GeneratePrivateKey()
	if err != nil {
		f.Fatal(err)
	}
	address := PubkeyToAddress(key.PublicKey())

	f.Add([]byte("hello"))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, msg []byte) {
		digest := Keccak256(msg)
		sig, err := key.Sign(digest)
		require.NoError(t, err)

		ok, err := VerifyAddress(digest, sig, address)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

// go test -run '^$' -fuzz '^FuzzVerifyAddressWrongAddress$' -fuzztime=10s ./core
func FuzzVerifyAddressWrongAddress(f *testing.F) {
	key, err := types.GeneratePrivateKey()
	if err != nil {
		f.Fatal(err)
	}
	digest := Keccak256([]byte("hello"))
	sig, err := key.Sign(digest)
	if err != nil {
		f.Fatal(err)
	}

	f.Add(make([]byte, 20))
	f.Add([]byte{0xff})

	f.Fuzz(func(t *testing.T, addrBytes []byte) {
		addr := types.NewAddressFromBytes(addrBytes)
		_, err := VerifyAddress(digest, sig, addr)
		require.NoError(t, err)
	})
}

func FuzzLegacyDecodeCanonical(f *testing.F) {
	f.Add(mustDecodeHexF("f86c098504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a0499aa1110848b179aa0f228e20faa3ba68b350e1feeab49638c6b8ce40ea56aea0053ec9b43dcea26f8d10b43a44bdfafae5b1b26462367921079005d4d274e06d"))
	f.Add(mustDecodeHexF("f865018477359400825208943535353535353535353535353535353535353535823039801ba08b658f1eda41f671154751160cfd4dc026e427c48a49d1393a8e700cac8f6b2ca03d4e25002bde45fa449cbf4a3f649a465a3376b956e48757860e8cc5ae5df45f"))
	f.Add([]byte{0xc0})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var tx LegacyTx
		if err := tx.DecodeRLP(data); err != nil {
			return
		}
		enc, err := tx.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, data, enc)
	})
}

func FuzzLegacySignRoundTrip(f *testing.F) {
	f.Add(uint64(0), []byte{}, uint64(21000), []byte{}, []byte{}, []byte{}, uint64(1))
	f.Add(uint64(9), []byte{0x04, 0xa8}, uint64(21000), make([]byte, 20), []byte{0x01}, []byte{0xaa}, uint64(0))
	f.Add(uint64(1<<63), make([]byte, 32), uint64(1<<63), make([]byte, 20), make([]byte, 32), make([]byte, 100), uint64(1<<62))

	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	if err != nil {
		f.Fatal(err)
	}
	want := PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, nonce uint64, gasPrice []byte, gasLimit uint64, to, value, data []byte, chainID uint64) {
		if len(gasPrice) > 32 || len(value) > 32 {
			t.Skip()
		}
		tx := LegacyTx{
			ChainID:  new(big.Int).SetUint64(chainID),
			Nonce:    nonce,
			GasPrice: new(big.Int).SetBytes(gasPrice),
			GasLimit: gasLimit,
			Value:    new(big.Int).SetBytes(value),
			Data:     data,
		}
		if len(to) == types.AddressLength {
			tx.To = types.NewAddressFromBytes(to)
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
		require.Equal(t, string(tx.Data), string(got.Data))
		require.True(t, tx.To.Equal(got.To))
		require.True(t, tx.Signature.Equal(got.Signature))

		again, err := got.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, raw, again)

		digest, err := got.SigningHash()
		require.NoError(t, err)
		ok, err := VerifyAddress(digest, got.Signature, want)
		require.NoError(t, err)
		require.True(t, ok)

		sender, err := got.Sender()
		require.NoError(t, err)
		require.True(t, want.Equal(sender))
	})
}

func FuzzDynamicFeeDecodeCanonical(f *testing.F) {
	f.Add(mustDecodeHexF("02f87301098459682f008506fc23ac00825208943535353535353535353535353535353535353535880de0b6b3a764000080c001a0a2a0cb81cd4d8e240f38279012ee02b696002ebd27203a64d4acd6cc60d2f139a029de4d469feeadf9d8bd087fbd31c50c015928b4ac387409758415b8e301fe57"))
	f.Add(mustDecodeHexF("02f9011f83aa36a782012c020782ea6094353535353535353535353535353535353535353580b844a9059cbb00000000000000000000000035353535353535353535353535353535353535350000000000000000000000000000000000000000000000000000000000000001f872f859943535353535353535353535353535353535353535f842a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000001d6940000000000000000000000000000000000000001c001a077d78238660067a518b7725cf7d64aa65a5cfd1ac0b13f175eb5c0e9b7cceb4ea002ffc76be8ff4bca0096f173a39df34a7afea6529b2d9393d74e4b213c78e70c"))
	f.Add(mustDecodeHexF("02f85c0180843b9aca008477359400830493e08080856080604052c080a0fb01d4a851187ba170117585b16e34daf1669fbf9cfc7e49a7b491ee65d948b0a0085bad6f450392057b227c6af7206f570a6df58f554d78906bfb3e8b12d94b17"))
	f.Add([]byte{0x02, 0xc0})
	f.Add([]byte{0x02})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var tx DynamicFeeTx
		if err := tx.DecodeRLP(data); err != nil {
			return
		}
		enc, err := tx.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, data, enc)
	})
}

func FuzzDynamicFeeSignRoundTrip(f *testing.F) {
	f.Add(uint64(0), []byte{}, []byte{}, uint64(21000), []byte{}, []byte{}, []byte{}, []byte{}, uint64(1))
	f.Add(uint64(9), []byte{0x59, 0x68}, []byte{0x06, 0xfc}, uint64(21000), make([]byte, 20), []byte{0x01}, []byte{0xaa}, make([]byte, 52), uint64(0))
	f.Add(uint64(1<<63), make([]byte, 32), make([]byte, 32), uint64(1<<63), make([]byte, 20), make([]byte, 32), make([]byte, 100), make([]byte, 20+32*3), uint64(1<<62))

	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	if err != nil {
		f.Fatal(err)
	}
	want := PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, nonce uint64, tip, fee []byte, gasLimit uint64, to, value, data, access []byte, chainID uint64) {
		if len(tip) > 32 || len(fee) > 32 || len(value) > 32 {
			t.Skip()
		}
		tx := DynamicFeeTx{
			ChainID:   new(big.Int).SetUint64(chainID),
			Nonce:     nonce,
			GasTipCap: new(big.Int).SetBytes(tip),
			GasFeeCap: new(big.Int).SetBytes(fee),
			GasLimit:  gasLimit,
			Value:     new(big.Int).SetBytes(value),
			Data:      data,
		}
		if len(to) == types.AddressLength {
			tx.To = types.NewAddressFromBytes(to)
		}
		for len(access) >= types.AddressLength {
			entry := AccessTuple{Address: types.NewAddressFromBytes(access[:types.AddressLength])}
			access = access[types.AddressLength:]
			for len(access) >= types.HashLength && len(entry.StorageKeys) < 3 {
				entry.StorageKeys = append(entry.StorageKeys, types.NewHashFromBytes(access[:types.HashLength]))
				access = access[types.HashLength:]
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

		again, err := got.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, raw, again)

		digest, err := got.SigningHash()
		require.NoError(t, err)
		ok, err := VerifyAddress(digest, got.Signature, want)
		require.NoError(t, err)
		require.True(t, ok)

		sender, err := got.Sender()
		require.NoError(t, err)
		require.True(t, want.Equal(sender))
	})
}

func FuzzSetCodeDecodeCanonical(f *testing.F) {
	f.Add(mustDecodeHexF("04f8ca01098459682f008506fc23ac00830186a09435353535353535353535353535353535353535358080c0f85cf85a019400000000000000000000000000000000000000aa8080a0042b827292e0c095b5b0e5a9f9253145fc04ee3c68b2236bda2ad646a963b5aea064a02c24a0fd6805523ff5487f38f172ac68865c4ffcd526906a477d7bca6f0180a0335bc722c1669ba60019f08c914067da968205259e8e1697eba5fa390693dac2a00490ef675b056b6920d7a326de3ebf39dc74bccc0e0ddb79e768e076f2fc6052"))
	f.Add(mustDecodeHexF("04f9015c0103020783030d409435353535353535353535353535353535353535358203e884deadbeeff838f7943535353535353535353535353535353535353535e1a00000000000000000000000000000000000000000000000000000000000000001f8b8f85a809400000000000000000000000000000000000000bb0501a01fcc43d62d241a9d1f7613a6df05c297d181e2e892cea3f24aaecc17b954ac2ca033741c06a948496d549a8a92e4146270d29df0f9324710adfb1f8e3fb138ecb2f85a809400000000000000000000000000000000000000cc0680a0feff899cb9e37f4c02d812efbdd18e73e526e15d54bf44f3271044ff57ac478ea078cdfd8952054fe026820a1b17285659e4e29b7d3a50ea1c9370ba63ded74fa901a097d39a855b0d4f8927641c91db37d0f2c61757adebe8fbe433290dbd78c70852a05b1919260f78788897b417bef00a46ade8ddb62b0592d158062ed3497c40db9d"))
	f.Add(mustDecodeHexF("04f8ce83aa36a701010182ea609435353535353535353535353535353535353535358080c0f867f86583aa36a794000000000000000000000000000000000000000088fffffffffffffffe01a016bf793051be0e6d2b11af7f9f7d290c655f41f258161bdce78f67b26883d684a01c23983457a5e6e01d20a298f7e993b5b45a93e1cfc3915657aef546489408c580a0947bc8d76c122fe524b575a562e738e75462835cf6b1ebb20c405e513c0faacaa073193038ca623154657e244b5365e30c1a14346d4da3fa01ff51b352d130af98"))
	f.Add([]byte{0x04, 0xc0})
	f.Add([]byte{0x04})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var tx SetCodeTx
		if err := tx.DecodeRLP(data); err != nil {
			return
		}
		enc, err := tx.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, data, enc)
	})
}

func FuzzSetCodeSignRoundTrip(f *testing.F) {
	f.Add(uint64(0), []byte{}, []byte{}, uint64(21000), make([]byte, 20), []byte{}, []byte{}, make([]byte, 52), uint64(1), make([]byte, 3*28))
	f.Add(uint64(9), []byte{0x59, 0x68}, []byte{0x06, 0xfc}, uint64(100000), make([]byte, 20), []byte{0x01}, []byte{0xaa}, make([]byte, 52), uint64(0), make([]byte, 28))
	f.Add(uint64(1<<63), make([]byte, 32), make([]byte, 32), uint64(1<<63), make([]byte, 20), make([]byte, 32), make([]byte, 100), make([]byte, 20+32*3), uint64(1<<62), make([]byte, 3*28))

	key, err := types.NewPrivateKeyFromHex(legacyTestKey)
	if err != nil {
		f.Fatal(err)
	}
	authKey, err := types.NewPrivateKeyFromHex(secondTestKey)
	if err != nil {
		f.Fatal(err)
	}
	want := PubkeyToAddress(key.PublicKey())
	wantAuthority := PubkeyToAddress(authKey.PublicKey())

	f.Fuzz(func(t *testing.T, nonce uint64, tip, fee []byte, gasLimit uint64, to, value, data, access []byte, chainID uint64, auths []byte) {
		if len(tip) > 32 || len(fee) > 32 || len(value) > 32 || len(to) != types.AddressLength {
			t.Skip()
		}
		tx := SetCodeTx{
			ChainID:   new(big.Int).SetUint64(chainID),
			Nonce:     nonce,
			GasTipCap: new(big.Int).SetBytes(tip),
			GasFeeCap: new(big.Int).SetBytes(fee),
			GasLimit:  gasLimit,
			To:        types.NewAddressFromBytes(to),
			Value:     new(big.Int).SetBytes(value),
			Data:      data,
		}
		for len(access) >= types.AddressLength {
			entry := AccessTuple{Address: types.NewAddressFromBytes(access[:types.AddressLength])}
			access = access[types.AddressLength:]
			for len(access) >= types.HashLength && len(entry.StorageKeys) < 3 {
				entry.StorageKeys = append(entry.StorageKeys, types.NewHashFromBytes(access[:types.HashLength]))
				access = access[types.HashLength:]
			}
			tx.AccessList = append(tx.AccessList, entry)
		}
		for len(auths) >= 28 && len(tx.AuthList) < 4 {
			auth := Authorization{
				ChainID: new(big.Int).SetBytes(auths[:8]),
				Address: types.NewAddressFromBytes(auths[8:28]),
				Nonce:   nonce ^ chainID,
			}
			auths = auths[28:]
			require.NoError(t, auth.Sign(authKey))
			tx.AuthList = append(tx.AuthList, auth)
		}
		if len(tx.AuthList) == 0 {
			t.Skip()
		}

		require.NoError(t, tx.Sign(key))
		raw, err := tx.EncodeRLP()
		require.NoError(t, err)

		var got SetCodeTx
		require.NoError(t, got.DecodeRLP(raw))
		requireSameSetCode(t, tx, got)
		require.True(t, tx.Signature.Equal(got.Signature))

		again, err := got.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, raw, again)

		digest, err := got.SigningHash()
		require.NoError(t, err)
		ok, err := VerifyAddress(digest, got.Signature, want)
		require.NoError(t, err)
		require.True(t, ok)

		sender, err := got.Sender()
		require.NoError(t, err)
		require.True(t, want.Equal(sender))

		for i := range got.AuthList {
			authority, err := got.AuthList[i].Authority()
			require.NoError(t, err)
			require.True(t, wantAuthority.Equal(authority))
		}
	})
}

func FuzzDecodeTransaction(f *testing.F) {
	f.Add(mustDecodeHexF("f86c098504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a0499aa1110848b179aa0f228e20faa3ba68b350e1feeab49638c6b8ce40ea56aea0053ec9b43dcea26f8d10b43a44bdfafae5b1b26462367921079005d4d274e06d"))
	f.Add(mustDecodeHexF("02f87301098459682f008506fc23ac00825208943535353535353535353535353535353535353535880de0b6b3a764000080c001a0a2a0cb81cd4d8e240f38279012ee02b696002ebd27203a64d4acd6cc60d2f139a029de4d469feeadf9d8bd087fbd31c50c015928b4ac387409758415b8e301fe57"))
	f.Add(mustDecodeHexF("04f8ca01098459682f008506fc23ac00830186a09435353535353535353535353535353535353535358080c0f85cf85a019400000000000000000000000000000000000000aa8080a0042b827292e0c095b5b0e5a9f9253145fc04ee3c68b2236bda2ad646a963b5aea064a02c24a0fd6805523ff5487f38f172ac68865c4ffcd526906a477d7bca6f0180a0335bc722c1669ba60019f08c914067da968205259e8e1697eba5fa390693dac2a00490ef675b056b6920d7a326de3ebf39dc74bccc0e0ddb79e768e076f2fc6052"))
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add([]byte{0x02})
	f.Add([]byte{0x03, 0xc0})
	f.Add([]byte{0x04, 0xc0})
	f.Add([]byte{0xc0})
	f.Add([]byte{0xff})

	f.Fuzz(func(t *testing.T, data []byte) {
		tx, err := DecodeTransaction(data)
		if err != nil {
			require.Nil(t, tx)
			return
		}
		enc, err := tx.EncodeRLP()
		require.NoError(t, err)
		require.Equal(t, data, enc)
	})
}
