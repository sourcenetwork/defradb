// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/planner"

	"github.com/fxamacker/cbor/v2"
	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
)

type UpdateOpt struct{}
type CreateOpt struct{}

var (
	ErrInvalidTarget         = errors.New("The doc targeter is an unknown type")
	ErrTargetEmpty           = errors.New("The doc targeter cannot be empty")
	ErrInvalidUpdater        = errors.New("The doc updater is an unknown type")
	ErrUpdateEmpty           = errors.New("The doc update cannot be empty")
	ErrInvalidMergeValueType = errors.New("The type of value in the merge patch doesn't match the schema")
)

func (c *Collection) Create2(doc *document.SimpleDocument, opts ...CreateOpt) error {

	return nil
}

func (c *Collection) save2(txn *Txn, doc *document.SimpleDocument) error {

	return nil
}

// Update2 updates the given doc. It will scan through the field/value pairs
// and find those marked for update, and apply the appropriate update.
// Update only works on root level field/value pairs. So not foreign or related
// types can be updated. If you wish to update sub types, use UpdateWith, and supply
// an update payload in the form of a Patch or a Merge Patch.
func (c *Collection) Update2(doc *document.SimpleDocument, opts ...client.UpdateOpt) error {
	return nil
}

// UpdateWith updates a target document using the given updater type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Update.
// Eg: UpdateWithFilter or UpdateWithKey
func (c *Collection) UpdateWith(target interface{}, updater interface{}, opts ...client.UpdateOpt) error {
	switch t := target.(type) {
	case string, map[string]interface{}, *parser.Filter:
		_, err := c.UpdateWithFilter(t, updater, opts...)
		return err
	case key.DocKey:
		_, err := c.UpdateWithKey(t, updater, opts...)
		return err
	case []key.DocKey:
		_, err := c.UpdateWithKeys(t, updater, opts...)
		return err
	case *document.SimpleDocument:
		return c.UpdateWithDoc(t, updater, opts...)
	case []*document.SimpleDocument:
		return c.UpdateWithDocs(t, updater, opts...)
	default:
		return ErrInvalidTarget
	}
}

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) UpdateWithFilter(filter interface{}, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(txn)
	res, err := c.updateWithFilter(txn, filter, updater, opts...)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(txn)
}

// UpdateWithKey updates using a DocKey to target a single document for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) UpdateWithKey(key key.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(txn)
	res, err := c.updateWithKey(txn, key, updater, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(txn)
}

// UpdateWithKeys is the same as UpdateWithKey but accepts multiple keys as a slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) UpdateWithKeys(keys []key.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	txn, err := c.getTxn(false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(txn)
	res, err := c.updateWithKeys(txn, keys, updater, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(txn)
}

// UpdateWithDoc updates targeting the supplied document.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) UpdateWithDoc(doc *document.SimpleDocument, updater interface{}, opts ...client.UpdateOpt) error {
	return nil
}

// UpdateWithDocs updates all the supplied documents in the slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) UpdateWithDocs(docs []*document.SimpleDocument, updater interface{}, opts ...client.UpdateOpt) error {
	return nil
}

func (c *Collection) updateWithKey(txn *Txn, key key.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isPatch = false
	default:
		return nil, ErrInvalidUpdater
	}

	doc, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	v, err := doc.ToMap()
	if err != nil {
		return nil, err
	}

	if isPatch {
		// todo
	} else {
		err = c.applyMerge(txn, v, patch.(map[string]interface{}))
	}
	if err != nil {
		return nil, nil
	}

	results := &client.UpdateResult{
		Count:   1,
		DocKeys: []string{key.String()},
	}
	return results, nil
}

func (c *Collection) updateWithKeys(txn *Txn, keys []key.DocKey, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	// fmt.Println("updating keys:", keys)
	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isPatch = false
	default:
		return nil, ErrInvalidUpdater
	}

	results := &client.UpdateResult{
		DocKeys: make([]string, len(keys)),
	}
	for i, key := range keys {
		doc, err := c.Get(key)
		if err != nil {
			fmt.Println("error getting key to update:", key)
			return nil, err
		}
		v, err := doc.ToMap()
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else {
			err = c.applyMerge(txn, v, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, nil
		}

		results.DocKeys[i] = key.String()
		results.Count++
	}
	return results, nil
}

