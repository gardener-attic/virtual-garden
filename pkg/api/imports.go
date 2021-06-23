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

package api

import (
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Imports defines the structure for the required configuration values from other components.
type Imports struct {
	// Cluster is the kubeconfig of the hosting cluster into which the virtual garden shall be installed.
	Cluster lsv1alpha1.Target `json:"cluster" yaml:"cluster"`
	// HostingCluster contains settings for the hosting cluster that runs the virtual garden.
	HostingCluster HostingCluster `json:"hostingCluster" yaml:"hostingCluster"`
	// VirtualGarden contains configuration for the virtual garden cluster.
	VirtualGarden VirtualGarden `json:"virtualGarden" yaml:"virtualGarden"`
}

// HostingCluster contains settings for the hosting cluster that runs the virtual garden.
type HostingCluster struct {
	// Namespace is a namespace in the hosting cluster into which the virtual garden shall be installed.
	Namespace string `json:"namespace" yaml:"namespace"`
	// InfrastructureProvider is the provider type of the underlying infrastructure of the hosting cluster.
	InfrastructureProvider InfrastructureProviderType `json:"infrastructureProvider" yaml:"infrastructureProvider"`
}

// VirtualGarden contains configuration for the virtual garden cluster.
type VirtualGarden struct {
	// ETCD contains configuration for the etcd that is used by the virtual garden kube-apiserver.
	// +optional
	ETCD *ETCD `json:"etcd,omitempty" yaml:"etcd,omitempty"`
	// KubeAPIServer contains configuration for the virtual garden kube-apiserver.
	// +optional
	KubeAPIServer *KubeAPIServer `json:"kubeAPIServer,omitempty" yaml:"kubeAPIServer,omitempty"`

	// DeleteNamespace controls if the namespace should be deleted
	DeleteNamespace bool `json:"deleteNamespace,omitempty" yaml:"deleteNamespace,omitempty"`

	PriorityClassName string `json:"priorityClassName,omitempty" yaml:"priorityClassName,omitempty"`
}

// ETCD contains configuration for the etcd that is used by the virtual garden kube-apiserver.
type ETCD struct {
	// StorageClassName allows to overwrite the default storage class name for etcd.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	// Backup contains configuration for the backup of the main etcd for the virtual garden.
	// +optional
	Backup *ETCDBackup `json:"backup,omitempty" yaml:"backup,omitempty"`

	HVPAEnabled bool `json:"hvpaEnabled,omitempty" yaml:"hvpaEnabled,omitempty"`

	// HandleETCDPersistentVolumes defines whether the PV(C)s that are getting automatically created by the etcd
	// statefulset shall be handled or not (false by default). If true then they will be deleted when the virtual
	// garden is deleted. Otherwise, they will remain in the system for manual cleanup (to prevent data loss).
	HandleETCDPersistentVolumes bool `json:"handleETCDPersistentVolumes,omitempty" yaml:"handleETCDPersistentVolumes,omitempty"`
}

// ETCDBackup contains configuration for the backup of the main etcd for the virtual garden.
type ETCDBackup struct {
	// InfrastructureProvider is the provider type of the underlying infrastructure for the blob storage bucket.
	InfrastructureProvider InfrastructureProviderType `json:"infrastructureProvider" yaml:"infrastructureProvider"`
	// Region is the name of the region in which the blob storage bucket shall be created.
	Region string `json:"region" yaml:"region"`
	// BucketName is the name of the blob storage bucket.
	BucketName string `json:"bucketName" yaml:"bucketName"`
	// Credentials contain the credentials that shall be used for the creation of the blob storage
	// bucket.
	Credentials *Credentials `json:"credentials" yaml:"credentials"`
}

// KubeAPIServer contains configuration for the virtual garden kube-apiserver.
type KubeAPIServer struct {
	Replicas int `json:"replicas,omitempty" yaml:"replicas,omitempty"`

	// SNI contains configuration for SNI settings for the virtual garden.
	// +optional
	SNI *SNI `json:"sni,omitempty" yaml:"sni,omitempty"`

	DnsAccessDomain      string               `json:"dnsAccessDomain,omitempty" yaml:"dnsAccessDomain,omitempty"`
	GardenerControlplane GardenerControlplane `json:"gardenerControlplane,omitempty" yaml:"gardenerControlplane,omitempty"`

	AuditWebhookConfig       AuditWebhookConfig `json:"auditWebhookConfig,omitempty" yaml:"auditWebhookConfig,omitempty"`
	AuditWebhookBatchMaxSize string             `json:"auditWebhookBatchMaxSize,omitempty" yaml:"auditWebhookBatchMaxSize,omitempty"`

	SeedAuthorizer SeedAuthorizer `json:"seedAuthorizer,omitempty" yaml:"seedAuthorizer,omitempty"`

	HVPAEnabled bool        `json:"hvpaEnabled,omitempty" yaml:"hvpaEnabled,omitempty"`
	HVPA        *HvpaConfig `json:"hvpa,omitempty" yaml:"hvpa,omitempty"`

	EventTTL      *string `json:"eventTTL,omitempty" yaml:"eventTTL,omitempty"`
	OidcIssuerURL *string `json:"oidcIssuerURL,omitempty" yaml:"oidcIssuerURL,omitempty"`

	AdditionalVolumeMounts []corev1.VolumeMount `json:"additionalVolumeMounts,omitempty" yaml:"additionalVolumeMounts,omitempty"`
	AdditionalVolumes      []corev1.Volume      `json:"additionalVolumes,omitempty" yaml:"additionalVolumes,omitempty"`

	HorizontalPodAutoscaler *HorizontalPodAutoscaler `json:"horizontalPodAutoscaler,omitempty" yaml:"horizontalPodAutoscaler,omitempty"`
}

type HorizontalPodAutoscaler struct {
	DownscaleStabilization  string `json:"downscaleStabilization,omitempty" yaml:"downscaleStabilization,omitempty"`
	ReadinessDelay          string `json:"readinessDelay,omitempty" yaml:"readinessDelay,omitempty"`
	CpuInitializationPeriod string `json:"cpuInitializationPeriod,omitempty" yaml:"cpuInitializationPeriod,omitempty"`
	SyncPeriod              string `json:"syncPeriod,omitempty" yaml:"syncPeriod,omitempty"`
	Tolerance               string `json:"tolerance,omitempty" yaml:"tolerance,omitempty"`
}

type GardenerControlplane struct {
	ValidatingWebhookEnabled bool `json:"validatingWebhookEnabled,omitempty" yaml:"validatingWebhookEnabled,omitempty"`
	MutatingWebhookEnabled   bool `json:"mutatingWebhookEnabled,omitempty" yaml:"mutatingWebhookEnabled,omitempty"`
}

type AuditWebhookConfig struct {
	Config string `json:"config,omitempty" yaml:"config,omitempty"`
}

type SeedAuthorizer struct {
	Enabled                  bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty" yaml:"certificateAuthorityData,omitempty"`
}

// SNI contains configuration for SNI settings for the virtual garden.
type SNI struct {
	// Hostname is the hostname for the virtual garden kube-apiserver. It is used to create DNS entries
	// pointing to it.
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// DNSClass is the DNS class that shall be used to create the DNS entries for the given hostnames.
	// +optional
	DNSClass *string `json:"dnsClass,omitempty" yaml:"dnsClass,omitempty"`
	// TTL is the time-to-live for the DNS entries created for the given hostnames.
	// +optional
	TTL *int32 `json:"ttl,omitempty" yaml:"ttl,omitempty"`
	// SecretName
	// +optional
	SecretName string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
}

// Credentials contains key-value pairs for credentials for a certain endpoint type.
type Credentials struct {
	// Data contains key-value pairs with the credentials information. The keys are specific for the credentials type.
	Data map[string]string `json:"data" yaml:"data"`
}

// InfrastructureProviderType is a string alias.
type InfrastructureProviderType string

const (
	// InfrastructureProviderAlicloud is a constant for the Alicloud infrastructure provider.
	InfrastructureProviderAlicloud InfrastructureProviderType = "alicloud"
	// InfrastructureProviderAWS is a constant for the AWS infrastructure provider.
	InfrastructureProviderAWS InfrastructureProviderType = "aws"
	// InfrastructureProviderGCP is a constant for the GCP infrastructure provider.
	InfrastructureProviderGCP InfrastructureProviderType = "gcp"
	// InfrastructureProviderFake is a constant for fake infrastructure provider.
	InfrastructureProviderFake InfrastructureProviderType = "fake"
)
