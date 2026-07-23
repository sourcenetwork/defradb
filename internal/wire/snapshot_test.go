// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wire

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nested struct {
	Inner string
}

type structFields struct {
	Name       string
	Count      int
	Ptr        *nested
	List       []nested
	Blob       []byte
	unexported string //nolint:unused
}

type containers struct {
	Bytes    []byte
	Rows     [][]string
	ByKey    map[string]nested
	FixedArr [3]byte
}

type wireIface interface {
	Do(string) int
}

// schemaValue declares its schema with a value receiver; schemaPtr with a
// pointer receiver. Both must be rendered from the schema, not their Go fields.
type schemaValue struct{ Ignored int }

func (schemaValue) IPLDSchemaBytes() []byte { return []byte("type schemaValue struct { a String }") }

type schemaPtr struct{ Ignored int }

func (*schemaPtr) IPLDSchemaBytes() []byte { return []byte("type schemaPtr struct { b Int }") }

func snap(t ...reflect.Type) string { return snapshotOf(t) }

// TestSnapshot_StructFields renders exported fields with their types and skips
// unexported ones, since only exported fields are encoded.
func TestSnapshot_StructFields(t *testing.T) {
	out := snap(reflect.TypeFor[structFields]())
	assert.Contains(t, out, "Name string")
	assert.Contains(t, out, "Count int")
	assert.Contains(t, out, "Blob []uint8")
	assert.NotContains(t, out, "unexported", "unexported fields are not on the wire")
}

// TestSnapshot_ReachesNested renders a nested named struct even though only the
// parent was passed, so a change to the nested type is covered.
func TestSnapshot_ReachesNested(t *testing.T) {
	out := snap(reflect.TypeFor[structFields]())
	require.Contains(t, out, typePath(reflect.TypeFor[nested]())+" struct")
	assert.Contains(t, out, "Inner string")
}

// TestSnapshot_Containers renders slice, nested slice, map, and array field
// types structurally so a change to any element type is visible.
func TestSnapshot_Containers(t *testing.T) {
	out := snap(reflect.TypeFor[containers]())
	assert.Contains(t, out, "Bytes []uint8")
	assert.Contains(t, out, "Rows [][]string")
	assert.Contains(t, out, "FixedArr [3]uint8")
	assert.Contains(t, out, "ByKey map[string]")
}

// TestSnapshot_Interface renders an interface's method set.
func TestSnapshot_Interface(t *testing.T) {
	out := snap(reflect.TypeFor[wireIface]())
	assert.Contains(t, out, typePath(reflect.TypeFor[wireIface]())+" interface")
	assert.Contains(t, out, "Do func(string) int")
}

// TestSnapshot_IPLDSchema renders a type from its declared schema, for both a
// value and a pointer receiver, and does not render its Go fields.
func TestSnapshot_IPLDSchema(t *testing.T) {
	out := snap(reflect.TypeFor[schemaValue](), reflect.TypeFor[schemaPtr]())
	assert.Contains(t, out, typePath(reflect.TypeFor[schemaValue]())+" ipld")
	assert.Contains(t, out, "type schemaValue struct { a String }")
	assert.Contains(t, out, "type schemaPtr struct { b Int }")
	assert.NotContains(t, out, "Ignored", "IPLD types are rendered from schema, not fields")
}

// TestSnapshot_Deterministic renders the same output regardless of root order,
// so map-iteration order in the registry cannot perturb the snapshot.
func TestSnapshot_Deterministic(t *testing.T) {
	a := snap(reflect.TypeFor[structFields](), reflect.TypeFor[containers]())
	b := snap(reflect.TypeFor[containers](), reflect.TypeFor[structFields]())
	assert.Equal(t, a, b)
}

// TestSnapshot_PointerRootDerefs treats a *T root the same as T.
func TestSnapshot_PointerRootDerefs(t *testing.T) {
	assert.Equal(t,
		snap(reflect.TypeFor[structFields]()),
		snap(reflect.TypeFor[*structFields]()))
}
