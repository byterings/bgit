package export

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/byterings/bgit/core/models"
)

func buildArchiveBytes(manifest models.ExportManifest, configData []byte, keyFiles []archiveKeyFile, createdAt time.Time) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := writeManifestEntry(tarWriter, manifest, createdAt); err != nil {
		return nil, err
	}
	if err := writeDirEntry(tarWriter, PayloadDir, createdAt); err != nil {
		return nil, err
	}
	if err := writeDirEntry(tarWriter, PayloadConfigDir, createdAt); err != nil {
		return nil, err
	}
	if err := writeFileEntry(tarWriter, PayloadConfigPath, configData, 0600, createdAt); err != nil {
		return nil, err
	}
	if err := writeDirEntry(tarWriter, PayloadKeysDir, createdAt); err != nil {
		return nil, err
	}
	for _, keyFile := range keyFiles {
		if err := writeFileEntry(tarWriter, keyFile.Path, keyFile.Data, keyFile.Mode, createdAt); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize export archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize export compression: %w", err)
	}

	return buf.Bytes(), nil
}

func writeManifestEntry(tw *tar.Writer, manifest models.ExportManifest, createdAt time.Time) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode export manifest: %w", err)
	}
	data = append(data, '\n')
	return writeFileEntry(tw, ManifestPath, data, 0600, createdAt)
}

func writeDirEntry(tw *tar.Writer, path string, createdAt time.Time) error {
	header := &tar.Header{
		Name:     path + "/",
		Typeflag: tar.TypeDir,
		Mode:     0700,
		ModTime:  createdAt,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write archive directory %s: %w", path, err)
	}
	return nil
}

func writeFileEntry(tw *tar.Writer, path string, data []byte, mode int64, createdAt time.Time) error {
	header := &tar.Header{
		Name:    path,
		Mode:    mode,
		Size:    int64(len(data)),
		ModTime: createdAt,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write archive header %s: %w", path, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("failed to write archive file %s: %w", path, err)
	}
	return nil
}

type payloadContents struct {
	ConfigData []byte
	Keys       map[string]payloadKeyPair
}

type payloadKeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

func readPayloadContents(archiveBytes []byte) (*payloadContents, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read archive compression: %w", err)
	}
	defer gzipReader.Close()

	contents := &payloadContents{
		Keys: make(map[string]payloadKeyPair),
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read archive entry: %w", err)
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		switch {
		case header.Name == PayloadConfigPath:
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read archived config: %w", err)
			}
			contents.ConfigData = data
		case strings.HasPrefix(header.Name, PayloadKeysDir+"/"):
			if err := readPayloadKeyEntry(tarReader, contents, header.Name); err != nil {
				return nil, err
			}
		}
	}

	if len(contents.ConfigData) == 0 {
		return nil, fmt.Errorf("archive is missing %s", PayloadConfigPath)
	}
	return contents, nil
}

func readPayloadKeyEntry(r io.Reader, contents *payloadContents, path string) error {
	name := strings.TrimPrefix(path, PayloadKeysDir+"/")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("invalid archived key path %q", path)
	}

	isPublic := strings.HasSuffix(name, ".pub")
	alias := strings.TrimSuffix(name, ".pub")
	if err := validateKeyAlias(alias); err != nil {
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read archived key %s: %w", path, err)
	}

	pair := contents.Keys[alias]
	if isPublic {
		pair.PublicKey = data
	} else {
		pair.PrivateKey = data
	}
	contents.Keys[alias] = pair
	return nil
}
