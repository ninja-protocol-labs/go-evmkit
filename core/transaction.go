package core

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ninja-protocol-labs/go-evmkit/core/rlp"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

var (
	_ Transaction = (*LegacyTx)(nil)
	_ Transaction = (*DynamicFeeTx)(nil)
	_ Transaction = (*SetCodeTx)(nil)
)

var (
	// ErrUnsigned is returned when a transaction has no Signature.
	ErrUnsigned = errors.New("core: transaction is unsigned")

	// ErrInvalidTxType is returned for an empty encoding or a type other than 0, 2 or 4,
	// or one that does not match the receiver.
	ErrInvalidTxType = errors.New("core: invalid transaction type")

	// ErrInvalidInteger is returned for an integer field that is negative or above 256 bits.
	ErrInvalidInteger = errors.New("core: integer field must be an unsigned 256-bit integer")

	// ErrInvalidTransaction is returned for a structurally invalid transaction.
	ErrInvalidTransaction = errors.New("core: invalid transaction")
)

// authorizationMagic prefixes the EIP-7702 authorization signing payload.
const authorizationMagic byte = 0x05

var (
	bigTwo      = big.NewInt(2)
	eip155VBase = big.NewInt(35)
)

// TxType is the EIP-2718 transaction type byte.
type TxType byte

const (
	// LegacyTxType is the pre-EIP-2718 transaction, protected by EIP-155.
	LegacyTxType TxType = 0x00
	// DynamicFeeTxType is the EIP-1559 transaction.
	DynamicFeeTxType TxType = 0x02
	// SetCodeTxType is the EIP-7702 transaction.
	SetCodeTxType TxType = 0x04
)

// Transaction is implemented by *LegacyTx, *DynamicFeeTx and *SetCodeTx.
type Transaction interface {
	Type() TxType
	// SigningHash returns the digest to sign, ignoring Signature.
	SigningHash() (*types.Hash, error)
	// Sign signs SigningHash and stores the result in Signature.
	Sign(Signer) error
	// Hash returns keccak256 of the signed encoding.
	Hash() (*types.Hash, error)
	// Sender recovers the signer's address.
	Sender() (*types.Address, error)
	// EncodeRLP returns the signed wire encoding, or ErrUnsigned.
	EncodeRLP() ([]byte, error)
	// DecodeRLP parses a signed wire encoding, including Signature.
	DecodeRLP([]byte) error
}

// DecodeTransaction parses a signed transaction, choosing its type from the first byte.
func DecodeTransaction(raw []byte) (Transaction, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: empty encoding", ErrInvalidTxType)
	}

	var tx Transaction
	switch first := raw[0]; {
	case first >= 0xc0:
		tx = new(LegacyTx)
	case first == byte(DynamicFeeTxType):
		tx = new(DynamicFeeTx)
	case first == byte(SetCodeTxType):
		tx = new(SetCodeTx)
	default:
		return nil, fmt.Errorf("%w: first byte 0x%02x", ErrInvalidTxType, first)
	}

	if err := tx.DecodeRLP(raw); err != nil {
		return nil, err
	}
	return tx, nil
}

// AccessTuple is an EIP-2930 access list entry.
type AccessTuple struct {
	Address     *types.Address
	StorageKeys []*types.Hash
}

// AccessList is an EIP-2930 access list.
type AccessList []AccessTuple

// Authorization is an EIP-7702 authorization. A ChainID of 0 is valid on every chain.
type Authorization struct {
	ChainID   *big.Int
	Address   *types.Address
	Nonce     uint64
	Signature *types.Signature
}

func (a *Authorization) validate() error {
	if err := checkUint256(a.ChainID); err != nil {
		return fmt.Errorf("chainId: %w", err)
	}
	if a.Address == nil {
		return fmt.Errorf("%w: authorization has no address", ErrInvalidTransaction)
	}
	return nil
}

// SigningHash returns keccak256(0x05 || rlp([chainId, address, nonce])).
func (a *Authorization) SigningHash() (*types.Hash, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}
	enc, err := rlp.Encode(&rlp.AuthorizationSigRLP{
		ChainID: rlp.BigOrZero(a.ChainID),
		Address: *rlp.FromAddress(a.Address),
		Nonce:   a.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("core: authorization signing hash: %w", err)
	}
	return Keccak256(append([]byte{authorizationMagic}, enc...)), nil
}

