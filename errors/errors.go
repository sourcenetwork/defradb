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

import (
	"bytes"
	"errors"
	"runtime"

	goErrors "github.com/go-errors/errors"
)

// todo: make this configurable:
// https://github.com/sourcenetwork/defradb/issues/733
const MaxStackDepth int = 50

type KV struct {
	key   string
	value any
}

func NewKV(key string, value any) KV {
	return KV{
		key:   key,
		value: value,
	}
}

// New creates a new Defra error, suffixing any key-value
// pairs provided.
//
// A stacktrace will be yielded if formatting with a `+`, e.g `fmt.Sprintf("%+v", err)`.
// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//go:noinline
func New(message string, keyvals ...KV) error {
	return newError(message, 1, keyvals...)
}

// Wrap creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided.
// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//go:noinline
func Wrap(message string, inner error, keyvals ...KV) error {
	err := newError(message, 1, keyvals...)
	err.inner = inner
	return err
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//go:noinline
func WithStack(err error, keyvals ...KV) error {
	return withStackTrace(err.Error(), 1, keyvals...)
}

// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//go:noinline
func newError(message string, depthToSkip int, keyvals ...KV) *defraError {
	return withStackTrace(message, depthToSkip+1, keyvals...)
}

// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//go:noinline
func withStackTrace(message string, depthToSkip int, keyvals ...KV) *defraError {
	stackBuffer := make([]uintptr, MaxStackDepth)

	// Skip the first X frames as they are part of this library (and dependencies) and are
	// best hidden, also account for any parent calls within this library.
	const depthFromHereToSkip int = 2
	length := runtime.Callers(depthFromHereToSkip+depthToSkip, stackBuffer[:])
	stack := stackBuffer[:length]
	stackText := toString(stack)

	return &defraError{
		message:    message,
		stacktrace: stackText,
		kvs:        keyvals,
	}
}

func toString(stack []uintptr) string {
	buf := bytes.Buffer{}

	for _, pc := range stack {
		frame := goErrors.NewStackFrame(pc)
		buf.WriteString(frame.String())
	}
	return buf.String()
}
