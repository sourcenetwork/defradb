// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"encoding/json"

	"github.com/sourcenetwork/immutable"
)

// collectionVersionDisplay is the JSON-marshalable form of a [CollectionVersion] with each
// field's Kind and Typ, and each index's Kind, rendered as their string forms (e.g. "String", "[Int!]", "lww", "ordered")
// rather than the numeric IDs used on the wire and on disk.
//
// It is intentionally not a MarshalJSON on [CollectionVersion]: the latter is marshalled
// as-is when persisted and when evaluating JSON-patch operations, both of which must keep
// emitting the numeric form. The display form round-trips: unmarshalling it back into a
// [CollectionVersion] yields the original.
//
// [CollectionVersion] is embedded so fields added to it carry through automatically; only
// Fields and Indexes are shadowed (encoding/json picks the shallower fields and suppresses the embedded ones).
type collectionVersionDisplay struct {
	CollectionVersion

	Fields  []collectionFieldDisplay
	Indexes []collectionIndexDisplay
}

// collectionFieldDisplay mirrors [CollectionFieldDescription]; the field set and order must
// match its unmarshal counterpart so the rendered output round-trips losslessly.
type collectionFieldDisplay struct {
	FieldID      string
	Name         string
	Kind         json.RawMessage
	Typ          string
	RelationName immutable.Option[string]
	IsPrimary    bool
	DefaultValue any
	Size         int
}

// collectionIndexDisplay mirrors [IndexDescription]; the field set and order must match its
// unmarshal counterpart so the rendered output round-trips losslessly.
type collectionIndexDisplay struct {
	Name            string
	ID              uint32
	Fields          []IndexedFieldDescription
	Kind            string
	KindDescription json.RawMessage
	Unique          bool
}

// Display returns this [IndexDescription] as JSON in which its Kind is rendered as a
// human-readable string (e.g. "ordered", "vector") rather than the numeric ID used on
// the wire and on disk.
func (d IndexDescription) Display() (json.RawMessage, error) {
	d = d.Normalize()
	kindDescription, err := json.Marshal(d.KindDescription)
	if err != nil {
		return nil, err
	}
	return json.Marshal(collectionIndexDisplay{
		Name:            d.Name,
		ID:              d.ID,
		Fields:          d.Fields,
		Kind:            d.Kind.String(),
		KindDescription: kindDescription,
		Unique:          d.Unique,
	})
}

// listIndexesResultDisplay mirrors [ListIndexesResult] for display output.
type listIndexesResultDisplay struct {
	CollectionName string
	Description    json.RawMessage
	Execution      ActionExecution
}

// Display returns this [ListIndexesResult] as JSON with its Description.Kind rendered as a string.
func (r ListIndexesResult) Display() (json.RawMessage, error) {
	desc, err := r.Description.Display()
	if err != nil {
		return nil, err
	}
	return json.Marshal(listIndexesResultDisplay{
		CollectionName: r.CollectionName,
		Description:    desc,
		Execution:      r.Execution,
	})
}

// DisplayListIndexesResults applies [ListIndexesResult.Display] to each of the given results.
func DisplayListIndexesResults(results []ListIndexesResult) ([]json.RawMessage, error) {
	display := make([]json.RawMessage, len(results))
	for i, r := range results {
		raw, err := r.Display()
		if err != nil {
			return nil, err
		}
		display[i] = raw
	}
	return display, nil
}

// Display returns this [CollectionVersion] as JSON in which each field's Kind and Typ,
// and each index's Kind, are rendered as human-readable strings (e.g. "String", "[Int!]",
// "lww", "ordered") rather than the numeric IDs used on the wire and on disk.
//
// It is intended for introspection and display, e.g. by `collection describe`. The result is
// a [json.RawMessage] so callers can embed it in a larger document without re-encoding; for a
// printable string use string(raw). The output round-trips back into a [CollectionVersion].
func (col CollectionVersion) Display() (json.RawMessage, error) {
	fields := make([]collectionFieldDisplay, len(col.Fields))
	for i, f := range col.Fields {
		kind, err := MarshalFieldKindToJSON(f.Kind)
		if err != nil {
			return nil, err
		}
		fields[i] = collectionFieldDisplay{
			FieldID:      f.FieldID,
			Name:         f.Name,
			Kind:         kind,
			Typ:          f.Typ.String(),
			RelationName: f.RelationName,
			IsPrimary:    f.IsPrimary,
			DefaultValue: f.DefaultValue,
			Size:         f.Size,
		}
	}
	indexes := make([]collectionIndexDisplay, len(col.Indexes))
	for i, index := range col.Indexes {
		index = index.Normalize()
		kindDescription, err := json.Marshal(index.KindDescription)
		if err != nil {
			return nil, err
		}
		indexes[i] = collectionIndexDisplay{
			Name:            index.Name,
			ID:              index.ID,
			Fields:          index.Fields,
			Kind:            index.Kind.String(),
			KindDescription: kindDescription,
			Unique:          index.Unique,
		}
	}
	return json.Marshal(collectionVersionDisplay{
		CollectionVersion: col,
		Fields:            fields,
		Indexes:           indexes,
	})
}

// DisplayCollectionVersions applies [CollectionVersion.Display] to each of the given versions.
func DisplayCollectionVersions(cols []CollectionVersion) ([]json.RawMessage, error) {
	display := make([]json.RawMessage, len(cols))
	for i, col := range cols {
		raw, err := col.Display()
		if err != nil {
			return nil, err
		}
		display[i] = raw
	}
	return display, nil
}