func (a *Authorization) Sign(s Signer) error {
	digest, err := a.SigningHash()
	if err != nil {
		return err
	}
	sig, err := s.Sign(digest)
	if err != nil {
		return fmt.Errorf("core: sign authorization: %w", err)
	}
	a.Signature = sig
	return nil
}

// Authority recovers the address that signed the authorization.
func (a *Authorization) Authority() (*types.Address, error) {
	return recoverSigner(a.Signature, a.SigningHash)
}

// LegacyTx is a type 0 transaction. A ChainID of 0 means no EIP-155 replay protection.
type LegacyTx struct {
	ChainID  *big.Int
	Nonce    uint64
	GasPrice *big.Int
	GasLimit uint64
	To       *types.Address
	Value    *big.Int
	Data     []byte

	Signature *types.Signature // nil until signed
}

func (*LegacyTx) Type() TxType {
	return LegacyTxType
}

func (tx *LegacyTx) validate() error {
	if err := checkUint256(tx.ChainID); err != nil {
		return fmt.Errorf("chainId: %w", err)
	}
	if err := checkUint256(tx.GasPrice); err != nil {
		return fmt.Errorf("gasPrice: %w", err)
	}
	if err := checkUint256(tx.Value); err != nil {
		return fmt.Errorf("value: %w", err)
	}
	return nil
}

// SigningHash returns keccak256(rlp([nonce, gasPrice, gasLimit, to, value, data, chainId, 0, 0])),
// without the last three fields when ChainID is 0.
func (tx *LegacyTx) SigningHash() (*types.Hash, error) {
	if err := tx.validate(); err != nil {
		return nil, err
	}
	chainID := rlp.BigOrZero(tx.ChainID)

	var payload any
	if chainID.Sign() == 0 {
		payload = &rlp.LegacyUnprotectedRLP{
			Nonce:    tx.Nonce,
			GasPrice: rlp.BigOrZero(tx.GasPrice),
			GasLimit: tx.GasLimit,
			To:       rlp.FromAddress(tx.To),
			Value:    rlp.BigOrZero(tx.Value),
			Data:     tx.Data,
		}
	} else {
		payload = &rlp.LegacyProtectedRLP{
			Nonce:    tx.Nonce,
			GasPrice: rlp.BigOrZero(tx.GasPrice),
			GasLimit: tx.GasLimit,
			To:       rlp.FromAddress(tx.To),
			Value:    rlp.BigOrZero(tx.Value),
			Data:     tx.Data,
			ChainID:  chainID,
		}
	}

	enc, err := rlp.Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("core: legacy tx signing hash: %w", err)
	}
	return Keccak256(enc), nil
}

func (tx *LegacyTx) Sign(s Signer) error {
	digest, err := tx.SigningHash()
	if err != nil {
		return err
	}
	sig, err := s.Sign(digest)
	if err != nil {
		return fmt.Errorf("core: sign legacy tx: %w", err)
	}
	tx.Signature = sig
	return nil
}

// EncodeRLP returns rlp([nonce, gasPrice, gasLimit, to, value, data, v, r, s]),
// with v = chainId*2 + 35 + recoveryId, or 27 + recoveryId when ChainID is 0.
func (tx *LegacyTx) EncodeRLP() ([]byte, error) {
	if tx.Signature.IsZero() {
		return nil, ErrUnsigned
	}
	if err := tx.validate(); err != nil {
		return nil, err
	}
	if _, err := rlp.RecoveryID(tx.Signature); err != nil {
		return nil, err
	}

	chainID := rlp.BigOrZero(tx.ChainID)
	v := tx.Signature.LegacyV()
	if chainID.Sign() != 0 {
		v = tx.Signature.EIP155V(chainID)
	}

	enc, err := rlp.Encode(rlp.LegacySignedRLP{
		Nonce:    tx.Nonce,
		GasPrice: rlp.BigOrZero(tx.GasPrice),
		GasLimit: tx.GasLimit,
		To:       rlp.FromAddress(tx.To),
		Value:    rlp.BigOrZero(tx.Value),
		Data:     tx.Data,
		V:        v,
		R:        tx.Signature.R(),
		S:        tx.Signature.S(),
	})
	if err != nil {
		return nil, fmt.Errorf("core: encode legacy tx: %w", err)
	}
	return enc, nil
}

func (tx *LegacyTx) Hash() (*types.Hash, error) {
	enc, err := tx.EncodeRLP()
	if err != nil {
		return nil, err
	}
	return Keccak256(enc), nil
}

func (tx *LegacyTx) Sender() (*types.Address, error) {
	return recoverSigner(tx.Signature, tx.SigningHash)
}

