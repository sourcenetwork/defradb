package db

import (
	"context"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
)

type CollectionIndex interface {
	Save(core.DataStoreKey, client.Value) error
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

func (c *collectionSimpleIndex) Save(core.DataStoreKey, client.Value) error {
	return nil
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

func generateIndexName(col client.Collection, fields []client.IndexedFieldDescription) string {
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
	if desc.Name == "" {
		desc.Name = generateIndexName(c, desc.Fields)
	}
	colIndex := NewCollectionIndex(c, desc)
	return colIndex, nil
}
