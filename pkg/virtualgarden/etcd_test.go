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
	"fmt"
	"io/ioutil"

	"github.com/gardener/virtual-garden/pkg/api"
	mockclient "github.com/gardener/virtual-garden/pkg/mock/controller-runtime/client"
	"github.com/gardener/virtual-garden/pkg/provider/fake"

	"github.com/gardener/gardener/pkg/utils"
	secretutils "github.com/gardener/gardener/pkg/utils/secrets"
	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed templates/testdata/etcd-bootstrap.sh
var bootstrapScript string

//go:embed templates/testdata/etcd-config-main.yaml
var configMain string

//go:embed templates/testdata/etcd-config-events.yaml
var configEvents string

var _ = Describe("Etcd", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient

		ctx              = context.TODO()
		namespace        = "foo"
		storageClassName = "fast"
		fakeErr          = fmt.Errorf("fail")

		infrastructureStorageClassProvisioner = "FaKe"
		infrastructureStorageClassParameters  = map[string]string{"foo": "bar"}

		bucketName       = "main-backup"
		backupSecretData = map[string][]byte{"foo": []byte("bar")}
		op               *operation
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)

		op = &operation{
			client:                 c,
			log:                    &logrus.Logger{Out: ioutil.Discard},
			namespace:              namespace,
			infrastructureProvider: fake.NewInfrastructureProvider(infrastructureStorageClassProvisioner, infrastructureStorageClassParameters),
			backupProvider:         fake.NewBackupProvider(backupSecretData),
			imports: &api.Imports{
				VirtualGarden: api.VirtualGarden{
					ETCD: &api.ETCD{
						Backup: &api.ETCDBackup{
							BucketName: bucketName,
						},
						StorageClassName:            &storageClassName,
						HVPAEnabled:                 true,
						HandleETCDPersistentVolumes: true,
					},
					PriorityClassName: "garden-controlplane",
				},
			},
			imageRefs: api.ImageRefs{
				ETCDImage:              "eu.gcr.io/sap-se-gcr-k8s-public/quay_io/coreos/etcd:v3.3.17",
				ETCDBackupRestoreImage: "eu.gcr.io/sap-se-gcr-k8s-public/eu_gcr_io/gardener-project/gardener/etcdbrctl:v0.9.1",
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DeployETCD", func() {
		var (
			storageClassObjectMeta metav1.ObjectMeta
			secretCAObjectMeta     metav1.ObjectMeta
			secretClientObjectMeta metav1.ObjectMeta
			secretBackupObjectMeta metav1.ObjectMeta

			serviceMainObjectMeta      metav1.ObjectMeta
			configMapMainObjectMeta    metav1.ObjectMeta
			secretServerMainObjectMeta metav1.ObjectMeta
			statefulSetMainObjectMeta  metav1.ObjectMeta
			hvpaMainObjectMeta         metav1.ObjectMeta

			serviceEventsObjectMeta      metav1.ObjectMeta
			configMapEventsObjectMeta    metav1.ObjectMeta
			secretServerEventsObjectMeta metav1.ObjectMeta
			statefulSetEventsObjectMeta  metav1.ObjectMeta
			hvpaEventsObjectMeta         metav1.ObjectMeta

			secretChecksumCA           string
			secretChecksumClient       string
			secretChecksumBackup       string
			secretChecksumServerMain   string
			secretChecksumServerEvents string
			configMapChecksumMain      string
			configMapChecksumEvents    string

			labelsMain = map[string]string{
				"app":       "virtual-garden",
				"component": "etcd",
				"role":      "main",
			}
			labelsEvents = map[string]string{
				"app":       "virtual-garden",
				"component": "etcd",
				"role":      "events",
			}

			serviceFor = func(role string) *corev1.Service {
				var (
					objectMeta metav1.ObjectMeta
					labels     map[string]string
				)

				if role == ETCDRoleMain {
					objectMeta = serviceMainObjectMeta
					labels = labelsMain
				} else if role == ETCDRoleEvents {
					objectMeta = serviceEventsObjectMeta
					labels = labelsEvents
				}

				return &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      objectMeta.Name,
						Namespace: objectMeta.Namespace,
						Labels:    labels,
					},
					Spec: corev1.ServiceSpec{
						Type:            corev1.ServiceTypeClusterIP,
						SessionAffinity: corev1.ServiceAffinityNone,
						Selector:        labels,
						Ports: []corev1.ServicePort{
							{
								Name:       "client",
								Protocol:   corev1.ProtocolTCP,
								Port:       2379,
								TargetPort: intstr.FromInt(2379),
							},
							{
								Name:       "backup-client",
								Protocol:   corev1.ProtocolTCP,
								Port:       8080,
								TargetPort: intstr.FromInt(8080),
							},
						},
					},
				}
			}
			configMapFor = func(role string) (*corev1.ConfigMap, string) {
				var (
					objectMeta metav1.ObjectMeta
					labels     map[string]string
					config     string
				)

				if role == ETCDRoleMain {
					objectMeta = configMapMainObjectMeta
					labels = labelsMain
					config = configMain
				} else if role == ETCDRoleEvents {
					objectMeta = configMapEventsObjectMeta
					labels = labelsEvents
					config = configEvents
				}

				data := map[string]string{
					"bootstrap.sh":  bootstrapScript,
					"etcd.conf.yml": config,
				}

				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      objectMeta.Name,
						Namespace: objectMeta.Namespace,
						Labels:    labels,
					},
					Data: data,
				}, utils.ComputeChecksum(data)
			}
			statefulSetFor = func(role string, backup *api.ETCDBackup) *appsv1.StatefulSet {
				var (
					objectMeta               metav1.ObjectMeta
					labels                   map[string]string
					serviceName              string
					configMapName            string
					secretServerCertName     string
					checksumConfigMap        string
					checksumServerCert       string
					pvcName                  string
					storageSize              string
					storageClassName         *string
					env                      []corev1.EnvVar
					additionalPodAnnotations map[string]string
					additionalVolumes        []corev1.Volume
					additionalVolumeMounts   []corev1.VolumeMount
					additionalCommand        []string
				)

				if role == ETCDRoleMain {
					objectMeta = statefulSetMainObjectMeta
					labels = labelsMain
					serviceName = serviceMainObjectMeta.Name
					configMapName = configMapMainObjectMeta.Name
					secretServerCertName = secretServerMainObjectMeta.Name
					checksumConfigMap = configMapChecksumMain
					checksumServerCert = secretChecksumServerMain
					pvcName = "main-virtual-garden-etcd"
					storageSize = "25Gi"
					storageClassName = &storageClassObjectMeta.Name
					if backup != nil {
						env = append([]corev1.EnvVar{
							{
								Name:  "STORAGE_CONTAINER",
								Value: bucketName,
							},
						}, fake.FakeEnv...)
						additionalPodAnnotations = map[string]string{"checksum/secret-etcd-backup": secretChecksumBackup}
						additionalVolumes = []corev1.Volume{
							{
								Name: "backup-credentials",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: secretBackupObjectMeta.Name,
									},
								},
							},
						}
						additionalVolumeMounts = []corev1.VolumeMount{
							{
								Name:      "backup-credentials",
								MountPath: "/var/etcd/backup",
							},
						}
						additionalCommand = []string{
							"--schedule=0 */24 * * *",
							"--defragmentation-schedule=0 1 * * *",
							"--storage-provider=" + fake.FakeProviderName,
							"--store-prefix=virtual-garden-etcd-main",
							"--delta-snapshot-period=5m",
							"--delta-snapshot-memory-limit=104857600",
							"--embedded-etcd-quota-bytes=8589934592",
						}
					}
				} else if role == ETCDRoleEvents {
					objectMeta = statefulSetEventsObjectMeta
					labels = labelsEvents
					serviceName = serviceEventsObjectMeta.Name
					configMapName = configMapEventsObjectMeta.Name
					secretServerCertName = secretServerEventsObjectMeta.Name
					checksumConfigMap = configMapChecksumEvents
					checksumServerCert = secretChecksumServerEvents
					pvcName = "virtual-garden-etcd-events"
					storageSize = "10Gi"
				}

				return &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      objectMeta.Name,
						Namespace: objectMeta.Namespace,
						Labels:    labels,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: pointer.Int32Ptr(1),
						Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
						ServiceName:    serviceName,
						UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: utils.MergeStringMaps(map[string]string{
									"checksum/configmap-etcd-bootstrap-config": checksumConfigMap,
									"checksum/secret-etcd-ca":                  secretChecksumCA,
									"checksum/secret-etcd-client":              secretChecksumClient,
									"checksum/secret-etcd-server":              checksumServerCert,
								}, additionalPodAnnotations),
								Labels: labels,
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
															Values: []string{
																Prefix,
															},
														},
														{
															Key:      "component",
															Operator: metav1.LabelSelectorOpIn,
															Values: []string{
																"etcd",
															},
														},
													},
												},
												TopologyKey: corev1.LabelHostname,
											},
										},
									},
								},
								PriorityClassName: "garden-controlplane",
								Containers: []corev1.Container{
									{
										Name:            "etcd",
										Image:           "eu.gcr.io/sap-se-gcr-k8s-public/quay_io/coreos/etcd:v3.3.17",
										ImagePullPolicy: corev1.PullIfNotPresent,
										Command:         []string{"/bootstrap/bootstrap.sh"},
										ReadinessProbe: &corev1.Probe{
											Handler: corev1.Handler{
												HTTPGet: &corev1.HTTPGetAction{
													Path:   "/healthz",
													Port:   intstr.FromInt(8080),
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
														"--cacert=/var/etcd/ssl/ca/ca.crt",
														"--cert=/var/etcd/ssl/client/tls.crt",
														"--key=/var/etcd/ssl/client/tls.key",
														fmt.Sprintf("--endpoints=https://%s-0:%d", objectMeta.Name, 2379),
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
												Name:          "server",
												ContainerPort: 2380,
												Protocol:      corev1.ProtocolTCP,
											},
											{
												Name:          "client",
												ContainerPort: 2379,
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
												Name:      pvcName,
												MountPath: "/var/etcd/data",
											},
											{
												Name:      "bootstrap-config",
												MountPath: "/bootstrap",
											},
											{
												Name:      "ca-cert",
												MountPath: "/var/etcd/ssl/ca",
											},
											{
												Name:      "server-cert",
												MountPath: "/var/etcd/ssl/server",
											},
											{
												Name:      "client-cert",
												MountPath: "/var/etcd/ssl/client",
											},
										},
									},
									{
										Name:            "backup-restore",
										Image:           "eu.gcr.io/sap-se-gcr-k8s-public/eu_gcr_io/gardener-project/gardener/etcdbrctl:v0.9.1",
										ImagePullPolicy: corev1.PullIfNotPresent,
										Command: append([]string{
											"etcdbrctl",
											"server",
											"--data-dir=/var/etcd/data/new.etcd",
											"--cacert=/var/etcd/ssl/ca/ca.crt",
											"--cert=/var/etcd/ssl/client/tls.crt",
											"--key=/var/etcd/ssl/client/tls.key",
											"--insecure-transport=false",
											"--insecure-skip-tls-verify=false",
											fmt.Sprintf("--endpoints=https://%s-0:2379", objectMeta.Name),
											"--etcd-connection-timeout=5m",
											"--garbage-collection-period=12h",
											"--snapstore-temp-directory=/var/etcd/data/temp",
										}, additionalCommand...),
										Ports: []corev1.ContainerPort{
											{
												Name:          "server",
												ContainerPort: 8080,
												Protocol:      corev1.ProtocolTCP,
											},
										},
										Env: env,
										VolumeMounts: append([]corev1.VolumeMount{
											{
												Name:      pvcName,
												MountPath: "/var/etcd/data",
											},
											{
												Name:      "ca-cert",
												MountPath: "/var/etcd/ssl/ca",
											},
											{
												Name:      "client-cert",
												MountPath: "/var/etcd/ssl/client",
											},
										}, additionalVolumeMounts...),
									},
								},
								Volumes: append([]corev1.Volume{
									{
										Name: "bootstrap-config",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: configMapName,
												},
												DefaultMode: pointer.Int32Ptr(356),
											},
										},
									},
									{
										Name: "ca-cert",
										VolumeSource: corev1.VolumeSource{
											Secret: &corev1.SecretVolumeSource{
												SecretName: secretCAObjectMeta.Name,
											},
										},
									},
									{
										Name: "server-cert",
										VolumeSource: corev1.VolumeSource{
											Secret: &corev1.SecretVolumeSource{
												SecretName: secretServerCertName,
											},
										},
									},
									{
										Name: "client-cert",
										VolumeSource: corev1.VolumeSource{
											Secret: &corev1.SecretVolumeSource{
												SecretName: secretClientObjectMeta.Name,
											},
										},
									},
								}, additionalVolumes...),
							},
						},
						VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: pvcName,
								},
								Spec: corev1.PersistentVolumeClaimSpec{
									AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
									StorageClassName: storageClassName,
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceStorage: resource.MustParse(storageSize),
										},
									},
								},
							},
						},
					},
				}
			}
			hvpaFor = func(role string) *hvpav1alpha1.Hvpa {
				var (
					hvpaContainerScalingOff = autoscalingv1beta2.ContainerScalingModeOff
					objectMeta              metav1.ObjectMeta
					statefulSetName         string
				)

				if role == ETCDRoleMain {
					objectMeta = hvpaMainObjectMeta
					statefulSetName = statefulSetMainObjectMeta.Name
				} else if role == ETCDRoleEvents {
					objectMeta = hvpaEventsObjectMeta
					statefulSetName = statefulSetEventsObjectMeta.Name
				}

				return &hvpav1alpha1.Hvpa{
					ObjectMeta: metav1.ObjectMeta{
						Name:      objectMeta.Name,
						Namespace: objectMeta.Namespace,
					},
					Spec: hvpav1alpha1.HvpaSpec{
						Replicas: pointer.Int32Ptr(1),
						TargetRef: &autoscalingv2beta1.CrossVersionObjectReference{
							APIVersion: "apps/v1",
							Kind:       "StatefulSet",
							Name:       statefulSetName,
						},
						Hpa: hvpav1alpha1.HpaSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"role": "etcd-hpa-" + role},
							},
							Deploy: false,
							Template: hvpav1alpha1.HpaTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{"role": "etcd-hpa-" + role},
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
						},
						Vpa: hvpav1alpha1.VpaSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"role": "etcd-vpa-" + role},
							},
							Deploy: true,
							Template: hvpav1alpha1.VpaTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{"role": "etcd-vpa-" + role},
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
												Mode:          &hvpaContainerScalingOff,
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
						},
						WeightBasedScalingIntervals: []hvpav1alpha1.WeightBasedScalingInterval{
							{
								VpaWeight:         hvpav1alpha1.VpaOnly,
								StartReplicaCount: 1,
								LastReplicaCount:  1,
							},
						},
					},
				}
			}
		)

		BeforeEach(func() {
			storageClassObjectMeta = objectMeta(ETCDStorageClassName(op.imports.VirtualGarden.ETCD), "")
			secretCAObjectMeta = objectMeta(ETCDSecretNameCACertificate, namespace)
			secretClientObjectMeta = objectMeta(ETCDSecretNameClientCertificate, namespace)
			secretBackupObjectMeta = objectMeta(ETCDSecretNameBackup, namespace)

			serviceMainObjectMeta = objectMeta(ETCDServiceName(ETCDRoleMain), namespace)
			configMapMainObjectMeta = objectMeta(ETCDConfigMapName(ETCDRoleMain), namespace)
			secretServerMainObjectMeta = objectMeta(ETCDSecretNameServerCertificate(ETCDRoleMain), namespace)
			statefulSetMainObjectMeta = objectMeta(ETCDStatefulSetName(ETCDRoleMain), namespace)
			hvpaMainObjectMeta = objectMeta(ETCDHVPAName(ETCDRoleMain), namespace)

			serviceEventsObjectMeta = objectMeta(ETCDServiceName(ETCDRoleEvents), namespace)
			configMapEventsObjectMeta = objectMeta(ETCDConfigMapName(ETCDRoleEvents), namespace)
			secretServerEventsObjectMeta = objectMeta(ETCDSecretNameServerCertificate(ETCDRoleEvents), namespace)
			statefulSetEventsObjectMeta = objectMeta(ETCDStatefulSetName(ETCDRoleEvents), namespace)
			hvpaEventsObjectMeta = objectMeta(ETCDHVPAName(ETCDRoleEvents), namespace)
		})

		It("should correctly deploy all etcd resources (w/ backup, w/ hvpa)", func() {
			gomock.InOrder(
				c.EXPECT().Get(ctx, client.ObjectKey{Name: storageClassObjectMeta.Name}, gomock.AssignableToTypeOf(&storagev1.StorageClass{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&storagev1.StorageClass{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(&storagev1.StorageClass{
						ObjectMeta:           storageClassObjectMeta,
						AllowVolumeExpansion: pointer.BoolPtr(true),
						Provisioner:          infrastructureStorageClassProvisioner,
						Parameters:           infrastructureStorageClassParameters,
					}))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretCAObjectMeta.Name, Namespace: secretCAObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					Expect(o.ObjectMeta).To(Equal(secretCAObjectMeta))
					Expect(o.Type).To(Equal(corev1.SecretTypeOpaque))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificateCA))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyPrivateKeyCA))
					secretChecksumCA = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretClientObjectMeta.Name, Namespace: secretClientObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					Expect(o.ObjectMeta).To(Equal(secretClientObjectMeta))
					Expect(o.Type).To(Equal(corev1.SecretTypeTLS))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificateCA))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificate))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyPrivateKey))
					secretChecksumClient = utils.ComputeChecksum(o.Data)
				}),

				c.EXPECT().Get(ctx, client.ObjectKey{Name: serviceMainObjectMeta.Name, Namespace: serviceMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Service{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(serviceFor(ETCDRoleMain)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: configMapMainObjectMeta.Name, Namespace: configMapMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					configMap, checksum := configMapFor(ETCDRoleMain)
					Expect(obj).To(Equal(configMap))
					configMapChecksumMain = checksum
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretServerMainObjectMeta.Name, Namespace: secretServerMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					verifyServerCertSecret(o, secretServerMainObjectMeta)
					secretChecksumServerMain = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretBackupObjectMeta.Name, Namespace: secretBackupObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					secretChecksumBackup = utils.ComputeChecksum(backupSecretData)
					Expect(obj).To(Equal(&corev1.Secret{
						ObjectMeta: secretBackupObjectMeta,
						Type:       corev1.SecretTypeOpaque,
						Data:       backupSecretData,
					}))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: statefulSetMainObjectMeta.Name, Namespace: statefulSetMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(statefulSetFor(ETCDRoleMain, op.imports.VirtualGarden.ETCD.Backup)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: hvpaMainObjectMeta.Name, Namespace: hvpaMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&hvpav1alpha1.Hvpa{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&hvpav1alpha1.Hvpa{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(hvpaFor(ETCDRoleMain)))
				}),

				c.EXPECT().Get(ctx, client.ObjectKey{Name: serviceEventsObjectMeta.Name, Namespace: serviceEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Service{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(serviceFor(ETCDRoleEvents)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: configMapEventsObjectMeta.Name, Namespace: configMapEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					configMap, checksum := configMapFor(ETCDRoleEvents)
					Expect(obj).To(Equal(configMap))
					configMapChecksumEvents = checksum
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretServerEventsObjectMeta.Name, Namespace: secretServerEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					verifyServerCertSecret(o, secretServerEventsObjectMeta)
					secretChecksumServerEvents = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: statefulSetEventsObjectMeta.Name, Namespace: statefulSetEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(statefulSetFor(ETCDRoleEvents, nil)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: hvpaEventsObjectMeta.Name, Namespace: hvpaEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&hvpav1alpha1.Hvpa{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&hvpav1alpha1.Hvpa{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(hvpaFor(ETCDRoleEvents)))
				}),
			)

			Expect(op.DeployETCD(ctx)).To(Succeed())
		})

		It("should correctly deploy all etcd resources (w/o backup, w/o hvpa)", func() {
			op.imports.VirtualGarden.ETCD.Backup = nil
			op.imports.VirtualGarden.ETCD.HVPAEnabled = false

			gomock.InOrder(
				c.EXPECT().Get(ctx, client.ObjectKey{Name: storageClassObjectMeta.Name}, gomock.AssignableToTypeOf(&storagev1.StorageClass{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&storagev1.StorageClass{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(&storagev1.StorageClass{
						ObjectMeta:           storageClassObjectMeta,
						AllowVolumeExpansion: pointer.BoolPtr(true),
						Provisioner:          infrastructureStorageClassProvisioner,
						Parameters:           infrastructureStorageClassParameters,
					}))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretCAObjectMeta.Name, Namespace: secretCAObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					Expect(o.ObjectMeta).To(Equal(secretCAObjectMeta))
					Expect(o.Type).To(Equal(corev1.SecretTypeOpaque))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificateCA))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyPrivateKeyCA))
					secretChecksumCA = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretClientObjectMeta.Name, Namespace: secretClientObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					Expect(o.ObjectMeta).To(Equal(secretClientObjectMeta))
					Expect(o.Type).To(Equal(corev1.SecretTypeTLS))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificateCA))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyCertificate))
					Expect(o.Data).To(HaveKey(secretutils.DataKeyPrivateKey))
					secretChecksumClient = utils.ComputeChecksum(o.Data)
				}),

				c.EXPECT().Get(ctx, client.ObjectKey{Name: serviceMainObjectMeta.Name, Namespace: serviceMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Service{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(serviceFor(ETCDRoleMain)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: configMapMainObjectMeta.Name, Namespace: configMapMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					configMap, checksum := configMapFor(ETCDRoleMain)
					Expect(obj).To(Equal(configMap))
					configMapChecksumMain = checksum
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretServerMainObjectMeta.Name, Namespace: secretServerMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					verifyServerCertSecret(o, secretServerMainObjectMeta)
					secretChecksumServerMain = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: statefulSetMainObjectMeta.Name, Namespace: statefulSetMainObjectMeta.Namespace}, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(statefulSetFor(ETCDRoleMain, op.imports.VirtualGarden.ETCD.Backup)))
				}),

				c.EXPECT().Get(ctx, client.ObjectKey{Name: serviceEventsObjectMeta.Name, Namespace: serviceEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Service{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(serviceFor(ETCDRoleEvents)))
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: configMapEventsObjectMeta.Name, Namespace: configMapEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.ConfigMap{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					configMap, checksum := configMapFor(ETCDRoleEvents)
					Expect(obj).To(Equal(configMap))
					configMapChecksumEvents = checksum
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: secretServerEventsObjectMeta.Name, Namespace: secretServerEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&corev1.Secret{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")).Times(2),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&corev1.Secret{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					o, ok := obj.(*corev1.Secret)
					Expect(ok).To(BeTrue())
					verifyServerCertSecret(o, secretServerEventsObjectMeta)
					secretChecksumServerEvents = utils.ComputeChecksum(o.Data)
				}),
				c.EXPECT().Get(ctx, client.ObjectKey{Name: statefulSetEventsObjectMeta.Name, Namespace: statefulSetEventsObjectMeta.Namespace}, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Return(apierrors.NewNotFound(schema.GroupResource{}, "")),
				c.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&appsv1.StatefulSet{})).Do(func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) {
					Expect(obj).To(Equal(statefulSetFor(ETCDRoleEvents, nil)))
				}),
			)

			Expect(op.DeployETCD(ctx)).To(Succeed())
		})
	})

	Describe("#DeleteETCD", func() {
		It("should correctly delete all etcd resources (w/ backup, w/ pvc handling, w/ hvpa)", func() {
			gomock.InOrder(
				c.EXPECT().Delete(ctx, &hvpav1alpha1.Hvpa{ObjectMeta: objectMeta(ETCDHVPAName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &appsv1.StatefulSet{ObjectMeta: objectMeta(ETCDStatefulSetName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: objectMeta(ETCDPersistentVolumeClaimName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameBackup, namespace)}),
				c.EXPECT().Delete(ctx, &corev1.ConfigMap{ObjectMeta: objectMeta(ETCDConfigMapName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Service{ObjectMeta: objectMeta(ETCDServiceName(ETCDRoleMain), namespace)}),

				c.EXPECT().Delete(ctx, &hvpav1alpha1.Hvpa{ObjectMeta: objectMeta(ETCDHVPAName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &appsv1.StatefulSet{ObjectMeta: objectMeta(ETCDStatefulSetName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: objectMeta(ETCDPersistentVolumeClaimName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.ConfigMap{ObjectMeta: objectMeta(ETCDConfigMapName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Service{ObjectMeta: objectMeta(ETCDServiceName(ETCDRoleEvents), namespace)}),

				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameCACertificate, namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameServerCertificate(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameServerCertificate(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameClientCertificate, namespace)}),
				c.EXPECT().List(ctx, &appsv1.StatefulSetList{}),
				c.EXPECT().Delete(ctx, &storagev1.StorageClass{ObjectMeta: objectMeta(ETCDStorageClassName(op.imports.VirtualGarden.ETCD), "")}),
			)

			Expect(op.DeleteETCD(ctx)).To(Succeed())
		})

		It("should correctly delete all etcd resources (w/o backup, w/o pvc handling, w/o hvpa)", func() {
			op.imports.VirtualGarden.ETCD.Backup = nil
			op.imports.VirtualGarden.ETCD.HVPAEnabled = false
			op.imports.VirtualGarden.ETCD.HandleETCDPersistentVolumes = false

			gomock.InOrder(
				c.EXPECT().Delete(ctx, &appsv1.StatefulSet{ObjectMeta: objectMeta(ETCDStatefulSetName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.ConfigMap{ObjectMeta: objectMeta(ETCDConfigMapName(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Service{ObjectMeta: objectMeta(ETCDServiceName(ETCDRoleMain), namespace)}),

				c.EXPECT().Delete(ctx, &appsv1.StatefulSet{ObjectMeta: objectMeta(ETCDStatefulSetName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.ConfigMap{ObjectMeta: objectMeta(ETCDConfigMapName(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Service{ObjectMeta: objectMeta(ETCDServiceName(ETCDRoleEvents), namespace)}),

				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameCACertificate, namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameServerCertificate(ETCDRoleMain), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameServerCertificate(ETCDRoleEvents), namespace)}),
				c.EXPECT().Delete(ctx, &corev1.Secret{ObjectMeta: objectMeta(ETCDSecretNameClientCertificate, namespace)}),
				c.EXPECT().List(ctx, &appsv1.StatefulSetList{}).DoAndReturn(func(_ context.Context, actual *appsv1.StatefulSetList, _ ...client.ListOption) error {
					*actual = appsv1.StatefulSetList{Items: []appsv1.StatefulSet{{}}}
					return nil
				}),
			)

			Expect(op.DeleteETCD(ctx)).To(Succeed())
		})
	})

	Describe("#OtherVirtualGardensInHostingCluster", func() {
		It("should return the list error", func() {
			c.EXPECT().List(ctx, &appsv1.StatefulSetList{}).Return(fakeErr)

			otherVirtualGardens, err := OtherVirtualGardensInHostingCluster(ctx, c, namespace)
			Expect(otherVirtualGardens).To(BeFalse())
			Expect(err).To(MatchError(fakeErr))
		})

		It("should return false (no other statefulsets)", func() {
			c.EXPECT().List(ctx, &appsv1.StatefulSetList{})

			otherVirtualGardens, err := OtherVirtualGardensInHostingCluster(ctx, c, namespace)
			Expect(otherVirtualGardens).To(BeFalse())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return false (only statefulsets in the same namespace)", func() {
			c.EXPECT().List(ctx, &appsv1.StatefulSetList{}).DoAndReturn(func(_ context.Context, actual *appsv1.StatefulSetList, _ ...client.ListOption) error {
				*actual = appsv1.StatefulSetList{
					Items: []appsv1.StatefulSet{{ObjectMeta: metav1.ObjectMeta{Namespace: namespace}}},
				}
				return nil
			})

			otherVirtualGardens, err := OtherVirtualGardensInHostingCluster(ctx, c, namespace)
			Expect(otherVirtualGardens).To(BeFalse())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return true", func() {
			c.EXPECT().List(ctx, &appsv1.StatefulSetList{}).DoAndReturn(func(_ context.Context, actual *appsv1.StatefulSetList, _ ...client.ListOption) error {
				*actual = appsv1.StatefulSetList{
					Items: []appsv1.StatefulSet{{}},
				}
				return nil
			})

			otherVirtualGardens, err := OtherVirtualGardensInHostingCluster(ctx, c, namespace)
			Expect(otherVirtualGardens).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func objectMeta(name, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: name, Namespace: namespace}
}

func verifyServerCertSecret(obj *corev1.Secret, objectMeta metav1.ObjectMeta) {
	Expect(obj.ObjectMeta).To(Equal(objectMeta))
	Expect(obj.Type).To(Equal(corev1.SecretTypeTLS))
	Expect(obj.Data).To(HaveKey(secretutils.DataKeyCertificateCA))
	Expect(obj.Data).To(HaveKey(secretutils.DataKeyCertificate))
	Expect(obj.Data).To(HaveKey(secretutils.DataKeyPrivateKey))
}