// DecodeRLP parses a signed legacy transaction. v of 27 or 28 means ChainID 0;
// otherwise ChainID is (v - 35) / 2.
func (tx *LegacyTx) DecodeRLP(data []byte) error {
	var dec rlp.LegacySignedRLP
	if err := rlp.Decode(data, &dec); err != nil {
		return fmt.Errorf("core: decode legacy tx: %w", err)
	}

	chainID := new(big.Int)
	var recID byte
	switch {
	case dec.V.IsUint64() && (dec.V.Uint64() == 27 || dec.V.Uint64() == 28):
		recID = byte(dec.V.Uint64() - 27)
	case dec.V.Cmp(eip155VBase) >= 0:
		rest := new(big.Int).Sub(dec.V, eip155VBase)
		var odd *big.Int
		chainID, odd = new(big.Int).DivMod(rest, bigTwo, new(big.Int))
		recID = byte(odd.Uint64())
		if chainID.Sign() == 0 {
			return fmt.Errorf("core: decode legacy tx: invalid v %s", dec.V)
		}
	default:
		return fmt.Errorf("core: decode legacy tx: invalid v %s", dec.V)
	}

	sig, err := rlp.SignatureFromRS(dec.R, dec.S, recID)
	if err != nil {
		return fmt.Errorf("core: decode legacy tx: %w", err)
	}

	decoded := LegacyTx{
		ChainID:   chainID,
		Nonce:     dec.Nonce,
		GasPrice:  dec.GasPrice,
		GasLimit:  dec.GasLimit,
		To:        rlp.ToAddress(dec.To),
		Value:     dec.Value,
		Data:      dec.Data,
		Signature: sig,
	}
	if err := decoded.validate(); err != nil {
		return fmt.Errorf("core: decode legacy tx: %w", err)
	}
	*tx = decoded
	return nil
}

// DynamicFeeTx is a type 2 (EIP-1559) transaction.
type DynamicFeeTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int // maxPriorityFeePerGas
	GasFeeCap  *big.Int // maxFeePerGas
	GasLimit   uint64
	To         *types.Address
	Value      *big.Int
	Data       []byte
	AccessList AccessList

	Signature *types.Signature // nil until signed
}

func (*DynamicFeeTx) Type() TxType {
	return DynamicFeeTxType
}

func (tx *DynamicFeeTx) validate() error {
	if err := checkUint256(tx.ChainID); err != nil {
		return fmt.Errorf("chainId: %w", err)
	}
	if err := checkUint256(tx.GasTipCap); err != nil {
		return fmt.Errorf("gasTipCap: %w", err)
	}
	if err := checkUint256(tx.GasFeeCap); err != nil {
		return fmt.Errorf("gasFeeCap: %w", err)
	}
	if err := checkUint256(tx.Value); err != nil {
		return fmt.Errorf("value: %w", err)
	}
	return nil
}

func (tx *DynamicFeeTx) unsignedRLP() (*rlp.DynamicFeeRLP, error) {
	if err := tx.validate(); err != nil {
		return nil, err
	}
	accessList, err := accessListToRLP(tx.AccessList)
	if err != nil {
		return nil, err
	}
	return &rlp.DynamicFeeRLP{
		ChainID:    rlp.BigOrZero(tx.ChainID),
		Nonce:      tx.Nonce,
		GasTipCap:  rlp.BigOrZero(tx.GasTipCap),
		GasFeeCap:  rlp.BigOrZero(tx.GasFeeCap),
		GasLimit:   tx.GasLimit,
		To:         rlp.FromAddress(tx.To),
		Value:      rlp.BigOrZero(tx.Value),
		Data:       tx.Data,
		AccessList: accessList,
	}, nil
}

// SigningHash returns keccak256(0x02 || rlp([chainId, nonce, gasTipCap, gasFeeCap,
// gasLimit, to, value, data, accessList])).
func (tx *DynamicFeeTx) SigningHash() (*types.Hash, error) {
	payload, err := tx.unsignedRLP()
	if err != nil {
		return nil, err
	}
	enc, err := rlp.Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("core: dynamic fee tx signing hash: %w", err)
	}
	return Keccak256(append([]byte{byte(DynamicFeeTxType)}, enc...)), nil
}

func (tx *DynamicFeeTx) Sign(s Signer) error {
	digest, err := tx.SigningHash()
	if err != nil {
		return err
	}
	sig, err := s.Sign(digest)
	if err != nil {
		return fmt.Errorf("core: sign dynamic fee tx: %w", err)
	}
	tx.Signature = sig
	return nil
}

