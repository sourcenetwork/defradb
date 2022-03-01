// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document/key"
)

var (
//log = logging.MustNewLogger("defra.bench")
)

func runQueryBenchGet(b *testing.B, ctx context.Context, fixture fixtures.Generator, docCount int, query string, doSync bool) error {
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close(ctx)

	dockeys, err := benchutils.BackfillBenchmarkDB(b, ctx, collections, fixture, docCount, 0, doSync)
	if err != nil {
		return err
	}

	return runQueryBenchGetSync(b, ctx, db, docCount, dockeys, query)
}

func runQueryBenchGetSync(
	b *testing.B,
	ctx context.Context,
	db *db.DB,
	docCount int,
	dockeys [][]key.DocKey,
	query string,
) error {
	// run any preprocessing on the query before execution (mostly just dockey insertion if needed)
	query = formatQuery(b, query, dockeys)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := db.ExecQuery(ctx, query)
		if len(res.Errors) > 0 {
			return fmt.Errorf("Query error: %v", res.Errors)
		}

		// leave comments for debug!!
		// l := len(res.Data.([]map[string]interface{}))
		// if l != opCount {
		// 	return fmt.Errorf("Invalid response, returned data doesn't match length, expected %v actual %v", docCount, l)
		// }
		// log.Info(ctx, "", logging.NewKV("Response", res))
	}
	b.StopTimer()

	return nil
}

func formatQuery(b *testing.B, query string, dockeys [][]key.DocKey) string {
	numPlaceholders := strings.Count(query, "{{dockey}}")
	if numPlaceholders == 0 {
		return query
	}
	// create a copy of dockeys since we'll be mutating it
	dockeysCopy := dockeys[:]

	// b.Logf("formatting query, replacing %v instances", numPlaceholders)
	// b.Logf("Query before: %s", query)

	if len(dockeysCopy) < numPlaceholders {
		b.Fatalf("Invalid number of query placeholders, max is %v requested is %v", len(dockeys), numPlaceholders)
	}

	for i := 0; i < numPlaceholders; i++ {
		// pick a random dockey, needs to be unique accross all
		// loop iterations, so remove the selected one so the next
		// iteration cant potentially pick it.
		rIndex := rand.Intn(len(dockeysCopy))
		key := dockeysCopy[rIndex][0]

		// remove selected key
		dockeysCopy = append(dockeysCopy[:rIndex], dockeysCopy[rIndex+1:]...)

		// replace
		query = strings.Replace(query, "{{dockey}}", key.String(), 1)
	}

	// b.Logf("Query After: %s", query)
	return query
}
