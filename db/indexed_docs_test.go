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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/db/fetcher"
	fetcherMocks "github.com/sourcenetwork/defradb/db/fetcher/mocks"
)

const testValuePrefix = "v"
const testNilValue = "n"

type userDoc struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Weight float64 `json:"weight"`
}

type productDoc struct {
	ID        int     `json:"id"`
	Price     float64 `json:"price"`
	Category  string  `json:"category"`
	Available bool    `json:"available"`
}

func (f *indexTestFixture) saveDocToCollection(doc *client.Document, col client.Collection) {
	err := col.Create(f.ctx, doc)
	require.NoError(f.t, err)
	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
}

func (f *indexTestFixture) newUserDoc(name string, age int) *client.Document {
	d := userDoc{Name: name, Age: age, Weight: 154.1}
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data)
	require.NoError(f.t, err)
	return doc
}

func (f *indexTestFixture) newProdDoc(id int, price float64, cat string) *client.Document {
	d := productDoc{ID: id, Price: price, Category: cat}
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data)
	require.NoError(f.t, err)
	return doc
}

// indexKeyBuilder is a helper for building index keys that can be turned into a string.
// The format of the non-unique index key is: "/<collection_id>/<index_id>/<value>/<doc_id>"
// Example: "/5/1/12/bae-61cd6879-63ca-5ca9-8731-470a3c1dac69"
type indexKeyBuilder struct {
	f         *indexTestFixture
	colName   string
	fieldName string
	doc       *client.Document
	values    []string
	isUnique  bool
}

func newIndexKeyBuilder(f *indexTestFixture) *indexKeyBuilder {
	return &indexKeyBuilder{f: f}
}

func (b *indexKeyBuilder) Col(colName string) *indexKeyBuilder {
	b.colName = colName
	return b
}

// Field sets the field name for the index key.
// If the field name is not set, the index key will contain only collection id.
// When building a key it will it will find the field id to use in the key.
func (b *indexKeyBuilder) Field(fieldName string) *indexKeyBuilder {
	b.fieldName = fieldName
	return b
}

// Doc sets the document for the index key.
// For non-unique index keys, it will try to find the field value in the document
// corresponding to the field name set in the builder.
// As the last value in the index key, it will use the document id.
func (b *indexKeyBuilder) Doc(doc *client.Document) *indexKeyBuilder {
	b.doc = doc
	return b
}

// Values sets the values for the index key.
// It will override the field values stored in the document.
func (b *indexKeyBuilder) Values(values ...string) *indexKeyBuilder {
	b.values = values
	return b
}

func (b *indexKeyBuilder) Unique() *indexKeyBuilder {
	b.isUnique = true
	return b
}

func (b *indexKeyBuilder) Build() core.IndexDataStoreKey {
	key := core.IndexDataStoreKey{}

	if b.colName == "" {
		return key
	}

	cols, err := b.f.db.getAllCollections(b.f.ctx, b.f.txn)
	require.NoError(b.f.t, err)
	var collection client.Collection
	for _, col := range cols {
		if col.Name() == b.colName {
			collection = col
			break
		}
	}
	if collection == nil {
		panic(errors.New("collection not found"))
	}
	key.CollectionID = strconv.Itoa(int(collection.ID()))

	if b.fieldName == "" {
		return key
	}

	indexes, err := collection.GetIndexes(b.f.ctx)
	require.NoError(b.f.t, err)
	for _, index := range indexes {
		if index.Fields[0].Name == b.fieldName {
			key.IndexID = strconv.Itoa(int(index.ID))
			break
		}
	}

	if b.doc != nil {
		var fieldStrVal string
		if len(b.values) == 0 {
			fieldVal, err := b.doc.Get(b.fieldName)
			require.NoError(b.f.t, err)
			fieldStrVal = fmt.Sprintf("%s%v", testValuePrefix, fieldVal)
		} else {
			fieldStrVal = b.values[0]
		}

		key.FieldValues = []string{fieldStrVal, testValuePrefix + b.doc.Key().String()}
	} else if len(b.values) > 0 {
		key.FieldValues = b.values
	}

	return key
}

