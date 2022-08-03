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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bxcodec/faker"
	"github.com/cip8/autoname"
)

type ID string

type relation string

const (
	oneToOne   = relation("one-to-one")
	oneToMany  = relation("one-to-many")
	fixtureTag = "fixture"
)

var (
	// A map of names => schema groups.
	//
	// The order of types is *very* important, as it
	// needs to follow the dependancy order of related types.
	// Since they will be inserted in order.
	registeredFixtures = map[string][]interface{}{
		"user_simple":           {User{}},
		"book_publisher_author": {Book{}, Author{}, Publisher{}},
	}
)

func init() {
	faker.AddProvider("title", titleGenerator)
}

func titleGenerator(v reflect.Value) (interface{}, error) {
	return strings.Title(autoname.Generate(" ")), nil
}

type Generator struct {
	ctx context.Context

	schema string
	types  []interface{}
	tags   []tag

	// an array of ratios.
	ratios []float64
}

func ForSchema(ctx context.Context, schemaName string) (Generator, error) {
	types := registeredFixtures[schemaName]
	g := Generator{
		ctx:    ctx,
		schema: schemaName,
	}

	// dependancy graph for related types
	var depGraph Graph
	for _, t := range types {
		v := reflect.TypeOf(t)
		deps, err := dependants(v)
		if err != nil {
			return g, fmt.Errorf("couldn't get dependants for %s: %w", v.Name(), err)
		}
		depGraph = append(depGraph, NewNode(v.Name(), t, deps...))
	}

	var err error
	depGraph, err = resolveGraph(depGraph)
	if err != nil {
		return g, fmt.Errorf("dependancy graph resolution: %w", err)
	}

	// reorder types array based on dependancy graph
	depTypes := make([]interface{}, len(types))
	for i, t := range types {
		for _, n := range depGraph {
			typeName := reflect.TypeOf(t).Name()
			if typeName == n.name {
				depTypes[i] = t
				break
			}
		}
	}

	// compute ratios
	for i, t := range depTypes {
		typeOf := reflect.TypeOf(t)
		for j := 0; j < typeOf.NumField(); j++ {
			f := typeOf.Field(j)

			// if we have a related type, we need to
			if isStructPointer(f.Type) || isSliceStructPointer(f.Type) {
				fixtureTagStr, ok := f.Tag.Lookup(fixtureTag)
				if !ok {
					return g, fmt.Errorf("missing fixture tag for field %d", j)
				}

				tg, err := parseTag(fixtureTagStr)
				if err != nil {
					return g, err
				}
				g.tags[i] = tg

				// start with average. Todo: min/max
				// r, err := tg.avgRatio()
				// if err != nil {
				// 	return g, fmt.Errorf("average ratio for %s: %w", f.Type.Name(), err)
				// }

			}
		}
	}

	g.types = depTypes

	return g, nil
}

// Types returns the defined types for this fixture set
func (g Generator) Types() []interface{} {
	return g.types
}

// Type returns type at the given index in the fixture set
func (g Generator) Type(index int) interface{} {
	return g.types[index]
}

// TypeName returns the name of the type at the given index
// in the fixture set
func (g Generator) TypeName(index int) string {
	return reflect.TypeOf(g.types[index]).Name()
}

// HasRelatedTypes returns if any of the types defined
// in the generator are related to one another
func (g Generator) HasRelatedTypes() bool {
	for _, t := range g.types {
		typeOf := reflect.TypeOf(t)
		for i := 0; i < typeOf.NumField(); i++ {
			f := typeOf.Field(i)
			if isStructPointer(f.Type) || isSliceStructPointer(f.Type) {
				return true
			}
		}
	}
	return false
}

