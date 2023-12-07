// Copyright 2022 Democratized Data Foundation
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
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/planner"
)

// UpdateWith updates a target document using the given updater type. Target
// can be a Filter statement, a single docKey, a single document,
// an array of docKeys, or an array of documents.
// If you want more type safety, use the respective typed versions of Update.
// Eg: UpdateWithFilter or UpdateWithKey
func (c *collection) UpdateWith(
	ctx context.Context,
	target any,
	updater string,
) (*client.UpdateResult, error) {
	switch t := target.(type) {
	case string, map[string]any, *request.Filter:
		return c.UpdateWithFilter(ctx, t, updater)
	case client.DocKey:
		return c.UpdateWithKey(ctx, t, updater)
	case []client.DocKey:
		return c.UpdateWithKeys(ctx, t, updater)
	default:
		return nil, client.ErrInvalidUpdateTarget
	}
}

// UpdateWithFilter updates using a filter to target documents for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithFilter(
	ctx context.Context,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithFilter(ctx, txn, filter, updater)
	if err != nil {
		return nil, err
	}
	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKey updates using a DocKey to target a single document for update.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKey(
	ctx context.Context,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKey(ctx, txn, key, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

// UpdateWithKeys is the same as UpdateWithKey but accepts multiple keys as a slice.
// An updater value is provided, which could be a string Patch, string Merge Patch
// or a parsed Patch, or parsed Merge Patch.
func (c *collection) UpdateWithKeys(
	ctx context.Context,
	keys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)
	res, err := c.updateWithKeys(ctx, txn, keys, updater)
	if err != nil {
		return nil, err
	}

	return res, c.commitImplicitTxn(ctx, txn)
}

func (c *collection) updateWithKey(
	ctx context.Context,
	txn datastore.Txn,
	key client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
	}

	doc, err := c.Get(ctx, key, false)
	if err != nil {
		return nil, err
	}

	if isPatch {
		// todo
	} else {
		err = c.applyMergeToDoc(doc, parsedUpdater.GetObject())
	}
	if err != nil {
		return nil, err
	}

	_, err = c.save(ctx, txn, doc, false)
	if err != nil {
		return nil, err
	}

	results := &client.UpdateResult{
		Count:   1,
		DocKeys: []string{key.String()},
	}
	return results, nil
}

func (c *collection) updateWithKeys(
	ctx context.Context,
	txn datastore.Txn,
	keys []client.DocKey,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	if parsedUpdater.Type() == fastjson.TypeArray {
		isPatch = true
	} else if parsedUpdater.Type() != fastjson.TypeObject {
		return nil, client.ErrInvalidUpdater
	}

	results := &client.UpdateResult{
		DocKeys: make([]string, len(keys)),
	}
	for i, key := range keys {
		doc, err := c.Get(ctx, key, false)
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else {
			err = c.applyMergeToDoc(doc, parsedUpdater.GetObject())
		}
		if err != nil {
			return nil, err
		}

		_, err = c.save(ctx, txn, doc, false)
		if err != nil {
			return nil, err
		}

		results.DocKeys[i] = key.String()
		results.Count++
	}
	return results, nil
}

func (c *collection) updateWithFilter(
	ctx context.Context,
	txn datastore.Txn,
	filter any,
	updater string,
) (*client.UpdateResult, error) {
	parsedUpdater, err := fastjson.Parse(updater)
	if err != nil {
		return nil, err
	}

	isPatch := false
	isMerge := false
	switch parsedUpdater.Type() {
	case fastjson.TypeArray:
		isPatch = true
	case fastjson.TypeObject:
		isMerge = true
	default:
		return nil, client.ErrInvalidUpdater
	}

	// Make a selection plan that will scan through only the documents with matching filter.
	selectionPlan, err := c.makeSelectionPlan(ctx, txn, filter)
	if err != nil {
		return nil, err
	}

	err = selectionPlan.Init()
	if err != nil {
		return nil, err
	}

	if err = selectionPlan.Start(); err != nil {
		return nil, err
	}

	// If the plan isn't properly closed at any exit point log the error.
	defer func() {
		if err := selectionPlan.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close the selection plan, after filter update", err)
		}
	}()

	results := &client.UpdateResult{
		DocKeys: make([]string, 0),
	}

	docMap := selectionPlan.DocumentMap()

	// Keep looping until results from the selection plan have been iterated through.
	for {
		next, nextErr := selectionPlan.Next()
		if nextErr != nil {
			return nil, err
		}
		// if theres no more records from the request, jump out of the loop
		if !next {
			break
		}

		// Get the document, and apply the patch
		docAsMap := docMap.ToMap(selectionPlan.Value())
		doc, err := client.NewDocFromMap(docAsMap)
		if err != nil {
			return nil, err
		}

		if isPatch {
			// todo
		} else if isMerge { // else is fine here
			err = c.applyMergeToDoc(doc, parsedUpdater.GetObject())
		}
		if err != nil {
			return nil, err
		}

		_, err = c.save(ctx, txn, doc, false)
		if err != nil {
			return nil, err
		}

		// add successful updated doc to results
		results.DocKeys = append(results.DocKeys, doc.Key().String())
		results.Count++
	}

	return results, nil
}