func (f *indexTestFixture) getPrefixFromDataStore(prefix string) [][]byte {
	q := query.Query{Prefix: prefix}
	res, err := f.txn.Datastore().Query(f.ctx, q)
	require.NoError(f.t, err)

	var keys [][]byte
	for r := range res.Next() {
		keys = append(keys, r.Entry.Value)
	}
	return keys
}

func (f *indexTestFixture) mockTxn() *mocks.MultiStoreTxn {
	mockedTxn := mocks.NewTxnWithMultistore(f.t)

	systemStoreOn := mockedTxn.MockSystemstore.EXPECT()
	f.resetSystemStoreStubs(systemStoreOn)
	f.stubSystemStore(systemStoreOn)

	f.txn = mockedTxn
	return mockedTxn
}

func (*indexTestFixture) resetSystemStoreStubs(systemStoreOn *mocks.DSReaderWriter_Expecter) {
	systemStoreOn.Query(mock.Anything, mock.Anything).Unset()
	systemStoreOn.Get(mock.Anything, mock.Anything).Unset()
	systemStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Unset()
}

func (f *indexTestFixture) stubSystemStore(systemStoreOn *mocks.DSReaderWriter_Expecter) {
	desc := getUsersIndexDescOnName()
	desc.ID = 1
	indexOnNameDescData, err := json.Marshal(desc)
	require.NoError(f.t, err)

	colIndexKey := core.NewCollectionIndexKey(f.users.Description().Name, "")
	matchPrefixFunc := func(q query.Query) bool { return q.Prefix == colIndexKey.ToDS().String() }

	systemStoreOn.Query(mock.Anything, mock.MatchedBy(matchPrefixFunc)).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t, indexOnNameDescData), nil)
	systemStoreOn.Query(mock.Anything, mock.Anything).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t), nil)

	colKey := core.NewCollectionKey(f.users.Name())
	systemStoreOn.Get(mock.Anything, colKey.ToDS()).Maybe().Return([]byte(userColVersionID), nil)

	colVersionIDKey := core.NewCollectionSchemaVersionKey(userColVersionID)
	colDesc := getUsersCollectionDesc()
	colDesc.ID = 1
	for i := range colDesc.Schema.Fields {
		colDesc.Schema.Fields[i].ID = client.FieldID(i)
	}
	colDescBytes, err := json.Marshal(colDesc)
	require.NoError(f.t, err)
	systemStoreOn.Get(mock.Anything, colVersionIDKey.ToDS()).Maybe().Return(colDescBytes, nil)

	colIndexOnNameKey := core.NewCollectionIndexKey(f.users.Description().Name, testUsersColIndexName)
	systemStoreOn.Get(mock.Anything, colIndexOnNameKey.ToDS()).Maybe().Return(indexOnNameDescData, nil)

	sequenceKey := core.NewSequenceKey(fmt.Sprintf("%s/%d", core.COLLECTION_INDEX, f.users.ID()))
	systemStoreOn.Get(mock.Anything, sequenceKey.ToDS()).Maybe().Return([]byte{0, 0, 0, 0, 0, 0, 0, 1}, nil)

	systemStoreOn.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)

	systemStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

	systemStoreOn.Has(mock.Anything, mock.Anything).Maybe().Return(false, nil)

	systemStoreOn.Delete(mock.Anything, mock.Anything).Maybe().Return(nil)
}

func TestNonUnique_IfDocIsAdded_ShouldBeIndexed(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)
	//f.commitTxn()

	key := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfFailsToStoredIndexedDoc_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	key := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

	mockTxn := f.mockTxn()

	dataStoreOn := mockTxn.MockDatastore.EXPECT()
	dataStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Unset()
	dataStoreOn.Put(mock.Anything, key.ToDS(), mock.Anything).Return(errors.New("error"))
	dataStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(f.t, err, NewErrFailedToStoreIndexedField("name", nil))
}

