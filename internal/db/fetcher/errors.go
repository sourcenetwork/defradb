// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFieldIdNotFound            string = "unable to find CollectionFieldDescription for given FieldId"
	errFailedToSeek               string = "seek failed"
	errFailedToMergeState         string = "failed merging state"
	errVFetcherFailedToFindBlock  string = "(version fetcher) failed to find block in blockstore"
	errVFetcherFailedToGetBlock   string = "(version fetcher) failed to get block in blockstore"
	errVFetcherFailedToWriteBlock string = "(version fetcher) failed to write block to blockstore"
	errVFetcherFailedToDecodeNode string = "(version fetcher) failed to decode protobuf"
	errVFetcherFailedToGetDagLink string = "(version fetcher) failed to get node link from DAG"
	errFailedToGetDagNode         string = "failed to get DAG Node"
	errMissingMapper              string = "missing document mapper"
	errInvalidInOperatorValue     string = "invalid _in/_nin value"
	errInvalidFilterOperator      string = "invalid filter operator is provided"
	errNotSupportedKindByIndex    string = "kind is not supported by index"
	errUnexpectedTypeValue        string = "unexpected type value"
	errCreateDocIterator          string = "failed to create document iterator"
	errIterateDocuments           string = "failed to iterate documents"
	errParseDocumentKey           string = "failed to parse document key"
	errGetDocumentValue           string = "failed to get document value"
	errIterateDocFields           string = "failed to iterate document fields"
	errParseFieldKey              string = "failed to parse field key"
	errGetFieldValue              string = "failed to get field value"
	errCreateIndexIterator        string = "failed to create index iterator"
	errIterateIndex               string = "failed to iterate index"
	errDecodeIndexKey             string = "failed to decode index key"
	errGetIndexValue              string = "failed to get index value"
	errGetIndexEntry              string = "failed to get index entry"
	errGetNextIndexEntry          string = "failed to get next index entry"
	errCreateHeadIterator         string = "failed to create headstore iterator"
	errIterateHeads               string = "failed to iterate heads"
	errParseHeadKey               string = "failed to parse headstore key"
	errDecodeDocField             string = "failed to decode document field"
	errCopyVersionedData          string = "failed to copy versioned data"
	errCreateVersionIterator      string = "failed to create version data iterator"
)

var (
	ErrFieldIdNotFound            = errors.New(errFieldIdNotFound)
	ErrFailedToSeek               = errors.New(errFailedToSeek)
	ErrFailedToMergeState         = errors.New(errFailedToMergeState)
	ErrVFetcherFailedToFindBlock  = errors.New(errVFetcherFailedToFindBlock)
	ErrVFetcherFailedToGetBlock   = errors.New(errVFetcherFailedToGetBlock)
	ErrVFetcherFailedToWriteBlock = errors.New(errVFetcherFailedToWriteBlock)
	ErrVFetcherFailedToDecodeNode = errors.New(errVFetcherFailedToDecodeNode)
	ErrVFetcherFailedToGetDagLink = errors.New(errVFetcherFailedToGetDagLink)
	ErrFailedToGetDagNode         = errors.New(errFailedToGetDagNode)
	ErrMissingMapper              = errors.New(errMissingMapper)
	ErrInvalidInOperatorValue     = errors.New(errInvalidInOperatorValue)
	ErrInvalidFilterOperator      = errors.New(errInvalidFilterOperator)
	ErrUnexpectedTypeValue        = errors.New(errUnexpectedTypeValue)
)

// NewErrFieldIdNotFound returns an error indicating that the given FieldId was not found.
func NewErrFieldIdNotFound(fieldId uint32) error {
	return errors.New(errFieldIdNotFound, errors.NewKV("FieldId", fieldId))
}

// NewErrFailedToSeek returns an error indicating that the given target could not be seeked to.
func NewErrFailedToSeek(target any, inner error) error {
	return errors.Wrap(errFailedToSeek, inner, errors.NewKV("Target", target))
}

// NewErrFailedToMergeState returns an error indicating that the given state could not be merged.
func NewErrFailedToMergeState(inner error) error {
	return errors.Wrap(errFailedToMergeState, inner)
}

// NewErrVFetcherFailedToFindBlock returns an error indicating that the given block could not be found.
func NewErrVFetcherFailedToFindBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToFindBlock, inner)
}

// NewErrVFetcherFailedToGetBlock returns an error indicating that the given block could not be retrieved.
func NewErrVFetcherFailedToGetBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToGetBlock, inner)
}

