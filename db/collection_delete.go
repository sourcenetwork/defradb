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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/planner"
)

type DeleteOpt struct{}

var (
	ErrInvalidDeleteTarget = errors.New("The doc targeter is an unknown type")
	ErrInvalidDeleter      = errors.New("The doc deleter is an unknown type")
	ErrDeleteEmpty         = errors.New("The doc delete cannot be empty")
)

// Delete2 deletes the given doc. It will scan through the field/value pairs
// and find those marked for delete, and apply the appropriate delete.
// Delete only works on root level field/value pairs. So not foreign or related
// types can be deleted. If you wish to delete sub types, use DeleteWith, and supply
// an delete payload in the form of a Patch or a Merge Patch.
func (c *Collection) Delete2(doc *document.SimpleDocument, opts ...client.DeleteOpt) error {
	return nil
}

// DeleteWith deletes a target document using the given deleter type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Delete.
// Eg: DeleteWithFilter or DeleteWithKey
func (c *Collection) DeleteWith(ctx context.Context, target interface{}, deleter interface{}, opts ...client.DeleteOpt) error {
	switch t := target.(type) {
	case string, map[string]interface{}, *parser.Filter:
		_, err := c.DeleteWithFilter(ctx, t, deleter, opts...)
		return err
	case key.DocKey:
		_, err := c.DeleteWithKey(ctx, t, deleter, opts...)
		return err
	case []key.DocKey:
		_, err := c.DeleteWithKeys(ctx, t, deleter, opts...)
		return err
	case *document.SimpleDocument:
		return c.DeleteWithDoc(t, deleter, opts...)
	case []*document.SimpleDocument:
		return c.DeleteWithDocs(t, deleter, opts...)
	default:
		return ErrInvalidDeleteTarget
	}
}

// DeleteWithFilter deletes using a filter to target documents for delete.
// A deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithFilter(ctx context.Context, filter interface{}, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.deleteWithFilter(ctx, txn, filter, deleter, opts...)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithKey deletes using a DocKey to target a single document for delete.
// A deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithKey(ctx context.Context, key key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.deleteWithKey(ctx, txn, key, deleter, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithKeys is the same as DeleteWithKey but accepts multiple keys as a slice.
// A deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithKeys(ctx context.Context, keys []key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.deleteWithKeys(ctx, txn, keys, deleter, opts...)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// DeleteWithDoc deletes targeting the supplied document.
// A deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithDoc(doc *document.SimpleDocument, deleter interface{}, opts ...client.DeleteOpt) error {
	return nil
}

// DeleteWithDocs deletes all the supplied documents in the slice.
// A deleter value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *Collection) DeleteWithDocs(docs []*document.SimpleDocument, deleter interface{}, opts ...client.DeleteOpt) error {
	return nil
}

func (c *Collection) deleteWithKey(ctx context.Context, txn *Txn, key key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	patch, err := parseDeleter(deleter)
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
		return nil, ErrInvalidDeleter
	}

	doc, err := c.Get(ctx, key)
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
		err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
	}
	if err != nil {
		return nil, err
	}

	results := &client.DeleteResult{
		Count:   1,
		DocKeys: []string{key.String()},
	}
	return results, nil
}

func (c *Collection) deleteWithKeys(ctx context.Context, txn *Txn, keys []key.DocKey, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	fmt.Println("updating keys:", keys)
	patch, err := parseDeleter(deleter)
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
		return nil, ErrInvalidDeleter
	}

	results := &client.DeleteResult{
		DocKeys: make([]string, len(keys)),
	}
	for i, key := range keys {
		doc, err := c.Get(ctx, key)
		if err != nil {
			fmt.Println("error getting key to delete:", key)
			return nil, err
		}
		v, err := doc.ToMap()
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else {
			err = c.applyMerge(ctx, txn, v, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, nil
		}

		results.DocKeys[i] = key.String()
		results.Count++
	}
	return results, nil
}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
// func makeCollectionDataKey(collectionID uint32) ds.Key {
// 	return collectionNs.ChildString(name)
// }

// makeQuery constructs a simple query of the collection using the given filter.
// currently it doesn't support any other query operation other than filters.
// (IE: No limit, order, etc)
// Additionally it only queries for the root scalar fields of the object
func (c *Collection) makeSelectionDeleteQuery(
	ctx context.Context,
	txn *Txn,
	filter interface{},
	opts ...client.DeleteOpt) (planner.Query, error) {
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
	slct, err := c.makeSelectDeleteLocal(f)
	if err != nil {
		return nil, err
	}

	return c.db.queryExecutor.MakeSelectQuery(ctx, c.db, txn, slct)
}

func (c *Collection) deleteWithFilter(ctx context.Context, txn *Txn, filter interface{}, deleter interface{}, opts ...client.DeleteOpt) (*client.DeleteResult, error) {
	patch, err := parseDeleter(deleter)
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
		return nil, ErrInvalidDeleter
	}

	// scan through docs with filter
	query, err := c.makeSelectionDeleteQuery(ctx, txn, filter, opts...)
	if err != nil {
		return nil, err
	}
	if err := query.Start(); err != nil {
		return nil, err
	}

	results := &client.DeleteResult{
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
			err = c.applyMerge(ctx, txn, doc, patch.(map[string]interface{}))
		}
		if err != nil {
			return nil, err
		}

		// add succesful deleted doc to results
		results.DocKeys = append(results.DocKeys, doc["_key"].(string))
		results.Count++
	}

	return results, nil
}

func (c *Collection) makeSelectDeleteLocal(filter *parser.Filter) (*parser.Select, error) {
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

type DeleteResult struct {
	Count   int64
	DocKeys []string
}

func parseDeleter(deleter interface{}) (patcher, error) {
	switch v := deleter.(type) {
	case string:
		return parseDeleterString(v)
	case []interface{}:
		return parseDeleterSlice(v)
	case []map[string]interface{}, map[string]interface{}:
		return patcher(v), nil
	case nil:
		return nil, ErrDeleteEmpty
	default:
		return nil, ErrInvalidDeleter
	}
}

func parseDeleterString(v string) (patcher, error) {
	if v == "" {
		return nil, ErrDeleteEmpty
	}
	var i interface{}
	if err := json.Unmarshal([]byte(v), &i); err != nil {
		return nil, err
	}
	return parseDeleter(i)
}

// converts an []interface{} to []map[string]interface{}
// which is required to be an array of Patch Ops
func parseDeleterSlice(v []interface{}) (patcher, error) {
	if len(v) == 0 {
		return nil, ErrDeleteEmpty
	}

	patches := make([]map[string]interface{}, len(v))
	for i, patch := range v {
		p, ok := patch.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidDeleter
		}
		patches[i] = p
	}

	return parseDeleter(patches)
}

/*

filter := NewFilterFromString("Name: {_eq: 'bob'}")

filter := db.NewQuery().And()

*/
