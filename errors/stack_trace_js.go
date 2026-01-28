// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package errors

// withStackTrace creates a `defraError` with a stacktrace and the given key-value pairs.
//
// The stacktrace will skip the top `depthToSkip` frames, allowing frames/calls generated from
// within this package to not polute the resultant stacktrace.
//
// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//
//go:noinline
func withStackTrace(message string, depthToSkip int, keyvals ...KV) *defraError {
	// stack tracing does not work in WASM builds
	return &defraError{
		message: message,
		kvs:     keyvals,
	}
}
