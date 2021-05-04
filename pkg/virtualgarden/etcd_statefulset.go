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
	secretsutil "github.com/gardener/gardener/pkg/utils/secrets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	etcdContainerName                 = "etcd"
	backupRestoreSidecarContainerName = "backup-restore"

	etcdConfigMapVolumeName         = "bootstrap-config"
	etcdCACertificateVolumeName     = "ca-cert"
	etcdServerCertificateVolumeName = "server-cert"
	etcdClientCertificateVolumeName = "client-cert"

	etcdConfigMapVolumeMountPath = "/bootstrap"

	etcdContainerPortServerName       = "server"
	etcdContainerPortClientName       = "client"
	etcdContainerPortServer     int32 = 2380
	etcdContainerPortClient     int32 = 2379
)

// ETCDStatefulSetName returns the name of the etcd statefulset for the given role.
func ETCDStatefulSetName(role string) string {
	return Prefix + "-etcd-" + role
}

// ETCDPersistentVolumeClaimName returns the name of the etcd persistent volume claim for the given role.
func ETCDPersistentVolumeClaimName(role string) string {
	return fmt.Sprintf("%s-%s-0", ETCDDataVolumeName(role), ETCDStatefulSetName(role))
}

// ETCDDataVolumeName returns the name of the etcd persistent volume claim for the given role.
func ETCDDataVolumeName(role string) string {
	if role == ETCDRoleMain {
		return fmt.Sprintf("%s-%s-etcd", role, Prefix)
	}
	return ETCDStatefulSetName(role)
}

