// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (c *Collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	body, err := doc.String()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(body))
	if err != nil {
		return err
	}

	setDocEncryptionFlagIfNeeded(req, opt)
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	var docIDs []string
	if err := c.http.requestJson(req, &docIDs); err != nil {
		return err
	}
	if err := setDocumentIDs([]*client.Document{doc}, docIDs); err != nil {
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
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	var docMapList []json.RawMessage
	for _, doc := range docs {
		docMap, err := doc.ToJSONPatch()
		if err != nil {
			return err
		}
		docMapList = append(docMapList, docMap)
	}

	body, err := json.Marshal(docMapList)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	setDocEncryptionFlagIfNeeded(req, opt)
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	var docIDs []string
	if err := c.http.requestJson(req, &docIDs); err != nil {
		return err
	}
	if err := setDocumentIDs(docs, docIDs); err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func setDocumentIDs(docs []*client.Document, docIDs []string) error {
	if len(docIDs) != len(docs) {
		return client.NewErrUnexpectedType[[]string]("docIDs", docIDs)
	}
	for i, docIDString := range docIDs {
		docID, err := client.NewDocIDFromString(docIDString)
		if err != nil {
			return err
		}
		client.ApplySavedDocumentID(docs[i], docID)
	}
	return nil
}

func setDocEncryptionFlagIfNeeded(req *http.Request, opt *options.AddDocumentOptions) {
	q := req.URL.Query()
	if opt.EncryptDoc {
		q.Set(docEncryptParam, "true")
	}
	if len(opt.EncryptedFields) > 0 {
		q.Set(docEncryptFieldsParam, strings.Join(opt.EncryptedFields, ","))
	}
	if len(q) > 0 {
		req.URL.RawQuery = q.Encode()
	}
}

func setDocSigningFlagIfNeeded(req *http.Request, enableSigning immutable.Option[bool]) {
	if !enableSigning.HasValue() {
		return
	}
	q := req.URL.Query()
	q.Set(docEnableSigningParam, strconv.FormatBool(enableSigning.Value()))
	req.URL.RawQuery = q.Encode()
}

func (c *Collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "document", doc.ID().String())

	body, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	_, err = c.http.request(req)
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
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	if !doc.ID().IsValid() {
		return c.AddDocument(ctx, doc, opts...)
	}

	opt := utils.NewOptions(opts...)

	getOpts := options.GetDocument()
	if opt.GetIdentity().HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}
	_, err := c.GetDocument(ctx, doc.ID(), getOpts.SetShowDeleted(true))
	if err == nil {
		updateOpts := options.UpdateDocument()
		if opt.GetIdentity().HasValue() {
			updateOpts.SetIdentity(opt.GetIdentity().Value())
		}
		if opt.EnableSigning.HasValue() {
			updateOpts.SetEnableSigning(opt.EnableSigning.Value())
		}
		return c.UpdateDocument(ctx, doc, updateOpts)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return c.AddDocument(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "document", docID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return false, err
	}
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	_, err = c.http.request(req)
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
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	_, err := c.GetDocument(ctx, docID)
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
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	request := UpdateCollectionRequest{
		Filter:  filter,
		Updater: updater,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	var result client.UpdateResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	request := DeleteCollectionRequest{
		Filter: filter,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	setDocSigningFlagIfNeeded(req, opt.EnableSigning)

	var result client.DeleteResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	if c.txn.HasValue() {
		ctx = datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}

	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	query := url.Values{}
	if opt.ShowDeleted {
		query.Add("show_deleted", "true")
	}

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "document", docID.String())
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}

	data, err := c.http.request(req)
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
