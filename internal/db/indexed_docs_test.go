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
	"encoding/json"
	"errors"
	"testing"

	ipfsDatastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	fetcherMocks "github.com/sourcenetwork/defradb/internal/db/fetcher/mocks"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type userDoc struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Weight  float64  `json:"weight"`
	Numbers []int    `json:"numbers"`
	Hobbies []string `json:"hobbies"`
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
	f.commitTxn()
	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
}

func (f *indexTestFixture) deleteDocFromCollection(docID client.DocID, col client.Collection) {
	res, err := col.Delete(f.ctx, docID)
	require.NoError(f.t, err)
	require.True(f.t, res)
	f.commitTxn()
	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
}

func (f *indexTestFixture) newUserDoc(name string, age int, col client.Collection) *client.Document {
	d := userDoc{Name: name, Age: age, Weight: 154.1}
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data, col.Definition())
	require.NoError(f.t, err)
	return doc
}

func (f *indexTestFixture) newCustomUserDoc(d userDoc, col client.Collection) *client.Document {
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data, col.Definition())
	require.NoError(f.t, err)
	return doc
}

func (f *indexTestFixture) newProdDoc(id int, price float64, cat string, col client.Collection) *client.Document {
	d := productDoc{ID: id, Price: price, Category: cat}
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data, col.Definition())
	require.NoError(f.t, err)
	return doc
}

// indexKeyBuilder is a helper for building index keys that can be turned into a string.
// The format of the non-unique index key is: "/<collection_id>/<index_id>/<value>/<doc_id>"
// Example: "/5/1/12/bae-61cd6879-63ca-5ca9-8731-470a3c1dac69"
type indexKeyBuilder struct {
	f                *indexTestFixture
	colName          string
	fieldsNames      []string
	descendingFields []bool
	doc              *client.Document
	isUnique         bool
	arrayFieldValues map[string]any
}

func newIndexKeyBuilder(f *indexTestFixture) *indexKeyBuilder {
	return &indexKeyBuilder{f: f, arrayFieldValues: make(map[string]any)}
}

func (b *indexKeyBuilder) Col(colName string) *indexKeyBuilder {
	b.colName = colName
	return b
}

// Fields sets the fields names for the index key.
// If the field name is not set, the index key will contain only collection id.
// When building a key it will it will find the field id to use in the key.
func (b *indexKeyBuilder) Fields(fieldsNames ...string) *indexKeyBuilder {
	b.fieldsNames = fieldsNames
	return b
}

// ArrayFieldVal sets the value for the array field.
// The value should be of a single element of the array, as index indexes array fields by each element.
// If ArrayFieldVal is not set and index array field is present, it will take array first element as a value.
func (b *indexKeyBuilder) ArrayFieldVal(fieldName string, val any) *indexKeyBuilder {
	b.arrayFieldValues[fieldName] = val
	return b
}

