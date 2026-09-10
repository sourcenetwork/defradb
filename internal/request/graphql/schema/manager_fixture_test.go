// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/testfixtures"
)

// TestGeneratedSchemaMatchesCommittedFixtures preserves the public schema
// surface captured before the code-first generator was removed. It runs with
// the standard Go test suite and requires neither graphql-go nor Node.
func TestGeneratedSchemaMatchesCommittedFixtures(t *testing.T) {
	for _, fixture := range testfixtures.Fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			generated, err := testfixtures.GenerateSDL(t.Context(), fixture.SDL)
			require.NoError(t, err)
			committed, err := os.ReadFile(testfixtures.Path(fixture.Name))
			require.NoError(t, err)

			require.Equal(t, schemaSurface(t, committed), schemaSurface(t, generated))
		})
	}
}

func schemaSurface(t *testing.T, sdl []byte) map[string][]string {
	t.Helper()
	document, report := astparser.ParseGraphqlDocumentBytes(sdl)
	require.False(t, report.HasErrors(), report.Error())
	astnormalization.NormalizeDefinition(&document, &report)
	require.False(t, report.HasErrors(), report.Error())
	var data introspection.Data
	introspection.NewGenerator().Generate(&document, &report, &data)
	require.False(t, report.HasErrors(), report.Error())
	result := make(map[string][]string, len(data.Schema.Types)+1)
	for _, gqlType := range data.Schema.Types {
		members := []string{"kind:" + gqlType.Kind.String()}
		for _, field := range gqlType.Fields {
			args := make([]string, len(field.Args))
			for index, arg := range field.Args {
				args[index] = fixtureInputValue(arg)
			}
			sort.Strings(args)
			members = append(members, fmt.Sprintf("field:%s(%s):%s:deprecated=%t:%v",
				field.Name, strings.Join(args, ","), fixtureTypeRef(field.Type),
				field.IsDeprecated, fixtureOptionalString(field.DeprecationReason)))
		}
		for _, field := range gqlType.InputFields {
			members = append(members, "input:"+fixtureInputValue(field))
		}
		for _, value := range gqlType.EnumValues {
			members = append(members, fmt.Sprintf("enum:%s:deprecated=%t:%v",
				value.Name, value.IsDeprecated, fixtureOptionalString(value.DeprecationReason)))
		}
		for _, implemented := range gqlType.Interfaces {
			members = append(members, "interface:"+fixtureTypeRef(implemented))
		}
		for _, possible := range gqlType.PossibleTypes {
			members = append(members, "possible:"+fixtureTypeRef(possible))
		}
		sort.Strings(members)
		result["type:"+gqlType.Name] = members
	}
	for _, directive := range data.Schema.Directives {
		args := make([]string, len(directive.Args))
		for index, arg := range directive.Args {
			args[index] = fixtureInputValue(arg)
		}
		sort.Strings(args)
		locations := append([]string(nil), directive.Locations...)
		sort.Strings(locations)
		result["directive:"+directive.Name] = []string{
			fmt.Sprintf("args:%s", strings.Join(args, ",")),
			fmt.Sprintf("locations:%s", strings.Join(locations, ",")),
			fmt.Sprintf("repeatable:%t", directive.IsRepeatable),
		}
	}
	return result
}

func fixtureInputValue(value introspection.InputValue) string {
	return fmt.Sprintf("%s:%s:default=%v:deprecated=%t:%v", value.Name, fixtureTypeRef(value.Type),
		fixtureOptionalString(value.DefaultValue), value.IsDeprecated, fixtureOptionalString(value.DeprecationReason))
}

func fixtureOptionalString(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}

func fixtureTypeRef(ref introspection.TypeRef) string {
	if ref.Name != nil {
		return *ref.Name
	}
	if ref.OfType == nil {
		return ""
	}
	name := fixtureTypeRef(*ref.OfType)
	switch ref.Kind.String() {
	case "NON_NULL":
		return name + "!"
	case "LIST":
		return "[" + name + "]"
	default:
		return name
	}
}
