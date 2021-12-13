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
	"fmt"

	"net"

	"github.com/gardener/gardener/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"
)

const (
	KubeApiServerSecretNameAggregatorCACertificate          = Prefix + "-kube-aggregator-ca"
	KubeApiServerSecretNameAggregatorClientCertificate      = Prefix + "-kube-aggregator"
	KubeApiServerSecretNameApiServerCACertificate           = Prefix + "-kube-apiserver-ca"
	KubeApiServerSecretNameApiServerServerCertificate       = Prefix + "-kube-apiserver"
	KubeApiServerSecretNameKubeControllerManagerCertificate = Prefix + "-kube-controller-manager"
	KubeApiServerSecretNameClientAdminCertificate           = Prefix + "-kubeconfig-for-admin"
	KubeApiServerSecretNameMetricsScraperCertificate        = Prefix + "-metrics-scraper"
	KubeApiServerSecretNameOidcAuthenticationWebhookConfig  = Prefix + "-kube-apiserver-authentication-webhook-config"
)

func (o *operation) deployKubeAPIServerCertificates(ctx context.Context, loadbalancer string, checksums map[string]string) (
	*secretsutil.Certificate,
	error,
) {
	o.log.Infof("Deploying secrets containing kube-apiserver certificates")

	aggregatorCACertificate, err := o.deployKubeApiServerAggregatorCACertificate(ctx, checksums)
	if err != nil {
		return nil, err
	}

	_, err = o.deployKubeApiServerAggregatorClientCertificate(ctx, aggregatorCACertificate, checksums)
	if err != nil {
		return nil, err
	}

	apiServerCACertificate, err := o.deployKubeApiServerApiServerCACertificate(ctx, checksums)
	if err != nil {
		return nil, err
	}

	_, err = o.deployKubeApiServerApiServerTLSServingCertificate(ctx, apiServerCACertificate, loadbalancer, checksums)
	if err != nil {
		return nil, err
	}

	_, err = o.deployKubeApiServerKubeControllerManagerClientCertificate(ctx, apiServerCACertificate, checksums)
	if err != nil {
		return nil, err
	}

	_, err = o.deployKubeApiServerClientAdminCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return nil, err
	}

	_, err = o.deployKubeApiServerMetricsScraperCertificate(ctx, apiServerCACertificate)
	if err != nil {
		return nil, err
	}

	oidcAuthenticationWebhookCert, err := o.deployKubeApiServerSecretOidcAuthenticationWebhookConfig(ctx, apiServerCACertificate, checksums)
	if err != nil {
		return nil, err
	}

	return oidcAuthenticationWebhookCert, nil
}

func (o *operation) deleteKubeAPIServerCertificates(ctx context.Context) error {
	o.log.Infof("Deleting secrets containing kube-apiserver certificates")

	for _, name := range []string{
		KubeApiServerSecretNameAggregatorCACertificate,
		KubeApiServerSecretNameAggregatorClientCertificate,
		KubeApiServerSecretNameApiServerCACertificate,
		KubeApiServerSecretNameApiServerServerCertificate,
		KubeApiServerSecretNameKubeControllerManagerCertificate,
		KubeApiServerSecretNameClientAdminCertificate,
		KubeApiServerSecretNameMetricsScraperCertificate,
		KubeApiServerSecretNameOidcAuthenticationWebhookConfig,
	} {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
		if err := o.client.Delete(ctx, secret); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (o *operation) deployKubeApiServerApiServerCACertificate(ctx context.Context, checksums map[string]string) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameApiServerCACertificate,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:kube-apiserver",
	}

	cert, checksum, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, nil)
	if err != nil {
		return nil, err
	}

	o.exports.KubeApiserverCaPem = string(cert.CertificatePEM)

	checksums[ChecksumKeyKubeAPIServerCA] = checksum
	return cert, err
}