// NewErrVFetcherFailedToWriteBlock returns an error indicating that the given block could not be written.
func NewErrVFetcherFailedToWriteBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToWriteBlock, inner)
}

// NewErrVFetcherFailedToDecodeNode returns an error indicating that the given node could not be decoded.
func NewErrVFetcherFailedToDecodeNode(inner error) error {
	return errors.Wrap(errVFetcherFailedToDecodeNode, inner)
}

// NewErrVFetcherFailedToGetDagLink returns an error indicating that the given DAG link
// could not be retrieved.
func NewErrVFetcherFailedToGetDagLink(inner error) error {
	return errors.Wrap(errVFetcherFailedToGetDagLink, inner)
}

// NewErrFailedToGetDagNode returns an error indicating that the given DAG node could not be retrieved.
func NewErrFailedToGetDagNode(inner error) error {
	return errors.Wrap(errFailedToGetDagNode, inner)
}

// NewErrInvalidInOperatorValue returns an error indicating that the given value is invalid for the _in/_nin operator.
func NewErrInvalidInOperatorValue(inner error) error {
	return errors.Wrap(errInvalidInOperatorValue, inner)
}

// NewErrInvalidFilterOperator returns an error indicating that the given filter operator is invalid.
func NewErrInvalidFilterOperator(operator string) error {
	return errors.New(errInvalidFilterOperator, errors.NewKV("Operator", operator))
}

// NewErrNotSupportedKindByIndex returns an error indicating that the given kind is not supported by index.
func NewErrNotSupportedKindByIndex(kind client.FieldKind) error {
	return errors.New(errNotSupportedKindByIndex, errors.NewKV("Kind", kind.String()))
}

// NewErrUnexpectedTypeValue returns an error indicating that the given value is of an unexpected type.
func NewErrUnexpectedTypeValue[T any](value any) error {
	var t T
	return errors.New(errUnexpectedTypeValue, errors.NewKV("Value", value), errors.NewKV("Type", fmt.Sprintf("%T", t)))
}

func NewErrCreateDocIterator(inner error) error {
	return errors.Wrap(errCreateDocIterator, inner)
}

func NewErrIterateDocuments(inner error) error {
	return errors.Wrap(errIterateDocuments, inner)
}

func NewErrParseDocumentKey(inner error) error {
	return errors.Wrap(errParseDocumentKey, inner)
}

func NewErrGetDocumentValue(inner error) error {
	return errors.Wrap(errGetDocumentValue, inner)
}

func NewErrIterateDocFields(inner error) error {
	return errors.Wrap(errIterateDocFields, inner)
}

func NewErrParseFieldKey(inner error) error {
	return errors.Wrap(errParseFieldKey, inner)
}

func NewErrGetFieldValue(inner error) error {
	return errors.Wrap(errGetFieldValue, inner)
}

func NewErrCreateIndexIterator(inner error, indexName string) error {
	return errors.Wrap(errCreateIndexIterator, inner, errors.NewKV("IndexName", indexName))
}

func NewErrIterateIndex(inner error, indexName string) error {
	return errors.Wrap(errIterateIndex, inner, errors.NewKV("IndexName", indexName))
}

func NewErrDecodeIndexKey(inner error, indexName string) error {
	return errors.Wrap(errDecodeIndexKey, inner, errors.NewKV("IndexName", indexName))
}

func NewErrGetIndexValue(inner error, indexName string) error {
	return errors.Wrap(errGetIndexValue, inner, errors.NewKV("IndexName", indexName))
}

func NewErrGetIndexEntry(inner error, indexKey string) error {
	return errors.Wrap(errGetIndexEntry, inner, errors.NewKV("IndexKey", indexKey))
}

func NewErrGetNextIndexEntry(inner error, indexName string) error {
	return errors.Wrap(errGetNextIndexEntry, inner, errors.NewKV("IndexName", indexName))
}

func NewErrCreateHeadIterator(inner error) error {
	return errors.Wrap(errCreateHeadIterator, inner)
}

func NewErrIterateHeads(inner error) error {
	return errors.Wrap(errIterateHeads, inner)
}

func NewErrParseHeadKey(inner error) error {
	return errors.Wrap(errParseHeadKey, inner)
}

func NewErrDecodeDocField(inner error, field string) error {
	return errors.Wrap(errDecodeDocField, inner, errors.NewKV("Field", field))
}

func NewErrCopyVersionedData(inner error) error {
	return errors.Wrap(errCopyVersionedData, inner)
}

func NewErrCreateVersionIterator(inner error) error {
	return errors.Wrap(errCreateVersionIterator, inner)
}
