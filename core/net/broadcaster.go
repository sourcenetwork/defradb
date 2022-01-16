package net

import "time"

// A Broadcaster provides a way to send (notify) an opaque payload to
// all replicas and to retrieve payloads broadcasted.
type Broadcaster interface {
	// Send broadcasts a message without blocking
	Send(v interface{}) error

	// SendWithTimeout broadcasts a message, blocks upto timeout duration
	SendWithTimeout(v interface{}, d time.Duration) error

	// // Obtain the next payload received from the network.
	// Listen() Listener
}

// // Listener allows clients to subscribe to a Broadcaster channel
// type Listener interface {
// 	// Discard closes the listener channel
// 	Discard()

// 	// Channel returns the underlying channel
// 	Channel() <-chan interface{}
// }
