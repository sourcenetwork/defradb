package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

type CollectionIndex interface {
	Save(context.Context, datastore.Txn, *client.Document) error
	Name() string
	Description() client.IndexDescription
}

func getFieldValConverter(kind client.FieldKind) func(any) ([]byte, error) {
	switch kind {
	case client.FieldKind_STRING:
		return func(val any) ([]byte, error) {
			return []byte(val.(string)), nil
		}
	case client.FieldKind_INT:
		return func(val any) ([]byte, error) {
			intVal, ok := val.(int64)
			if !ok {
				return nil, errors.New("invalid int value")
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}
	case client.FieldKind_FLOAT:
		return func(val any) ([]byte, error) {
			floatVal, ok := val.(float64)
			if !ok {
				return nil, errors.New("invalid float value")
			}
			return []byte(strconv.FormatFloat(floatVal, 'f', -1, 64)), nil
		}
	case client.FieldKind_BOOL:
		return func(val any) ([]byte, error) {
			boolVal, ok := val.(bool)
			if !ok {
				return nil, errors.New("invalid bool value")
			}
			var intVal int64 = 0
			if boolVal {
				intVal = 1
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}
	case client.FieldKind_DATETIME:
		return func(val any) ([]byte, error) {
			timeStrVal, ok := val.(string)
			if !ok {
				return nil, errors.New("invalid datetime value")
			}
			_, err := time.Parse(time.RFC3339, timeStrVal)
			if err != nil {
				return nil, err
			}
			return []byte(timeStrVal), nil
		}
	default:
		panic("there is no test for this case")
	}
}

func NewCollectionIndex(
	collection client.Collection,
	desc client.IndexDescription,
) CollectionIndex {
	index := &collectionSimpleIndex{collection: collection, desc: desc}
	schema := collection.Description().Schema
	fieldID := schema.GetFieldKey(desc.Fields[0].Name)
	field := schema.Fields[fieldID]
	index.convertFunc = getFieldValConverter(field.Kind)
	return index
}

type collectionSimpleIndex struct {
	collection  client.Collection
	desc        client.IndexDescription
	convertFunc func(any) ([]byte, error)
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (i *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	indexedFieldName := i.desc.Fields[0].Name
	fieldVal, err := doc.Get(indexedFieldName)
	if err != nil {
		return nil
	}

	data, err := i.convertFunc(fieldVal)
	if err != nil {
		return NewErrCanNotIndexInvalidFieldValue(err)
	}
	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = strconv.Itoa(int(i.collection.ID()))
	indexDataStoreKey.IndexID = strconv.Itoa(int(i.desc.ID))
	indexDataStoreKey.FieldValues = []string{string(data), doc.Key().String()}
	keyStr := indexDataStoreKey.ToDS()
	err = txn.Datastore().Put(ctx, keyStr, []byte{})
	if err != nil {
		return NewErrFailedToStoreIndexedField(indexDataStoreKey.IndexID, err)
	}
	return nil
}

func (i *collectionSimpleIndex) Name() string {
	return i.desc.Name
}

func (i *collectionSimpleIndex) Description() client.IndexDescription {
	return i.desc
}

func validateIndexDescription(desc client.IndexDescription) error {
	if desc.ID != 0 {
		return NewErrNonZeroIndexIDProvided(desc.ID)
	}
	if len(desc.Fields) == 0 {
		return ErrIndexMissingFields
	}
	if len(desc.Fields) == 1 && desc.Fields[0].Direction == client.Descending {
		return ErrIndexSingleFieldWrongDirection
	}
	for i := range desc.Fields {
		if desc.Fields[i].Name == "" {
			return ErrIndexFieldMissingName
		}
		if desc.Fields[i].Direction == "" {
			desc.Fields[i].Direction = client.Ascending
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
	sb.WriteString(col.Name())
	sb.WriteByte('_')
	sb.WriteString(fields[0].Name)
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
	if c.isIndexCached {
		c.indexes = append(c.indexes, index)
	}
	return index.Description(), nil
}

func (c *collection) DropIndex(ctx context.Context, indexName string) error {
	key := core.NewCollectionIndexKey(c.Name(), indexName)

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	err = txn.Systemstore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if c.isIndexCached {
		for i := range c.indexes {
			if c.indexes[i].Name() == indexName {
				c.indexes = append(c.indexes[:i], c.indexes[i+1:]...)
				break
			}
		}
	}
	return nil
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

func (c *collection) getIndexes(ctx context.Context, txn datastore.Txn) ([]CollectionIndex, error) {
	if c.isIndexCached {
		return c.indexes, nil
	}

	prefix := core.NewCollectionIndexKey(c.Name(), "")
	if txn == nil {
		var err error
		txn, err = c.getTxn(ctx, true)
		if err != nil {
			return nil, err
		}
	}
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	indexes := make([]CollectionIndex, 0)
	for res := range q.Next() {
		if res.Error != nil {
			return nil, res.Error
		}

		var indexDesc client.IndexDescription
		err = json.Unmarshal(res.Value, &indexDesc)
		if err != nil {
			return nil, NewErrInvalidStoredIndex(err)
		}
		colIndex := NewCollectionIndex(c, indexDesc)
		indexes = append(indexes, colIndex)
	}

	c.indexes = indexes
	c.isIndexCached = true
	return indexes, nil
}

func (c *collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	indexes, err := c.getIndexes(ctx, nil)
	if err != nil {
		return nil, err
	}
	indexDescriptions := make([]client.IndexDescription, 0, len(indexes))
	for _, index := range indexes {
		indexDescriptions = append(indexDescriptions, index.Description())
	}

	return indexDescriptions, nil
}

func (c *collection) createIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	err := validateIndexDescription(desc)
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

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}

	colSeq, err := c.db.getSequence(ctx, txn, fmt.Sprintf("%s/%d", core.COLLECTION_INDEX, c.ID()))
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next(ctx, txn)
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)

	buf, err := json.Marshal(desc)
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
