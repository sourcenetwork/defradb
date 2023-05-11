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

func (f *indexTestFixture) getNonUniqueIndexKey() core.IndexDataStoreKey {
	key := core.IndexDataStoreKey{
		CollectionID: strconv.Itoa(int(f.users.ID())),
		IndexID:      strconv.Itoa(1),
	}
	return key
}

func (f *indexTestFixture) getNonUniqueDocIndexKey(
	doc *client.Document,
	fieldName string,
) core.IndexDataStoreKey {
	key := f.getNonUniqueIndexKey()

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

	key := f.getNonUniqueDocIndexKey(doc, "name")

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

func TestNonUnique_IfFailsToStoredIndexedDoc_Error(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.newUserDoc("John", 21)
	key := f.getNonUniqueDocIndexKey(doc, "name")

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

	key := f.getNonUniqueIndexKey()
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

	key := f.getNonUniqueDocIndexKey(doc, "age")

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}
