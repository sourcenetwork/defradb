package schema

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
	gql "github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
)

func newTestGenerator() *Generator {
	sm, _ := NewSchemaManager()
	return sm.NewGenerator()
}

func Test_Generator_NewGenerator_HasManager(t *testing.T) {
	sm, _ := NewSchemaManager()
	g := sm.NewGenerator()
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
					"_key": &gql.Field{
						Name: "_key",
						Type: gql.ID,
					},
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.String,
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_SingleNonNullScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String!
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.NewNonNull(gql.String),
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_SingleListScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: [String]
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.NewList(gql.String),
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_SingleListNonNullScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: [String!]
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.NewList(gql.NewNonNull(gql.String)),
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_SingleNonNullListScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: [String]!
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.NewNonNull(gql.NewList(gql.String)),
					},
				},
			}),
		})
}

func Test_Generator_buildTypesFromAST_SingleNonNullListNonNullScalarField(t *testing.T) {
	runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: [String!]!
		}
		`,
		[]*gql.Object{
			gql.NewObject(gql.ObjectConfig{
				Name: "MyObject",
				Fields: gql.Fields{
					"myField": &gql.Field{
						Name: "myField",
						Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(gql.String))),
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
			boolField: Boolean
			intField: Int
			floatField: Float
			dateTimeField: DateTime
			idField: ID
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
					"boolField": &gql.Field{
						Name: "boolField",
						Type: gql.Boolean,
					},
					"intField": &gql.Field{
						Name: "intField",
						Type: gql.Int,
					},
					"floatField": &gql.Field{
						Name: "floatField",
						Type: gql.Float,
					},
					"dateTimeField": &gql.Field{
						Name: "dateTimeField",
						Type: gql.DateTime,
					},
					"idField": &gql.Field{
						Name: "idField",
						Type: gql.ID,
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

func Test_Generator_buildTypesFromAST_MissingObject(t *testing.T) {
	myObj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
		},
	})
	err := runTestConfigForbuildTypesFromASTSuite(t,
		`
		type MyObject {
			myField: String
		}

		type OtherObject {
			otherField: UndefinedObject
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

	// make sure we get back the *correct* error.
	if err != nil && !strings.Contains(err.Error(), "field type must be Output Type but got: <nil>") {
		t.Error("buildTypesFromAST didn't fail on UndefinedObject:", err)
	}
}

func runTestConfigForbuildTypesFromASTSuite(t *testing.T, schema string, typeDefs []*gql.Object) error {
	g := newTestGenerator()

	// // parse to AST
	// source := source.NewSource(&source.Source{
	// 	Body: []byte(schema),
	// })
	// doc, err := parser.Parse(parser.ParseParams{
	// 	Source: source,
	// })
	// if err != nil {
	// 	return errors.Wrap(err, "Failed to parse schema string")
	// }
	// // assert.NoError(t, err, "Failed to parse schema string")

	// err = g.buildTypesFromAST(doc)
	_, err := g.FromSDL(schema)
	if err != nil {
		return errors.Wrap(err, "Failed to build types from AST")
	}
	// assert.NoError(t, err, "Failed to build types from AST")

	for i, objDef := range typeDefs {
		objName := objDef.Name()
		myObject, exists := g.manager.schema.TypeMap()[objDef.Name()]
		if !exists {
			return errors.New(fmt.Sprintf("%s type doesn't exist in the schema manager TypeMap", objName))
		}
		if myObject.Error() != nil {
			return errors.Wrapf(myObject.Error(), "%s contains an internal error", objName)
		}
		if !reflect.DeepEqual(myObject, g.typeDefs[i]) {
			// add the assert here for its object diff output
			assert.Equal(t, myObject, g.typeDefs[i], "TypeMap object doesn't match typeDef object")
			return errors.New("TypeMap object doesn't match typeDef object")
		}

		myObjectActual := myObject.(*gql.Object)
		spew.Dump(myObjectActual.Fields())
		myObjectActual.Fields() // call Fields() to trigger the defineFields() function
		// to resolve the FieldsThunker

		if myObject.Error() != nil {
			return errors.Wrap(myObject.Error(), fmt.Sprintf("%s contains an internal error from the Fields() > definFields() call", objName))
		}
		// assert.NoErrorf(t, myObjectActual.Error(), "%s contains an internal error from the defineFields() call", objName)

		assert.Equal(t, objDef.Name(), myObjectActual.Name(), "Mismatched object names from buildTypesFromAST")
		fmt.Println("expected vs actual objects:")
		fmt.Println(objDef.Fields())
		fmt.Println(myObjectActual.Fields())
		for _, fieldActual := range myObjectActual.Fields() {
			fieldExpected, ok := objDef.Fields()[fieldActual.Name]
			assert.True(t, ok, "Failed to find expected field for matching actual field")

			assert.Equal(t, fieldExpected.Name, fieldActual.Name, "Mismatch object field names")
			assert.Equal(t, fieldExpected.Type.Name(), fieldActual.Type.Name(), "Mismatch object field types")
		}
	}

	return nil
}

