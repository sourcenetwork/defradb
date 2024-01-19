// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
genopenapi is a tool to generate and print an OpenAPI specification.
*/
package main

import (
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/http"
)

func main() {
	router, err := http.NewApiRouter()
	if err != nil {
		panic(err)
	}
	json, err := router.OpenAPI().MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Fprint(os.Stdout, string(json))
}
