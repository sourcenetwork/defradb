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
	"github.com/sourcenetwork/defradb/internal/utils"
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

func (c *Collection) Add(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	args := makeDocAddArgs(c, opts...)

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

func (c *Collection) AddMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.CollectionAddOptions],
) error {
	args := makeDocAddArgs(c, opts...)

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

func makeDocAddArgs(
	c *Collection,
	opts ...options.Enumerable[options.CollectionAddOptions],
) []string {
	args := []string{"client", "collection", "add"}
	args = append(args, "--name", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	if opt.EncryptDoc {
		args = append(args, "--encrypt")
	}
	if len(opt.EncryptedFields) > 0 {
		args = append(args, "--encrypt-fields", strings.Join(opt.EncryptedFields, ","))
	}

	return args
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.CollectionUpdateOptions],
) error {
	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}

	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--docID", doc.ID().String())
	args = append(args, "--updater", string(document))

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.CollectionSaveOptions],
) error {
	getOpts := options.CollectionGet()
	opt := utils.NewOptions(opts...)
	if opt.Identity.HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}
	_, err := c.Get(ctx, doc.ID(), getOpts.SetShowDeleted(true))
	if err == nil {
		updateOpts := options.CollectionUpdate()
		opt := utils.NewOptions(opts...)
		if opt.GetIdentity().HasValue() {
			updateOpts.SetIdentity(opt.GetIdentity().Value())
		}
		return c.Update(ctx, doc, updateOpts)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		opt := utils.NewOptions(opts...)
		addOpt := options.CollectionAdd().
			SetEncryptDoc(opt.EncryptDoc).
			SetEncryptedFields(opt.EncryptedFields)
		if opt.GetIdentity().HasValue() {
			addOpt.SetIdentity(opt.GetIdentity().Value())
		}
		return c.Add(ctx, doc, addOpt)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionDeleteOptions],
) (bool, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--docID", docID.String())

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionExistsOptions],
) (bool, error) {
	getOpts := options.CollectionGet()
	opt := utils.NewOptions(opts...)
	if opt.GetIdentity().HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}

	_, err := c.Get(ctx, docID, getOpts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.CollectionUpdateWithFilterOptions],
) (*client.UpdateResult, error) {
	args := []string{"client", "collection", "update"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, "--updater", updater)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.CollectionDeleteWithFilterOptions],
) (*client.DeleteResult, error) {
	args := []string{"client", "collection", "delete"}
	args = append(args, "--name", c.Version().Name)

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	args = append(args, "--filter", string(filterJSON))

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.CollectionGetOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)

	args := []string{"client", "collection", "get"}
	args = append(args, "--name", c.Version().Name)
	args = append(args, docID.String())

	if opt.ShowDeleted {
		args = append(args, "--show-deleted")
	}
	args = appendIdentityArg(args, opt.GetIdentity())

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
	opts ...options.Enumerable[options.CollectionGetAllDocIDsOptions],
) (<-chan client.DocIDResult, error) {
	args := []string{"client", "collection", "docIDs"}
	args = append(args, "--name", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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

func (c *Collection) AddIndex(
	ctx context.Context,
	indexDesc client.IndexAddRequest,
	opts ...options.Enumerable[options.CollectionAddIndexOptions],
) (index client.IndexDescription, err error) {
	args := []string{"client", "index", "add"}
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

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.CollectionDeleteIndexOptions],
) error {
	args := []string{"client", "index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--name", indexName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionListIndexesOptions],
) ([]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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

// AddEncryptedIndex implements client.Collection.
func (c *Collection) AddEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.AddEncryptedIndexOptions],
) (index client.EncryptedIndexDescription, err error) {
	args := []string{"client", "encrypted-index", "add"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", indexDesc.FieldName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	args := []string{"client", "encrypted-index", "list"}
	args = append(args, "--collection", c.Version().Name)
	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	args := []string{"client", "encrypted-index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", fieldName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions],
) error {
	args := []string{"client", "collection", "truncate"}
	args = append(args, "--name", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	return err
}