func (o *operation) deployETCDStatefulSet(
	ctx context.Context,
	role string,
	checksums map[string]string,
	storageCapacity string,
	storageClassName *string,
	storageProviderName string,
	environment []corev1.EnvVar,
) error {
	sts := emptyETCDStatefulSet(o.namespace, role)

	var (
		backupConfigParameters []string
		backupEnvironment      []corev1.EnvVar
		backupVolumes          []corev1.Volume
		backupVolumeMounts     []corev1.VolumeMount
	)

	if storageProviderName != "" {
		backupConfigParameters = []string{
			"--schedule=0 */24 * * *",
			"--defragmentation-schedule=0 1 * * *",
			"--storage-provider=" + storageProviderName,
			"--store-prefix=" + sts.Name,
			"--delta-snapshot-period=5m",
			"--delta-snapshot-memory-limit=104857600", // 100 MB
			"--embedded-etcd-quota-bytes=8589934592",  // 8 GB
		}
		backupEnvironment = append([]corev1.EnvVar{
			{
				Name:  "STORAGE_CONTAINER",
				Value: o.imports.VirtualGarden.ETCD.Backup.BucketName,
			},
		}, environment...)
		backupVolumes = []corev1.Volume{
			{
				Name: etcdVolumeNameBackupSecret,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: ETCDSecretNameBackup,
					},
				},
			},
		}
		backupVolumeMounts = []corev1.VolumeMount{
			{
				Name:      etcdVolumeNameBackupSecret,
				MountPath: ETCDVolumeMountPathBackupSecret,
			},
		}
	}

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, sts, func() error {
		sts.Labels = utils.MergeStringMaps(sts.Labels, etcdLabels(role))
		sts.Spec.Replicas = pointer.Int32Ptr(1)
		sts.Spec.Selector = &metav1.LabelSelector{MatchLabels: etcdLabels(role)}
		sts.Spec.ServiceName = ETCDServiceName(role)
		sts.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType}
		sts.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: checksums,
				Labels:      etcdLabels(role),
			},
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{
												Prefix,
											},
										},
										{
											Key:      "component",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{
												"etcd",
											},
										},
									},
								},
								TopologyKey:   corev1.LabelHostname,
							},
						},
					},
				},
				PriorityClassName: "garden-controlplane",
				Containers: []corev1.Container{
					{
						Name:            etcdContainerName,
						Image:           "eu.gcr.io/sap-se-gcr-k8s-public/quay_io/coreos/etcd:v3.3.17",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{etcdConfigMapVolumeMountPath + "/" + ETCDConfigMapDataKeyBootstrapScript},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/healthz",
									Port:   intstr.FromInt(etcdServiceBackupRestoreSidecarPort),
									Scheme: corev1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: int32(5),
							PeriodSeconds:       int32(5),
						},
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"/bin/sh",
										"-ec",
										"ETCDCTL_API=3",
										"etcdctl",
										fmt.Sprintf("--cacert=%s/%s", etcdCACertificateVolumeMountPath, secretsutil.DataKeyCertificateCA),
										fmt.Sprintf("--cert=%s/%s", etcdClientCertificateVolumeMountPath, secretsutil.DataKeyCertificate),
										fmt.Sprintf("--key=%s/%s", etcdClientCertificateVolumeMountPath, secretsutil.DataKeyPrivateKey),
										fmt.Sprintf("--endpoints=https://%s-0:%d", sts.Name, etcdServiceClientPort),
										"get",
										"foo",
									},
								},
							},
							InitialDelaySeconds: int32(15),
							PeriodSeconds:       int32(5),
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          etcdContainerPortServerName,
								ContainerPort: etcdContainerPortServer,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          etcdContainerPortClientName,
								ContainerPort: etcdContainerPortClient,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("500Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("8Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      ETCDDataVolumeName(role),
								MountPath: etcdDataVolumeMountPath,
							},
							{
								Name:      etcdConfigMapVolumeName,
								MountPath: etcdConfigMapVolumeMountPath,
							},
							{
								Name:      etcdCACertificateVolumeName,
								MountPath: etcdCACertificateVolumeMountPath,
							},
							{
								Name:      etcdServerCertificateVolumeName,
								MountPath: etcdServerCertificateVolumeMountPath,
							},
							{
								Name:      etcdClientCertificateVolumeName,
								MountPath: etcdClientCertificateVolumeMountPath,
							},
						},
					},
					{
						Name:            backupRestoreSidecarContainerName,
						Image:           "eu.gcr.io/sap-se-gcr-k8s-public/eu_gcr_io/gardener-project/gardener/etcdbrctl:v0.9.1",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command: append([]string{
							"etcdbrctl",
							"server",
							fmt.Sprintf("--data-dir=%s", etcdDataDir),
							fmt.Sprintf("--cacert=%s/%s", etcdCACertificateVolumeMountPath, secretsutil.DataKeyCertificateCA),
							fmt.Sprintf("--cert=%s/%s", etcdClientCertificateVolumeMountPath, secretsutil.DataKeyCertificate),
							fmt.Sprintf("--key=%s/%s", etcdClientCertificateVolumeMountPath, secretsutil.DataKeyPrivateKey),
							"--insecure-transport=false",
							"--insecure-skip-tls-verify=false",
							fmt.Sprintf("--endpoints=https://%s-0:%d", sts.Name, etcdServiceClientPort),
							"--etcd-connection-timeout=5m",
							"--garbage-collection-period=12h",
							fmt.Sprintf("--snapstore-temp-directory=%s/temp", etcdDataVolumeMountPath),
						}, backupConfigParameters...),
						Ports: []corev1.ContainerPort{
							{
								Name:          etcdContainerPortServerName,
								ContainerPort: int32(etcdServiceBackupRestoreSidecarPort),
								Protocol:      corev1.ProtocolTCP,
							},
						},
						Env: backupEnvironment,
						VolumeMounts: append([]corev1.VolumeMount{
							{
								Name:      ETCDDataVolumeName(role),
								MountPath: etcdDataVolumeMountPath,
							},
							{
								Name:      etcdCACertificateVolumeName,
								MountPath: etcdCACertificateVolumeMountPath,
							},
							{
								Name:      etcdClientCertificateVolumeName,
								MountPath: etcdClientCertificateVolumeMountPath,
							},
						}, backupVolumeMounts...),
					},
				},
				Volumes: append([]corev1.Volume{
					{
						Name: etcdConfigMapVolumeName,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: ETCDConfigMapName(role),
								},
								DefaultMode: pointer.Int32Ptr(356),
							},
						},
					},
					{
						Name: etcdCACertificateVolumeName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: ETCDSecretNameCACertificate,
							},
						},
					},
					{
						Name: etcdServerCertificateVolumeName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: ETCDSecretNameServerCertificate(role),
							},
						},
					},
					{
						Name: etcdClientCertificateVolumeName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: ETCDSecretNameClientCertificate,
							},
						},
					},
				}, backupVolumes...),
			},
		}
		sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: ETCDDataVolumeName(role),
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					StorageClassName: storageClassName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(storageCapacity),
						},
					},
				},
			},
		}
		return nil
	})
	return err
}

func (o *operation) deleteETCDStatefulSet(ctx context.Context, role string) error {
	if err := o.client.Delete(ctx, emptyETCDStatefulSet(o.namespace, role)); client.IgnoreNotFound(err) != nil {
		return err
	}

	if o.handleETCDPersistentVolumes {
		return client.IgnoreNotFound(o.client.Delete(ctx, emptyETCDPersistentVolumeClaim(o.namespace, role)))
	}

	return nil
}

func emptyETCDStatefulSet(namespace, role string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: ETCDStatefulSetName(role), Namespace: namespace}}
}

func emptyETCDPersistentVolumeClaim(namespace, role string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: ETCDPersistentVolumeClaimName(role), Namespace: namespace}}
}
