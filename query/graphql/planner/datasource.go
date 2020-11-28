package planner

import (
	"encoding/json"
	"errors"

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
		return planSource{}, errors.New("collection name cannot be empty")
	}
	colDesc, err := p.getCollectionDesc(collection)
	if err != nil {
		return planSource{}, err
	}

	scan := p.Scan()
	scan.initCollection(colDesc)

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
	buf, err := p.txn.Systemstore().Get(key.ToDS())
	if err != nil {
		return desc, err
	}

	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return desc, err
	}

	return desc, nil
}
