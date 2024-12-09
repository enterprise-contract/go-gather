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

package zip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/enterprise-contract/go-gather/expand"
	"github.com/enterprise-contract/go-gather/internal/helpers"
)

func TestZipMethod(){
	fmt.Println("Testing Zip Method")
}

// ZipExpander provides functionality to extract ZIP archives.
type ZipExpander struct {
	FileSizeLimit int64
	FilesLimit    int
}

// Expand extracts a ZIP file to the specified destination directory.
// It handles tilde expansion, enforces file size limits, and ensures secure extraction.
func (z *ZipExpander) Expand(ctx context.Context, src, dst string, dir bool, umask os.FileMode) error {
	// Expand tilde in source and destination paths
	var err error
	if src, err = helpers.ExpandTilde(src); err != nil {
		return fmt.Errorf("failed to expand source path: %w", err)
	}
	if dst, err = helpers.ExpandTilde(dst); err != nil {
		return fmt.Errorf("failed to expand destination path: %w", err)
	}

	// Open the ZIP archive
	archive, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file %q: %w", src, err)
	}
	defer archive.Close()

	// Prepare a buffer for copying file contents
	const bufferSize = 32 * 1024 // 32 KB
	buffer := make([]byte, bufferSize)

	// Iterate over files in the archive
	for _, f := range archive.File {
		// Enforce file size limit if set
		if z.FileSizeLimit > 0 && f.FileInfo().Size() > z.FileSizeLimit {
			return fmt.Errorf("file %q exceeds size limit of %d bytes", f.Name, z.FileSizeLimit)
		}

		// Construct full file path
		filePath := filepath.Join(dst, f.Name)

		// Prevent Zip Slip vulnerability
		if ok, err := helpers.IsSafePath(filePath, dst); !ok {
			return fmt.Errorf("zip file (%s) would escape destination directory: %w", f.Name, err)
		}

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		// Handle directories
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, umask); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", filePath, err)
			}
			continue
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), umask); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", filepath.Dir(filePath), err)
		}

		// Extract the file
		if err := z.extractFile(f, filePath, buffer); err != nil {
			return err
		}
	}

	return nil
}

// extractFile handles the extraction of a single file from the ZIP archive.
func (z *ZipExpander) extractFile(f *zip.File, filePath string, buffer []byte) error {
	// Open the source file within the archive
	srcFile, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", f.Name, err)
	}
	defer srcFile.Close()

	// Open the destination file
	dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", filePath, err)
	}
	defer dstFile.Close()

	// Copy contents using a buffer
	if _, err := io.CopyBuffer(dstFile, srcFile, buffer); err != nil {
		return fmt.Errorf("failed to copy file %q: %w", f.Name, err)
	}

	return nil
}

// Matcher checks if the extension matches supported formats.
func (z *ZipExpander) Matcher(extension string) bool {
	return strings.Contains(extension, ".zip")
}

func init() {
	expand.RegisterExpander(&ZipExpander{})
}