// EncodeRLP returns 0x02 || rlp([chainId, nonce, gasTipCap, gasFeeCap,
// gasLimit, to, value, data, accessList, yParity, r, s]).
func (tx *DynamicFeeTx) EncodeRLP() ([]byte, error) {
	if tx.Signature.IsZero() {
		return nil, ErrUnsigned
	}
	yParity, err := rlp.RecoveryID(tx.Signature)
	if err != nil {
		return nil, err
	}
	unsigned, err := tx.unsignedRLP()
	if err != nil {
		return nil, err
	}

	enc, err := rlp.Encode(&rlp.DynamicFeeSignedRLP{
		ChainID:    unsigned.ChainID,
		Nonce:      unsigned.Nonce,
		GasTipCap:  unsigned.GasTipCap,
		GasFeeCap:  unsigned.GasFeeCap,
		GasLimit:   unsigned.GasLimit,
		To:         unsigned.To,
		Value:      unsigned.Value,
		Data:       unsigned.Data,
		AccessList: unsigned.AccessList,
		V:          new(big.Int).SetUint64(uint64(yParity)),
		R:          tx.Signature.R(),
		S:          tx.Signature.S(),
	})
	if err != nil {
		return nil, fmt.Errorf("core: encode dynamic fee tx: %w", err)
	}
	return append([]byte{byte(DynamicFeeTxType)}, enc...), nil
}

func (tx *DynamicFeeTx) Hash() (*types.Hash, error) {
	enc, err := tx.EncodeRLP()
	if err != nil {
		return nil, err
	}
	return Keccak256(enc), nil
}

func (tx *DynamicFeeTx) Sender() (*types.Address, error) {
	return recoverSigner(tx.Signature, tx.SigningHash)
}

// DecodeRLP parses a signed type 2 transaction.
func (tx *DynamicFeeTx) DecodeRLP(data []byte) error {
	if len(data) == 0 || data[0] != byte(DynamicFeeTxType) {
		return fmt.Errorf("core: decode dynamic fee tx: %w", ErrInvalidTxType)
	}

	var dec rlp.DynamicFeeSignedRLP
	if err := rlp.Decode(data[1:], &dec); err != nil {
		return fmt.Errorf("core: decode dynamic fee tx: %w", err)
	}
	if !dec.V.IsUint64() || dec.V.Uint64() > 1 {
		return fmt.Errorf("core: decode dynamic fee tx: invalid yParity %s", dec.V)
	}

	sig, err := rlp.SignatureFromRS(dec.R, dec.S, byte(dec.V.Uint64()))
	if err != nil {
		return fmt.Errorf("core: decode dynamic fee tx: %w", err)
	}

	decoded := DynamicFeeTx{
		ChainID:    dec.ChainID,
		Nonce:      dec.Nonce,
		GasTipCap:  dec.GasTipCap,
		GasFeeCap:  dec.GasFeeCap,
		GasLimit:   dec.GasLimit,
		To:         rlp.ToAddress(dec.To),
		Value:      dec.Value,
		Data:       dec.Data,
		AccessList: accessListFromRLP(dec.AccessList),
		Signature:  sig,
	}
	if err := decoded.validate(); err != nil {
		return fmt.Errorf("core: decode dynamic fee tx: %w", err)
	}
	*tx = decoded
	return nil
}

// SetCodeTx is a type 4 (EIP-7702) transaction. To and AuthList are required.
type SetCodeTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int // maxPriorityFeePerGas
	GasFeeCap  *big.Int // maxFeePerGas
	GasLimit   uint64
	To         *types.Address
	Value      *big.Int
	Data       []byte
	AccessList AccessList
	AuthList   []Authorization

	Signature *types.Signature // nil until signed
}

func (*SetCodeTx) Type() TxType {
	return SetCodeTxType
}

