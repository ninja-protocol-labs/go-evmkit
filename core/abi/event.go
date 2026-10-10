package abi

import (
	"fmt"
	"strings"

	"github.com/ninja-protocol-labs/go-lib-cryptography/keccak"
)

// Topic is a 32-byte log topic slot: topics[0] (the event signature hash,
// unless anonymous) or one indexed parameter's value/hash.
type Topic [32]byte

// EventParam is one parameter of an event declaration: a type plus whether
// it's indexed (stored as a topic) or not (stored in the log's data).
type EventParam struct {
	Type    Type
	Indexed bool
}

// Event describes a Solidity event declaration (e.g. "event Transfer(address
// indexed from, address indexed to, uint256 value)"): which parameters are
// indexed (and so live in topics) vs. not (and so live in the log's data).
type Event struct {
	Name      string
	Inputs    []EventParam
	Anonymous bool

	// set once by NewEvent and never written afterwards, so concurrent
	// reads are safe. Nil when Anonymous, since anonymous events have no
	// signature topic.
	t0 *Topic
}

// NewEvent builds an Event from its name, parameters, and whether it's
// anonymous (anonymous events don't reserve topics[0] for the signature
// hash, leaving one more topic slot for indexed parameters).
func NewEvent(name string, inputs []EventParam, anonymous bool) *Event {
	e := &Event{
		Name:      name,
		Inputs:    inputs,
		Anonymous: anonymous,
	}
	if !anonymous {
		t0 := e.computeTopic0()
		e.t0 = &t0
	}
	return e
}

// Signature returns the event's canonical signature, e.g.
// "Transfer(address,address,uint256)" — indexed-ness doesn't affect it.
func (e *Event) Signature() string {
	names := make([]string, len(e.Inputs))
	for i := range e.Inputs {
		names[i] = e.Inputs[i].Type.String()
	}
	return e.Name + "(" + strings.Join(names, ",") + ")"
}

// Topic0 returns topics[0]: keccak256(Signature()). Anonymous events have none
// ok reports whether one exists.
func (e *Event) Topic0() (topic Topic, ok bool) {
	if e.t0 == nil {
		return Topic{}, false
	}
	return *e.t0, true
}

// Matches reports whether topic0 is e's Topic0 (e.g. a log's first topic); always false for an Anonymous event.
func (e *Event) Matches(topic0 Topic) bool {
	t0, ok := e.Topic0()
	return ok && topic0 == t0
}

// Decode decodes a log's topics and data according to Inputs, in
// declaration order. An indexed value-type parameter (bool/uintN/intN/
// address/bytesN) decodes to its Go value same as a non-indexed one; an
// indexed dynamic-type parameter (string/bytes/array/slice/tuple) can't be
// recovered from its topic — keccak256 is one-way — so it decodes to the
// raw Topic hash instead. Non-anonymous events expect topics[0] to be
// Topic0().
func (e *Event) Decode(topics []Topic, data []byte) ([]any, error) {
	var tIdx int
	if !e.Anonymous {
		if len(topics) == 0 {
			return nil, fmt.Errorf("%w: expected topics[0], got none", ErrByteLengthMismatch)
		}
		if t, _ := e.Topic0(); topics[0] != t {
			return nil, fmt.Errorf("%w: expected %x, got %x", ErrTopicMismatch, t, topics[0])
		}
		tIdx = 1
	}

	v := make([]any, len(e.Inputs))
	var (
		dTyp Types
		dIdx []int
	)
	for i, p := range e.Inputs {
		if !p.Indexed {
			dTyp = append(dTyp, p.Type)
			dIdx = append(dIdx, i)
			continue
		}
		if tIdx >= len(topics) {
			return nil, fmt.Errorf("%w: missing topic for indexed parameter %d", ErrByteLengthMismatch, i)
		}
		if p.Type.IsDynamic() {
			v[i] = topics[tIdx]
		} else {
			dv, err := decodeValue(p.Type, topics[tIdx][:])
			if err != nil {
				return nil, fmt.Errorf("abi: decode event %s: %w", e.Name, err)
			}
			v[i] = dv
		}
		tIdx++
	}
	if tIdx != len(topics) {
		return nil, fmt.Errorf("%w: expected %d topics, got %d", ErrByteLengthMismatch, tIdx, len(topics))
	}

	dataVals, err := Unpack(dTyp, data)
	if err != nil {
		return nil, fmt.Errorf("abi: decode event %s: %w", e.Name, err)
	}
	for i, idx := range dIdx {
		v[idx] = dataVals[i]
	}

	return v, nil
}

func (e *Event) computeTopic0() Topic {
	digest := keccak.Hash256([]byte(e.Signature())).Bytes()
	var t Topic
	copy(t[:], digest[:])
	return t
}
