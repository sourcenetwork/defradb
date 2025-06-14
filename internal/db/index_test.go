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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

const (
	usersColName    = "Users"
	productsColName = "Products"

	usersNameFieldName    = "name"
	usersAgeFieldName     = "age"
	usersWeightFieldName  = "weight"
	usersNumbersFieldName = "numbers"
	usersHobbiesFieldName = "hobbies"
	usersCustomFieldName  = "custom"

	testUsersColIndexName    = "user_name_index"
	testUsersColIndexAge     = "user_age_index"
	testUsersColIndexWeight  = "user_weight_index"
	testUsersColIndexNumbers = "user_numbers_index"
	testUsersColIndexCustom  = "user_custom_index"
)

type indexTestFixture struct {
	ctx   context.Context
	db    *DB
	txn   datastore.Txn
	users client.Collection
	t     *testing.T
}

func (f *indexTestFixture) addUsersCollection() client.Collection {
	if f.users != nil {
		return f.users
	}

	_, err := f.db.AddSchema(
		f.ctx,
		fmt.Sprintf(
			`type %s {
				%s: String
				%s: Int
				%s: Float
				%s: [Int!]
				%s: [String!]
				%s: JSON
			}`,
			usersColName,
			usersNameFieldName,
			usersAgeFieldName,
			usersWeightFieldName,
			usersNumbersFieldName,
			usersHobbiesFieldName,
			usersCustomFieldName,
		),
	)
	require.NoError(f.t, err)

	col, err := f.db.GetCollectionByName(f.ctx, usersColName)
	require.NoError(f.t, err)

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	return col
}

func newIndexTestFixtureBare(t *testing.T) *indexTestFixture {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	return &indexTestFixture{
		ctx: ctx,
		db:  db,
		txn: txn,
		t:   t,
	}
}

func newIndexTestFixture(t *testing.T) *indexTestFixture {
	f := newIndexTestFixtureBare(t)
	f.users = f.addUsersCollection()
	return f
}

func (f *indexTestFixture) createCollectionIndex(
	desc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	return f.createCollectionIndexFor(f.users, desc)
}

func getUsersIndexDescOnName() client.IndexCreateRequest {
	return client.IndexCreateRequest{
		Name: testUsersColIndexName,
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName},
		},
	}
}

func getUsersIndexDescOnAge() client.IndexCreateRequest {
	return client.IndexCreateRequest{
		Name: testUsersColIndexAge,
		Fields: []client.IndexedFieldDescription{
			{Name: usersAgeFieldName},
		},
	}
}

func (f *indexTestFixture) createUserCollectionIndexOnName() client.IndexDescription {
	newIndexDesc, err := f.createCollectionIndexFor(f.users, getUsersIndexDescOnName())
	require.NoError(f.t, err)
	return newIndexDesc
}

func (f *indexTestFixture) createUserCollectionIndexOnAge() client.IndexDescription {
	newDesc, err := f.createCollectionIndexFor(f.users, getUsersIndexDescOnAge())
	require.NoError(f.t, err)
	return newDesc
}

func (f *indexTestFixture) commitTxn() {
	err := f.txn.Commit(f.ctx)
	require.NoError(f.t, err)
	txn, err := f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
	f.txn = txn
}

func (f *indexTestFixture) createCollectionIndexFor(
	col client.Collection,
	desc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	ctx := InitContext(f.ctx, f.txn)
	index, err := col.CreateIndex(ctx, desc)
	if err == nil {
		f.commitTxn()
	}
	return index, err
}

func TestCreateIndex_IfFieldsIsEmpty_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	_, err := f.createCollectionIndex(client.IndexCreateRequest{
		Name: "some_index_name",
	})
	assert.EqualError(t, err, errIndexMissingFields)
}

func TestCreateIndex_IfValidInput_CreateIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	desc := client.IndexCreateRequest{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName},
		},
	}
	resultDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, desc.Name, resultDesc.Name)
	assert.Equal(t, desc.Fields, resultDesc.Fields)
	assert.Equal(t, desc.Unique, resultDesc.Unique)
}

