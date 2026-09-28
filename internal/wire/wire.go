// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package wire records the types that a node sends to another node.
//
// Each wire type registers itself from its own package's init, giving one source
// of truth for what crosses the wire. The snapshot check reads the registered set
// to catch a field-shape change that would break communication with an older node.
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
