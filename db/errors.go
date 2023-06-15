// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetHeads               string = "failed to get document heads"
	errFailedToCreateCollectionQuery  string = "failed to create collection prefix query"
	errFailedToGetCollection          string = "failed to get collection"
	errDocVerification                string = "the document verification failed"
	errAddingP2PCollection            string = "cannot add collection ID"
	errRemovingP2PCollection          string = "cannot remove collection ID"
	errAddCollectionWithPatch         string = "unknown collection, adding collections via patch is not supported"
	errCollectionIDDoesntMatch        string = "CollectionID does not match existing"
	errSchemaIDDoesntMatch            string = "SchemaID does not match existing"
	errCannotModifySchemaName         string = "modifying the schema name is not supported"
	errCannotSetVersionID             string = "setting the VersionID is not supported. It is updated automatically"
	errCannotSetFieldID               string = "explicitly setting a field ID value is not supported"
	errCannotAddRelationalField       string = "the adding of new relation fields is not yet supported"
	errDuplicateField                 string = "duplicate field"
	errCannotMutateField              string = "mutating an existing field is not supported"
	errCannotMoveField                string = "moving fields is not currently supported"
	errInvalidCRDTType                string = "only default or LWW (last writer wins) CRDT types are supported"
	errCannotDeleteField              string = "deleting an existing field is not supported"
	errFieldKindNotFound              string = "no type found for given name"
	errDocumentAlreadyExists          string = "a document with the given dockey already exists"
	errDocumentDeleted                string = "a document with the given dockey has been deleted"
	errIndexMissingFields             string = "index missing fields"
	errNonZeroIndexIDProvided         string = "non-zero index ID provided"
	errIndexFieldMissingName          string = "index field missing name"
	errIndexFieldMissingDirection     string = "index field missing direction"
	errIndexSingleFieldWrongDirection string = "wrong direction for index with a single field"
	errIndexWithNameAlreadyExists     string = "index with name already exists"
	errInvalidStoredIndex             string = "invalid stored index"
	errInvalidStoredIndexKey          string = "invalid stored index key"
	errNonExistingFieldForIndex       string = "creating an index on a non-existing property"
	errCollectionDoesntExisting       string = "collection with given name doesn't exist"
	errFailedToStoreIndexedField      string = "failed to store indexed field"
	errFailedToReadStoredIndexDesc    string = "failed to read stored index description"
	errCanNotDeleteIndexedField       string = "can not delete indexed field"
	errCanNotAddIndexWithPatch        string = "adding indexes via patch is not supported"
	errCanNotDropIndexWithPatch       string = "dropping indexes via patch is not supported"
	errCanNotChangeIndexWithPatch     string = "changing indexes via patch is not supported"
	errIndexWithNameDoesNotExists     string = "index with name doesn't exists"
	errInvalidFieldValue              string = "invalid field value"
	errUnsupportedIndexFieldType      string = "unsupported index field type"
	errIndexDescriptionHasNoFields    string = "index description has no fields"
	errIndexDescHasNonExistingField   string = "index description has non existing field"
)

