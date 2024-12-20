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

package git

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/enterprise-contract/go-gather/gather"
	"github.com/enterprise-contract/go-gather/metadata"
)

type GitGatherer struct {
	GitMetadata
	Authenticator SSHAuthenticator
}

type GitMetadata struct {
	Path         string
	CommitHash   string
	Author       string
	Timestamp    string
	LatestCommit string
}

// SSHAuthenticator represents an interface for authenticating SSH connections.
type SSHAuthenticator interface {
	// NewSSHAgentAuth returns a new SSH agent authentication method for the given user.
	// It returns an instance of transport.AuthMethod and an error if any.
	NewSSHAgentAuth(user string) (transport.AuthMethod, error)
}

// RealSSHAuthenticator represents an implementation of the SSHAuthenticator interface.
type RealSSHAuthenticator struct{}

func (g *GitGatherer) Matcher(uri string) bool {
	terms := []string{"git@", "git://", "git::", ".git"}

	for _, term := range terms {
		if strings.Contains(uri, term) {
			return true
		}
	}
	return false
}

func (g *GitGatherer) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	src, ref, subdir, depth, err := processUrl(src)

	if err != nil {
		return nil, fmt.Errorf("failed to process URL: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:      src,
		Progress: nil,
	}

	// if we have a subdir, set no checkout to true and depth to 1
	if subdir != "" {
		cloneOpts.Depth = 1
		cloneOpts.NoCheckout = true
	}

	// If we have a ref and it isn't a hash, set the reference name in the clone options
	if len(ref) > 0 && !plumbing.IsHash(ref) {
		cloneOpts.ReferenceName = plumbing.ReferenceName(ref)
	}

	if depth != "" {
		cloneOpts.Depth, err = strconv.Atoi(depth)
		if err != nil {
			return nil, fmt.Errorf("failed to parse depth: %w", err)
		}
	}

	repo, err := git.PlainCloneContext(ctx, dst, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	if subdir != "" {
		w, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get repository worktree: %w", err)
		}
		err = w.Checkout(&git.CheckoutOptions{
			SparseCheckoutDirectories: []string{subdir},
			Branch:                    plumbing.NewBranchReferenceName(head.Name().Short()),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to checkout repository: %w", err)
		}
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}

	g.Path = dst
	g.CommitHash = commit.Hash.String()
	g.Author = commit.Author.Name
	g.Timestamp = commit.Author.When.Format(time.RFC3339)
	g.LatestCommit = commit.Hash.String()

	return &g.GitMetadata, nil
}

func (g *GitMetadata) Get() interface{} {
	return g
}

// NewSSHAgentAuth returns an AuthMethod that uses the SSH agent for authentication.
// It uses the specified user as the username for authentication.
func (r *RealSSHAuthenticator) NewSSHAgentAuth(user string) (transport.AuthMethod, error) {
	return ssh.NewSSHAgentAuth(user)
}

// extractKeyFromQuery extracts the value of the specified key from the query parameters and extracts a subdir, if present.
func extractKeyFromQuery(q url.Values, key string, subdir *string) string {
	value := q.Get(key)
	if strings.Contains(value, "//") {
		parts := strings.SplitN(value, "//", 2)
		*subdir = parts[1]
		q.Del(key)
		return parts[0]
	}
	q.Del(key)
	return value
}

func processUrl(rawSource string) (src, ref, subdir, depth string, err error) {
	for _, prefix := range []string{"git://", "git::"} {
		rawSource = strings.TrimPrefix(rawSource, prefix)
	}
	src = rawSource

	if !strings.HasPrefix(src, "git@") {
		src = "https://" + src
	}

	parsedUrl, err := giturls.Parse(src)
	if err != nil {
		return src, ref, subdir, depth, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Parse the URL again with the url package to extract the query parameters, etc.
	u, err := url.Parse(parsedUrl.String())
	if err != nil {
		return src, ref, subdir, depth, fmt.Errorf("failed to reparse URL: %w", err)
	}

	// Extract the ref, subdir, and depth from the query parameters
	q := u.Query()
	ref = extractKeyFromQuery(q, "ref", &subdir)
	depth = extractKeyFromQuery(q, "depth", &subdir)
	u.RawQuery = q.Encode()

	// If the path contains "//", split it to get the actual path and subdir
	if strings.Contains(u.Path, "//") {
		parts := strings.SplitN(u.Path, "//", 2)
		u.Path = parts[0]
		subdir = parts[1]
	}

	// If the path does not end with ".git", append it
	if !strings.HasSuffix(u.Path, ".git") {
		u.Path += ".git"
	}

	// Return the URL, ref, subdir, and depth
	return u.String(), ref, subdir, depth, nil
}

func init() {
	gather.RegisterGatherer(&GitGatherer{})
}
