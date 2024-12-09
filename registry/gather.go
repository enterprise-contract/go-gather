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

package registry

import (
	"github.com/enterprise-contract/go-gather/gather"
	_ "github.com/enterprise-contract/go-gather/gather/file"
	_ "github.com/enterprise-contract/go-gather/gather/git"
	_ "github.com/enterprise-contract/go-gather/gather/http"
	_ "github.com/enterprise-contract/go-gather/gather/oci"
)

func GetGatherer(uri string) (gather.Gatherer, error) {
	return gather.GetGatherer(uri)
}
