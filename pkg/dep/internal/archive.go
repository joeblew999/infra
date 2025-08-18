package internal

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/log"
)

// ExtractArchive extracts various archive types to destination directory
func ExtractArchive(archivePath, destDir string) error {
	// Determine archive type by file extension
	switch {
	case hasExtension(archivePath, ".zip"):
		return Unzip(archivePath, destDir)
	case hasExtension(archivePath, ".tar.gz", ".tgz"):
		return UntarGz(archivePath, destDir)
	case hasExtension(archivePath, ".tar.bz2", ".tbz2"):
		return UntarBz2(archivePath, destDir)
	case hasExtension(archivePath, ".tar.xz"):
		return fmt.Errorf("tar.xz extraction not yet implemented")
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

// hasExtension checks if file has any of the given extensions
func hasExtension(filename string, extensions ...string) bool {
	for _, ext := range extensions {
		if len(filename) >= len(ext) && filename[len(filename)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// Unzip extracts a zip archive to a destination directory
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", fpath, err)
		}

		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open output file %s: %w", fpath, err)
		}
		defer out.Close()

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip archive: %w", err)
		}
		defer rc.Close()

		_, err = io.Copy(out, rc)
		if err != nil {
			return fmt.Errorf("failed to copy content from zip to file: %w", err)
		}
	}
	return nil
}

// UntarGz extracts a .tar.gz archive to a destination directory
func UntarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		fpath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fpath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), os.FileMode(0755)); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", fpath, err)
			}
			out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to open output file %s: %w", fpath, err)
			}
			defer out.Close()
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("failed to copy content from tar to file: %w", err)
			}
		default:
			log.Warn("Skipping unsupported tar entry type", "type", header.Typeflag, "name", header.Name)
		}
	}
	return nil
}

// UntarBz2 extracts a .tar.bz2 archive to a destination directory
func UntarBz2(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.bz2 file: %w", err)
	}
	defer file.Close()

	br := bzip2.NewReader(file)
	tr := tar.NewReader(br)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		fpath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fpath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), os.FileMode(0755)); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", fpath, err)
			}
			out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to open output file %s: %w", fpath, err)
			}
			defer out.Close()
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("failed to copy content from tar to file: %w", err)
			}
		default:
			log.Warn("Skipping unsupported tar entry type", "type", header.Typeflag, "name", header.Name)
		}
	}
	return nil
}