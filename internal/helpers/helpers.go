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

package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsSafePath returns a boolean indicating whether the filePath is within dst,
// along with an error if not.
func IsSafePath(filePath, dst string) (bool, error) {
	// Convert dst to an absolute path
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return false, fmt.Errorf("failed to resolve absolute destination path: %v", err)
	}
	// Ensure dst ends with a path separator to match only subdirectories
	absDst = filepath.Clean(absDst) + string(os.PathSeparator)

	// Convert filePath to an absolute path
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve absolute file path: %v", err)
	}

	// Resolve any symlinks in absFilePath for additional security
	resolvedFilePath, err := filepath.EvalSymlinks(absFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	// Check if resolvedFilePath is within absDst
	if !strings.HasPrefix(resolvedFilePath, absDst) {
		return false, fmt.Errorf("illegal file path: %s", filePath)
	}

	return true, nil
}

// copyReader copies a reader to a file. If fileSizeLimit is greater than 0, it will limit the size of the file.
func CopyReader(src io.Reader, dst string, mode os.FileMode, fileSizeLimit int64) error {
	dstF, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", dst, err)
	}
	defer dstF.Close()

	if fileSizeLimit > 0 {
		src = io.LimitReader(src, fileSizeLimit)
	}

	_, err = io.Copy(dstF, src)
	if err != nil {
		return fmt.Errorf("failed to copy file %s: %w", dst, err)
	}

	return os.Chmod(dst, mode)
}

func GetDirectorySize(dir string) (int64, error) {
	var size int64
	dir, err := ExpandTilde(dir)
	if err != nil {
		return 0, fmt.Errorf("failed to expand directory path: %w", err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}
	return size, nil
}

func ExpandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	return filepath.Join(homeDir, path[1:]), nil
}

// CopyDir copies the contents of the src directory to dst directory
func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source directory info: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	_, err = os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dst, srcInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// containsDotDot checks if the filepath value v contains a ".." entry.
// This will check filepath components by splitting along / or \. This
// function is copied directly from the Go net/http implementation.
func ContainsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlash) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlash(r rune) bool { return r == '/' || r == '\\' }
