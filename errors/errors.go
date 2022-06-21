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
	"runtime"

	goErrors "github.com/go-errors/errors"
)

const InnerErrorKey string = "Inner"
const StackKey string = "Stack"

// todo: thread safety stuff (atomics?)
var MaxStackDepth = 50
var WithStackTrace = true

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
	if WithStackTrace {
		return withStackTrace(message, keyvals...)
	}
	return withoutStackTrace(message, keyvals...)
}

// Wrap creates a new error of the given message that contains
// the given inner error, suffixing any key-value pairs provided.
func Wrap(message string, inner error, keyvals ...KV) error {
	newKeyVals := keyvals
	newKeyVals = append(newKeyVals, NewKV(InnerErrorKey, inner))
	return Error(message, newKeyVals...)
}

func withoutStackTrace(message string, keyvals ...KV) error {
	panic("todo")
}

// ErrorS creates a new Defra error, suffixing any key-value
// pairs provided, and a stacktrace.
func withStackTrace(message string, keyvals ...KV) error {
	stackBuffer := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stackBuffer[:])
	stack := stackBuffer[:length]
	stackText := toString(stack)

	newKeyVals := keyvals
	newKeyVals = append(newKeyVals, NewKV(StackKey, stackText))

	return withoutStackTrace(message, newKeyVals...)
}

func toString(stack []uintptr) []byte {
	buf := bytes.Buffer{}

	for _, pc := range stack {
		frame := goErrors.NewStackFrame(pc)
		buf.WriteString(frame.String())
	}
	return buf.Bytes()
}
