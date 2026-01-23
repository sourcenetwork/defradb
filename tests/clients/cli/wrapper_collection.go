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
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
)

// appendIdentityArg extracts identity from an immutable.Option and appends --identity flag to args
// if identity is present and is a FullIdentity with a private key.
func appendIdentityArg(args []string, ident immutable.Option[identity.Identity]) []string {
	if !ident.HasValue() {
		return args
	}
	if fullIdent, ok := ident.Value().(identity.FullIdentity); ok {
		rawIdent := fullIdent.IntoRawIdentity()
		if rawIdent.PrivateKey != "" {
			args = append(args, "--identity", rawIdent.PrivateKey)
		}
	}
	return args
}

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	cmd *cliWrapper
	def client.CollectionVersion
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
}

func (c *Collection) CollectionID() string {
	return c.Version().CollectionID
}

func (c *Collection) Create(
	ctx context.Context,
	doc *client.Document,
	opts ...*options.CollectionCreateOptions,
) error {
	args := makeDocCreateArgs(c, opts)

	document, err := doc.String()
	if err != nil {
		return err
	}
	args = append(args, document)

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
	opts ...*options.CollectionCreateOptions,
) error {
	args := makeDocCreateArgs(c, opts)

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
	c *Collection,
	opts []*options.CollectionCreateOptions,
) []string {
	args := []string{"client", "collection", "create"}
	args = append(args, "--name", c.Version().Name)

	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]
		args = appendIdentityArg(args, opt.GetIdentity())
		if opt.EncryptDoc {
			args = append(args, "--encrypt")
		}
		if len(opt.EncryptedFields) > 0 {
			args = append(args, "--encrypt-fields", strings.Join(opt.EncryptedFields, ","))
		}
	}

	return args
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
	opts ...*options.CollectionUpdateOptions,
) error {
	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--docID", doc.ID().String())
	args = append(args, "--updater", string(document))

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

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
	opts ...*options.CollectionSaveOptions,
) error {
	var getOpts []*options.CollectionGetOptions
	if len(opts) > 0 && opts[0] != nil && opts[0].GetIdentity().HasValue() {
		getOpts = []*options.CollectionGetOptions{
			options.CollectionGet().SetIdentity(opts[0].GetIdentity().Value()),
		}
	}

	_, err := c.Get(ctx, doc.ID(), true, getOpts...)
	if err == nil {
		var updateOpts []*options.CollectionUpdateOptions
		if len(opts) > 0 && opts[0] != nil && opts[0].GetIdentity().HasValue() {
			updateOpts = []*options.CollectionUpdateOptions{
				options.CollectionUpdate().SetIdentity(opts[0].GetIdentity().Value()),
			}
		}
		return c.Update(ctx, doc, updateOpts...)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		var createOpts []*options.CollectionCreateOptions
		if len(opts) > 0 && opts[0] != nil {
			createOpt := options.CollectionCreate().
				SetEncryptDoc(opts[0].EncryptDoc).
				SetEncryptedFields(opts[0].EncryptedFields)
			if opts[0].GetIdentity().HasValue() {
				createOpt.SetIdentity(opts[0].GetIdentity().Value())
			}
			createOpts = []*options.CollectionCreateOptions{createOpt}
		}
		return c.Create(ctx, doc, createOpts...)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
	opts ...*options.CollectionDeleteOptions,
) (bool, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--docID", docID.String())

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

	_, err := c.cmd.execute(ctx, args)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
	opts ...*options.CollectionExistsOptions,
) (bool, error) {
	var getOpts []*options.CollectionGetOptions
	if len(opts) > 0 && opts[0] != nil && opts[0].GetIdentity().HasValue() {
		getOpts = []*options.CollectionGetOptions{
			options.CollectionGet().SetIdentity(opts[0].GetIdentity().Value()),
		}
	}

	_, err := c.Get(ctx, docID, false, getOpts...)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...*options.CollectionUpdateWithFilterOptions,
) (*client.UpdateResult, error) {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--updater", updater)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

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
	opts ...*options.CollectionDeleteWithFilterOptions,
) (*client.DeleteResult, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Version().Name)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

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
	opts ...*options.CollectionGetOptions,
) (*client.Document, error) {
	args := []string{"client", "collection", "get"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, docID.String())

	if showDeleted {
		args = append(args, "--show-deleted")
	}

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	doc, err := client.NewDocWithID(ctx, docID, c.Version())
	if err != nil {
		return nil, err
	}
	err = doc.SetWithJSON(ctx, data)
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
	opts ...*options.CollectionGetAllDocIDsOptions,
) (<-chan client.DocIDResult, error) {
	args := []string{"client", "collection", "docIDs"}
	args = append(args, "--name", c.Version().Name)

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

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
				docIDResult.Err = errors.New(res.Error)
			}
			docIDCh <- docIDResult
		}
	}()

	return docIDCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexCreateRequest,
	opts ...*options.CollectionCreateIndexOptions,
) (index client.IndexDescription, err error) {
	args := []string{"client", "index", "create"}
	args = append(args, "--collection", c.Version().Name)
	if indexDesc.Name != "" {
		args = append(args, "--name", indexDesc.Name)
	}
	if indexDesc.Unique {
		args = append(args, "--unique")
	}

	fields := make([]string, len(indexDesc.Fields))
	orders := make([]bool, len(indexDesc.Fields))

	for i := range indexDesc.Fields {
		fields[i] = indexDesc.Fields[i].Name
		orders[i] = indexDesc.Fields[i].Descending
	}

	orderedFields := make([]string, len(fields))

	for i := range fields {
		if orders[i] {
			orderedFields[i] = fields[i] + ":DESC"
		} else {
			orderedFields[i] = fields[i] + ":ASC"
		}
	}

	args = append(args, "--fields", strings.Join(orderedFields, ","))

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

func (c *Collection) DropIndex(
	ctx context.Context,
	indexName string,
	opts ...*options.CollectionDropIndexOptions,
) error {
	args := []string{"client", "index", "drop"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--name", indexName)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) GetIndexes(
	ctx context.Context,
	opts ...*options.CollectionGetIndexesOptions,
) ([]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.Version().Name)

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

// CreateEncryptedIndex implements client.Collection.
func (c *Collection) CreateEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
) (index client.EncryptedIndexDescription, err error) {
	args := []string{"client", "encrypted-index", "create"}
	args = append(args, "--collection", c.Version().Name)

	args = append(args, "--field", indexDesc.FieldName)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

// ListEncryptedIndexes implements client.Collection.
func (c *Collection) ListEncryptedIndexes(ctx context.Context) ([]client.EncryptedIndexDescription, error) {
	args := []string{"client", "encrypted-index", "list"}
	args = append(args, "--collection", c.Version().Name)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes []client.EncryptedIndexDescription
	if err := json.Unmarshal(data, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

// DeleteEncryptedIndex implements client.Collection.
func (c *Collection) DeleteEncryptedIndex(ctx context.Context, fieldName string) error {
	args := []string{"client", "encrypted-index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", fieldName)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) Truncate(ctx context.Context, opts ...*options.CollectionTruncateOptions) error {
	args := []string{"client", "collection", "truncate"}
	args = append(args, "--name", c.Version().Name)

	if len(opts) > 0 && opts[0] != nil {
		args = appendIdentityArg(args, opts[0].GetIdentity())
	}

	_, err := c.cmd.execute(ctx, args)
	return err
}
