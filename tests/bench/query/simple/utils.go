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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	benchutils "github.com/sourcenetwork/defradb/tests/bench"
	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

var (
// log = logging.MustNewLogger("bench")
)

func RunQueryBenchGet(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	docCount int,
	query string,
	doSync bool,
) error {
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	listOfDocIDs, err := benchutils.BackfillBenchmarkDB(
		b,
		ctx,
		collections,
		fixture,
		docCount,
		0,
		doSync,
	)
	if err != nil {
		return err
	}

	return runQueryBenchGetSync(b, ctx, db, docCount, listOfDocIDs, query)
}

func runQueryBenchGetSync(
	b *testing.B,
	ctx context.Context,
	db client.DB,
	docCount int,
	listOfDocIDs [][]client.DocID,
	query string,
) error {
	// run any preprocessing on the query before execution (mostly just docID insertion if needed)
	query = formatQuery(b, query, listOfDocIDs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := db.ExecRequest(ctx, acpIdentity.NoIdentity, query)
		if len(res.GQL.Errors) > 0 {
			return errors.New(fmt.Sprintf("Query error: %v", res.GQL.Errors))
		}

		// leave comments for debug!!
		// l := len(res.Data.([]map[string]any))
		// if l != opCount {
		// 	return errors.Wrap(
		// "Invalid response, returned data doesn't match length, expected %v actual %v",
		// docCount, l)
		// }
		// log.Info(ctx, "", logging.NewKV("Response", res))
	}
	b.StopTimer()

	return nil
}

func formatQuery(b *testing.B, query string, listOfDocIDs [][]client.DocID) string {
	numPlaceholders := strings.Count(query, "{{docID}}")
	if numPlaceholders == 0 {
		return query
	}
	// create a copy of docIDs since we'll be mutating it
	docIDsCopy := listOfDocIDs[:]

	// b.Logf("formatting query, replacing %v instances", numPlaceholders)
	// b.Logf("Query before: %s", query)

	if len(docIDsCopy) < numPlaceholders {
		b.Fatalf(
			"Invalid number of query placeholders, max is %v requested is %v",
			len(listOfDocIDs),
			numPlaceholders,
		)
	}

	for i := 0; i < numPlaceholders; i++ {
		// pick a random docID, needs to be unique accross all
		// loop iterations, so remove the selected one so the next
		// iteration cant potentially pick it.
		rIndex := rand.Intn(len(docIDsCopy))
		docID := docIDsCopy[rIndex][0]

		// remove selected docID
		docIDsCopy = append(docIDsCopy[:rIndex], docIDsCopy[rIndex+1:]...)

		// replace
		query = strings.Replace(query, "{{docID}}", docID.String(), 1)
	}

	// b.Logf("Query After: %s", query)
	return query
}
