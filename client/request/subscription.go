// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import (
	"github.com/sourcenetwork/immutable"
)

// ObjectSubscription is a field on the SubscriptionType
// of a graphql request. It includes all the possible
// arguments
type ObjectSubscription struct {
	Field
	ChildSelect

	Filterable

	// Collection is the target collection name
	Collection string
}

// ToSelect returns a basic Select object, with the same Name, Alias, and Fields as
// the Subscription object. Used to create a Select planNode for the event stream return objects.
func (m ObjectSubscription) ToSelect(docID, cid string) *Select {
	return &Select{
		Field: Field{
			Name:  m.Collection,
			Alias: m.Alias,
		},
		DocIDsFilter: DocIDsFilter{
			DocIDs: immutable.Some([]string{docID}),
		},
		CIDFilter: CIDFilter{
			immutable.Some(cid),
		},
		ChildSelect: m.ChildSelect,
		Filterable:  m.Filterable,
	}
}
