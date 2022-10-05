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
func New(message string, keyvals ...KV) error {
	return newError(message, keyvals...)
}

// Wrap creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided.
func Wrap(message string, inner error, keyvals ...KV) error {
	err := newError(message, keyvals...)
	err.inner = inner
	return err
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func WithStack(err error, keyvals ...KV) error {
	return withStackTrace(err.Error(), keyvals...)
}

func newError(message string, keyvals ...KV) *defraError {
	return withStackTrace(message, keyvals...)
}

func withStackTrace(message string, keyvals ...KV) *defraError {
	stackBuffer := make([]uintptr, MaxStackDepth)
	// Skip the first X frames as they are part of this library (and dependencies) and are
	// best hidden.
	length := runtime.Callers(4, stackBuffer[:])
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
