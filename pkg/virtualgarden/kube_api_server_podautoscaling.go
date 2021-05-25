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

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	VirtualGardenKubeControllerManagerName = "virtual-garden-kube-controller-manager"
)

func (o *operation) deployKubeAPIServerPodAutoscaling(ctx context.Context) error {
	o.log.Infof("Deploying manifests for pod autoscaling for the kube-apiserver")

	if err := o.deployKubeApiServerHvpa(ctx); err != nil {
		return err
	}

	if err := o.deployKubeApiServerVpa(ctx); err != nil {
		return err
	}

	if err := o.deployKubeApiServerVpaForControllerManager(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteKubeAPIServerPodAutoscaling(ctx context.Context) error {
	o.log.Infof("Deleting manifests for pod autoscaling for the kube-apiserver")

	if err := o.deleteKubeApiServerHvpa(ctx); err != nil {
		return err
	}

	if err := o.deleteKubeApiServerVpa(ctx); err != nil {
		return err
	}

	if err := o.deleteKubeApiServerVpaForControllerManager(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeApiServerVpaForControllerManager(ctx context.Context) error {
	o.log.Infof("Deploying vpa for controller manager for the kube-apiserver")
	if o.imports.VirtualGarden.KubeAPIServer.HVPAEnabled {
		return nil
	}

	vpa := o.emptyKubeAPIServerVpa(VirtualGardenKubeControllerManagerName)

	auto := autoscalingv1beta2.UpdateModeAuto

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, vpa, func() error {
		vpa.Spec = autoscalingv1beta2.VerticalPodAutoscalerSpec{
			ResourcePolicy: &autoscalingv1beta2.PodResourcePolicy{
				ContainerPolicies: []autoscalingv1beta2.ContainerResourcePolicy{
					{
						ContainerName: "kube-controller-manager",
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("10G"),
							corev1.ResourceCPU:    resource.MustParse("5"),
						},
						MinAllowed: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("400M"),
							corev1.ResourceCPU:    resource.MustParse("400m"),
						},
					},
				},
			},
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       VirtualGardenKubeControllerManagerName,
				APIVersion: "apps/v1",
			},
			UpdatePolicy: &autoscalingv1beta2.PodUpdatePolicy{
				UpdateMode: &auto,
			},
		}

		return nil
	})

	return err
}

func (o *operation) deleteKubeApiServerVpaForControllerManager(ctx context.Context) error {
	o.log.Infof("Delete vpa for controller manager for the kube-apiserver")

	hvpa := o.emptyKubeAPIServerVpa(VirtualGardenKubeControllerManagerName)

	if err := o.client.Delete(ctx, hvpa); client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeApiServerVpa(ctx context.Context) error {
	o.log.Infof("Deploying vpa for the kube-apiserver")
	if o.imports.VirtualGarden.KubeAPIServer.HVPAEnabled {
		return nil
	}

	vpa := o.emptyKubeAPIServerVpa(KubeAPIServerServiceName)

	auto := autoscalingv1beta2.UpdateModeAuto

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, vpa, func() error {
		vpa.Spec = autoscalingv1beta2.VerticalPodAutoscalerSpec{
			ResourcePolicy: &autoscalingv1beta2.PodResourcePolicy{
				ContainerPolicies: []autoscalingv1beta2.ContainerResourcePolicy{
					{
						ContainerName: "kube-apiserver",
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("25G"),
							corev1.ResourceCPU:    resource.MustParse("8"),
						},
						MinAllowed: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("400M"),
							corev1.ResourceCPU:    resource.MustParse("400m"),
						},
					},
				},
			},
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       KubeAPIServerServiceName,
				APIVersion: "apps/v1",
			},
			UpdatePolicy: &autoscalingv1beta2.PodUpdatePolicy{
				UpdateMode: &auto,
			},
		}

		return nil
	})

	return err
}

