package export

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/byterings/bgit/core/models"
)

func writeArchive(path string, manifest models.ExportManifest, configData []byte, createdAt time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
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

	gzipWriter := gzip.NewWriter(tempFile)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := writeManifestEntry(tarWriter, manifest, createdAt); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tempFile.Close()
		return err
	}
	if err := writeDirEntry(tarWriter, PayloadDir, createdAt); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tempFile.Close()
		return err
	}
	if err := writeDirEntry(tarWriter, PayloadConfigDir, createdAt); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tempFile.Close()
		return err
	}
	if err := writeFileEntry(tarWriter, PayloadConfigPath, configData, 0600, createdAt); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tempFile.Close()
		return err
	}
	if err := writeDirEntry(tarWriter, PayloadKeysDir, createdAt); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tempFile.Close()
		return err
	}

	if err := tarWriter.Close(); err != nil {
		gzipWriter.Close()
		tempFile.Close()
		return fmt.Errorf("failed to finalize export archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to finalize export compression: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close export archive: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to move export archive into place: %w", err)
	}

	return nil
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
