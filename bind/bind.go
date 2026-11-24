// Package bind is the path every contract client shares: reads go through Call, writes through Transact, and multi-call previews through Simulate.
package bind

import "errors"

var (
	ErrInvalidOptions = errors.New("invalid options")
	ErrInvalidRequest = errors.New("invalid request")
)
