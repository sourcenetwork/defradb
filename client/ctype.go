// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// CType indicates CRDT type.
type CType byte

// Available CRDT types.
//
// If a CRDT type is not declared here, we do not support it!
// CRDT types here may not be valid in all contexts.
const (
	NONE_CRDT = CType(iota) // reserved none type
	LWW_REGISTER
	OBJECT
	COMPOSITE
)
