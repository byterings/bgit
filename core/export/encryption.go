package export

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/byterings/bgit/core/models"
	"golang.org/x/crypto/argon2"
)

const (
	envelopeMagic           = "BGITEX10"
	argonTimeCost    uint32 = 3
	argonMemoryKiB   uint32 = 64 * 1024
	argonKeyLength   uint32 = 32
	argonParallelism uint8  = 4
	saltLength              = 16
)

func encryptArchive(archiveBytes []byte, toolVersion string, createdAt time.Time, password string) (*models.ExportEnvelopeHeader, []byte, error) {
	salt, err := randomBytes(saltLength)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argonTimeCost, argonMemoryKiB, argonParallelism, argonKeyLength)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AEAD: %w", err)
	}

	nonce, err := randomBytes(gcm.NonceSize())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	header := &models.ExportEnvelopeHeader{
		FormatVersion: ArchiveFormatV1,
		BgitVersion:   toolVersion,
		CreatedAt:     createdAt.Format(time.RFC3339),
		Payload: models.ExportPayloadDescriptor{
			LayoutVersion: ArchiveLayoutV1,
			Compression:   "tar+gzip",
		},
		Encryption: models.ExportEncryptionHeader{
			KDF: models.ExportKDFHeader{
				Algorithm:   "Argon2id",
				SaltHex:     hex.EncodeToString(salt),
				TimeCost:    argonTimeCost,
				MemoryKiB:   argonMemoryKiB,
				Parallelism: argonParallelism,
				KeyLength:   argonKeyLength,
			},
			Cipher: models.ExportCipherHeader{
				Algorithm: "AES-256-GCM",
				NonceHex:  hex.EncodeToString(nonce),
			},
		},
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode export header: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, archiveBytes, headerBytes)
	return header, ciphertext, nil
}

func decryptArchiveFile(path string, password string) ([]byte, *models.ExportEnvelopeHeader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read archive: %w", err)
	}

	header, ciphertext, headerBytes, err := parseEnvelope(data)
	if err != nil {
		return nil, nil, err
	}

	if header.FormatVersion != ArchiveFormatV1 {
		return nil, nil, fmt.Errorf("unsupported export format version %q", header.FormatVersion)
	}
	if header.Payload.LayoutVersion != ArchiveLayoutV1 {
		return nil, nil, fmt.Errorf("unsupported payload layout version %q", header.Payload.LayoutVersion)
	}
	if header.Payload.Compression != "tar+gzip" {
		return nil, nil, fmt.Errorf("unsupported payload compression %q", header.Payload.Compression)
	}
	if header.Encryption.KDF.Algorithm != "Argon2id" {
		return nil, nil, fmt.Errorf("unsupported KDF %q", header.Encryption.KDF.Algorithm)
	}
	if header.Encryption.Cipher.Algorithm != "AES-256-GCM" {
		return nil, nil, fmt.Errorf("unsupported cipher %q", header.Encryption.Cipher.Algorithm)
	}

	salt, err := hex.DecodeString(header.Encryption.KDF.SaltHex)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid archive salt: %w", err)
	}
	nonce, err := hex.DecodeString(header.Encryption.Cipher.NonceHex)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid archive nonce: %w", err)
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		header.Encryption.KDF.TimeCost,
		header.Encryption.KDF.MemoryKiB,
		header.Encryption.KDF.Parallelism,
		header.Encryption.KDF.KeyLength,
	)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AEAD: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, nil, fmt.Errorf("invalid archive nonce length")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, headerBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt archive: invalid password or corrupted archive")
	}

	return plaintext, header, nil
}

func parseEnvelope(data []byte) (*models.ExportEnvelopeHeader, []byte, []byte, error) {
	if len(data) < len(envelopeMagic)+4 {
		return nil, nil, nil, fmt.Errorf("invalid bgit archive: file is too small")
	}
	if string(data[:len(envelopeMagic)]) != envelopeMagic {
		return nil, nil, nil, fmt.Errorf("invalid bgit archive: unsupported file header")
	}

	headerLen := binary.BigEndian.Uint32(data[len(envelopeMagic) : len(envelopeMagic)+4])
	headerStart := len(envelopeMagic) + 4
	headerEnd := headerStart + int(headerLen)
	if headerLen == 0 || headerEnd > len(data) {
		return nil, nil, nil, fmt.Errorf("invalid bgit archive: bad header length")
	}

	headerBytes := data[headerStart:headerEnd]
	var header models.ExportEnvelopeHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode archive header: %w", err)
	}

	ciphertext := data[headerEnd:]
	if len(ciphertext) == 0 {
		return nil, nil, nil, fmt.Errorf("invalid bgit archive: missing encrypted payload")
	}

	return &header, ciphertext, headerBytes, nil
}

func writeEncryptedArchive(path string, header *models.ExportEnvelopeHeader, ciphertext []byte) error {
	if header == nil {
		return fmt.Errorf("export header is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("failed to encode export header: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(path), "bgit-export-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp export archive: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	if err := tempFile.Chmod(0600); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to set temp archive permissions: %w", err)
	}

	if _, err := tempFile.Write([]byte(envelopeMagic)); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write export header magic: %w", err)
	}
	if err := writeUint32(tempFile, uint32(len(headerBytes))); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write export header length: %w", err)
	}
	if _, err := tempFile.Write(headerBytes); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write export header: %w", err)
	}
	if _, err := tempFile.Write(ciphertext); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write encrypted payload: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to sync export archive: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close export archive: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to move export archive into place: %w", err)
	}
	return nil
}

func randomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeUint32(w io.Writer, value uint32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], value)
	_, err := w.Write(buf[:])
	return err
}
