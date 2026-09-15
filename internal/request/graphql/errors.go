// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"errors"
	"fmt"
)

var ErrGraphQLProtocol = errors.New("graphql protocol error")

type protocolError string

func (e protocolError) Error() string        { return string(e) }
func (e protocolError) Is(target error) bool { return target == ErrGraphQLProtocol }

type syntaxError struct{ cause error }

func (e syntaxError) Error() string        { return "Syntax Error GraphQL: " + e.cause.Error() }
func (e syntaxError) Unwrap() error        { return e.cause }
func (e syntaxError) Is(target error) bool { return target == ErrGraphQLProtocol }

func newErrGraphQLSyntax(cause error) error { return syntaxError{cause: cause} }
func newErrMustProvideOperation() error     { return protocolError("Must provide an operation.") }
func newErrMustProvideOperationName() error {
	return protocolError("Must provide operation name if query contains multiple operations.")
}
func newErrUnknownOperation(name string) error {
	return protocolError(fmt.Sprintf("Unknown operation named %q.", name))
}
