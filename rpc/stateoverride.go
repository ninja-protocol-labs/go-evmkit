package rpc

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"slices"
	"strconv"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

var ErrInvalidStateOverride = errors.New("invalid state override")

// StateOverride is the per-account state set passed as the third argument of eth_call and eth_estimateGas.
type StateOverride struct {
	accounts   map[string]*accountOverride
	nilAccount bool
}

// accountOverride holds one account's overrides; a nil field leaves that part of the account unchanged.
type accountOverride struct {
	balance   *big.Int
	nonce     *uint64
	code      []byte
	state     map[[32]byte][32]byte
	stateDiff map[[32]byte][32]byte
}

// NewStateOverride returns an empty StateOverride.
func NewStateOverride() *StateOverride {
	return &StateOverride{
		accounts: make(map[string]*accountOverride),
	}
}

// WithBalance sets account's balance; nil removes a previously set balance.
func (o *StateOverride) WithBalance(account *types.Address, balance *big.Int) *StateOverride {
	a := o.account(account)
	a.balance = nil
	if balance != nil {
		a.balance = new(big.Int).Set(balance)
	}
	return o
}

// WithNonce sets account's nonce.
func (o *StateOverride) WithNonce(account *types.Address, nonce uint64) *StateOverride {
	o.account(account).nonce = &nonce
	return o
}

// WithCode sets account's code; nil removes a previously set code and an empty slice clears the code.
func (o *StateOverride) WithCode(account *types.Address, code []byte) *StateOverride {
	a := o.account(account)
	a.code = nil
	if code != nil {
		a.code = append([]byte{}, code...)
	}
	return o
}

// WithState adds slot to account's full storage replacement, zeroing every slot not set.
func (o *StateOverride) WithState(account *types.Address, slot, value [32]byte) *StateOverride {
	a := o.account(account)
	if a.state == nil {
		a.state = make(map[[32]byte][32]byte)
	}
	a.state[slot] = value
	return o
}

// WithStateDiff overwrites slot in account's storage, leaving other slots unchanged.
func (o *StateOverride) WithStateDiff(account *types.Address, slot, value [32]byte) *StateOverride {
	a := o.account(account)
	if a.stateDiff == nil {
		a.stateDiff = make(map[[32]byte][32]byte)
	}
	a.stateDiff[slot] = value
	return o
}

// Validate reports a nil account, a balance outside uint256, or state and stateDiff on the same account.
func (o *StateOverride) Validate() error {
	if o == nil {
		return nil
	}
	if o.nilAccount {
		return fmt.Errorf("%w: account is nil", ErrInvalidStateOverride)
	}
	for _, key := range slices.Sorted(maps.Keys(o.accounts)) {
		a := o.accounts[key]
		if a.balance != nil && a.balance.Sign() < 0 {
			return fmt.Errorf("%w: %s.balance is negative", ErrInvalidStateOverride, key)
		}
		if a.balance != nil && a.balance.BitLen() > 256 {
			return fmt.Errorf("%w: %s.balance exceeds uint256", ErrInvalidStateOverride, key)
		}
		if a.state != nil && a.stateDiff != nil {
			return fmt.Errorf("%w: %s: state and stateDiff are both set", ErrInvalidStateOverride, key)
		}
	}
	return nil
}

// MarshalJSON validates o and encodes it as the deterministic address → account object nodes expect.
func (o *StateOverride) MarshalJSON() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	out := make(map[string]map[string]any, len(o.accounts))
	for key, a := range o.accounts {
		fields := make(map[string]any)
		if a.balance != nil {
			fields["balance"] = "0x" + a.balance.Text(16)
		}
		if a.nonce != nil {
			fields["nonce"] = "0x" + strconv.FormatUint(*a.nonce, 16)
		}
		if a.code != nil {
			fields["code"] = encoding.Hex.EncodePrefixed(a.code)
		}
		if a.state != nil {
			fields["state"] = encodeStorage(a.state)
		}
		if a.stateDiff != nil {
			fields["stateDiff"] = encodeStorage(a.stateDiff)
		}
		out[key] = fields
	}
	return json.Marshal(out, json.Deterministic(true))
}

func (o *StateOverride) account(account *types.Address) *accountOverride {
	if account == nil {
		o.nilAccount = true
		return new(accountOverride)
	}
	if o.accounts == nil {
		o.accounts = make(map[string]*accountOverride)
	}
	key := encoding.Hex.EncodePrefixed(account.Bytes())
	a, ok := o.accounts[key]
	if !ok {
		a = new(accountOverride)
		o.accounts[key] = a
	}
	return a
}

func encodeStorage(storage map[[32]byte][32]byte) map[string]string {
	out := make(map[string]string, len(storage))
	for slot, value := range storage {
		out[encoding.Hex.EncodePrefixed(slot[:])] = encoding.Hex.EncodePrefixed(value[:])
	}
	return out
}
