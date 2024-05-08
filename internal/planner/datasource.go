// Copyright 2022 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type planSource struct {
	collection client.Collection
	plan       planNode
}

func (p *Planner) getSource(parsed *mapper.Select) (planSource, error) {
	// for now, we only handle simple collection scannodes
	return p.getCollectionScanPlan(parsed)
}

func (p *Planner) getCollectionScanPlan(mapperSelect *mapper.Select) (planSource, error) {
	col, err := p.db.GetCollectionByName(p.ctx, mapperSelect.CollectionName)
	if err != nil {
		return planSource{}, err
	}

	var plan planNode
	if len(col.Description().QuerySources()) > 0 {
		var err error
		plan, err = p.View(mapperSelect, col)
		if err != nil {
			return planSource{}, err
		}
	} else {
		var err error
		plan, err = p.Scan(mapperSelect, col.Description())
		if err != nil {
			return planSource{}, err
		}
	}

	return planSource{
		plan:       plan,
		collection: col,
	}, nil
}