// @todo: should store as nil value?
func TestNonUnique_IfDocDoesNotHaveIndexedField_SkipIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	data, err := json.Marshal(struct {
		Age    int     `json:"age"`
		Weight float64 `json:"weight"`
	}{Age: 21, Weight: 154.1})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data)
	require.NoError(f.t, err)

	err = f.users.Create(f.ctx, doc)
	require.NoError(f.t, err)

	key := newIndexKeyBuilder(f).Col(usersColName).Build()
	prefixes := f.getPrefixFromDataStore(key.ToString())
	assert.Len(t, prefixes, 0)
}

func TestNonUnique_IfSystemStorageHasInvalidIndexDescription_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)

	mockTxn := f.mockTxn().ClearSystemStore()
	systemStoreOn := mockTxn.MockSystemstore.EXPECT()
	systemStoreOn.Query(mock.Anything, mock.Anything).
		Return(mocks.NewQueryResultsWithValues(t, []byte("invalid")), nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(t, err, NewErrInvalidStoredIndex(nil))
}

func TestNonUnique_IfSystemStorageFailsToReadIndexDesc_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)

	testErr := errors.New("test error")

	mockTxn := f.mockTxn().ClearSystemStore()
	systemStoreOn := mockTxn.MockSystemstore.EXPECT()
	systemStoreOn.Query(mock.Anything, mock.Anything).
		Return(nil, testErr)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(t, err, testErr)
}

