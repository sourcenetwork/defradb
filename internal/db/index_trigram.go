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
	indextrigram "github.com/sourcenetwork/defradb/internal/index/trigram"
)

// TrigramFieldGenerator fans a string field value out into one ordinary non-unique index entry per
// distinct trigram. A nil value has no trigrams and therefore no entries.
type TrigramFieldGenerator struct{}

func (g *TrigramFieldGenerator) Generate(
	value client.NormalValue,
	f func(client.NormalValue) error,
) error {
	str, ok := value.String()
	if !ok {
		nillableStr, isString := value.NillableString()
		if !isString || !nillableStr.HasValue() {
			return nil
		}
		str = nillableStr.Value()
	}
	for _, value := range indextrigram.Extract(str) {
		if err := f(client.NewNormalString(value)); err != nil {
			return err
		}
	}
	return nil
}

// newCollectionTrigramIndex intentionally reuses the ordinary non-unique writer and key layout. It
// differs only in field-value generation; the planner and fetcher still treat it as a distinct kind.
func newCollectionTrigramIndex(base collectionBaseIndex) (client.CollectionIndex, error) {
	if _, ok := base.desc.GetTrigram(); !ok {
		return nil, NewErrCorruptedIndexKindDescription(base.desc.Name, base.desc.Kind)
	}
	base.fieldGenerators = []FieldIndexGenerator{&TrigramFieldGenerator{}}
	return &collectionSimpleIndex{collectionBaseIndex: base}, nil
}
