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

type FileGatherer struct{}

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

	// Remove the file:: prefix if it exists
	src = strings.TrimPrefix(src, "file::")

	// Parse the source URI
	parsedSrc, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}

	// Expand the source path
	src, err = helpers.ExpandTilde(parsedSrc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand source path: %w", err)
	}

	// Check if the source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %w", err)
	}

	// Check if the source is a directory, if so we copy the entire
	// directory to the destination and return metadata.
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

	// Check if our source is a compressed file.
	if !helpers.IsDir(src) {
		if ok, e, err := expand.IsCompressedFile(src); ok && err == nil {
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

	// Todo: Figure out how to make this flexible for more savers
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

	// Check that the src file exists
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

	// Create a reader for the source file
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