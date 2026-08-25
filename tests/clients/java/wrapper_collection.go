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

//go:build javaclient

package java

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	def client.CollectionVersion
	w   *Wrapper
	txn immutable.Option[datastore.Txn]
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

func setCtxTxnFromCollection(ctx context.Context, c *Collection) context.Context {
	if c.txn.HasValue() {
		return datastore.CtxSetTxn(ctx, c.txn.Value())
	}
	return ctx
}

func (c *Collection) NewIndex(
	ctx context.Context,
	indexDesc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	orderedFields := make([]string, len(indexDesc.Fields))
	for i, f := range indexDesc.Fields {
		order := "ASC"
		if f.Descending {
			order = "DESC"
		}
		orderedFields[i] = f.Name + ":" + order
	}

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "NewIndexNative",
		newArgs().argStr(indexDesc.Name).argStr(strings.Join(orderedFields, ",")).argBool(indexDesc.Unique).
			collOpts(c.def.Name, "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return client.IndexDescription{}, err
	}
	if err := res.asError(); err != nil {
		return client.IndexDescription{}, err
	}
	return unmarshalResult[client.IndexDescription](res.Value)
}

func (c *Collection) DeleteIndex(
	ctx context.Context, indexName string, opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "DeleteIndexNative",
		newArgs().argStr(indexName).collOpts(c.def.Name, "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (c *Collection) ListIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.ListIndexesResult, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "ListIndexesNative",
		newArgs().collOpts(c.def.Name, "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return []client.ListIndexesResult{}, err
	}
	if err := res.asError(); err != nil {
		return []client.ListIndexesResult{}, err
	}
	return unmarshalResult[[]client.ListIndexesResult](res.Value)
}

func (c *Collection) NewEncryptedIndex(
	ctx context.Context, req client.EncryptedIndexDescription, opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "NewEncryptedIndexNative",
		newArgs().argStr(c.def.Name).argStr(req.FieldName).argLong(idH))
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	if err := res.asError(); err != nil {
		return client.EncryptedIndexDescription{}, err
	}
	return unmarshalResult[client.EncryptedIndexDescription](res.Value)
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context, fieldName string, opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "DeleteEncryptedIndexNative",
		newArgs().argStr(c.def.Name).argStr(fieldName).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "ListEncryptedIndexesNative", newArgs().argStr(c.def.Name).argLong(idH))
	if err != nil {
		return []client.EncryptedIndexDescription{}, err
	}
	if err := res.asError(); err != nil {
		return []client.EncryptedIndexDescription{}, err
	}
	return unmarshalResult[[]client.EncryptedIndexDescription](res.Value)
}

func (c *Collection) Truncate(ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions]) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "TruncateCollectionNative",
		newArgs().collOpts(c.def.Name, "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (c *Collection) AddDocument(
	ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	addOpts := utils.NewOptions(opts...)
	encryptedFields := ""
	if len(addOpts.EncryptedFields) > 0 {
		encryptedFields = strings.Join(addOpts.EncryptedFields, ",")
	}

	docJSONBytes, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	idH := identityHandle(addOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "AddDocumentNative",
		newArgs().argStr(string(docJSONBytes)).argBool(addOpts.EncryptDoc).argStr(encryptedFields).
			collOpts(c.def.Name, "", "", false, addOpts.EnableSigning).argLong(idH))
	if err != nil {
		return err
	}
	if err := res.asError(); err != nil {
		return err
	}
	if err := setDocumentIDsFromJSON([]*client.Document{doc}, []byte(res.Value)); err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) AddManyDocuments(
	ctx context.Context, docs []*client.Document, opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	addOpts := utils.NewOptions(opts...)
	encryptedFields := ""
	if len(addOpts.EncryptedFields) > 0 {
		encryptedFields = strings.Join(addOpts.EncryptedFields, ",")
	}

	jsonDocs := make([]json.RawMessage, 0, len(docs))
	for _, doc := range docs {
		b, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf(errFmtDocumentToJSONFailed, err)
		}
		jsonDocs = append(jsonDocs, b)
	}
	docJSONBytes, err := json.Marshal(jsonDocs)
	if err != nil {
		return err
	}

	idH := identityHandle(addOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "AddDocumentNative",
		newArgs().argStr(string(docJSONBytes)).argBool(addOpts.EncryptDoc).argStr(encryptedFields).
			collOpts(c.def.Name, "", "", false, addOpts.EnableSigning).argLong(idH))
	if err != nil {
		return err
	}
	if err := res.asError(); err != nil {
		return err
	}
	if err := setDocumentIDsFromJSON(docs, []byte(res.Value)); err != nil {
		return err
	}
	for _, doc := range docs {
		doc.Clean()
	}
	return nil
}