func (o *operation) deployKubeApiServerApiServerTLSServingCertificate(ctx context.Context, caCertificate *secretsutil.Certificate,
	loadbalancer string, checksums map[string]string) (*secretsutil.Certificate, error) {

	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameApiServerServerCertificate,
		CertType:   secretsutil.ServerCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":server:kube-apiserver",
		DNSNames: []string{
			"localhost",
			KubeAPIServerServiceName,
			KubeAPIServerServiceName + "." + o.namespace,
			KubeAPIServerServiceName + "." + o.namespace + ".svc",
			KubeAPIServerServiceName + "." + o.namespace + ".svc.cluster",
			KubeAPIServerServiceName + "." + o.namespace + ".svc.cluster.local",
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster",
			"kubernetes.default.svc.cluster.local",
		},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("100.64.0.1"),
		},
	}

	dnsAccessDomain := o.imports.VirtualGarden.KubeAPIServer.DnsAccessDomain
	if len(dnsAccessDomain) > 0 {
		certConfig.DNSNames = append(certConfig.DNSNames, fmt.Sprintf("api.%s", dnsAccessDomain), fmt.Sprintf("gardener.%s", dnsAccessDomain))
	}

	loadbalancerIP := net.ParseIP(loadbalancer)
	if loadbalancerIP != nil && len(loadbalancerIP.String()) != 0 {
		certConfig.IPAddresses = append(certConfig.IPAddresses, loadbalancerIP)
	} else {
		// because it is not an ip address, we assume it is a DNS domain
		certConfig.DNSNames = append(certConfig.DNSNames, loadbalancer)
	}

	cert, checksum, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, nil)
	if err != nil {
		return nil, err
	}

	checksums[ChecksumKeyKubeAPIServerServer] = checksum
	return cert, err
}

func (o *operation) deployKubeApiServerKubeControllerManagerClientCertificate(ctx context.Context, caCertificate *secretsutil.Certificate, checksums map[string]string) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameKubeControllerManagerCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: "system:kube-controller-manager",
	}

	kubeconfigGen := &kubeconfigGenerator{
		user:   "kube-controller-manager",
		server: "https://virtual-garden-kube-apiserver:443",
	}

	cert, checksum, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, kubeconfigGen)
	if err != nil {
		return nil, err
	}

	checksums[ChecksumKeyKubeControllerManagerClient] = checksum
	return cert, err
}

func (o *operation) deployKubeApiServerClientAdminCertificate(ctx context.Context, caCertificate *secretsutil.Certificate,
	loadBalancer string) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameClientAdminCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":client:admin",
		Organization: []string{
			"system:masters",
		},
	}

	kubeconfigGen := &kubeconfigGenerator{
		user:   "admin",
		server: o.infrastructureProvider.GetKubeAPIServerURL(o.imports.VirtualGarden.KubeAPIServer, loadBalancer),
	}

	cert, _, kubeconfig, err := deployCertificate(ctx, o.client, o.namespace, certConfig, kubeconfigGen)
	if err != nil {
		return nil, err
	}

	o.exports.KubeconfigYaml = string(kubeconfig)

	return cert, err
}

