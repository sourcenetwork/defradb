// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package errors

const InnerErrorKey string = "Inner"

type KV struct {
	key   string
	value interface{}
}

func NewKV(key string, value interface{}) KV {
	return KV{
		key:   key,
		value: value,
	}
}

// Error creates a new Defra error, suffixing any key-value
// pairs provided.
//
// Does not include a stacktrace.
func Error(message string, keyvals ...KV) error {
	panic("todo")
}

// ErrorS creates a new Defra error, suffixing any key-value
// pairs provided, and a stacktrace.
func ErrorS(message string, keyvals ...KV) error {
	panic("todo")
}

// Wrap creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided.
//
// Does not include a stacktrace.
func Wrap(message string, inner error, keyvals ...KV) error {
	newKeyVals := keyvals
	newKeyVals = append(newKeyVals, NewKV(InnerErrorKey, inner))
	return Error(message, newKeyVals...)
}

// WrapS creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided
// and a stacktrace.
func WrapS(message string, inner error, keyvals ...KV) error {
	newKeyVals := keyvals
	newKeyVals = append(newKeyVals, NewKV(InnerErrorKey, inner))
	return Error(message, newKeyVals...)
}
