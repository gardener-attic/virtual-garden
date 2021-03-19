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

package gcp_test

import (
	"encoding/json"
	"fmt"

	. "github.com/gardener/virtual-garden/pkg/provider/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials", func() {
	Describe("#", func() {
		var (
			projectID          string
			serviceAccountData string
		)

		BeforeEach(func() {
			projectID = "project"
			serviceAccountData = fmt.Sprintf(`{"project_id": "%s"}`, projectID)
		})

		Describe("#ExtractServiceAccountProjectID", func() {
			It("should fail unmarshalling the document", func() {
				serviceAccountData = "zy"

				actualProjectID, err := ExtractServiceAccountProjectID(serviceAccountData)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&json.SyntaxError{}))
				Expect(actualProjectID).To(BeEmpty())
			})

			It("should fail extracting the project ID", func() {
				serviceAccountData = "{}"

				actualProjectID, err := ExtractServiceAccountProjectID(serviceAccountData)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("no service account")))
				Expect(actualProjectID).To(BeEmpty())
			})

			It("should correctly extract the project ID", func() {
				actualProjectID, err := ExtractServiceAccountProjectID(serviceAccountData)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualProjectID).To(Equal(projectID))
			})
		})
	})
})
