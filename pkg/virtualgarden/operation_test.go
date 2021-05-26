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

package virtualgarden

import (
	"github.com/gardener/virtual-garden/pkg/api"
	mockclient "github.com/gardener/virtual-garden/pkg/mock/controller-runtime/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Operation", func() {
	Describe("#NewOperation", func() {
		It("should return the correct operation object", func() {
			var (
				c         = mockclient.NewMockClient(gomock.NewController(GinkgoT()))
				log       = logrus.New()
				namespace = "foo"
				imports   = &api.Imports{
					HostingCluster: api.HostingCluster{InfrastructureProvider: api.InfrastructureProviderGCP},
					VirtualGarden: api.VirtualGarden{
						CreateNamespace: true,
					},
				}
				imageRefs = &api.ImageRefs{
					ETCDImage:                  "eu.gcr.io/sap-se-gcr-k8s-public/quay_io/coreos/etcd:v3.3.17",
					ETCDBackupRestoreImage:     "eu.gcr.io/sap-se-gcr-k8s-public/eu_gcr_io/gardener-project/gardener/etcdbrctl:v0.9.1",
					KubeControllerManagerImage: "",
					KubeAPIServerImage:         "",
				}
			)

			operationInterface, err := NewOperation(c, log, namespace, imports, imageRefs)
			Expect(err).NotTo(HaveOccurred())

			op, ok := operationInterface.(*operation)
			Expect(ok).To(BeTrue())
			Expect(op.client).To(Equal(c))
			Expect(op.log).To(Equal(log))
			Expect(op.namespace).To(Equal(namespace))
			Expect(op.imports.VirtualGarden.CreateNamespace).To(Equal(imports.VirtualGarden.CreateNamespace))
			Expect(op.imports.VirtualGarden.DeleteNamespace).To(Equal(false))
			Expect(op.imports).To(Equal(imports))
		})
	})
})
