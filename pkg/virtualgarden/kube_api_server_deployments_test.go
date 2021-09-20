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
	It("Should create deployment virtual-garden-kube-apiserver", func() {
		var (
			namespaceName          = "apiserverdeployment"
			staticTokenHealthCheck = "testStaticTokenHealthCheck"
			checksums              = map[string]string{
				ChecksumKeyKubeAPIServerAuditPolicyConfig: "testChecksum1",
				ChecksumKeyKubeAPIServerEncryptionConfig:  "testChecksum2",
			}
			kubeApiServerImage = "testImage"
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

		imports := getImportsForDeploymentTest()

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                imports,
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{KubeAPIServerImage: kubeApiServerImage},
		}

		// deploy Deployment
		err = operation.deployKubeAPIServerDeployment(ctx, checksums, staticTokenHealthCheck)
		Expect(err).NotTo(HaveOccurred())

		// check Deployment
		deployment := operation.emptyDeployment(KubeAPIServerDeploymentNameAPIServer)
		objectKey := client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}
		err = testenv.Client.Get(ctx, objectKey, deployment)
		Expect(err).NotTo(HaveOccurred())

		Expect(deployment.Labels).To(Equal(kubeAPIServerLabels()))
		Expect(deployment.Spec.Replicas).To(Equal(pointer.Int32(int32(imports.VirtualGarden.KubeAPIServer.Replicas))))
		Expect(deployment.Spec.Selector.MatchLabels).To(Equal(kubeAPIServerLabels()))
		Expect(deployment.Spec.Template.Annotations).To(Equal(checksums))
		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(kubeApiServerImage))
		Expect(deployment.Spec.Template.Spec.PriorityClassName).To(Equal(imports.VirtualGarden.PriorityClassName))
		Expect(deployment.Spec.Template.Spec.ServiceAccountName).To(Equal(KubeAPIServerServiceName))

		// delete Deployment
		err = operation.deleteDeployments(ctx)
		Expect(err).NotTo(HaveOccurred())
		deploymentList := appsv1.DeploymentList{}
		Expect(testenv.Client.List(ctx, &deploymentList)).To(Succeed())
		Expect(deploymentList.Items).To(BeEmpty())
	})
})

func getImportsForDeploymentTest() *api.Imports {
	return &api.Imports{
		Cluster:        lsv1alpha1.Target{},
		HostingCluster: api.HostingCluster{},
		VirtualGarden: api.VirtualGarden{
			ETCD: nil,
			KubeAPIServer: &api.KubeAPIServer{
				Replicas:        2,
				SNI:             nil,
				DnsAccessDomain: "com.our.test",
				GardenerControlplane: api.GardenerControlplane{
					ValidatingWebhookEnabled: true,
					MutatingWebhookEnabled:   true,
				},
				ServiceAccountKeyPem:     pointer.String("test-service-account-key"),
				AuditWebhookConfig:       api.AuditWebhookConfig{Config: "testconfig"},
				AuditWebhookBatchMaxSize: "",
				SeedAuthorizer: api.SeedAuthorizer{
					Enabled:                  true,
					CertificateAuthorityData: "test-ca-data",
				},
				HVPAEnabled:             false,
				HVPA:                    nil,
				EventTTL:                nil,
				OidcIssuerURL:           nil,
				AdditionalVolumeMounts:  nil,
				AdditionalVolumes:       nil,
				HorizontalPodAutoscaler: nil,
			},
			DeleteNamespace:   false,
			PriorityClassName: "",
		},
	}
}
