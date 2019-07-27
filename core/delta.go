package core

// Delta represents a delta-state update to delta-CRDT
// They are serialized to and from Protobuf (or CBOR)
type Delta interface {
	GetPriority() uint64
	SetPriotiy(uint64)
	Marshal() ([]byte, error)
}
