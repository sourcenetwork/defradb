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
	errFailedToGetHeads                         string = "failed to get document heads"
	errFailedToCreateCollectionQuery            string = "failed to create collection prefix query"
	errFailedToGetCollection                    string = "failed to get collection"
	errFailedToGetAllCollections                string = "failed to get all collections"
	errDocVerification                          string = "the document verification failed"
	errAddingP2PCollection                      string = "cannot add collection ID"
	errRemovingP2PCollection                    string = "cannot remove collection ID"
	errAddCollectionWithPatch                   string = "adding collections via patch is not supported"
	errCollectionIDDoesntMatch                  string = "CollectionID does not match existing"
	errSchemaRootDoesntMatch                    string = "SchemaRoot does not match existing"
	errCannotModifySchemaName                   string = "modifying the schema name is not supported"
	errCannotSetVersionID                       string = "setting the VersionID is not supported"
	errRelationalFieldInvalidRelationType       string = "invalid RelationType"
	errRelationalFieldMissingIDField            string = "missing id field for relation object field"
	errRelationalFieldMissingRelationName       string = "missing relation name"
	errPrimarySideOnMany                        string = "cannot set the many side of a relation as primary"
	errBothSidesPrimary                         string = "both sides of a relation cannot be primary"
	errRelatedFieldKindMismatch                 string = "invalid Kind of the related field"
	errRelatedFieldRelationTypeMismatch         string = "invalid RelationType of the related field"
	errRelationalFieldIDInvalidType             string = "relational id field of invalid kind"
	errDuplicateField                           string = "duplicate field"
	errCannotMutateField                        string = "mutating an existing field is not supported"
	errCannotMoveField                          string = "moving fields is not currently supported"
	errCannotDeleteField                        string = "deleting an existing field is not supported"
	errFieldKindNotFound                        string = "no type found for given name"
	errFieldKindDoesNotMatchFieldSchema         string = "field Kind does not match field Schema"
	errDocumentAlreadyExists                    string = "a document with the given ID already exists"
	errDocumentDeleted                          string = "a document with the given ID has been deleted"
	errIndexMissingFields                       string = "index missing fields"
	errNonZeroIndexIDProvided                   string = "non-zero index ID provided"
	errIndexFieldMissingName                    string = "index field missing name"
	errIndexFieldMissingDirection               string = "index field missing direction"
	errIndexWithNameAlreadyExists               string = "index with name already exists"
	errInvalidStoredIndex                       string = "invalid stored index"
	errInvalidStoredIndexKey                    string = "invalid stored index key"
	errNonExistingFieldForIndex                 string = "creating an index on a non-existing property"
	errCollectionDoesntExisting                 string = "collection with given name doesn't exist"
	errFailedToStoreIndexedField                string = "failed to store indexed field"
	errFailedToReadStoredIndexDesc              string = "failed to read stored index description"
	errCanNotDeleteIndexedField                 string = "can not delete indexed field"
	errCanNotAddIndexWithPatch                  string = "adding indexes via patch is not supported"
	errCanNotDropIndexWithPatch                 string = "dropping indexes via patch is not supported"
	errCanNotChangeIndexWithPatch               string = "changing indexes via patch is not supported"
	errIndexWithNameDoesNotExists               string = "index with name doesn't exists"
	errCorruptedIndex                           string = "corrupted index. Please delete and recreate the index"
	errInvalidFieldValue                        string = "invalid field value"
	errUnsupportedIndexFieldType                string = "unsupported index field type"
	errIndexDescriptionHasNoFields              string = "index description has no fields"
	errFieldOrAliasToFieldNotExist              string = "The given field or alias to field does not exist"
	errCreateFile                               string = "failed to create file"
	errRemoveFile                               string = "failed to remove file"
	errOpenFile                                 string = "failed to open file"
	errCloseFile                                string = "failed to close file"
	errFailedtoCloseQueryReqAllIDs              string = "failed to close query requesting all docIDs"
	errFailedToReadByte                         string = "failed to read byte"
	errFailedToWriteString                      string = "failed to write string"
	errJSONDecode                               string = "failed to decode JSON"
	errDocFromMap                               string = "failed to create a new doc from map"
	errDocCreate                                string = "failed to save a new doc to collection"
	errDocUpdate                                string = "failed to update doc to collection"
	errExpectedJSONObject                       string = "expected JSON object"
	errExpectedJSONArray                        string = "expected JSON array"
	errOneOneAlreadyLinked                      string = "target document is already linked to another document"
	errIndexDoesNotMatchName                    string = "the index used does not match the given name"
	errCanNotIndexNonUniqueFields               string = "can not index a doc's field(s) that violates unique index"
	errInvalidViewQuery                         string = "the query provided is not valid as a View"
	errCollectionAlreadyExists                  string = "collection already exists"
	errMultipleActiveCollectionVersions         string = "multiple versions of same collection cannot be active"
	errCollectionSourcesCannotBeAddedRemoved    string = "collection sources cannot be added or removed"
	errCollectionSourceIDMutated                string = "collection source ID cannot be mutated"
	errCollectionIndexesCannotBeMutated         string = "collection indexes cannot be mutated"
	errCollectionFieldsCannotBeMutated          string = "collection fields cannot be mutated"
	errCollectionRootIDCannotBeMutated          string = "collection root ID cannot be mutated"
	errCollectionSchemaVersionIDCannotBeMutated string = "collection schema version ID cannot be mutated"
	errCollectionIDCannotBeZero                 string = "collection ID cannot be zero"
	errCollectionsCannotBeDeleted               string = "collections cannot be deleted"
)