// Fields sets the fields names for the index key.
func (b *indexKeyBuilder) DescendingFields(descending ...bool) *indexKeyBuilder {
	b.descendingFields = descending
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

// Unique sets the index key to be unique.
func (b *indexKeyBuilder) Unique() *indexKeyBuilder {
	b.isUnique = true
	return b
}

func (b *indexKeyBuilder) Build() core.IndexDataStoreKey {
	key := core.IndexDataStoreKey{}

	if b.colName == "" {
		return key
	}

	ctx := SetContextTxn(b.f.ctx, b.f.txn)
	cols, err := b.f.db.getCollections(ctx, client.CollectionFetchOptions{})
	require.NoError(b.f.t, err)
	var collection client.Collection
	for _, col := range cols {
		if col.Name().Value() == b.colName {
			collection = col
			break
		}
	}
	if collection == nil {
		panic(errors.New("collection not found"))
	}
	key.CollectionID = collection.ID()

	if len(b.fieldsNames) == 0 {
		return key
	}

	indexes, err := collection.GetIndexes(b.f.ctx)
	require.NoError(b.f.t, err)
indexLoop:
	for _, index := range indexes {
		if len(index.Fields) == len(b.fieldsNames) {
			for i := range index.Fields {
				if index.Fields[i].Name != b.fieldsNames[i] {
					continue indexLoop
				}
			}
			key.IndexID = index.ID
			break indexLoop
		}
	}

	if b.doc != nil {
		hasNilValue := false
		for i, fieldName := range b.fieldsNames {
			fieldValue, err := b.doc.GetValue(fieldName)
			if err != nil {
				if !errors.Is(err, client.ErrFieldNotExist) {
					require.NoError(b.f.t, err)
				}
			}
			var val client.NormalValue
			if fieldValue != nil {
				val = fieldValue.NormalValue()
			} else {
				kind := client.FieldKind_NILLABLE_STRING
				if fieldName == usersAgeFieldName {
					kind = client.FieldKind_NILLABLE_INT
				} else if fieldName == usersWeightFieldName {
					kind = client.FieldKind_NILLABLE_FLOAT
				}
				val, err = client.NewNormalNil(kind)
				require.NoError(b.f.t, err)
			}
			if val.IsNil() {
				hasNilValue = true
			} else if val.IsArray() {
				if arrVal, ok := b.arrayFieldValues[fieldName]; ok {
					if normVal, ok := arrVal.(client.NormalValue); ok {
						val = normVal
					} else {
						val, err = client.NewNormalValue(arrVal)
						require.NoError(b.f.t, err, "given value is not a normal value")
					}
				} else {
					arrVals, err := client.ToArrayOfNormalValues(val)
					require.NoError(b.f.t, err)
					require.Greater(b.f.t, len(arrVals), 0, "empty array can not be indexed")
					val = arrVals[0]
				}
			}
			descending := false
			if i < len(b.descendingFields) {
				descending = b.descendingFields[i]
			}
			key.Fields = append(key.Fields, core.IndexedField{Value: val, Descending: descending})
		}

		if !b.isUnique || hasNilValue {
			key.Fields = append(key.Fields, core.IndexedField{Value: client.NewNormalString(b.doc.ID().String())})
		}
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
	if f.users == nil {
		f.users = f.addUsersCollection()
	}
	desc := getUsersIndexDescOnName()
	desc.ID = 1
	indexOnNameDescData, err := json.Marshal(desc)
	require.NoError(f.t, err)

	colIndexKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), "")
	matchPrefixFunc := func(q query.Query) bool {
		return q.Prefix == colIndexKey.ToDS().String()
	}

	systemStoreOn.Query(mock.Anything, mock.MatchedBy(matchPrefixFunc)).
		RunAndReturn(func(context.Context, query.Query) (query.Results, error) {
			return mocks.NewQueryResultsWithValues(f.t, indexOnNameDescData), nil
		}).Maybe()
	systemStoreOn.Query(mock.Anything, mock.MatchedBy(matchPrefixFunc)).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t, indexOnNameDescData), nil)
	systemStoreOn.Query(mock.Anything, mock.Anything).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t), nil)

	colIndexOnNameKey := core.NewCollectionIndexKey(immutable.Some(f.users.ID()), testUsersColIndexName)
	systemStoreOn.Get(mock.Anything, colIndexOnNameKey.ToDS()).Maybe().Return(indexOnNameDescData, nil)

	if f.users != nil {
		sequenceKey := core.NewIndexIDSequenceKey(f.users.ID())
		systemStoreOn.Get(mock.Anything, sequenceKey.ToDS()).Maybe().Return([]byte{0, 0, 0, 0, 0, 0, 0, 1}, nil)
	}

	systemStoreOn.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)

	systemStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

	systemStoreOn.Has(mock.Anything, mock.Anything).Maybe().Return(false, nil)

	systemStoreOn.Delete(mock.Anything, mock.Anything).Maybe().Return(nil)
}

func TestNonUnique_IfDocIsAdded_ShouldBeIndexed(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfDocIsDeleted_ShouldRemoveIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)
	f.deleteDocFromCollection(doc.ID(), f.users)

	userNameKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Build()
	assert.Len(t, f.getPrefixFromDataStore(userNameKey.ToString()), 0)
}

func TestNonUnique_IfDocWithDescendingOrderIsAdded_ShouldBeIndexed(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexDesc := getUsersIndexDescOnName()
	indexDesc.Fields[0].Descending = true
	_, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).DescendingFields(true).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfDocDoesNotHaveIndexedField_SkipIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	data, err := json.Marshal(struct {
		Age    int     `json:"age"`
		Weight float64 `json:"weight"`
	}{Age: 21, Weight: 154.1})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data, f.users.Definition())
	require.NoError(f.t, err)

	err = f.users.Create(f.ctx, doc)
	require.NoError(f.t, err)

	key := newIndexKeyBuilder(f).Col(usersColName).Build()
	prefixes := f.getPrefixFromDataStore(key.ToString())
	assert.Len(t, prefixes, 0)
}

