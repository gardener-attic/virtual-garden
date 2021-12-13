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
	_ "embed"
	"encoding/json"

	"github.com/gardener/gardener/pkg/utils"

	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apiserverv1 "k8s.io/apiserver/pkg/apis/apiserver/v1"

	// auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	KubeApiServerConfigMapAdmission   = Prefix + "-kube-apiserver-admission-config"
	KubeApiServerConfigMapAuditPolicy = "kube-apiserver-audit-policy-config"
)

func (o *operation) deployKubeAPIServerConfigMaps(ctx context.Context, checksums map[string]string) error {
	o.log.Infof("Deploying configmaps for the kube-apiserver")

	if err := o.deployKubeApiServerConfigMapAdmission(ctx, checksums); err != nil {
		return err
	}

	if err := o.deployKubeApiServerConfigMapAuditPolicy(ctx, checksums); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteKubeAPIServerConfigMaps(ctx context.Context) error {
	o.log.Infof("Deleting configmaps for the kube-apiserver")

	for _, name := range []string{
		KubeApiServerConfigMapAdmission,
		KubeApiServerConfigMapAuditPolicy,
	} {
		configmap := o.emptyConfigMap(name)
		if err := o.client.Delete(ctx, configmap); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

//go:embed resources/audit_policy.yaml
var auditPolicy []byte

func (o *operation) deployKubeApiServerConfigMapAuditPolicy(ctx context.Context, checksums map[string]string) error {
	auditPolicyYaml := string(auditPolicy)

	configMap := o.emptyConfigMap(KubeApiServerConfigMapAuditPolicy)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, configMap, func() error {
		if len(configMap.Data) == 0 {
			configMap.Data = make(map[string]string)
		}

		configMap.Data[AuditPolicyYamlKey] = auditPolicyYaml
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerAuditPolicyConfig] = utils.ComputeChecksum(configMap.Data)
	return nil
}

func (o *operation) deployKubeApiServerConfigMapAdmission(ctx context.Context, checksums map[string]string) error {
	const (
		validatingAdmissionWebhookConfigName     = "ValidatingAdmissionWebhook"
		validatingAdmissionWebhookKubeconfigPath = "/var/run/secrets/admission-kubeconfig/validating-webhook"
		mutatingAdmissionWebhookConfigName       = "MutatingAdmissionWebhook"
		mutatingAdmissionWebhookKubeconfigPath   = "/var/run/secrets/admission-kubeconfig/mutating-webhook"
	)

	controlplane := o.imports.VirtualGarden.KubeAPIServer.GardenerControlplane
	if !(o.isWebhookEnabled()) {
		return nil
	}

	configMap := o.emptyConfigMap(KubeApiServerConfigMapAdmission)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, configMap, func() error {
		configMap.Labels = map[string]string{
			LabelKeyApp:       Prefix,
			LabelKeyComponent: "kube-apiserver",
		}

		admissionConfig := apiserverv1.AdmissionConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionConfiguration",
				APIVersion: "apiserver.config.k8s.io/v1",
			},
			Plugins: []apiserverv1.AdmissionPluginConfiguration{},
		}

		if controlplane.ValidatingWebhook.Kubeconfig != "" {
			config, err := o.newAdmissionPluginConfiguration(validatingAdmissionWebhookConfigName, validatingAdmissionWebhookKubeconfigPath)
			if err != nil {
				return err
			}

			admissionConfig.Plugins = append(admissionConfig.Plugins, *config)
		}

		if controlplane.MutatingWebhook.Kubeconfig != "" {
			config, err := o.newAdmissionPluginConfiguration(mutatingAdmissionWebhookConfigName, mutatingAdmissionWebhookKubeconfigPath)
			if err != nil {
				return err
			}

			admissionConfig.Plugins = append(admissionConfig.Plugins, *config)
		}

		admissionConfigYAML, err := yaml.Marshal(admissionConfig)
		if err != nil {
			return err
		}

		if len(configMap.Data) == 0 {
			configMap.Data = make(map[string]string)
		}

		configMap.Data[ConfigurationYamlKey] = string(admissionConfigYAML)
		return nil
	})
	if err != nil {
		return err
	}

	checksums[ChecksumKeyKubeAPIServerAdmissionConfig] = utils.ComputeChecksum(configMap.Data)
	return nil
}

func (o *operation) newAdmissionPluginConfiguration(name, kubeConfigPath string) (*apiserverv1.AdmissionPluginConfiguration, error) {
	config := map[string]string{
		"apiVersion":     "apiserver.config.k8s.io/v1",
		"kind":           "WebhookAdmissionConfiguration",
		"kubeConfigFile": kubeConfigPath,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &apiserverv1.AdmissionPluginConfiguration{
		Name: name,
		Configuration: &runtime.Unknown{
			Raw: configJSON,
		},
	}, nil
}

func (o *operation) emptyConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
