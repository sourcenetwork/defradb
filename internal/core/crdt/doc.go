// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package crdt implements a collection of CRDT types specifically to be used in DefraDB, and use the Delta-State CRDT
architecture to update and replicate state. It is based on the go Merkle-CRDT project.

Conflict-Free Replicated Data Types (CRDT) are a data structure which can be replicated across multiple computers in a
network, where the replicas can be updated independently and concurrently without coordination between the replicas and
are able to deterministically converge to the same state.

The CRDTs shall satisfy the ReplicatedData interface which is a single merge function which given two states of the
same data type will merge into a single state.

Unless explicitly enabling the entire state to be fully loaded into memory as an object, all data will reside inside
the BadgerDB datastore.

In general, each CRDT type will be implemented independently, and oblivious to its underlying datastore, and to how it
will be structured as Merkle-CRDT. Instead they will focus on their core semantics and implementation and will be
wrapped in handlers to ensure state persistence to DBs, DAG creation, and replication to peers.
*/
package crdt