func TestNonUnique_IfIndexIntField_StoreIt(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfMultipleCollectionsWithIndexes_StoreIndexWithCollectionID(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.addUsersCollection()
	products := f.getProductsCollectionDesc()

	_, err := f.createCollectionIndexFor(users.Name().Value(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(products.Name().Value(), getProductsIndexDescOnCategory())
	require.NoError(f.t, err)
	f.commitTxn()

	userDoc := f.newUserDoc("John", 21, users)
	prodDoc := f.newProdDoc(1, 3, "games", products)

	err = users.Create(f.ctx, userDoc)
	require.NoError(f.t, err)
	err = products.Create(f.ctx, prodDoc)
	require.NoError(f.t, err)
	f.commitTxn()

	userDocID := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(userDoc).Build()
	prodDocID := newIndexKeyBuilder(f).Col(productsColName).Fields(productsCategoryFieldName).Doc(prodDoc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, userDocID.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, prodDocID.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfMultipleIndexes_StoreIndexWithIndexID(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	nameKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()
	ageKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, nameKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, ageKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

// func TestNonUnique_StoringIndexedFieldValueOfDifferentTypes(t *testing.T) {
// 	f := newIndexTestFixtureBare(t)

// 	now := time.Now()
// 	nowStr := now.Format(time.RFC3339)

// 	testCase := []struct {
// 		Name      string
// 		FieldKind client.FieldKind
// 		// FieldVal is the value the index will receive for serialization
// 		FieldVal   any
// 		ShouldFail bool
// 	}{
// 		{Name: "invalid int", FieldKind: client.FieldKind_INT, FieldVal: "invalid", ShouldFail: true},
// 		{Name: "invalid float", FieldKind: client.FieldKind_FLOAT, FieldVal: "invalid", ShouldFail: true},
// 		{Name: "invalid bool", FieldKind: client.FieldKind_BOOL, FieldVal: "invalid", ShouldFail: true},
// 		{Name: "invalid datetime", FieldKind: client.FieldKind_DATETIME, FieldVal: nowStr[1:], ShouldFail: true},
// 		{Name: "invalid datetime type", FieldKind: client.FieldKind_DATETIME, FieldVal: 1, ShouldFail: true},
// 		{Name: "invalid blob", FieldKind: client.FieldKind_BLOB, FieldVal: "invalid", ShouldFail: true},
// 		{Name: "invalid blob type", FieldKind: client.FieldKind_BLOB, FieldVal: 1, ShouldFail: true},

// 		{Name: "valid int", FieldKind: client.FieldKind_INT, FieldVal: 12},
// 		{Name: "valid float", FieldKind: client.FieldKind_FLOAT, FieldVal: 36.654},
// 		{Name: "valid bool true", FieldKind: client.FieldKind_BOOL, FieldVal: true},
// 		{Name: "valid bool false", FieldKind: client.FieldKind_BOOL, FieldVal: false},
// 		{Name: "valid datetime string", FieldKind: client.FieldKind_DATETIME, FieldVal: nowStr},
// 		{Name: "valid empty string", FieldKind: client.FieldKind_STRING, FieldVal: ""},
// 		{Name: "valid blob type", FieldKind: client.FieldKind_BLOB, FieldVal: "00ff"},
// 	}

// 	for i, tc := range testCase {
// 		_, err := f.db.AddSchema(
// 			f.ctx,
// 			fmt.Sprintf(
// 				`type %s {
// 					field: %s
// 				}`,
// 				"testTypeCol"+strconv.Itoa(i),
// 				tc.FieldKind.String(),
// 			),
// 		)
// 		require.NoError(f.t, err)

// 		collection, err := f.db.GetCollectionByName(f.ctx, "testTypeCol"+strconv.Itoa(i))
// 		require.NoError(f.t, err)

// 		f.txn, err = f.db.NewTxn(f.ctx, false)
// 		require.NoError(f.t, err)

// 		indexDesc := client.IndexDescription{
// 			Fields: []client.IndexedFieldDescription{
// 				{Name: "field", Direction: client.Ascending},
// 			},
// 		}

// 		_, err = f.createCollectionIndexFor(collection.Name(), indexDesc)
// 		require.NoError(f.t, err)
// 		f.commitTxn()

// 		d := struct {
// 			Field any `json:"field"`
// 		}{Field: tc.FieldVal}
// 		data, err := json.Marshal(d)
// 		require.NoError(f.t, err)
// 		doc, err := client.NewDocFromJSON(data, collection.Schema())
// 		require.NoError(f.t, err)

// 		err = collection.Create(f.ctx, doc)
// 		f.commitTxn()
// 		if tc.ShouldFail {
// 			require.ErrorIs(f.t, err,
// 				NewErrInvalidFieldValue(tc.FieldKind, tc.FieldVal), "test case: %s", tc.Name)
// 		} else {
// 			assertMsg := fmt.Sprintf("test case: %s", tc.Name)
// 			require.NoError(f.t, err, assertMsg)

// 			keyBuilder := newIndexKeyBuilder(f).Col(collection.Name()).Field("field").Doc(doc)
// 			key := keyBuilder.Build()

// 			keyStr := key.ToDS()
// 			data, err := f.txn.Datastore().Get(f.ctx, keyStr)
// 			require.NoError(t, err, assertMsg)
// 			assert.Len(t, data, 0, assertMsg)
// 		}
// 	}
// }

func TestNonUnique_IfIndexedFieldIsNil_StoreItAsNil(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	docJSON, err := json.Marshal(struct {
		Age int `json:"age"`
	}{Age: 44})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(docJSON, f.users.Definition())
	require.NoError(f.t, err)

	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUniqueCreate_ShouldIndexExistingDocs(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	doc1 := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc1, f.users)
	doc2 := f.newUserDoc("Islam", 18, f.users)
	f.saveDocToCollection(doc2, f.users)

	f.createUserCollectionIndexOnName()

	key1 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc1).Build()
	key2 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc2).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key1.ToDS())
	require.NoError(t, err, key1.ToString())
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, key2.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUniqueCreate_IfUponIndexingExistingDocsFetcherFails_ReturnError(t *testing.T) {
	testError := errors.New("test error")

	cases := []struct {
		Name           string
		PrepareFetcher func() fetcher.Fetcher
	}{
		{
			Name: "Fails to init",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Init(
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Unset()
				f.EXPECT().Init(
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to start",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Start(mock.Anything, mock.Anything).Unset()
				f.EXPECT().Start(mock.Anything, mock.Anything).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to fetch next decoded",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().FetchNext(mock.Anything).Unset()
				f.EXPECT().FetchNext(mock.Anything).Return(nil, fetcher.ExecInfo{}, testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to close",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().FetchNext(mock.Anything).Unset()
				f.EXPECT().FetchNext(mock.Anything).Return(nil, fetcher.ExecInfo{}, nil)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(testError)
				return f
			},
		},
	}

	for _, tc := range cases {
		f := newIndexTestFixture(t)
		defer f.db.Close()

		doc := f.newUserDoc("John", 21, f.users)
		f.saveDocToCollection(doc, f.users)

		f.users.(*collection).fetcherFactory = tc.PrepareFetcher
		key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

		_, err := f.users.CreateIndex(f.ctx, getUsersIndexDescOnName())
		require.ErrorIs(t, err, testError, tc.Name)

		_, err = f.txn.Datastore().Get(f.ctx, key.ToDS())
		require.Error(t, err, tc.Name)
	}
}

func TestNonUniqueCreate_IfDatastoreFailsToStoreIndex_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	fieldKeyString := core.DataStoreKey{
		CollectionRootID: f.users.Description().RootID,
	}.WithDocID(doc.ID().String()).
		WithFieldID("1").
		WithValueFlag().
		ToString()

	invalidKeyString := fieldKeyString + "/doesn't matter/"

	// Insert an invalid key within the document prefix, this will generate an error within the fetcher.
	err := f.db.multistore.Datastore().Put(f.ctx, ipfsDatastore.NewKey(invalidKeyString), []byte("doesn't matter"))
	require.NoError(f.t, err)

	_, err = f.users.CreateIndex(f.ctx, getUsersIndexDescOnName())
	require.ErrorIs(f.t, err, core.ErrFailedToGetFieldIdOfKey)
}

func TestNonUniqueDrop_ShouldDeleteStoredIndexedFields(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.addUsersCollection()
	_, err := f.createCollectionIndexFor(users.Name().Value(), getUsersIndexDescOnName())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name().Value(), getUsersIndexDescOnAge())
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name().Value(), getUsersIndexDescOnWeight())
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newUserDoc("John", 21, users), users)
	f.saveDocToCollection(f.newUserDoc("Islam", 23, users), users)

	products := f.getProductsCollectionDesc()
	_, err = f.createCollectionIndexFor(products.Name().Value(), getProductsIndexDescOnCategory())
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newProdDoc(1, 55, "games", products), products)

	userNameKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Build()
	userAgeKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName).Build()
	userWeightKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersWeightFieldName).Build()
	prodCatKey := newIndexKeyBuilder(f).Col(productsColName).Fields(productsCategoryFieldName).Build()

	err = f.dropIndex(usersColName, testUsersColIndexAge)
	require.NoError(f.t, err)

	assert.Len(t, f.getPrefixFromDataStore(userNameKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(userAgeKey.ToString()), 0)
	assert.Len(t, f.getPrefixFromDataStore(userWeightKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(prodCatKey.ToString()), 1)
}

func TestNonUniqueUpdate_ShouldDeleteOldValueAndStoreNewOne(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
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

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	for _, tc := range cases {
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

		err := doc.Set(usersNameFieldName, tc.NewValue)
		require.NoError(t, err)
		err = tc.Exec(doc)
		require.NoError(t, err)
		f.commitTxn()

		newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.Error(t, err)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.NoError(t, err)
	}
}

func TestNonUniqueUpdate_IfFailsToReadIndexDescription_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	err := doc.Set(usersNameFieldName, "Islam")
	require.NoError(t, err)

	// retrieve the collection without index cached
	ctx := SetContextTxn(f.ctx, f.txn)
	usersCol, err := f.db.getCollectionByName(ctx, usersColName)
	require.NoError(t, err)

	testErr := errors.New("test error")

	mockedTxn := f.mockTxn()
	mockedTxn.MockSystemstore = mocks.NewDSReaderWriter(t)
	mockedTxn.MockSystemstore.EXPECT().Query(mock.Anything, mock.Anything).Return(nil, testErr)
	mockedTxn.EXPECT().Systemstore().Unset()
	mockedTxn.EXPECT().Systemstore().Return(mockedTxn.MockSystemstore)
	mockedTxn.MockDatastore.EXPECT().Get(mock.Anything, mock.Anything).Unset()
	mockedTxn.MockDatastore.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte{}, nil)

	usersCol.(*collection).fetcherFactory = func() fetcher.Fetcher {
		return fetcherMocks.NewStubbedFetcher(t)
	}
	ctx = SetContextTxn(f.ctx, mockedTxn)
	err = usersCol.Update(ctx, doc)
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
				f.EXPECT().Init(
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Unset()
				f.EXPECT().Init(
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to start",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().Start(mock.Anything, mock.Anything).Unset()
				f.EXPECT().Start(mock.Anything, mock.Anything).Return(testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to fetch next decoded",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().FetchNext(mock.Anything).Unset()
				f.EXPECT().FetchNext(mock.Anything).Return(nil, fetcher.ExecInfo{}, testError)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(nil)
				return f
			},
		},
		{
			Name: "Fails to close",
			PrepareFetcher: func() fetcher.Fetcher {
				f := fetcherMocks.NewStubbedFetcher(t)
				f.EXPECT().FetchNext(mock.Anything).Unset()
				// By default the stubbed fetcher returns an empty, invalid document
				// here we need to make sure it reaches the Close call by overriding that default.
				f.EXPECT().FetchNext(mock.Anything).Maybe().Return(nil, fetcher.ExecInfo{}, nil)
				f.EXPECT().Close().Unset()
				f.EXPECT().Close().Return(testError)
				return f
			},
		},
	}

	for _, tc := range cases {
		t.Log(tc.Name)

		f := newIndexTestFixture(t)
		defer f.db.Close()
		f.createUserCollectionIndexOnName()

		doc := f.newUserDoc("John", 21, f.users)
		f.saveDocToCollection(doc, f.users)

		f.users.(*collection).fetcherFactory = tc.PrepareFetcher
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

		err := doc.Set(usersNameFieldName, "Islam")
		require.NoError(t, err, tc.Name)
		err = f.users.Update(f.ctx, doc)
		require.Error(t, err, tc.Name)

		newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.NoError(t, err, tc.Name)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.Error(t, err, tc.Name)
	}
}