func TestNonUnique_IfIndexIntField_StoreIt(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Field(usersAgeFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfMultipleCollectionsWithIndexes_StoreIndexWithCollectionID(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.createCollection(getUsersCollectionDesc())
	products := f.createCollection(getProductsCollectionDesc())

	_, err := f.createCollectionIndexFor(users.Name(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(products.Name(), getProductsIndexDescOnCategory())
	require.NoError(f.t, err)
	f.commitTxn()

	userDoc := f.newUserDoc("John", 21)
	prodDoc := f.newProdDoc(1, 3, "games")

	err = users.Create(f.ctx, userDoc)
	require.NoError(f.t, err)
	err = products.Create(f.ctx, prodDoc)
	require.NoError(f.t, err)
	f.commitTxn()

	userDocKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(userDoc).Build()
	prodDocKey := newIndexKeyBuilder(f).Col(productsColName).Field(productsCategoryFieldName).Doc(prodDoc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, userDocKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, prodDocKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfMultipleIndexes_StoreIndexWithIndexID(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)

	nameKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()
	ageKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersAgeFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, nameKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, ageKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_StoringIndexedFieldValueOfDifferentTypes(t *testing.T) {
	f := newIndexTestFixtureBare(t)

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	testCase := []struct {
		Name      string
		FieldKind client.FieldKind
		// FieldVal is the value the index will receive for serialization
		FieldVal   any
		ShouldFail bool
		// Stored is the value that is stored as part of the index value key
		Stored string
	}{
		{Name: "invalid int", FieldKind: client.FieldKind_INT, FieldVal: "invalid", ShouldFail: true},
		{Name: "invalid float", FieldKind: client.FieldKind_FLOAT, FieldVal: "invalid", ShouldFail: true},
		{Name: "invalid bool", FieldKind: client.FieldKind_BOOL, FieldVal: "invalid", ShouldFail: true},
		{Name: "invalid datetime", FieldKind: client.FieldKind_DATETIME, FieldVal: nowStr[1:], ShouldFail: true},

		{Name: "valid int", FieldKind: client.FieldKind_INT, FieldVal: 12, Stored: "12"},
		{Name: "valid float", FieldKind: client.FieldKind_FLOAT, FieldVal: 36.654, Stored: "36.654"},
		{Name: "valid bool true", FieldKind: client.FieldKind_BOOL, FieldVal: true, Stored: "1"},
		{Name: "valid bool false", FieldKind: client.FieldKind_BOOL, FieldVal: false, Stored: "0"},
		{Name: "valid datetime string", FieldKind: client.FieldKind_DATETIME, FieldVal: nowStr, Stored: nowStr},
		{Name: "valid empty string", FieldKind: client.FieldKind_STRING, FieldVal: "", Stored: ""},
	}

	for i, tc := range testCase {
		desc := client.CollectionDescription{
			Name: "testTypeCol" + strconv.Itoa(i),
			Schema: client.SchemaDescription{
				Fields: []client.FieldDescription{
					{
						Name: "_key",
						Kind: client.FieldKind_DocKey,
					},
					{
						Name: "field",
						Kind: tc.FieldKind,
						Typ:  client.LWW_REGISTER,
					},
				},
			},
		}

		collection := f.createCollection(desc)

		indexDesc := client.IndexDescription{
			Fields: []client.IndexedFieldDescription{
				{Name: "field", Direction: client.Ascending},
			},
		}

		_, err := f.createCollectionIndexFor(collection.Name(), indexDesc)
		require.NoError(f.t, err)
		f.commitTxn()

		d := struct {
			Field any `json:"field"`
		}{Field: tc.FieldVal}
		data, err := json.Marshal(d)
		require.NoError(f.t, err)
		doc, err := client.NewDocFromJSON(data)
		require.NoError(f.t, err)

		err = collection.Create(f.ctx, doc)
		f.commitTxn()
		if tc.ShouldFail {
			require.ErrorIs(f.t, err, NewErrCanNotIndexInvalidFieldValue(nil), "test case: %s", tc.Name)
		} else {
			assertMsg := fmt.Sprintf("test case: %s", tc.Name)
			require.NoError(f.t, err, assertMsg)

			keyBuilder := newIndexKeyBuilder(f).Col(collection.Name()).Field("field").Doc(doc)
			if tc.Stored != "" {
				keyBuilder.Values(testValuePrefix + tc.Stored)
			}
			key := keyBuilder.Build()

			keyStr := key.ToDS()
			data, err := f.txn.Datastore().Get(f.ctx, keyStr)
			require.NoError(t, err, assertMsg)
			assert.Len(t, data, 0, assertMsg)
		}
	}
}

func TestNonUnique_IfIndexedFieldIsNil_StoreItAsNil(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	docJSON, err := json.Marshal(struct {
		Age int `json:"age"`
	}{Age: 44})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(docJSON)
	require.NoError(f.t, err)

	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).
		Values(testNilValue).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUniqueDrop_ShouldDeleteStoredIndexedFields(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.createCollection(getUsersCollectionDesc())
	_, err := f.createCollectionIndexFor(users.Name(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name(), getUsersIndexDescOnAge())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name(), getUsersIndexDescOnWeight())
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newUserDoc("John", 21), users)
	f.saveDocToCollection(f.newUserDoc("Islam", 23), users)

	products := f.createCollection(getProductsCollectionDesc())
	_, err = f.createCollectionIndexFor(products.Name(), getProductsIndexDescOnCategory())
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newProdDoc(1, 55, "games"), products)

	userNameKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Build()
	userAgeKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersAgeFieldName).Build()
	userWeightKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersWeightFieldName).Build()
	prodCatKey := newIndexKeyBuilder(f).Col(productsColName).Field(productsCategoryFieldName).Build()

	err = f.dropIndex(usersColName, testUsersColIndexAge)
	require.NoError(f.t, err)

	assert.Len(t, f.getPrefixFromDataStore(userNameKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(userAgeKey.ToString()), 0)
	assert.Len(t, f.getPrefixFromDataStore(userWeightKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(prodCatKey.ToString()), 1)
}

func TestNonUniqueDrop_IfFailsToQueryDataStorage_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockDatastore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
	mockedTxn.EXPECT().Datastore().Unset()
	mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore)

	err := f.dropIndex(usersColName, testUsersColIndexName)
	require.ErrorIs(t, err, testErr)
}

func TestNonUniqueDrop_IfQueryIteratorFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockDatastore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(mocks.NewQueryResultsWithResults(t, query.Result{Error: testErr}), nil)
	mockedTxn.EXPECT().Datastore().Unset()
	mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore)

	err := f.dropIndex(usersColName, testUsersColIndexName)
	require.ErrorIs(t, err, testErr)
}

