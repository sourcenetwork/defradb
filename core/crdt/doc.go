package crdt

// Conflict-Free Replicated Data Types (CRDT)
// are a data structure which can be replicated across multiple computers in a network,
// where the replicas can be updated independently and concurrently without coordination
// between the replicas and are able to determinsitly converge to the same state.

// This package implements a collection of CRDT types specifically to be used in DefraDB,
// and use the Delta-State CRDT architecture to update and replicate state. It is based on
// the go Merkle-CRDT project

// The CRDTs shall satisfy the ReplicatedData interface which is a single merge function
// which given two states of the same data type will merge into a single state.

// Unless the explicitly enabling the entire state to be fully loaded into memory as an object, all data will reside
// inside the BadgerDB datastore.

// In general, each CRDT type will be implemented indepenant, and oblivious to its underlying datastore, and
// to how it will be structured as Merkle-CRDT. Instead they will focus on their core semantics
// and implementation and will be wrapped in handlers to ensure state persistence to DBs,
// DAG creation, and replication to peers.
