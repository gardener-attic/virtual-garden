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
	cryptorand "crypto/rand"
	_ "embed"
	"strings"

	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1 "k8s.io/apiserver/pkg/apis/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/gardener/gardener/pkg/utils"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

const (
	KubeApiServerSecretNameAdmissionKubeconfig = Prefix + "-kube-apiserver-admission-kubeconfig"
	KubeApiServerSecretNameAuditWebhookConfig  = "kube-apiserver-audit-webhook-config"
	KubeApiServerSecretNameAuthWebhookConfig   = Prefix + "-kube-apiserver-auth-webhook-config"
	KubeApiServerSecretNameStaticToken         = Prefix + "-kube-apiserver-static-token"
	KubeApiServerSecretNameEncryptionConfig    = Prefix + "-kube-apiserver-encryption-config"
	KubeApiServerSecretNameServiceAccountKey   = Prefix + "-service-account-key"
)

//go:embed resources/validating-webhook-kubeconfig.yaml
var validatingWebhookKubeconfig []byte

//go:embed resources/mutating-webhook-kubeconfig.yaml
var mutatingWebhookKubeconfig []byte

func (o *operation) deployKubeAPIServerSecrets(ctx context.Context, checksums map[string]string) (string, error) {
	o.log.Infof("Deploying secrets for the kube-apiserver")

	if err := o.deployKubeApiServerSecretAdmissionKubeconfig(ctx); err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretAuditWebhookConfig(ctx, checksums); err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretAuthWebhookConfig(ctx, checksums); err != nil {
		return "", err
	}

	staticTokenHealthCheck, err := o.deployKubeApiServerSecretStaticToken(ctx, checksums)
	if err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretEncryptionConfig(ctx, checksums); err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretServiceAccountKey(ctx, checksums); err != nil {
		return "", err
	}

	return staticTokenHealthCheck, nil
}

func (o *operation) deleteKubeAPIServerSecrets(ctx context.Context) error {
	o.log.Infof("Deleting secrets for the kube-apiserver")

	for _, name := range []string{
		KubeApiServerSecretNameAdmissionKubeconfig,
		KubeApiServerSecretNameAuditWebhookConfig,
		KubeApiServerSecretNameAuthWebhookConfig,
		KubeApiServerSecretNameStaticToken,
		KubeApiServerSecretNameEncryptionConfig,
		KubeApiServerSecretNameServiceAccountKey,
	} {
		secret := o.emptySecret(name)
		if err := o.client.Delete(ctx, secret); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (o *operation) deployKubeApiServerSecretAdmissionKubeconfig(ctx context.Context) error {
	if !o.isWebhookEnabled() {
		return nil
	}

	secret := o.emptySecret(KubeApiServerSecretNameAdmissionKubeconfig)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[ValidatingWebhookKey] = validatingWebhookKubeconfig
		secret.Data[MutatingWebhookKey] = mutatingWebhookKubeconfig
		return nil
	})
	return err
}

func (o *operation) deployKubeApiServerSecretAuditWebhookConfig(ctx context.Context, checksums map[string]string) error {
	config := o.getAPIServerAuditWebhookConfig()
	if len(config) == 0 {
		return nil
	}

	secret := o.emptySecret(KubeApiServerSecretNameAuditWebhookConfig)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[AuditWebhookConfigKey] = []byte(config)
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerAuditWebhookConfig] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) deployKubeApiServerSecretAuthWebhookConfig(ctx context.Context, checksums map[string]string) error {
	if !o.isSeedAuthorizerEnabled() {
		return nil
	}

	const (
		clusterName = "gardener-admission-controller"
		userName    = "virtual-garden-kube-apiserver"
	)

	cluster := v1.NamedCluster{
		Name: clusterName,
		Cluster: v1.Cluster{
			Server: "https://gardener-admission-controller.garden/webhooks/auth/seed",
		},
	}

	if len(o.imports.VirtualGarden.KubeAPIServer.SeedAuthorizer.CertificateAuthorityData) > 0 {
		cluster.Cluster.CertificateAuthorityData = []byte(o.imports.VirtualGarden.KubeAPIServer.SeedAuthorizer.CertificateAuthorityData)
	} else {
		cluster.Cluster.InsecureSkipTLSVerify = true
	}

	user := v1.NamedAuthInfo{
		Name: userName,
	}

	context := v1.NamedContext{
		Name: "auth-webhook",
		Context: v1.Context{
			Cluster:  clusterName,
			AuthInfo: userName,
		},
	}

	config := v1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: "auth-webhook",
		Clusters:       []v1.NamedCluster{cluster},
		AuthInfos:      []v1.NamedAuthInfo{user},
		Contexts:       []v1.NamedContext{context},
	}

	configYaml, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	secret := o.emptySecret(KubeApiServerSecretNameAuthWebhookConfig)
	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		secret.ObjectMeta.Labels = kubeAPIServerLabels()

		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[ConfigYamlKey] = configYaml
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerAuthWebhookConfig] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) deployKubeApiServerSecretStaticToken(ctx context.Context, checksums map[string]string) (string, error) {
	const (
		staticTokenSuffix = ",kube-apiserver-health-check,kube-apiserver-health-check,"
	)

	var staticTokenValue []byte

	secret := o.emptySecret(KubeApiServerSecretNameStaticToken)
	err := o.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}

		// secret does not exist: generate static token
		token, err2 := utils.GenerateRandomString(128)
		if err2 != nil {
			return "", err2
		}

		staticTokenValue = []byte(token + staticTokenSuffix)
	} else {
		// secret exists: use existing value
		staticTokenValue = secret.Data[StaticTokenKey]
	}

	returnToken := strings.TrimSuffix(string(staticTokenValue), staticTokenSuffix)

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[StaticTokenKey] = staticTokenValue
		return nil
	})
	if err != nil {
		return "", err
	}

	checksums[ChecksumKeyKubeAPIServerStaticToken] = utils.ComputeChecksum(secret.Data)
	return returnToken, nil
}

