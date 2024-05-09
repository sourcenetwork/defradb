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
Package merkle provides implementions of the Merkle data structures.
*/
package merkle

/*
CRDTs are composed of two structures, the payload and a clock. The
payload is the actual CRDT data which abides by the merge semantics.
The clock is a mechanism to provide a casual ordering of events, so
we can determine which event proceeded each other and apply the
various merge strategies.

MerkleCRDTs are similar, they contain a CRDT payload, but instead
of a logical or vector clock, it uses a MerkleClock.

MerkleClock is a Merkle DAG, which provides causal ordering of nodes
which link to previous nodes, with content addressable IDs creating a
fully linked graph of content. The linked graph of nodes creates a
natural history of events because a parent node contains a CID of a
child node, which ensures parents occurred AFTER a child.

	  A			  	   B			  C
	//////   link	//////	 link	//////
	//--//--------->//--//--------->//	//
	//////			//////			//////
     head							 tail

	The above diagram shows the ordering of events A, B, and C.

API:
	mc = NewMerkleClock(blockstore)
	event = mc.NewEvent(delta)
	mc.AddEvent(event) cid
	mc.HasEvent(cid)
	mc.
	extractDelta(node) delta
*/
