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
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Snapshot renders the shape of every registered type as deterministic text. A
// change to any field name, field type, field order, or IPLD discriminator
// changes the output, so a diff against the committed golden flags a wire-format
// change that must be intentional.
//
// Two wire layers are covered. CBOR message types are rendered from their Go
// fields, since the encoder keys fields by their Go name. IPLD block types are
// rendered from the schema they declare (including the CRDT union's wire-critical
// discriminator strings), since that schema, not the Go struct, is their wire
// shape.
//
// Types are sorted by path. A struct lists its fields; nested named structs are
// rendered at the top level too if reachable, so their shape is covered. An
// interface records its method set.
func Snapshot() string {
	return snapshotOf(Registered())
}

// snapshotOf renders the shape of the given roots plus their reachable named
// types. Split from Snapshot so it can be tested with an explicit set instead of
// the global registry.
func snapshotOf(roots []reflect.Type) string {
	var b strings.Builder
	for _, t := range sortedByPath(reachableFrom(roots)) {
		writeType(&b, t)
	}
	return b.String()
}

// reachableFrom returns roots plus every named struct reachable through their
// fields, so a shape change in a nested type is covered even if that type is not
// itself a root.
func reachableFrom(roots []reflect.Type) []reflect.Type {
	seen := map[reflect.Type]bool{}
	var out []reflect.Type
	var visit func(t reflect.Type)
	visit = func(t reflect.Type) {
		t = derefType(t)
		if t == nil || seen[t] {
			return
		}
		// Only named types (with a package path) have a shape worth recording.
		if t.PkgPath() == "" {
			// Descend through unnamed containers to reach named element types.
			switch t.Kind() {
			case reflect.Slice, reflect.Array, reflect.Pointer:
				visit(t.Elem())
			case reflect.Map:
				visit(t.Key())
				visit(t.Elem())
			}
			return
		}
		seen[t] = true
		out = append(out, t)
		// A type that declares an IPLD schema is rendered from that schema, so its
		// Go fields are not walked (they hold ipld/cid internals, not the shape).
		// Types the schema names (a union's variants, a linked block) are not
		// reached this way, so each must be registered on its own to be covered.
		if hasIPLDSchema(t) {
			return
		}
		switch t.Kind() {
		case reflect.Struct:
			for i := range t.NumField() {
				// Only exported fields are encoded, so only they are part of the
				// wire shape. Skipping the rest also avoids descending into a
				// stdlib type's unexported internals (e.g. time.Time).
				if f := t.Field(i); f.IsExported() {
					visit(f.Type)
				}
			}
		case reflect.Slice, reflect.Array, reflect.Pointer:
			// A named container (type Items []Item) still carries its element's
			// shape on the wire, so descend to record that element type.
			visit(t.Elem())
		case reflect.Map:
			visit(t.Key())
			visit(t.Elem())
		}
	}
	for _, t := range roots {
		visit(t)
	}
	return out
}

// writeType renders one named type. A type that declares an IPLD schema is
// rendered from that schema, which already describes its fields and, for the
// union, its wire-critical discriminators. Otherwise a struct is rendered as its
// ordered fields, an interface as its method set.
func writeType(b *strings.Builder, t reflect.Type) {
	if schema, ok := ipldSchema(t); ok {
		fmt.Fprintf(b, "%s ipld\n", typePath(t))
		for line := range strings.SplitSeq(strings.TrimSpace(schema), "\n") {
			// Normalize the schema literal's own indentation so the snapshot reads
			// and diffs cleanly regardless of source formatting, without touching
			// whitespace inside a quoted discriminator.
			fmt.Fprintf(b, "\t%s\n", collapseUnquoted(line))
		}
		return
	}
	switch t.Kind() {
	case reflect.Struct:
		fmt.Fprintf(b, "%s struct%s\n", typePath(t), structOptions(t))
		writeStructFields(b, t)
	case reflect.Interface:
		fmt.Fprintf(b, "%s interface\n", typePath(t))
		for _, m := range sortedMethods(t) {
			fmt.Fprintf(b, "\t%s\n", m)
		}
	default:
		fmt.Fprintf(b, "%s %s\n", typePath(t), t.Kind())
	}
}

// ipldSchemaMethod is the method an IPLD wire type uses to declare its schema.
const ipldSchemaMethod = "IPLDSchemaBytes"

// hasIPLDSchema reports whether t declares an IPLD schema.
func hasIPLDSchema(t reflect.Type) bool {
	_, ok := ipldSchema(t)
	return ok
}

// ipldSchema returns t's declared IPLD schema, checking both value and pointer
// receivers. Returns false if t declares no schema.
func ipldSchema(t reflect.Type) (string, bool) {
	for _, rt := range []reflect.Type{t, reflect.PointerTo(t)} {
		m, ok := rt.MethodByName(ipldSchemaMethod)
		if !ok || m.Type.NumIn() != 1 || m.Type.NumOut() != 1 {
			continue
		}
		v := reflect.New(t)
		if rt.Kind() != reflect.Pointer {
			v = v.Elem()
		}
		out := m.Func.Call([]reflect.Value{v})
		if b, ok := out[0].Interface().([]byte); ok {
			return string(b), true
		}
	}
	return "", false
}

