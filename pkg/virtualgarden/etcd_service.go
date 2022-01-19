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

	"github.com/gardener/gardener/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	etcdServiceClientPortName               = "client"
	etcdServiceClientPort                   = 2379
	etcdServiceBackupRestoreSidecarPortName = "backup-client"
	etcdServiceBackupRestoreSidecarPort     = 8080
)

// ETCDServiceName returns the name of the etcd server for the given role.
func ETCDServiceName(role string) string {
	return Prefix + "-etcd-" + role + "-client"
}

func (o *operation) deployETCDService(ctx context.Context, role string) error {
	service := emptyETCDService(o.namespace, role)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, service, func() error {
		service.Labels = utils.MergeStringMaps(service.Labels, etcdLabels(role))
		service.Spec.Type = corev1.ServiceTypeClusterIP
		service.Spec.SessionAffinity = corev1.ServiceAffinityNone
		service.Spec.Selector = etcdLabels(role)
		service.Spec.Ports = reconcileServicePorts(service.Spec.Ports, []corev1.ServicePort{
			{
				Name:       etcdServiceClientPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       etcdServiceClientPort,
				TargetPort: intstr.FromInt(etcdServiceClientPort),
			},
			{
				Name:       etcdServiceBackupRestoreSidecarPortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       etcdServiceBackupRestoreSidecarPort,
				TargetPort: intstr.FromInt(etcdServiceBackupRestoreSidecarPort),
			},
		})
		return nil
	})

	// export the etcd Url
	if role == ETCDRoleMain {
		o.exports.EtcdUrl = fmt.Sprintf("%s.%s.svc:%d", service.Name, service.Namespace, etcdServiceClientPort)
	}

	return err
}

func (o *operation) deleteETCDService(ctx context.Context, role string) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyETCDService(o.namespace, role)))
}

func emptyETCDService(namespace, role string) *corev1.Service {
	return &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: ETCDServiceName(role), Namespace: namespace}}
}