// GenerateFixtureDocs uses the faker fixture system to
// randomly generate a new set of documents matching the defined
// struct types within the context.
//
// withRelated is an optional bool that will
func (g Generator) GenerateDocs(withRelated ...bool) ([]string, error) {
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

func generateDoc(doc interface{}) error {
	return nil
}

// extractGQLFromType extracts a GraphQL SDL definition as a string
// from a given type struct
func ExtractGQLFromType(t interface{}) (string, error) {
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
		// process scalars, objects, slices of objects,
		// skip foreignKey (ID) fields
		var gqlType string
		if ftype == "ID" {
			continue
		} else if isStructPointer(f.Type) {
			// object
			gqlType = f.Type.Elem().Name()
			// check tag
			ft, err := getFixtureTag(f)
			if err != nil {
				return "", fmt.Errorf("bad fixture tag for %s: %w", fname, err)
			}

			if ft.rel == oneToOne {
				// explicitly seperate from the above if condition
				// so as to not trigger the else case if its not
				// a primary one-to-one
				if ft.isPrimary {
					gqlType = gqlType + " @primary"
				}
			} else if ft.rel == oneToMany {
				// one side of one-to-many
				// noop - seperated to cleanly handle the else case
				// without this branch
			} else {
				// err
				return "", fmt.Errorf("invalid relation type for object")
			}
		} else if isSliceStructPointer(f.Type) {
			// object slice
			fmt.Print("IS STRUCT POINTER:", f)
			gqlType = fmt.Sprintf("[%s]", f.Type.Elem().Elem().Name())
			// check tag
			ft, err := getFixtureTag(f)
			if err != nil {
				return "", fmt.Errorf("bad fixture tag for %s: %w", fname, err)
			}

			if ft.rel != oneToMany {
				return "", fmt.Errorf("slice of objects must be a one-to-many relationship")
			}
		} else if f.Type.Kind() == reflect.Slice {
			// scalar slice
			gqlType = gTypeToGQLType[f.Type.Elem().Name()]
			gqlType = fmt.Sprintf("[%s]", gqlType)
		} else {
			// scalar
			gqlType = gTypeToGQLType[ftype]
		}
		fmt.Fprintf(&buf, "\t%s: %s\n", fname, gqlType)
	}
	fmt.Fprint(&buf, "}")

	return buf.String(), nil
}

// func getGQLType(typ string) string {
// 	return gTypeToGQLType[typ]
// }

func getFixtureTag(f reflect.StructField) (tag, error) {
	fixtureTagStr := f.Tag.Get(fixtureTag)
	if fixtureTagStr == "" {
		return tag{}, fmt.Errorf("missing fixture tag")
	}

	return parseTag(fixtureTagStr)
}

// Fixture Tag
//
// format: fixture:"<relation>[,primary],<fill_rate>[,min_objects][,max-objects]"
// < _ > indicates required
// [ _ ] indicates optional (depending on relation)
type tag struct {
	rel        relation
	isPrimary  bool
	fillRate   float64
	minObjects int64
	maxObjects int64
}

func parseTag(t string) (tag, error) {
	if t == "" {
		return tag{}, fmt.Errorf("empty fixture tag")
	}

	tg := tag{
		minObjects: 1, // default is 1 not 0
		maxObjects: 1, // default is 1 not 0
	}
	parts := strings.Split(t, ",")
	remainderIndex := 1
	switch parts[0] {
	case string(oneToOne):
		tg.rel = oneToOne
		// check for primary
		if len(parts) > 1 && parts[1] == "primary" {
			tg.isPrimary = true
			remainderIndex = 2
			// required to have fill rate (remainderIndex + 3)
			if len(parts) < remainderIndex+1 {
				return tag{}, fmt.Errorf("missing required fixture args")
			}
		}

	case string(oneToMany):
		tg.rel = oneToMany
	default:
		return tag{}, fmt.Errorf("invalid relation for fixture tag: %s", parts[0])
	}

	// parse remaining values
	// fillrate, minObjects, maxObjects
	i := 0
	var err error
	for j, v := range parts[remainderIndex:] {
		switch i {
		case 0: //fill
			tg.fillRate, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return tag{}, fmt.Errorf("invalid fill rate %s: %w", v, err)
			}

			// fill rate between 0-1.0
			if tg.fillRate > 1.0 || 0 > tg.fillRate {
				return tag{}, fmt.Errorf("invalid fill rate, outside 0-1.0 range: %v", tg.fillRate)
			}
		case 1: //min
			tg.minObjects, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return tag{}, fmt.Errorf("invalid minObject value %s: %w", v, err)
			}
		case 2: //max
			tg.maxObjects, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return tag{}, fmt.Errorf("invalid maxObject value %s: %w", v, err)
			}
		default:
			return tag{}, fmt.Errorf("unknown tag parameter %s at index %v", v, j)
		}
		i++
	}

	return tg, nil
}

func (t tag) avgRatio() (ratio, error) {
	avg := float64(t.maxObjects+t.minObjects) / float64(2)
	return newRatio(1, avg/t.fillRate)
}

func (t tag) minRatio() (ratio, error) {
	return newRatio(1, float64(t.minObjects)/t.fillRate)
}

func (t tag) maxRatio() (ratio, error) {
	return newRatio(1, float64(t.maxObjects)/t.fillRate)
}
