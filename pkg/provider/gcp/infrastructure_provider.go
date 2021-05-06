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

package gcp

import (
	"fmt"

	"github.com/gardener/virtual-garden/pkg/api"

	corev1 "k8s.io/api/core/v1"
)

type infrastructureProvider struct{}

// NewInfrastructureProvider creates a new GCP infrastructure provider implementation.
func NewInfrastructureProvider() *infrastructureProvider {
	return &infrastructureProvider{}
}

func (b *infrastructureProvider) ComputeStorageClassConfiguration() (provisioner string, parameters map[string]string) {
	provisioner = "kubernetes.io/gce-pd"
	parameters = map[string]string{
		"type": "pd-ssd",
	}
	return
}

func (b *infrastructureProvider) GetLoadBalancer(service *corev1.Service) string {
	ingress := service.Status.LoadBalancer.Ingress
	if len(ingress) == 0 {
		return ""
	}

	return ingress[0].IP
}

func (b *infrastructureProvider) GetKubeAPIServerURL(kubeAPIServer *api.KubeAPIServer, _ string) string {
	return fmt.Sprintf("https://api.%s:443", kubeAPIServer.DnsAccessDomain)
}
