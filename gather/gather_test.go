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

package gather

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/enterprise-contract/go-gather/metadata"
)

type TestGatherer struct{}

func (t *TestGatherer) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	return nil, nil
}
func (t *TestGatherer) Matcher(uri string) bool {
	return strings.HasPrefix(uri, "test://")
}

type TestGathererA struct{}

func (t *TestGathererA) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	return nil, nil
}
func (t *TestGathererA) Matcher(uri string) bool {
	return strings.HasPrefix(uri, "testA://")
}

type TestGathererB struct{}

func (t *TestGathererB) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	return nil, nil
}
func (t *TestGathererB) Matcher(uri string) bool {
	return strings.HasPrefix(uri, "testB://")
}

func TestRegisterGatherer(t *testing.T) {
	scheme := "test://"
	RegisterGatherer(&TestGatherer{})

	gatherer, err := GetGatherer(scheme)
	if err != nil {
		t.Fatalf("expected gatherer to be registered, got error: %v", err)
	}

	if _, ok := gatherer.(*TestGatherer); !ok {
		t.Fatalf("expected gatherer of type *TestGatherer, got %T", gatherer)
	}
}

func TestRegisterMultipleGatherers(t *testing.T) {
	// Register multiple gatherers
	RegisterGatherer(&TestGathererA{})
	RegisterGatherer(&TestGathererB{})

	// Retrieve and validate each gatherer
	gathererA, err := GetGatherer("testA://")
	if err != nil {
		t.Fatalf("expected gathererA to be registered, got error: %v", err)
	}
	if _, ok := gathererA.(*TestGathererA); !ok {
		t.Fatalf("expected gatherer of type *TestGathererA, got %T", gathererA)
	}

	gathererB, err := GetGatherer("testB://")
	if err != nil {
		t.Fatalf("expected gathererB to be registered, got error: %v", err)
	}
	if _, ok := gathererB.(*TestGathererB); !ok {
		t.Fatalf("expected gatherer of type *TestGathererB, got %T", gathererB)
	}
}

func TestGetGatherer(t *testing.T) {
	RegisterGatherer(&TestGatherer{})
	gatherer, err := GetGatherer("test://")
	assert.NoError(t, err)
	assert.NotNil(t, gatherer)
	assert.IsType(t, &TestGatherer{}, gatherer)
}

func TestGetGathererError(t *testing.T) {
	RegisterGatherer(&TestGathererA{})
	RegisterGatherer(&TestGathererB{})

	_, err := GetGatherer("invalid://")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no gatherer found for URI: invalid://")
}
