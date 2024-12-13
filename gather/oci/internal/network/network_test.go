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

import "testing"

func TestHostname(t *testing.T) {
	cases := []struct {
		given    string
		expected string
	}{
		{
			given:    "example.com",
			expected: "example.com",
		},
		{
			given:    "example.com:12345",
			expected: "example.com",
		},
		{
			given:    "example.com/path",
			expected: "example.com",
		},
		{
			given:    "oci://example.com",
			expected: "example.com",
		},
		{
			given:    "oci://example.com:12345",
			expected: "example.com",
		},
		{
			given:    "oci://example.com:12345",
			expected: "example.com",
		},
		{
			given:    "oci://example.com:12345/path",
			expected: "example.com",
		},
		{
			given:    "example.com//path",
			expected: "example.com",
		},
		{
			given:    "example.com/p1/p2/p3",
			expected: "example.com",
		},
		{
			given:    "localhost",
			expected: "localhost",
		},
		{
			given:    "localhost:12345",
			expected: "localhost",
		},
		{
			given:    "localhost/path",
			expected: "localhost",
		},
		{
			given:    "localhost:12345/path",
			expected: "localhost",
		},
		{
			given:    "1.2.3.4",
			expected: "1.2.3.4",
		},
		{
			given:    "1.2.3.4:12345",
			expected: "1.2.3.4",
		},
		{
			given:    "oci://1.2.3.4",
			expected: "1.2.3.4",
		},
		{
			given:    "oci://1.2.3.4:12345",
			expected: "1.2.3.4",
		},
		{
			given:    "::1",
			expected: "::1",
		},
		{
			given:    "[::1]:12345",
			expected: "::1",
		},
		{
			given:    "[::1]:12345/path",
			expected: "::1",
		},
		{
			given:    "2001:db8:85a3:8d3:1319:8a2e:370:7348",
			expected: "2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			given:    "[2001:db8:85a3:8d3:1319:8a2e:370:7348]:12345",
			expected: "2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			given:    "[2001:db8:85a3:8d3:1319:8a2e:370:7348]:12345/path",
			expected: "2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
	}

	for _, c := range cases {
		t.Run(c.given, func(t *testing.T) {
			if got, expected := Hostname(c.given), c.expected; got != expected {
				t.Errorf("Hostname(%q) = %q, expected %q", c.given, got, c.expected)
			}
		})
	}
}
