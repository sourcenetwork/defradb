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

type Options struct {
	directives map[string]map[string][]string
}

func NewOptions() *Options {
	return &Options{
		directives: make(map[string]map[string][]string),
	}
}

func (o *Options) WithFieldDirective(typeName, field, directive string) *Options {
	if o.directives == nil {
		o.directives = make(map[string]map[string][]string)
	}
	if o.directives[typeName] == nil {
		o.directives[typeName] = make(map[string][]string)
	}
	o.directives[typeName][field] = append(o.directives[typeName][field], directive)
	return o
}

type Generator struct {
	ctx context.Context

	schema  string
	types   []any
	options *Options
}

func ForSchema(ctx context.Context, schemaName string, options *Options) Generator {
	return Generator{
		ctx:     ctx,
		schema:  schemaName,
		types:   registeredFixtures[schemaName],
		options: options,
	}
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

		directive := ""
		if g.options != nil {
			if dirsMap, ok := g.options.directives[name]; ok {
				if dirs, ok := dirsMap[fname]; ok {
					directive = " " + strings.Join(dirs, " ")
				}
			}
		}
		fmt.Fprintf(&buf, "\t%s: %s%s\n", fname, gqlType, directive)
	}
	fmt.Fprint(&buf, "}")

	return buf.String(), nil
}