func TestCreateIndex_IfFieldNameIsEmpty_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	desc := client.IndexCreateRequest{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: ""},
		},
	}
	_, err := f.createCollectionIndex(desc)
	assert.EqualError(t, err, errIndexFieldMissingName)
}

func TestCreateIndex_IfFieldHasNoDirection_DefaultToAsc(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	desc := client.IndexCreateRequest{
		Name:   "some_index_name",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.False(t, newDesc.Fields[0].Descending)
}

func TestCreateIndex_IfIndexWithNameAlreadyExists_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	name := "some_index_name"
	desc1 := client.IndexCreateRequest{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	desc2 := client.IndexCreateRequest{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersAgeFieldName}},
	}
	_, err := f.createCollectionIndex(desc1)
	assert.NoError(t, err)
	_, err = f.createCollectionIndex(desc2)
	assert.ErrorIs(t, err, NewErrIndexWithNameAlreadyExists(name))
}

func TestCreateIndex_IfGeneratedNameMatchesExisting_AddIncrement(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	name := usersColName + "_" + usersAgeFieldName + "_ASC"
	desc1 := client.IndexCreateRequest{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	desc2 := client.IndexCreateRequest{
		Name:   name + "_2",
		Fields: []client.IndexedFieldDescription{{Name: usersWeightFieldName}},
	}
	desc3 := client.IndexCreateRequest{
		Name:   "",
		Fields: []client.IndexedFieldDescription{{Name: usersAgeFieldName}},
	}
	_, err := f.createCollectionIndex(desc1)
	assert.NoError(t, err)
	_, err = f.createCollectionIndex(desc2)
	assert.NoError(t, err)
	newDesc3, err := f.createCollectionIndex(desc3)
	assert.NoError(t, err)
	assert.Equal(t, name+"_3", newDesc3.Name)
}

func TestCreateIndex_IfPropertyDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	const field = "non_existing_field"
	desc := client.IndexCreateRequest{
		Fields: []client.IndexedFieldDescription{{Name: field}},
	}

	_, err := f.createCollectionIndex(desc)
	assert.ErrorIs(t, err, NewErrNonExistingFieldForIndex(field))
}

func TestCreateIndex_IfProvideInvalidIndexName_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexDesc := getUsersIndexDescOnName()
	indexDesc.Name = "!"
	_, err := f.users.CreateIndex(f.ctx, indexDesc)
	require.ErrorIs(t, err, schema.NewErrIndexWithInvalidName(indexDesc.Name))
}

func TestCreateIndex_ShouldUpdateCollectionsDescription(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indOnName, err := f.users.CreateIndex(f.ctx, getUsersIndexDescOnName())
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{indOnName}, f.users.Version().Indexes)

	indOnAge, err := f.users.CreateIndex(f.ctx, getUsersIndexDescOnAge())
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{indOnName, indOnAge},
		f.users.Version().Indexes)
}

func TestCollectionGetIndexes_ShouldReturnIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	indexes, err := f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)

	require.Equal(t, 1, len(indexes))
	assert.Equal(t, testUsersColIndexName, indexes[0].Name)
}

func TestCollectionGetIndexes_IfInvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	indexes, err := f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	require.Len(t, indexes, 2)
	require.ElementsMatch(t,
		[]string{testUsersColIndexName, testUsersColIndexAge},
		[]string{indexes[0].Name, indexes[1].Name},
	)
	require.ElementsMatch(t, []uint32{1, 2}, []uint32{indexes[0].ID, indexes[1].ID})
}

func TestCollectionGetIndexes_IfIndexIsCreated_ReturnUpdateIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	indexes, err := f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, indexes, 1)

	_, err = f.users.CreateIndex(f.ctx, getUsersIndexDescOnAge())
	assert.NoError(t, err)

	indexes, err = f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, indexes, 2)
}

