// Copyright 2023 Democratized Data Foundation
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
	"testing"

	"github.com/stretchr/testify/require"
)

// An in-process caller that predates the kind carrying fields sets only the deprecated top-level
// Fields. Nothing normalizes a struct built in memory, so the read has to fall back to it.
func TestIndexDescription_CompatFieldsOnly_ReadsThrough(t *testing.T) {
	d := IndexDescription{
		Name:            "x",
		ID:              1,
		Fields:          []IndexedFieldDescription{{Name: "age"}},
		Kind:            IndexKindOrdered,
		KindDescription: &OrderedIndexDescription{Unique: true},
	}
	require.Equal(t, []string{"age"}, d.fieldNames())
}

// A descriptor built in memory carries directions only on the deprecated field. Reading it must not
// reduce them to names, or a descending composite index silently becomes ascending.
func TestIndexDescription_CompatFieldsOnly_KeepsDirection(t *testing.T) {
	d := IndexDescription{
		Name: "x", ID: 1, Kind: IndexKindOrdered,
		Fields: []IndexedFieldDescription{{Name: "age", Descending: true}},
	}
	require.Equal(t, d.Fields, d.GetFields())
}
