// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/db"
)

func TestNodeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []NodeOpt{
		WithStoreOpts(WithPath(t.TempDir())),
		WithDatabaseOpts(db.WithUpdateEvents()),
	}

	node, err := NewNode(ctx, opts...)
	require.NoError(t, err)

	err = node.Start(ctx)
	require.NoError(t, err)

	<-time.After(5 * time.Second)

	err = node.Close(ctx)
	require.NoError(t, err)
}
