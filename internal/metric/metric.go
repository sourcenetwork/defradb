// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metric

import (
	"context"
	"runtime"
	"strings"
)

// Tracer is used to create span telemetry.
type Tracer interface {
	// Start creates a new span.
	Start(context.Context) (context.Context, Span)
}

// Span represents a node in a function call tree.
type Span interface {
	// End completes the span.
	End()
}

// callerInfo returns the calling package name and calling func name.
func callerInfo() (string, string) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "", ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", ""
	}
	name := fn.Name()
	index := strings.LastIndex(name, ".")
	if index < 0 {
		return "", ""
	}
	return name[:index], name[index+1:]
}
