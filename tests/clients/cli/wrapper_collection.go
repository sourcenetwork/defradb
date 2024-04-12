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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	cmd *cliWrapper
	def client.CollectionDefinition
}

func (c *Collection) Description() client.CollectionDescription {
	return c.def.Description
}

func (c *Collection) Name() immutable.Option[string] {
	return c.Description().Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.def.Schema
}

func (c *Collection) ID() uint32 {
	return c.Description().ID
}

func (c *Collection) SchemaRoot() string {
	return c.Schema().Root
}

func (c *Collection) Definition() client.CollectionDefinition {
	return c.def
}

func (c *Collection) Create(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
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

func (c *Collection) CreateMany(
	ctx context.Context,
	identity immutable.Option[string],
	docs []*client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	docMapList := make([]map[string]any, len(docs))
	for i, doc := range docs {
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

func (c *Collection) Update(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, "--docID", doc.ID().String())

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

func (c *Collection) Save(
	ctx context.Context,
	identity immutable.Option[string],
	doc *client.Document,
) error {
	_, err := c.Get(ctx, identity, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, identity, doc)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return c.Create(ctx, identity, doc)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (bool, error) {
	res, err := c.DeleteWithDocID(ctx, identity, docID)
	if err != nil {
		return false, err
	}
	return res.Count == 1, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (bool, error) {
	_, err := c.Get(ctx, identity, docID, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWith(
	ctx context.Context,
	identity immutable.Option[string],
	target any,
	updater string,
) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, identity, t, updater)
	case client.DocID:
		return c.UpdateWithDocID(ctx, identity, t, updater)
	case []client.DocID:
		return c.UpdateWithDocIDs(ctx, identity, t, updater)
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
	identity immutable.Option[string],
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, "--updater", updater)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	return c.updateWith(ctx, args)
}

func (c *Collection) UpdateWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, "--docID", docID.String())
	args = append(args, "--updater", updater)

	return c.updateWith(ctx, args)
}

func (c *Collection) UpdateWithDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
	docIDs []client.DocID,
	updater string,
) (*client.UpdateResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, "--updater", updater)

	strDocIDs := make([]string, len(docIDs))
	for i, v := range docIDs {
		strDocIDs[i] = v.String()
	}
	args = append(args, "--docID", strings.Join(strDocIDs, ","))

	return c.updateWith(ctx, args)
}

func (c *Collection) DeleteWith(
	ctx context.Context,
	identity immutable.Option[string],
	target any,
) (*client.DeleteResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.DeleteWithFilter(ctx, identity, t)
	case client.DocID:
		return c.DeleteWithDocID(ctx, identity, t)
	case []client.DocID:
		return c.DeleteWithDocIDs(ctx, identity, t)
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

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	identity immutable.Option[string],
	filter any,
) (*client.DeleteResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	return c.deleteWith(ctx, args)
}

func (c *Collection) DeleteWithDocID(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
) (*client.DeleteResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, "--docID", docID.String())

	return c.deleteWith(ctx, args)
}

func (c *Collection) DeleteWithDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
	docIDs []client.DocID,
) (*client.DeleteResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	strDocIDs := make([]string, len(docIDs))
	for i, v := range docIDs {
		strDocIDs[i] = v.String()
	}
	args = append(args, "--docID", strings.Join(strDocIDs, ","))

	return c.deleteWith(ctx, args)
}

func (c *Collection) Get(
	ctx context.Context,
	identity immutable.Option[string],
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "get"}
	args = append(args, "--name", c.Description().Name.Value())

	if identity.HasValue() {
		args = append(args, "--identity", identity.Value())
	}

	args = append(args, docID.String())

	if showDeleted {
		args = append(args, "--show-deleted")
	}

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithID(docID, c.Schema())
	err = doc.SetWithJSON(data)
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
	identity immutable.Option[string],
) (<-chan client.DocIDResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "docIDs"}
	args = append(args, "--name", c.Description().Name.Value())

	stdOut, _, err := c.cmd.executeStream(ctx, args)
	if err != nil {
		return nil, err
	}
	docIDCh := make(chan client.DocIDResult)

	go func() {
		dec := json.NewDecoder(stdOut)
		defer close(docIDCh)

		for {
			var res http.DocIDResult
			if err := dec.Decode(&res); err != nil {
				return
			}
			docID, err := client.NewDocIDFromString(res.DocID)
			if err != nil {
				return
			}
			docIDResult := client.DocIDResult{
				ID: docID,
			}
			if res.Error != "" {
				docIDResult.Err = fmt.Errorf(res.Error)
			}
			docIDCh <- docIDResult
		}
	}()

	return docIDCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexDescription,
) (index client.IndexDescription, err error) {
	if !c.Description().Name.HasValue() {
		return client.IndexDescription{}, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "index", "create"}
	args = append(args, "--collection", c.Description().Name.Value())
	if indexDesc.Name != "" {
		args = append(args, "--name", indexDesc.Name)
	}
	if indexDesc.Unique {
		args = append(args, "--unique")
	}

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
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "index", "drop"}
	args = append(args, "--collection", c.Description().Name.Value())
	args = append(args, "--name", indexName)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.Description().Name.Value())

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

func (c *Collection) CreateDocIndex(context.Context, *client.Document) error {
	return ErrMethodIsNotImplemented
}

func (c *Collection) UpdateDocIndex(ctx context.Context, oldDoc, newDoc *client.Document) error {
	return ErrMethodIsNotImplemented
}

func (c *Collection) DeleteDocIndex(context.Context, *client.Document) error {
	return ErrMethodIsNotImplemented
}