func TestNonUniqueUpdate_IfFailsToUpdateIndex_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)
	f.commitTxn()

	validKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName).Doc(doc).Build()
	err := f.txn.Datastore().Delete(f.ctx, validKey.ToDS())
	require.NoError(f.t, err)
	f.commitTxn()

	err = doc.Set(usersAgeFieldName, 23)
	require.NoError(t, err)
	err = f.users.Update(f.ctx, doc)
	require.ErrorIs(t, err, ErrCorruptedIndex)
}

func TestNonUniqueUpdate_ShouldPassToFetcherOnlyRelevantFields(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()
	f.createUserCollectionIndexOnAge()

	f.users.(*collection).fetcherFactory = func() fetcher.Fetcher {
		f := fetcherMocks.NewStubbedFetcher(t)
		f.EXPECT().Init(
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Unset()
		f.EXPECT().Init(
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
			RunAndReturn(func(
				ctx context.Context,
				identity immutable.Option[acpIdentity.Identity],
				txn datastore.Txn,
				acp immutable.Option[acp.ACP],
				col client.Collection,
				fields []client.FieldDefinition,
				filter *mapper.Filter,
				mapping *core.DocumentMapping,
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
	doc := f.newUserDoc("John", 21, f.users)
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
				ds.Has(mock.Anything, mock.Anything).Maybe().Return(true, nil)
				ds.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)
			},
		},
		{
			Name: "Store new value",
			StubDataStore: func(ds *mocks.DSReaderWriter_Expecter) {
				ds.Delete(mock.Anything, mock.Anything).Maybe().Return(nil)
				ds.Get(mock.Anything, mock.Anything).Maybe().Return([]byte{}, nil)
				ds.Has(mock.Anything, mock.Anything).Maybe().Return(true, nil)
				ds.Put(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(testErr)
			},
		},
	}

	for _, tc := range cases {
		t.Log(tc.Name)

		f := newIndexTestFixture(t)
		defer f.db.Close()
		f.createUserCollectionIndexOnName()

		doc := f.newUserDoc("John", 21, f.users)
		err := doc.Set(usersNameFieldName, "Islam")
		require.NoError(t, err)

		encodedDoc := shimEncodedDocument{
			key:             []byte(doc.ID().String()),
			schemaVersionID: f.users.Schema().VersionID,
		}

		f.users.(*collection).fetcherFactory = func() fetcher.Fetcher {
			df := fetcherMocks.NewStubbedFetcher(t)
			df.EXPECT().FetchNext(mock.Anything).Unset()
			df.EXPECT().FetchNext(mock.Anything).Return(&encodedDoc, fetcher.ExecInfo{}, nil)
			return df
		}

		mockedTxn := f.mockTxn()
		mockedTxn.MockDatastore = mocks.NewDSReaderWriter(f.t)
		tc.StubDataStore(mockedTxn.MockDatastore.EXPECT())
		mockedTxn.EXPECT().Datastore().Unset()
		mockedTxn.EXPECT().Datastore().Return(mockedTxn.MockDatastore).Maybe()

		ctx := SetContextTxn(f.ctx, mockedTxn)
		err = f.users.Update(ctx, doc)
		require.ErrorIs(t, err, testErr)
	}
}

func TestNonUpdate_IfIndexedFieldWasNil_ShouldDeleteIt(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnName()

	docJSON, err := json.Marshal(struct {
		Age int `json:"age"`
	}{Age: 44})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(docJSON, f.users.Definition())
	require.NoError(f.t, err)

	f.saveDocToCollection(doc, f.users)

	oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

	err = doc.Set(usersNameFieldName, "John")
	require.NoError(f.t, err)

	err = f.users.Update(f.ctx, doc)
	require.NoError(f.t, err)
	f.commitTxn()

	newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Doc(doc).Build()

	_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
	require.NoError(t, err)
	_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
	require.Error(t, err)
}

type shimEncodedDocument struct {
	key             []byte
	schemaVersionID string
	status          client.DocumentStatus
	properties      map[client.FieldDefinition]any
}

var _ fetcher.EncodedDocument = (*shimEncodedDocument)(nil)

func (encdoc *shimEncodedDocument) ID() []byte {
	return encdoc.key
}

func (encdoc *shimEncodedDocument) SchemaVersionID() string {
	return encdoc.schemaVersionID
}

func (encdoc *shimEncodedDocument) Status() client.DocumentStatus {
	return encdoc.status
}

func (encdoc *shimEncodedDocument) Properties(onlyFilterProps bool) (map[client.FieldDefinition]any, error) {
	return encdoc.properties, nil
}

func (encdoc *shimEncodedDocument) Reset() {
	encdoc.key = nil
	encdoc.schemaVersionID = ""
	encdoc.status = 0
	encdoc.properties = map[client.FieldDefinition]any{}
}

func TestUniqueCreate_ShouldIndexExistingDocs(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	doc1 := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc1, f.users)
	doc2 := f.newUserDoc("Islam", 18, f.users)
	f.saveDocToCollection(doc2, f.users)

	f.createUserCollectionUniqueIndexOnName()

	key1 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Unique().Doc(doc1).Build()
	key2 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Unique().Doc(doc2).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key1.ToDS())
	require.NoError(t, err, key1.ToString())
	assert.Equal(t, data, []byte(doc1.ID().String()))
	data, err = f.txn.Datastore().Get(f.ctx, key2.ToDS())
	require.NoError(t, err)
	assert.Equal(t, data, []byte(doc2.ID().String()))
}

