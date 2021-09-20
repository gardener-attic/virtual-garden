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
	"fmt"

	. "github.com/gardener/virtual-garden/pkg/provider/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("BackupProvider", func() {
	const projectID = "project"

	var (
		credentialsData    map[string]string
		serviceAccountData = fmt.Sprintf(`{"project_id": "%s"}`, projectID)
	)

	BeforeEach(func() {
		credentialsData = map[string]string{DataKeyServiceAccountJSON: serviceAccountData}
	})

	Describe("#ComputeETCDBackupConfiguration", func() {
		etcdBackupSecretVolumeMountPath := "/foo/bar"

		It("should return the correct backup values", func() {
			provider, err := NewBackupProvider(credentialsData, "", "")
			Expect(err).To(BeNil())

			storageProviderName, secretData, environment := provider.ComputeETCDBackupConfiguration(etcdBackupSecretVolumeMountPath, "")
			Expect(storageProviderName).To(Equal("GCS"))
			Expect(secretData).To(Equal(map[string][]byte{DataKeyServiceAccountJSON: []byte(serviceAccountData)}))
			Expect(environment).To(Equal([]corev1.EnvVar{{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: etcdBackupSecretVolumeMountPath + "/" + DataKeyServiceAccountJSON,
			}}))
		})
	})
})
