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

package file

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/enterprise-contract/go-gather/expand"
	"github.com/enterprise-contract/go-gather/gather"
	"github.com/enterprise-contract/go-gather/internal/helpers"
	"github.com/enterprise-contract/go-gather/metadata"
)

type FileGatherer struct{
	FSMetadata
}

type FSMetadata struct {
	Path      string
	Size      int64
	Timestamp string
}

type FileSaver struct {
	FSMetadata
}

func (f *FileGatherer) Matcher(uri string) bool {
	prefixes := []string{"file://", "file::", "/", "./", "../"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}
	return false
}

func (f *FileGatherer) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	src = strings.TrimPrefix(src, "file::")

	parsedSrc, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}

	src, err = helpers.ExpandTilde(parsedSrc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand source path: %w", err)
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %w", err)
	}

	if helpers.IsDir(src) {
		if err := helpers.CopyDir(src, dst); err != nil {
			return nil, fmt.Errorf("failed to copy directory: %w", err)
		}
		dirSize, err := helpers.GetDirectorySize(dst)
		if err != nil {
			return nil, err
		}

		return &FSMetadata{
			Path: 	dst,
			Size: 	dirSize,
			Timestamp: time.Now().String(),
		}, nil
	}

	if !helpers.IsDir(src) {
		if ok, format, err := expand.IsCompressedFile(src); ok && err == nil {
			var e expand.Expander
			switch format {
			case "zip":
				e = expand.GetExpander("zip")
			case "tar":
				e = expand.GetExpander("tar")
			case "gzip":
				if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
					e = expand.GetExpander("tar.gz")
				} else {
					e = expand.GetExpander("gzip")
				}
			case "bzip2":
				if strings.HasSuffix(src, ".tar.bz2") || strings.HasSuffix(src, ".tbz2") {
					e = expand.GetExpander("tar.bz2")
				} else {
					e = expand.GetExpander("bzip2")
				}
			default:
				return nil, fmt.Errorf("compressed file found, but no expander available")
			}

			err := e.Expand(ctx, src, dst, true, 0755)
			if err != nil {
				return nil, err
			}
			dirSize, err := helpers.GetDirectorySize(dst)
			if err != nil {
				return nil, err
			}
	
			fm := &FSMetadata{
				Path:      dst,
				Size:      dirSize,
				Timestamp: time.Now().String(),
			}
			return fm, err
		} else if err != nil {
			return nil, fmt.Errorf("failed to determine if source is a compressed file: %w", err)
		}
	}

	// TODO: Figure out how to make this flexible for more saver types?
	fsaver := FileSaver{}
	return fsaver.save(ctx, src, dst, false)
}

func (f *FSMetadata) Get() interface{} {
	return f
}

// save copies from a filesystem source to a filesystem destination. 
// If append is true, the file will be appended to the destination.
func (f *FileSaver) save(ctx context.Context, source string, destination string, append bool) (metadata.Metadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var dstFile *os.File
	var err error

	src, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}

	if _, err := os.Stat(src.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %w", err)
	}

	dst, err := url.Parse(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to parse destination URI: %w", err)
	}

	if append {
		dstFile, err = os.OpenFile(dst.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	} else {
		dstFile, err = os.Create(dst.Path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}

	return &FSMetadata{
		Path:      dst.Path,
		Size:      f.Size,
		Timestamp: f.Timestamp,
	}, nil
}

func init() {
	gather.RegisterGatherer(&FileGatherer{})
}