var (
	ErrFailedToGetHeads              = errors.New(errFailedToGetHeads)
	ErrFailedToCreateCollectionQuery = errors.New(errFailedToCreateCollectionQuery)
	ErrFailedToGetCollection         = errors.New(errFailedToGetCollection)
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key.
	ErrDocVerification         = errors.New(errDocVerification)
	ErrSubscriptionsNotAllowed = errors.New("server does not accept subscriptions")
	ErrDeleteTargetEmpty       = errors.New("the doc delete targeter cannot be empty")
	ErrDeleteEmpty             = errors.New("the doc delete cannot be empty")
	ErrUpdateTargetEmpty       = errors.New("the doc update targeter cannot be empty")
	ErrUpdateEmpty             = errors.New("the doc update cannot be empty")
	ErrInvalidMergeValueType   = errors.New(
		"the type of value in the merge patch doesn't match the schema",
	)
	ErrMissingDocFieldToUpdate        = errors.New("missing document field to update")
	ErrDocMissingKey                  = errors.New("document is missing key")
	ErrMergeSubTypeNotSupported       = errors.New("merge doesn't support sub types yet")
	ErrInvalidFilter                  = errors.New("invalid filter")
	ErrInvalidOpPath                  = errors.New("invalid patch op path")
	ErrDocumentAlreadyExists          = errors.New(errDocumentAlreadyExists)
	ErrDocumentDeleted                = errors.New(errDocumentDeleted)
	ErrUnknownCRDTArgument            = errors.New("invalid CRDT arguments")
	ErrUnknownCRDT                    = errors.New("unknown crdt")
	ErrSchemaFirstFieldDocKey         = errors.New("collection schema first field must be a DocKey")
	ErrCollectionAlreadyExists        = errors.New("collection already exists")
	ErrCollectionNameEmpty            = errors.New("collection name can't be empty")
	ErrSchemaIDEmpty                  = errors.New("schema ID can't be empty")
	ErrSchemaVersionIDEmpty           = errors.New("schema version ID can't be empty")
	ErrKeyEmpty                       = errors.New("key cannot be empty")
	ErrAddingP2PCollection            = errors.New(errAddingP2PCollection)
	ErrRemovingP2PCollection          = errors.New(errRemovingP2PCollection)
	ErrAddCollectionWithPatch         = errors.New(errAddCollectionWithPatch)
	ErrCollectionIDDoesntMatch        = errors.New(errCollectionIDDoesntMatch)
	ErrSchemaIDDoesntMatch            = errors.New(errSchemaIDDoesntMatch)
	ErrCannotModifySchemaName         = errors.New(errCannotModifySchemaName)
	ErrCannotSetVersionID             = errors.New(errCannotSetVersionID)
	ErrCannotSetFieldID               = errors.New(errCannotSetFieldID)
	ErrCannotAddRelationalField       = errors.New(errCannotAddRelationalField)
	ErrDuplicateField                 = errors.New(errDuplicateField)
	ErrCannotMutateField              = errors.New(errCannotMutateField)
	ErrCannotMoveField                = errors.New(errCannotMoveField)
	ErrInvalidCRDTType                = errors.New(errInvalidCRDTType)
	ErrCannotDeleteField              = errors.New(errCannotDeleteField)
	ErrFieldKindNotFound              = errors.New(errFieldKindNotFound)
	ErrIndexMissingFields             = errors.New(errIndexMissingFields)
	ErrIndexFieldMissingName          = errors.New(errIndexFieldMissingName)
	ErrIndexFieldMissingDirection     = errors.New(errIndexFieldMissingDirection)
	ErrIndexSingleFieldWrongDirection = errors.New(errIndexSingleFieldWrongDirection)
	ErrCanNotChangeIndexWithPatch     = errors.New(errCanNotChangeIndexWithPatch)
)

// NewErrFailedToGetHeads returns a new error indicating that the heads of a document
// could not be obtained.
func NewErrFailedToGetHeads(inner error) error {
	return errors.Wrap(errFailedToGetHeads, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed.
func NewErrFailedToCreateCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCreateCollectionQuery, inner)
}

// NewErrInvalidStoredIndex returns a new error indicating that the stored
// index in the database is invalid.
func NewErrInvalidStoredIndex(inner error) error {
	return errors.Wrap(errInvalidStoredIndex, inner)
}

// NewErrInvalidStoredIndexKey returns a new error indicating that the stored
// index in the database is invalid.
func NewErrInvalidStoredIndexKey(key string) error {
	return errors.New(errInvalidStoredIndexKey, errors.NewKV("Key", key))
}

// NewErrNonExistingFieldForIndex returns a new error indicating the attempt to create an index
// on a non-existing field.
func NewErrNonExistingFieldForIndex(field string) error {
	return errors.New(errNonExistingFieldForIndex, errors.NewKV("Field", field))
}

