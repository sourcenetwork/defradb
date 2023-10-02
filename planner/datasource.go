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
	"github.com/sourcenetwork/defradb/planner/mapper"
)

type planSource struct {
	collection client.Collection
	plan       planNode
}

func (p *Planner) getSource(parsed *mapper.Select) (planSource, error) {
	// for now, we only handle simple collection scannodes
	return p.getCollectionScanPlan(parsed)
}

func (p *Planner) getCollectionScanPlan(parsed *mapper.Select) (planSource, error) {
	col, err := p.db.GetCollectionByName(p.ctx, parsed.CollectionName)
	if err != nil {
		return planSource{}, err
	}

	scan, err := p.Scan(parsed)
	if err != nil {
		return planSource{}, err
	}

	return planSource{
		plan:       scan,
		collection: col,
	}, nil
}