func (tx *SetCodeTx) validate() error {
	if err := checkUint256(tx.ChainID); err != nil {
		return fmt.Errorf("chainId: %w", err)
	}
	if err := checkUint256(tx.GasTipCap); err != nil {
		return fmt.Errorf("gasTipCap: %w", err)
	}
	if err := checkUint256(tx.GasFeeCap); err != nil {
		return fmt.Errorf("gasFeeCap: %w", err)
	}
	if err := checkUint256(tx.Value); err != nil {
		return fmt.Errorf("value: %w", err)
	}
	if tx.To == nil {
		return fmt.Errorf("%w: set code transaction requires a destination", ErrInvalidTransaction)
	}
	if len(tx.AuthList) == 0 {
		return fmt.Errorf("%w: set code transaction requires an authorization", ErrInvalidTransaction)
	}
	for i := range tx.AuthList {
		auth := &tx.AuthList[i]
		if err := auth.validate(); err != nil {
			return fmt.Errorf("authorization %d: %w", i, err)
		}
		if auth.Signature.IsZero() {
			return fmt.Errorf("authorization %d: %w", i, ErrUnsigned)
		}
		if _, err := rlp.RecoveryID(auth.Signature); err != nil {
			return fmt.Errorf("authorization %d: %w", i, err)
		}
	}
	return nil
}

func (tx *SetCodeTx) unsignedRLP() (*rlp.SetCodeRLP, error) {
	if err := tx.validate(); err != nil {
		return nil, err
	}
	accessList, err := accessListToRLP(tx.AccessList)
	if err != nil {
		return nil, err
	}

	authList := make([]rlp.AuthorizationRLP, len(tx.AuthList))
	for i, auth := range tx.AuthList {
		yParity, _ := rlp.RecoveryID(auth.Signature)
		authList[i] = rlp.AuthorizationRLP{
			ChainID: rlp.BigOrZero(auth.ChainID),
			Address: *rlp.FromAddress(auth.Address),
			Nonce:   auth.Nonce,
			V:       yParity,
			R:       auth.Signature.R(),
			S:       auth.Signature.S(),
		}
	}

	return &rlp.SetCodeRLP{
		ChainID:    rlp.BigOrZero(tx.ChainID),
		Nonce:      tx.Nonce,
		GasTipCap:  rlp.BigOrZero(tx.GasTipCap),
		GasFeeCap:  rlp.BigOrZero(tx.GasFeeCap),
		GasLimit:   tx.GasLimit,
		To:         *rlp.FromAddress(tx.To),
		Value:      rlp.BigOrZero(tx.Value),
		Data:       tx.Data,
		AccessList: accessList,
		AuthList:   authList,
	}, nil
}

// SigningHash returns keccak256(0x04 || rlp([chainId, nonce, gasTipCap, gasFeeCap,
// gasLimit, to, value, data, accessList, authorizationList])). Every
// authorization must already be signed.
func (tx *SetCodeTx) SigningHash() (*types.Hash, error) {
	payload, err := tx.unsignedRLP()
	if err != nil {
		return nil, err
	}
	enc, err := rlp.Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("core: set code tx signing hash: %w", err)
	}
	return Keccak256(append([]byte{byte(SetCodeTxType)}, enc...)), nil
}

func (tx *SetCodeTx) Sign(s Signer) error {
	digest, err := tx.SigningHash()
	if err != nil {
		return err
	}
	sig, err := s.Sign(digest)
	if err != nil {
		return fmt.Errorf("core: sign set code tx: %w", err)
	}
	tx.Signature = sig
	return nil
}

// EncodeRLP returns 0x04 || rlp([chainId, nonce, gasTipCap, gasFeeCap, gasLimit,
// to, value, data, accessList, authorizationList, yParity, r, s]).
func (tx *SetCodeTx) EncodeRLP() ([]byte, error) {
	if tx.Signature.IsZero() {
		return nil, ErrUnsigned
	}
	yParity, err := rlp.RecoveryID(tx.Signature)
	if err != nil {
		return nil, err
	}
	unsigned, err := tx.unsignedRLP()
	if err != nil {
		return nil, err
	}

	enc, err := rlp.Encode(&rlp.SetCodeSignedRLP{
		ChainID:    unsigned.ChainID,
		Nonce:      unsigned.Nonce,
		GasTipCap:  unsigned.GasTipCap,
		GasFeeCap:  unsigned.GasFeeCap,
		GasLimit:   unsigned.GasLimit,
		To:         unsigned.To,
		Value:      unsigned.Value,
		Data:       unsigned.Data,
		AccessList: unsigned.AccessList,
		AuthList:   unsigned.AuthList,
		V:          new(big.Int).SetUint64(uint64(yParity)),
		R:          tx.Signature.R(),
		S:          tx.Signature.S(),
	})
	if err != nil {
		return nil, fmt.Errorf("core: encode set code tx: %w", err)
	}
	return append([]byte{byte(SetCodeTxType)}, enc...), nil
}

func (tx *SetCodeTx) Hash() (*types.Hash, error) {
	enc, err := tx.EncodeRLP()
	if err != nil {
		return nil, err
	}
	return Keccak256(enc), nil
}

