package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func (f *indexTestFixture) saveToUsers(doc *client.Document) {
	err := f.users.Create(f.ctx, doc)
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

func (f *indexTestFixture) getNonUniqueIndexKey(colName string) core.IndexDataStoreKey {
	cols, err := f.db.getAllCollections(f.ctx, f.txn)
	require.NoError(f.t, err)
	colID := -1
	for _, col := range cols {
		if col.Name() == colName {
			colID = int(col.ID())
			break
		}
	}
	if colID == -1 {
		panic(errors.New("collection not found"))
	}

	key := core.IndexDataStoreKey{
		CollectionID: strconv.Itoa(colID),
		IndexID:      strconv.Itoa(1),
	}
	return key
}

func (f *indexTestFixture) getNonUniqueDocIndexKey(
	doc *client.Document,
	colName, fieldName string,
) core.IndexDataStoreKey {
	key := f.getNonUniqueIndexKey(colName)

	fieldVal, err := doc.Get(fieldName)
	require.NoError(f.t, err)
	fieldStrVal := fmt.Sprintf("%v", fieldVal)

	key.FieldValues = []string{fieldStrVal, doc.Key().String()}

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

	indexOnNameDescData, err := json.Marshal(getUsersIndexDescOnName())
	require.NoError(f.t, err)

	systemStoreOn := mockedTxn.MockSystemstore.EXPECT()

	colIndexKey := core.NewCollectionIndexKey(f.users.Description().Name, "")
	matchPrefixFunc := func(q query.Query) bool { return q.Prefix == colIndexKey.ToDS().String() }

	systemStoreOn.Query(mock.Anything, mock.Anything).Unset()
	systemStoreOn.Query(mock.Anything, mock.MatchedBy(matchPrefixFunc)).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t, indexOnNameDescData), nil)
	systemStoreOn.Query(mock.Anything, mock.Anything).Maybe().
		Return(mocks.NewQueryResultsWithValues(f.t), nil)

	systemStoreOn.Get(mock.Anything, mock.Anything).Unset()
	colIndexOnNameKey := core.NewCollectionIndexKey(f.users.Description().Name, testUsersColIndexName)
	systemStoreOn.Get(mock.Anything, colIndexOnNameKey.ToDS()).Return(indexOnNameDescData, nil)
	systemStoreOn.Get(mock.Anything, mock.Anything).Return([]byte{}, nil)

	f.txn = mockedTxn
	return mockedTxn
}

func TestNonUnique_IfDocIsAdded_ShouldBeIndexed(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	f.saveToUsers(doc)

	key := f.getNonUniqueDocIndexKey(doc, usersColName, usersNameFieldName)

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfFailsToStoredIndexedDoc_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	key := f.getNonUniqueDocIndexKey(doc, usersColName, usersNameFieldName)

	mockTxn := f.mockTxn()

	dataStoreOn := mockTxn.MockDatastore.EXPECT()
	dataStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Unset()
	dataStoreOn.Put(mock.Anything, key.ToDS(), mock.Anything).Return(errors.New("error"))
	dataStoreOn.Put(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(f.t, err, NewErrFailedToStoreIndexedField("name", nil))
}

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

	key := f.getNonUniqueIndexKey(usersColName)
	prefixes := f.getPrefixFromDataStore(key.ToString())
	assert.Len(t, prefixes, 0)
}

func TestNonUnique_IfSystemStorageHasInvalidIndexDescription_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	indexDesc := f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)

	mockTxn := f.mockTxn()
	systemStoreOn := mockTxn.MockSystemstore.EXPECT()
	systemStoreOn.Get(mock.Anything, mock.Anything).Unset()
	colIndexKey := core.NewCollectionIndexKey(f.users.Description().Name, indexDesc.Name)
	systemStoreOn.Get(mock.Anything, colIndexKey.ToDS()).Return([]byte("invalid"), nil)
	systemStoreOn.Get(mock.Anything, mock.Anything).Return([]byte{}, nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(t, err, NewErrInvalidStoredIndex(nil))
}

func TestNonUnique_IfSystemStorageFailsToReadIndexDesc_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)

	mockTxn := f.mockTxn()
	systemStoreOn := mockTxn.MockSystemstore.EXPECT()
	systemStoreOn.Get(mock.Anything, mock.Anything).Unset()
	colIndexKey := core.NewCollectionIndexKey(f.users.Description().Name, testUsersColIndexName)
	systemStoreOn.Get(mock.Anything, colIndexKey.ToDS()).Return([]byte{}, errors.New("error"))
	systemStoreOn.Get(mock.Anything, mock.Anything).Return([]byte{}, nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.ErrorIs(t, err, NewErrFailedToReadStoredIndexDesc(nil))
}

func TestNonUnique_IfIndexIntField_StoreIt(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnAge()

	doc := f.newUserDoc("John", 21)
	f.saveToUsers(doc)

	key := f.getNonUniqueDocIndexKey(doc, usersColName, usersAgeFieldName)

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

	userDocKey := f.getNonUniqueDocIndexKey(userDoc, usersColName, usersNameFieldName)
	prodDocKey := f.getNonUniqueDocIndexKey(prodDoc, productsColName, productsCategoryFieldName)

	data, err := f.txn.Datastore().Get(f.ctx, userDocKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
	data, err = f.txn.Datastore().Get(f.ctx, prodDocKey.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}
