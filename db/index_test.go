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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

const (
	usersColName    = "Users"
	productsColName = "Products"

	usersNameFieldName   = "name"
	usersAgeFieldName    = "age"
	usersWeightFieldName = "weight"

	productsIDFieldName        = "id"
	productsPriceFieldName     = "price"
	productsCategoryFieldName  = "category"
	productsAvailableFieldName = "available"

	testUsersColIndexName = "user_name"
	testUsersColIndexAge  = "user_age"
)

type indexTestFixture struct {
	ctx      context.Context
	db       *implicitTxnDB
	txn      datastore.Txn
	users    client.Collection
	products client.Collection
	t        *testing.T
}

func getUsersCollectionDesc() client.CollectionDescription {
	return client.CollectionDescription{
		Name: usersColName,
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: usersNameFieldName,
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: usersAgeFieldName,
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: usersWeightFieldName,
					Kind: client.FieldKind_FLOAT,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}
}

func getProductsCollectionDesc() client.CollectionDescription {
	return client.CollectionDescription{
		Name: productsColName,
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: productsIDFieldName,
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: productsPriceFieldName,
					Kind: client.FieldKind_FLOAT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: productsCategoryFieldName,
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: productsAvailableFieldName,
					Kind: client.FieldKind_BOOL,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}
}

func newIndexTestFixtureBare(t *testing.T) *indexTestFixture {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
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
	f.users = f.createCollection(getUsersCollectionDesc())
	return f
}

func (f *indexTestFixture) createCollectionIndex(
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	return f.createCollectionIndexFor(f.users.Name(), desc)
}

func getUsersIndexDescOnName() client.IndexDescription {
	return client.IndexDescription{
		Name: testUsersColIndexName,
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Ascending},
		},
	}
}

func getUsersIndexDescOnAge() client.IndexDescription {
	return client.IndexDescription{
		Name: testUsersColIndexAge,
		Fields: []client.IndexedFieldDescription{
			{Name: usersAgeFieldName, Direction: client.Ascending},
		},
	}
}

func getProductsIndexDescOnCategory() client.IndexDescription {
	return client.IndexDescription{
		Name: testUsersColIndexAge,
		Fields: []client.IndexedFieldDescription{
			{Name: productsCategoryFieldName, Direction: client.Ascending},
		},
	}
}

func (f *indexTestFixture) createUserCollectionIndexOnName() client.IndexDescription {
	newDesc, err := f.createCollectionIndexFor(f.users.Name(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	f.commitTxn()
	return newDesc
}

func (f *indexTestFixture) createUserCollectionIndexOnAge() client.IndexDescription {
	newDesc, err := f.createCollectionIndexFor(f.users.Name(), getUsersIndexDescOnAge())
	require.NoError(f.t, err)
	f.commitTxn()
	return newDesc
}

func (f *indexTestFixture) dropIndex(colName, indexName string) error {
	return f.db.dropCollectionIndex(f.ctx, f.txn, colName, indexName)
}

func (f *indexTestFixture) dropAllIndexes(colName string) error {
	col := (f.users.WithTxn(f.txn)).(*collection)
	return col.dropAllIndexes(f.ctx)
}

func (f *indexTestFixture) countIndexPrefixes(colName, indexName string) int {
	prefix := core.NewCollectionIndexKey(usersColName, indexName)
	q, err := f.txn.Systemstore().Query(f.ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	assert.NoError(f.t, err)
	defer func() {
		err := q.Close()
		assert.NoError(f.t, err)
	}()

	count := 0
	for res := range q.Next() {
		if res.Error != nil {
			assert.NoError(f.t, err)
		}
		count++
	}
	return count
}

func (f *indexTestFixture) commitTxn() {
	err := f.txn.Commit(f.ctx)
	require.NoError(f.t, err)
	txn, err := f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
	f.txn = txn
}

func (f *indexTestFixture) createCollectionIndexFor(
	collectionName string,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	newDesc, err := f.db.createCollectionIndex(f.ctx, f.txn, collectionName, desc)
	//if err != nil {
	//return newDesc, err
	//}
	//f.txn, err = f.db.NewTxn(f.ctx, false)
	//assert.NoError(f.t, err)
	return newDesc, err
}

func (f *indexTestFixture) getAllIndexes() ([]client.CollectionIndexDescription, error) {
	return f.db.getAllCollectionIndexes(f.ctx, f.txn)
}

func (f *indexTestFixture) getCollectionIndexes(colName string) ([]client.IndexDescription, error) {
	return f.db.getCollectionIndexes(f.ctx, f.txn, colName)
}

func (f *indexTestFixture) createCollection(
	desc client.CollectionDescription,
) client.Collection {
	col, err := f.db.createCollection(f.ctx, f.txn, desc)
	assert.NoError(f.t, err)
	err = f.txn.Commit(f.ctx)
	assert.NoError(f.t, err)
	f.txn, err = f.db.NewTxn(f.ctx, false)
	assert.NoError(f.t, err)
	return col
}

func TestCreateIndex_IfFieldsIsEmpty_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	_, err := f.createCollectionIndex(client.IndexDescription{
		Name: "some_index_name",
	})
	assert.EqualError(t, err, errIndexMissingFields)
}

func TestCreateIndex_IfIndexDescriptionIDIsNotZero_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	for _, id := range []uint32{1, 20, 999} {
		desc := client.IndexDescription{
			Name: "some_index_name",
			ID:   id,
			Fields: []client.IndexedFieldDescription{
				{Name: usersNameFieldName, Direction: client.Ascending},
			},
		}
		_, err := f.createCollectionIndex(desc)
		assert.ErrorIs(t, err, NewErrNonZeroIndexIDProvided(0))
	}
}

func TestCreateIndex_IfValidInput_CreateIndex(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Ascending},
		},
	}
	resultDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, desc.Name, resultDesc.Name)
	assert.Equal(t, desc, resultDesc)
}

