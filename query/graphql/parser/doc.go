// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package parser provides a structured proxy to the underlying GraphQL AST and parser.
Additionally it evaluates the parsed filter conditions on a document.

Given an already parsed GraphQL ast.Document, this package can further parse the document into the
DefraDB GraphQL Query structure, representing, Select statements, fields, filters, arguments, directives, etc.
*/
package parser
