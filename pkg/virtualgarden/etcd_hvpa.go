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

	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ETCDHVPAName returns the name of the HVPA object for the given role.
func ETCDHVPAName(role string) string {
	return Prefix + "-etcd-" + role
}

func (o *operation) deployETCDHVPA(ctx context.Context, role string) error {
	var (
		containerScalingOff = autoscalingv1beta2.ContainerScalingModeOff
		hvpa                = emptyETCDHVPA(o.namespace, role)
	)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, hvpa, func() error {
		hvpa.Spec.Replicas = pointer.Int32Ptr(1)
		hvpa.Spec.TargetRef = &autoscalingv2beta1.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Name:       ETCDStatefulSetName(role),
		}
		hvpa.Spec.Hpa = hvpav1alpha1.HpaSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: etcdHPALabels(role),
			},
			Deploy: false,
			Template: hvpav1alpha1.HpaTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: etcdHPALabels(role),
				},
				Spec: hvpav1alpha1.HpaTemplateSpec{
					MinReplicas: pointer.Int32Ptr(1),
					MaxReplicas: 1,
					Metrics: []autoscalingv2beta1.MetricSpec{
						{
							Resource: &autoscalingv2beta1.ResourceMetricSource{
								Name:                     corev1.ResourceMemory,
								TargetAverageUtilization: pointer.Int32Ptr(80),
							},
						},
						{
							Resource: &autoscalingv2beta1.ResourceMetricSource{
								Name:                     corev1.ResourceCPU,
								TargetAverageUtilization: pointer.Int32Ptr(80),
							},
						},
					},
				},
			},
		}
		hvpa.Spec.Vpa = hvpav1alpha1.VpaSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: etcdVPALabels(role),
			},
			Deploy: true,
			Template: hvpav1alpha1.VpaTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: etcdVPALabels(role),
				},
				Spec: hvpav1alpha1.VpaTemplateSpec{
					ResourcePolicy: &autoscalingv1beta2.PodResourcePolicy{
						ContainerPolicies: []autoscalingv1beta2.ContainerResourcePolicy{
							{
								ContainerName: etcdContainerName,
								MaxAllowed: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("4"),
									corev1.ResourceMemory: resource.MustParse("30G"),
								},
								MinAllowed: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("700M"),
								},
							},
							{
								ContainerName: backupRestoreSidecarContainerName,
								Mode:          &containerScalingOff,
							},
						},
					},
				},
			},
			ScaleUp: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: pointer.StringPtr(hvpav1alpha1.UpdateModeAuto),
				},
				StabilizationDuration: pointer.StringPtr("5m"),
				MinChange: hvpav1alpha1.ScaleParams{
					CPU: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("1"),
						Percentage: pointer.Int32Ptr(80),
					},
					Memory: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("2G"),
						Percentage: pointer.Int32Ptr(80),
					},
				},
			},
			ScaleDown: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: pointer.StringPtr(hvpav1alpha1.UpdateModeOff),
				},
				StabilizationDuration: pointer.StringPtr("15m"),
				MinChange: hvpav1alpha1.ScaleParams{
					CPU: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("1"),
						Percentage: pointer.Int32Ptr(80),
					},
					Memory: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("2G"),
						Percentage: pointer.Int32Ptr(80),
					},
				},
			},
			LimitsRequestsGapScaleParams: hvpav1alpha1.ScaleParams{
				CPU: hvpav1alpha1.ChangeParams{
					Value:      pointer.StringPtr("2"),
					Percentage: pointer.Int32Ptr(40),
				},
				Memory: hvpav1alpha1.ChangeParams{
					Value:      pointer.StringPtr("3G"),
					Percentage: pointer.Int32Ptr(40),
				},
			},
		}
		hvpa.Spec.WeightBasedScalingIntervals = []hvpav1alpha1.WeightBasedScalingInterval{
			{
				VpaWeight:         hvpav1alpha1.VpaOnly,
				StartReplicaCount: 1,
				LastReplicaCount:  1,
			},
		}
		return nil
	})
	return err
}

func (o *operation) deleteETCDHVPA(ctx context.Context, role string) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyETCDHVPA(o.namespace, role)))
}

func emptyETCDHVPA(namespace, role string) *hvpav1alpha1.Hvpa {
	return &hvpav1alpha1.Hvpa{ObjectMeta: metav1.ObjectMeta{Name: ETCDHVPAName(role), Namespace: namespace}}
}

func etcdVPALabels(role string) map[string]string {
	return map[string]string{"role": "etcd-vpa-" + role}
}

func etcdHPALabels(role string) map[string]string {
	return map[string]string{"role": "etcd-hpa-" + role}
}
