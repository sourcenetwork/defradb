// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package wire records the types that a node sends to another node over CBOR.
//
// Each wire type registers itself from its own package's init, so the set has
// one source of truth that both runtime code and the wirecheck linter read: the
// linter fails a cbor.Marshal of a type that was never registered here, and a
// later snapshot check reads the registered set to catch a field-shape change.
package wire

import (
	"reflect"
	"sync"
)

var (
	mu         sync.Mutex
	registered = map[reflect.Type]struct{}{}
)

// Register records T as a type that crosses the wire. Call it from an init in the
// package that owns T. Registering the same type twice is a no-op.
func Register[T any]() {
	mu.Lock()
	defer mu.Unlock()
	registered[reflect.TypeFor[T]()] = struct{}{}
}

// MarkLocal records that T is CBOR-encoded only for local storage, never sent to
// a peer. It exists so the wirecheck linter stops flagging a local encode in a
// wire package, and so the decision that T is local is written down next to T.
// It does not add T to the wire set.
func MarkLocal[T any]() {}

// IsRegistered reports whether T has been registered.
func IsRegistered[T any]() bool {
	mu.Lock()
	defer mu.Unlock()
	_, ok := registered[reflect.TypeFor[T]()]
	return ok
}

// Registered returns the registered types, for the snapshot check.
func Registered() []reflect.Type {
	mu.Lock()
	defer mu.Unlock()
	types := make([]reflect.Type, 0, len(registered))
	for t := range registered {
		types = append(types, t)
	}
	return types
}
