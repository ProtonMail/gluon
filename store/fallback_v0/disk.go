package fallback_v0

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"

	"github.com/ProtonMail/gluon/store"
)

type onDiskStoreV0 struct {
	compressor Compressor
}

func NewOnDiskStoreV0() store.Fallback {
	return &onDiskStoreV0{compressor: nil}
}

func NewOnDiskStoreV0WithCompressor(c Compressor) store.Fallback {
	return &onDiskStoreV0{compressor: c}
}

func (d *onDiskStoreV0) Read(gcm cipher.AEAD, reader io.Reader) ([]byte, error) {
	enc, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	b, err := gcm.Open(nil, enc[:gcm.NonceSize()], enc[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	if d.compressor != nil {
		dec, err := d.compressor.Decompress(b)
		if err != nil {
			return nil, err
		}

		b = dec
	}

	return b, nil
}

func (d *onDiskStoreV0) Write(gcm cipher.AEAD, filepath string, data []byte) error {
	nonce := make([]byte, gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	if d.compressor != nil {
		enc, err := d.compressor.Compress(data)
		if err != nil {
			return err
		}

		data = enc
	}

	return os.WriteFile(
		filepath,
		gcm.Seal(nonce, nonce, data, nil),
		0o600,
	)
}
