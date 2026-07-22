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
// change to any field name, field type, or field order changes the output, so a
// diff against the committed golden snapshot flags a wire-format change that
// must be intentional. The CBOR encoder keys struct fields by their Go name, so
// the field names this records are the wire contract.
//
// Each registered type is rendered top to bottom, types sorted by path. A struct
// lists its fields; nested named structs are rendered once at the top level too
// if reachable, so their own shape is covered. An interface records its method
// set: its concrete implementers are registered and rendered separately.
func Snapshot() string {
	var b strings.Builder
	for _, t := range sortedByPath(reachableTypes()) {
		writeType(&b, t)
	}
	return b.String()
}

// reachableTypes returns the registered types plus every named struct reachable
// through their fields, so a shape change in a nested type is covered even if
// that type is not itself registered.
func reachableTypes() []reflect.Type {
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
		if t.Kind() == reflect.Struct {
			for i := range t.NumField() {
				visit(t.Field(i).Type)
			}
		}
	}
	for _, t := range Registered() {
		visit(t)
	}
	return out
}

// writeType renders one named type: a struct as its ordered fields, an interface
// as its method set, anything else as its kind.
func writeType(b *strings.Builder, t reflect.Type) {
	switch t.Kind() {
	case reflect.Struct:
		fmt.Fprintf(b, "%s struct\n", typePath(t))
		for i := range t.NumField() {
			f := t.Field(i)
			fmt.Fprintf(b, "\t%s %s\n", f.Name, typeName(f.Type))
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
