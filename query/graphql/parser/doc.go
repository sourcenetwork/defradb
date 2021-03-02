// Package parser provides a structured proxy to the underlying
// GraphQL AST and parser. Additionally it evaluates the parsed
// filter conditions on a document.
//
// Given an already parsed GraphQL ast.Document, this package
// can further parse the document into the DefraDB GraphQL
// Query structure, representing, Select statements, fields,
// filters, arguments, directives, etc.
package parser
