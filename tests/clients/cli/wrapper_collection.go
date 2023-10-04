// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	cmd    *cliWrapper
	desc   client.CollectionDescription
	schema client.SchemaDescription
}

func (c *Collection) Description() client.CollectionDescription {
	return c.desc
}

func (c *Collection) Name() string {
	return c.desc.Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.schema
}

func (c *Collection) ID() uint32 {
	return c.desc.ID
}

func (c *Collection) SchemaID() string {
	return c.schema.SchemaID
}

func (c *Collection) Create(ctx context.Context, doc *client.Document) error {
	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.desc.Name)

	// We must call this here, else the doc key on the given object will not match
	// that of the document saved in the database
	err := doc.RemapAliasFieldsAndDockey(c.Schema().Fields)
	if err != nil {
		return err
	}
	document, err := doc.String()
	if err != nil {
		return err
	}
	args = append(args, string(document))

	_, err = c.cmd.execute(ctx, args)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(ctx context.Context, docs []*client.Document) error {
	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.desc.Name)

	docMapList := make([]map[string]any, len(docs))
	for i, doc := range docs {
		// We must call this here, else the doc key on the given object will not match
		// that of the document saved in the database
		err := doc.RemapAliasFieldsAndDockey(c.Schema().Fields)
		if err != nil {
			return err
		}
		docMap, err := doc.ToMap()
		if err != nil {
			return err
		}
		docMapList[i] = docMap
	}
	documents, err := json.Marshal(docMapList)
	if err != nil {
		return err
	}
	args = append(args, string(documents))

	_, err = c.cmd.execute(ctx, args)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(ctx context.Context, doc *client.Document) error {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, "--key", doc.Key().String())

	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	args = append(args, string(document))

	_, err = c.cmd.execute(ctx, args)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(ctx context.Context, doc *client.Document) error {
	_, err := c.Get(ctx, doc.Key(), true)
	if err == nil {
		return c.Update(ctx, doc)
	}
	if errors.Is(err, client.ErrDocumentNotFound) {
		return c.Create(ctx, doc)
	}
	return err
}

func (c *Collection) Delete(ctx context.Context, docKey client.DocKey) (bool, error) {
	res, err := c.DeleteWithKey(ctx, docKey)
	if err != nil {
		return false, err
	}
	return res.Count == 1, nil
}

func (c *Collection) Exists(ctx context.Context, docKey client.DocKey) (bool, error) {
	_, err := c.Get(ctx, docKey, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWith(ctx context.Context, target any, updater string) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, t, updater)
	case client.DocKey:
		return c.UpdateWithKey(ctx, t, updater)
	case []client.DocKey:
		return c.UpdateWithKeys(ctx, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

func (c *Collection) updateWith(
	ctx context.Context,
	args []string,
) (*client.UpdateResult, error) {
	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var res client.UpdateResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, "--updater", updater)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	return c.updateWith(ctx, args)
}

func (c *Collection) UpdateWithKey(
	ctx context.Context,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, "--key", key.String())
	args = append(args, "--updater", updater)

	return c.updateWith(ctx, args)
}

func (c *Collection) UpdateWithKeys(
	ctx context.Context,
	docKeys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, "--updater", updater)

	keys := make([]string, len(docKeys))
	for i, v := range docKeys {
		keys[i] = v.String()
	}
	args = append(args, "--key", strings.Join(keys, ","))

	return c.updateWith(ctx, args)
}

func (c *Collection) DeleteWith(ctx context.Context, target any) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, t)
	case client.DocKey:
		return c.DeleteWithKey(ctx, t)
	case []client.DocKey:
		return c.DeleteWithKeys(ctx, t)
	default:
		return nil, client.ErrInvalidDeleteTarget
	}
}

func (c *Collection) deleteWith(
	ctx context.Context,
	args []string,
) (*client.DeleteResult, error) {
	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var res client.DeleteResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) DeleteWithFilter(ctx context.Context, filter any) (*client.DeleteResult, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.desc.Name)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	return c.deleteWith(ctx, args)
}

func (c *Collection) DeleteWithKey(ctx context.Context, docKey client.DocKey) (*client.DeleteResult, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, "--key", docKey.String())

	return c.deleteWith(ctx, args)
}

func (c *Collection) DeleteWithKeys(ctx context.Context, docKeys []client.DocKey) (*client.DeleteResult, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.desc.Name)

	keys := make([]string, len(docKeys))
	for i, v := range docKeys {
		keys[i] = v.String()
	}
	args = append(args, "--key", strings.Join(keys, ","))

	return c.deleteWith(ctx, args)
}

func (c *Collection) Get(ctx context.Context, key client.DocKey, showDeleted bool) (*client.Document, error) {
	args := []string{"client", "collection", "get"}
	args = append(args, "--name", c.desc.Name)
	args = append(args, key.String())

	if showDeleted {
		args = append(args, "--show-deleted")
	}

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var docMap map[string]any
	if err := json.Unmarshal(data, &docMap); err != nil {
		return nil, err
	}
	return client.NewDocFromMap(docMap)
}

func (c *Collection) WithTxn(tx datastore.Txn) client.Collection {
	return &Collection{
		cmd:  c.cmd.withTxn(tx),
		desc: c.desc,
	}
}

func (c *Collection) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	args := []string{"client", "collection", "keys"}
	args = append(args, "--name", c.desc.Name)

	stdOut, _, err := c.cmd.executeStream(ctx, args)
	if err != nil {
		return nil, err
	}
	docKeyCh := make(chan client.DocKeysResult)

	go func() {
		dec := json.NewDecoder(stdOut)
		defer close(docKeyCh)

		for {
			var res http.DocKeyResult
			if err := dec.Decode(&res); err != nil {
				return
			}
			key, err := client.NewDocKeyFromString(res.Key)
			if err != nil {
				return
			}
			docKey := client.DocKeysResult{
				Key: key,
			}
			if res.Error != "" {
				docKey.Err = fmt.Errorf(res.Error)
			}
			docKeyCh <- docKey
		}
	}()

	return docKeyCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexDescription,
) (index client.IndexDescription, err error) {
	args := []string{"client", "index", "create"}
	args = append(args, "--collection", c.desc.Name)
	args = append(args, "--name", indexDesc.Name)

	fields := make([]string, len(indexDesc.Fields))
	for i := range indexDesc.Fields {
		fields[i] = indexDesc.Fields[i].Name
	}
	args = append(args, "--fields", strings.Join(fields, ","))

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	args := []string{"client", "index", "drop"}
	args = append(args, "--collection", c.desc.Name)
	args = append(args, "--name", indexName)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.desc.Name)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes []client.IndexDescription
	if err := json.Unmarshal(data, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}