func TestUnique_IfIndexedFieldIsNil_StoreItAsNil(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionUniqueIndexOnName()

	docJSON, err := json.Marshal(struct {
		Age int `json:"age"`
	}{Age: 44})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(docJSON, f.users.Definition())
	require.NoError(f.t, err)

	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Unique().Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestUniqueDrop_ShouldDeleteStoredIndexedFields(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.addUsersCollection()
	_, err := f.createCollectionIndexFor(users.Name().Value(), makeUnique(getUsersIndexDescOnName()))
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name().Value(), makeUnique(getUsersIndexDescOnAge()))
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newUserDoc("John", 21, users), users)
	f.saveDocToCollection(f.newUserDoc("Islam", 23, users), users)

	userNameKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Build()
	userAgeKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName).Build()

	err = f.dropIndex(usersColName, testUsersColIndexAge)
	require.NoError(f.t, err)

	assert.Len(t, f.getPrefixFromDataStore(userNameKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(userAgeKey.ToString()), 0)
}

func TestUniqueUpdate_ShouldDeleteOldValueAndStoreNewOne(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionUniqueIndexOnName()

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

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	for _, tc := range cases {
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Unique().Doc(doc).Build()

		err := doc.Set(usersNameFieldName, tc.NewValue)
		require.NoError(t, err)
		err = tc.Exec(doc)
		require.NoError(t, err)
		f.commitTxn()

		newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName).Unique().Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.Error(t, err)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.NoError(t, err)
	}
}