func TestCollectionGetIndexes_IfIndexIsDropped_ReturnUpdateIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	indexes, err := f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, indexes, 2)

	err = f.users.DropIndex(f.ctx, testUsersColIndexName)
	assert.NoError(t, err)

	indexes, err = f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	assert.Equal(t, indexes[0].Name, testUsersColIndexAge)

	err = f.users.DropIndex(f.ctx, testUsersColIndexAge)
	assert.NoError(t, err)

	indexes, err = f.users.GetIndexes(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, indexes, 0)
}

func TestCollectionGetIndexes_ShouldReturnIndexesInOrderedByName(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	const (
		num             = 30
		fieldNamePrefix = "field_"
		indexNamePrefix = "index_"
	)

	toSuffix := func(i int) string {
		return fmt.Sprintf("%02d", i)
	}

	builder := strings.Builder{}
	builder.WriteString("type testCollection {\n")

	for i := 1; i <= num; i++ {
		_, err := builder.WriteString(fieldNamePrefix)
		require.NoError(f.t, err)

		_, err = builder.WriteString(toSuffix(i))
		require.NoError(f.t, err)

		_, err = builder.WriteString(": String\n")
		require.NoError(f.t, err)
	}
	_, err := builder.WriteString("}")
	require.NoError(f.t, err)

	_, err = f.db.AddSchema(
		f.ctx,
		builder.String(),
	)
	require.NoError(f.t, err)

	collection, err := f.db.GetCollectionByName(f.ctx, "testCollection")
	require.NoError(f.t, err)

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
	for i := 1; i <= num; i++ {
		iStr := toSuffix(i)
		indexDesc := client.IndexCreateRequest{
			Name: indexNamePrefix + iStr,
			Fields: []client.IndexedFieldDescription{
				{Name: fieldNamePrefix + iStr},
			},
		}

		_, err := f.createCollectionIndexFor(collection, indexDesc)
		require.NoError(t, err)
	}

	indexes, err := collection.GetIndexes(f.ctx)
	require.NoError(t, err)
	require.Len(t, indexes, num)

	for i := 1; i <= num; i++ {
		assert.Equal(t, indexNamePrefix+toSuffix(i), indexes[i-1].Name, "i = %d", i)
	}
}

func TestDropIndex_ShouldUpdateCollectionsDescription(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	ctx := InitContext(f.ctx, f.txn)
	_, err := f.users.CreateIndex(ctx, getUsersIndexDescOnName())
	require.NoError(t, err)
	indOnAge, err := f.users.CreateIndex(ctx, getUsersIndexDescOnAge())
	require.NoError(t, err)
	f.commitTxn()

	err = f.users.DropIndex(f.ctx, testUsersColIndexName)
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{indOnAge},
		f.users.Version().Indexes)

	err = f.users.DropIndex(f.ctx, testUsersColIndexAge)
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{}, f.users.Version().Indexes)
}

func TestDropIndex_IfIndexWithNameDoesNotExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	const name = "not_existing_index"
	err := f.users.DropIndex(f.ctx, name)
	require.ErrorIs(t, err, NewErrIndexWithNameDoesNotExists(name))
}

func TestNewCollectionIndex_IfDescriptionHasNoFields_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := getUsersIndexDescOnName()
	desc.Fields = nil
	descWithID := client.IndexDescription{
		Name:   desc.Name,
		ID:     1,
		Fields: desc.Fields,
		Unique: desc.Unique,
	}
	_, err := NewCollectionIndex(f.users, descWithID)
	require.ErrorIs(t, err, NewErrIndexDescHasNoFields(descWithID))
}

func TestNewCollectionIndex_IfDescriptionHasNonExistingField_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := getUsersIndexDescOnName()
	desc.Fields[0].Name = "non_existing_field"
	descWithID := client.IndexDescription{
		Name:   desc.Name,
		ID:     1,
		Fields: desc.Fields,
		Unique: desc.Unique,
	}
	_, err := NewCollectionIndex(f.users, descWithID)
	require.ErrorIs(t, err, client.NewErrFieldNotExist(desc.Fields[0].Name))
}
