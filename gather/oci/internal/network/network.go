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

package network

import (
	"net"
	"strings"
)

/* This code is sourced from the open-policy-agent/conftest project. */
func Hostname(ref string) string {
	if ref == "" {
		return ref
	}

	ref = strings.TrimPrefix(ref, "oci://")

	slash := strings.IndexByte(ref, '/')
	if slash > 0 {
		ref = ref[0:slash]
	}

	host, _, err := net.SplitHostPort(ref)
	if err == nil {
		return host
	}

	return ref
}

func IsLoopback(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0:0:0:0:0:0:0:1" {
		// fast path
		return true
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			return true
		}
	}

	return false
}
