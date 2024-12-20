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
	expander "github.com/enterprise-contract/go-gather/expand"
	_ "github.com/enterprise-contract/go-gather/expand/bzip2"
	_ "github.com/enterprise-contract/go-gather/expand/tar"
	_ "github.com/enterprise-contract/go-gather/expand/zip"
)

func GetExpander(extension string) expander.Expander {
	return expander.GetExpander(extension)
}
