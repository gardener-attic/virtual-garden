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
)

const (
	KubeApiServerSecretNameAdmissionKubeconfig = Prefix + "-kube-apiserver-admission-kubeconfig"
	KubeApiServerSecretNameAuditWebhookConfig  = "kube-apiserver-audit-webhook-config"
	KubeApiServerSecretNameBasicAuth           = Prefix + "-kube-apiserver-basic-auth"
	KubeApiServerSecretNameEncryptionConfig    = Prefix + "-kube-apiserver-encryption-config"
	KubeApiServerSecretNameServiceAccountKey   = Prefix + "-service-account-key"

	pwUsers = ",admin,admin,system:masters"
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

	basicAuthPw, err := o.deployKubeApiServerSecretBasicAuth(ctx, checksums)
	if err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretEncryptionConfig(ctx, checksums); err != nil {
		return "", err
	}

	if err := o.deployKubeApiServerSecretServiceAccountKey(ctx, checksums); err != nil {
		return "", err
	}

	return basicAuthPw, nil
}

func (o *operation) deleteKubeAPIServerSecrets(ctx context.Context) error {
	o.log.Infof("Deleting secrets for the kube-apiserver")

	for _, name := range []string{
		KubeApiServerSecretNameAdmissionKubeconfig,
		KubeApiServerSecretNameAuditWebhookConfig,
		KubeApiServerSecretNameBasicAuth,
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
		secret.Data["validating-webhook"] = validatingWebhookKubeconfig
		secret.Data["mutating-webhook"] = mutatingWebhookKubeconfig
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
		secret.Data["audit-webhook-config.yaml"] = []byte(config)
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerAuditWebhookConfig] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) deployKubeApiServerSecretBasicAuth(ctx context.Context, checksums map[string]string) (string, error) {
	const basicAuthKey = "basic_auth.csv"

	var basicAuthValue []byte

	secret := o.emptySecret(KubeApiServerSecretNameBasicAuth)
	err := o.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}

		// secret does not exist: generate password
		pw, err2 := utils.GenerateRandomString(32)
		if err2 != nil {
			return "", err2
		}

		basicAuthValue = []byte(pw + pwUsers)
	} else {
		// secret exists: use existing value
		basicAuthValue = secret.Data[basicAuthKey]
	}

	returnPw := strings.TrimSuffix(string(basicAuthValue), pwUsers)

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[basicAuthKey] = basicAuthValue
		return nil
	})
	if err != nil {
		return "", err
	}

	checksums[ChecksumKeyKubeAPIServerBasicAuth] = utils.ComputeChecksum(secret.Data)
	return returnPw, nil
}

func (o *operation) deployKubeApiServerSecretEncryptionConfig(ctx context.Context, checksums map[string]string) error {
	const encryptionConfigKey = "encryption-config.yaml"

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
		encryptionConfigValue = secret.Data[encryptionConfigKey]
	}

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[encryptionConfigKey] = encryptionConfigValue
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

// NEEDS TO BE REWORKED
// Issues:
// - the data map of the secret has another key than usual for a private key
// - the secret contains only the private key, but not the certificate
// Therefore the loading fails if the secret does already exist.
func (o *operation) deployKubeApiServerSecretServiceAccountKey(ctx context.Context, checksums map[string]string) error {
	const key = "service_account.key"

	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameServiceAccountKey,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:kube-apiserver",
		DNSNames: []string{
			"virtual-garden:ca:kube-apiserver",
		},
	}

	var value []byte

	secret := o.emptySecret(KubeApiServerSecretNameServiceAccountKey)
	err := o.client.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}

		// secret does not exist: generate certificate
		cert, err := certConfig.GenerateCertificate()
		if err != nil {
			return err
		}

		value = cert.PrivateKeyPEM
	} else {
		// secret exists: use existing value
		value = secret.Data[key]
	}

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[key] = value
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyServiceAccountKey] = utils.ComputeChecksum(secret.Data)
	return nil
}

func (o *operation) emptySecret(name string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
