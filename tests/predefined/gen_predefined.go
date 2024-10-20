// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package predefined

import (
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/gen"
)

// CreateFromSDL generates documents for GraphQL SDL from a predefined list
// of docs that might include nested docs.
// The SDL is parsed to get the list of fields, and the docs
// are created with the fields parsed from the SDL.
// This allows us to have only one large list of docs with predefined
// fields, and create SDLs with different fields from it.
func CreateFromSDL(gqlSDL string, docsList DocsList) ([]gen.GeneratedDoc, error) {
	resultDocs := make([]gen.GeneratedDoc, 0, len(docsList.Docs))
	typeDefsByName, err := gen.ParseSDL(gqlSDL)
	if err != nil {
		return nil, err
	}

	defs := make([]client.CollectionDefinition, 0, len(typeDefsByName))
	for _, def := range typeDefsByName {
		defs = append(defs, def)
	}

	generator := docGenerator{
		types:           typeDefsByName,
		definitionCache: client.NewDefinitionCache(defs),
	}

	for _, doc := range docsList.Docs {
		docs, err := generator.generateRelatedDocs(doc, docsList.ColName)
		if err != nil {
			return nil, err
		}
		resultDocs = append(resultDocs, docs...)
	}
	return resultDocs, nil
}

// Create generates documents from a predefined list
// of docs that might include nested docs.
//
// For example it can be used to generate docs from this list:
//
//		gen.DocsList{
//			ColName: "User",
//			Docs: []map[string]any{
//				{
//					"name":     "Shahzad",
//					"age":      20,
//					"devices": []map[string]any{
//						{
//							"model": "iPhone Xs",
//						},
//					},
//				},
//			},
//	 ...
//
// It will generator documents for `User` collection replicating the given structure, i.e.
// creating devices as related secondary documents.
func Create(defs []client.CollectionDefinition, docsList DocsList) ([]gen.GeneratedDoc, error) {
	resultDocs := make([]gen.GeneratedDoc, 0, len(docsList.Docs))
	typeDefs := make(map[string]client.CollectionDefinition)
	for _, col := range defs {
		typeDefs[col.Description.Name.Value()] = col
	}

	generator := docGenerator{
		types:           typeDefs,
		definitionCache: client.NewDefinitionCache(defs),
	}

	for _, doc := range docsList.Docs {
		docs, err := generator.generateRelatedDocs(doc, docsList.ColName)
		if err != nil {
			return nil, err
		}
		resultDocs = append(resultDocs, docs...)
	}
	return resultDocs, nil
}

type docGenerator struct {
	types           map[string]client.CollectionDefinition
	definitionCache client.DefinitionCache
}

// toRequestedDoc removes the fields that are not in the schema of the collection.
//
// This is typically called on user/test provided seed documents to remove any non-existent
// fields before generating documents from them.
// It doesn't not modify the original doc.
func toRequestedDoc(doc map[string]any, typeDef *client.CollectionDefinition) map[string]any {
	result := make(map[string]any)
	for _, field := range typeDef.GetFields() {
		if field.IsRelation() || field.Name == request.DocIDFieldName {
			continue
		}
		result[field.Name] = doc[field.Name]
	}
	for name, val := range doc {
		if strings.HasSuffix(name, request.RelatedObjectID) {
			result[name] = val
		}
	}
	return result
}

