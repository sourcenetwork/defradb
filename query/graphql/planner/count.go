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

// Consider moving this file into an `aggregate` sub-package to keep them organized,
// or moving all aggregates to within an do-all `aggregate` node when adding the next few
// aggregates in.

import (
	"reflect"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type countNode struct {
	documentIterator

	p    *Planner
	plan planNode

	sourceProperty string
	virtualFieldId string

	filter *parser.Filter
}

func (p *Planner) Count(field *parser.Select, host *parser.Select) (*countNode, error) {
	source, err := field.GetAggregateSource(host)
	if err != nil {
		return nil, err
	}

	return &countNode{
		p:              p,
		sourceProperty: source.HostProperty,
		virtualFieldId: field.Name,
		filter:         field.Filter,
	}, nil
}

func (n *countNode) Init() error {
	return n.plan.Init()
}

func (n *countNode) Start() error           { return n.plan.Start() }
func (n *countNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *countNode) Close() error           { return n.plan.Close() }
func (n *countNode) Source() planNode       { return n.plan }

func (n *countNode) Next() (bool, error) {
	hasValue, err := n.plan.Next()
	if err != nil || !hasValue {
		return hasValue, err
	}

	n.currentValue = n.plan.Value()

	// Can just scan for now, can be replaced later by something fancier if needed
	var count int
	if property, hasProperty := n.currentValue[n.sourceProperty]; hasProperty {
		v := reflect.ValueOf(property)
		switch v.Kind() {
		// v.Len will panic if v is not one of these types, we don't want it to panic
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
			count = v.Len()
			// For now, we only support count filters internally to support averages
			// so this is fine here now, but may need to be moved later once external
			// count filter support is added.
			if count > 0 && n.filter != nil {
				docArray, isDocArray := property.([]core.Doc)
				if isDocArray {
					count = 0
					for _, doc := range docArray {
						passed, err := parser.RunFilter(doc, n.filter, n.p.evalCtx)
						if err != nil {
							return false, err
						}
						if passed {
							count += 1
						}
					}
				}
			}
		}
	}

	n.currentValue[n.virtualFieldId] = count
	return true, nil
}

func (n *countNode) SetPlan(p planNode) { n.plan = p }
