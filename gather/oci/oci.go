package oci

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/enterprise-contract/go-gather/gather"
	r "github.com/enterprise-contract/go-gather/internal/oci/registry"
	"github.com/enterprise-contract/go-gather/metadata"
)

type OCIGatherer struct{}

type OCIMetadata struct {
	Path      string
	Digest    string
	Timestamp string
}

func (o *OCIMetadata) Get() interface{} {
	return o
}

func (o *OCIMetadata) GetDigest() string {
	return o.Digest
}

var Transport http.RoundTripper = http.DefaultTransport

var orasCopy = oras.Copy

func (o *OCIGatherer) Gather(ctx context.Context, source, dst string) (metadata.Metadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.Contains(source, "localhost") {
		source = strings.ReplaceAll(source, "localhost", "127.0.0.1")
	}

	// Parse the source URI
	repo := ociURLParse(source)

	// Get the artifact reference
	ref, err := registry.ParseReference(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// If the reference is empty, set it to "latest"
	if ref.Reference == "" {
		ref.Reference = "latest"
		repo = ref.String()
	}

	// Create the repository client
	src, err := remote.NewRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository client: %w", err)
	}

	// Setup the client for the repository
	if err := r.SetupClient(src, Transport); err != nil {
		return nil, fmt.Errorf("failed to setup repository client: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file store
	fileStore, err := file.New(dst)
	if err != nil {
		return nil, fmt.Errorf("file store: %w", err)
	}
	defer fileStore.Close()

	// Copy the artifact to the file store
	a, err := orasCopy(ctx, src, repo, fileStore, "", oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("pulling policy: %w", err)
	}

	// Simulate metadata gathering for OCI image
	return &OCIMetadata{
		Path:      dst,
		Digest:    a.Digest.String(),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (o *OCIGatherer) Matcher(uri string) bool {
	prefixes := []string{"oci://", "oci::"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}
	return false
}

func init() {
	gather.RegisterGatherer(&OCIGatherer{})
}

func ociURLParse(source string) string {
	if strings.Contains(source, "::") {
		source = strings.Split(source, "::")[1]
	}

	scheme, src, found := strings.Cut(source, "://")
	if !found {
		src = scheme
	}
	return src
}
