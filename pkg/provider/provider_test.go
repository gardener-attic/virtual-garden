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
	"io/ioutil"

	"github.com/gardener/virtual-garden/pkg/provider/alicloud"

	"github.com/gardener/virtual-garden/pkg/api"
	. "github.com/gardener/virtual-garden/pkg/provider"
	"github.com/gardener/virtual-garden/pkg/provider/gcp"
	"github.com/sirupsen/logrus"

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
			fooType = api.InfrastructureProviderType("foo")
			log     = &logrus.Logger{Out: ioutil.Discard}
		)

		It("should succeed for alicloud", func() {
			credentials := api.Credentials{
				Data: map[string]string{
					alicloud.DataKeyAccessKeyID:     "test-id",
					alicloud.DataKeyAccessKeySecret: "test-secret",
					alicloud.DataKeyStorageEndpoint: "test-endpoint",
				},
			}

			provider, err := NewBackupProvider(api.InfrastructureProviderAlicloud, &credentials, "", "", log)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should succeed for gcp", func() {
			credentials := api.Credentials{
				Data: map[string]string{gcp.DataKeyServiceAccountJSON: "{\"project_id\": \"my-project\"}"},
			}

			provider, err := NewBackupProvider(api.InfrastructureProviderGCP, &credentials, "", "", log)
			Expect(err).To(BeNil())
			Expect(provider).NotTo(BeNil())
		})

		It("should fail for unsupported providers", func() {
			credentials := api.Credentials{}

			provider, err := NewBackupProvider(fooType, &credentials, "", "", log)
			Expect(err).To(MatchError(ContainSubstring("unsupported")))
			Expect(provider).To(BeNil())
		})
	})
})
