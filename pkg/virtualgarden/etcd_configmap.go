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
	"bytes"
	"context"
	_ "embed"
	"text/template"

	"github.com/gardener/gardener/pkg/utils"
	secretutils "github.com/gardener/gardener/pkg/utils/secrets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// ETCDConfigMapDataKeyBootstrapScript is a constant for a key in a configmap data that contains the bootstrap
	// script.
	ETCDConfigMapDataKeyBootstrapScript = "bootstrap.sh"
	// ETCDConfigMapDataKeyConfiguration is a constant for a key in a configmap data that contains the etcd config.
	ETCDConfigMapDataKeyConfiguration = "etcd.conf.yml"

	etcdCertificatesVolumeMountPath      = "/var/etcd/ssl"
	etcdCACertificateVolumeMountPath     = etcdCertificatesVolumeMountPath + "/ca"
	etcdServerCertificateVolumeMountPath = etcdCertificatesVolumeMountPath + "/server"
	etcdClientCertificateVolumeMountPath = etcdCertificatesVolumeMountPath + "/client"

	etcdDataVolumeMountPath = "/var/etcd/data"
	etcdDataDir             = etcdDataVolumeMountPath + "/new.etcd"
)

// ETCDConfigMapName returns the name of the etcd bootstrap configmap for the given role.
func ETCDConfigMapName(role string) string {
	return Prefix + "-etcd-" + role + "-bootstrap"
}

var (
	etcdBootstrapScriptTemplate *template.Template
	etcdConfigurationTemplate   *template.Template
)

var (
	//go:embed templates/etcd-bootstrap.sh.gtpl
	bootstrapScriptTemplate string
	//go:embed templates/etcd-config.yaml.gtpl
	configurationTemplate string
)

func init() {
	var err error

	etcdBootstrapScriptTemplate, err = template.New(ETCDConfigMapDataKeyConfiguration).Parse(bootstrapScriptTemplate)
	if err != nil {
		panic(err)
	}

	etcdConfigurationTemplate, err = template.New(ETCDConfigMapDataKeyConfiguration).Parse(configurationTemplate)
	if err != nil {
		panic(err)
	}
}

func (o *operation) deployETCDConfigMap(ctx context.Context, role string) (string, error) {
	var bootstrapScript bytes.Buffer
	if err := etcdBootstrapScriptTemplate.Execute(&bootstrapScript, map[string]interface{}{
		"BackupRestoreSidecarServicePort": etcdServiceBackupRestoreSidecarPort,
		"DataDir":                         etcdDataVolumeMountPath,
	}); err != nil {
		return "", err
	}

	var configuration bytes.Buffer
	if err := etcdConfigurationTemplate.Execute(&configuration, map[string]interface{}{
		"Role":            role,
		"DataDir":         etcdDataDir,
		"CACertPath":      etcdCACertificateVolumeMountPath + "/" + secretutils.DataKeyCertificateCA,
		"ServerCertPath":  etcdServerCertificateVolumeMountPath + "/" + secretutils.DataKeyCertificate,
		"ServerKeyPath":   etcdServerCertificateVolumeMountPath + "/" + secretutils.DataKeyPrivateKey,
		"ETCDServicePort": etcdServiceClientPort,
	}); err != nil {
		return "", err
	}

	bootstrapConfigMap := emptyETCDConfigMap(o.namespace, role)
	if _, err := controllerutil.CreateOrUpdate(ctx, o.client, bootstrapConfigMap, func() error {
		bootstrapConfigMap.Labels = utils.MergeStringMaps(bootstrapConfigMap.Labels, etcdLabels(role))
		bootstrapConfigMap.Data = map[string]string{
			ETCDConfigMapDataKeyBootstrapScript: bootstrapScript.String(),
			ETCDConfigMapDataKeyConfiguration:   configuration.String(),
		}
		return nil
	}); err != nil {
		return "", err
	}

	return utils.ComputeChecksum(bootstrapConfigMap.Data), nil
}

func (o *operation) deleteETCDConfigMap(ctx context.Context, role string) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyETCDConfigMap(o.namespace, role)))
}

func emptyETCDConfigMap(namespace, role string) *corev1.ConfigMap {
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: ETCDConfigMapName(role), Namespace: namespace}}
}
