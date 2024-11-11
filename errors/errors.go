// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package errors provides the internal error system.
*/
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

// KV is a key-value pair.
type KV struct {
	key   string
	value any
}

// NewKV creates a new key-value pair.
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
//
//go:noinline
func New(message string, keyvals ...KV) error {
	return withStackTrace(message, 1, keyvals...)
}

// Wrap creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided.
// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//
//go:noinline
func Wrap(message string, inner error, keyvals ...KV) error {
	err := withStackTrace(message, 1, keyvals...)
	err.inner = inner
	return err
}

// Is reports whether any error in err's tree matches target.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly
// calling its Unwrap() error or Unwrap() []error method. When err wraps multiple
// errors, Is examines err followed by a depth-first traversal of its children.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See [syscall.Errno.Is] for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call [Unwrap] on either.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
//
// A non-nil error returned by Join implements the Unwrap() []error method.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// This function will not be inlined by the compiler as it will spoil any stacktrace
// generated.
//
//go:noinline
func WithStack(err error, keyvals ...KV) error {
	return withStackTrace(err.Error(), 1, keyvals...)
}

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
