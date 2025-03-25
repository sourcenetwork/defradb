// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package telemetry

import (
	"context"
)

var (
	_ Tracer = (*noopTracer)(nil)
	_ Span   = (*noopSpan)(nil)
)

type noopTracer struct{}

func NewTracer() Tracer {
	return &noopTracer{}
}

func (t noopTracer) Start(ctx context.Context) (context.Context, Span) {
	return ctx, &noopSpan{}
}

type noopSpan struct{}

func (s *noopSpan) End() {}

func ConfigureTelemetry(ctx context.Context, version string) error {
	return nil
}
