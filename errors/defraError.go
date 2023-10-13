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
	"errors"
	"fmt"
	"io"
	"strings"
)

const StackKey string = "Stack"

var (
	_ error         = (*defraError)(nil)
	_ fmt.Formatter = (*defraError)(nil)
)

type defraError struct {
	message    string
	inner      error
	stacktrace string
	kvs        []KV
}

func (e *defraError) Error() string {
	builder := strings.Builder{}
	builder.WriteString(e.message)

	if e.inner != nil {
		builder.WriteString(": ")
		builder.WriteString(e.inner.Error())
	}

	if len(e.kvs) > 0 {
		builder.WriteString(".")
	}

	for i, kv := range e.kvs {
		builder.WriteString(" ")
		builder.WriteString(kv.key)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprint(kv.value))
		if i < len(e.kvs)-1 {
			builder.WriteString(",")
		}
	}

	return builder.String()
}

func (e *defraError) Is(other error) bool {
	var otherDefraError *defraError
	if errors.As(other, &otherDefraError) {
		return e.message == otherDefraError.message
	}
	otherString := other.Error()
	return e.message == otherString || e.Error() == otherString || errors.Is(e.inner, other)
}

func (e *defraError) Unwrap() error {
	return e.inner
}

// Format writes the error into the given state.
//
// Currently the following runes are supported: `v[+]` (+ also writes out the stacktrace), `s`, `q`.
func (e *defraError) Format(f fmt.State, verb rune) {
	errorString := e.Error()
	switch verb {
	case 'v':
		_, _ = io.WriteString(f, errorString)

		if f.Flag('+') {
			if len(errorString) > 0 && errorString[len(errorString)-1] != '.' {
				_, _ = io.WriteString(f, ".")
			}

			_, _ = io.WriteString(f, " "+StackKey+": ")
			_, _ = io.WriteString(f, e.stacktrace)
		}
	case 's':
		_, _ = io.WriteString(f, errorString)
	case 'q':
		_, _ = fmt.Fprintf(f, "%q", errorString)
	}
}
