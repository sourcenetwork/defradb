package db

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

type CollectionIndex interface {
	Save(context.Context, datastore.Txn, core.DataStoreKey, any) error
	Name() string
	Description() client.IndexDescription
}

func NewCollectionIndex(
	collection client.Collection,
	desc client.IndexDescription,
) CollectionIndex {
	return &collectionSimpleIndex{collection: collection, desc: desc}
}

type collectionSimpleIndex struct {
	collection client.Collection
	desc       client.IndexDescription
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (c *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	key core.DataStoreKey,
	val any,
) error {
	data := val.(string)
	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = strconv.Itoa(int(c.collection.ID()))
	indexDataStoreKey.IndexID = "1"
	indexDataStoreKey.FieldValues = []string{data, key.DocKey}
	return txn.Datastore().Put(ctx, indexDataStoreKey.ToDS(), []byte{})
}

func (c *collectionSimpleIndex) Name() string {
	return c.desc.Name
}

func (c *collectionSimpleIndex) Description() client.IndexDescription {
	return c.desc
}

func validateIndexDescriptionFields(fields []client.IndexedFieldDescription) error {
	if len(fields) == 0 {
		return ErrIndexMissingFields
	}
	if len(fields) == 1 && fields[0].Direction == client.Descending {
		return ErrIndexSingleFieldWrongDirection
	}
	for i := range fields {
		if fields[i].Name == "" {
			return ErrIndexFieldMissingName
		}
		if fields[i].Direction == "" {
			fields[i].Direction = client.Ascending
		}
	}
	return nil
}

func generateIndexName(col client.Collection, fields []client.IndexedFieldDescription, inc int) string {
	sb := strings.Builder{}
	direction := "ASC"
	//if fields[0].Direction == client.Descending {
	//direction = "DESC"
	//}
	sb.WriteString(strings.ToLower(col.Name()))
	sb.WriteByte('_')
	sb.WriteString(strings.ToLower(fields[0].Name))
	sb.WriteByte('_')
	sb.WriteString(direction)
	if inc > 1 {
		sb.WriteByte('_')
		sb.WriteString(strconv.Itoa(inc))
	}
	return sb.String()
}

func (c *collection) CreateIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	index, err := c.createIndex(ctx, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return index.Description(), nil
}

func (c *collection) DropIndex(ctx context.Context, indexName string) error {
	key := core.NewCollectionIndexKey(c.Name(), indexName)

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	return txn.Systemstore().Delete(ctx, key.ToDS())
}

func (c *collection) dropAllIndexes(ctx context.Context) error {
	prefix := core.NewCollectionIndexKey(c.Name(), "")
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	for res := range q.Next() {
		if res.Error != nil {
			return res.Error
		}
		err = txn.Systemstore().Delete(ctx, ds.NewKey(res.Key))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	return nil, nil
}

func (c *collection) createIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	err := validateIndexDescriptionFields(desc.Fields)
	if err != nil {
		return nil, err
	}

	err = c.checkExistingFields(ctx, desc.Fields)
	if err != nil {
		return nil, err
	}

	indexKey, err := c.processIndexName(ctx, &desc)
	if err != nil {
		return nil, err
	}

	buf, err := json.Marshal(desc)
	if err != nil {
		return nil, err
	}

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	err = txn.Systemstore().Put(ctx, indexKey.ToDS(), buf)
	if err != nil {
		return nil, err
	}
	colIndex := NewCollectionIndex(c, desc)
	return colIndex, nil
}

func (c *collection) checkExistingFields(
	ctx context.Context,
	fields []client.IndexedFieldDescription,
) error {
	collectionFields := c.Description().Schema.Fields
	for _, field := range fields {
		found := false
		fieldLower := strings.ToLower(field.Name)
		for _, colField := range collectionFields {
			if fieldLower == strings.ToLower(colField.Name) {
				found = true
				break
			}
		}
		if !found {
			return NewErrNonExistingFieldForIndex(field.Name)
		}
	}
	return nil
}

func (c *collection) processIndexName(
	ctx context.Context,
	desc *client.IndexDescription,
) (core.CollectionIndexKey, error) {
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return core.CollectionIndexKey{}, err
	}

	var indexKey core.CollectionIndexKey
	if desc.Name == "" {
		nameIncrement := 1
		for {
			desc.Name = generateIndexName(c, desc.Fields, nameIncrement)
			indexKey = core.NewCollectionIndexKey(c.Name(), desc.Name)
			exists, err := txn.Systemstore().Has(ctx, indexKey.ToDS())
			if err != nil {
				return core.CollectionIndexKey{}, err
			}
			if !exists {
				break
			}
			nameIncrement++
		}
	} else {
		indexKey = core.NewCollectionIndexKey(c.Name(), desc.Name)
		exists, err := txn.Systemstore().Has(ctx, indexKey.ToDS())
		if err != nil {
			return core.CollectionIndexKey{}, err
		}
		if exists {
			return core.CollectionIndexKey{}, ErrIndexWithNameAlreadyExists
		}
	}
	return indexKey, nil
}