// generatePrimary generates primary docs for the given secondary doc and adds foreign docID
// to the secondary doc to reference the primary docs.
func (this *docGenerator) generatePrimary(
	secDocMap map[string]any,
	secType *client.CollectionDefinition,
) (map[string]any, []gen.GeneratedDoc, error) {
	result := []gen.GeneratedDoc{}
	requestedSecondary := toRequestedDoc(secDocMap, secType)
	for _, secDocField := range secType.GetFields() {
		if secDocField.IsRelation() && secDocField.IsPrimaryRelation {
			if secDocMapField, hasField := secDocMap[secDocField.Name]; hasField {
				primaryDef, _ := client.GetDefinition(this.definitionCache, *secType, secDocField.Kind)
				primType := this.types[primaryDef.GetName()]

				primDocMap, subResult, err := this.generatePrimary(
					secDocMap[secDocField.Name].(map[string]any), &primType)
				if err != nil {
					return nil, nil, NewErrFailedToGenerateDoc(err)
				}
				primDoc, err := client.NewDocFromMap(primDocMap, primType)
				if err != nil {
					return nil, nil, NewErrFailedToGenerateDoc(err)
				}
				docID := primDoc.ID().String()
				requestedSecondary[secDocField.Name+request.RelatedObjectID] = docID
				subResult = append(subResult, gen.GeneratedDoc{Col: &primType, Doc: primDoc})
				result = append(result, subResult...)

				secondaryDocs, err := this.generateSecondaryDocs(
					secDocMapField.(map[string]any), docID, &primType, secType.Description.Name.Value())
				if err != nil {
					return nil, nil, err
				}
				result = append(result, secondaryDocs...)
			}
		}
	}
	return requestedSecondary, result, nil
}

// generateRelatedDocs generates related docs (primary and secondary) for the given doc and
// adds foreign docID to the given doc to reference the primary docs.
func (this *docGenerator) generateRelatedDocs(docMap map[string]any, typeName string) ([]gen.GeneratedDoc, error) {
	typeDef := this.types[typeName]

	// create first primary docs and link them to the given doc so that we can define
	// docID for the complete document.
	requested, result, err := this.generatePrimary(docMap, &typeDef)
	if err != nil {
		return nil, err
	}
	doc, err := client.NewDocFromMap(requested, typeDef)
	if err != nil {
		return nil, NewErrFailedToGenerateDoc(err)
	}

	result = append(result, gen.GeneratedDoc{Col: &typeDef, Doc: doc})

	secondaryDocs, err := this.generateSecondaryDocs(docMap, doc.ID().String(), &typeDef, "")
	if err != nil {
		return nil, err
	}
	result = append(result, secondaryDocs...)

	return result, nil
}

func (this *docGenerator) generateSecondaryDocs(
	primaryDocMap map[string]any,
	docID string,
	primaryType *client.CollectionDefinition,
	parentTypeName string,
) ([]gen.GeneratedDoc, error) {
	result := []gen.GeneratedDoc{}
	for _, field := range primaryType.GetFields() {
		if field.IsRelation() && !field.IsPrimaryRelation {
			if _, hasProp := primaryDocMap[field.Name]; hasProp {
				otherDef, _ := client.GetDefinition(this.definitionCache, *primaryType, field.Kind)
				if parentTypeName == "" || parentTypeName != otherDef.GetName() {
					docs, err := this.generateSecondaryDocsForField(
						primaryDocMap, *primaryType, &field, docID)
					if err != nil {
						return nil, err
					}

					result = append(result, docs...)
				}
			}
		}
	}
	return result, nil
}

// generateSecondaryDocsForField generates secondary docs for the given field of a primary doc.
func (this *docGenerator) generateSecondaryDocsForField(
	primaryDoc map[string]any,
	primaryType client.CollectionDefinition,
	relField *client.FieldDefinition,
	primaryDocID string,
) ([]gen.GeneratedDoc, error) {
	result := []gen.GeneratedDoc{}

	relTypeDef, _ := client.GetDefinition(this.definitionCache, primaryType, relField.Kind)

	primaryPropName := ""
	for _, relDocField := range relTypeDef.GetFields() {
		relDocDef, _ := client.GetDefinition(this.definitionCache, relTypeDef, relDocField.Kind)

		if relDocDef.GetName() == primaryType.GetName() && relDocField.IsPrimaryRelation {
			primaryPropName = relDocField.Name + request.RelatedObjectID
			switch relVal := primaryDoc[relField.Name].(type) {
			case []map[string]any:
				for _, relDoc := range relVal {
					relDoc[primaryPropName] = primaryDocID
					actions, err := this.generateRelatedDocs(relDoc, relTypeDef.Description.Name.Value())
					if err != nil {
						return nil, err
					}
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocID
				actions, err := this.generateRelatedDocs(relVal, relTypeDef.Description.Name.Value())
				if err != nil {
					return nil, err
				}
				result = append(result, actions...)
			}
		}
	}
	return result, nil
}
