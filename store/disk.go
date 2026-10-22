package store

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
	"github.com/pierrec/lz4/v4"
	"github.com/sirupsen/logrus"
)

type onDiskStore struct {
	path     string
	gcm      cipher.AEAD
	sem      *Semaphore
	fallback Fallback
}

func NewCipher(pass []byte) (cipher.AEAD, error) {
	aes, err := aes.NewCipher(hash(pass))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func NewOnDiskStore(path string, pass []byte, opt ...Option) (Store, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	gcm, err := NewCipher(pass)
	if err != nil {
		return nil, err
	}

	store := &onDiskStore{
		path: path,
		gcm:  gcm,
	}

	for _, opt := range opt {
		opt.config(store)
	}

	return store, nil
}

const blockSize = 64 * 4096
const (
	// legacyStoreVersion files seal every block with the same nonce and are only read.
	legacyStoreVersion = uint32(1)
	storeVersion       = uint32(2)
)

func (c *onDiskStore) Set(messageID imap.InternalMessageID, in io.Reader) error {
	if err := os.MkdirAll(c.path, 0o700); err != nil {
		return err
	}

	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	fullPath := filepath.Join(c.path, messageID.String())

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer file.Close()

	if written, err := file.Write(storeHeaderBytes); err != nil {
		return err
	} else if written != len(storeHeaderBytes) {
		return fmt.Errorf("failed to write store header to file")
	}

	reader, writer := io.Pipe()
	defer writer.Close()

	compressor := lz4.NewWriter(writer)
	if err := compressor.Apply(lz4.BlockSizeOption(lz4.Block64Kb), lz4.ChecksumOption(false)); err != nil {
		return fmt.Errorf("failed to set compressor options: %w", err)
	}

	go func() {
		_, err := compressor.ReadFrom(in)
		compressor.Close()
		writer.CloseWithError(err)
	}()

	encryptionOverhead := c.gcm.Overhead()
	encryptedBlockSized := getEncryptedBlockSize(c.gcm, blockSize)

	compressedBlock := make([]byte, blockSize)
	encryptedBlock := make([]byte, encryptedBlockSized)

	// Write nonce to file.
	if bytesWritten, err := file.Write(nonce); err != nil || bytesWritten != len(nonce) {
		return fmt.Errorf("failed to write nonce to file: %w", err)
	}

	blockNonceBuf := make([]byte, len(nonce))

	// Write encrypted blocks.
	for blockIndex := uint64(0); ; blockIndex++ {
		// Read at least blockSize from the compressor.
		bytesRead, err := io.ReadAtLeast(reader, compressedBlock, blockSize)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			// If there less bytes than expected we reached the end of the file.
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				return err
			}
		}

		// Encrypt the compressed block.
		encryptedBlock = encryptedBlock[:0] // Reset slice.
		encrypted := c.gcm.Seal(encryptedBlock, blockNonce(blockNonceBuf, nonce, blockIndex), compressedBlock[0:bytesRead], nil)
		encryptedLen := bytesRead + encryptionOverhead

		// Write to disk.
		if bytesWritten, err := file.Write(encrypted[0:encryptedLen]); err != nil || bytesWritten != encryptedLen {
			return fmt.Errorf("failed to write block to disk: %w", err)
		}
	}

	return nil
}

