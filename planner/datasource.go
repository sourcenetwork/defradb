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
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// sourceInfo stores info about the data source
type sourceInfo struct {
	collectionDescription client.CollectionDescription
	// and more
}

type planSource struct {
	info sourceInfo
	plan planNode
}

func (p *Planner) getSource(parsed *mapper.Select) (planSource, error) {
	// for now, we only handle simple collection scannodes
	return p.getCollectionScanPlan(parsed)
}

func (p *Planner) getCollectionScanPlan(parsed *mapper.Select) (planSource, error) {
	colDesc, err := p.getCollectionDesc(parsed.CollectionName)
	if err != nil {
		return planSource{}, err
	}

	scan, err := p.Scan(parsed)
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

func (p *Planner) getCollectionDesc(name string) (client.CollectionDescription, error) {
	collectionKey := core.NewCollectionKey(name)
	var desc client.CollectionDescription
	schemaVersionIdBytes, err := p.txn.Systemstore().Get(p.ctx, collectionKey.ToDS())
	if err != nil {
		return desc, errors.Wrap("failed to get collection description", err)
	}

	schemaVersionId := string(schemaVersionIdBytes)
	schemaVersionKey := core.NewCollectionSchemaVersionKey(schemaVersionId)
	buf, err := p.txn.Systemstore().Get(p.ctx, schemaVersionKey.ToDS())
	if err != nil {
		return desc, err
	}

	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return desc, err
	}

	return desc, nil
}
