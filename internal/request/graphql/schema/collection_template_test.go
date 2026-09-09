// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	"github.com/sourcenetwork/defradb/client"
)

func TestRenderCollectionSchemaSDL(t *testing.T) {
	collections := []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
			{Name: "age", Kind: client.FieldKind_NILLABLE_INT},
			{Name: "verified", Kind: client.FieldKind_NILLABLE_BOOL},
		},
	}}

	dynamic, err := renderCollectionSchemaSDL(collections)
	require.NoError(t, err)
	require.Contains(t, dynamic, "truncate_User")
	require.Contains(t, dynamic, "upsert_User")
	require.Contains(t, dynamic, "age: IntOperatorBlock")
	require.Contains(t, dynamic, "field: UserNumericFieldsArg!")

	document, report := astparser.ParseGraphqlDocumentString(
		strings.TrimSpace(defaultSchemaSDL) + "\n\n" + dynamic,
	)
	require.False(t, report.HasErrors(), report.Error())
	astnormalization.NormalizeDefinition(&document, &report)
	require.False(t, report.HasErrors(), report.Error())

	validationReport := operationreport.Report{}
	state := astvalidation.DefaultDefinitionValidator().Validate(&document, &validationReport)
	require.Equal(t, astvalidation.Valid, state, validationReport.Error())
}

// TestCollectionTemplateSimpleSurfaceParity checks the public template surface
// using graphql-go-tools introspection. Descriptions and ordering are ignored.
func TestCollectionTemplateSimpleSurfaceParity(t *testing.T) {
	assertCollectionTemplateSurfaceParity(t, `
		type User {
			name: String
			age: Int
			verified: Boolean
			points: Int @crdt(type: pncounter)
		}`)
}

