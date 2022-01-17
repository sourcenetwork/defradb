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

// Consider moving this file into an `aggregate` sub-package to keep them organized,
// or moving all aggregates to within an do-all `aggregate` node when adding the next few
// aggregates in.

import (
	"reflect"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type countNode struct {
	p    *Planner
	plan planNode

	sourceProperty string
	virtualFieldId string
}

func (p *Planner) Count(c *parser.PropertyTransformation) (*countNode, error) {
	var sourceProperty string
	if len(c.Source) == 1 {
		sourceProperty = c.Source[0]
	} else {
		sourceProperty = ""
	}

	return &countNode{
		p:              p,
		sourceProperty: sourceProperty,
		virtualFieldId: c.Destination,
	}, nil
}

func (n *countNode) Init() error {
	return n.plan.Init()
}

func (n *countNode) Start() error           { return n.plan.Start() }
func (n *countNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *countNode) Close() error           { return n.plan.Close() }
func (n *countNode) Source() planNode       { return n.plan }

func (n *countNode) Values() map[string]interface{} {
	value := n.plan.Values()

	// Can just scan for now, can be replaced later by something fancier if needed
	var count int
	if property, hasProperty := value[n.sourceProperty]; hasProperty {
		v := reflect.ValueOf(property)
		switch v.Kind() {
		// v.Len will panic if v is not one of these types, we don't want it to panic
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
			count = v.Len()
		}
	}

	value[n.virtualFieldId] = count

	return value
}

func (n *countNode) Next() (bool, error) {
	return n.plan.Next()
}