func TestCreateIndex_IfFieldNameIsEmpty_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: "", Direction: client.Ascending},
		},
	}
	_, err := f.createCollectionIndex(desc)
	assert.EqualError(t, err, errIndexFieldMissingName)
}

func TestCreateIndex_IfFieldHasNoDirection_DefaultToAsc(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Name:   "some_index_name",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, client.Ascending, newDesc.Fields[0].Direction)
}

func TestCreateIndex_IfNameIsNotSpecified_GenerateWithLowerCase(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	colDesc := getUsersCollectionDesc()
	const colName = "UsErS"
	const fieldName = "NaMe"
	colDesc.Name = colName
	colDesc.Schema.Name = colName // Which one should we use?
	colDesc.Schema.Fields[1].Name = fieldName
	f.users = f.createCollection(colDesc)

	desc := client.IndexDescription{
		Name: "",
		Fields: []client.IndexedFieldDescription{
			{Name: fieldName, Direction: client.Ascending},
		},
	}

	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, colName+"_"+fieldName+"_ASC", newDesc.Name)
}

func TestCreateIndex_IfSingleFieldInDescOrder_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Descending},
		},
	}
	_, err := f.createCollectionIndex(desc)
	assert.EqualError(t, err, errIndexSingleFieldWrongDirection)
}

func TestCreateIndex_IfIndexWithNameAlreadyExists_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "some_index_name"
	desc1 := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	desc2 := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersAgeFieldName}},
	}
	_, err := f.createCollectionIndex(desc1)
	assert.NoError(t, err)
	_, err = f.createCollectionIndex(desc2)
	assert.EqualError(t, err, errIndexWithNameAlreadyExists)
}

func TestCreateIndex_IfGeneratedNameMatchesExisting_AddIncrement(t *testing.T) {
	f := newIndexTestFixture(t)

	name := usersColName + "_" + usersAgeFieldName + "_ASC"
	desc1 := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	desc2 := client.IndexDescription{
		Name:   name + "_2",
		Fields: []client.IndexedFieldDescription{{Name: usersWeightFieldName}},
	}
	desc3 := client.IndexDescription{
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

func TestCreateIndex_ShouldSaveToSystemStorage(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "users_age_ASC"
	desc := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)

	key := core.NewCollectionIndexKey(f.users.Name(), name)
	data, err := f.txn.Systemstore().Get(f.ctx, key.ToDS())
	assert.NoError(t, err)
	var deserialized client.IndexDescription
	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)
	assert.Equal(t, desc, deserialized)
}

func TestCreateIndex_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "users_age_ASC"
	desc := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}

	f.db.Close(f.ctx)

	_, err := f.createCollectionIndex(desc)
	assert.Error(t, err)
}

func TestCreateIndex_IfCollectionDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}

	_, err := f.createCollectionIndexFor(productsColName, desc)
	assert.ErrorIs(t, err, NewErrCollectionDoesntExist(usersColName))
}

func TestCreateIndex_IfPropertyDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	const field = "non_existing_field"
	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: field}},
	}

	_, err := f.createCollectionIndex(desc)
	assert.ErrorIs(t, err, NewErrNonExistingFieldForIndex(field))
}

func TestGetIndexes_ShouldReturnListOfAllExistingIndexes(t *testing.T) {
	f := newIndexTestFixture(t)

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.createCollection(getProductsCollectionDesc())
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}
	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	indexes, err := f.getAllIndexes()
	assert.NoError(t, err)

	require.Equal(t, 2, len(indexes))
	usersIndexIndex := 0
	if indexes[0].CollectionName != usersColName {
		usersIndexIndex = 1
	}
	assert.Equal(t, usersIndexDesc, indexes[usersIndexIndex].Index)
	assert.Equal(t, usersColName, indexes[usersIndexIndex].CollectionName)
	assert.Equal(t, productsIndexDesc, indexes[1-usersIndexIndex].Index)
	assert.Equal(t, productsColName, indexes[1-usersIndexIndex].CollectionName)
}

