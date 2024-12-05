package bzip2

import (
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/enterprise-contract/go-gather/expand"
	"github.com/enterprise-contract/go-gather/internal/helpers"
)

type Bzip2Expander struct {
	FileSizeLimit int64
}

func (b *Bzip2Expander) Expand(ctx context.Context, src, dst string, dir bool, umask os.FileMode) error {
	var err error
	if src, err = helpers.ExpandTilde(src); err != nil {
		return fmt.Errorf("failed to expand source path: %w", err)
	}
	if dst, err = helpers.ExpandTilde(dst); err != nil {
		return fmt.Errorf("failed to expand destination path: %w", err)
	}

	//Bzip2 doesn't support directories
	if dir {
		return fmt.Errorf("bzip2 compression can only extract to a single file")
	}

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open bzip2 file %q: %w", src, err)
	}
	defer file.Close()

	bzipReader := bzip2.NewReader(file)

	outputFile := dst
	if filepath.Base(dst) == "" {
		outputFile = filepath.Join(dst, filepath.Base(src))
	}

	if ok := helpers.ContainsDotDot(outputFile); ok {
		return fmt.Errorf("bzip2 file would escape destination directory")
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), umask); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", filepath.Dir(outputFile), err)
	}

	outFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", outputFile, err)
	}
	defer outFile.Close()

	const bufferSize = 32 * 1024 // 32 KB
	buffer := make([]byte, bufferSize)
	if _, err := io.CopyBuffer(outFile, bzipReader, buffer); err != nil {
		return fmt.Errorf("failed to decompress file %q: %w", src, err)
	}

	return nil
}

// Matcher checks if the extension matches supported formats.
func (b *Bzip2Expander) Matcher(extension string) bool {
	return (strings.Contains(extension, "bz2") || strings.Contains(extension, "bzip2")) && !strings.Contains(extension, "tar")
}

func init() {
	expand.RegisterExpander(&Bzip2Expander{})
}