// applyMergeToDoc applies the given json merge to the given Defra doc.
//
// It does not save the document.
func (c *collection) applyMergeToDoc(
	doc *client.Document,
	merge *fastjson.Object,
) error {
	mergeMap := make(map[string]*fastjson.Value)
	merge.Visit(func(k []byte, v *fastjson.Value) {
		mergeMap[string(k)] = v
	})

	for mfield, mval := range mergeMap {
		fd, isValidField := c.Schema().GetField(mfield)
		if !isValidField {
			return client.NewErrFieldNotExist(mfield)
		}

		if fd.Kind == client.FieldKind_FOREIGN_OBJECT {
			fd, isValidField = c.Schema().GetField(mfield + request.RelatedObjectID)
			if !isValidField {
				return client.NewErrFieldNotExist(mfield)
			}
		}

		cborVal, err := validateFieldSchema(mval, fd)
		if err != nil {
			return err
		}

		err = doc.Set(fd.Name, cborVal)
		if err != nil {
			return err
		}
	}

	return nil
}

// isSecondaryIDField returns true if the given field description represents a secondary relation field ID.
func (c *collection) isSecondaryIDField(fieldDesc client.FieldDescription) (client.FieldDescription, bool) {
	if fieldDesc.RelationType != client.Relation_Type_INTERNAL_ID {
		return client.FieldDescription{}, false
	}

	relationFieldDescription, valid := c.Schema().GetField(
		strings.TrimSuffix(fieldDesc.Name, request.RelatedObjectID),
	)
	return relationFieldDescription, valid && !relationFieldDescription.IsPrimaryRelation()
}

// patchPrimaryDoc patches the (primary) document linked to from the document of the given dockey via the
// given (secondary) relationship field description (hosted on the collection of the document matching the
// given dockey).
//
// The given field value should be the string representation of the dockey of the primary document to be
// patched.
func (c *collection) patchPrimaryDoc(
	ctx context.Context,
	txn datastore.Txn,
	secondaryCollectionName string,
	relationFieldDescription client.FieldDescription,
	docKey string,
	fieldValue string,
) error {
	primaryDockey, err := client.NewDocKeyFromString(fieldValue)
	if err != nil {
		return err
	}

	primaryCol, err := c.db.getCollectionByName(ctx, txn, relationFieldDescription.Schema)
	if err != nil {
		return err
	}
	primaryCol = primaryCol.WithTxn(txn)
	primarySchema := primaryCol.Schema()

	primaryField, ok := primaryCol.Description().GetFieldByRelation(
		relationFieldDescription.RelationName,
		secondaryCollectionName,
		relationFieldDescription.Name,
		&primarySchema,
	)
	if !ok {
		return client.NewErrFieldNotExist(relationFieldDescription.RelationName)
	}

	primaryIDField, ok := primaryCol.Schema().GetField(primaryField.Name + request.RelatedObjectID)
	if !ok {
		return client.NewErrFieldNotExist(primaryField.Name + request.RelatedObjectID)
	}

	doc, err := primaryCol.Get(
		ctx,
		primaryDockey,
		false,
	)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	// If the document doesn't exist then there is nothing to update.
	if doc == nil {
		return nil
	}

	existingVal, err := doc.GetValue(primaryIDField.Name)
	if err != nil && !errors.Is(err, client.ErrFieldNotExist) {
		return err
	}

	if existingVal != nil && existingVal.Value() != "" && existingVal.Value() != docKey {
		return NewErrOneOneAlreadyLinked(docKey, fieldValue, relationFieldDescription.RelationName)
	}

	err = doc.Set(primaryIDField.Name, docKey)
	if err != nil {
		return err
	}

	err = primaryCol.Update(ctx, doc)
	if err != nil {
		return err
	}

	return nil
}

