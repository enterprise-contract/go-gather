package tar

import (
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/google/safearchive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enterprise-contract/go-gather/expand"
	"github.com/enterprise-contract/go-gather/internal/helpers"
)

type TarExpander struct {
	FileSizeLimit int64
	FilesLimit    int
}

func (t *TarExpander) Expand(ctx context.Context, src, dst string, dir bool, umask os.FileMode) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %s", src)
	}
	defer input.Close()

	if strings.Contains(src, "tar.gz") || strings.Contains(src, "tgz") {
		if err = extractTarGz(input, dst, dir, umask, t.FileSizeLimit, t.FilesLimit); err != nil {
			return fmt.Errorf("failed to extract tar.gz file: %s", err)
		}
	} else if strings.Contains(src, "tar.bz2") || strings.Contains(src, "tbz2") {
		if err = extractTarBz(input, dst, dir, umask, t.FileSizeLimit, t.FilesLimit); err != nil {
			return fmt.Errorf("failed to extract tar.bz2 file: %s", err)
		}
	} else {
		if err = untar(input, dst, src, dir, umask, t.FileSizeLimit, t.FilesLimit); err != nil {
			return fmt.Errorf("failed to untar file: %s", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to get destination directory size: %s", dst)
	}

	return nil
}

func (t *TarExpander) Matcher(fileName string) bool {
	extensions := []string{"tar", "tgz", "tbz2"}
	for _, ext := range extensions {
		if strings.Contains(fileName, ext) {
			return true
		}
	}
	return false
}

// extractTarBz is a helper function that extracts a tarball compressed with bzip2 to a destination directory
func extractTarBz(input io.Reader, dst string, dir bool, umask os.FileMode, fileSizeLimit int64, filesLimit int) error {
	bzr := bzip2.NewReader(input)
	return untar(bzr, dst, "", dir, umask, fileSizeLimit, filesLimit)
}

// extractTarGz is a helper function that extracts a tarball compressed with gzip to a destination directory
func extractTarGz(input io.Reader, dst string, dir bool, umask os.FileMode, fileSizeLimit int64, filesLimit int) error {
	gzr, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer gzr.Close()

	return untar(gzr, dst, "", dir, umask, fileSizeLimit, filesLimit)
}

// untar is a helper function that untars a tarball to a destination directory
func untar(input io.Reader, dst, src string, dir bool, umask os.FileMode, fileSizeLimit int64, filesLimit int) error {
	tarReader := tar.NewReader(input)
	finished := false

	dirHeaders := []*tar.Header{}
	now := time.Now()

	var (
		fileSize   int64
		filesCount int
	)

	for {
		if filesLimit > 0 {
			filesCount++
			if filesCount > filesLimit {
				return fmt.Errorf("tar file contains more files than the %d allowed: %d", filesCount, filesLimit)
			}
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			if !finished {
				// Empty archive
				return fmt.Errorf("tar file is empty: %s", src)
			}
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeXGlobalHeader || header.Typeflag == tar.TypeXHeader {
			continue
		}

		fPath := dst

		if dir {
			if dir {
				// if helpers.ContainsDotDot(header.Name) {
				// 	return fmt.Errorf("tar file (%s) would escape destination directory", header.Name)
				// }
				fPath = filepath.Join(dst, header.Name) // nolint:gosec
			}
		}

		fileInfo := header.FileInfo()
		fileSize += fileInfo.Size()

		if fileSizeLimit > 0 && fileSize > fileSizeLimit {
			return fmt.Errorf("tar file size exceeds the %d limit: %d", fileSizeLimit, fileSize)
		}

		if fileInfo.IsDir() {
			if !dir {
				return fmt.Errorf("expected a file (%s), got a directory: %s", src, fPath)
			}

			if err := os.MkdirAll(fPath, umask); err != nil {
				return fmt.Errorf("failed to create directory (%s): %s", fPath, err)
			}

			dirHeaders = append(dirHeaders, header)

			continue
		} else {
			destPath := filepath.Dir(fPath)

			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				if err := os.MkdirAll(destPath, umask); err != nil {
					return fmt.Errorf("failed to create directory (%s): %s", destPath, err)
				}
			}
		}

		if !dir && finished {
			return fmt.Errorf("tar file contains more than one file: %s", src)
		}

		finished = true

		err = helpers.CopyReader(tarReader, fPath, umask, fileSizeLimit)
		if err != nil {
			return err
		}

		aTime, mTime := now, now

		if header.AccessTime.Unix() > 0 {
			aTime = header.AccessTime
		}

		if header.ModTime.Unix() > 0 {
			mTime = header.ModTime
		}

		if err := os.Chtimes(fPath, aTime, mTime); err != nil {
			return fmt.Errorf("failed to change file times (%s): %s", fPath, err)
		}
	}

	for _, dirHeader := range dirHeaders {
		path := filepath.Join(dst, dirHeader.Name) // nolint:gosec
		// Chmod the directory
		if err := os.Chmod(path, dirHeader.FileInfo().Mode()); err != nil {
			return fmt.Errorf("failed to change directory permissions (%s): %s", path, err)
		}

		// Set the access and modification times
		aTime, mTime := now, now

		if dirHeader.AccessTime.Unix() > 0 {
			aTime = dirHeader.AccessTime
		}
		if dirHeader.ModTime.Unix() > 0 {
			mTime = dirHeader.ModTime
		}
		if err := os.Chtimes(path, aTime, mTime); err != nil {
			return fmt.Errorf("failed to change directory times (%s): %s", path, err)
		}
	}
	return nil
}

func init() {
	expand.RegisterExpander(&TarExpander{})
}