// writeStructFields renders a struct's exported fields as wire lines. An
// embedded struct with no renaming tag is flattened by the encoder, so its own
// fields are rendered inline rather than as a nested field.
func writeStructFields(b *strings.Builder, t reflect.Type) {
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key, opts, skip := cborField(f)
		if skip {
			continue
		}
		if f.Anonymous && opts == "" && !hasNameTag(f) && derefType(f.Type).Kind() == reflect.Struct {
			writeStructFields(b, derefType(f.Type))
			continue
		}
		// The encoded key, not the Go field name, is the wire contract: the encoder
		// honors a cbor or json tag. Options (omitempty, keyasint) are recorded too
		// since they change the encoding.
		fmt.Fprintf(b, "\t%s %s%s\n", key, typeName(f.Type), opts)
	}
}

// structOptions returns a struct-level encoding marker (e.g. toarray) declared by
// a blank field tag, or empty. Such a field flips the whole struct between map
// and positional-array encoding, so it is part of the wire shape.
func structOptions(t reflect.Type) string {
	if t.Kind() != reflect.Struct {
		return ""
	}
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Name != "_" {
			continue
		}
		if _, rest, _ := strings.Cut(f.Tag.Get("cbor"), ","); rest != "" {
			return " [" + rest + "]"
		}
	}
	return ""
}

// hasNameTag reports whether a field's cbor or json tag renames it (a leading
// name before any comma). An embedded field with such a tag is nested, not
// flattened.
func hasNameTag(f reflect.StructField) bool {
	tag := f.Tag.Get("cbor")
	if tag == "" {
		tag = f.Tag.Get("json")
	}
	name, _, _ := strings.Cut(tag, ",")
	return name != "" && name != "-"
}

// collapseUnquoted trims a line and collapses runs of whitespace to a single
// space, but leaves whitespace inside double-quoted spans untouched so distinct
// discriminators like "a  b" and "a b" stay distinct.
func collapseUnquoted(line string) string {
	var out strings.Builder
	inQuote := false
	prevSpace := false
	for _, r := range strings.TrimSpace(line) {
		if r == '"' {
			inQuote = !inQuote
		}
		if !inQuote && (r == ' ' || r == '\t') {
			if !prevSpace {
				out.WriteByte(' ')
			}
			prevSpace = true
			continue
		}
		out.WriteRune(r)
		prevSpace = false
	}
	return out.String()
}

// cborField resolves how a struct field is encoded: its wire key, an options
// suffix (e.g. ",omitempty"), and whether it is skipped entirely. It follows the
// encoder's precedence, checking the cbor tag then the json tag, so a rename via
// either tag changes the snapshot.
func cborField(f reflect.StructField) (key, opts string, skip bool) {
	tag := f.Tag.Get("cbor")
	if tag == "" {
		tag = f.Tag.Get("json")
	}
	name, rest, _ := strings.Cut(tag, ",")
	if name == "-" && rest == "" {
		return "", "", true
	}
	if name == "" {
		name = f.Name
	}
	if rest != "" {
		opts = " [" + rest + "]"
	}
	return name, opts, false
}

// typePath is the package-qualified name of a named type.
func typePath(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.PkgPath() + "." + t.Name()
}

// typeName renders a field's type. A named struct or interface is named by path,
// since its own shape is rendered as its own entry. A named container (e.g.
// type Links [][]byte) is named by path plus its underlying shape, so a change to
// its element type is still visible. Unnamed types render structurally.
func typeName(t reflect.Type) string {
	if t.PkgPath() != "" {
		switch t.Kind() {
		case reflect.Struct, reflect.Interface:
			return typePath(t)
		default:
			return typePath(t) + "=" + underlyingName(t)
		}
	}
	return underlyingName(t)
}

// underlyingName renders a type structurally, ignoring any name it has.
func underlyingName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Pointer:
		return "*" + typeName(t.Elem())
	case reflect.Slice:
		return "[]" + typeName(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), typeName(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", typeName(t.Key()), typeName(t.Elem()))
	default:
		return t.Kind().String()
	}
}

// sortedMethods returns an interface's method signatures, sorted for stability.
func sortedMethods(t reflect.Type) []string {
	methods := make([]string, 0, t.NumMethod())
	for i := range t.NumMethod() {
		m := t.Method(i)
		methods = append(methods, m.Name+" "+m.Type.String())
	}
	sort.Strings(methods)
	return methods
}

// sortedByPath sorts types by their package-qualified name for a stable snapshot.
func sortedByPath(types []reflect.Type) []reflect.Type {
	sort.Slice(types, func(i, j int) bool {
		return typePath(types[i]) < typePath(types[j])
	})
	return types
}

// derefType unwraps a single pointer.
func derefType(t reflect.Type) reflect.Type {
	if t != nil && t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}
