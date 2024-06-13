package encryption

type DocCipher struct {
	encryptionKey string
}

func NewDocCipher() *DocCipher {
	return &DocCipher{}
}

func (d *DocCipher) setKey(encryptionKey string) {
	d.encryptionKey = encryptionKey
}

func (d *DocCipher) Encrypt(docID string, fieldID int, plainText []byte) ([]byte, error) {
	return EncryptAES(plainText, []byte(d.encryptionKey))
}

func (d *DocCipher) Decrypt(docID string, fieldID int, cipherText []byte) ([]byte, error) {
	return DecryptAES(cipherText, []byte(d.encryptionKey))
}