func (c *Collection) updateWithFilter(txn *Txn, filter interface{}, updater interface{}, opts ...client.UpdateOpt) (*client.UpdateResult, error) {
	patch, err := parseUpdater(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	isMerge := false
	switch patch.(type) {
	case []map[string]interface{}:
		isPatch = true
	case map[string]interface{}:
		isMerge = true
	default:
		return nil, ErrInvalidUpdater
	}

	// scan through docs with filter
	query, err := c.makeSelectionQuery(txn, filter, opts...)
	if err != nil {
		return nil, err
	}
	if err := query.Start(); err != nil {
		return nil, err
	}

	results := &client.UpdateResult{
		DocKeys: make([]string, 0),
	}

	// loop while we still have results from the filter query
	for {
		next, err := query.Next()
		if err != nil {
			return nil, err
		}
		// if theres no more records from the query, jump out of the loop
		if !next {
			break
		}

		// Get the document, and apply the patch
		doc := query.Values()
		if isPatch {
			err = c.applyPatch(txn, doc, patch.([]map[string]interface{}))
		} else if isMerge { // else is fine here
			err = c.applyMerge(txn, doc, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, err
		}

		// add succesful updated doc to results
		results.DocKeys = append(results.DocKeys, doc["_key"].(string))
		results.Count++
	}

	return results, nil
}

// func (c *Collection) updateWithFilterPatch(txn *Txn, filter map[string]interface{}, patch []map[string]interface{}, opts ...client.UpdateOpt) (*UpdateResult, error) {
// 	// scan through docs with filter
// 	query, err := c.makeQuery(filter, opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := query.Start(); err != nil {
// 		return nil, err
// 	}

// 	// loop while we still have results from the filter query
// 	for {
// 		next, err := query.Next()
// 		if err != nil {
// 			return nil, err
// 		}

// 		// if theres no more records from the query, jump out of the loop
// 		if !next {
// 			break
// 		}

// 		// Get the document, and apply the patch
// 		doc := query.Values()
// 	}

// 	// loop through patch ops
// 	// apply each
// 	// if op is a sub field, get target collection and docID, call c.applyUpdateWithPatch()
// 	return nil, nil
// }

// func (c *Collection) updateWithFilterMergePatch(txn *Txn, filter map[string]interface{}, merge map[string]interface{}, opts ...client.UpdateOpt) (*UpdateResult, error) {
// 	// loop through the fields of merge patch
// 	// apply
// 	return nil, nil
// }

func (c *Collection) applyPatch(txn *Txn, doc map[string]interface{}, patch []map[string]interface{}) error {
	for _, op := range patch {
		path, ok := op["path"].(string)
		if !ok {
			return errors.New("Missing document field to update")
		}

		targetCollection, _, err := c.getCollectionForPatchOpPath(txn, path)
		if err != nil {
			return err
		}

		key, err := c.getTargetKeyForPatchPath(txn, doc, path)
		if err != nil {
			return err
		}
		field, val, ok := getValFromDocForPatchPath(doc, path)
		if err := targetCollection.applyPatchOp(txn, key, field, val, op); err != nil {
			return err
		}
	}

	// comleted patch update
	return nil
}

func (c *Collection) applyPatchOp(txn *Txn, dockey string, field string, currentVal interface{}, patchOp map[string]interface{}) error {
	return nil
}

func (c *Collection) applyMerge(txn *Txn, doc map[string]interface{}, merge map[string]interface{}) error {
	keyStr, ok := doc["_key"].(string)
	if !ok {
		return errors.New("Document is missing key")
	}
	key := ds.NewKey(keyStr)
	links := make([]core.DAGLink, 0)
	for mfield, mval := range merge {
		if _, ok := mval.(map[string]interface{}); ok {
			return ErrInvalidMergeValueType
		}

		fd, valid := c.desc.GetField(mfield)
		if !valid {
			return errors.New("Invalid field in Patch")
		}

		cval, err := validateFieldSchema(mval, fd)
		if err != nil {
			return err
		}

		val := document.NewCBORValue(fd.Typ, cval)
		fieldKey := c.getFieldKey(key, mfield)
		c, err := c.saveDocValue(txn, c.getPrimaryIndexDocKey(fieldKey), val)
		if err != nil {
			return err
		}
		// links[mfield] = c
		links = append(links, core.DAGLink{
			Name: mfield,
			Cid:  c,
		})
	}

	// Update CompositeDAG
	em, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		return err
	}
	buf, err := em.Marshal(merge)
	if err != nil {
		return err
	}
	if _, err := c.saveValueToMerkleCRDT(txn, c.getPrimaryIndexDocKey(key), core.COMPOSITE, buf, links); err != nil {
		return err
	}

	// if this a a Batch masked as a Transaction
	// commit our writes so we can see them.
	// Batches don't maintain serializability, or
	// linearization, or any other transaction
	// semantics, which the user already knows
	// otherwise they wouldn't use a datastore
	// that doesnt support proper transactions.
	// So lets just commit, and keep going.
	// @todo: Change this on the Txn.BatchShim
	// structure
	if txn.IsBatch() {
		if err := txn.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldSchema takes a given value as an interface,
// and ensures it matches the supplied field description.
// It will do any minor parsing, like dates, and return
// the typed value again as an interface.
func validateFieldSchema(val interface{}, field base.FieldDescription) (interface{}, error) {
	var cval interface{}
	var err error
	var ok bool
	switch field.Kind {
	case base.FieldKind_DocKey, base.FieldKind_STRING:
		cval, ok = val.(string)
	case base.FieldKind_BOOL:
		cval, ok = val.(bool)
	case base.FieldKind_FLOAT, base.FieldKind_DECIMNAL:
		cval, ok = val.(float64)
	case base.FieldKind_DATE:
		var sval string
		sval, ok = val.(string)
		cval, err = time.Parse(time.RFC3339, sval)
	case base.FieldKind_INT:
		fval, ok := val.(float64)
		if !ok {
			return nil, ErrInvalidMergeValueType
		}
		cval = int64(fval)
	case base.FieldKind_OBJECT, base.FieldKind_OBJECT_ARRAY,
		base.FieldKind_FOREIGN_OBJECT, base.FieldKind_FOREIGN_OBJECT_ARRAY:
		err = errors.New("Merge doesn't support sub types yet")
	}

	if !ok {
		return nil, ErrInvalidMergeValueType
	}
	if err != nil {
		return nil, err
	}

	return cval, err
}

func (c *Collection) applyMergePatchOp(txn *Txn, docKey string, field string, currentVal interface{}, targetVal interface{}) error {
	return nil
}

// makeQuery constructs a simple query of the collection using the given filter.
// currently it doesn't support any other query operation other than filters.
// (IE: No limit, order, etc)
// Additionally it only queries for the root scalar fields of the object
func (c *Collection) makeSelectionQuery(txn *Txn, filter interface{}, opts ...client.UpdateOpt) (planner.Query, error) {
	var f *parser.Filter
	var err error
	switch fval := filter.(type) {
	case string:
		if fval == "" {
			return nil, errors.New("Invalid filter")
		}
		f, err = parser.NewFilterFromString(fval)
		if err != nil {
			return nil, err
		}
	case *parser.Filter:
		f = fval
	default:
		return nil, errors.New("Invalid filter")
	}
	if filter == "" {
		return nil, errors.New("Invalid filter")
	}
	slct, err := c.makeSelectLocal(f)
	if err != nil {
		return nil, err
	}

	return c.db.queryExecutor.MakeSelectQuery(c.db, txn, slct)
}

func (c *Collection) makeSelectLocal(filter *parser.Filter) (*parser.Select, error) {
	slct := &parser.Select{
		Name:   c.Name(),
		Filter: filter,
		Fields: make([]parser.Selection, len(c.desc.Schema.Fields)),
	}

	for i, fd := range c.Schema().Fields {
		if fd.IsObject() {
			continue
		}
		slct.Fields[i] = parser.Field{Name: fd.Name}
	}

	return slct, nil
}

// getTypeAndCollectionForPatch parses the Patch op path values
// and compares it against the collection schema.
// If its within the schema, then patchIsSubType is false
// subTypeName is empty.
// If the target type is an array, isArray is true.
// May need to query the database for other schema types
// which requires a db transaction. It is recommended
// to use collection.WithTxn(txn) for this function call.
func (c *Collection) getCollectionForPatchOpPath(txn *Txn, path string) (col *Collection, isArray bool, err error) {
	return nil, false, nil
}

// getTargetKeyForPatchPath walks through the given doc and Patch path.
// It returns the
func (c *Collection) getTargetKeyForPatchPath(txn *Txn, doc map[string]interface{}, path string) (string, error) {
	_, length := splitPatchPath(path)
	if length == 0 {
		return "", errors.New("Invalid patch op path")
	} else if length > 0 {

	}
	return "", nil
}

func splitPatchPath(path string) ([]string, int) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	pathParts := strings.Split(path, "/")
	return pathParts, len(pathParts)
}

func getValFromDocForPatchPath(doc map[string]interface{}, path string) (string, interface{}, bool) {
	pathParts, length := splitPatchPath(path)
	if length == 0 {
		return "", nil, false
	}
	return getMapProp(doc, pathParts, length)
}

func getMapProp(doc map[string]interface{}, paths []string, length int) (string, interface{}, bool) {
	val, ok := doc[paths[0]]
	if !ok {
		return "", nil, false
	}
	if length > 1 {
		doc, ok := val.(map[string]interface{})
		if !ok {
			return "", nil, false
		}
		return getMapProp(doc, paths[1:], length-1)
	}
	return paths[0], val, true
}

type UpdateResult struct {
	Count   int64
	DocKeys []string
}

type patcher interface{}

func parseUpdater(updater interface{}) (patcher, error) {
	switch v := updater.(type) {
	case string:
		return parseUpdaterString(v)
	case []interface{}:
		return parseUpdaterSlice(v)
	case []map[string]interface{}, map[string]interface{}:
		return patcher(v), nil
	case nil:
		return nil, ErrUpdateEmpty
	default:
		return nil, ErrInvalidUpdater
	}
}

func parseUpdaterString(v string) (patcher, error) {
	if v == "" {
		return nil, ErrUpdateEmpty
	}
	var i interface{}
	if err := json.Unmarshal([]byte(v), &i); err != nil {
		return nil, err
	}
	return parseUpdater(i)
}

// converts an []interface{} to []map[string]interface{}
// which is required to be an array of Patch Ops
func parseUpdaterSlice(v []interface{}) (patcher, error) {
	if len(v) == 0 {
		return nil, ErrUpdateEmpty
	}

	patches := make([]map[string]interface{}, len(v))
	for i, patch := range v {
		p, ok := patch.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidUpdater
		}
		patches[i] = p
	}

	return parseUpdater(patches)
}

/*

filter := NewFilterFromString("Name: {_eq: 'bob'}")

filter := db.NewQuery().And()

*/
