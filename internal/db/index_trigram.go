// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
)

// TrigramFieldGenerator fans a string field value out into one index entry per distinct
// trigram of it. This is the same shape as ArrayFieldGenerator - one field value, many
// entries - so the trigram index is the ordinary non-unique index over these values.
type TrigramFieldGenerator struct{}

func (g *TrigramFieldGenerator) Generate(value client.NormalValue, f func(client.NormalValue) error) error {
	str, ok := value.String()
	if !ok {
		nillableStr, isString := value.NillableString()
		if !isString || !nillableStr.HasValue() {
			// A nil value has no trigrams, so it gets no entry. Nothing is lost: a trigram index
			// only serves substring and pattern matching, and neither matches a nil value.
			return nil
		}
		str = nillableStr.Value()
	}

	for _, trigram := range core.Trigrams(str) {
		if err := f(client.NewNormalString(trigram)); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	// A trigram index writes one ordinary key per (trigram, document) with an empty value, which
	// is exactly what the non-unique ordered index does once the generator has fanned the value
	// out. It therefore needs no implementation of its own.
	registerIndexKind(client.IndexKindTrigram, func(base collectionBaseIndex) client.CollectionIndex {
		return &collectionSimpleIndex{collectionBaseIndex: base}
	})
}

// validateTrigramIndex checks a new trigram index request against the collection.
func validateTrigramIndex(def client.CollectionVersion, req client.NewIndexRequest) error {
	if req.Unique {
		return ErrTrigramIndexUnique
	}
	if len(req.Fields) != 1 {
		return ErrTrigramIndexNotSingleField
	}
	field, found := def.GetFieldByName(req.Fields[0].Name)
	if !found {
		return NewErrNonExistingFieldForIndex(req.Fields[0].Name)
	}
	if field.Kind != client.FieldKind_STRING && field.Kind != client.FieldKind_NILLABLE_STRING {
		return NewErrTrigramIndexFieldNotString(field.Name, field.Kind.String())
	}
	return nil
}
