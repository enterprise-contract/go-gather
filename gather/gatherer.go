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
	"fmt"

	"github.com/enterprise-contract/go-gather/metadata"
)

type Gatherer interface {
	Gather(ctx context.Context, src, dst string) (metadata.Metadata, error)
	Matcher(uri string) bool
}

var gatherers []Gatherer

func GetGatherer(uri string) (Gatherer, error) {
	for _, gatherer := range gatherers {
		if gatherer.Matcher(uri) {
			return gatherer, nil
		}
	}
	return nil, fmt.Errorf("no gatherer found for URI: %s", uri)
}

func RegisterGatherer(g Gatherer) {
	gatherers = append(gatherers, g)
}
