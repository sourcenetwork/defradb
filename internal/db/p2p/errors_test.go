// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/errors"
)

// A block-fetch timeout must be identifiable as such and must not be confused with the generic
// linked-block load failure, so callers and operators can tell "the peer was too slow" apart
// from a decode/storage error.
func TestBlockSyncTimeoutError_IsDistinctAndWrapsCause(t *testing.T) {
	err := NewErrBlockSyncTimeout(context.DeadlineExceeded, "bafyLink")

	assert.True(t, errors.Is(err, ErrBlockSyncTimeout), "should be identifiable as a block-sync timeout")
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "should preserve the deadline-exceeded cause")

	generic := NewErrLoadLinkedBlock(context.DeadlineExceeded)
	assert.False(t, errors.Is(generic, ErrBlockSyncTimeout),
		"the generic load error must not masquerade as a block-sync timeout")
}
