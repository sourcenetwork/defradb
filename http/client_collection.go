// Copyright 2023 Democratized Data Foundation
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
	"strings"

	sse "github.com/vito/go-sse/sse"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*Collection)(nil)

// Collection implements the client.Collection interface over HTTP.
type Collection struct {
	http *httpClient
	def  client.CollectionVersion
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
	opts ...options.Lister[options.CollectionCreateOptions],
) error {
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

	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Lister[options.CollectionCreateOptions],
) error {
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

	_, err = c.http.request(req)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func setDocEncryptionFlagIfNeeded(req *http.Request, opt *options.CollectionCreateOptions) {
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

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Lister[options.CollectionUpdateOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, doc.ID().String())

	body, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Lister[options.CollectionSaveOptions],
) error {
	opt := utils.NewOptions(opts...)

	getOpts := options.CollectionGet()
	if opt.GetIdentity().HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}
	_, err := c.Get(ctx, doc.ID(), getOpts.SetShowDeleted(true))
	if err == nil {
		updateOpts := options.CollectionUpdate()
		if opt.GetIdentity().HasValue() {
			updateOpts.SetIdentity(opt.GetIdentity().Value())
		}
		return c.Update(ctx, doc, updateOpts)
	}
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		createOpts := options.CollectionCreate().
			SetEncryptDoc(opt.EncryptDoc).
			SetEncryptedFields(opt.EncryptedFields)

		if opt.GetIdentity().HasValue() {
			createOpts.SetIdentity(opt.GetIdentity().Value())
		}

		return c.Create(ctx, doc, createOpts)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Lister[options.CollectionDeleteOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, docID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return false, err
	}

	_, err = c.http.request(req)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Lister[options.CollectionExistsOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	_, err := c.Get(ctx, docID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Lister[options.CollectionUpdateWithFilterOptions],
) (*client.UpdateResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	request := CollectionUpdateRequest{
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

	var result client.UpdateResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Lister[options.CollectionDeleteWithFilterOptions],
) (*client.DeleteResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	request := CollectionDeleteRequest{
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

	var result client.DeleteResult
	if err := c.http.requestJson(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Lister[options.CollectionGetOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	query := url.Values{}
	if opt.ShowDeleted {
		query.Add("show_deleted", "true")
	}

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, docID.String())
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

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
	opts ...options.Lister[options.CollectionGetAllDocIDsOptions],
) (<-chan client.DocIDResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}

	err = c.http.setDefaultHeaders(req)
	if err != nil {
		return nil, err
	}

	res, err := c.http.client.Do(req)
	if err != nil {
		return nil, err
	}
	docIDCh := make(chan client.DocIDResult)

	go func() {
		eventReader := sse.NewReadCloser(res.Body)
		// ignore close errors because the status
		// and body of the request are already
		// checked and it cannot be handled properly
		defer eventReader.Close() //nolint:errcheck
		defer close(docIDCh)

		for {
			evt, err := eventReader.Next()
			if err != nil {
				return
			}
			var res DocIDResult
			if err := json.Unmarshal(evt.Data, &res); err != nil {
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
	opts ...options.Lister[options.CollectionCreateIndexOptions],
) (client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes")

	body, err := json.Marshal(&indexDesc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return client.IndexDescription{}, err
	}
	var index client.IndexDescription
	if err := c.http.requestJson(req, &index); err != nil {
		return client.IndexDescription{}, err
	}
	return index, nil
}

func (c *Collection) DropIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Lister[options.CollectionDropIndexOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes", indexName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Collection) GetIndexes(
	ctx context.Context,
	opts ...options.Lister[options.CollectionGetIndexesOptions],
) ([]client.IndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes []client.IndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Collection) CreateEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
) (client.EncryptedIndexDescription, error) {
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes")

	body, err := json.Marshal(&indexDesc)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	var index client.EncryptedIndexDescription
	if err := c.http.requestJson(req, &index); err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return index, nil
}

func (c *Collection) ListEncryptedIndexes(ctx context.Context) ([]client.EncryptedIndexDescription, error) {
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes []client.EncryptedIndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Collection) DeleteEncryptedIndex(ctx context.Context, fieldName string) error {
	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "encrypted-indexes", fieldName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}

func (c *Collection) Truncate(ctx context.Context, opts ...options.Lister[options.CollectionTruncateOptions]) error {
	opt := utils.NewOptions(opts...)
	ctx = identity.WithContext(ctx, opt.GetIdentity())

	methodURL := c.http.apiURL.JoinPath("collections", c.Version().Name, "truncate")

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}
