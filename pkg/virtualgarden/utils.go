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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/ghodss/yaml"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"

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

func deployCertificate(ctx context.Context, c client.Client, namespace string, certConfig *secretsutil.CertificateSecretConfig,
	kubeconfigGen *kubeconfigGenerator) (*secretsutil.Certificate, string, []byte, error) {
	objectKey := client.ObjectKey{Name: certConfig.Name, Namespace: namespace}

	cert, err := loadOrGenerateCertificateSecret(ctx, c, objectKey, certConfig)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Loading or generating the certificate in secret (%s/%s) failed: %w",
			objectKey.Namespace, objectKey.Name, err)
	}

	checksum, kubeconfig, err := createOrUpdateCertificateSecret(ctx, c, objectKey, cert, kubeconfigGen)
	if err != nil {
		return nil, "", nil, err
	}

	return cert, checksum, kubeconfig, nil
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

func createOrUpdateCertificateSecret(ctx context.Context, c client.Client, objectKey client.ObjectKey,
	certificate *secretsutil.Certificate, kubeconfigGenerator *kubeconfigGenerator) (string, []byte, error) {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: objectKey.Name, Namespace: objectKey.Namespace}}

	var kubeconfig []byte

	if _, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		secret.Type = corev1.SecretTypeOpaque
		if certificate.CA != nil {
			secret.Type = corev1.SecretTypeTLS
		}
		secret.Data = certificate.SecretData()

		if kubeconfigGenerator != nil {
			var tmperr error
			kubeconfig, tmperr = kubeconfigGenerator.addKubeconfigToSecretData(certificate, secret.Data)
			if tmperr != nil {
				return tmperr
			}
		}

		return nil
	}); err != nil {
		return "", nil, err
	}

	return utils.ComputeChecksum(secret.Data), kubeconfig, nil
}

type kubeconfigGenerator struct {
	user   string
	server string
}

func (k *kubeconfigGenerator) addKubeconfigToSecretData(certificate *secretsutil.Certificate, secretData map[string][]byte) ([]byte, error) {
	kubeconfig, err := yaml.Marshal(k.createKubeconfig(certificate))
	if err != nil {
		return nil, err
	}

	secretData[SecretKeyKubeconfig] = kubeconfig
	return kubeconfig, nil
}

func (k *kubeconfigGenerator) createKubeconfig(certificate *secretsutil.Certificate) *v1.Config {
	return &v1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: "virtual-garden",
		Contexts: []v1.NamedContext{
			{
				Name: "virtual-garden",
				Context: v1.Context{
					Cluster:  "virtual-garden",
					AuthInfo: k.user,
				},
			},
		},
		Clusters: []v1.NamedCluster{
			{
				Name: "virtual-garden",
				Cluster: v1.Cluster{
					Server:                   k.server,
					CertificateAuthorityData: certificate.CA.CertificatePEM,
				},
			},
		},
		AuthInfos: []v1.NamedAuthInfo{
			{
				Name: k.user,
				AuthInfo: v1.AuthInfo{
					ClientCertificateData: certificate.CertificatePEM,
					ClientKeyData:         certificate.PrivateKeyPEM,
				},
			},
		},
	}
}

func volumeWithSecretSource(volumeName, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func volumeWithConfigMapSource(volumeName, configMapName string) corev1.Volume {
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
}

func waitForDeploymentReady(ctx context.Context, c client.Client, deployment *appsv1.Deployment) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err := wait.PollImmediateUntil(10*time.Second, func() (done bool, err error) {
		if err := c.Get(ctx, client.ObjectKeyFromObject(deployment), deployment); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return false, err
			}
			return false, nil
		}

		replicas := int32(1)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}

		ready := deployment.Generation == deployment.Status.ObservedGeneration &&
			replicas == deployment.Status.ReadyReplicas &&
			replicas == deployment.Status.UpdatedReplicas &&
			replicas == deployment.Status.AvailableReplicas

		return ready, nil
	}, timeoutCtx.Done())

	if err != nil {
		return fmt.Errorf("error while waiting for deployment %s to become ready: %w", deployment.Name, err)
	}

	return nil
}

func waitForStatefulSetReady(ctx context.Context, c client.Client, statefulSet *appsv1.StatefulSet) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err := wait.PollImmediateUntil(2*time.Second, func() (done bool, err error) {
		if err := c.Get(ctx, client.ObjectKeyFromObject(statefulSet), statefulSet); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return false, err
			}
			return false, nil
		}

		replicas := int32(1)
		if statefulSet.Spec.Replicas != nil {
			replicas = *statefulSet.Spec.Replicas
		}

		ready := statefulSet.Generation == statefulSet.Status.ObservedGeneration &&
			replicas == statefulSet.Status.ReadyReplicas &&
			replicas == statefulSet.Status.UpdatedReplicas &&
			replicas == statefulSet.Status.CurrentReplicas

		return ready, nil
	}, timeoutCtx.Done())

	if err != nil {
		return fmt.Errorf("error while waiting for statefulset %s to become ready: %w", statefulSet.Name, err)
	}

	return nil
}
