package encryption

type DocCipher struct {
}

const testKey = "examplekey1234567890examplekey12"

func NewDocCipher() *DocCipher {
	return &DocCipher{}
}

func (d *DocCipher) Encrypt(docID string, fieldID int, plainText []byte) ([]byte, error) {
	return EncryptAES(plainText, []byte(testKey))
}

func (d *DocCipher) Decrypt(docID string, fieldID int, cipherText []byte) ([]byte, error) {
	return DecryptAES(cipherText, []byte(testKey))
}
