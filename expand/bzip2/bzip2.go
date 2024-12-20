// Copyright The Enterprise Contract Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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

	// Bzip2 doesn't support directories
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

	// Track total decompressed size to avoid decompression bombs.
	var totalBytes int64
	for {
		n, err := bzipReader.Read(buffer)
		if n > 0 {
			if totalBytes+int64(n) > b.FileSizeLimit {
				return fmt.Errorf("decompressed file exceeds size limit of %d bytes", b.FileSizeLimit)
			}
			if _, writeErr := outFile.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write decompressed data: %w", writeErr)
			}
			totalBytes += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error during decompression: %w", err)
		}
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
