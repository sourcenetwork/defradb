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
	"strings"
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

// selfCycle references itself; cycleA and cycleB reference each other. The walk
// must terminate on both via the seen set.
type selfCycle struct {
	Self *selfCycle
	Name string
}

type cycleA struct{ B *cycleB }
type cycleB struct{ A *cycleA }

// twoRefs reaches nested by two different fields; nested must render exactly once.
type twoRefs struct {
	First  nested
	Second nested
}

// tagged fields carry encoder tags: cbor wins over json, json renames, "-" skips,
// options are recorded.
type tagged struct {
	Renamed  string `cbor:"wire_name"`
	FromJSON string `json:"json_name"`
	Both     string `cbor:"cbor_wins" json:"json_loses"` // cbor tag takes precedence
	Opt      string `cbor:"opt,omitempty"`
	Skipped  string `cbor:"-"`
	DashName string `cbor:"-,omitempty"` // literal key "-", not skipped
	Plain    string
}

// embedder embeds Meta with no tag: CBOR flattens Meta's fields into it. The
// nested variant instead nests because the cbor tag renames the embed.
type Meta struct {
	M1 string
	M2 string
}

type embedder struct {
	Meta
	Own string
}

type nestedEmbedder struct {
	Meta `cbor:"meta"`
	Own  string
}

// arrayEncoded flips to positional-array encoding via a blank toarray marker.
type arrayEncoded struct {
	_ struct{} `cbor:",toarray"`
	X string
	Y string
}

// namedElem is a named container whose element is a struct; the element's fields
// must still be recorded.
type namedElem []elem
type elem struct{ Value string }
type holdsNamedElem struct{ E namedElem }

// namedList is a named container: a change to its element type must still show.
type namedList [][]byte

type holdsNamedList struct {
	L namedList
}

// schemaSpaced declares a discriminator whose quoted value contains a double
// space, which must survive normalization.
type schemaSpaced struct{}

func (schemaSpaced) IPLDSchemaBytes() []byte {
	return []byte("type schemaSpaced union {\n\t| A \"a  b\"\n} representation keyed")
}

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
	// The full map type including the value, so a value-type change is detected.
	assert.Contains(t, out, "ByKey map[string]"+typePath(reflect.TypeFor[nested]()))
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

// TestSnapshot_Cycles terminates on self-referential and mutually-referential
// types rather than looping, and still traverses their fields. A regression that
// recurses before marking a type seen would hang; one that skips fields would
// drop the referenced type.
func TestSnapshot_Cycles(t *testing.T) {
	self := snap(reflect.TypeFor[selfCycle]())
	assert.Contains(t, self, "Self *"+typePath(reflect.TypeFor[selfCycle]()))
	assert.Contains(t, self, "Name string")

	mutual := snap(reflect.TypeFor[cycleA]())
	assert.Contains(t, mutual, typePath(reflect.TypeFor[cycleA]())+" struct")
	assert.Contains(t, mutual, typePath(reflect.TypeFor[cycleB]())+" struct")
}

// TestSnapshot_RendersEachTypeOnce renders a type reachable by two paths exactly
// once, so a dedup regression would show up as a repeated block.
func TestSnapshot_RendersEachTypeOnce(t *testing.T) {
	out := snap(reflect.TypeFor[twoRefs]())
	header := typePath(reflect.TypeFor[nested]()) + " struct"
	assert.Equal(t, 1, strings.Count(out, header), "nested must render exactly once")
}

// TestSnapshot_EncoderTags records the encoded key, not the Go field name: a cbor
// tag wins, a json tag renames, "-" skips, and options are recorded.
func TestSnapshot_EncoderTags(t *testing.T) {
	out := snap(reflect.TypeFor[tagged]())
	assert.Contains(t, out, "wire_name string")
	assert.Contains(t, out, "json_name string")
	assert.Contains(t, out, "cbor_wins string", "cbor tag wins over json when both are present")
	assert.NotContains(t, out, "json_loses", "the json tag must not win over the cbor tag")
	assert.Contains(t, out, "opt string [omitempty]")
	assert.Contains(t, out, "Plain string")
	assert.Contains(t, out, "- string [omitempty]", `"-,omitempty" is a literal key, not a skip`)
	assert.NotContains(t, out, "Renamed", "the Go name must not appear once tagged")
	assert.NotContains(t, out, "Skipped", "a - tag drops the field")
}

// TestSnapshot_EmbeddedFlatten renders an untagged embedded struct's fields
// inline (as CBOR flattens them), but a tag-renamed embed as a nested field, so
// switching between the two is detected.
func TestSnapshot_EmbeddedFlatten(t *testing.T) {
	flat := snap(reflect.TypeFor[embedder]())
	assert.Contains(t, flat, "M1 string")
	assert.Contains(t, flat, "M2 string")
	assert.Contains(t, flat, "Own string")

	nested := snap(reflect.TypeFor[nestedEmbedder]())
	assert.Contains(t, nested, "meta "+typePath(reflect.TypeFor[Meta]()),
		"a tag-renamed embed is a nested field, not flattened")
}

// TestSnapshot_ToArray records a struct-level toarray marker even though it lives
// on an unexported blank field, since it changes map vs array encoding.
func TestSnapshot_ToArray(t *testing.T) {
	out := snap(reflect.TypeFor[arrayEncoded]())
	assert.Contains(t, out, "struct [toarray]")
}

// TestSnapshot_NamedContainerElement records the fields of a struct reached only
// through a named container, so a change to that element is detected.
func TestSnapshot_NamedContainerElement(t *testing.T) {
	out := snap(reflect.TypeFor[holdsNamedElem]())
	assert.Contains(t, out, typePath(reflect.TypeFor[elem]())+" struct")
	assert.Contains(t, out, "Value string")
}

// TestSnapshot_NamedContainer renders a named container's underlying shape, so a
// change to its element type is visible.
func TestSnapshot_NamedContainer(t *testing.T) {
	out := snap(reflect.TypeFor[holdsNamedList]())
	assert.Contains(t, out, typePath(reflect.TypeFor[namedList]())+"=[][]uint8")
}

// TestSnapshot_QuotedWhitespace keeps whitespace inside a quoted discriminator,
// so "a  b" and "a b" do not collapse to the same text.
func TestSnapshot_QuotedWhitespace(t *testing.T) {
	out := snap(reflect.TypeFor[schemaSpaced]())
	assert.Contains(t, out, `"a  b"`)
}
