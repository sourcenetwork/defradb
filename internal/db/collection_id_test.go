// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// A collection's generated ID must not depend on the order unrelated collections are written in.
// "Foo" is a prefix of the unrelated "FooBar", so if the collection-set order is nondeterministic
// the head scan for "Foo" can read "FooBar"'s head, changing Foo's ID between identical runs.
func TestAddCollectionDeterministicIDsWithSharedNamePrefix(t *testing.T) {
	ctx := context.Background()

	const sdl = `
		type Foo { name: String }
		type FooBar { name: String }
	`

	fooID := func() string {
		db, err := newBadgerDB(ctx)
		require.NoError(t, err)
		defer db.Close()

		cols, err := db.AddCollection(ctx, sdl)
		require.NoError(t, err)
		for _, c := range cols {
			if c.Name == "Foo" {
				return c.VersionID
			}
		}
		require.Fail(t, `collection "Foo" not returned`)
		return ""
	}

	ids := map[string]struct{}{}
	for i := 0; i < 30; i++ {
		ids[fooID()] = struct{}{}
	}
	require.Len(t, ids, 1, "Foo's collection ID must be identical across runs")
}
