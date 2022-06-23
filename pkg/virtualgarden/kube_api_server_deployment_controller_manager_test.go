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
	"context"

	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/virtual-garden/pkg/api"
	"github.com/gardener/virtual-garden/pkg/provider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Api Server deployment test", func() {
	It("Should create deployment virtual-garden-kube-controller-manager", func() {
		var (
			namespaceName = "apiservercontrollermanagerdeployment"
			checksums     = map[string]string{
				ChecksumKeyKubeAPIServerAuditPolicyConfig: "testChecksum1",
				ChecksumKeyKubeAPIServerEncryptionConfig:  "testChecksum2",
			}
			kubeControllerManagerImage = "testImage"
		)

		ctx := context.Background()
		defer ctx.Done()

		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		err := testenv.Client.Create(ctx, &namespace)
		Expect(err).NotTo(HaveOccurred())

		infrastructureProvider, err := provider.NewInfrastructureProvider(api.InfrastructureProviderGCP)
		Expect(err).NotTo(HaveOccurred())

		imports := getImportsForControllerManagerDeploymentTest()

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                imports,
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{KubeControllerManagerImage: kubeControllerManagerImage},
		}

		// deploy Deployment
		err = operation.deployKubeAPIServerDeploymentControllerManager(ctx, checksums)
		Expect(err).NotTo(HaveOccurred())

		// check Deployment
		deployment := operation.emptyDeployment(KubeAPIServerDeploymentNameControllerManager)
		objectKey := client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}
		err = testenv.Client.Get(ctx, objectKey, deployment)
		Expect(err).NotTo(HaveOccurred())

		// delete Deployment
		err = operation.deleteDeployments(ctx)
		Expect(err).NotTo(HaveOccurred())
		deploymentList := appsv1.DeploymentList{}
		Expect(testenv.Client.List(ctx, &deploymentList)).To(Succeed())
		Expect(deploymentList.Items).To(BeEmpty())
	})
})

func getImportsForControllerManagerDeploymentTest() *api.Imports {
	return &api.Imports{
		RuntimeCluster:         lsv1alpha1.Target{},
		RuntimeClusterSettings: api.ClusterSettings{},
		VirtualGarden: api.VirtualGarden{
			ETCD: nil,
			KubeAPIServer: &api.KubeAPIServer{
				Replicas:        2,
				SNI:             nil,
				DnsAccessDomain: "com.our.test",
				GardenerControlplane: api.GardenerControlplane{
					ValidatingWebhook: api.AdmissionWebhookConfig{
						Token: api.AdmissionWebhookTokenConfig{Enabled: true},
					},
					MutatingWebhook: api.AdmissionWebhookConfig{
						Token: api.AdmissionWebhookTokenConfig{Enabled: true},
					},
				},
				ServiceAccountKeyPem:     pointer.String("test-service-account-key"),
				AuditWebhookConfig:       api.AuditWebhookConfig{Config: "testconfig"},
				AuditWebhookBatchMaxSize: "",
				SeedAuthorizer: api.SeedAuthorizer{
					Enabled:                  true,
					CertificateAuthorityData: "test-ca-data",
				},
				EventTTL:               nil,
				OidcIssuerURL:          nil,
				AdditionalVolumeMounts: nil,
				AdditionalVolumes:      nil,
			},
			DeleteNamespace:   false,
			PriorityClassName: "",
		},
	}
}