func (c *onDiskStore) Get(messageID imap.InternalMessageID) ([]byte, error) {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	file, err := os.Open(filepath.Join(c.path, messageID.String()))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fileSize int64

	header := make([]byte, len(storeHeaderBytes))
	if _, err := io.ReadFull(file, header); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) && c.fallback != nil {
			result, err := c.readFromFallback(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read from fallback: %w", err)
			}

			return result, nil
		}

		return nil, err
	}

	perBlockNonce := bytes.Equal(header, storeHeaderBytes)

	if !perBlockNonce && !bytes.Equal(header, legacyStoreHeaderBytes) {
		if c.fallback != nil {
			result, err := c.readFromFallback(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read from fallback: %w", err)
			}

			return result, nil
		}

		return nil, fmt.Errorf("file is not a valid store file")
	}

	if stat, err := file.Stat(); err == nil {
		fileSize = stat.Size() - int64(len(storeHeaderBytes))
	}

	nonce := make([]byte, c.gcm.NonceSize())

	// Read nonce from file.
	if _, err := io.ReadFull(file, nonce); err != nil {
		return nil, fmt.Errorf("failed to read nonce: %w", err)
	}

	reader, writer := io.Pipe()

	encryptionOverhead := c.gcm.Overhead()
	encryptedBlockSize := getEncryptedBlockSize(c.gcm, blockSize)

	go func() {
		defer writer.Close()

		readBuffer := make([]byte, encryptedBlockSize)
		decryptBuffer := make([]byte, blockSize)
		blockNonceBuf := make([]byte, len(nonce))
		totalBytesRead := 0

		for blockIndex := uint64(0); ; blockIndex++ {
			// Read up to encryptedBlockSize bytes from disk.
			bytesRead, err := file.Read(readBuffer)
			if err != nil {
				if !errors.Is(err, io.EOF) && bytesRead != 0 {
					writer.CloseWithError(err)
				}

				return
			}

			// Decrypt read bytes.
			decryptBuffer = decryptBuffer[:0] // Reset slice.

			readNonce := nonce
			if perBlockNonce {
				readNonce = blockNonce(blockNonceBuf, nonce, blockIndex)
			}

			decrypted, err := c.gcm.Open(decryptBuffer, readNonce, readBuffer[0:bytesRead], nil)
			if err != nil {
				writer.CloseWithError(fmt.Errorf("failed to decrypt block (offset:%v): %w", totalBytesRead, err))
				return
			}

			decryptedLen := bytesRead - encryptionOverhead

			// Write to pipe so they can be decompressed.
			if _, err := writer.Write(decrypted[0:decryptedLen]); err != nil {
				writer.CloseWithError(err)
				return
			}

			totalBytesRead += bytesRead
		}
	}()

	decompressor := lz4.NewReader(reader)

	var b bytes.Buffer

	b.Grow(int(fileSize))

	if _, err := decompressor.WriteTo(&b); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

func (c *onDiskStore) Delete(messageIDs ...imap.InternalMessageID) error {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	for _, messageID := range messageIDs {
		if err := os.Remove(filepath.Join(c.path, messageID.String())); err != nil {
			return err
		}
	}

	return nil
}

func (c *onDiskStore) List() ([]imap.InternalMessageID, error) {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	var ids []imap.InternalMessageID

	if err := filepath.Walk(c.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		id, err := imap.InternalMessageIDFromString(info.Name())
		if err != nil {
			logrus.WithError(err).Errorf("Invalid id file in cache: %v", info.Name())
		}

		ids = append(ids, id)

		return nil
	}); err != nil {
		return nil, err
	}

	return ids, nil
}

func (c *onDiskStore) Close() error {
	return nil
}

type OnDiskStoreBuilder struct{}

func (*OnDiskStoreBuilder) New(path, userID string, passphrase []byte) (Store, error) {
	storePath := filepath.Join(path, userID)

	return NewOnDiskStore(storePath, passphrase)
}

func (*OnDiskStoreBuilder) Delete(path, userID string) error {
	storePath := filepath.Join(path, userID)

	return os.RemoveAll(storePath)
}

func (c *onDiskStore) readFromFallback(file *os.File) ([]byte, error) {
	if pos, err := file.Seek(0, 0); err != nil {
		return nil, err
	} else if pos != 0 {
		return nil, fmt.Errorf("failed to rewind file to start")
	}

	return c.fallback.Read(c.gcm, file)
}

func getEncryptedBlockSize(aead cipher.AEAD, blockSize int) int {
	return blockSize + aead.Overhead()
}

// blockNonce writes the nonce for the given block into dst: the file nonce with the block index
// XORed into its last 8 bytes, so that no two blocks of a file are sealed under the same nonce.
func blockNonce(dst, nonce []byte, blockIndex uint64) []byte {
	copy(dst, nonce)

	var counter [8]byte

	binary.BigEndian.PutUint64(counter[:], blockIndex)

	tail := dst[len(dst)-len(counter):]
	for i := range counter {
		tail[i] ^= counter[i]
	}

	return dst
}

func makeGluonHeaderBytes(headerVersion uint32) []byte {
	const StoreHeaderID = "GLUON-CACHE"

	version := make([]byte, 4)

	binary.LittleEndian.PutUint32(version, headerVersion)

	return append([]byte(StoreHeaderID), version...)
}

var (
	storeHeaderBytes       = makeGluonHeaderBytes(storeVersion)
	legacyStoreHeaderBytes = makeGluonHeaderBytes(legacyStoreVersion)
)
