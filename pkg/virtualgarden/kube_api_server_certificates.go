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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"
)

const (
	KubeApiServerSecretNameAggregatorCACertificate     = Prefix + "-kube-aggregator-ca"
	KubeApiServerSecretNameAggregatorClientCertificate = Prefix + "-kube-aggregator"

	KubeApiServerSecretNameApiServerCACertificate     = Prefix + "-kube-apiserver-ca"
	KubeApiServerSecretNameApiServerServerCertificate = Prefix + "-kube-apiserver"

	KubeApiServerSecretNameKubeControllerManagerCertificate = Prefix + "-kube-controller-manager"
	KubeApiServerSecretNameClientAdminCertificate           = Prefix + "-kubeconfig-for-admin"
	KubeApiServerSecretNameMetricsScraperCertificate        = Prefix + "-metrics-scraper"
)

func (o *operation) deployKubeAPIServerCertificates(ctx context.Context, loadbalancer string) error {
	o.log.Infof("Deploying secrets containing kube-apiserver certificates")

	aggregatorCACertificate, aggregatorCACertChecksum, err := o.deployKubeApiServerAggregatorCACertificate(ctx)
	if err != nil {
		return err
	}

	_, aggregatorClientCertChecksum, err := o.deployKubeApiServerAggregatorClientCertificate(ctx, aggregatorCACertificate)
	if err != nil {
		return err
	}

	apiServerCACertificate, apiServerCACertChecksum, err := o.deployKubeApiServerApiServerCACertificate(ctx)
	if err != nil {
		return err
	}

	_, apiServerServerCertChecksum, err := o.deployKubeApiServerApiServerServerCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return err
	}

	_, kubeControllerManagerClientCertChecksum, err := o.deployKubeApiServerKubeControllerManagerClientCertificate(ctx, apiServerCACertificate)
	if err != nil {
		return err
	}

	_, clientAdminCertChecksum, err := o.deployKubeApiServerClientAdminCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return err
	}

	_, metricsScraperCertChecksum, err := o.deployKubeApiServerMetricsScraperCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return err
	}

	// temporarily
	fmt.Println("Checksums: ",
		aggregatorCACertChecksum,
		aggregatorClientCertChecksum,
		apiServerCACertChecksum,
		apiServerServerCertChecksum,
		kubeControllerManagerClientCertChecksum,
		clientAdminCertChecksum,
		metricsScraperCertChecksum,
	)

	return nil
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
	} {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
		if err := o.client.Delete(ctx, secret); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (o *operation) deployKubeApiServerApiServerCACertificate(ctx context.Context) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameApiServerCACertificate,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:kube-apiserver",
	}
	return o.deployCertificate(ctx, certConfig, nil)
}

func (o *operation) deployKubeApiServerApiServerServerCertificate(ctx context.Context, caCertificate *secretsutil.Certificate,
	loadbalancer string) (*secretsutil.Certificate, string, error) {
	dnsAccessDomain := o.imports.VirtualGarden.KubeAPIServer.DnsAccessDomain

	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameApiServerServerCertificate,
		CertType:   secretsutil.ServerClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":server:kube-apiserver",
		DNSNames: []string{
			"127.0.0.1",
			"localhost",
			"100.64.0.1",
			"virtual-garden-kube-apiserver",
			"virtual-garden-kube-apiserver.garden",
			"virtual-garden-kube-apiserver.garden.svc",
			"virtual-garden-kube-apiserver.garden.svc.cluster",
			"virtual-garden-kube-apiserver.garden.svc.cluster.local",
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster",
			"kubernetes.default.svc.cluster.local",
			loadbalancer,
			fmt.Sprintf("api.%s", dnsAccessDomain),
			fmt.Sprintf("gardener.%s", dnsAccessDomain),
		},
	}
	return o.deployCertificate(ctx, certConfig, nil)
}

func (o *operation) deployKubeApiServerKubeControllerManagerClientCertificate(ctx context.Context, caCertificate *secretsutil.Certificate) (*secretsutil.Certificate, string, error) {
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

	return o.deployCertificate(ctx, certConfig, kubeconfigGen)
}

func (o *operation) deployKubeApiServerClientAdminCertificate(ctx context.Context, caCertificate *secretsutil.Certificate,
	loadBalancer string) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameClientAdminCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":client:admin",
	}

	// o.imports.VirtualGarden.KubeAPIServer could be nil

	kubeconfigGen := &kubeconfigGenerator{
		user:   "admin",
		server: o.infrastructureProvider.GetKubeAPIServerURL(o.imports.VirtualGarden.KubeAPIServer, loadBalancer),
	}

	return o.deployCertificate(ctx, certConfig, kubeconfigGen)
}

func (o *operation) deployKubeApiServerMetricsScraperCertificate(ctx context.Context, caCertificate *secretsutil.Certificate,
	loadBalancer string) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameMetricsScraperCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":client:metrics-scraper",
	}

	return o.deployCertificate(ctx, certConfig, nil)
}

func (o *operation) deployKubeApiServerAggregatorCACertificate(ctx context.Context) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameAggregatorCACertificate,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:kube-aggregator",
	}
	return o.deployCertificate(ctx, certConfig, nil)
}

func (o *operation) deployKubeApiServerAggregatorClientCertificate(ctx context.Context,
	caCertificate *secretsutil.Certificate) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       KubeApiServerSecretNameAggregatorClientCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":aggregator-client:kube-aggregator",
	}
	return o.deployCertificate(ctx, certConfig, nil)
}
