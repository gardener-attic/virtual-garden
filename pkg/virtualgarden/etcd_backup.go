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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// ETCDSecretNameBackup is a constant for the name of a Kubernetes secret that contains the backup secret for the
	// main etcd.
	ETCDSecretNameBackup = Prefix + "-etcd-" + ETCDRoleMain + "-backup"
	// ETCDVolumeMountPathBackupSecret is a constant for the mount path of the etcd backup secret volume.
	ETCDVolumeMountPathBackupSecret = "/var/etcd/backup"

	etcdVolumeNameBackupSecret = "backup-credentials"
)

// DeployBackupBucket deploys a new backup bucket that will store the data of the main etcd.
func (o *operation) DeployBackupBucket(ctx context.Context) error {
	return o.backupProvider.CreateBucket(ctx)
}

// DeleteBackupBucket deletes the configured backup bucket that will store the data of the main etcd.
func (o *operation) DeleteBackupBucket(ctx context.Context) error {
	return o.backupProvider.DeleteBucket(ctx)
}

func (o *operation) deployETCDBackupSecret(ctx context.Context, secretData map[string][]byte) (string, error) {
	backupSecret := emptyETCDBackupSecret(o.namespace)

	if _, err := controllerutil.CreateOrUpdate(ctx, o.client, backupSecret, func() error {
		backupSecret.Type = corev1.SecretTypeOpaque
		backupSecret.Data = secretData
		return nil
	}); err != nil {
		return "", err
	}

	return utils.ComputeChecksum(backupSecret.Data), nil
}

func (o *operation) deleteETCDBackupSecret(ctx context.Context) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyETCDBackupSecret(o.namespace)))
}

func emptyETCDBackupSecret(namespace string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: ETCDSecretNameBackup, Namespace: namespace}}
}
