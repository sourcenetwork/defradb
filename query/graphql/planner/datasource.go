// Copyright 2020 Source Inc.
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
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/db/base"
)

// sourceInfo stores info about the data source
type sourceInfo struct {
	collectionDescription base.CollectionDescription
	// and more
}

type planSource struct {
	info sourceInfo
	plan planNode
}

// datasource is a set of utilities for constructing scan/index/join nodes
// from a given query statement
func (p *Planner) getSource(collection string, versioned bool) (planSource, error) {
	// for now, we only handle simple collection scannodes
	return p.getCollectionScanPlan(collection, versioned)
}

// @todo: Add field selection
func (p *Planner) getCollectionScanPlan(collection string, versioned bool) (planSource, error) {
	if collection == "" {
		return planSource{}, fmt.Errorf("collection name cannot be empty")
	}
	colDesc, err := p.getCollectionDesc(collection)
	if err != nil {
		return planSource{}, err
	}

	scan := p.Scan(versioned)
	err = scan.initCollection(colDesc)
	if err != nil {
		return planSource{}, err
	}

	return planSource{
		plan: scan,
		info: sourceInfo{
			collectionDescription: colDesc,
		},
	}, nil
}

func (p *Planner) getCollectionDesc(name string) (base.CollectionDescription, error) {
	key := base.MakeCollectionSystemKey(name)
	var desc base.CollectionDescription
	buf, err := p.txn.Systemstore().Get(p.ctx, key.ToDS())
	if err != nil {
		return desc, fmt.Errorf("Failed to get collection description: %w", err)
	}

	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return desc, err
	}

	return desc, nil
}
