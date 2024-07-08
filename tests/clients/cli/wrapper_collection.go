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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/encryption"
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
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := makeDocCreateArgs(ctx, c)

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
	docs []*client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := makeDocCreateArgs(ctx, c)

	docStrings := make([]string, len(docs))
	for i, doc := range docs {
		docStr, err := doc.String()
		if err != nil {
			return err
		}
		docStrings[i] = docStr
	}
	args = append(args, "["+strings.Join(docStrings, ",")+"]")

	_, err := c.cmd.execute(ctx, args)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func makeDocCreateArgs(
	ctx context.Context,
	c *Collection,
) []string {
	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.Description().Name.Value())

	encConf := encryption.GetContextConfig(ctx)
	if encConf.HasValue() {
		if encConf.Value().IsEncrypted {
			args = append(args, "--encrypt")
		}
		if len(encConf.Value().EncryptedFields) > 0 {
			args = append(args, "--encrypt-fields", strings.Join(encConf.Value().EncryptedFields, ","))
		}
	}

	return args
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	if !c.Description().Name.HasValue() {
		return client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())
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
	doc *client.Document,
) error {
	_, err := c.Get(ctx, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, doc)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return c.Create(ctx, doc)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Description().Name.Value())
	args = append(args, "--docID", docID.String())

	_, err := c.cmd.execute(ctx, args)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	_, err := c.Get(ctx, docID, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Description().Name.Value())
	args = append(args, "--updater", updater)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

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

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Description().Name.Value())

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

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

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	if !c.Description().Name.HasValue() {
		return nil, client.ErrOperationNotPermittedOnNamelessCols
	}

	args := []string{"client", "collection", "get"}
	args = append(args, "--name", c.Description().Name.Value())
	args = append(args, docID.String())

	if showDeleted {
		args = append(args, "--show-deleted")
	}

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	doc := client.NewDocWithID(docID, c.Definition())
	err = doc.SetWithJSON(data)
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,

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
