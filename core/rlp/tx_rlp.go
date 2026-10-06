package rlp

import (
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

type LegacyUnprotectedRLP struct {
	Nonce    uint64
	GasPrice *big.Int
	GasLimit uint64
	To       *[types.AddressLength]byte `rlp:"nil"`
	Value    *big.Int
	Data     []byte
}

type LegacyProtectedRLP struct {
	Nonce    uint64
	GasPrice *big.Int
	GasLimit uint64
	To       *[types.AddressLength]byte `rlp:"nil"`
	Value    *big.Int
	Data     []byte
	ChainID  *big.Int
	Zero1    uint64
	Zero2    uint64
}

type LegacySignedRLP struct {
	Nonce    uint64
	GasPrice *big.Int
	GasLimit uint64
	To       *[types.AddressLength]byte `rlp:"nil"`
	Value    *big.Int
	Data     []byte
	V        *big.Int
	R        *big.Int
	S        *big.Int
}

type AccessTupleRLP struct {
	Address     [types.AddressLength]byte
	StorageKeys [][types.HashLength]byte
}

type DynamicFeeRLP struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	GasLimit   uint64
	To         *[types.AddressLength]byte `rlp:"nil"`
	Value      *big.Int
	Data       []byte
	AccessList []AccessTupleRLP
}

type DynamicFeeSignedRLP struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	GasLimit   uint64
	To         *[types.AddressLength]byte `rlp:"nil"`
	Value      *big.Int
	Data       []byte
	AccessList []AccessTupleRLP
	V          *big.Int
	R          *big.Int
	S          *big.Int
}

type AuthorizationSigRLP struct {
	ChainID *big.Int
	Address [types.AddressLength]byte
	Nonce   uint64
}

type AuthorizationRLP struct {
	ChainID *big.Int
	Address [types.AddressLength]byte
	Nonce   uint64
	V       uint8
	R       *big.Int
	S       *big.Int
}

type SetCodeRLP struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	GasLimit   uint64
	To         [types.AddressLength]byte
	Value      *big.Int
	Data       []byte
	AccessList []AccessTupleRLP
	AuthList   []AuthorizationRLP
}

type SetCodeSignedRLP struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	GasLimit   uint64
	To         [types.AddressLength]byte
	Value      *big.Int
	Data       []byte
	AccessList []AccessTupleRLP
	AuthList   []AuthorizationRLP
	V          *big.Int
	R          *big.Int
	S          *big.Int
}
