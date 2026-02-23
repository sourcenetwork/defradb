// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// isOrderingByRelation returns true if the ordering involves a relation field.
// This is detected by checking if any order condition has more than one field index
// (indicating traversal through a relation).
func (r *primaryObjectsRetriever) isOrderingByRelation() bool {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return true
		}
	}
	return false
}

// orderingRelFieldIsPrimary returns true if the fetched doc is the primary side of the
// ordering relation (i.e. stores the FK). When true, orphans can be identified directly
// via FK IS NULL. When false, orphans can only be identified by exclusion from join results.
func (r *primaryObjectsRetriever) orderingRelFieldIsPrimary() bool {
	_, relFieldIndex := r.getOrderingInfo()
	fieldName, ok := r.primaryScan.documentMapping.TryToFindNameFromIndex(relFieldIndex)
	if !ok {
		return false
	}
	fieldDef, ok := r.primarySide.col.Version().GetFieldByName(fieldName)
	if !ok {
		return false
	}
	return fieldDef.IsPrimary
}

// getOrderingInfo returns the sort direction and relation field index if the ordering involves a relation field.
func (r *primaryObjectsRetriever) getOrderingInfo() (*mapper.SortDirection, int) {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return &order.Direction, order.FieldIndexes[0]
		}
	}
	return nil, 0
}
