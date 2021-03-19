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

	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ETCDSecretNameCACertificate is a constant for the name of a Kubernetes secret that contains the CA certificate
	// for etcd.
	ETCDSecretNameCACertificate = Prefix + "-etcd-ca"
	// ETCDSecretNameClientCertificate is a constant for the name of a Kubernetes secret that contains the client
	// certificate for etcd.
	ETCDSecretNameClientCertificate = Prefix + "-etcd-client"
)

// ETCDSecretNameServerCertificate returns the name of a Kubernetes secret that contains the server certificate for etcd
// for the given role.
func ETCDSecretNameServerCertificate(role string) string {
	return fmt.Sprintf("%s-etcd-%s-server", Prefix, role)
}

func (o *operation) deployETCDCACertificate(ctx context.Context) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       ETCDSecretNameCACertificate,
		CertType:   secretsutil.CACert,
		CommonName: Prefix + ":ca:etcd",
	}
	return o.deployCertificate(ctx, certConfig)
}

func (o *operation) deployETCDServerCertificate(ctx context.Context, caCertificate *secretsutil.Certificate, role string) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       ETCDSecretNameServerCertificate(role),
		CertType:   secretsutil.ServerClientCert,
		SigningCA:  caCertificate,
		CommonName: fmt.Sprintf("%s:server:etcd:%s", Prefix, role),
		DNSNames: []string{
			fmt.Sprintf("%s-etcd-%s-0", Prefix, role),
			fmt.Sprintf("%s-etcd-%s-client.%s", Prefix, role, o.namespace),
			fmt.Sprintf("%s-etcd-%s-client.%s.svc", Prefix, role, o.namespace),
			fmt.Sprintf("%s-etcd-%s-client.%s.svc.cluster", Prefix, role, o.namespace),
			fmt.Sprintf("%s-etcd-%s-client.%s.svc.cluster.local", Prefix, role, o.namespace),
		},
	}
	return o.deployCertificate(ctx, certConfig)
}

func (o *operation) deployETCDClientCertificate(ctx context.Context, caCertificate *secretsutil.Certificate) (*secretsutil.Certificate, string, error) {
	certConfig := &secretsutil.CertificateSecretConfig{
		Name:       ETCDSecretNameClientCertificate,
		CertType:   secretsutil.ClientCert,
		SigningCA:  caCertificate,
		CommonName: Prefix + ":client:etcd",
	}
	return o.deployCertificate(ctx, certConfig)
}

func (o *operation) deleteETCDCertificateSecrets(ctx context.Context) error {
	for _, name := range []string{
		ETCDSecretNameCACertificate,
		ETCDSecretNameServerCertificate(ETCDRoleMain),
		ETCDSecretNameServerCertificate(ETCDRoleEvents),
		ETCDSecretNameClientCertificate,
	} {
		if err := o.client.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (o *operation) deployCertificate(ctx context.Context, certConfig *secretsutil.CertificateSecretConfig) (*secretsutil.Certificate, string, error) {
	objectKey := client.ObjectKey{Name: certConfig.Name, Namespace: o.namespace}

	cert, err := loadOrGenerateCertificateSecret(ctx, o.client, objectKey, certConfig)
	if err != nil {
		return nil, "", err
	}

	checksum, err := createOrUpdateCertificateSecret(ctx, o.client, objectKey, cert)
	if err != nil {
		return nil, "", err
	}

	return cert, checksum, nil
}