func TestGetIndexes_IfInvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	indexKey := core.NewCollectionIndexKey(usersColName, "users_name_index")
	err := f.txn.Systemstore().Put(f.ctx, indexKey.ToDS(), []byte("invalid"))
	assert.NoError(t, err)

	_, err = f.getAllIndexes()
	assert.ErrorIs(t, err, NewErrInvalidStoredIndex(nil))
}

func TestGetIndexes_IfInvalidIndexKeyIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	indexKey := core.NewCollectionIndexKey(usersColName, "users_name_index")
	key := ds.NewKey(indexKey.ToString() + "/invalid")
	desc := client.IndexDescription{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Ascending},
		},
	}
	descData, _ := json.Marshal(desc)
	err := f.txn.Systemstore().Put(f.ctx, key, descData)
	assert.NoError(t, err)

	_, err = f.getAllIndexes()
	assert.ErrorIs(t, err, NewErrInvalidStoredIndexKey(key.String()))
}

func TestGetCollectionIndexes_ShouldReturnListOfCollectionIndexes(t *testing.T) {
	f := newIndexTestFixture(t)

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.createCollection(getProductsCollectionDesc())
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}
	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	userIndexes, err := f.getCollectionIndexes(usersColName)
	assert.NoError(t, err)
	require.Equal(t, 1, len(userIndexes))
	assert.Equal(t, usersIndexDesc, userIndexes[0])

	productIndexes, err := f.getCollectionIndexes(productsColName)
	assert.NoError(t, err)
	require.Equal(t, 1, len(productIndexes))
	assert.Equal(t, productsIndexDesc, productIndexes[0])
}

func TestGetCollectionIndexes_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	f.db.Close(f.ctx)

	_, err := f.getCollectionIndexes(usersColName)
	assert.Error(t, err)
}

func TestGetCollectionIndexes_IfInvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	indexKey := core.NewCollectionIndexKey(usersColName, "users_name_index")
	err := f.txn.Systemstore().Put(f.ctx, indexKey.ToDS(), []byte("invalid"))
	assert.NoError(t, err)

	_, err = f.getCollectionIndexes(usersColName)
	assert.ErrorIs(t, err, NewErrInvalidStoredIndex(nil))
}

func TestDropIndex_ShouldDeleteIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	desc := f.createUserCollectionIndexOnName()

	err := f.dropIndex(usersColName, desc.Name)
	assert.NoError(t, err)

	indexKey := core.NewCollectionIndexKey(usersColName, desc.Name)
	_, err = f.txn.Systemstore().Get(f.ctx, indexKey.ToDS())
	assert.Error(t, err)
}

func TestDropIndex_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	desc := f.createUserCollectionIndexOnName()

	f.db.Close(f.ctx)

	err := f.dropIndex(productsColName, desc.Name)
	assert.Error(t, err)
}

func TestDropIndex_IfCollectionDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	err := f.dropIndex(productsColName, "any_name")
	assert.ErrorIs(t, err, NewErrCollectionDoesntExist(usersColName))
}

func TestDropAllIndex_ShouldDeleteAllIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	_, err := f.createCollectionIndexFor(usersColName, client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Ascending},
		},
	})
	assert.NoError(f.t, err)

	_, err = f.createCollectionIndexFor(usersColName, client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: usersAgeFieldName, Direction: client.Ascending},
		},
	})
	assert.NoError(f.t, err)

	assert.Equal(t, 2, f.countIndexPrefixes(usersColName, ""))

	err = f.dropAllIndexes(usersColName)
	assert.NoError(t, err)

	assert.Equal(t, 0, f.countIndexPrefixes(usersColName, ""))
}

func TestDropAllIndexes_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	f.db.Close(f.ctx)

	err := f.dropAllIndexes(usersColName)
	assert.Error(t, err)
}

func TestCreateIndex_WithMultipleCollectionsAndIndexes_AssignIncrementedIDPerCollection(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.createCollection(getUsersCollectionDesc())
	products := f.createCollection(getProductsCollectionDesc())

	makeIndex := func(fieldName string) client.IndexDescription {
		return client.IndexDescription{
			Fields: []client.IndexedFieldDescription{
				{Name: fieldName, Direction: client.Ascending},
			},
		}
	}

	createIndexAndAssert := func(col client.Collection, fieldName string, expectedID uint32) {
		desc, err := f.createCollectionIndexFor(col.Name(), makeIndex(fieldName))
		require.NoError(t, err)
		assert.Equal(t, expectedID, desc.ID)
		seqKey := core.NewSequenceKey(fmt.Sprintf("%s/%d", core.COLLECTION_INDEX, col.ID()))
		storedSeqKey, err := f.txn.Systemstore().Get(f.ctx, seqKey.ToDS())
		assert.NoError(t, err)
		storedSeqVal := binary.BigEndian.Uint64(storedSeqKey)
		assert.Equal(t, expectedID, uint32(storedSeqVal))
	}

	createIndexAndAssert(users, usersNameFieldName, 1)
	createIndexAndAssert(users, usersAgeFieldName, 2)
	createIndexAndAssert(products, productsIDFieldName, 1)
	createIndexAndAssert(products, productsCategoryFieldName, 2)
}