func TestCompositeCreate_ShouldIndexExistingDocs(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	doc1 := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc1, f.users)
	doc2 := f.newUserDoc("Islam", 18, f.users)
	f.saveDocToCollection(doc2, f.users)

	f.createUserCollectionIndexOnNameAndAge()

	key1 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).Doc(doc1).Build()
	key2 := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).Doc(doc2).Build()

	ds := f.txn.Datastore()
	data, err := ds.Get(f.ctx, key1.ToDS())
	require.NoError(t, err, key1.ToString())
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, key2.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestComposite_IfIndexedFieldIsNil_StoreItAsNil(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnNameAndAge()

	docJSON, err := json.Marshal(struct {
		Age int `json:"age"`
	}{Age: 44})
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(docJSON, f.users.Definition())
	require.NoError(f.t, err)

	f.saveDocToCollection(doc, f.users)

	key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).
		Doc(doc).Build()

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestUniqueComposite_IfNilUpdateToValue_ShouldUpdateIndexStored(t *testing.T) {
	testCases := []struct {
		Name   string
		Doc    string
		Update string
		Action func(*client.Document) error
	}{
		{
			Name:   "/nil/44/docID -> /John/44",
			Doc:    `{"age": 44}`,
			Update: `{"name": "John", "age": 44}`,
			Action: func(doc *client.Document) error {
				return doc.Set(usersNameFieldName, "John")
			},
		},
		{
			Name:   "/Islam/33 -> /Islam/nil/docID",
			Doc:    `{"name": "Islam", "age": 33}`,
			Update: `{"name": "Islam", "age": null}`,
			Action: func(doc *client.Document) error {
				return doc.Set(usersAgeFieldName, nil)
			},
		},
		{
			Name:   "/Andy/nil/docID -> /nil/22/docID",
			Doc:    `{"name": "Andy"}`,
			Update: `{"name": null, "age": 22}`,
			Action: func(doc *client.Document) error {
				return errors.Join(doc.Set(usersNameFieldName, nil), doc.Set(usersAgeFieldName, 22))
			},
		},
		{
			Name:   "/nil/55/docID -> /nil/nil/docID",
			Doc:    `{"age": 55}`,
			Update: `{"name": null, "age": null}`,
			Action: func(doc *client.Document) error {
				return doc.Set(usersAgeFieldName, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			f := newIndexTestFixture(t)
			defer f.db.Close()

			indexDesc := makeUnique(addFieldToIndex(getUsersIndexDescOnName(), usersAgeFieldName))
			_, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
			require.NoError(f.t, err)
			f.commitTxn()

			doc, err := client.NewDocFromJSON([]byte(tc.Doc), f.users.Definition())
			require.NoError(f.t, err)

			f.saveDocToCollection(doc, f.users)

			oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).
				Doc(doc).Unique().Build()

			require.NoError(t, doc.SetWithJSON([]byte(tc.Update)))

			newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).
				Doc(doc).Unique().Build()

			require.NoError(t, f.users.Update(f.ctx, doc), tc.Name)
			f.commitTxn()

			_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
			require.Error(t, err, oldKey.ToString(), oldKey.ToDS(), tc.Name)
			_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
			require.NoError(t, err, newKey.ToString(), newKey.ToDS(), tc.Name)
		})
	}
}

