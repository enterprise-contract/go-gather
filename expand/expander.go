package expand

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
)

/* package expander provides an interface for expanders to implement. Expanders are used to expand compressed files. */

// Known magic numbers for common compressed file formats
var magicNumbers = map[string][]byte{
	"gzip":  {0x1f, 0x8b},
	"zip":   {0x50, 0x4b, 0x03, 0x04},
	"tar":   {0x75, 0x73, 0x74, 0x61, 0x72},
	"bzip2": {0x42, 0x5a, 0x68},
	"xz":    {0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00},
	"7z":    {0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c},
}

type Expander interface {
	Expand(ctx context.Context, source string, destination string, dir bool, umask os.FileMode) error
	Matcher(extension string) bool
}

var expanders []Expander

func GetExpander(extension string) (Expander) {
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

func IsCompressedFile(filename string) (bool, Expander, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes
	header := make([]byte, 10) // maximum length of magic numbers
	_, err = file.Read(header)
	if err != nil {
		return false, nil, fmt.Errorf("could not read file header: %w", err)
	}

	// Check against known magic numbers
	for format, magic := range magicNumbers {
		if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
			switch format {
			case "zip":
				return true, GetExpander("zip"), nil
			case "tar":
				return true, GetExpander("tar"), nil
			case "gzip":
				if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
					return true, GetExpander("tar.gz"), nil
				} else {
					return true, GetExpander("gzip"), nil
				}
			case "bzip2":
				if strings.HasSuffix(filename, ".tar.bz2") || strings.HasSuffix(filename, ".tbz2") {
					return true, GetExpander("tar.bz2"), nil
				} else {
					return true, GetExpander("bzip2"), nil
				}
			default:
				return false, nil, nil
			}
		}
	}
	return false, nil, nil
}
