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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorIs(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)

	assert.Equal(t, true, Is(Wrap("wrapped error", err), err))
}

func TestErrorIsDefraError(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)

	assert.ErrorIs(t, err, errors.New(errorMessage), err)
}

func TestErrorWithStack(t *testing.T) {
	t.Parallel()
	errorMessage := "gndjdhs"
	err := errors.New(errorMessage)

	errWithStack := WithStack(err)

	result := fmt.Sprintf("%+v", errWithStack)

	// Assert that the first line starts with the error message and contains this [test] function's stacktrace-line -
	// including file, line number, and function reference. An exact string match should not be used as the stacktrace
	// is machine dependent.
	assert.Regexp(t, fmt.Sprintf("^%s\\. Stack: .*\\/.*\\/errors\\/errors_test\\.go:[0-9]+ \\([a-zA-Z0-9]*\\)", errorMessage), result)

	// Assert that the error contains this function's name, and a print-out of the generating line.
	assert.Regexp(t, "TestErrorWithStack: errWithStack := WithStack\\(err\\)", result)

	// Assert that the next line of the stacktrace is also present.
	assert.Regexp(t, ".*\\/testing/testing.go:[0-9]+ \\([a-zA-Z0-9]*\\)", result)
}

func TestErrorWrap(t *testing.T) {
	t.Parallel()
	const errorMessage1 string = "gndjdhs"
	const errorMessage2 string = "nhdfbgshna"

	err1 := New(errorMessage1)
	err2 := Wrap(errorMessage2, err1)

	assert.ErrorIs(t, err2, errors.New(errorMessage1))
}

func TestErrorUnwrap(t *testing.T) {
	t.Parallel()
	const errorMessage1 string = "gndjdhs"
	const errorMessage2 string = "nhdfbgshna"

	err1 := New(errorMessage1)
	err2 := Wrap(errorMessage2, err1)

	unwrapped := errors.Unwrap(err2)

	assert.ErrorIs(t, unwrapped, errors.New(errorMessage1))
}

func TestErrorAs(t *testing.T) {
	t.Parallel()
	const errorMessage1 string = "gndjdhs"
	const errorMessage2 string = "nhdfbgshna"

	err1 := New(errorMessage1)
	err2 := fmt.Errorf("%s: %w", errorMessage2, err1)

	target := &defraError{}
	isErr1 := errors.As(err2, &target)

	assert.True(t, isErr1)
	assert.ErrorIs(t, target, errors.New(errorMessage1))
}

func TestErrorFmts(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)
	result := fmt.Sprintf("%s", err)

	assert.Equal(t, errorMessage, result)
}

func TestErrorFmtq(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)
	result := fmt.Sprintf("%q", err)

	assert.Equal(t, "\""+errorMessage+"\"", result)
}

func TestErrorFmtvWithoutStacktrace(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)
	result := fmt.Sprintf("%v", err)

	assert.Equal(t, errorMessage, result)
}

func TestErrorFmtsWithKvp(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage, NewKV("Kv1", 1))
	result := fmt.Sprintf("%s", err)

	assert.Equal(t, errorMessage+". Kv1: 1", result)
}

func TestErrorFmtsWithManyKvps(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage, NewKV("Kv1", 1), NewKV("Kv2", "2"))
	result := fmt.Sprintf("%s", err)

	assert.Equal(t, errorMessage+". Kv1: 1, Kv2: 2", result)
}

func TestErrorFmtvWithStacktrace(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage)

	result := fmt.Sprintf("%+v", err)

	// Assert that the first line starts with the error message and contains this [test] function's stacktrace-line -
	// including file, line number, and function reference. An exact string match should not be used as the stacktrace
	// is machine dependent.
	assert.Regexp(t, fmt.Sprintf("^%s\\. Stack: .*\\/.*\\/errors\\/errors_test\\.go:[0-9]+ \\([a-zA-Z0-9]*\\)", errorMessage), result)

	// Assert that the error contains this function's name, and a print-out of the generating line.
	assert.Regexp(t, "TestErrorFmtvWithStacktrace: err := New\\(errorMessage\\)", result)

	// Assert that the next line of the stacktrace is also present.
	assert.Regexp(t, ".*\\/testing/testing.go:[0-9]+ \\([a-zA-Z0-9]*\\)", result)
}

func TestErrorFmtvWithStacktraceAndKvps(t *testing.T) {
	t.Parallel()
	const errorMessage string = "gndjdhs"

	err := New(errorMessage, NewKV("Kv1", 1), NewKV("Kv2", "2"))
	result := fmt.Sprintf("%+v", err)

	// Assert that the first line starts with the error message and contains this [test] function's stacktrace-line -
	// including file, line number, and function reference. An exact string match should not be used as the stacktrace
	// is machine dependent.
	assert.Regexp(t, fmt.Sprintf("^%s\\. Kv1: 1, Kv2: 2\\. Stack: .*\\/.*\\/errors\\/errors_test\\.go:[0-9]+ \\([a-zA-Z0-9]*\\)", errorMessage), result)

	// Assert that the error contains this function's name, and a print-out of the generating line.
	assert.Regexp(t, "TestErrorFmtvWithStacktraceAndKvps: err := New\\(errorMessage, NewKV\\(\"Kv1\", 1\\), NewKV\\(\"Kv2\", \"2\"\\)\\)", result)

	// Assert that the next line of the stacktrace is also present.
	assert.Regexp(t, ".*\\/testing/testing.go:[0-9]+ \\([a-zA-Z0-9]*\\)", result)
}
