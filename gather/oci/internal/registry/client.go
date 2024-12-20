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
	"net/http"

	"github.com/spf13/viper"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/enterprise-contract/go-gather/gather/oci/internal/network"
)

/* This code is sourced from the open-policy-agent/conftest project. */

func SetupClient(repository *remote.Repository, transport http.RoundTripper) error {
	registry := repository.Reference.Host()

	// If `--tls=false` was provided or accessing the registry via loopback with
	// `--tls` flag was not set to true
	forceTLS := viper.IsSet("tls") && viper.GetBool("tls")
	forcePlain := viper.IsSet("tls") && !viper.GetBool("tls")
	if forcePlain || (network.IsLoopback(network.Hostname(registry)) && !forceTLS) {
		// Docker by default accesses localhost using plaintext HTTP
		repository.PlainHTTP = true
	}

	httpClient := &http.Client{
		Transport: retry.NewTransport(transport),
	}

	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{
		AllowPlaintextPut:        true,
		DetectDefaultNativeStore: true,
	})
	if err != nil {
		return err
	}

	client := &auth.Client{
		Client:     httpClient,
		Credential: credentials.Credential(store),
		Cache:      auth.NewCache(),
	}
	client.SetUserAgent("conftest")

	repository.Client = client

	return nil
}
