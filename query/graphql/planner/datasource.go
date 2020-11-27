package planner

import (
	"encoding/json"
	"errors"

	"github.com/sourcenetwork/defradb/core"
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
func (p *Planner) getSource(collection string) (planSource, error) {
	// for now, we only handle simple collection scannodes
	return p.getCollectionScanPlan(collection)
}

// @todo: Add field selection
func (p *Planner) getCollectionScanPlan(collection string) (planSource, error) {
	if collection == "" {
		return nil, errors.New("collection name cannot be empty")
	}
	colDesc, err := p.getCollectionDesc(collection)
	if err != nil {
		return nil, err
	}

	scan := p.Scan()
	scan.initCollection(colDesc)

	return planSource{
		plan: scan,
		info: sourceInfo{
			collectionDescription: colDesc,
		},
	}
}

func (p *Planner) getCollectionDesc(name string) (base.CollectionDescription, error) {
	key := core.MakeCollectionSystemKey(name)
	buf, err := p.txn.Systemstore().Get(key.ToDS())
	if err != nil {
		return nil, err
	}

	var desc base.CollectionDescription
	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return nil, err
	}

	return desc, nil
}
