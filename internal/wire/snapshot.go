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
		if hasIPLDSchema(t) {
			return
		}
		// Only struct fields are walked. A named type carried by an interface
		// method is not reached here; today no wire interface method carries one.
		if t.Kind() == reflect.Struct {
			for i := range t.NumField() {
				// Only exported fields are encoded, so only they are part of the
				// wire shape. Skipping the rest also avoids descending into a
				// stdlib type's unexported internals (e.g. time.Time).
				if f := t.Field(i); f.IsExported() {
					visit(f.Type)
				}
			}
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
			// Collapse the schema literal's own indentation to one tab so the
			// snapshot reads and diffs cleanly regardless of source formatting.
			fmt.Fprintf(b, "\t%s\n", strings.Join(strings.Fields(line), " "))
		}
		return
	}
	switch t.Kind() {
	case reflect.Struct:
		fmt.Fprintf(b, "%s struct\n", typePath(t))
		for i := range t.NumField() {
			// Only exported fields are encoded, so only they are the wire shape.
			if f := t.Field(i); f.IsExported() {
				fmt.Fprintf(b, "\t%s %s\n", f.Name, typeName(f.Type))
			}
		}
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

// typePath is the package-qualified name of a named type.
func typePath(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.PkgPath() + "." + t.Name()
}

// typeName renders a field's type: named types by full path, containers
// structurally, so a change to either is visible.
func typeName(t reflect.Type) string {
	if t.PkgPath() != "" {
		return typePath(t)
	}
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
		return t.String()
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
