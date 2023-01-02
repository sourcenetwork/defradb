// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import "fmt"

// todo - confirm use of int - there may be some benefit to using string (direct use of json patch ops)
// which might be more understandable to users.
type PatchOp int

const (
	// IMO these names and the names of the types that implement `Patch` should very strictly stick to
	// the names used in the JSON patch spec: https://jsonpatch.com/ - should save users a fair amount
	// of head-scratching, and will provide natural consistency with any document level json-patch
	// support that we might add later (e.g. for inline arrays).
	PatchOpAdd    = 1
	PatchOpMove   = 2
	PatchOpRemove = 3
)

// interface helps minimize the failing surface area when compared to a struct
// for example providing a value to a remove op, or failing to provide a dst path
// to a move op.  See client/db.go for usage.
//
// Types implementing this interface might be converted to a single concrete type
// internally (unimportant to decide that now IMO), similar to how Requests work.
// Before doing all the generator/databasey stuff.
type Patch interface {
	Op() PatchOp
	// question - I'm tempted to type this out instead of using raw []string, I think I prefer
	// that but have no strong feelings - what are your thoughts?
	Path() []string
}

var _ Patch = (*patchAdd)(nil)

type patchAdd struct {
	path []string
	// any type is consistent with request-recult (document/map[string]any)
	value any
}

func AppendSchema(schema string) Patch {
	return &patchAdd{
		path: []string{
			"-",
		},
		value: schema,
	}
}

func AppendField(schemaName string, fieldName string, fieldType string) Patch {
	return &patchAdd{
		path: []string{
			schemaName,
			"-",
		},
		value: fmt.Sprintf("%s: %s", fieldName, fieldType),
	}
}

func AddFieldAfter(schemaName string, afterFieldName string, fieldName string, fieldType string) Patch {
	return &patchAdd{
		path: []string{
			schemaName,
			afterFieldName,
		},
		value: fmt.Sprintf("%s: %s", fieldName, fieldType),
	}
}

func (p *patchAdd) Op() PatchOp {
	return PatchOpAdd
}

func (p *patchAdd) Path() []string {
	return p.path
}

var _ Patch = (*patchMove)(nil)

// move is consistent with json patch naming, rename is not
type patchMove struct {
	from []string
	path []string
}

func MoveField(schemaName string, src string, dst string) Patch {
	return &patchMove{
		from: []string{
			schemaName,
			src,
		},
		path: []string{
			schemaName,
			dst,
		},
	}
}

func MoveSchema(src string, dst string) Patch {
	return &patchMove{
		from: []string{
			src,
		},
		path: []string{
			dst,
		},
	}
}

func (p *patchMove) Op() PatchOp {
	return PatchOpMove
}

func (p *patchMove) Path() []string {
	return p.path
}

var _ Patch = (*patchRemove)(nil)

type patchRemove struct {
	path []string
}

func RemoveField(schemaName string, fieldName string) Patch {
	return &patchRemove{
		path: []string{
			schemaName,
			fieldName,
		},
	}
}

func RemoveSchema(schemaName string) Patch {
	return &patchRemove{
		path: []string{
			schemaName,
		},
	}
}

func (p *patchRemove) Op() PatchOp {
	return PatchOpRemove
}

func (p *patchRemove) Path() []string {
	return p.path
}
