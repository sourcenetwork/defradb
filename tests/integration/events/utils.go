// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

import (
	"context"
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type TestCase struct {
	Description string

	// docs is a map from Collection name, to a list
	// of docs in stringified JSON format
	Docs map[string][]string

	// The collection calls to make after subscribing to the events channel.
	//
	// Any errors generated within these calls are of no interest to this test
	// framework and should be handled as desired by the caller.
	CollectionCalls map[string][]func(client.Collection)

	// The daabase calls to make after subscribing to the events channel.
	//
	// Any errors generated within these calls are of no interest to this test
	// framework and should be handled as desired by the caller.
	DatabaseCalls []func(context.Context, client.DB)

	// The update events expected during the specified calls. The length will
	// be asserted (within timeout).
	ExpectedUpdates []ExpectedUpdate
}

// A struct holding properties that can be asserted upon.
//
// Properties with a `None` value will not be asserted. If all properties
// are `None` the Update event will still be expected and will contribute
// to the asserted count.
type ExpectedUpdate struct {
	DocKey immutable.Option[string]
	// The expected Cid, as a string (results in much more readable errors)
	Cid      immutable.Option[string]
	SchemaID immutable.Option[string]
	Priority immutable.Option[uint64]
}

type dbInfo interface {
	DB() client.DB
}

const eventTimeout = 100 * time.Millisecond

func ExecuteQueryTestCase(
	t *testing.T,
	schema string,
	testCase TestCase,
) {
	var err error
	ctx := context.Background()

	var dbi dbInfo
	dbi, err = testUtils.NewBadgerMemoryDB(ctx, db.WithUpdateEvents())
	if err != nil {
		t.Fatal(err)
	}

	db := dbi.DB()

	err = db.AddSchema(ctx, schema)
	if err != nil {
		t.Fatal(err)
	}

	setupDatabase(ctx, t, db, testCase)

	testRoutineClosedChan := make(chan struct{})
	closeTestRoutineChan := make(chan struct{})
	eventsChan, err := db.Events().Updates.Value().Subscribe()
	if err != nil {
		t.Fatal(err)
	}

	indexOfNextExpectedUpdate := 0
	go func() {
		for {
			select {
			case update := <-eventsChan:
				if indexOfNextExpectedUpdate >= len(testCase.ExpectedUpdates) {
					assert.Fail(t, "More events recieved than were expected", update)
					testRoutineClosedChan <- struct{}{}
					return
				}

				expectedEvent := testCase.ExpectedUpdates[indexOfNextExpectedUpdate]
				assertIfExpected(t, expectedEvent.Cid, update.Cid.String())
				assertIfExpected(t, expectedEvent.DocKey, update.DocKey)
				assertIfExpected(t, expectedEvent.Priority, update.Priority)
				assertIfExpected(t, expectedEvent.SchemaID, update.SchemaID)

				indexOfNextExpectedUpdate++
			case <-closeTestRoutineChan:
				return
			}
		}
	}()

	for collectionName, collectionCallSet := range testCase.CollectionCalls {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if err != nil {
			t.Fatal(err)
		}

		for _, collectionCall := range collectionCallSet {
			collectionCall(col)
		}
	}

	for _, databaseCall := range testCase.DatabaseCalls {
		databaseCall(ctx, db)
	}

	select {
	case <-time.After(eventTimeout):
		// Trigger an exit from the go routine monitoring the eventsChan.
		// As well as being a bit cleaner it stops the `--race` flag from
		// (rightly) seeing the assert of indexOfNextExpectedUpdate as a
		// data race.
		closeTestRoutineChan <- struct{}{}
	case <-testRoutineClosedChan:
		// no-op - just allow the host func to continue
	}

	// This is expressed verbosely, as `len(testCase.ExpectedUpdates) == indexOfNextExpectedUpdate`
	// is less easy to understand than the below
	indexOfLastExpectedUpdate := len(testCase.ExpectedUpdates) - 1
	indexOfLastAssertedUpdate := indexOfNextExpectedUpdate - 1
	assert.Equal(t, indexOfLastExpectedUpdate, indexOfLastAssertedUpdate)
}

func setupDatabase(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	testCase TestCase,
) {
	for collectionName, docs := range testCase.Docs {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if err != nil {
			t.Fatal(err)
		}

		for _, docStr := range docs {
			doc, err := client.NewDocFromJSON([]byte(docStr))
			if err != nil {
				t.Fatal(err)
			}
			err = col.Save(ctx, doc)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

// assertIfExpected asserts that the given values are Equal, if the expected parameter
// has a value.  Otherwise this function will do nothing.
func assertIfExpected[T any](t *testing.T, expected immutable.Option[T], actual T) {
	if expected.HasValue() {
		assert.Equal(t, expected.Value(), actual)
	}
}
