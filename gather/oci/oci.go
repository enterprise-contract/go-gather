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

type OCIGatherer struct {
	OCIMetadata
}

type OCIMetadata struct {
	Path      string
	Digest    string
	Timestamp string
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

	o.Digest = a.Digest.String()
	o.Path = dst
	o.Timestamp = time.Now().Format(time.RFC3339)

	return &o.OCIMetadata, nil
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

func (o *OCIMetadata) Get() interface{} {
	return o
}

func (o *OCIMetadata) GetDigest() string {
	return o.Digest
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

func init() {
	gather.RegisterGatherer(&OCIGatherer{})
}
