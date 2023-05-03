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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

const (
	usersColName    = "Users"
	productsColName = "Products"
)

type indexTestFixture struct {
	ctx        context.Context
	db         *implicitTxnDB
	txn        datastore.Txn
	collection client.Collection
	t          *testing.T
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
					Name: "name",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "age",
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "weight",
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
					Name: "price",
					Kind: client.FieldKind_FLOAT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "description",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}
}

func newIndexTestFixture(t *testing.T) *indexTestFixture {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)
	txn, err := db.NewTxn(ctx, false)
	assert.NoError(t, err)

	f := &indexTestFixture{
		ctx: ctx,
		db:  db,
		txn: txn,
		t:   t,
	}
	f.collection = f.createCollection(getUsersCollectionDesc())
	return f
}

func (f *indexTestFixture) createCollectionIndex(
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	return f.createCollectionIndexFor(f.collection.Name(), desc)
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

func TestCreateIndex_IfValidInput_CreateIndex(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Name: "some_index_name",
		Fields: []client.IndexedFieldDescription{
			{Name: "name", Direction: client.Ascending},
		},
	}
	resultDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, resultDesc.Name, desc.Name)
	assert.Equal(t, resultDesc, desc)
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
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, newDesc.Fields[0].Direction, client.Ascending)
}

func TestCreateIndex_IfNameIsNotSpecified_GenerateWithLowerCase(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Name: "",
		Fields: []client.IndexedFieldDescription{
			{Name: "Name", Direction: client.Ascending},
		},
	}
	f.collection.Description().Schema.Fields[1].Name = "Name"
	newDesc, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)
	assert.Equal(t, newDesc.Name, "users_name_ASC")
}

func TestCreateIndex_IfSingleFieldInDescOrder_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{
			{Name: "name", Direction: client.Descending},
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
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	desc2 := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
	}
	_, err := f.createCollectionIndex(desc1)
	assert.NoError(t, err)
	_, err = f.createCollectionIndex(desc2)
	assert.EqualError(t, err, errIndexWithNameAlreadyExists)
}

func TestCreateIndex_IfGeneratedNameMatchesExisting_AddIncrement(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "users_age_ASC"
	desc1 := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	desc2 := client.IndexDescription{
		Name:   name + "_2",
		Fields: []client.IndexedFieldDescription{{Name: "weight"}},
	}
	desc3 := client.IndexDescription{
		Name:   "",
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
	}
	_, err := f.createCollectionIndex(desc1)
	assert.NoError(t, err)
	_, err = f.createCollectionIndex(desc2)
	assert.NoError(t, err)
	newDesc3, err := f.createCollectionIndex(desc3)
	assert.NoError(t, err)
	assert.Equal(t, newDesc3.Name, name+"_3")
}

func TestCreateIndex_ShouldSaveToSystemStorage(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "users_age_ASC"
	desc := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	_, err := f.createCollectionIndex(desc)
	assert.NoError(t, err)

	key := core.NewCollectionIndexKey(f.collection.Name(), name)
	data, err := f.txn.Systemstore().Get(f.ctx, key.ToDS())
	assert.NoError(t, err)
	var deserialized client.IndexDescription
	err = json.Unmarshal(data, &deserialized)
	assert.NoError(t, err)
	assert.Equal(t, deserialized, desc)
}

func TestCreateIndex_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	name := "users_age_ASC"
	desc := client.IndexDescription{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}

	f.db.Close(f.ctx)

	_, err := f.createCollectionIndex(desc)
	assert.Error(t, err)
}

func TestCreateIndex_IfCollectionDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: "price"}},
	}

	_, err := f.createCollectionIndexFor(productsColName, desc)
	assert.Error(t, err)
}

func TestCreateIndex_IfPropertyDoesntExist_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	const prop = "non_existing_property"
	desc := client.IndexDescription{
		Fields: []client.IndexedFieldDescription{{Name: prop}},
	}

	_, err := f.createCollectionIndex(desc)
	assert.ErrorIs(t, err, NewErrNonExistingFieldForIndex(prop))
}

func TestGetIndexes_ShouldReturnListOfAllExistingIndexes(t *testing.T) {
	f := newIndexTestFixture(t)

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.createCollection(getProductsCollectionDesc())
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: "price"}},
	}
	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	indexes, err := f.getAllIndexes()
	assert.NoError(t, err)

	require.Equal(t, len(indexes), 2)
	usersIndexIndex := 0
	if indexes[0].CollectionName != usersColName {
		usersIndexIndex = 1
	}
	assert.Equal(t, indexes[usersIndexIndex].Index, usersIndexDesc)
	assert.Equal(t, indexes[usersIndexIndex].CollectionName, usersColName)
	assert.Equal(t, indexes[1-usersIndexIndex].Index, productsIndexDesc)
	assert.Equal(t, indexes[1-usersIndexIndex].CollectionName, productsColName)
}

func TestGetCollectionIndexes_ShouldReturnListOfCollectionIndexes(t *testing.T) {
	f := newIndexTestFixture(t)

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.createCollection(getProductsCollectionDesc())
	productsIndexDesc := client.IndexDescription{
		Name:   "products_description_index",
		Fields: []client.IndexedFieldDescription{{Name: "price"}},
	}
	_, err = f.createCollectionIndexFor(productsColName, productsIndexDesc)
	assert.NoError(t, err)

	userIndexes, err := f.getCollectionIndexes(usersColName)
	assert.NoError(t, err)
	require.Equal(t, len(userIndexes), 1)
	assert.Equal(t, userIndexes[0], usersIndexDesc)

	productIndexes, err := f.getCollectionIndexes(productsColName)
	assert.NoError(t, err)
	require.Equal(t, len(productIndexes), 1)
	assert.Equal(t, productIndexes[0], productsIndexDesc)
}

func TestGetCollectionIndexes_IfStorageFails_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	usersIndexDesc := client.IndexDescription{
		Name:   "users_name_index",
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	}
	_, err := f.createCollectionIndexFor(usersColName, usersIndexDesc)
	assert.NoError(t, err)

	f.db.Close(f.ctx)

	_, err = f.getCollectionIndexes(usersColName)
	assert.Error(t, err)
}

func TestGetCollectionIndexes_InvalidIndexIsStored_ReturnError(t *testing.T) {
	f := newIndexTestFixture(t)

	indexKey := core.NewCollectionIndexKey(usersColName, "users_name_index")
	err := f.txn.Systemstore().Put(f.ctx, indexKey.ToDS(), []byte("invalid"))
	assert.NoError(t, err)

	_, err = f.getCollectionIndexes(usersColName)
	assert.ErrorIs(t, err, NewErrInvalidStoredIndex(nil))
}
