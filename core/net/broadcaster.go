package net

// A Broadcaster provides a way to send (notify) an opaque payload to
// all replicas and to retrieve payloads broadcasted.
type Broadcaster interface {
	// Send payload to other replicas.
	Broadcast(topic []byte, buf []byte) error
	// Obtain the next payload received from the network.
	Next(topic []byte) ([]byte, error)
}
