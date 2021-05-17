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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (o *operation) deployMisc(ctx context.Context) error {
	o.log.Infof("Deploying misc for the kube-apiserver")

	if err := o.deployKubeAPIServerPodDisruptionBudget(ctx); err != nil {
		return err
	}

	if err := o.deployKubeAPIServerServiceAccount(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteMisc(ctx context.Context) error {

	if err := o.deleteKubeAPIServerPodDisruptionBudget(ctx); err != nil {
		return err
	}

	if err := o.deleteKubeAPIServerServiceAccount(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeAPIServerPodDisruptionBudget(ctx context.Context) error {
	o.log.Infof("Deploying PodDisruptionBudget for the kube-apiserver")

	minAvailable := intstr.FromInt(2)

	budget := o.emptyPodDisruptionBudget()

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, budget, func() error {
		budget.Spec = policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: getKubeAPIServerServiceLabels(),
			},
		}
		return nil
	})

	return err
}

func (o *operation) deployKubeAPIServerServiceAccount(ctx context.Context) error {
	o.log.Infof("Deploying service account for the kube-apiserver")

	serviceccount := o.emptyServiceAccount()

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, serviceccount, func() error {
		serviceccount.ObjectMeta.Labels = getKubeAPIServerServiceLabels()
		return nil
	})

	return err
}

func (o *operation) deleteKubeAPIServerPodDisruptionBudget(ctx context.Context) error {
	o.log.Infof("Delete PodDisruptionBudget for the kube-apiserver")

	budget := o.emptyPodDisruptionBudget()

	if err := o.client.Delete(ctx, budget); client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}

func (o *operation) deleteKubeAPIServerServiceAccount(ctx context.Context) error {
	o.log.Infof("Delete service account for the kube-apiserver")

	serviceAccount := o.emptyServiceAccount()

	if err := o.client.Delete(ctx, serviceAccount); client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}

func (o *operation) emptyPodDisruptionBudget() *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Name: "virtual-garden-kube-apiserver", Namespace: o.namespace}}
}

func (o *operation) emptyServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "virtual-garden-kube-apiserver", Namespace: o.namespace}}
}