func TestCollectionTemplateRelationSurfaceParity(t *testing.T) {
	for _, test := range []struct {
		name string
		sdl  string
	}{
		{
			name: "one",
			sdl: `
				type Book {
					name: String
					rating: Float
					author: Author @primary
				}
				type Author {
					name: String
					age: Int
					verified: Boolean
					published: Book
				}`,
		},
		{
			name: "many",
			sdl: `
				type Book {
					name: String
					rating: Float
					author: Author
				}
				type Author {
					name: String
					age: Int
					verified: Boolean
					published: [Book]
				}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertCollectionTemplateSurfaceParity(t, test.sdl)
		})
	}
}

func TestCollectionTemplateScalarSurfaceParity(t *testing.T) {
	assertCollectionTemplateSurfaceParity(t, `
		type Scalars {
			boolRequired: Boolean!
			boolOptional: Boolean
			intRequired: Int!
			intOptional: Int
			float32Required: Float32!
			float32Optional: Float32
			float64Required: Float64!
			float64Optional: Float64
			stringRequired: String!
			stringOptional: String
			dateRequired: DateTime!
			dateOptional: DateTime
			blobRequired: Blob!
			blobOptional: Blob
			jsonRequired: JSON!
			jsonOptional: JSON
			boolItemsRequired: [Boolean!]
			boolItemsOptional: [Boolean]
			intItemsRequired: [Int!]
			intItemsOptional: [Int]
			float32ItemsRequired: [Float32!]
			float32ItemsOptional: [Float32]
			float64ItemsRequired: [Float64!]
			float64ItemsOptional: [Float64]
			stringItemsRequired: [String!]
			stringItemsOptional: [String]
			dateItemsRequired: [DateTime!]
			dateItemsOptional: [DateTime]
		}`)
}

func TestCollectionTemplateEmbeddedSurfaceParity(t *testing.T) {
	assertCollectionTemplateSurfaceParity(t, `
		interface Embedded {
			name: String
		}
	`)
}

func TestCollectionTemplateReadOnlyRootParity(t *testing.T) {
	collections := []client.CollectionVersion{{
		Name:  "UserView",
		Query: immutable.Some(client.QuerySource{}),
		Fields: []client.CollectionFieldDescription{
			{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
		},
	}}
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)
	err = manager.Generate(context.Background(), collections)
	require.NoError(t, err)
	templateSchema := introspectDocument(t, manager.Definition())
	require.Contains(t, typeMembers(templateSchema.TypeByName("Query")), "UserView")
	require.Contains(t, typeMembers(templateSchema.TypeByName("Subscription")), "UserView")
	require.NotContains(t, typeMembers(templateSchema.TypeByName("Mutation")), "add_UserView")
	require.NotContains(t, typeMembers(templateSchema.TypeByName("Mutation")), "update_UserView")
	require.NotContains(t, typeMembers(templateSchema.TypeByName("Mutation")), "delete_UserView")
}

func assertCollectionTemplateSurfaceParity(t *testing.T, sdl string) {
	t.Helper()
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)
	parsed, err := manager.ParseSDL(sdl)
	require.NoError(t, err)
	collections := make([]client.CollectionVersion, len(parsed))
	for index := range parsed {
		collections[index] = parsed[index].Definition
	}
	dynamic, err := renderCollectionSchemaSDL(collections)
	require.NoError(t, err)

	templateSchema := introspectSDL(t, strings.TrimSpace(defaultSchemaSDL)+"\n\n"+dynamic)
	for _, collection := range collections {
		foundCollectionType := false
		for _, gqlType := range templateSchema.Types {
			if !strings.HasPrefix(gqlType.Name, collection.Name) {
				continue
			}
			foundCollectionType = true
			require.NotEmpty(t, typeSurface(templateSchema.TypeByName(gqlType.Name)), gqlType.Name)
		}
		require.True(t, foundCollectionType, collection.Name)
	}
	model, err := newCollectionTemplateModel(collections, false)
	require.NoError(t, err)
	for _, filter := range model.ElementFilters {
		require.NotEmpty(t, typeSurface(templateSchema.TypeByName(filter.Name)), filter.Name)
	}
	for _, rootName := range []string{"Query", "Mutation", "Subscription"} {
		require.NotEmpty(t, typeSurface(templateSchema.TypeByName(rootName)), rootName)
	}
}

func introspectSDL(t *testing.T, sdl string) introspection.Schema {
	t.Helper()
	document, report := astparser.ParseGraphqlDocumentString(sdl)
	require.False(t, report.HasErrors(), report.Error())
	astnormalization.NormalizeDefinition(&document, &report)
	require.False(t, report.HasErrors(), report.Error())
	return introspectDocument(t, &document)
}

func introspectDocument(t *testing.T, document *wgast.Document) introspection.Schema {
	t.Helper()
	var report operationreport.Report
	var data introspection.Data
	introspection.NewGenerator().Generate(document, &report, &data)
	require.False(t, report.HasErrors(), report.Error())
	return data.Schema
}

func typeMembers(gqlType *introspection.FullType) []string {
	if gqlType == nil {
		return nil
	}
	members := make([]string, 0, len(gqlType.Fields)+len(gqlType.InputFields)+len(gqlType.EnumValues))
	for _, field := range gqlType.Fields {
		members = append(members, field.Name)
	}
	for _, field := range gqlType.InputFields {
		members = append(members, field.Name)
	}
	for _, value := range gqlType.EnumValues {
		members = append(members, value.Name)
	}
	sort.Strings(members)
	return members
}

func typeSurface(gqlType *introspection.FullType) []string {
	if gqlType == nil {
		return nil
	}
	members := make([]string, 0, len(gqlType.Fields)+len(gqlType.InputFields)+len(gqlType.EnumValues))
	for _, field := range gqlType.Fields {
		args := make([]string, len(field.Args))
		for index, arg := range field.Args {
			args[index] = arg.Name + ":" + typeRefString(arg.Type)
		}
		sort.Strings(args)
		members = append(members, field.Name+"("+strings.Join(args, ",")+"):"+typeRefString(field.Type))
	}
	for _, field := range gqlType.InputFields {
		members = append(members, field.Name+":"+typeRefString(field.Type))
	}
	for _, value := range gqlType.EnumValues {
		members = append(members, value.Name)
	}
	sort.Strings(members)
	return members
}

func typeRefString(ref introspection.TypeRef) string {
	if ref.Name != nil {
		return *ref.Name
	}
	if ref.OfType == nil {
		return ""
	}
	typeName := typeRefString(*ref.OfType)
	switch ref.Kind.String() {
	case "NON_NULL":
		return typeName + "!"
	case "LIST":
		return "[" + typeName + "]"
	default:
		return typeName
	}
}

func TestCollectionTemplateModelRelationFields(t *testing.T) {
	collections := []client.CollectionVersion{
		{Name: "Author"},
		{
			Name: "Book",
			Fields: []client.CollectionFieldDescription{{
				Name:      "author",
				Kind:      &client.NamedKind{Name: "Author"},
				IsPrimary: true,
			}},
		},
	}

	dynamic, err := renderCollectionSchemaSDL(collections)
	require.NoError(t, err)
	require.Contains(t, dynamic, "_authorID: ID")
	require.Contains(t, dynamic, "author(filter: AuthorFilterArg): Author")
	require.Contains(t, dynamic, "author: AuthorFilterArg")
	require.Contains(t, dynamic, "author: AuthorOrderArg")
}

func TestCollectionTemplateModelSelfKinds(t *testing.T) {
	setID := "books-and-authors"
	collections := []client.CollectionVersion{
		{
			Name:          "Author",
			CollectionSet: immutable.Some(client.CollectionSetDescription{CollectionSetID: setID, RelativeID: 0}),
			Fields: []client.CollectionFieldDescription{{
				Name: "books", Kind: client.NewSelfKind("1", true),
			}},
		},
		{
			Name:          "Book",
			CollectionSet: immutable.Some(client.CollectionSetDescription{CollectionSetID: setID, RelativeID: 1}),
			Fields: []client.CollectionFieldDescription{{
				Name: "author", Kind: client.NewSelfKind("0", false), IsPrimary: true,
			}},
		},
	}

	dynamic, err := renderCollectionSchemaSDL(collections)
	require.NoError(t, err)
	require.Contains(t, dynamic, "books(docID: [ID!], filter: BookFilterArg")
	require.Contains(t, dynamic, "author(filter: AuthorFilterArg): Author")
	require.Contains(t, dynamic, "_authorID: ID")
}

func TestCollectionTemplateEncryptedSurfaceParity(t *testing.T) {
	collections := []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: "email", Kind: client.FieldKind_NILLABLE_STRING},
			{Name: "age", Kind: client.FieldKind_NILLABLE_INT},
		},
		EncryptedIndexes: []client.EncryptedIndexDescription{
			{FieldName: "email", Type: client.EncryptedIndexTypeEquality},
		},
	}}
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)
	err = manager.Generate(context.Background(), collections)
	require.NoError(t, err)
	templateSchema := introspectDocument(t, manager.Definition())
	require.Equal(t, []string{"email:StringEncryptedFilterArg"},
		typeSurface(templateSchema.TypeByName("UserEncryptedFilterArg")))
	require.Contains(t, typeSurface(templateSchema.TypeByName("StringEncryptedFilterArg")), "_eq:String")
	require.Contains(t, typeMembers(templateSchema.TypeByName("Query")), "encrypted_User")
}
