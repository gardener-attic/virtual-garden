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
	"strconv"
	"strings"

	"github.com/gardener/virtual-garden/pkg/api/helper"

	"github.com/gardener/gardener/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// KubeAPIServerServiceName is a constant for the service name for the kube-apiserver of the virtual cluster.
	KubeAPIServerServiceName = "virtual-garden-kube-apiserver"

	kubeAPIServerServicePortName = "kube-apiserver"
	kubeAPIServerServicePort     = 443
)

var kubeAPIServerServiceLabels = map[string]string{
	"app":       Prefix,
	"component": "kube-apiserver",
}

// DeployKubeAPIServerService deploys the service object for the virtual garden kube-apiserver.
func (o *operation) DeployKubeAPIServerService(ctx context.Context) error {
	service := emptyKubeAPIServerService(o.namespace)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, service, func() error {
		if helper.KubeAPIServerSNIEnabled(o.imports.VirtualGarden.KubeAPIServer) {
			annotations := map[string]string{
				"dns.gardener.cloud/dnsnames": strings.Join(o.imports.VirtualGarden.KubeAPIServer.Exposure.SNI.Hostnames, ","),
			}
			if o.imports.VirtualGarden.KubeAPIServer.Exposure.SNI.DNSClass != nil {
				annotations["dns.gardener.cloud/class"] = *o.imports.VirtualGarden.KubeAPIServer.Exposure.SNI.DNSClass
			}
			if o.imports.VirtualGarden.KubeAPIServer.Exposure.SNI.TTL != nil {
				annotations["dns.gardener.cloud/ttl"] = strconv.Itoa(int(*o.imports.VirtualGarden.KubeAPIServer.Exposure.SNI.TTL))
			}
			service.Annotations = utils.MergeStringMaps(service.Annotations, annotations)
		} else {
			delete(service.Annotations, "dns.gardener.cloud/dnsnames")
			delete(service.Annotations, "dns.gardener.cloud/class")
			delete(service.Annotations, "dns.gardener.cloud/ttl")
		}

		service.Labels = utils.MergeStringMaps(service.Labels, kubeAPIServerServiceLabels)
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
		service.Spec.Selector = kubeAPIServerServiceLabels
		service.Spec.Ports = reconcileServicePorts(service.Spec.Ports, []corev1.ServicePort{
			{
				Name:       kubeAPIServerServicePortName,
				Protocol:   corev1.ProtocolTCP,
				Port:       kubeAPIServerServicePort,
				TargetPort: intstr.FromInt(kubeAPIServerServicePort),
			},
		})

		return nil
	})
	return err
}

// DeleteKubeAPIServerService deletes the service object for the virtual garden kube-apiserver.
func (o *operation) DeleteKubeAPIServerService(ctx context.Context) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyKubeAPIServerService(o.namespace)))
}

func emptyKubeAPIServerService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeAPIServerServiceName,
			Namespace: namespace,
		},
	}
}
