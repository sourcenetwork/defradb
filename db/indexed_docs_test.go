package db

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userDoc struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Weight float64 `json:"weight"`
}

func (f *indexTestFixture) createUserDoc(name string, age int) *client.Document {
	d := userDoc{Name: name, Age: age, Weight: 154.1}
	data, err := json.Marshal(d)
	require.NoError(f.t, err)

	doc, err := client.NewDocFromJSON(data)
	if err != nil {
		f.t.Error(err)
		return nil
	}
	err = f.collection.Create(f.ctx, doc)
	require.NoError(f.t, err)
	f.txn, err = f.db.NewTxn(f.ctx, false)
	require.NoError(f.t, err)
	return doc
}

func (f *indexTestFixture) getNonUniqueDocIndex(
	doc *client.Document,
	fieldName string,
) core.IndexDataStoreKey {
	colDesc := f.collection.Description()
	field, ok := colDesc.GetField(fieldName)
	require.True(f.t, ok)

	fieldVal, err := doc.Get(fieldName)
	require.NoError(f.t, err)
	fieldStrVal, ok := fieldVal.(string)
	require.True(f.t, ok)

	key := core.IndexDataStoreKey{
		CollectionID: strconv.Itoa(int(f.collection.ID())),
		IndexID:      strconv.Itoa(int(field.ID)),
		FieldValues:  []string{fieldStrVal, doc.Key().String()},
	}
	return key
}

func TestNonUnique_IfDocIsAdded_ShouldBeIndexed(t *testing.T) {
	f := newIndexTestFixture(t)
	f.createUserCollectionIndexOnName()

	doc := f.createUserDoc("John", 21)

	key := f.getNonUniqueDocIndex(doc, "name")

	data, err := f.txn.Datastore().Get(f.ctx, key.ToDS())
	require.NoError(t, err)
	assert.Len(t, data, 0)
}
