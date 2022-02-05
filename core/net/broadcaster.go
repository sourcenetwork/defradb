package net

import "time"

// A Broadcaster provides a way to send (notify) an opaque payload to
// all replicas and to retrieve payloads broadcasted.
type Broadcaster interface {
	// Send broadcasts a message without blocking
	Send(v interface{}) error

	// SendWithTimeout broadcasts a message, blocks upto timeout duration
	SendWithTimeout(v interface{}, d time.Duration) error
}
