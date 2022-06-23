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

	policyv1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

var _ = Describe("Api Server misc test", func() {
	It("Should create a PodDisruptionBudget for the kube-apiserver", func() {
		namespaceName := "apiserverpoddisruptionbudget"

		ctx := context.Background()
		defer ctx.Done()

		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		err := testenv.Client.Create(ctx, &namespace)
		Expect(err).NotTo(HaveOccurred())

		infrastructureProvider, err := provider.NewInfrastructureProvider(api.InfrastructureProviderGCP)
		Expect(err).NotTo(HaveOccurred())

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                getImportsForMiscTest(),
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{},
		}

		// deploy PodDisruptionBudget
		err = operation.deployKubeAPIServerPodDisruptionBudget(ctx)
		Expect(err).NotTo(HaveOccurred())

		// check PodDisruptionBudget
		podDisruptionBudget := operation.emptyPodDisruptionBudget()
		objectKey := client.ObjectKey{Name: podDisruptionBudget.Name, Namespace: podDisruptionBudget.Namespace}
		err = testenv.Client.Get(ctx, objectKey, podDisruptionBudget)
		Expect(err).NotTo(HaveOccurred())
		Expect(podDisruptionBudget.Spec.Selector.MatchLabels).To(Equal(kubeAPIServerLabels()))
		Expect(podDisruptionBudget.Spec.MinAvailable).NotTo(BeNil())
		Expect(podDisruptionBudget.Spec.MinAvailable.IntVal).To(Equal(int32(2)))

		// delete PodDisruptionBudget
		err = operation.deleteKubeAPIServerPodDisruptionBudget(ctx)
		Expect(err).NotTo(HaveOccurred())

		podDisruptionBudgetList := policyv1.PodDisruptionBudgetList{}
		Expect(testenv.Client.List(ctx, &podDisruptionBudgetList)).To(Succeed())
		Expect(podDisruptionBudgetList.Items).To(BeEmpty())
	})

	It("Should create a ServiceAccount for the kube-apiserver", func() {
		namespaceName := "apiserverserviceaccount"

		ctx := context.Background()
		defer ctx.Done()

		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		err := testenv.Client.Create(ctx, &namespace)
		Expect(err).NotTo(HaveOccurred())

		infrastructureProvider, err := provider.NewInfrastructureProvider(api.InfrastructureProviderGCP)
		Expect(err).NotTo(HaveOccurred())

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                getImportsForMiscTest(),
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{},
		}

		// deploy PodDisruptionBudget
		err = operation.deployKubeAPIServerServiceAccount(ctx)
		Expect(err).NotTo(HaveOccurred())

		// check ServiceAccount
		serviceAccount := operation.emptyServiceAccount()
		objectKey := client.ObjectKey{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}
		err = testenv.Client.Get(ctx, objectKey, serviceAccount)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceAccount.Labels).To(Equal(kubeAPIServerLabels()))

		// delete PodDisruptionBudget
		err = operation.deleteKubeAPIServerServiceAccount(ctx)
		Expect(err).NotTo(HaveOccurred())

		serviceAccount = operation.emptyServiceAccount()
		err = testenv.Client.Get(ctx, objectKey, serviceAccount)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
})

func getImportsForMiscTest() *api.Imports {
	return &api.Imports{
		RuntimeCluster:         lsv1alpha1.Target{},
		RuntimeClusterSettings: api.ClusterSettings{},
		VirtualGarden: api.VirtualGarden{
			ETCD: nil,
			KubeAPIServer: &api.KubeAPIServer{
				Replicas:        0,
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
