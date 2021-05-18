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
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (o *operation) deployDeployments(ctx context.Context) error {
	o.log.Infof("Deploying deployments for the kube-apiserver")

	if err := o.deployKubeAPIServerDeployment(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteDeployments(ctx context.Context) error {
	o.log.Infof("Deleting deployments for the kube-apiserver")

	if err := o.deleteKubeAPIServerDeployment(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeAPIServerDeployment(ctx context.Context) error {
	o.log.Infof("Deploying PodDisruptionBudget for the kube-apiserver")

	deployment := o.emptyDeployment(KubeAPIServerServiceName)

	// compute particular values
	apiServerImports := o.imports.VirtualGarden.KubeAPIServer

	replicas := pointer.Int32Ptr(int32(apiServerImports.Replicas))
	if apiServerImports.HVPAEnabled {
		replicas = apiServerImports.HVPA.GetMinReplicas()
	}

	annotations, err := o.computeApiServerAnnotations(ctx)
	if err != nil {
		return err
	}

	// create/update
	_, err = controllerutil.CreateOrUpdate(ctx, o.client, deployment, func() error {
		deployment.ObjectMeta.Labels = getKubeAPIServerServiceLabels()

		deployment.Spec = appsv1.DeploymentSpec{
			RevisionHistoryLimit: pointer.Int32Ptr(0),
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: getKubeAPIServerServiceLabels(),
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
				Spec:       corev1.PodSpec{},
			},
		}
		return nil
	})

	return err
}

func (o *operation) computeApiServerAnnotations(ctx context.Context) (map[string]string, error) {
	result := map[string]string{}



	return result, nil
}

func (o *operation) emptyDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
