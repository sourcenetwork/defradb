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

func SetDocEncContext(ctx context.Context, encryptionKey string) context.Context {
	cipher, ok := TryGetContextDocEnc(ctx)
	if !ok {
		cipher = NewDocCipher()
		ctx = context.WithValue(ctx, docEncContextKey{}, cipher)
	}
	cipher.setKey(encryptionKey)
	return ctx
}