func TestCompositeDrop_ShouldDeleteStoredIndexedFields(t *testing.T) {
	f := newIndexTestFixtureBare(t)
	users := f.addUsersCollection()
	_, err := f.createCollectionIndexFor(users.Name().Value(), addFieldToIndex(getUsersIndexDescOnName(), usersAgeFieldName))
	require.NoError(f.t, err)
	_, err = f.createCollectionIndexFor(users.Name().Value(), addFieldToIndex(getUsersIndexDescOnAge(), usersWeightFieldName))
	require.NoError(f.t, err)
	f.commitTxn()

	f.saveDocToCollection(f.newUserDoc("John", 21, users), users)
	f.saveDocToCollection(f.newUserDoc("Islam", 23, users), users)

	userNameAgeKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).Build()
	userAgeWeightKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersAgeFieldName, usersWeightFieldName).Build()

	err = f.dropIndex(usersColName, testUsersColIndexAge)
	require.NoError(f.t, err)

	assert.Len(t, f.getPrefixFromDataStore(userNameAgeKey.ToString()), 2)
	assert.Len(t, f.getPrefixFromDataStore(userAgeWeightKey.ToString()), 0)
}

func TestCompositeUpdate_ShouldDeleteOldValueAndStoreNewOne(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()
	f.createUserCollectionIndexOnNameAndAge()

	cases := []struct {
		Name     string
		Field    string
		NewValue any
		Exec     func(doc *client.Document) error
	}{
		{
			Name:     "update first",
			NewValue: "Islam",
			Field:    usersNameFieldName,
			Exec: func(doc *client.Document) error {
				return f.users.Update(f.ctx, doc)
			},
		},
		{
			Name:     "save first",
			NewValue: "Andy",
			Field:    usersNameFieldName,
			Exec: func(doc *client.Document) error {
				return f.users.Save(f.ctx, doc)
			},
		},
		{
			Name:     "update second",
			NewValue: 33,
			Field:    usersAgeFieldName,
			Exec: func(doc *client.Document) error {
				return f.users.Update(f.ctx, doc)
			},
		},
		{
			Name:     "save second",
			NewValue: 36,
			Field:    usersAgeFieldName,
			Exec: func(doc *client.Document) error {
				return f.users.Save(f.ctx, doc)
			},
		},
	}

	doc := f.newUserDoc("John", 21, f.users)
	f.saveDocToCollection(doc, f.users)

	for _, tc := range cases {
		oldKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).Doc(doc).Build()

		err := doc.Set(tc.Field, tc.NewValue)
		require.NoError(t, err)
		err = tc.Exec(doc)
		require.NoError(t, err)
		f.commitTxn()

		newKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNameFieldName, usersAgeFieldName).Doc(doc).Build()

		_, err = f.txn.Datastore().Get(f.ctx, oldKey.ToDS())
		require.Error(t, err)
		_, err = f.txn.Datastore().Get(f.ctx, newKey.ToDS())
		require.NoError(t, err)
		f.commitTxn()
	}
}

