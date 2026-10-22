package store

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/pierrec/lz4/v4"
	"github.com/stretchr/testify/require"
)

func TestOnDiskStore_BlocksDoNotShareNonce(t *testing.T) {
	dir := t.TempDir()

	s, err := NewOnDiskStore(dir, []byte("pass"))
	require.NoError(t, err)

	data := make([]byte, 3*blockSize)
	_, err = rand.Read(data)
	require.NoError(t, err)

	id := imap.NewInternalMessageID()
	require.NoError(t, s.Set(id, bytes.NewReader(data)))

	raw, err := os.ReadFile(filepath.Join(dir, id.String()))
	require.NoError(t, err)

	gcm := s.(*onDiskStore).gcm //nolint:forcetypeassert
	nonce := raw[len(storeHeaderBytes) : len(storeHeaderBytes)+gcm.NonceSize()]
	blocks := raw[len(storeHeaderBytes)+gcm.NonceSize():]
	encryptedBlockSize := getEncryptedBlockSize(gcm, blockSize)

	require.Greater(t, len(blocks), encryptedBlockSize)

	// The second block must not authenticate under the nonce stored in the file header.
	_, err = gcm.Open(nil, nonce, blocks[encryptedBlockSize:min(2*encryptedBlockSize, len(blocks))], nil)
	require.Error(t, err)

	read, err := s.Get(id)
	require.NoError(t, err)
	require.Equal(t, data, read)
}

func TestOnDiskStore_ReadsLegacyVersion(t *testing.T) {
	dir := t.TempDir()

	s, err := NewOnDiskStore(dir, []byte("pass"))
	require.NoError(t, err)

	gcm := s.(*onDiskStore).gcm //nolint:forcetypeassert

	data := make([]byte, 3*blockSize)
	_, err = rand.Read(data)
	require.NoError(t, err)

	var compressed bytes.Buffer

	compressor := lz4.NewWriter(&compressed)
	require.NoError(t, compressor.Apply(lz4.BlockSizeOption(lz4.Block64Kb), lz4.ChecksumOption(false)))
	_, err = io.Copy(compressor, bytes.NewReader(data))
	require.NoError(t, err)
	require.NoError(t, compressor.Close())

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	require.NoError(t, err)

	// Version 1 files seal every block with the same nonce.
	file := append(append([]byte{}, legacyStoreHeaderBytes...), nonce...)

	for plain := compressed.Bytes(); len(plain) > 0; {
		n := min(blockSize, len(plain))
		file = gcm.Seal(file, nonce, plain[:n], nil)
		plain = plain[n:]
	}

	id := imap.NewInternalMessageID()
	require.NoError(t, os.WriteFile(filepath.Join(dir, id.String()), file, 0o600))

	read, err := s.Get(id)
	require.NoError(t, err)
	require.Equal(t, data, read)
}
