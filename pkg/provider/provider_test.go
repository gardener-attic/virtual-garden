// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package provider_test

import (
	"github.com/gardener/virtual-garden/pkg/api"
	. "github.com/gardener/virtual-garden/pkg/provider"
	"github.com/gardener/virtual-garden/pkg/provider/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider", func() {
	Describe("#NewInfrastructureProvider", func() {
		It("should succeed for alicloud", func() {
			provider, err := NewInfrastructureProvider(api.InfrastructureProviderAlicloud)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should succeed for aws", func() {
			provider, err := NewInfrastructureProvider(api.InfrastructureProviderAWS)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should succeed for gcp", func() {
			provider, err := NewInfrastructureProvider(api.InfrastructureProviderGCP)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should fail for unsupported providers", func() {
			provider, err := NewInfrastructureProvider("foo")
			Expect(err).To(MatchError(ContainSubstring("unsupported")))
			Expect(provider).To(BeNil())
		})
	})

	Describe("#NewBackupProvider", func() {
		var (
			fooType        = api.InfrastructureProviderType("foo")
			barType        = api.InfrastructureProviderType("bar")
			credentialsRef = "foo"
			credentials    map[string]api.Credentials
		)

		BeforeEach(func() {
			credentials = map[string]api.Credentials{credentialsRef: {}}
		})

		It("should fail if the referenced credentials cannot be found", func() {
			provider, err := NewBackupProvider("foo", credentials, "")
			Expect(err).To(MatchError(ContainSubstring("could not find referenced credentials")))
			Expect(provider).To(BeNil())
		})

		It("should fail if the referenced credentials type is not equal provider type", func() {
			credentials[credentialsRef] = api.Credentials{Type: barType}

			provider, err := NewBackupProvider(fooType, credentials, credentialsRef)
			Expect(err).To(MatchError(ContainSubstring("does not match provider type")))
			Expect(provider).To(BeNil())
		})

		It("should fail for alicloud", func() {
			credentials[credentialsRef] = api.Credentials{Type: api.InfrastructureProviderAlicloud}

			provider, err := NewBackupProvider(api.InfrastructureProviderAlicloud, credentials, credentialsRef)
			Expect(err).To(MatchError(ContainSubstring("unsupported")))
			Expect(provider).To(BeNil())
		})

		It("should fail for aws", func() {
			credentials[credentialsRef] = api.Credentials{Type: api.InfrastructureProviderAWS}

			provider, err := NewBackupProvider(api.InfrastructureProviderAWS, credentials, credentialsRef)
			Expect(err).To(MatchError(ContainSubstring("unsupported")))
			Expect(provider).To(BeNil())
		})

		It("should succeed for gcp", func() {
			credentials[credentialsRef] = api.Credentials{
				Type: api.InfrastructureProviderGCP,
				Data: map[string]string{gcp.DataKeyServiceAccountJSON: "{\"project_id\": \"my-project\"}"},
			}

			provider, err := NewBackupProvider(api.InfrastructureProviderGCP, credentials, credentialsRef)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should fail for unsupported providers", func() {
			credentials[credentialsRef] = api.Credentials{Type: fooType}

			provider, err := NewBackupProvider(fooType, credentials, credentialsRef)
			Expect(err).To(MatchError(ContainSubstring("unsupported")))
			Expect(provider).To(BeNil())
		})
	})
})
