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
	"time"

	"github.com/stretchr/testify/assert"
)

// The per-block fetch timeout should default to the node setting, and a positive per-request
// override carried on the context should take precedence. A zero or negative override is ignored
// so a caller cannot accidentally disable the timeout.
func TestBlockSyncTimeout_OverrideResolution(t *testing.T) {
	p := &P2P{syncBlockLinkTimeout: 5 * time.Second}

	assert.Equal(t, 5*time.Second, p.blockSyncTimeout(context.Background()),
		"with no override the node default should be used")

	overridden := WithBlockSyncTimeout(context.Background(), 30*time.Second)
	assert.Equal(t, 30*time.Second, p.blockSyncTimeout(overridden),
		"a positive override should take precedence over the node default")

	zeroOverride := WithBlockSyncTimeout(context.Background(), 0)
	assert.Equal(t, 5*time.Second, p.blockSyncTimeout(zeroOverride),
		"a non-positive override should be ignored in favour of the node default")
}
