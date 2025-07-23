// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
	"github.com/sourcenetwork/defradb/client"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	def client.CollectionDefinition
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def.Version
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) Schema() client.SchemaDescription {
	return c.def.Schema
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
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
	opts ...client.DocCreateOption,
) error {

	isEncrypted := isEncryptedFromDocCreateOption(opts)
	encryptedFields := encryptedFieldsFromDocCreateOptions(opts)

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.def.GetName()
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	docJSONbytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	cJSON := string(docJSONbytes)

	result := cbindings.CollectionCreate(cJSON, isEncrypted, encryptedFields, copts)

	if result.Status != 0 {
		return errors.New(result.Error)
	}

	doc.Clean()
	return nil
}

func (c *Collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
	opts ...client.DocCreateOption,
) error {
	isEncrypted := isEncryptedFromDocCreateOption(opts)
	encryptedFields := encryptedFieldsFromDocCreateOptions(opts)

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.def.GetName()
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	var jsonDocs []json.RawMessage
	for _, doc := range docs {
		b, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to convert document to JSON: %w", err)
		}
		jsonDocs = append(jsonDocs, b)
	}
	docJSONbytes, err := json.Marshal(jsonDocs)
	if err != nil {
		return err
	}
	cJSON := string(docJSONbytes)

	result := cbindings.CollectionCreate(cJSON, isEncrypted, encryptedFields, copts)

	if result.Status != 0 {
		return errors.New(result.Error)
	}

	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func (c *Collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	docID := doc.ID().String()
	filter := ""
	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	updater := string(document)

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = c.Schema().VersionID
	copts.Name = ""
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionUpdate(docID, filter, updater, copts)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	doc.Clean()
	return nil
}

func (c *Collection) Save(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	_, err := c.Get(ctx, doc.ID(), true)
	if err == nil {
		return c.Update(ctx, doc)
	}
	if strings.Contains(err.Error(), client.ErrDocumentNotFoundOrNotAuthorized.Error()) {
		return c.Create(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) Delete(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	docIDStr := docID.String()
	filter := ""

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.def.GetName()
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionDelete(docIDStr, filter, copts)

	if result.Status != 0 {
		return false, errors.New(result.Error)
	}
	return true, nil
}

func (c *Collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	docIDStr := docID.String()
	var cShowDeleted bool = false

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = ""
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionGet(docIDStr, cShowDeleted, copts)

	if result.Status != 0 {
		return false, errors.New(result.Error)
	}
	return true, nil
}

func (c *Collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	docID := ""
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filterStr := string(filterJSON)

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.def.GetName()
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionUpdate(docID, filterStr, updater, copts)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var res client.UpdateResult
	retString := []byte(result.Value)
	if err := json.Unmarshal(retString, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) DeleteWithFilter(
	ctx context.Context,
	filter any,
) (*client.DeleteResult, error) {
	docID := ""
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	filterStr := string(filterJSON)

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.def.GetName()
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionDelete(docID, filterStr, copts)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var res client.DeleteResult
	retString := []byte(result.Value)
	if err := json.Unmarshal(retString, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	docIDStr := docID.String()

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = c.Version().Name
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionGet(docIDStr, showDeleted, copts)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	jsonStr := result.Value
	doc, err := client.NewDocWithID(docID, c.Definition())
	if err != nil {
		return nil, err
	}
	err = doc.SetWithJSON([]byte(jsonStr))
	if err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}

func (c *Collection) GetAllDocIDs(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {

	var copts cbindings.GoCOptions
	copts.TxID = txnIDFromContext(ctx)
	copts.Version = ""
	copts.CollectionID = ""
	copts.Name = ""
	copts.Identity = identityFromContext(ctx)
	copts.GetInactive = 0

	result := cbindings.CollectionListDocIDs(copts)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	docIDCh := make(chan client.DocIDResult)

	go func() {
		defer close(docIDCh)

		// Extract and decode JSON from C-Result
		var rawResults []struct {
			DocID string `json:"docID"`
			Error string `json:"error"`
		}

		if err := json.Unmarshal([]byte(result.Value), &rawResults); err != nil {
			docIDCh <- client.DocIDResult{Err: fmt.Errorf("failed to parse docIDs: %w", err)}
			return
		}

		// Send each result to the channel
		for _, r := range rawResults {
			docID, err := client.NewDocIDFromString(r.DocID)
			res := client.DocIDResult{
				ID: docID,
			}
			if err != nil {
				res.Err = err
			}
			if r.Error != "" {
				res.Err = errors.New(r.Error)
			}
			docIDCh <- res
		}
	}()

	return docIDCh, nil
}

func (c *Collection) CreateIndex(
	ctx context.Context,
	indexDesc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	txnID := txnIDFromContext(ctx)
	name := c.def.GetName()

	// Build the Fields string
	orderedFields := make([]string, len(indexDesc.Fields))
	for i, f := range indexDesc.Fields {
		order := "ASC"
		if f.Descending {
			order = "DESC"
		}
		orderedFields[i] = f.Name + ":" + order
	}
	fields := strings.Join(orderedFields, ",")

	result := cbindings.IndexCreate(name, indexDesc.Name, fields, indexDesc.Unique, txnID)

	if result.Status != 0 {
		return client.IndexDescription{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to client.IndexDescription
	retRes, err := unmarshalResult[client.IndexDescription](result.Value)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return retRes, nil
}

func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	txnID := txnIDFromContext(ctx)
	name := c.def.GetName()

	result := cbindings.IndexDrop(name, indexName, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (c *Collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	txnID := txnIDFromContext(ctx)
	name := c.def.GetName()

	result := cbindings.IndexList(name, txnID)

	if result.Status != 0 {
		return []client.IndexDescription{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to []client.IndexDescription
	retRes, err := unmarshalResult[[]client.IndexDescription](result.Value)
	if err != nil {
		return []client.IndexDescription{}, err
	}
	return retRes, nil
}