// NewErrCanNotReadCollection returns a new error indicating the collection doesn't exist.
func NewErrCanNotReadCollection(colName string, inner error) error {
	return errors.Wrap(errCollectionDoesntExisting, inner, errors.NewKV("Collection", colName))
}

// NewErrFailedToStoreIndexedField returns a new error indicating that the indexed field
// could not be stored.
func NewErrFailedToStoreIndexedField(key string, inner error) error {
	return errors.Wrap(errFailedToStoreIndexedField, inner, errors.NewKV("Key", key))
}

// NewErrFailedToReadStoredIndexDesc returns a new error indicating that the stored index
// description could not be read.
func NewErrFailedToReadStoredIndexDesc(inner error) error {
	return errors.Wrap(errFailedToReadStoredIndexDesc, inner)
}

// NewCanNotDeleteIndexedField returns a new error a failed attempt to delete an indexed field
func NewCanNotDeleteIndexedField(inner error) error {
	return errors.Wrap(errCanNotDeleteIndexedField, inner)
}

// NewErrNonZeroIndexIDProvided returns a new error indicating that a non-zero index ID was
// provided.
func NewErrNonZeroIndexIDProvided(indexID uint32) error {
	return errors.New(errNonZeroIndexIDProvided, errors.NewKV("ID", indexID))
}

// NewErrFailedToGetCollection returns a new error indicating that the collection could not
// be obtained.
func NewErrFailedToGetCollection(name string, inner error) error {
	return errors.Wrap(errFailedToGetCollection, inner, errors.NewKV("Name", name))
}

// NewErrDocVerification returns a new error indicating that the document verification failed.
func NewErrDocVerification(expected string, actual string) error {
	return errors.New(
		errDocVerification,
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

// NewErrAddingP2PCollection returns a new error indicating that adding a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrAddingP2PCollection(inner error) error {
	return errors.Wrap(errAddingP2PCollection, inner)
}

// NewErrRemovingP2PCollection returns a new error indicating that removing a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrRemovingP2PCollection(inner error) error {
	return errors.Wrap(errRemovingP2PCollection, inner)
}

func NewErrAddCollectionWithPatch(name string) error {
	return errors.New(
		errAddCollectionWithPatch,
		errors.NewKV("Name", name),
	)
}

func NewErrCollectionIDDoesntMatch(name string, existingID, proposedID uint32) error {
	return errors.New(
		errCollectionIDDoesntMatch,
		errors.NewKV("Name", name),
		errors.NewKV("ExistingID", existingID),
		errors.NewKV("ProposedID", proposedID),
	)
}

func NewErrSchemaIDDoesntMatch(name, existingID, proposedID string) error {
	return errors.New(
		errSchemaIDDoesntMatch,
		errors.NewKV("Name", name),
		errors.NewKV("ExistingID", existingID),
		errors.NewKV("ProposedID", proposedID),
	)
}

func NewErrCannotModifySchemaName(existingName, proposedName string) error {
	return errors.New(
		errCannotModifySchemaName,
		errors.NewKV("ExistingName", existingName),
		errors.NewKV("ProposedName", proposedName),
	)
}

func NewErrCannotSetFieldID(name string, id client.FieldID) error {
	return errors.New(
		errCannotSetFieldID,
		errors.NewKV("Field", name),
		errors.NewKV("ID", id),
	)
}

func NewErrCannotAddRelationalField(name string, kind client.FieldKind) error {
	return errors.New(
		errCannotAddRelationalField,
		errors.NewKV("Field", name),
		errors.NewKV("Kind", kind),
	)
}

func NewErrFieldKindNotFound(kind string) error {
	return errors.New(
		errFieldKindNotFound,
		errors.NewKV("Kind", kind),
	)
}

