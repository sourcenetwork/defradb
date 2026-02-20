// Copyright 2026 Democratized Data Foundation
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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAddSchema_Concurrent_NoRace verifies that multiple goroutines calling AddSchema
// simultaneously do not race on the GQL schema manager pointer swap (issue #4545).
// Run with: go test -race ./internal/db/ -run TestAddSchema_Concurrent_NoRace
func TestAddSchema_Concurrent_NoRace(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// start barrier so all goroutines attempt AddSchema as close to simultaneously as possible
	start := make(chan struct{})

	for i := range goroutines {
		i := i
		go func() {
			defer wg.Done()
			<-start
			schema := fmt.Sprintf(`
				type ConcurrentType%d {
					name: String
				}
			`, i)
			// Errors are acceptable (e.g. duplicate names on retry), but panics and
			// data races are not.
			_, _ = db.AddSchema(ctx, schema)
		}()
	}

	close(start)
	wg.Wait()
}

// TestAddSchema_Concurrent_WithConcurrentReads verifies that concurrent reads (ExecRequest)
// do not race with concurrent schema mutations.
// Run with: go test -race ./internal/db/ -run TestAddSchema_Concurrent_WithConcurrentReads
func TestAddSchema_Concurrent_WithConcurrentReads(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Seed an initial schema so ExecRequest has something to validate against.
	_, err = db.AddSchema(ctx, `type Book { title: String }`)
	require.NoError(t, err)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	start := make(chan struct{})

	// writer goroutines: mutate the schema
	for i := range goroutines {
		i := i
		go func() {
			defer wg.Done()
			<-start
			schema := fmt.Sprintf(`
				type ReadRaceType%d {
					value: Int
				}
			`, i)
			_, _ = db.AddSchema(ctx, schema)
		}()
	}

	// reader goroutines: execute queries concurrently with schema mutations
	for range goroutines {
		go func() {
			defer wg.Done()
			<-start
			// A simple introspection query exercises the parser's schemaManager read path.
			res := db.ExecRequest(ctx, `{ __typename }`)
			// Errors are expected (schema may not yet be ready), but data races must not occur.
			_ = res
		}()
	}

	close(start)
	wg.Wait()
}
