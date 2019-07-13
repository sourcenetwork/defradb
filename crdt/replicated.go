package crdt

// ReplicatedData is a data type that allows concurrent writers
// to deterministicly merge other replicated data so as to
// converge on the same state
type ReplicatedData interface {
	Merge(data ReplicatedData)
}
