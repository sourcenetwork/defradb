// Copyright 2022 Democratized Data Foundation
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

	"github.com/stretchr/testify/assert"
)

func TestGetCollectionByNameReturnsErrorGivenNonExistantCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = db.GetCollectionByName(ctx, "doesNotExist")
	assert.EqualError(t, err, "datastore: key not found")
}

func TestGetCollectionByNameReturnsErrorGivenEmptyString(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = db.GetCollectionByName(ctx, "")
	assert.EqualError(t, err, "collection name can't be empty")
}
