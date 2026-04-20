// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSigningOverrideFromContext_WithNoOverride_ReturnsNone(t *testing.T) {
	ctx := context.Background()
	result := SigningOverrideFromContext(ctx)
	assert.False(t, result.HasValue())
}

func TestSigningOverrideFromContext_WithTrueOverride_ReturnsTrue(t *testing.T) {
	ctx := ContextWithSigningOverride(context.Background(), true)
	result := SigningOverrideFromContext(ctx)
	assert.True(t, result.HasValue())
	assert.True(t, result.Value())
}

func TestSigningOverrideFromContext_WithFalseOverride_ReturnsFalse(t *testing.T) {
	ctx := ContextWithSigningOverride(context.Background(), false)
	result := SigningOverrideFromContext(ctx)
	assert.True(t, result.HasValue())
	assert.False(t, result.Value())
}

func TestSigningOverride_DoesNotAffectEnabledSigning(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithSigningOverride(ctx, true)
	ctx = ContextWithEnabledSigning(ctx)

	override := SigningOverrideFromContext(ctx)
	assert.True(t, override.HasValue())
	assert.True(t, override.Value())

	enabled, _ := EnabledSigningFromContext(ctx)
	assert.True(t, enabled)
}

func TestEnabledSigning_DoesNotAffectSigningOverride(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithEnabledSigning(ctx)

	override := SigningOverrideFromContext(ctx)
	assert.False(t, override.HasValue())
}