func TestNonUniqueDrop_ShouldCloseQueryIterator(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	mockedTxn := f.mockTxn()

	mockedTxn.MockDatastore = mocks.NewDSReaderWriter(f.t)
	mockedTxn.EXPECT().Datastore().Unset()
	mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore).Maybe()
	queryResults := mocks.NewQueryResultsWithValues(f.t)
	queryResults.EXPECT().Close().Unset()
	queryResults.EXPECT().Close().Return(nil)
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(queryResults, nil)

	err := f.dropIndex(usersColName, testUsersColIndexName)
	assert.NoError(t, err)
}

func TestNonUniqueDrop_IfDatastoreFailsToDelete_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	mockedTxn := f.mockTxn()
	mockedTxn.MockDatastore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockDatastore.EXPECT().Query(mock.Anything, mock.Anything).
		Return(mocks.NewQueryResultsWithValues(t, []byte{}), nil)
	mockedTxn.MockDatastore.EXPECT().Delete(mock.Anything, mock.Anything).
		Return(errors.New("error"))
	mockedTxn.EXPECT().Datastore().Unset()
	mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore)

	err := f.dropIndex(usersColName, testUsersColIndexName)
	require.ErrorIs(t, err, NewCanNotDeleteIndexedField(nil))
}

func TestNonUniqueUpdate_ShouldDeleteOldValueAndStoreNewOne(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	cases := []struct {
		Name     string
		NewValue string
		Exec     func(doc *client.Document) error
	}{
		{
			Name:     "update",
			NewValue: "Islam",
			Exec: func(doc *client.Document) error {
				return f.users.Update(f.ctx, doc)
			},
		},
		{
			Name:     "save",
			NewValue: "Andy",
			Exec: func(doc *client.Document) error {
				return f.users.Save(f.ctx, doc)
			},
		},
	}

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)

	for _, tc := range cases {
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

		err := doc.Set(usersNameFieldName, tc.NewValue)
		require.NoError(t, err)
		err = tc.Exec(doc)
		require.NoError(t, err)
		f.commitTxn()

		newKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.Error(t, err)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.NoError(t, err)
	}
}

func TestNonUniqueUpdate_IfFailsToReadIndexDescription_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)

	err := doc.Set(usersNameFieldName, "Islam")
	require.NoError(t, err)

	// retrieve the collection without index cached
	usersCol, err := f.db.getCollectionByName(f.ctx, f.txn, usersColName)
	require.NoError(t, err)

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore)
	mockedTxn.MockDatastore.EXPECT().Get(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockDatastore.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte{}, nil)

	err = usersCol.WithTxn(mockedTxn).Update(f.ctx, doc)
	require.ErrorIs(t, err, testErr)
}

func TestNonUniqueUpdate_IfFetcherFails_ReturnError(t *testing.T) {
	testError := errors.New("test error")

	cases := []struct {
		Name           string
		PrepareFetcher func() fetcher.Fetcher
	}{
		{
			Name: "Fails to init",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Unset()
				f.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to start",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Start(mock.Anything, mock.Anything, mock.Anything).Unset()
				f.EXPECT().Start(mock.Anything, mock.Anything, mock.Anything).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to fetch next decoded",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().FetchNextDecoded(mock.Anything).Unset()
				f.EXPECT().FetchNextDecoded(mock.Anything).Return(nil, testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to close",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(testError)
				return f
			},
		},
	}

	for _, tc := range cases {
		f := newIndexTestFixture(t)
		f.createUserCollectionIndexOnName()

		doc := f.newUserDoc("John", 21)
		f.saveDocToCollection(doc, f.users)

		f.users.fetcherFactory = tc.PrepareFetcher
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

		err := doc.Set(usersNameFieldName, "Islam")
		require.NoError(t, err, tc.Name)
		err = f.users.Update(f.ctx, doc)
		require.Error(t, err, tc.Name)

		newKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersNameFieldName).Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.NoError(t, err, tc.Name)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.Error(t, err, tc.Name)
	}
}

