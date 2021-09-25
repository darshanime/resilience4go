package bulkhead

import "errors"

var (
	// ErrFull is returned if bulkhead full even after max wait duration
	ErrFull = errors.New("bulkhead full error")
)