func setDocumentIDsFromJSON(docs []*client.Document, data []byte) error {
	var docIDs []string
	if err := json.Unmarshal(data, &docIDs); err != nil {
		return err
	}
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

func (c *Collection) UpdateDocument(
	ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	ctx = setCtxTxnFromCollection(ctx, c)

	document, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}

	updateOpts := utils.NewOptions(opts...)
	idH := identityHandle(updateOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "UpdateDocumentNative",
		newArgs().argStr(doc.ID().String()).argStr("").argStr(string(document)).
			collOpts("", "", c.CollectionID(), false, updateOpts.EnableSigning).argLong(idH))
	if err != nil {
		return err
	}
	if err := res.asError(); err != nil {
		return err
	}
	doc.Clean()
	return nil
}

// SaveDocument emulates the direct Go client's atomic save (GetDocument then Add/UpdateDocument
// in a single transaction) by explicitly opening one of its own when the caller hasn't already
// attached one (via a transaction-bound Collection, c.txn), the same pattern Wrapper.AddCollection
// uses. Without this, each of the three calls below would open and commit its own implicit
// transaction, leaving a window in which another writer can add or delete the document between
// the initial GetDocument and the resulting Add/UpdateDocument call.
func (c *Collection) SaveDocument(
	ctx context.Context, doc *client.Document, opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	if !doc.ID().IsValid() {
		return c.AddDocument(ctx, doc, opts...)
	}

	var txn datastore.Txn
	hadTxn := c.txn.HasValue()
	if hadTxn {
		txn = c.txn.Value()
	} else {
		clientTxn, err := c.w.NewTxn(false)
		if err != nil {
			return err
		}
		var ok bool
		txn, ok = clientTxn.(datastore.Txn)
		if !ok {
			return errors.New(errCastClientTxnFailed)
		}
		defer txn.Discard()
		ctx = datastore.CtxSetTxn(ctx, txn)
	}

	saveOpt := utils.NewOptions(opts...)
	getOpts := options.GetDocument().SetShowDeleted(true)
	if saveOpt.Identity.HasValue() {
		getOpts.SetIdentity(saveOpt.Identity.Value())
	}
	_, err := c.GetDocument(ctx, doc.ID(), getOpts)
	switch {
	case err == nil:
		updateOpts := options.UpdateDocument()
		if saveOpt.Identity.HasValue() {
			updateOpts.SetIdentity(saveOpt.Identity.Value())
		}
		err = c.UpdateDocument(ctx, doc, updateOpts)
	case errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized):
		err = c.AddDocument(ctx, doc, opts...)
	}
	if err != nil {
		return err
	}

	if !hadTxn {
		return txn.Commit()
	}
	return nil
}

func (c *Collection) DeleteDocument(
	ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	deleteOpts := utils.NewOptions(opts...)
	idH := identityHandle(deleteOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "DeleteDocumentNative",
		newArgs().argStr(docID.String()).argStr("").collOpts(c.def.Name, "", "", false, deleteOpts.EnableSigning).argLong(idH))
	if err != nil {
		return false, err
	}
	if err := res.asError(); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) ExistsDocument(
	ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "GetDocumentNative",
		newArgs().argStr(docID.String()).argBool(false).
			collOpts("", c.def.VersionID, c.def.CollectionID, false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return false, err
	}
	if err := res.asError(); err != nil {
		if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateDocumentsWithFilter(
	ctx context.Context, filter any, updater string, opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}

	updateOpts := utils.NewOptions(opts...)
	idH := identityHandle(updateOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "UpdateDocumentNative",
		newArgs().argStr("").argStr(string(filterJSON)).argStr(updater).
			collOpts(c.def.Name, "", "", false, updateOpts.EnableSigning).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	var updateRes client.UpdateResult
	if err := json.Unmarshal([]byte(res.Value), &updateRes); err != nil {
		return nil, err
	}
	return &updateRes, nil
}

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context, filter any, opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}

	deleteOpts := utils.NewOptions(opts...)
	idH := identityHandle(deleteOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "DeleteDocumentNative",
		newArgs().argStr("").argStr(string(filterJSON)).collOpts(c.def.Name, "", "", false, deleteOpts.EnableSigning).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	var deleteRes client.DeleteResult
	if err := json.Unmarshal([]byte(res.Value), &deleteRes); err != nil {
		return nil, err
	}
	return &deleteRes, nil
}

func (c *Collection) GetDocument(
	ctx context.Context, docID client.DocID, opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	ctx = setCtxTxnFromCollection(ctx, c)

	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(c.w, ctx, "GetDocumentNative",
		newArgs().argStr(docID.String()).argBool(opt.ShowDeleted).collOpts(c.def.Name, "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}

	doc, err := client.NewDocWithID(ctx, docID, c.Version())
	if err != nil {
		return nil, err
	}
	if err := doc.SetWithJSON(ctx, []byte(res.Value)); err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}
