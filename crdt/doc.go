package crdt

// Conflict-Free Replicated Data Types (CRDT)
// are a data structure which can be replicated across multiple computers in a network,
// where the replicas can be updated independently and concurrently without coordination
// between the replicas and are able to determinsitly converge to the same state.

// This package implements a collection of CRDT types specifically to be used in DefraDB,
// and use the Delta-State CRDT architecture to update and replicate state.

// The CRDTs shall satisfy the ReplicatedData interface which is a single merge function
// which given two states of the same data type will merge into a single state.
