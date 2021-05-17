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
	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (o *operation) deployKubeAPIServerPodAutoscaling(ctx context.Context) error {
	o.log.Infof("Deploying manifests for pod autoscaling for the kube-apiserver")

	if err := o.deployKubeApiServerHvpa(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeApiServerHvpa(ctx context.Context) error {
	o.log.Infof("Deploying hvpa for the kube-apiserver")

	o.emptyKubeAPIServerHvpa()

	return nil
}

func (o *operation) deleteKubeAPIServerPodAutoscaling(ctx context.Context) error {
	o.log.Infof("Deleting manifests for pod autoscaling for the kube-apiserver")

	if err := o.deleteKubeApiServerHvpa(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteKubeApiServerHvpa(ctx context.Context) error {
	o.log.Infof("Delete hvpa for the kube-apiserver")

	hvpa := o.emptyKubeAPIServerHvpa()

	if err := o.client.Delete(ctx, hvpa); client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}

func (o *operation) emptyKubeAPIServerHvpa() *hvpav1alpha1.Hvpa {
	return &hvpav1alpha1.Hvpa{ObjectMeta: metav1.ObjectMeta{Name: "virtual-garden-kube-apiserver", Namespace: o.namespace}}
}