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

	"github.com/gardener/gardener/pkg/utils"
	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func reconcileServicePorts(existingPorts []corev1.ServicePort, desiredPorts []corev1.ServicePort) []corev1.ServicePort {
	var out []corev1.ServicePort

	for _, desiredPort := range desiredPorts {
		var port corev1.ServicePort

		for _, existingPort := range existingPorts {
			if existingPort.Name == desiredPort.Name {
				port = existingPort
				break
			}
		}

		port.Name = desiredPort.Name
		port.Protocol = desiredPort.Protocol
		port.Port = desiredPort.Port
		port.TargetPort = desiredPort.TargetPort

		out = append(out, port)
	}

	return out
}

func loadOrGenerateCertificateSecret(ctx context.Context, c client.Client, objectKey client.ObjectKey, certificateConfig *secretsutil.CertificateSecretConfig) (*secretsutil.Certificate, error) {
	secret := &corev1.Secret{}
	if err := c.Get(ctx, objectKey, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		return certificateConfig.GenerateCertificate()
	}

	dataKeyPrivateKey, dataKeyCertificate := secretsutil.DataKeyPrivateKey, secretsutil.DataKeyCertificate
	if certificateConfig.CertType == secretsutil.CACert {
		dataKeyPrivateKey, dataKeyCertificate = secretsutil.DataKeyPrivateKeyCA, secretsutil.DataKeyCertificateCA
	}

	certificate, err := secretsutil.LoadCertificate(objectKey.Name, secret.Data[dataKeyPrivateKey], secret.Data[dataKeyCertificate])
	if err != nil {
		return nil, err
	}
	certificate.CA = certificateConfig.SigningCA

	return certificate, nil
}

func createOrUpdateCertificateSecret(ctx context.Context, c client.Client, objectKey client.ObjectKey, certificate *secretsutil.Certificate) (string, error) {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: objectKey.Name, Namespace: objectKey.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		secret.Type = corev1.SecretTypeOpaque
		if certificate.CA != nil {
			secret.Type = corev1.SecretTypeTLS
		}
		secret.Data = certificate.SecretData()
		return nil
	}); err != nil {
		return "", err
	}

	return utils.ComputeChecksum(secret.Data), nil
}