// validateFieldSchema takes a given value as an interface,
// and ensures it matches the supplied field description.
// It will do any minor parsing, like dates, and return
// the typed value again as an interface.
func validateFieldSchema(val *fastjson.Value, field client.FieldDescription) (any, error) {
	switch field.Kind {
	case client.FieldKind_DocKey, client.FieldKind_STRING:
		return getString(val)

	case client.FieldKind_STRING_ARRAY:
		return getArray(val, getString)

	case client.FieldKind_NILLABLE_STRING_ARRAY:
		return getNillableArray(val, getString)

	case client.FieldKind_BOOL:
		return getBool(val)

	case client.FieldKind_BOOL_ARRAY:
		return getArray(val, getBool)

	case client.FieldKind_NILLABLE_BOOL_ARRAY:
		return getNillableArray(val, getBool)

	case client.FieldKind_FLOAT:
		return getFloat64(val)

	case client.FieldKind_FLOAT_ARRAY:
		return getArray(val, getFloat64)

	case client.FieldKind_NILLABLE_FLOAT_ARRAY:
		return getNillableArray(val, getFloat64)

	case client.FieldKind_DATETIME:
		// @TODO: Requires Typed Document refactor
		// to handle this correctly.
		// For now, we will persist DateTime as a
		// RFC3339 string
		// see https://github.com/sourcenetwork/defradb/issues/935
		return getString(val)

	case client.FieldKind_INT:
		return getInt64(val)

	case client.FieldKind_INT_ARRAY:
		return getArray(val, getInt64)

	case client.FieldKind_NILLABLE_INT_ARRAY:
		return getNillableArray(val, getInt64)

	case client.FieldKind_FOREIGN_OBJECT, client.FieldKind_FOREIGN_OBJECT_ARRAY:
		return nil, NewErrFieldOrAliasToFieldNotExist(field.Name)

	case client.FieldKind_BLOB:
		return getString(val)
	}

	return nil, client.NewErrUnhandledType("FieldKind", field.Kind)
}

func getString(v *fastjson.Value) (string, error) {
	b, err := v.StringBytes()
	return string(b), err
}

func getBool(v *fastjson.Value) (bool, error) {
	return v.Bool()
}

func getFloat64(v *fastjson.Value) (float64, error) {
	return v.Float64()
}

func getInt64(v *fastjson.Value) (int64, error) {
	return v.Int64()
}

func getArray[T any](
	val *fastjson.Value,
	typeGetter func(*fastjson.Value) (T, error),
) ([]T, error) {
	if val.Type() == fastjson.TypeNull {
		return nil, nil
	}

	valArray, err := val.Array()
	if err != nil {
		return nil, err
	}

	arr := make([]T, len(valArray))
	for i, arrItem := range valArray {
		if arrItem.Type() == fastjson.TypeNull {
			continue
		}
		arr[i], err = typeGetter(arrItem)
		if err != nil {
			return nil, err
		}
	}

	return arr, nil
}

func getNillableArray[T any](
	val *fastjson.Value,
	typeGetter func(*fastjson.Value) (T, error),
) ([]*T, error) {
	if val.Type() == fastjson.TypeNull {
		return nil, nil
	}

	valArray, err := val.Array()
	if err != nil {
		return nil, err
	}

	arr := make([]*T, len(valArray))
	for i, arrItem := range valArray {
		if arrItem.Type() == fastjson.TypeNull {
			continue
		}
		v, err := typeGetter(arrItem)
		if err != nil {
			return nil, err
		}
		arr[i] = &v
	}

	return arr, nil
}

// makeSelectionPlan constructs a simple read-only plan of the collection using the given filter.
// currently it doesn't support any other operations other than filters.
// (IE: No limit, order, etc)
// Additionally it only requests for the root scalar fields of the object
func (c *collection) makeSelectionPlan(
	ctx context.Context,
	txn datastore.Txn,
	filter any,
) (planner.RequestPlan, error) {
	var f immutable.Option[request.Filter]
	var err error
	switch fval := filter.(type) {
	case string:
		if fval == "" {
			return nil, ErrInvalidFilter
		}

		f, err = c.db.parser.NewFilterFromString(c.Name(), fval)
		if err != nil {
			return nil, err
		}
	case immutable.Option[request.Filter]:
		f = fval
	default:
		return nil, ErrInvalidFilter
	}

	slct, err := c.makeSelectLocal(f)
	if err != nil {
		return nil, err
	}

	planner := planner.New(ctx, c.db.WithTxn(txn), txn)
	return planner.MakePlan(&request.Request{
		Queries: []*request.OperationDefinition{
			{
				Selections: []request.Selection{
					slct,
				},
			},
		},
	})
}

func (c *collection) makeSelectLocal(filter immutable.Option[request.Filter]) (*request.Select, error) {
	slct := &request.Select{
		Field: request.Field{
			Name: c.Name(),
		},
		Filter: filter,
		Fields: make([]request.Selection, 0),
	}

	for _, fd := range c.Schema().Fields {
		if fd.IsObject() {
			continue
		}
		slct.Fields = append(slct.Fields, &request.Field{
			Name: fd.Name,
		})
	}

	return slct, nil
}
