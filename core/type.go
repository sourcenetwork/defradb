// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package core

// CType indicates CRDT type
// @todo: Migrate core/crdt.Type and merkle/crdt.Type to unifiied /core.CRDTType
type CType byte

const (
	//no lint
	NONE_CRDT = CType(iota) // reserved none type
	LWW_REGISTER
	OBJECT
	COMPOSITE
)

var (
	ByteToType = map[byte]CType{
		byte(0): NONE_CRDT,
		byte(1): LWW_REGISTER,
		byte(2): OBJECT,
		byte(3): COMPOSITE,
	}
)

// reserved names
const (
	HEAD         = "_head"
	COMPOSITE_ID = "C"
)