func Test_Generator_genType_Filter_SingleScalar(t *testing.T) {
	g := newTestGenerator()
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func Test_Generator_genType_Filter_MultiScalar(t *testing.T) {
	g := newTestGenerator()
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func Test_Generator_genType_Filter_SingleObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"other": &gql.Field{
				Name: "other",
				Type: other,
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func Test_Generator_genType_Filter_CompositeObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
			"other": &gql.Field{
				Name: "other",
				Type: other,
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func Test_Generator_genType_Filter_MultiObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	another := gql.NewObject(gql.ObjectConfig{
		Name: "AnotherObject",
		Fields: gql.Fields{
			"anotherField": &gql.Field{
				Name: "anotherField",
				Type: gql.Int,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: another,
			},
			"other": &gql.Field{
				Name: "other",
				Type: other,
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func Test_Generator_genType_Filter_SingleListObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"other": &gql.Field{
				Name: "other",
				Type: gql.NewList(other),
			},
		},
	})

	runTestConfigForGenTypeFilterSuite(t, g, obj)
}

func runTestConfigForGenTypeFilterSuite(t *testing.T, g *Generator, obj *gql.Object) {
	filterInput := g.genTypeFilterArgInput(obj)
	// generate the any sub object base arg input
	for _, field := range obj.Fields() {
		if !gql.IsLeafType(field.Type) {
			unwrappedFieldType := unwrapType(field.Type)
			base := g.genTypeFilterBaseArgInput(unwrappedFieldType.(*gql.Object))
			err := g.manager.schema.AppendType(base)
			assert.NoError(t, err, "Failed to generate sub object base arg input types")
		}
	}
	err := g.manager.schema.AppendType(filterInput)
	assert.NoError(t, err, "Failed to append type to TypeMap")

	assert.Equal(t, genTypeName(obj, "FilterArg"), filterInput.Name(), "Generated FilterInput type has incorrect name")
	assert.NoError(t, filterInput.Error(), "FilterInput type has an internal error")

	fields := filterInput.Fields()
	// conditional fields
	assert.Equal(t, filterInput, fields["_not"].Type, "_not fields of FilterInput type don't match")
	assert.Equal(t, gql.NewList(filterInput), fields["_and"].Type, "_and fields of FilterInput type don't match")
	assert.Equal(t, gql.NewList(filterInput), fields["_or"].Type, "_or fields of FilterInput type don't match")

	// object fields
	for _, field := range obj.Fields() {
		filterField, exists := fields[field.Name]
		assert.True(t, exists, "Missing field on FilterInput: %s", field.Name)
		assert.Equal(t, field.Name, filterField.Name(), "%s field name doesn't match", field.Name)
		if gql.IsLeafType(field.Type) { // leaf types (enums + scalars)
			block := g.manager.schema.TypeMap()[genTypeName(field.Type, "OperatorBlock")]
			assert.Equal(t, block, filterField.Type, "%s field doesn't match expected", field.Name)
		} else { // objects
			// DO
			unwrappedFieldType := unwrapType(field.Type)
			block := g.manager.schema.TypeMap()[genTypeName(unwrappedFieldType, "FilterBaseArg")]
			// unwrappedBlock := unwrapType(block)
			assert.Equal(t, block, filterField.Type, "%s field doesn't match expected", field.Name)
		}
	}
}

func Test_Generator_genType_FieldEnum_SingleScalar(t *testing.T) {
	g := newTestGenerator()
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
		},
	})

	runTestConfigForGenTypeFieldsEnum(t, g, obj)
}

func Test_Generator_genType_FieldEnum_MultiScalar(t *testing.T) {
	g := newTestGenerator()
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"myField": &gql.Field{
				Name: "myField",
				Type: gql.String,
			},
			"otherField": &gql.Field{
				Name: "myField",
				Type: gql.Int,
			},
		},
	})

	runTestConfigForGenTypeFieldsEnum(t, g, obj)
}

func Test_Generator_genType_FieldEnum_SingleObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"other": &gql.Field{
				Name: "other",
				Type: other,
			},
		},
	})

	runTestConfigForGenTypeFieldsEnum(t, g, obj)
}

func Test_Generator_genType_FieldEnum_MultiObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	another := gql.NewObject(gql.ObjectConfig{
		Name: "AnotherObject",
		Fields: gql.Fields{
			"anotherField": &gql.Field{
				Name: "anotherField",
				Type: gql.Int,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"other": &gql.Field{
				Name: "other",
				Type: other,
			},
			"another": &gql.Field{
				Name: "another",
				Type: another,
			},
		},
	})

	runTestConfigForGenTypeFieldsEnum(t, g, obj)
}

func Test_Generator_genType_FieldEnum_SingleListObject(t *testing.T) {
	g := newTestGenerator()
	other := gql.NewObject(gql.ObjectConfig{
		Name: "OtherObject",
		Fields: gql.Fields{
			"otherField": &gql.Field{
				Name: "otherField",
				Type: gql.Boolean,
			},
		},
	})
	obj := gql.NewObject(gql.ObjectConfig{
		Name: "MyObject",
		Fields: gql.Fields{
			"other": &gql.Field{
				Name: "other",
				Type: gql.NewList(other),
			},
		},
	})

	runTestConfigForGenTypeFieldsEnum(t, g, obj)
}

func runTestConfigForGenTypeFieldsEnum(t *testing.T, g *Generator, obj *gql.Object) {
	fieldEnum := g.genTypeFieldsEnum(obj)

	assert.Equal(t, len(obj.Fields()), len(fieldEnum.Values()), "Mismatched number of fields for object field enum, want %v, got %v", len(obj.Fields()), len(fieldEnum.Values()))
	for _, field := range obj.Fields() {
		assert.NotNil(t, fieldEnum.ParseValue(field.Name), "Missing field enum for field %s", field.Name)
	}
}

// unwrap List or NonNull types
func unwrapType(t gql.Type) gql.Type {
	switch unwrapped := t.(type) {
	case *gql.List:
		return unwrapType(unwrapped.OfType)
	case *gql.NonNull:
		return unwrapType(unwrapped.OfType)
	default:
		return t
	}
}

func isEqualTypes(t1, t2 gql.Type) bool {
	// equal names
	if t1.Name() != t2.Name() {
		return false
	}
	// prob more things too :/
	return true
}
