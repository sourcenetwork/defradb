package encryption

import "context"

// docEncContextKey is the key type for document encryption context values.
type docEncContextKey struct{}

// TryGetContextDocEnc returns a document encryption and a bool indicating if
// it was retrieved from the given context.
func TryGetContextDocEnc(ctx context.Context) (*DocCipher, bool) {
	d, ok := ctx.Value(docEncContextKey{}).(*DocCipher)
	return d, ok
}

// NewContext returns a new context with the document encryption value set.
//
// This will overwrite any previously set transaction value.
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, docEncContextKey{}, NewDocCipher())
}