func (o *operation) deployKubeApiServerSecretOidcAuthenticationWebhookConfig(ctx context.Context,
	caCertificate *secretsutil.Certificate, checksums map[string]string) (cert *secretsutil.Certificate, err error) {
	if !o.isOidcWebhookAuthenticatorEnabled() {
		return nil, nil
	}

	secret := o.emptySecret(KubeApiServerSecretNameOidcAuthenticationWebhookConfig)
	secretKey := client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}
	err = o.client.Get(ctx, secretKey, secret)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}

		certConfig := &secretsutil.CertificateSecretConfig{
			Name:       KubeApiServerSecretNameOidcAuthenticationWebhookConfig,
			CertType:   secretsutil.ClientCert,
			SigningCA:  caCertificate,
			CommonName: "apiserver-oidc-webhook-authenticator-operator",
		}

		cert, err = certConfig.GenerateCertificate()
		if err != nil {
			return nil, err
		}

		kubeconfig := o.createKubeconfigForOidcWebhook(cert)

		kubeconfigData, err := yaml.Marshal(kubeconfig)
		if err != nil {
			return nil, err
		}

		secret = o.emptySecret(KubeApiServerSecretNameOidcAuthenticationWebhookConfig)
		_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
			if secret.Data == nil {
				secret.Data = make(map[string][]byte)
			}
			secret.Data[SecretKeyKubeconfigYaml] = kubeconfigData
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		kubeconfigData, ok := secret.Data[SecretKeyKubeconfigYaml]
		if !ok {
			return nil, fmt.Errorf("Secret %s does not contain key %s",
				KubeApiServerSecretNameOidcAuthenticationWebhookConfig, SecretKeyKubeconfigYaml)
		}

		kubeconfig := &v1.Config{}
		if err = yaml.Unmarshal(kubeconfigData, kubeconfig); err != nil {
			return nil, err
		}

		user := extractUser(kubeconfig, UserVirtualGardenKubeApiServer)
		if user == nil {
			return nil, fmt.Errorf("Kubeconfig in secret %s does not contain user %s",
				KubeApiServerSecretNameOidcAuthenticationWebhookConfig, UserVirtualGardenKubeApiServer)
		}

		cert, err = secretsutil.LoadCertificate(KubeApiServerSecretNameOidcAuthenticationWebhookConfig, user.AuthInfo.ClientKeyData, user.AuthInfo.ClientCertificateData)
		if err != nil {
			return nil, err
		}
	}

	checksums[ChecksumKeyKubeAPIServerOidcAuthenticationWebhookConfig] = utils.ComputeChecksum(secret.Data)
	return cert, nil
}

func extractUser(kubeconfig *v1.Config, name string) *v1.NamedAuthInfo {
	for i := range kubeconfig.AuthInfos {
		user := &kubeconfig.AuthInfos[i]
		if user.Name == name {
			return user
		}
	}
	return nil
}

func (o *operation) createKubeconfigForOidcWebhook(cert *secretsutil.Certificate) *v1.Config {
	return &v1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: "authentication-webhook",
		Clusters: []v1.NamedCluster{
			{
				Name: "oidc-webhook-authenticator",
				Cluster: v1.Cluster{
					CertificateAuthorityData: []byte(o.imports.VirtualGarden.KubeAPIServer.OidcWebhookAuthenticator.CertificateAuthorityData),
					Server:                   "https://oidc-webhook-authenticator.garden/validate-token",
				},
			},
		},
		AuthInfos: []v1.NamedAuthInfo{
			{
				Name: UserVirtualGardenKubeApiServer,
				AuthInfo: v1.AuthInfo{
					ClientCertificateData: cert.CertificatePEM,
					ClientKeyData:         cert.PrivateKeyPEM,
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: "authentication-webhook",
				Context: v1.Context{
					Cluster:  "oidc-webhook-authenticator",
					AuthInfo: "virtual-garden-kube-apiserver",
				},
			},
		},
	}
}

func (o *operation) deployKubeApiServerMetricsScraperCertificate(
	ctx context.Context,
	caCertificate *secretsutil.Certificate) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameMetricsScraperCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":client:metrics-scraper",
	}

	cert, _, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, nil)
	if err != nil {
		return nil, err
	}

	return cert, err
}

func (o *operation) deployKubeApiServerAggregatorCACertificate(ctx context.Context, checksums map[string]string) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameAggregatorCACertificate,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:kube-aggregator",
	}

	cert, checksum, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, nil)
	if err != nil {
		return nil, err
	}

	checksums[ChecksumKeyKubeAggregatorCA] = checksum
	return cert, err
}

func (o *operation) deployKubeApiServerAggregatorClientCertificate(ctx context.Context,
	caCertificate *secretsutil.Certificate, checksums map[string]string) (*secretsutil.Certificate, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameAggregatorClientCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":aggregator-client:kube-aggregator",
	}

	cert, checksum, _, err := deployCertificate(ctx, o.client, o.namespace, certConfig, nil)
	if err != nil {
		return nil, err
	}

	checksums[ChecksumKeyKubeAggregatorClient] = checksum
	return cert, err
}
