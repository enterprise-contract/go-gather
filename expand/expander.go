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

package expand

import (
	"bytes"
	"context"
	"fmt"
	"os"
)

/* package expander provides an interface for expanders to implement. Expanders are used to expand compressed files. */

type Expander interface {
	Expand(ctx context.Context, source string, destination string, dir bool, umask os.FileMode) error
	Matcher(extension string) bool
}

var expanders []Expander

func GetExpander(extension string) Expander {
	for _, expander := range expanders {
		if expander.Matcher(extension) {
			return expander
		}
	}
	return nil
}

func RegisterExpander(e Expander) {
	expanders = append(expanders, e)
}

// Known magic numbers for common compressed file formats
var magicNumbers = map[string][]byte{
	"gzip":  {0x1f, 0x8b},
	"zip":   {0x50, 0x4b, 0x03, 0x04},
	"tar":   {0x75, 0x73, 0x74, 0x61, 0x72},
	"bzip2": {0x42, 0x5a, 0x68},
	"xz":    {0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00},
	"7z":    {0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c},
}

func IsCompressedFile(filename string) (bool, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, "", fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes
	header := make([]byte, 10) // maximum length of magic numbers
	_, err = file.Read(header)
	if err != nil {
		return false, "", fmt.Errorf("could not read file header: %w", err)
	}

	// Check against known magic numbers
	for format, magic := range magicNumbers {
		if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
			return true, format, nil
		}
	}
	return false, "", nil
}
