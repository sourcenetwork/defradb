// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import (
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
)

// Operation represents an operation such as query or mutation.
//
// It wraps child Selects belonging to this operation.
type Operation struct {
	// The document mapping for this select, describing how items yielded
	// for this select can be accessed and rendered.
	*core.DocumentMapping

	// Selects is the list of selections in the operation.
	Selects []*Select

	// Mutations is the list of mutations in the operation.
	Mutations []*Mutation

	// CommitSelects is the list of commit selections in the operation.
	CommitSelects []*CommitSelect
}

// addSelection adds a new selection to the operation's document mapping.
// The request.Field is used as the key for the core.RenderKey and the Select
// document mapping is added as a child field.
func (o *Operation) addSelection(i int, f request.Field, s Select) {
	renderKey := core.RenderKey{Index: i}
	if f.Alias.HasValue() {
		renderKey.Key = f.Alias.Value()
	} else {
		renderKey.Key = f.Name
	}
	o.DocumentMapping.Add(i, s.Name)
	o.DocumentMapping.SetChildAt(i, s.DocumentMapping)
	o.DocumentMapping.RenderKeys = append(o.DocumentMapping.RenderKeys, renderKey)
}
