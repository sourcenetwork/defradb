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
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// the final field select and render
type renderNode struct {
	p    *Planner
	plan planNode

	renderInfo topLevelRenderInfo
}

type topLevelRenderInfo struct {
	children []renderInfo
}

type renderInfo struct {
	sourceFieldName      string
	destinationFieldName string
	children             []renderInfo
}

func (p *Planner) render(parsed *parser.Select) *renderNode {
	return &renderNode{
		p:          p,
		renderInfo: buildTopLevelRenderInfo(parsed),
	}
}

func buildTopLevelRenderInfo(parsed parser.Selection) topLevelRenderInfo {
	childSelections := parsed.GetSelections()

	info := topLevelRenderInfo{
		children: make([]renderInfo, len(childSelections)),
	}

	for i, selection := range childSelections {
		info.children[i] = buildRenderInfo(selection)
	}

	return info
}

func buildRenderInfo(parsed parser.Selection) renderInfo {
	childSelections := parsed.GetSelections()
	sourceFieldName := parsed.GetName()
	alias := parsed.GetAlias()

	var destinationFieldName string
	if alias == "" {
		destinationFieldName = sourceFieldName
	} else {
		destinationFieldName = alias
	}

	info := renderInfo{
		sourceFieldName:      sourceFieldName,
		destinationFieldName: destinationFieldName,
		children:             make([]renderInfo, len(childSelections)),
	}

	for i, selection := range childSelections {
		info.children[i] = buildRenderInfo(selection)
	}

	return info
}

func (n *renderNode) Init() error  { return n.plan.Init() }
func (n *renderNode) Start() error { return n.plan.Start() }
func (n *renderNode) Next() (bool, error) {
	hasNext, err := n.plan.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	doc := n.plan.Value()
	if doc == nil {
		return n.Next()
	}

	if _, isHidden := doc[parser.HiddenFieldName]; isHidden {
		return n.Next()
	}
	return hasNext, err
}
func (n *renderNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *renderNode) Close() error           { return n.plan.Close() }
func (n *renderNode) Source() planNode       { return n.plan }

// we only need to implement the Values() func of the planNode
// interface since the embedded baseNode implements the rest
func (r *renderNode) Value() map[string]interface{} {
	doc := r.plan.Value()
	if doc == nil {
		return doc
	}

	result := map[string]interface{}{}
	for _, renderInfo := range r.renderInfo.children {
		renderInfo.render(doc, result)
	}
	return result
}

// Renders the source document into the destination document using the given renderInfo.
// Function recursively handles any nested children defined in the render info.
func (r *renderInfo) render(src map[string]interface{}, destination map[string]interface{}) {
	var resultValue interface{}
	if val, ok := src[r.sourceFieldName]; ok {
		switch v := val.(type) {
		// If the current property is itself a map, we should render any properties of the child
		case map[string]interface{}:
			inner := map[string]interface{}{}

			if _, isHidden := v[parser.HiddenFieldName]; isHidden {
				return
			}

			for _, child := range r.children {
				child.render(v, inner)
			}
			resultValue = inner
		// If the current property is an array of maps, we should render each child map
		case []map[string]interface{}:
			subdocs := make([]map[string]interface{}, 0)
			for _, subv := range v {
				if _, isHidden := subv[parser.HiddenFieldName]; isHidden {
					continue
				}

				inner := map[string]interface{}{}
				for _, child := range r.children {
					child.render(subv, inner)
				}
				subdocs = append(subdocs, inner)
			}
			resultValue = subdocs
		default:
			resultValue = v
		}
	} else {
		resultValue = nil
	}

	destination[r.destinationFieldName] = resultValue
}
