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

	"github.com/gardener/virtual-garden/pkg/api"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ETCDStorageClassName returns the storage class name for etcd.
func ETCDStorageClassName(etcd *api.ETCD) string {
	if etcd != nil && etcd.StorageClassName != nil {
		return *etcd.StorageClassName
	}
	return Prefix + ".gardener.cloud-fast"
}

func (o *operation) deployETCDStorageClass(ctx context.Context) error {
	var (
		provisioner, parameters = o.infrastructureProvider.ComputeStorageClassConfiguration()
		storageClass            = o.emptyETCDStorageClass()
	)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, storageClass, func() error {
		storageClass.AllowVolumeExpansion = pointer.BoolPtr(true)
		storageClass.Provisioner = provisioner
		storageClass.Parameters = parameters
		return nil
	})
	return err
}

func (o *operation) deleteETCDStorageClass(ctx context.Context) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, o.emptyETCDStorageClass()))
}

func (o *operation) emptyETCDStorageClass() *storagev1.StorageClass {
	return &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: ETCDStorageClassName(o.imports.VirtualGarden.ETCD)}}
}