func TestArrayIndex_IfDocIsAdded_ShouldIndexAllArrayElements(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnNumbers()

	numbersArray := []int{1, 2, 3}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	for _, num := range numbersArray {
		key := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNumbersFieldName).
			ArrayFieldVal(usersNumbersFieldName, num).Doc(doc).Build()

		data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
		require.NoError(t, err)
		assert.Len(t, data, 0)
	}
}

func TestArrayIndex_IfDocIsDeleted_ShouldRemoveIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnNumbers()

	numbersArray := []int{1, 2, 3}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	userNumbersKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNumbersFieldName).Build()
	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), len(numbersArray))

	f.deleteDocFromCollection(doc.ID(), f.users)

	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), 0)
}

func TestArrayIndex_IfDocIsDeletedButOneArrayElementHasNoIndexRecord_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	f.createUserCollectionIndexOnNumbers()

	numbersArray := []int{1, 2, 3}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	userNumbersKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNumbersFieldName).
		ArrayFieldVal(usersNumbersFieldName, 2).Doc(doc).Build()

	err := f.txn.Datastore().Delete(f.ctx, userNumbersKey.ToDS())
	require.NoError(t, err)
	f.commitTxn()

	res, err := f.users.Delete(f.ctx, doc.ID())
	require.Error(f.t, err)
	require.False(f.t, res)
}

func TestArrayIndex_With2ArrayFieldsIfDocIsDeleted_ShouldRemoveIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexDesc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: usersNumbersFieldName},
			{Name: usersHobbiesFieldName},
		},
	}

	_, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)

	numbersArray := []int{1, 2}
	hobbiesArray := []string{"reading", "swimming"}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray, Hobbies: hobbiesArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	userNumbersKey := newIndexKeyBuilder(f).Col(usersColName).
		Fields(usersNumbersFieldName, usersHobbiesFieldName).Build()
	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), len(numbersArray)*len(hobbiesArray))

	f.deleteDocFromCollection(doc.ID(), f.users)

	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), 0)
}

func TestArrayIndex_With2ArrayFieldsIfDocIsDeletedButOneArrayElementHasNoIndexRecord_ShouldRemoveIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexDesc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: usersNumbersFieldName},
			{Name: usersHobbiesFieldName},
		},
	}

	_, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)

	numbersArray := []int{1, 2}
	hobbiesArray := []string{"reading", "swimming"}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray, Hobbies: hobbiesArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	userNumbersKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNumbersFieldName, usersHobbiesFieldName).
		ArrayFieldVal(usersNumbersFieldName, 2).ArrayFieldVal(usersHobbiesFieldName, "swimming").Doc(doc).Build()

	err = f.txn.Datastore().Delete(f.ctx, userNumbersKey.ToDS())
	require.NoError(t, err)
	f.commitTxn()

	res, err := f.users.Delete(f.ctx, doc.ID())
	require.Error(f.t, err)
	require.False(f.t, res)
}

func TestArrayIndex_WithUniqueIndexIfDocIsDeleted_ShouldRemoveIndex(t *testing.T) {
	f := newIndexTestFixture(t)
	defer f.db.Close()

	indexDesc := client.IndexDescription{
		Unique: true,
		Fields: []client.IndexedFieldDescription{
			{Name: usersNumbersFieldName},
		},
	}

	_, err := f.createCollectionIndexFor(f.users.Name().Value(), indexDesc)
	require.NoError(f.t, err)

	numbersArray := []int{1, 2, 3}
	doc := f.newCustomUserDoc(userDoc{Name: "John", Numbers: numbersArray}, f.users)
	f.saveDocToCollection(doc, f.users)

	userNumbersKey := newIndexKeyBuilder(f).Col(usersColName).Fields(usersNumbersFieldName).Unique().Build()
	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), len(numbersArray))

	f.deleteDocFromCollection(doc.ID(), f.users)

	assert.Len(t, f.getPrefixFromDataStore(userNumbersKey.ToString()), 0)
}
