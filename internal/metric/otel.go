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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ Tracer = (*otelTracer)(nil)
	_ Span   = (*otelSpan)(nil)
)

type otelTracer struct {
	inner trace.Tracer
}

func NewTracer() Tracer {
	name := callingPackageName()
	tracer := otel.Tracer(name)
	return &otelTracer{tracer}
}

func (t otelTracer) Start(ctx context.Context) (context.Context, Span) {
	name := callingFuncName()
	ctx, span := t.inner.Start(ctx, name)
	return ctx, &otelSpan{span}
}

type otelSpan struct {
	inner trace.Span
}

func (s *otelSpan) End() {
	s.inner.End()
}