func (o *operation) deleteKubeApiServerVpa(ctx context.Context) error {
	o.log.Infof("Delete vpa for the kube-apiserver")

	hvpa := o.emptyKubeAPIServerVpa(KubeAPIServerServiceName)

	if err := o.client.Delete(ctx, hvpa); client.IgnoreNotFound(err) != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeApiServerHvpa(ctx context.Context) error {
	o.log.Infof("Deploying hvpa for the kube-apiserver")
	if !o.imports.VirtualGarden.KubeAPIServer.HVPAEnabled || o.imports.VirtualGarden.KubeAPIServer.Replicas == 0 {
		return nil
	}

	hvpaConfig := o.imports.VirtualGarden.KubeAPIServer.HVPA

	maxReplicas := hvpaConfig.GetMaxReplicas(6)
	minReplicas := hvpaConfig.GetMinReplicas()

	hvpa := o.emptyKubeAPIServerHvpa()

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, hvpa, func() error {
		hvpa.Spec.Replicas = pointer.Int32Ptr(1)

		hvpa.Spec.MaintenanceTimeWindow = nil
		if hvpaConfig != nil {
			hvpa.Spec.MaintenanceTimeWindow = o.imports.VirtualGarden.KubeAPIServer.HVPA.MaintenanceWindow
		}

		hvpa.Spec.Hpa = hvpav1alpha1.HpaSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"role": "virtual-gardener-apiserver-hpa",
				},
			},
			Deploy: true,
			ScaleUp: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: pointer.StringPtr(hvpav1alpha1.UpdateModeAuto),
				},
			},
			ScaleDown: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: pointer.StringPtr(hvpav1alpha1.UpdateModeAuto),
				},
			},
			Template: hvpav1alpha1.HpaTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"role": "virtual-gardener-apiserver-hpa",
					},
				},
				Spec: hvpav1alpha1.HpaTemplateSpec{
					MinReplicas: minReplicas,
					MaxReplicas: maxReplicas,
					Metrics: []autoscalingv2beta1.MetricSpec{
						{
							Resource: &autoscalingv2beta1.ResourceMetricSource{
								Name:                     corev1.ResourceMemory,
								TargetAverageUtilization: hvpaConfig.GetTargetAverageUtilizationMemory(80),
							},
						},
						{
							Resource: &autoscalingv2beta1.ResourceMetricSource{
								Name:                     corev1.ResourceCPU,
								TargetAverageUtilization: hvpaConfig.GetTargetAverageUtilizationCpu(80),
							},
						},
					},
				},
			},
		}
		hvpa.Spec.Vpa = hvpav1alpha1.VpaSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"role": "virtual-gardener-apiserver-vpa",
				},
			},
			Deploy: true,
			ScaleUp: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: hvpaConfig.GetVpaScaleUpMode(hvpav1alpha1.UpdateModeAuto),
				},
				MinChange: hvpaConfig.GetVpaScaleUpMinChange(hvpav1alpha1.ScaleParams{
					CPU: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("300m"),
						Percentage: pointer.Int32Ptr(80),
					},
					Memory: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("200M"),
						Percentage: pointer.Int32Ptr(80),
					},
				}),
				StabilizationDuration: hvpaConfig.GetVpaScaleUpStabilisationDuration("3m"),
			},
			ScaleDown: hvpav1alpha1.ScaleType{
				UpdatePolicy: hvpav1alpha1.UpdatePolicy{
					UpdateMode: hvpaConfig.GetVpaScaleDownMode(hvpav1alpha1.UpdateModeAuto),
				},
				MinChange: hvpaConfig.GetVpaScaleDownMinChange(hvpav1alpha1.ScaleParams{
					CPU: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("600m"),
						Percentage: pointer.Int32Ptr(80),
					},
					Memory: hvpav1alpha1.ChangeParams{
						Value:      pointer.StringPtr("600M"),
						Percentage: pointer.Int32Ptr(80),
					},
				}),
				StabilizationDuration: hvpaConfig.GetVpaScaleDownStabilisationDuration("15m"),
			},
			LimitsRequestsGapScaleParams: hvpaConfig.GetLimitsRequestsGapScaleParams(hvpav1alpha1.ScaleParams{
				CPU: hvpav1alpha1.ChangeParams{
					Value:      pointer.StringPtr("1"),
					Percentage: pointer.Int32Ptr(40),
				},
				Memory: hvpav1alpha1.ChangeParams{
					Value:      pointer.StringPtr("1G"),
					Percentage: pointer.Int32Ptr(40),
				},
			}),
			Template: hvpav1alpha1.VpaTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"role": "virtual-gardener-apiserver-vpa",
					},
				},
				Spec: hvpav1alpha1.VpaTemplateSpec{
					ResourcePolicy: &autoscalingv1beta2.PodResourcePolicy{
						ContainerPolicies: []autoscalingv1beta2.ContainerResourcePolicy{
							{
								ContainerName: kubeAPIServerContainerName,
								MaxAllowed: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("8"),
									corev1.ResourceMemory: resource.MustParse("25G"),
								},
								MinAllowed: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("400M"),
								},
							},
						},
					},
				},
			},
		}

		hvpa.Spec.WeightBasedScalingIntervals = []hvpav1alpha1.WeightBasedScalingInterval{}
		if maxReplicas > *minReplicas {
			hvpa.Spec.WeightBasedScalingIntervals = append(hvpa.Spec.WeightBasedScalingIntervals, hvpav1alpha1.WeightBasedScalingInterval{
				VpaWeight:         hvpav1alpha1.HpaOnly,
				StartReplicaCount: *minReplicas,
				LastReplicaCount:  maxReplicas - 1,
			})
		}
		hvpa.Spec.WeightBasedScalingIntervals = append(hvpa.Spec.WeightBasedScalingIntervals, hvpav1alpha1.WeightBasedScalingInterval{
			VpaWeight:         hvpav1alpha1.VpaOnly,
			StartReplicaCount: maxReplicas,
			LastReplicaCount:  maxReplicas,
		})

		hvpa.Spec.TargetRef = &autoscalingv2beta1.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       KubeAPIServerServiceName,
		}

		return nil
	})

	return err
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
	return &hvpav1alpha1.Hvpa{ObjectMeta: metav1.ObjectMeta{Name: KubeAPIServerServiceName, Namespace: o.namespace}}
}

func (o *operation) emptyKubeAPIServerVpa(name string) *autoscalingv1beta2.VerticalPodAutoscaler {
	return &autoscalingv1beta2.VerticalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