func TestNonUniqueUpdate_IfFailsToUpdateIndex_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)
	f.commitTxn()

	validKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersAgeFieldName).Doc(doc).Build()
	invalidKey := newIndexKeyBuilder(f).Col(usersColName).Field(usersAgeFieldName).Doc(doc).
		Values("invalid").Build()

	err := f.txn.Datastore().Delete(f.ctx, validKey.ToDS())
	require.NoError(f.t, err)
	err = f.txn.Datastore().Put(f.ctx, invalidKey.ToDS(), []byte{})
	require.NoError(f.t, err)
	f.commitTxn()

	err = doc.Set(usersAgeFieldName, 23)
	require.NoError(t, err)
	err = f.users.Update(f.ctx, doc)
	require.Error(t, err)
}

func TestNonUniqueUpdate_ShouldPassToFetcherOnlyRelevantFields(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	f.users.fetcherFactory = func() fetcher.Fetcher {
		f := fetcherMocks.NewStubbedFetcher(t)
		f.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Unset()
		f.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(
				col *client.CollectionDescription,
				fields []*client.FieldDescription,
				reverse, showDeleted bool,
			) error {
				require.Equal(t, 2, len(fields))
				require.ElementsMatch(t,
					[]string{usersNameFieldName, usersAgeFieldName},
					[]string{fields[0].Name, fields[1].Name})
				return errors.New("early exit")
			})
		return f
	}
	doc := f.newUserDoc("John", 21)
	f.saveDocToCollection(doc, f.users)

	err := doc.Set(usersNameFieldName, "Islam")
	require.NoError(t, err)
	_ = f.users.Update(f.ctx, doc)
}

func TestNonUniqueUpdate_IfDatastoreFails_ReturnError(t *testing.T) {
	testErr := errors.New("error")

	cases := []struct {
		Name          string
		StubDataStore func(*mocks.DSReaderWriter_Expecter)
	}{
		{
			Name: "Delete old value",
			StubDataStore: func(ds *mocks.DSReaderWriter_Expecter) {
				ds.Delete(mock.Anything, mock.Anything).Return(testErr)
				ds.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)
			},
		},
		{
			Name: "Store new value",
			StubDataStore: func(ds *mocks.DSReaderWriter_Expecter) {
				ds.Delete(mock.Anything, mock.Anything).Maybe().Return(nil)
				ds.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)
				ds.Put(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(testErr)
			},
		},
	}

	for _, tc := range cases {
		f := newIndexTestFixture(t)
		f.createUserCollectionIndexOnName()

		doc := f.newUserDoc("John", 21)

		f.users.fetcherFactory = func() fetcher.Fetcher {
			df := fetcherMocks.NewStubbedFetcher(t)
			df.EXPECT().FetchNextDecoded(mock.Anything).Unset()
			df.EXPECT().FetchNextDecoded(mock.Anything).Return(doc, nil)
			return df
		}

		f.saveDocToCollection(doc, f.users)

		err := doc.Set(usersNameFieldName, "Islam")
		require.NoError(t, err)

		mockedTxn := f.mockTxn()
		mockedTxn.MockDatastore = mocks.NewDSReaderWriter(f.t)
		tc.StubDataStore(mockedTxn.MockDatastore.EXPECT())
		mockedTxn.EXPECT().Datastore().Unset()
		mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore).Maybe()

		err = f.users.WithTxn(mockedTxn).Update(f.ctx, doc)
		require.ErrorIs(t, err, testErr)
	}
}
