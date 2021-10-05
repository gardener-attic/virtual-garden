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

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/virtual-garden/pkg/api"
	"github.com/gardener/virtual-garden/pkg/provider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Api Server create secrets test", func() {
	It("Should create api server secrets", func() {
		namespaceName := "apiserversecrets"

		// checkSecret reads the secret with the specified name and checks that its data section contains the given keys.
		checkSecret := func(ctx context.Context, secretName string, keys ...string) {
			objectKey := client.ObjectKey{Name: secretName, Namespace: namespaceName}
			secret := &v1.Secret{}
			err := testenv.Client.Get(ctx, objectKey, secret)
			Expect(err).To(BeNil())
			for _, key := range keys {
				Expect(secret.Data).To(HaveKey(key))
			}
		}

		ctx := context.Background()
		defer ctx.Done()

		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		err := testenv.Client.Create(ctx, &namespace)
		Expect(err).To(BeNil())

		imports := getImportsForApiServerSecretsTest()

		infrastructureProvider, err := provider.NewInfrastructureProvider(api.InfrastructureProviderGCP)
		Expect(err).To(BeNil())

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                &imports,
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{},
		}

		// deploy certificates
		checksums1 := make(map[string]string)
		staticToken1, err := operation.deployKubeAPIServerSecrets(ctx, checksums1)
		Expect(err).To(BeNil())
		Expect(staticToken1).NotTo(HaveLen(0))

		checkSecret(ctx, KubeApiServerSecretNameAdmissionKubeconfig, ValidatingWebhookKey, MutatingWebhookKey)
		checkSecret(ctx, KubeApiServerSecretNameAuditWebhookConfig, AuditWebhookConfigKey)
		checkSecret(ctx, KubeApiServerSecretNameAuthWebhookConfig, ConfigYamlKey)
		checkSecret(ctx, KubeApiServerSecretNameStaticToken, StaticTokenKey)
		checkSecret(ctx, KubeApiServerSecretNameEncryptionConfig, EncryptionConfigKey)
		checkSecret(ctx, KubeApiServerSecretNameServiceAccountKey, ServiceAccountKey)

		Expect(checksums1).To(HaveKey(ChecksumKeyKubeAPIServerAuditWebhookConfig))
		Expect(checksums1).To(HaveKey(ChecksumKeyKubeAPIServerAuthWebhookConfig))
		Expect(checksums1).To(HaveKey(ChecksumKeyKubeAPIServerStaticToken))
		Expect(checksums1).To(HaveKey(ChecksumKeyKubeAPIServerEncryptionConfig))
		Expect(checksums1).To(HaveKey(ChecksumKeyServiceAccountKey))

		// redeploy and check that secrets remain unchanged
		checksums2 := make(map[string]string)
		staticToken2, err := operation.deployKubeAPIServerSecrets(ctx, checksums2)
		Expect(err).To(BeNil())
		Expect(checksums1).To(Equal(checksums2))
		Expect(staticToken1).To(Equal(staticToken2))

		// delete secrets and check that they are gone
		Expect(operation.deleteKubeAPIServerSecrets(ctx)).To(Succeed())
		secretList := &v1.SecretList{}
		Expect(testenv.Client.List(ctx, secretList)).To(Succeed())
		Expect(secretList.Items).To(BeEmpty())
	})
})

func getImportsForApiServerSecretsTest() api.Imports {
	return api.Imports{
		Cluster:        lsv1alpha1.Target{},
		HostingCluster: api.HostingCluster{},
		VirtualGarden: api.VirtualGarden{
			ETCD: nil,
			KubeAPIServer: &api.KubeAPIServer{
				Replicas:        0,
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
