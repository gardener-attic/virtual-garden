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

	"github.com/gardener/virtual-garden/pkg/api"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (o *operation) deployKubeAPIServerDeploymentControllerManager(ctx context.Context, checksums map[string]string, basicAuthPw string) error {
	o.log.Infof("Deploying deployment virtual-garden-kube-controller-manager")

	deployment := o.emptyDeployment(KubeAPIServerDeploymentNameControllerManager)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, deployment, func() error {
		deployment.ObjectMeta.Labels = o.getKubeControllerManagerLabels()

		deployment.Spec = appsv1.DeploymentSpec{
			RevisionHistoryLimit: pointer.Int32Ptr(0),
			Replicas:             pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: o.getKubeControllerManagerLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: o.getKubeControllerManagerAnnotations(checksums),
					Labels: map[string]string{
						LabelKeyApp:                        Prefix,
						LabelKeyComponent:                  kubeControllerManager,
						"networking.gardener.cloud/to-dns": LabelValueAllowed,
						"networking.gardener.cloud/to-gardener-apiserver":     LabelValueAllowed,
						"networking.gardener.cloud/to-virtual-kube-apiserver": LabelValueAllowed,
					},
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken:  pointer.BoolPtr(false),
					PriorityClassName:             "garden-controlplane",
					Containers:                    o.getKubeControllerManagerContainers(),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					TerminationGracePeriodSeconds: pointer.Int64Ptr(30),
					Volumes:                       o.getKubeControllerManagerVolumes(),
				},
			},
		}
		return nil
	})

	return err
}

func (o *operation) getKubeControllerManagerAnnotations(checksums map[string]string) map[string]string {
	annotations := o.addChecksumsToAnnotations(checksums, []string{
		ChecksumKeyKubeAPIServerCA,
		ChecksumKeyKubeControllerManagerClient,
		ChecksumKeyServiceAccountKey,
	})
	return annotations
}

func (o *operation) getKubeControllerManagerLabels() map[string]string {
	return map[string]string{
		LabelKeyApp:       Prefix,
		LabelKeyComponent: kubeControllerManager,
	}
}

func (o *operation) getKubeControllerManagerContainers() []corev1.Container {
	return []corev1.Container{
		{
			Name:            kubeControllerManager,
			Image:           "eu.gcr.io/sap-se-gcr-k8s-public/k8s_gcr_io/kube-controller-manager:v1.18.14",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         o.getKubeControllerManagerCommand(),
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/healthz",
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: 10257},
						Scheme: corev1.URISchemeHTTPS,
					},
				},
				InitialDelaySeconds: 15,
				TimeoutSeconds:      15,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				FailureThreshold:    2,
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1024Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			VolumeMounts:             o.getKubeControllerManagerVolumeMounts(),
		},
	}
}

func (o *operation) getKubeControllerManagerCommand() []string {
	hpaConfig := o.getKubeControllerManagerHPAConfig()

	return []string{
		"/usr/local/bin/kube-controller-manager",
		"--authentication-kubeconfig=/srv/kubernetes/controller-manager/kubeconfig",
		"--authorization-kubeconfig=/srv/kubernetes/controller-manager/kubeconfig",
		"--cluster-signing-cert-file=/srv/kubernetes/ca/ca.crt",
		"--cluster-signing-key-file=/srv/kubernetes/ca/ca.key",
		"--controllers=namespace,serviceaccount,serviceaccount-token,clusterrole-aggregation,garbagecollector,csrapproving,csrcleaner,csrsigning,bootstrapsigner,tokencleaner,resourcequota",
		"--concurrent-gc-syncs=250",
		"--concurrent-namespace-syncs=100",
		"--concurrent-resource-quota-syncs=100",
		"--concurrent-serviceaccount-token-syncs=100",
		"--kubeconfig=/srv/kubernetes/controller-manager/kubeconfig",
		"--root-ca-file=/srv/kubernetes/ca/ca.crt",
		"--service-account-private-key-file=/srv/kubernetes/service-account-key/service_account.key",
		"--use-service-account-credentials=true",
		fmt.Sprintf("--horizontal-pod-autoscaler-downscale-stabilization=%s", hpaConfig.DownscaleStabilization),
		fmt.Sprintf("--horizontal-pod-autoscaler-initial-readiness-delay=%s", hpaConfig.ReadinessDelay),
		fmt.Sprintf("--horizontal-pod-autoscaler-cpu-initialization-period=%s", hpaConfig.CpuInitializationPeriod),
		fmt.Sprintf("--horizontal-pod-autoscaler-sync-period=%s", hpaConfig.SyncPeriod),
		fmt.Sprintf("--horizontal-pod-autoscaler-tolerance=%s", hpaConfig.Tolerance),
		"--v=5",
	}
}

func (o *operation) getKubeControllerManagerVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      volumeNameKubeAPIServerCA,
			MountPath: "/srv/kubernetes/ca",
		},
		{
			Name:      volumeNameKubeControllerManager,
			MountPath: "/srv/kubernetes/controller-manager",
		},
		{
			Name:      volumeNameServiceAccountKey,
			MountPath: "/srv/kubernetes/service-account-key",
		},
	}
}

func (o *operation) getKubeControllerManagerVolumes() []corev1.Volume {
	return []corev1.Volume{
		volumeWithSecretSource(volumeNameKubeAPIServerCA, KubeApiServerSecretNameApiServerCACertificate),
		volumeWithSecretSource(volumeNameKubeControllerManager, KubeApiServerSecretNameKubeControllerManagerCertificate),
		volumeWithSecretSource(volumeNameServiceAccountKey, KubeApiServerSecretNameServiceAccountKey),
	}
}

func (o *operation) getKubeControllerManagerHPAConfig() *api.HorizontalPodAutoscaler {
	// Start with the default values
	config := api.HorizontalPodAutoscaler{
		DownscaleStabilization:  "30m0s",
		ReadinessDelay:          "30s",
		CpuInitializationPeriod: "5m0s",
		SyncPeriod:              "30s",
		Tolerance:               "0.1",
	}

	// Overwrite with imported values
	importedConfig := o.imports.VirtualGarden.KubeAPIServer.HorizontalPodAutoscaler
	if importedConfig != nil {
		if len(importedConfig.DownscaleStabilization) != 0 {
			config.DownscaleStabilization = importedConfig.DownscaleStabilization
		}
		if len(importedConfig.ReadinessDelay) != 0 {
			config.ReadinessDelay = importedConfig.ReadinessDelay
		}
		if len(importedConfig.CpuInitializationPeriod) != 0 {
			config.CpuInitializationPeriod = importedConfig.CpuInitializationPeriod
		}
		if len(importedConfig.SyncPeriod) != 0 {
			config.SyncPeriod = importedConfig.SyncPeriod
		}
		if len(importedConfig.Tolerance) != 0 {
			config.Tolerance = importedConfig.Tolerance
		}
	}

	return &config
}
