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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

func TestWrapCollectionIndex_WithEmptyKind_ReturnsSimpleIndex(t *testing.T) {
	index, err := wrapCollectionIndex(collectionBaseIndex{desc: client.IndexDescription{}})

	require.NoError(t, err)
	assert.IsType(t, &collectionSimpleIndex{}, index)
}

func TestWrapCollectionIndex_WithEmptyKindAndUnique_ReturnsUniqueIndex(t *testing.T) {
	index, err := wrapCollectionIndex(collectionBaseIndex{desc: client.IndexDescription{Unique: true}})

	require.NoError(t, err)
	assert.IsType(t, &collectionUniqueIndex{}, index)
}

func TestWrapCollectionIndex_WithUnregisteredKind_ReturnsError(t *testing.T) {
	index, err := wrapCollectionIndex(collectionBaseIndex{desc: client.IndexDescription{Kind: client.IndexKindBM25}})

	assert.Nil(t, index)
	require.ErrorIs(t, err, ErrUnknownIndexKind)
}

func TestValidateIndexKind_WithUnregisteredKind_ReturnsError(t *testing.T) {
	require.NoError(t, validateIndexKind(client.NewIndexRequest{}))
	require.ErrorIs(t,
		validateIndexKind(client.NewIndexRequest{Kind: client.IndexKindTrigram}),
		ErrUnknownIndexKind,
	)
}

// An index description written before the Kind and Options fields existed has neither key, and
// must still resolve to the ordered key index without any migration.
func TestWrapCollectionIndex_WithDescriptionWrittenBeforeKindExisted_ReturnsSimpleIndex(t *testing.T) {
	storedJSON := `{"Name":"User_name_ASC","ID":1,"Fields":[{"Name":"name","Descending":false}],"Unique":false}`

	var desc client.IndexDescription
	require.NoError(t, json.Unmarshal([]byte(storedJSON), &desc))
	assert.Equal(t, "", desc.Kind)
	assert.Nil(t, desc.Options)

	index, err := wrapCollectionIndex(collectionBaseIndex{desc: desc})

	require.NoError(t, err)
	assert.IsType(t, &collectionSimpleIndex{}, index)
}

// Collection versions are shared between peers, so a version can name an index kind this build
// does not implement. Opening the collection must record that index as failed and carry on.
func TestNewCollection_WithIndexKindThisBuildDoesNotImplement_MarksIndexFailed(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	// Create the index under a kind registered here, then unregister the kind. What is left is a
	// stored collection version naming a kind this build does not implement.
	const kind = "test_only_kind"
	registerIndexKind(kind, func(base collectionBaseIndex) client.CollectionIndex {
		return &collectionSimpleIndex{collectionBaseIndex: base}
	})
	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Kind:   kind,
	})
	require.NoError(t, err)
	delete(indexKinds, kind)

	reopened, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	assert.Empty(t, reopened.(*collection).indexes)

	state := readIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
	assert.True(t, state.isFailed())
	assert.Contains(t, state.Reason, errUnknownIndexKind)
}

func TestNewErrUnknownIndexKind_MatchesSentinelAndNamesTheKind(t *testing.T) {
	err := NewErrUnknownIndexKind(client.IndexKindBM25)

	assert.True(t, errors.Is(err, ErrUnknownIndexKind))
	assert.Contains(t, err.Error(), client.IndexKindBM25)
}
