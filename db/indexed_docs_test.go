package db

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
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

func (f *indexTestFixture) getNonUniqueIndexKey(fieldName string) core.IndexDataStoreKey {
	colDesc := f.users.Description()
	field, ok := colDesc.GetField(fieldName)
	require.True(f.t, ok)

	key := core.IndexDataStoreKey{
		CollectionID: strconv.Itoa(int(f.users.ID())),
		IndexID:      strconv.Itoa(int(field.ID)),
	}
	return key
}

func (f *indexTestFixture) getNonUniqueDocIndexKey(
	doc *client.Document,
	fieldName string,
) core.IndexDataStoreKey {
	key := f.getNonUniqueIndexKey(fieldName)

	fieldVal, err := doc.Get(fieldName)
	require.NoError(f.t, err)
	fieldStrVal, ok := fieldVal.(string)
	require.True(f.t, ok)

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
	mockTxn.EXPECT().Rootstore().Unset()
	expect := mockTxn.MockDatastore.EXPECT()
	expect.Put(mock.Anything, mock.Anything, mock.Anything).Unset()
	expect.Put(mock.Anything, key.ToDS(), mock.Anything).Return(errors.New("error"))
	expect.Put(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := f.users.WithTxn(mockTxn).Create(f.ctx, doc)
	require.Error(f.t, err)
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

	key := f.getNonUniqueIndexKey("name")
	prefixes := f.getPrefixFromDataStore(key.ToString())
	assert.Len(t, prefixes, 0)
}
