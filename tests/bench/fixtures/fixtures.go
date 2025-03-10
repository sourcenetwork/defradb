// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/bxcodec/faker"

	"github.com/sourcenetwork/defradb/errors"
)

var (
	registeredFixtures = map[string][]any{
		"user_simple": {User{}},
	}
)

type Option func(*Generator)

func OptionFieldDirective(typeName, field, directive string) Option {
	return func(g *Generator) {
		if g.directives == nil {
			g.directives = make(map[string]map[string][]string)
		}
		if g.directives[typeName] == nil {
			g.directives[typeName] = make(map[string][]string)
		}
		g.directives[typeName][field] = append(g.directives[typeName][field], directive)
	}
}

type Generator struct {
	ctx context.Context

	// map of type name to field name to list of directives
	directives map[string]map[string][]string

	schema string
	types  []any
}

func ForSchema(ctx context.Context, schemaName string, options ...Option) Generator {
	g := Generator{
		ctx:    ctx,
		schema: schemaName,
		types:  registeredFixtures[schemaName],
	}

	for _, o := range options {
		o(&g)
	}

	return g
}

// Types returns the defined types for this fixture set
func (g Generator) Types() []any {
	return g.types
}

// Type returns type at the given index in the fixture set
func (g Generator) Type(index int) any {
	return g.types[index]
}

// TypeName returns the name of the type at the given index
// in the fixture set
func (g Generator) TypeName(index int) string {
	return reflect.TypeOf(g.types[index]).Name()
}

// GenerateFixtureDocs uses the faker fixture system to
// randomly generate a new set of documents matching the defined
// struct types within the context.
func (g Generator) GenerateDocs() ([]string, error) {
	results := make([]string, len(g.types))
	for i, t := range g.types {
		val := reflect.New(reflect.TypeOf(t)).Interface()

		// generate our new random struct and
		// write it to our reflected variable
		if err := faker.FakeData(val); err != nil {
			return nil, err
		}

		buf, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		results[i] = string(buf)
	}

	return results, nil
}

// extractGQLFromType extracts a GraphQL SDL definition as a string
// from a given type struct
func (g Generator) ExtractGQLFromType(t any) (string, error) {
	var buf bytes.Buffer

	if reflect.TypeOf(t).Kind() != reflect.Struct {
		return "", errors.New("given type is not a struct")
	}

	// get name
	tt := reflect.TypeOf(t)
	name := tt.Name()

	// write the GQL SDL object to the buffer, field by field
	fmt.Fprintf(&buf, "type %s {\n", name)
	for i := 0; i < tt.NumField(); i++ {
		// @todo: Handle non-scalar types
		f := tt.Field(i)
		fname := f.Name
		ftype := f.Type.Name()
		gqlType := gTypeToGQLType[ftype]

		directives := ""
		if g.directives != nil {
			if dirsMap, ok := g.directives[name]; ok {
				if dirs, ok := dirsMap[fname]; ok {
					directives = " " + strings.Join(dirs, " ")
				}
			}
		}
		// write field's name, type and directives
		fmt.Fprintf(&buf, "\t%s: %s%s\n", fname, gqlType, directives)
	}
	fmt.Fprint(&buf, "}")

	return buf.String(), nil
}
