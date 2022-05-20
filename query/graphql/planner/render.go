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

	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

// the final field select and render
type renderNode struct {
	documentIterator

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
		children: []renderInfo{},
	}

	for _, selection := range childSelections {
		if slct, isSelect := selection.(*parser.Select); isSelect && slct.Hidden {
			continue
		}
		info.children = append(info.children, buildRenderInfo(selection))
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
		children:             []renderInfo{},
	}

	for _, selection := range childSelections {
		if slct, isSelect := selection.(*parser.Select); isSelect && slct.Hidden {
			continue
		}
		info.children = append(info.children, buildRenderInfo(selection))
	}

	return info
}

func (n *renderNode) Kind() string { return "renderNode" }
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

	if _, isHidden := doc[parserTypes.HiddenFieldName]; isHidden {
		return n.Next()
	}

	n.currentValue = map[string]interface{}{}
	for _, renderInfo := range n.renderInfo.children {
		renderInfo.render(doc, n.currentValue)
	}

	return true, nil
}
func (n *renderNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *renderNode) Close() error           { return n.plan.Close() }
func (n *renderNode) Source() planNode       { return n.plan }

// Renders the source document into the destination document using the given renderInfo.
// Function recursively handles any nested children defined in the render info.
func (r *renderInfo) render(src map[string]interface{}, destination map[string]interface{}) {
	var resultValue interface{}
	if val, ok := src[r.sourceFieldName]; ok {
		switch v := val.(type) {
		// If the current property is itself a map, we should render any properties of the child
		case map[string]interface{}:
			inner := map[string]interface{}{}

			if _, isHidden := v[parserTypes.HiddenFieldName]; isHidden {
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
				if _, isHidden := subv[parserTypes.HiddenFieldName]; isHidden {
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
