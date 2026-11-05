// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package debug provides utilities for debugging and tracing code execution.
// It includes a hierarchical logger that visualizes call stacks with indentation.
package debug

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// PrefixWidth is the fixed width for tracer prefixes. All prefixes will be
// padded or truncated to this width for aligned output.
var PrefixWidth = 12

// ShowSourceLocation enables showing file:line information at the end of each log line.
var ShowSourceLocation = true

// globalIndent is the shared indentation level across all tracers.
var globalIndent atomic.Int32

// globalMu protects print operations to ensure atomic output.
var globalMu sync.Mutex

// Tracer provides hierarchical logging with indentation to visualize call stacks.
// It is safe for concurrent use from multiple goroutines.
// All Tracer instances share a global indentation level.
type Tracer struct {
	mu      sync.Mutex
	enabled bool
	prefix  string
}

// NewTracer creates a new Tracer with the given prefix.
// The prefix is prepended to all output lines.
func NewTracer(prefix string) *Tracer {
	return &Tracer{
		prefix:  prefix,
		enabled: true,
	}
}

// Enable turns on tracing output.
func (t *Tracer) Enable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = true
}

// Disable turns off tracing output.
func (t *Tracer) Disable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = false
}

// IsEnabled returns whether tracing is currently enabled.
func (t *Tracer) IsEnabled() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enabled
}

// Indent increases the global indentation level by one.
func (t *Tracer) Indent() {
	globalIndent.Add(1)
}

// Outdent decreases the global indentation level by one.
func (t *Tracer) Outdent() {
	if globalIndent.Load() > 0 {
		globalIndent.Add(-1)
	}
}

// indentStr returns the current indentation string.
func indentStr() string {
	indent := int(globalIndent.Load())
	var sb strings.Builder
	for i := 0; i < indent; i++ {
		sb.WriteString("    ")
	}
	return sb.String()
}

// Print outputs a formatted message with the current indentation.
func (t *Tracer) Print(format string, args ...any) {
	t.mu.Lock()
	enabled := t.enabled
	prefix := t.formattedPrefix()
	t.mu.Unlock()

	if !enabled {
		return
	}

	globalMu.Lock()
	defer globalMu.Unlock()
	fmt.Printf("%s%s"+format, append([]any{prefix, indentStr()}, args...)...)
}

// formattedPrefix returns the prefix formatted to fixed width (must be called with lock held).
func (t *Tracer) formattedPrefix() string {
	if t.prefix == "" {
		return strings.Repeat(" ", PrefixWidth+1) // +1 for ":"
	}
	// Add colon after prefix, then pad with spaces to fixed width
	p := t.prefix + ":"
	if len(p) > PrefixWidth+1 {
		p = p[:PrefixWidth+1]
	}
	// Pad to PrefixWidth+1 (for the colon) plus one space
	return fmt.Sprintf("%-*s ", PrefixWidth+1, p)
}

// Println outputs a formatted message with the current indentation, followed by a newline.
func (t *Tracer) Println(format string, args ...any) {
	if ShowSourceLocation {
		// Get source location before calling Print to ensure correct skip level
		loc := sourceLocation(1)
		t.Print(format+" | %s\n", append(args, loc)...)
	} else {
		t.Print(format+"\n", args...)
	}
}

// sourceLocation returns the file:line of the caller at the given skip level.
func sourceLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "???:0"
	}
	// Get just the filename, not full path
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// Enter logs entry into a function, increases indentation, and returns a function
// to be deferred that decreases indentation on exit.
// Usage:
//
//	defer tracer.Enter("MyFunc")()
//
// Or with object address:
//
//	defer tracer.Enter("MyFunc", myObject)()
func (t *Tracer) Enter(name string, obj ...any) func() {
	// Capture source location at call site
	loc := sourceLocation(1)
	// Print entry at current level with ">" marker, then indent for body
	if len(obj) > 0 {
		t.printEntry("> %s (%x)", loc, name, Addr(obj[0]))
	} else {
		t.printEntry("> %s", loc, name)
	}
	t.Indent()
	return func() {
		t.Outdent()
	}
}

// printEntry prints an entry line with a pre-captured source location.
func (t *Tracer) printEntry(format string, loc string, args ...any) {
	t.mu.Lock()
	enabled := t.enabled
	prefix := t.formattedPrefix()
	t.mu.Unlock()

	if !enabled {
		return
	}

	globalMu.Lock()
	defer globalMu.Unlock()
	if ShowSourceLocation {
		fmt.Printf("%s%s"+format+" | %s\n", append([]any{prefix, indentStr()}, append(args, loc)...)...)
	} else {
		fmt.Printf("%s%s"+format+"\n", append([]any{prefix, indentStr()}, args...)...)
	}
}

// EnterFunc logs entry into the calling function and returns a defer function.
// It automatically determines the caller's function name.
// Usage:
//
//	defer tracer.EnterFunc()()
//
// Or with object address:
//
//	defer tracer.EnterFunc(myObject)()
func (t *Tracer) EnterFunc(obj ...any) func() {
	name := callerName(2)
	return t.Enter(name, obj...)
}

// Addr returns the memory address of an object as a uintptr.
// Works with pointers, interfaces, maps, slices, channels, and functions.
func Addr(obj any) uintptr {
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if v.IsNil() {
			return 0
		}
		return v.Pointer()
	default:
		return 0
	}
}

// callerName returns the name of the caller at the given skip level.
func callerName(skip int) string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, pc)
	if n == 0 {
		return "unknown"
	}
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	// Extract just the function name without the full path
	name := frame.Function
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

// Default is a default global tracer that can be used for quick debugging.
var Default = NewTracer("")