func (tx *SetCodeTx) Sender() (*types.Address, error) {
	return recoverSigner(tx.Signature, tx.SigningHash)
}

// DecodeRLP parses a signed type 4 transaction.
func (tx *SetCodeTx) DecodeRLP(data []byte) error {
	if len(data) == 0 || data[0] != byte(SetCodeTxType) {
		return fmt.Errorf("core: decode set code tx: %w", ErrInvalidTxType)
	}

	var dec rlp.SetCodeSignedRLP
	if err := rlp.Decode(data[1:], &dec); err != nil {
		return fmt.Errorf("core: decode set code tx: %w", err)
	}
	if !dec.V.IsUint64() || dec.V.Uint64() > 1 {
		return fmt.Errorf("core: decode set code tx: invalid yParity %s", dec.V)
	}

	sig, err := rlp.SignatureFromRS(dec.R, dec.S, byte(dec.V.Uint64()))
	if err != nil {
		return fmt.Errorf("core: decode set code tx: %w", err)
	}

	authList := make([]Authorization, len(dec.AuthList))
	for i, auth := range dec.AuthList {
		authSig, err := rlp.SignatureFromRS(auth.R, auth.S, auth.V)
		if err != nil {
			return fmt.Errorf("core: decode set code tx: authorization %d: %w", i, err)
		}
		authList[i] = Authorization{
			ChainID:   auth.ChainID,
			Address:   types.NewAddressFromBytes(auth.Address[:]),
			Nonce:     auth.Nonce,
			Signature: authSig,
		}
	}

	decoded := SetCodeTx{
		ChainID:    dec.ChainID,
		Nonce:      dec.Nonce,
		GasTipCap:  dec.GasTipCap,
		GasFeeCap:  dec.GasFeeCap,
		GasLimit:   dec.GasLimit,
		To:         types.NewAddressFromBytes(dec.To[:]),
		Value:      dec.Value,
		Data:       dec.Data,
		AccessList: accessListFromRLP(dec.AccessList),
		AuthList:   authList,
		Signature:  sig,
	}
	if err := decoded.validate(); err != nil {
		return fmt.Errorf("core: decode set code tx: %w", err)
	}
	*tx = decoded
	return nil
}

// checkUint256 reports whether n, which may be nil, is an unsigned 256-bit integer.
func checkUint256(n *big.Int) error {
	if n == nil {
		return nil
	}
	if n.Sign() < 0 {
		return ErrInvalidInteger
	}
	if n.BitLen() > 256 {
		return ErrInvalidInteger
	}
	return nil
}

func accessListToRLP(al AccessList) ([]rlp.AccessTupleRLP, error) {
	out := make([]rlp.AccessTupleRLP, len(al))
	for i, entry := range al {
		if entry.Address == nil {
			return nil, fmt.Errorf("%w: access list entry %d has no address", ErrInvalidTransaction, i)
		}
		out[i].Address = *rlp.FromAddress(entry.Address)

		out[i].StorageKeys = make([][types.HashLength]byte, len(entry.StorageKeys))
		for j, key := range entry.StorageKeys {
			if key == nil {
				return nil, fmt.Errorf("%w: access list entry %d has a nil storage key", ErrInvalidTransaction, i)
			}
			copy(out[i].StorageKeys[j][:], key.Bytes())
		}
	}
	return out, nil
}

func accessListFromRLP(in []rlp.AccessTupleRLP) AccessList {
	if len(in) == 0 {
		return nil
	}
	out := make(AccessList, len(in))
	for i, entry := range in {
		out[i].Address = types.NewAddressFromBytes(entry.Address[:])
		out[i].StorageKeys = make([]*types.Hash, len(entry.StorageKeys))
		for j, key := range entry.StorageKeys {
			out[i].StorageKeys[j] = types.NewHashFromBytes(key[:])
		}
	}
	return out
}

func recoverSigner(sig *types.Signature, signingHash func() (*types.Hash, error)) (*types.Address, error) {
	if sig.IsZero() {
		return nil, ErrUnsigned
	}
	if _, err := rlp.RecoveryID(sig); err != nil {
		return nil, err
	}
	digest, err := signingHash()
	if err != nil {
		return nil, err
	}
	pub, err := sig.ECRecover(digest)
	if err != nil {
		return nil, fmt.Errorf("core: recover signer: %w", err)
	}
	return PubkeyToAddress(pub), nil
}