func (o *operation) deployKubeApiServerSecretEncryptionConfig(ctx context.Context, checksums map[string]string) error {
	var encryptionConfigValue []byte

	secret := o.emptySecret(KubeApiServerSecretNameEncryptionConfig)
	err := o.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}

		// secret does not exist: generate encryption config
		encryptionConfigValue, err = o.generateNewEncryptionConfig()
		if err != nil {
			return err
		}
	} else {
		// secret exists: use existing value
		encryptionConfigValue = secret.Data[EncryptionConfigKey]
	}

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[EncryptionConfigKey] = encryptionConfigValue
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerEncryptionConfig] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) generateNewEncryptionConfig() ([]byte, error) {
	secretBytes := make([]byte, 32)
	if _, err := cryptorand.Read(secretBytes); err != nil {
		return nil, err
	}

	secretString := utils.EncodeBase64(secretBytes)

	encryptionConfig := configv1.EncryptionConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EncryptionConfiguration",
			APIVersion: "apiserver.config.k8s.io/v1",
		},
		Resources: []configv1.ResourceConfiguration{
			{
				Resources: []string{
					"secrets",
				},
				Providers: []configv1.ProviderConfiguration{
					{
						AESCBC: &configv1.AESConfiguration{
							Keys: []configv1.Key{
								{
									Name:   "key",
									Secret: secretString,
								},
							},
						},
					},
					{
						Identity: &configv1.IdentityConfiguration{},
					},
				},
			},
		},
	}

	return yaml.Marshal(&encryptionConfig)
}

func (o *operation) deployKubeApiServerSecretServiceAccountKey(ctx context.Context, checksums map[string]string) error {
	var serviceAccountKey []byte

	secret := o.emptySecret(KubeApiServerSecretNameServiceAccountKey)

	if o.imports.VirtualGarden.KubeAPIServer.ServiceAccountKeyPem == nil ||
		len(*o.imports.VirtualGarden.KubeAPIServer.ServiceAccountKeyPem) == 0 {
		certConfig := &secretsutil.CertificateSecretConfig{
			Name:       KubeApiServerSecretNameServiceAccountKey,
			CertType:   secretsutil.CACert,
			CommonName: Prefix + ":ca:kube-apiserver",
			DNSNames: []string{
				"virtual-garden:ca:kube-apiserver",
			},
		}

		err := o.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}

			o.log.Info("Generating a new service account key")
			cert, err := certConfig.GenerateCertificate()
			if err != nil {
				return err
			}

			serviceAccountKey = cert.PrivateKeyPEM
		} else {
			// secret exists: use existing value
			serviceAccountKey = secret.Data[ServiceAccountKey]
		}
	} else {
		o.log.Info("Using the provided service account key")
		serviceAccountKey = []byte(*o.imports.VirtualGarden.KubeAPIServer.ServiceAccountKeyPem)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[ServiceAccountKey] = serviceAccountKey
		return nil
	})
	if err != nil {
		return err
	}

	o.exports.ServiceAccountKeyPem = string(serviceAccountKey)

	checksums[ChecksumKeyServiceAccountKey] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) emptySecret(name string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
