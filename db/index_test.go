// Copyright 2023 Democratized Data Foundation
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
	"strings"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/request/graphql/schema"
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

	testUsersColIndexName   = "user_name"
	testUsersColIndexAge    = "user_age"
	testUsersColIndexWeight = "user_weight"
)

type indexTestFixture struct {
	ctx   context.Context
	db    *implicitTxnDB
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
			}`,
			usersColName,
			usersNameFieldName,
			usersAgeFieldName,
			usersWeightFieldName,
		),
	)
	require.NoError(f.t, err)

	col, err := f.db.GetCollectionByName(f.ctx, usersColName)
	require.NoError(f.t, err)

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	return col
}

func (f *indexTestFixture) getProductsCollectionDesc() client.Collection {
	_, err := f.db.AddSchema(
		f.ctx,
		fmt.Sprintf(
			`type %s {
				%s: Int
				%s: Float
				%s: String
				%s: Boolean
			}`,
			productsColName,
			productsIDFieldName,
			productsPriceFieldName,
			productsCategoryFieldName,
			productsAvailableFieldName,
		),
	)
	require.NoError(f.t, err)

	col, err := f.db.GetCollectionByName(f.ctx, productsColName)
	require.NoError(f.t, err)

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	return col
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
	f.users = f.addUsersCollection()
	return f
}

func (f *indexTestFixture) createCollectionIndex(
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	return f.createCollectionIndexFor(f.users.Name().Value(), desc)
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

func getUsersIndexDescOnWeight() client.IndexDescription {
	return client.IndexDescription{
		Name: testUsersColIndexWeight,
		Fields: []client.IndexedFieldDescription{
			{Name: usersWeightFieldName, Direction: client.Ascending},
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
	newDesc, err := f.createCollectionIndexFor(f.users.Name().Value(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	return newDesc
}

func makeUnique(indexDesc client.IndexDescription) client.IndexDescription {
	indexDesc.Unique = true
	return indexDesc
}

func (f *indexTestFixture) createUserCollectionUniqueIndexOnName() client.IndexDescription {
	indexDesc := makeUnique(getUsersIndexDescOnName())
	newDesc, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)
	return newDesc
}

func addFieldToIndex(indexDesc client.IndexDescription, fieldName string) client.IndexDescription {
	indexDesc.Fields = append(indexDesc.Fields, client.IndexedFieldDescription{
		Name: fieldName, Direction: client.Ascending,
	})
	return indexDesc
}

func (f *indexTestFixture) createUserCollectionIndexOnNameAndAge() client.IndexDescription {
	indexDesc := addFieldToIndex(getUsersIndexDescOnName(), usersAgeFieldName)
	newDesc, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)
	return newDesc
}

func (f *indexTestFixture) createUserCollectionIndexOnAge() client.IndexDescription {
	newDesc, err := f.createCollectionIndexFor(f.users.Name().Value(), getUsersIndexDescOnAge())
	require.NoError(f.t, err)
	return newDesc
}

func (f *indexTestFixture) dropIndex(colName, indexName string) error {
	return f.db.dropCollectionIndex(f.ctx, f.txn, colName, indexName)
}

func (f *indexTestFixture) countIndexPrefixes(colName, indexName string) int {
	prefix := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), indexName)
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
	index, err := f.db.createCollectionIndex(f.ctx, f.txn, collectionName, desc)
	if err == nil {
		f.commitTxn()
	}
	return index, err
}

func (f *indexTestFixture) getAllIndexes() (map[client.CollectionName][]client.IndexDescription, error) {
	return f.db.getAllIndexes(f.ctx, f.txn)
}

func (f *indexTestFixture) getCollectionIndexes(colID uint32) ([]client.IndexDescription, error) {
	return f.db.fetchCollectionIndexDescriptions(f.ctx, f.txn, colID)
}

func TestCreateIndex_IfFieldsIsEmpty_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	_, err := f.createCollectionIndex(client.IndexDescription{
		Name: "some_index_name",
	})
	assert.EqualError(t, err, errIndexMissingFields)
}

func TestCreateIndex_IfIndexDescriptionIDIsNotZero_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

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
	defer f.db.Close()

	desc := client.IndexDescription{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: usersNameFieldName, Direction: client.Ascending},
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
	defer f.db.Close()

	desc := client.IndexDescription{
		Name:   "some_index_name",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, client.Ascending, newDesc.Fields[0].Direction)
}

func TestCreateIndex_IfSingleFieldInDescOrder_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

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
	defer f.db.Close()

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
	assert.ErrorIs(t, err, NewErrIndexWithNameAlreadyExists(name))
}

func TestCreateIndex_IfGeneratedNameMatchesExisting_AddIncrement(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

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
	defer f.db.Close()

	name := "users_age_ASC"
	desc := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)

	key := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), name)
	data, err := f.txn.Systemstore().Get(f.ctx, key.ToDS())
	assert.NoError(t, err)
	var deserialized client.IndexDescription
	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)
	desc.ID = 1
	assert.Equal(t, desc, deserialized)
}

func TestCreateIndex_IfCollectionDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}

	_, err := f.createCollectionIndexFor(productsColName, desc)
	assert.ErrorIs(t, err, NewErrCanNotReadCollection(usersColName, nil))
}

func TestCreateIndex_IfPropertyDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	const field = "non_existing_field"
	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: field}},
	}

	_, err := f.createCollectionIndex(desc)
	assert.ErrorIs(t, err, NewErrNonExistingFieldForIndex(field))
}

func TestCreateIndex_WithMultipleCollectionsAndIndexes_AssignIncrementedIDPerCollection(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.addUsersCollection()
	products := f.getProductsCollectionDesc()

	makeIndex := func(fieldName string) client.IndexDescription {
		return client.IndexDescription{
			Fields: []client.IndexedFieldDescription{
				{Name: fieldName, Direction: client.Ascending},
			},
		}
	}

	createIndexAndAssert := func(col client.Collection, fieldName string, expectedID uint32) {
		desc, err := f.createCollectionIndexFor(col.Name().Value(), makeIndex(fieldName))
		require.NoError(t, err)
		assert.Equal(t, expectedID, desc.ID)
		seqKey := core.NewIndexIDSequenceKey(col.ID())
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

func TestCreateIndex_IfFailsToCreateTxn_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	testErr := errors.New("test error")

	mockedRootStore := mocks.NewRootStore(t)
	mockedRootStore.On("Close").Return(nil)

	mockedRootStore.EXPECT().NewTransaction(mock.Anything, mock.Anything).Return(nil, testErr)
	f.db.rootstore = mockedRootStore

	_, err := f.users.CreateIndex(f.ctx, getUsersIndexDescOnName())
	require.ErrorIs(t, err, testErr)
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

	assert.ElementsMatch(t, []client.IndexDescription{indOnName}, f.users.Description().Indexes)

	indOnAge, err := f.users.CreateIndex(f.ctx, getUsersIndexDescOnAge())
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{indOnName, indOnAge},
		f.users.Description().Indexes)
}

func TestCreateIndex_IfAttemptToIndexOnUnsupportedType_ReturnError(t *testing.T) {
	f := newIndexTestFixtureBare(t)

	const unsupportedKind = client.FieldKind_BOOL_ARRAY

	_, err := f.db.AddSchema(
		f.ctx,
		`type testTypeCol {
			field: [Boolean!]
		}`,
	)
	require.NoError(f.t, err)

	collection, err := f.db.GetCollectionByName(f.ctx, "testTypeCol")
	require.NoError(f.t, err)

	indexDesc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: "field", Direction: client.Ascending},
		},
	}

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	_, err = f.createCollectionIndexFor(collection.Name().Value(), indexDesc)
	require.ErrorIs(f.t, err, NewErrUnsupportedIndexFieldType(unsupportedKind))
}

func TestGetIndexes_ShouldReturnListOfAllExistingIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.getProductsCollectionDesc()
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}
	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	indexes, err := f.getAllIndexes()
	assert.NoError(t, err)

	require.Equal(t, 2, len(indexes))

	assert.Equal(t, 1, len(indexes[usersColName]))
	assert.Equal(t, usersIndexDesc.Name, indexes[usersColName][0].Name)
	assert.Equal(t, 1, len(indexes[productsColName]))
	assert.Equal(t, productsIndexDesc.Name, indexes[productsColName][0].Name)
}

func TestGetIndexes_IfInvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), "users_name_index")
	err := f.txn.Systemstore().Put(f.ctx, indexKey.ToDS(), []byte("invalid"))
	assert.NoError(t, err)

	_, err = f.getAllIndexes()
	assert.ErrorIs(t, err, datastore.NewErrInvalidStoredValue(nil))
}

func TestGetIndexes_IfInvalidIndexKeyIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), "users_name_index")
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

func TestGetIndexes_IfSystemStoreFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := f.getAllIndexes()
	assert.ErrorIs(t, err, testErr)
}

func TestGetIndexes_IfSystemStoreFails_ShouldCloseIterator(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	q := mocks.NewQueryResultsWithValues(t)
	q.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(q, nil)

	_, _ = f.getAllIndexes()
}

func TestGetIndexes_IfSystemStoreQueryIteratorFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	q := mocks.NewQueryResultsWithResults(t, query.Result{Error: testErr})
	q.EXPECT().Close().Unset()
	q.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(q, nil)

	_, err := f.getAllIndexes()
	assert.ErrorIs(t, err, testErr)
}

func TestGetIndexes_IfSystemStoreHasInvalidData_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	q := mocks.NewQueryResultsWithValues(t, []byte("invalid"))
	q.EXPECT().Close().Unset()
	q.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(q, nil)

	_, err := f.getAllIndexes()
	assert.ErrorIs(t, err, datastore.NewErrInvalidStoredValue(nil))
}

func TestGetCollectionIndexes_ShouldReturnListOfCollectionIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: usersNameFieldName}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	products := f.getProductsCollectionDesc()
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: productsPriceFieldName}},
	}

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	userIndexes, err := f.getCollectionIndexes(f.users.ID())
	assert.NoError(t, err)
	require.Equal(t, 1, len(userIndexes))
	usersIndexDesc.ID = 1
	assert.Equal(t, usersIndexDesc, userIndexes[0])

	productIndexes, err := f.getCollectionIndexes(products.ID())
	assert.NoError(t, err)
	require.Equal(t, 1, len(productIndexes))
	productsIndexDesc.ID = 1
	assert.Equal(t, productsIndexDesc, productIndexes[0])
}

func TestGetCollectionIndexes_IfSystemStoreFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore)

	_, err := f.getCollectionIndexes(f.users.ID())
	assert.ErrorIs(t, err, testErr)
}

func TestGetCollectionIndexes_IfSystemStoreFails_ShouldCloseIterator(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	query := mocks.NewQueryResultsWithValues(t)
	query.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(query, nil)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore)

	_, _ = f.getCollectionIndexes(f.users.ID())
}

func TestGetCollectionIndexes_IfSystemStoreQueryIteratorFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(mocks.NewQueryResultsWithResults(t, query.Result{Error: testErr}), nil)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore)

	_, err := f.getCollectionIndexes(f.users.ID())
	assert.ErrorIs(t, err, testErr)
}

func TestGetCollectionIndexes_IfInvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), "users_name_index")
	err := f.txn.Systemstore().Put(f.ctx, indexKey.ToDS(), []byte("invalid"))
	assert.NoError(t, err)

	_, err = f.getCollectionIndexes(f.users.ID())
	assert.ErrorIs(t, err, datastore.NewErrInvalidStoredValue(nil))
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

func TestCollectionGetIndexes_ShouldCloseQueryIterator(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(f.t)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore).Maybe()
	queryResults := mocks.NewQueryResultsWithValues(f.t)
	queryResults.EXPECT().Close().Unset()
	queryResults.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(queryResults, nil)

	_, err := f.users.WithTxn(mockedTxn).GetIndexes(f.ctx)
	assert.NoError(t, err)
}

func TestCollectionGetIndexes_IfSystemStoreFails_ReturnError(t *testing.T) {
	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		ExpectedError      error
		GetMockSystemstore func(t *testing.T) *mocks.DSReaderWriter
	}{
		{
			Name:          "Query fails",
			ExpectedError: testErr,
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).Unset()
				store.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
				return store
			},
		},
		{
			Name:          "Query iterator fails",
			ExpectedError: testErr,
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).
					Return(mocks.NewQueryResultsWithResults(t, query.Result{Error: testErr}), nil)
				return store
			},
		},
		{
			Name:          "Query iterator returns invalid value",
			ExpectedError: datastore.NewErrInvalidStoredValue(nil),
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).
					Return(mocks.NewQueryResultsWithValues(t, []byte("invalid")), nil)
				return store
			},
		},
	}

	for _, testCase := range testCases {
		f := newIndexTestFixture(t)
		defer f.db.Close()

		f.createUserCollectionIndexOnName()

		mockedTxn := f.mockTxn()

		mockedTxn.MockSystemstore = testCase.GetMockSystemstore(t)
		mockedTxn.EXPECT().Systemstore().Unset()
		mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore).Maybe()

		_, err := f.users.WithTxn(mockedTxn).GetIndexes(f.ctx)
		require.ErrorIs(t, err, testCase.ExpectedError)
	}
}

func TestCollectionGetIndexes_IfFailsToCreateTxn_ShouldNotCache(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	testErr := errors.New("test error")

	workingRootStore := f.db.rootstore
	mockedRootStore := mocks.NewRootStore(t)
	f.db.rootstore = mockedRootStore
	mockedRootStore.EXPECT().NewTransaction(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := f.users.GetIndexes(f.ctx)
	require.ErrorIs(t, err, testErr)

	f.db.rootstore = workingRootStore

	indexes, err := f.users.GetIndexes(f.ctx)
	require.NoError(t, err)

	require.Equal(t, 1, len(indexes))
	assert.Equal(t, testUsersColIndexName, indexes[0].Name)
}

func TestCollectionGetIndexes_IfStoredIndexWithUnsupportedType_ReturnError(t *testing.T) {
	f := newIndexTestFixtureBare(t)

	const unsupportedKind = client.FieldKind_BOOL_ARRAY
	_, err := f.db.AddSchema(
		f.ctx,
		`type testTypeCol {
			name: String
			field: [Boolean!]
		}`,
	)
	require.NoError(f.t, err)

	collection, err := f.db.GetCollectionByName(f.ctx, "testTypeCol")
	require.NoError(f.t, err)

	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)

	indexDesc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: "field", Direction: client.Ascending},
		},
	}
	indexDescData, err := json.Marshal(indexDesc)
	require.NoError(t, err)

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(mocks.NewQueryResultsWithValues(t, indexDescData), nil)

	_, err = collection.WithTxn(mockedTxn).GetIndexes(f.ctx)
	require.ErrorIs(t, err, NewErrUnsupportedIndexFieldType(unsupportedKind))
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
		indexDesc := client.IndexDescription{
			Name: indexNamePrefix + iStr,
			Fields: []client.IndexedFieldDescription{
				{Name: fieldNamePrefix + iStr, Direction: client.Ascending},
			},
		}

		_, err := f.createCollectionIndexFor(collection.Name().Value(), indexDesc)
		require.NoError(t, err)
	}

	indexes, err := collection.GetIndexes(f.ctx)
	require.NoError(t, err)
	require.Len(t, indexes, num)

	for i := 1; i <= num; i++ {
		assert.Equal(t, indexNamePrefix+toSuffix(i), indexes[i-1].Name, "i = %d", i)
	}
}

func TestDropIndex_ShouldDeleteIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := f.createUserCollectionIndexOnName()

	err := f.dropIndex(usersColName, desc.Name)
	assert.NoError(t, err)

	indexKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), desc.Name)
	_, err = f.txn.Systemstore().Get(f.ctx, indexKey.ToDS())
	assert.Error(t, err)
}

func TestDropIndex_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := f.createUserCollectionIndexOnName()
	f.db.Close()

	err := f.dropIndex(productsColName, desc.Name)
	assert.Error(t, err)
}

func TestDropIndex_IfCollectionDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	err := f.dropIndex(productsColName, "any_name")
	assert.ErrorIs(t, err, NewErrCanNotReadCollection(usersColName, nil))
}

func TestDropIndex_IfFailsToCreateTxn_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	testErr := errors.New("test error")

	mockedRootStore := mocks.NewRootStore(t)
	mockedRootStore.On("Close").Return(nil)

	mockedRootStore.EXPECT().NewTransaction(mock.Anything, mock.Anything).Return(nil, testErr)
	f.db.rootstore = mockedRootStore

	err := f.users.DropIndex(f.ctx, testUsersColIndexName)
	require.ErrorIs(t, err, testErr)
}

func TestDropIndex_IfFailsToDeleteFromStorage_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn().ClearSystemStore()
	systemStoreOn := mockedTxn.MockSystemstore.EXPECT()
	systemStoreOn.Delete(mock.Anything, mock.Anything).Return(testErr)
	f.stubSystemStore(systemStoreOn)
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).Maybe().
		Return(mocks.NewQueryResultsWithValues(t), nil)

	err := f.users.WithTxn(mockedTxn).DropIndex(f.ctx, testUsersColIndexName)
	require.ErrorIs(t, err, testErr)
}

func TestDropIndex_ShouldUpdateCollectionsDescription(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	col := f.users.WithTxn(f.txn)
	_, err := col.CreateIndex(f.ctx, getUsersIndexDescOnName())
	require.NoError(t, err)
	indOnAge, err := col.CreateIndex(f.ctx, getUsersIndexDescOnAge())
	require.NoError(t, err)
	f.commitTxn()

	err = f.users.DropIndex(f.ctx, testUsersColIndexName)
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{indOnAge},
		f.users.Description().Indexes)

	err = f.users.DropIndex(f.ctx, testUsersColIndexAge)
	require.NoError(t, err)

	assert.ElementsMatch(t, []client.IndexDescription{}, f.users.Description().Indexes)
}

func TestDropIndex_IfIndexWithNameDoesNotExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	const name = "not_existing_index"
	err := f.users.DropIndex(f.ctx, name)
	require.ErrorIs(t, err, NewErrIndexWithNameDoesNotExists(name))
}

func TestDropIndex_IfSystemStoreFails_ReturnError(t *testing.T) {
	testErr := errors.New("test error")

	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnName()

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore).Maybe()

	err := f.users.WithTxn(mockedTxn).DropIndex(f.ctx, testUsersColIndexName)
	require.ErrorIs(t, err, testErr)
}

func TestDropAllIndexes_ShouldDeleteAllIndexes(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
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

	err = f.users.(*collection).dropAllIndexes(f.ctx, f.txn)
	assert.NoError(t, err)

	assert.Equal(t, 0, f.countIndexPrefixes(usersColName, ""))
}

func TestDropAllIndexes_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()
	f.db.Close()

	err := f.users.(*collection).dropAllIndexes(f.ctx, f.txn)
	assert.Error(t, err)
}

func TestDropAllIndexes_IfSystemStorageFails_ReturnError(t *testing.T) {
	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		ExpectedError      error
		GetMockSystemstore func(t *testing.T) *mocks.DSReaderWriter
	}{
		{
			Name:          "Query fails",
			ExpectedError: testErr,
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).Unset()
				store.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
				return store
			},
		},
		{
			Name:          "Query iterator fails",
			ExpectedError: testErr,
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).
					Return(mocks.NewQueryResultsWithResults(t, query.Result{Error: testErr}), nil)
				return store
			},
		},
		{
			Name:          "System storage fails to delete",
			ExpectedError: NewErrInvalidStoredIndex(nil),
			GetMockSystemstore: func(t *testing.T) *mocks.DSReaderWriter {
				store := mocks.NewDSReaderWriter(t)
				store.EXPECT().Query(mock.Anything, mock.Anything).
					Return(mocks.NewQueryResultsWithValues(t, []byte{}), nil)
				store.EXPECT().Delete(mock.Anything, mock.Anything).Maybe().Return(testErr)
				return store
			},
		},
	}

	for _, testCase := range testCases {
		f := newIndexTestFixture(t)
		defer f.db.Close()
		f.createUserCollectionIndexOnName()

		mockedTxn := f.mockTxn()

		mockedTxn.MockSystemstore = testCase.GetMockSystemstore(t)
		mockedTxn.EXPECT().Systemstore().Unset()
		mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore).Maybe()

		err := f.users.(*collection).dropAllIndexes(f.ctx, f.txn)
		assert.ErrorIs(t, err, testErr, testCase.Name)
	}
}

func TestDropAllIndexes_ShouldCloseQueryIterator(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	mockedTxn := f.mockTxn()

	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	q := mocks.NewQueryResultsWithValues(t, []byte{})
	q.EXPECT().Close().Unset()
	q.EXPECT().Close().Return(nil)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(q, nil)
	mockedTxn.MockSystemstore.EXPECT().Delete(mock.Anything, mock.Anything).Maybe().Return(nil)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore).Maybe()

	_ = f.users.(*collection).dropAllIndexes(f.ctx, f.txn)
}

func TestNewCollectionIndex_IfDescriptionHasNoFields_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := getUsersIndexDescOnName()
	desc.Fields = nil
	_, err := NewCollectionIndex(f.users, desc)
	require.ErrorIs(t, err, NewErrIndexDescHasNoFields(desc))
}

func TestNewCollectionIndex_IfDescriptionHasNonExistingField_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	desc := getUsersIndexDescOnName()
	desc.Fields[0].Name = "non_existing_field"
	_, err := NewCollectionIndex(f.users, desc)
	require.ErrorIs(t, err, client.NewErrFieldNotExist(desc.Fields[0].Name))
}
