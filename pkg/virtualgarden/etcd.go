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

	"github.com/gardener/virtual-garden/pkg/api/helper"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ETCDRoleMain is a constant for the 'main' role of etcd.
	ETCDRoleMain = "main"
	// ETCDRoleEvents is a constant for the 'events' role of etcd.
	ETCDRoleEvents = "events"
)

// DeployETCD generates a dedicated CA, two server certificates (for main and events etcd), and one client
// certificate. It deploys them as secrets together with the bootstrap config map. Finally, it creates a service for
// etcd that can be used by clients and deploys it as a stateful set with a persistent volume for its data.
func (o *operation) DeployETCD(ctx context.Context) error {
	o.log.Infof("Deploying the storage class for persistent volumes of etcd")
	if err := o.deployETCDStorageClass(ctx); err != nil {
		return err
	}

	o.log.Infof("Generating or loading CA, client certificates")
	caCertificate, caCertChecksum, err := o.deployETCDCACertificate(ctx)
	if err != nil {
		return err
	}
	_, clientCertChecksum, err := o.deployETCDClientCertificate(ctx, caCertificate)
	if err != nil {
		return err
	}

	for _, role := range []string{ETCDRoleMain, ETCDRoleEvents} {
		o.log.Infof("Deploying etcd service for role %q", role)
		if err := o.deployETCDService(ctx, role); err != nil {
			return err
		}

		o.log.Infof("Deploying etcd bootstrap configmap for role %q", role)
		configMapChecksum, err := o.deployETCDConfigMap(ctx, role)
		if err != nil {
			return err
		}

		o.log.Infof("Generating or loading server certificate for role %q", role)
		_, serverCertChecksum, err := o.deployETCDServerCertificate(ctx, caCertificate, role)
		if err != nil {
			return err
		}

		var (
			checksums = map[string]string{
				"checksum/configmap-etcd-bootstrap-config": configMapChecksum,
				"checksum/secret-etcd-ca":                  caCertChecksum,
				"checksum/secret-etcd-server":              serverCertChecksum,
				"checksum/secret-etcd-client":              clientCertChecksum,
			}

			// data volume settings
			storageCapacity  = "10Gi"
			storageClassName *string

			// backup settings
			storageProviderName string
			secretData          map[string][]byte
			environment         []corev1.EnvVar
		)

		if role == ETCDRoleMain {
			storageCapacity = "25Gi"
			storageClassName = pointer.StringPtr(ETCDStorageClassName(o.imports.VirtualGarden.ETCD))

			if helper.ETCDBackupEnabled(o.imports.VirtualGarden.ETCD) {
				o.log.Infof("Deploying etcd backup secret for role %q", role)
				storageProviderName, secretData, environment = o.backupProvider.ComputeETCDBackupConfiguration(ETCDVolumeMountPathBackupSecret)

				backupSecretChecksum, err := o.deployETCDBackupSecret(ctx, secretData)
				if err != nil {
					return err
				}
				checksums["checksum/secret-etcd-backup"] = backupSecretChecksum
			}
		}

		o.log.Infof("Deploying etcd statefulset for role %q", role)
		if err := o.deployETCDStatefulSet(ctx, role, checksums, storageCapacity, storageClassName, storageProviderName, environment); err != nil {
			return err
		}

		if helper.ETCDHVPAEnabled(o.imports.VirtualGarden.ETCD) {
			o.log.Infof("Deploying etcd HVPA for role %q if CRD is installed", role)
			if err := o.deployETCDHVPA(ctx, role); err != nil && !meta.IsNoMatchError(err) {
				return err
			}
		}
	}

	return nil
}

// DeleteETCD deletes etcd and all related resources.
func (o *operation) DeleteETCD(ctx context.Context) error {
	for _, role := range []string{ETCDRoleMain, ETCDRoleEvents} {
		if helper.ETCDHVPAEnabled(o.imports.VirtualGarden.ETCD) {
			o.log.Infof("Deleting etcd HVPA for role %q if CRD is installed", role)
			if err := o.deleteETCDHVPA(ctx, role); err != nil && !meta.IsNoMatchError(err) {
				return err
			}
		}

		o.log.Infof("Deleting etcd statefulset for role %q", role)
		if err := o.deleteETCDStatefulSet(ctx, role); err != nil {
			return err
		}

		if role == ETCDRoleMain && helper.ETCDBackupEnabled(o.imports.VirtualGarden.ETCD) {
			o.log.Infof("Deleting etcd backup secret for role %q", role)
			if err := o.deleteETCDBackupSecret(ctx); err != nil {
				return err
			}
		}

		o.log.Infof("Deleting etcd bootstrap configmap for role %q", role)
		if err := o.deleteETCDConfigMap(ctx, role); err != nil {
			return err
		}

		o.log.Infof("Deleting etcd service for role %q", role)
		if err := o.deleteETCDService(ctx, role); err != nil {
			return err
		}
	}

	o.log.Infof("Deleting all certificates related to etcd")
	if err := o.deleteETCDCertificateSecrets(ctx); err != nil {
		return err
	}

	mustKeepStorageClass, err := OtherVirtualGardensInHostingCluster(ctx, o.client, o.namespace)
	if err != nil {
		return err
	}

	if !mustKeepStorageClass {
		o.log.Infof("Deleting the storage class for persistent volumes of etcd")
		if err := o.deleteETCDStorageClass(ctx); err != nil {
			return err
		}
	}

	return nil
}

// OtherVirtualGardensInHostingCluster returns true if there are other statefulsets of virtual garden in the hosting
// cluster.
func OtherVirtualGardensInHostingCluster(ctx context.Context, c client.Client, namespace string) (bool, error) {
	statefulSetList := &appsv1.StatefulSetList{}
	if err := c.List(ctx, statefulSetList); err != nil {
		return false, err
	}

	for _, sts := range statefulSetList.Items {
		if sts.Namespace == namespace {
			continue
		}
		return true, nil
	}

	return false, nil
}

func etcdLabels(role string) map[string]string {
	return map[string]string{
		LabelKeyApp:       Prefix,
		LabelKeyComponent: "etcd",
		LabelKeyRole:      role,
	}
}
