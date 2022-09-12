// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package defaults

// makeInputObject retrned a properly made input field type
// using name (outer), name of type (inner), and types ofType.
func makeInputObject(
	name string,
	typeName any,
	ofType any,
) field {
	return field{
		"name": name,
		"type": field{
			"name":   typeName,
			"ofType": ofType,
		},
	}
}

func makeAuthorOrdering(name string) field {
	return makeInputObject(name, "Ordering", nil)
}

func MakeDefaultGroupArgsWithout(skipThese []string) []field {
	defaultsWithSomeSkipped := make(
		[]map[string]any,
		0,
		len(allDefaultGroupArgs)-len(skipThese),
	)

ArgLoop:
	for _, argAny := range allDefaultGroupArgs {
		arg, ok := argAny.(field)
		if !ok {
			panic("wrong usage of testing schema factory method.")
		}

		argName, found := arg["name"]
		if !found {
			panic("`name` not found in the default group arg.")
		}

		for _, toSkip := range skipThese {
			if toSkip == argName {
				continue ArgLoop
			}
		}

		// If we make it till here we, don't want to skip this arg.
		defaultsWithSomeSkipped = append(defaultsWithSomeSkipped, arg)

	}

	return defaultsWithSomeSkipped

}
