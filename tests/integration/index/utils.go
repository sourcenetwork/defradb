// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// createSchemaWithDocs returns UpdateSchema action and CreateDoc actions
// with the documents that match the schema.
// The schema is parsed to get the list of properties, and the docs
// are created with the same properties.
// This allows us to have only one large list of docs with predefined
// properties, and create schemas with different properties from it.
func createSchemaWithDocs(schema string) []any {
	userDocs := getUserDocs()
	resultActions := make([]any, 0, len(userDocs.docs)+1)
	resultActions = append(resultActions, testUtils.SchemaUpdate{Schema: schema})
	parser := schemaParser{}
	typeDefs := parser.parse(schema)
	for _, doc := range userDocs.docs {
		actions := makeCreateDocActions(doc, userDocs.colName, typeDefs)
		resultActions = append(resultActions, actions...)
	}
	return resultActions
}

func createDocJSON(doc map[string]any, typeDef *typeDefinition) (string, []propDefinition) {
	sb := strings.Builder{}
	relationProps := []propDefinition{}
	for _, prop := range typeDef.props {
		propName := prop.name
		format := `"%s": %v`
		if prop.isRelation {
			if !prop.isPrimary.Value() {
				if _, hasProp := doc[prop.name]; hasProp {
					relationProps = append(relationProps, prop)
				}
				continue
			} else {
				propName = propName + "_id"
			}
		}
		if _, isStr := doc[propName].(string); isStr {
			format = `"%s": "%v"`
		}
		if sb.Len() == 0 {
			sb.WriteString("{\n")
		} else {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(format, propName, doc[propName]))
	}
	sb.WriteString("\n}")
	return sb.String(), relationProps
}

func makeCreateDocActions(
	doc map[string]any,
	typeName string,
	types map[string]typeDefinition,
) []any {
	result := []any{}
	typeDef := types[typeName]

	docStr, relationProps := createDocJSON(doc, &typeDef)

	result = append(result, testUtils.CreateDoc{CollectionID: typeDef.index, Doc: docStr})
	if len(relationProps) > 0 {
		clientDoc, err := client.NewDocFromJSON([]byte(docStr))
		if err != nil {
			panic("Failed to create doc from JSON: " + err.Error())
		}
		docKey := clientDoc.Key().String()
		for _, relProp := range relationProps {
			actions := makeCreateDocActionForRelatedDocs(doc, typeName, &relProp, docKey, types)
			result = append(result, actions...)
		}
	}
	return result
}

func makeCreateDocActionForRelatedDocs(
	primaryDoc map[string]any,
	primaryTypeName string,
	relProp *propDefinition,
	primaryDocKey string,
	types map[string]typeDefinition,
) []any {
	result := []any{}
	relTypeDef := types[relProp.typeStr]
	primaryPropName := ""
	for _, relDocProp := range relTypeDef.props {
		if relDocProp.typeStr == primaryTypeName && relDocProp.isPrimary.Value() {
			primaryPropName = relDocProp.name + "_id"
			switch relVal := primaryDoc[relProp.name].(type) {
			case docsCollection:
				for _, relDoc := range relVal.docs {
					relDoc[primaryPropName] = primaryDocKey
					actions := makeCreateDocActions(relDoc, relTypeDef.name, types)
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocKey
				actions := makeCreateDocActions(relVal, relTypeDef.name, types)
				result = append(result, actions...)
			}
		}
	}
	return result
}

type propDefinition struct {
	name       string
	typeStr    string
	isArray    bool
	isRelation bool
	isPrimary  immutable.Option[bool]
}

type typeDefinition struct {
	name  string
	index int
	props map[string]propDefinition
}

type schemaParser struct {
	types             map[string]typeDefinition
	schemaLines       []string
	firstRelationType string
	currentTypeDef    typeDefinition
	relationTypesMap  map[string]map[string]string
}

func (p *schemaParser) parse(schema string) map[string]typeDefinition {
	p.types = make(map[string]typeDefinition)
	p.relationTypesMap = make(map[string]map[string]string)
	p.schemaLines = strings.Split(schema, "\n")
	p.findTypes()

	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.currentTypeDef = p.types[typeName]
			continue
		}
		if strings.HasPrefix(line, "}") {
			p.types[p.currentTypeDef.name] = p.currentTypeDef
			continue
		}
		pos := strings.Index(line, ":")
		if pos != -1 {
			p.defineProp(line, pos)
		}
	}
	p.resolvePrimaryRelations()
	return p.types
}

func (p *schemaParser) findTypes() {
	typeIndex := 0
	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.types[typeName] = typeDefinition{name: typeName, index: typeIndex, props: make(map[string]propDefinition)}
			typeIndex++
		}
	}
}

func (p *schemaParser) defineProp(line string, pos int) {
	prop := propDefinition{name: line[:pos]}
	prop.typeStr = strings.TrimSpace(line[pos+1:])
	typeEndPos := strings.Index(prop.typeStr, " ")
	if typeEndPos != -1 {
		prop.typeStr = prop.typeStr[:typeEndPos]
	}
	if prop.typeStr[0] == '[' {
		prop.isArray = true
		prop.typeStr = prop.typeStr[1 : len(prop.typeStr)-1]
	}
	if _, isRelation := p.types[prop.typeStr]; isRelation {
		prop.isRelation = true
		if prop.isArray {
			prop.isPrimary = immutable.Some(false)
		}
		relMap := p.relationTypesMap[prop.typeStr]
		if relMap == nil {
			relMap = make(map[string]string)
		}
		relMap[prop.name] = p.currentTypeDef.name
		p.relationTypesMap[prop.typeStr] = relMap
		if p.firstRelationType == "" {
			p.firstRelationType = p.currentTypeDef.name
		}
	}
	p.currentTypeDef.props[prop.name] = prop
}

func (p *schemaParser) resolvePrimaryRelations() {
	for typeName, relationProps := range p.relationTypesMap {
		typeDef := p.types[typeName]
		for _, prop := range typeDef.props {
			for relPropName, relPropType := range relationProps {
				if prop.typeStr == relPropType {
					relatedTypeDef := p.types[relPropType]
					relatedProp := relatedTypeDef.props[relPropName]
					if relatedProp.isPrimary.HasValue() {
						continue
					}
					prop.isPrimary = immutable.Some(typeName != p.firstRelationType)
					relatedProp.isPrimary = immutable.Some(typeName == p.firstRelationType)
					typeDef.props[prop.name] = prop
					relatedTypeDef.props[relPropName] = relatedProp
					p.types[relPropType] = relatedTypeDef
					delete(p.relationTypesMap, relPropType)
				}
			}
		}
		p.types[typeName] = typeDef
	}
}

func sendRequestAndExplain(
	reqBody string,
	results []map[string]any,
	asserter testUtils.ResultAsserter,
) []testUtils.Request {
	return []testUtils.Request{
		{
			Request: "query {" + reqBody + "}",
			Results: results,
		},
		{
			Request:  "query @explain(type: execute) {" + reqBody + "}",
			Asserter: asserter,
		},
	}
}