var (
	ErrFailedToGetCollection                    = errors.New(errFailedToGetCollection)
	ErrSubscriptionsNotAllowed                  = errors.New("server does not accept subscriptions")
	ErrInvalidFilter                            = errors.New("invalid filter")
	ErrCollectionAlreadyExists                  = errors.New(errCollectionAlreadyExists)
	ErrCollectionNameEmpty                      = errors.New("collection name can't be empty")
	ErrSchemaNameEmpty                          = errors.New("schema name can't be empty")
	ErrSchemaRootEmpty                          = errors.New("schema root can't be empty")
	ErrSchemaVersionIDEmpty                     = errors.New("schema version ID can't be empty")
	ErrKeyEmpty                                 = errors.New("key cannot be empty")
	ErrCannotSetVersionID                       = errors.New(errCannotSetVersionID)
	ErrIndexMissingFields                       = errors.New(errIndexMissingFields)
	ErrIndexFieldMissingName                    = errors.New(errIndexFieldMissingName)
	ErrCorruptedIndex                           = errors.New(errCorruptedIndex)
	ErrExpectedJSONObject                       = errors.New(errExpectedJSONObject)
	ErrExpectedJSONArray                        = errors.New(errExpectedJSONArray)
	ErrInvalidViewQuery                         = errors.New(errInvalidViewQuery)
	ErrCanNotIndexNonUniqueFields               = errors.New(errCanNotIndexNonUniqueFields)
	ErrMultipleActiveCollectionVersions         = errors.New(errMultipleActiveCollectionVersions)
	ErrCollectionSourcesCannotBeAddedRemoved    = errors.New(errCollectionSourcesCannotBeAddedRemoved)
	ErrCollectionSourceIDMutated                = errors.New(errCollectionSourceIDMutated)
	ErrCollectionIndexesCannotBeMutated         = errors.New(errCollectionIndexesCannotBeMutated)
	ErrCollectionFieldsCannotBeMutated          = errors.New(errCollectionFieldsCannotBeMutated)
	ErrCollectionRootIDCannotBeMutated          = errors.New(errCollectionRootIDCannotBeMutated)
	ErrCollectionSchemaVersionIDCannotBeMutated = errors.New(errCollectionSchemaVersionIDCannotBeMutated)
	ErrCollectionIDCannotBeZero                 = errors.New(errCollectionIDCannotBeZero)
	ErrCollectionsCannotBeDeleted               = errors.New(errCollectionsCannotBeDeleted)
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

// NewErrFailedToGetAllCollections returns a new error indicating that the collection list could not
// be obtained.
func NewErrFailedToGetAllCollections(inner error) error {
	return errors.Wrap(errFailedToGetAllCollections, inner)
}

// NewErrDocVerification returns a new error indicating that the document verification failed.
//
// This occurs when a documents contents fail the verification during a Create()
// call against the supplied Document ID (docID).
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

func NewErrAddCollectionIDWithPatch(id uint32) error {
	return errors.New(
		errAddCollectionWithPatch,
		errors.NewKV("ID", id),
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

func NewErrSchemaRootDoesntMatch(name, existingRoot, proposedRoot string) error {
	return errors.New(
		errSchemaRootDoesntMatch,
		errors.NewKV("Name", name),
		errors.NewKV("ExistingRoot", existingRoot),
		errors.NewKV("ProposedRoot", proposedRoot),
	)
}

func NewErrCannotModifySchemaName(existingName, proposedName string) error {
	return errors.New(
		errCannotModifySchemaName,
		errors.NewKV("ExistingName", existingName),
		errors.NewKV("ProposedName", proposedName),
	)
}

func NewErrRelationalFieldMissingIDField(name string, expectedName string) error {
	return errors.New(
		errRelationalFieldMissingIDField,
		errors.NewKV("Field", name),
		errors.NewKV("ExpectedIDFieldName", expectedName),
	)
}

func NewErrRelationalFieldMissingRelationName(name string) error {
	return errors.New(
		errRelationalFieldMissingRelationName,
		errors.NewKV("Field", name),
	)
}

func NewErrPrimarySideOnMany(name string) error {
	return errors.New(
		errPrimarySideOnMany,
		errors.NewKV("Field", name),
	)
}

func NewErrBothSidesPrimary(relationName string) error {
	return errors.New(
		errBothSidesPrimary,
		errors.NewKV("RelationName", relationName),
	)
}

func NewErrRelatedFieldKindMismatch(relationName string, expected client.FieldKind, actual client.FieldKind) error {
	return errors.New(
		errRelatedFieldKindMismatch,
		errors.NewKV("RelationName", relationName),
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

func NewErrRelationalFieldIDInvalidType(name string, expected, actual client.FieldKind) error {
	return errors.New(
		errRelationalFieldIDInvalidType,
		errors.NewKV("Field", name),
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

func NewErrFieldKindNotFound(name string, kind string) error {
	return errors.New(
		errFieldKindNotFound,
		errors.NewKV("Field", name),
		errors.NewKV("Kind", kind),
	)
}

func NewErrFieldKindDoesNotMatchFieldSchema(kind string, schema string) error {
	return errors.New(
		errFieldKindDoesNotMatchFieldSchema,
		errors.NewKV("Kind", kind),
		errors.NewKV("Schema", schema),
	)
}

func NewErrDuplicateField(name string) error {
	return errors.New(errDuplicateField, errors.NewKV("Name", name))
}

func NewErrCannotMutateField(name string) error {
	return errors.New(
		errCannotMutateField,
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

func NewErrCannotDeleteField(name string) error {
	return errors.New(
		errCannotDeleteField,
		errors.NewKV("Name", name),
	)
}

func NewErrDocumentAlreadyExists(docID string) error {
	return errors.New(
		errDocumentAlreadyExists,
		errors.NewKV("DocID", docID),
	)
}

func NewErrDocumentDeleted(docID string) error {
	return errors.New(
		errDocumentDeleted,
		errors.NewKV("DocID", docID),
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

// NewErrCorruptedIndex returns a new error indicating that an index with the
// given name has been corrupted.
func NewErrCorruptedIndex(indexName string) error {
	return errors.New(
		errCorruptedIndex,
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

// NewErrCreateFile returns a new error indicating there was a failure in creating a file.
func NewErrCreateFile(inner error, filepath string) error {
	return errors.Wrap(errCreateFile, inner, errors.NewKV("Filepath", filepath))
}

// NewErrOpenFile returns a new error indicating there was a failure in opening a file.
func NewErrOpenFile(inner error, filepath string) error {
	return errors.Wrap(errOpenFile, inner, errors.NewKV("Filepath", filepath))
}

// NewErrCloseFile returns a new error indicating there was a failure in closing a file.
func NewErrCloseFile(closeErr, other error) error {
	if other != nil {
		return errors.Wrap(errCloseFile, closeErr, errors.NewKV("Other error", other))
	}
	return errors.Wrap(errCloseFile, closeErr)
}

// NewErrRemoveFile returns a new error indicating there was a failure in removing a file.
func NewErrRemoveFile(removeErr, other error, filepath string) error {
	if other != nil {
		return errors.Wrap(
			errRemoveFile,
			removeErr,
			errors.NewKV("Other error", other),
			errors.NewKV("Filepath", filepath),
		)
	}
	return errors.Wrap(
		errRemoveFile,
		removeErr,
		errors.NewKV("Filepath", filepath),
	)
}

// NewErrFailedToReadByte returns a new error indicating there was a failure in read a byte
// from the Reader
func NewErrFailedToReadByte(inner error) error {
	return errors.Wrap(errFailedToReadByte, inner)
}

// NewErrFailedToWriteString returns a new error indicating there was a failure in writing
// a string to the Writer
func NewErrFailedToWriteString(inner error) error {
	return errors.Wrap(errFailedToWriteString, inner)
}

// NewErrJSONDecode returns a new error indicating there was a failure in decoding some JSON
// from the JSON decoder
func NewErrJSONDecode(inner error) error {
	return errors.Wrap(errJSONDecode, inner)
}

// NewErrDocFromMap returns a new error indicating there was a failure to create
// a new doc from a map
func NewErrDocFromMap(inner error) error {
	return errors.Wrap(errDocFromMap, inner)
}

// NewErrDocCreate returns a new error indicating there was a failure to save
// a new doc to a collection
func NewErrDocCreate(inner error) error {
	return errors.Wrap(errDocCreate, inner)
}

// NewErrDocUpdate returns a new error indicating there was a failure to update
// a doc to a collection
func NewErrDocUpdate(inner error) error {
	return errors.Wrap(errDocUpdate, inner)
}

func NewErrOneOneAlreadyLinked(documentId, targetId, relationName string) error {
	return errors.New(
		errOneOneAlreadyLinked,
		errors.NewKV("DocumentID", documentId),
		errors.NewKV("TargetID", targetId),
		errors.NewKV("RelationName", relationName),
	)
}

func NewErrIndexDoesNotMatchName(index, name string) error {
	return errors.New(
		errIndexDoesNotMatchName,
		errors.NewKV("Index", index),
		errors.NewKV("Name", name),
	)
}

func NewErrCanNotIndexNonUniqueFields(docID string, fieldValues ...errors.KV) error {
	kvPairs := make([]errors.KV, 0, len(fieldValues)+1)
	kvPairs = append(kvPairs, errors.NewKV("DocID", docID))
	kvPairs = append(kvPairs, fieldValues...)

	return errors.New(errCanNotIndexNonUniqueFields, kvPairs...)
}

func NewErrInvalidViewQueryCastFailed(query string) error {
	return errors.New(
		errInvalidViewQuery,
		errors.NewKV("Query", query),
		errors.NewKV("Reason", "Internal error, cast failed"),
	)
}

func NewErrInvalidViewQueryMissingQuery() error {
	return errors.New(
		errInvalidViewQuery,
		errors.NewKV("Reason", "No query provided"),
	)
}

func NewErrCollectionAlreadyExists(name string) error {
	return errors.New(
		errCollectionAlreadyExists,
		errors.NewKV("Name", name),
	)
}

func NewErrCollectionIDAlreadyExists(id uint32) error {
	return errors.New(
		errCollectionAlreadyExists,
		errors.NewKV("ID", id),
	)
}

func NewErrMultipleActiveCollectionVersions(name string, root uint32) error {
	return errors.New(
		errMultipleActiveCollectionVersions,
		errors.NewKV("Name", name),
		errors.NewKV("Root", root),
	)
}

func NewErrCollectionSourcesCannotBeAddedRemoved(colID uint32) error {
	return errors.New(
		errCollectionSourcesCannotBeAddedRemoved,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionSourceIDMutated(colID uint32, newSrcID uint32, oldSrcID uint32) error {
	return errors.New(
		errCollectionSourceIDMutated,
		errors.NewKV("CollectionID", colID),
		errors.NewKV("NewCollectionSourceID", newSrcID),
		errors.NewKV("OldCollectionSourceID", oldSrcID),
	)
}

func NewErrCollectionIndexesCannotBeMutated(colID uint32) error {
	return errors.New(
		errCollectionIndexesCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionFieldsCannotBeMutated(colID uint32) error {
	return errors.New(
		errCollectionFieldsCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionRootIDCannotBeMutated(colID uint32) error {
	return errors.New(
		errCollectionRootIDCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionSchemaVersionIDCannotBeMutated(colID uint32) error {
	return errors.New(
		errCollectionSchemaVersionIDCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionsCannotBeDeleted(colID uint32) error {
	return errors.New(
		errCollectionsCannotBeDeleted,
		errors.NewKV("CollectionID", colID),
	)
}