func NewErrDuplicateField(name string) error {
	return errors.New(errDuplicateField, errors.NewKV("Name", name))
}

func NewErrCannotMutateField(id client.FieldID, name string) error {
	return errors.New(
		errCannotMutateField,
		errors.NewKV("ID", id),
		errors.NewKV("ProposedName", name),
	)
}

func NewErrCannotMoveField(name string, proposedIndex, existingIndex int) error {
	return errors.New(
		errCannotMoveField,
		errors.NewKV("Name", name),
		errors.NewKV("ProposedIndex", proposedIndex),
		errors.NewKV("ExistingIndex", existingIndex),
	)
}

func NewErrInvalidCRDTType(name string, crdtType client.CType) error {
	return errors.New(
		errInvalidCRDTType,
		errors.NewKV("Name", name),
		errors.NewKV("CRDTType", crdtType),
	)
}

func NewErrCannotDeleteField(name string, id client.FieldID) error {
	return errors.New(
		errCannotDeleteField,
		errors.NewKV("Name", name),
		errors.NewKV("ID", id),
	)
}

func NewErrDocumentAlreadyExists(dockey string) error {
	return errors.New(
		errDocumentAlreadyExists,
		errors.NewKV("DocKey", dockey),
	)
}

func NewErrDocumentDeleted(dockey string) error {
	return errors.New(
		errDocumentDeleted,
		errors.NewKV("DocKey", dockey),
	)
}

// NewErrIndexWithNameAlreadyExists returns a new error indicating that an index with the
// given name already exists.
func NewErrIndexWithNameAlreadyExists(indexName string) error {
	return errors.New(
		errIndexWithNameAlreadyExists,
		errors.NewKV("Name", indexName),
	)
}

// NewErrIndexWithNameDoesNotExists returns a new error indicating that an index with the
// given name does not exist.
func NewErrIndexWithNameDoesNotExists(indexName string) error {
	return errors.New(
		errIndexWithNameDoesNotExists,
		errors.NewKV("Name", indexName),
	)
}

// NewErrCannotAddIndexWithPatch returns a new error indicating that an index cannot be added
// with a patch.
func NewErrCannotAddIndexWithPatch(proposedName string) error {
	return errors.New(
		errCanNotAddIndexWithPatch,
		errors.NewKV("ProposedName", proposedName),
	)
}

// NewErrCannotDropIndexWithPatch returns a new error indicating that an index cannot be dropped
// with a patch.
func NewErrCannotDropIndexWithPatch(indexName string) error {
	return errors.New(
		errCanNotDropIndexWithPatch,
		errors.NewKV("Name", indexName),
	)
}

// NewErrInvalidFieldValue returns a new error indicating that the given value is invalid for the
// given field kind.
func NewErrInvalidFieldValue(kind client.FieldKind, value any) error {
	return errors.New(
		errInvalidFieldValue,
		errors.NewKV("Kind", kind),
		errors.NewKV("Value", value),
	)
}

// NewErrUnsupportedIndexFieldType returns a new error indicating that the given field kind is not
// supported for indexing.
func NewErrUnsupportedIndexFieldType(kind client.FieldKind) error {
	return errors.New(
		errUnsupportedIndexFieldType,
		errors.NewKV("Kind", kind),
	)
}

// NewErrIndexDescHasNoFields returns a new error indicating that the given index
// description has no fields.
func NewErrIndexDescHasNoFields(desc client.IndexDescription) error {
	return errors.New(
		errIndexDescriptionHasNoFields,
		errors.NewKV("Description", desc),
	)
}

// NewErrIndexDescHasNonExistingField returns a new error indicating that the given index
// description points to a field that does not exist.
func NewErrIndexDescHasNonExistingField(desc client.IndexDescription, fieldName string) error {
	return errors.New(
		errIndexDescHasNonExistingField,
		errors.NewKV("Description", desc),
		errors.NewKV("Field name", fieldName),
	)
}
