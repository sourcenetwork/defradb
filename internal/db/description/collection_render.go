// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"encoding/json"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// RenderedCollectionVersion wraps a [client.CollectionVersion] for the `collection describe`
// output, rendering each field's Kind and Typ as human-readable strings (e.g. "String",
// "lww") rather than their numeric IDs. The rendered form round-trips back into a
// [client.CollectionVersion].
//
// This is deliberately a separate type rather than a MarshalJSON on [client.CollectionVersion]:
// that struct is marshalled as-is when persisted and when evaluating JSON-patch ops, both of
// which must keep emitting numbers.
//
// [client.CollectionVersion] is embedded (so new fields carry through automatically) with only
// Fields shadowed, encoding/json picks the shallower Fields and suppresses the embedded one.
type RenderedCollectionVersion struct {
	client.CollectionVersion

	Fields []renderedCollectionField
}

// renderedCollectionField mirrors [client.CollectionFieldDescription]; the field set and order
// must match its unmarshal counterpart so the rendered output round-trips losslessly.
type renderedCollectionField struct {
	FieldID      string
	Name         string
	Kind         json.RawMessage
	Typ          string
	RelationName immutable.Option[string]
	IsPrimary    bool
	DefaultValue any
	Size         int
}

// RenderCollectionVersion wraps a single [client.CollectionVersion] for human-readable output.
func RenderCollectionVersion(c client.CollectionVersion) (RenderedCollectionVersion, error) {
	fields := make([]renderedCollectionField, len(c.Fields))
	for i, f := range c.Fields {
		kind, err := client.MarshalFieldKindToJSON(f.Kind)
		if err != nil {
			return RenderedCollectionVersion{}, err
		}
		fields[i] = renderedCollectionField{
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
	return RenderedCollectionVersion{CollectionVersion: c, Fields: fields}, nil
}

// RenderCollectionVersions renders each of the given versions; see [RenderCollectionVersion].
func RenderCollectionVersions(cols []client.CollectionVersion) ([]RenderedCollectionVersion, error) {
	rendered := make([]RenderedCollectionVersion, len(cols))
	for i, c := range cols {
		r, err := RenderCollectionVersion(c)
		if err != nil {
			return nil, err
		}
		rendered[i] = r
	}
	return rendered, nil
}
