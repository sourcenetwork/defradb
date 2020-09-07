package graphql

import (
	"testing"

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
)

func newTestGenerator() *Generator {
	sm, _ := NewSchemaManager()
	return NewGenerator(sm)
}

func Test_Generator_NewGenerator_HasManager(t *testing.T) {
	sm, _ := NewSchemaManager()
	g := NewGenerator(sm)
	assert.Equal(t, sm, g.manager, "NewGenerator returned a different SchemaManager")
}

func Test_Generator_buildTypesFromAST_SingleScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.String,
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_MultiScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
			otherField: Boolean
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.String,
					},
					"otherField": &gql.Field{
						Name: "otherField",
						Type: gql.Boolean,
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_MultiObjectSingleScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
		}

		type OtherObject {
			otherField: Boolean
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.String,
					},
				},
			}),
			gql.NewObject(gql.ObjectConfig{
				Name: "OtherObject",
				Fields: gql.Fields{
					"otherField": &gql.Field{
						Name: "otherField",
						Type: gql.Boolean,
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_MultiObjectMultiScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
			secondary: Int
		}

		type OtherObject {
			otherField: Boolean
			tertiary: Float
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.String,
					},
					"secondary": &gql.Field{
						Name: "secondary",
						Type: gql.Int,
					},
				},
			}),
			gql.NewObject(gql.ObjectConfig{
				Name: "OtherObject",
				Fields: gql.Fields{
					"otherField": &gql.Field{
						Name: "otherField",
						Type: gql.Boolean,
					},
					"tertiary": &gql.Field{
						Name: "tertiary",
						Type: gql.Float,
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_MultiObjectSingleObjectField(t *testing.T) {
	myObj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
		},
	})
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
		}

		type OtherObject {
			otherField: MyObject
		}
		`,
		[]*gql.Object{
			myObj,
			gql.NewObject(gql.ObjectConfig{
				Name: "OtherObject",
				Fields: gql.Fields{
					"otherField": &gql.Field{
						Name: "otherField",
						Type: myObj,
					},
				},
			}),
		})
}

func runTestConfigForbuildTypesFromASTSuite(t *testing.T, schema string, typeDefs []*gql.Object) {
	g := newTestGenerator()

	// parse to AST
	source := source.NewSource(&source.Source{
		Body: []byte(schema),
	})
	doc, err := parser.Parse(parser.ParseParams{
		Source: source,
	})

	assert.NoError(t, err, "Failed to parse schema string")

	err = g.buildTypesFromAST(doc)
	assert.NoError(t, err, "Failed to build types from AST")

	for i, objDef := range typeDefs {
		objName := objDef.Name()
		myObject, exists := g.manager.schema.TypeMap()[objDef.Name()]
		assert.Truef(t, exists, "%s type doesn't exist in schema manager TypeMap", objName)
		assert.NoErrorf(t, myObject.Error(), "%s contains an internal error", objName)
		assert.Equal(t, myObject, g.typeDefs[i], "TypeMap object doesn't match typeDef object")

		myObjectActual := myObject.(*gql.Object)
		myObjectActual.Fields() // call Fields() to trigger the defineFields() function
		// to resolve the FieldsThunker

		assert.NoErrorf(t, myObjectActual.Error(), "%s contains an internal error from the defineFields() call", objName)

		assert.Equal(t, objDef.Name(), myObjectActual.Name(), "Mismatched object names from buildTypesFromAST")
		for _, fieldActual := range myObjectActual.Fields() {
			fieldExpected := objDef.Fields()[fieldActual.Name]

			assert.Equal(t, fieldExpected.Name, fieldActual.Name, "Mismatch object field names")
			assert.Equal(t, fieldExpected.Type.Name(), fieldActual.Type.Name(), "Mismatch object field types")
		}
	}
}
