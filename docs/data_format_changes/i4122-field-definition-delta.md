# Remove isArray from field definition delta

Change the FieldDefinitionDelta so that it doesn't need to specify if the field is an array type. Doing this caused the field definition block content to change and thus affected its CID, the one of the collection and thus the docIDs.