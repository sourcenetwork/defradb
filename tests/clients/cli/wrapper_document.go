// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cli

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (c *Collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
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

func (c *Collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
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
	opts ...options.Enumerable[options.AddDocumentOptions],
) []string {
	args := []string{"client", "document", "add"}
	args = append(args, "--collection-name", c.Version().Name)

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

func (c *Collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}

	args := []string{"client", "document", "update"}
	args = append(args, "--collection-name", c.Version().Name)
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

func (c *Collection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	getOpts := options.GetDocument()
	opt := utils.NewOptions(opts...)
	if opt.Identity.HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}
	_, err := c.GetDocument(ctx, doc.ID(), getOpts.SetShowDeleted(true))
	if err == nil {
		updateOpts := options.UpdateDocument()
		opt := utils.NewOptions(opts...)
		if opt.GetIdentity().HasValue() {
			updateOpts.SetIdentity(opt.GetIdentity().Value())
		}
		return c.UpdateDocument(ctx, doc, updateOpts)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		opt := utils.NewOptions(opts...)
		addOpt := options.AddDocument().
			SetEncryptDoc(opt.EncryptDoc).
			SetEncryptedFields(opt.EncryptedFields)
		if opt.GetIdentity().HasValue() {
			addOpt.SetIdentity(opt.GetIdentity().Value())
		}
		return c.AddDocument(ctx, doc, addOpt)
	}
	return err
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	args := []string{"client", "document", "delete"}
	args = append(args, "--collection-name", c.Version().Name)
	args = append(args, "--docID", docID.String())

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	getOpts := options.GetDocument()
	opt := utils.NewOptions(opts...)
	if opt.GetIdentity().HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}

	_, err := c.GetDocument(ctx, docID, getOpts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	args := []string{"client", "document", "update"}
	args = append(args, "--collection-name", c.Version().Name)
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

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	args := []string{"client", "document", "delete"}
	args = append(args, "--collection-name", c.Version().Name)

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

func (c *Collection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)

	args := []string{"client", "document", "get"}
	args = append(args, "--collection-name", c.Version().Name)
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
