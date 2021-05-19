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

import corev1 "k8s.io/api/core/v1"

// Imports defines the structure for the required configuration values from other components.
type Imports struct {
	// Cluster is the kubeconfig of the hosting cluster into which the virtual garden shall be installed.
	Cluster string `json:"cluster" yaml:"cluster"`
	// HostingCluster contains settings for the hosting cluster that runs the virtual garden.
	HostingCluster HostingCluster `json:"hostingCluster" yaml:"hostingCluster"`
	// VirtualGarden contains configuration for the virtual garden cluster.
	VirtualGarden VirtualGarden `json:"virtualGarden" yaml:"virtualGarden"`
	// Credentials maps names to credential pairs. Other structures shall reference those credentials using the names.
	// +optional
	Credentials map[string]Credentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// HostingCluster contains settings for the hosting cluster that runs the virtual garden.
type HostingCluster struct {
	// Kubeconfig is the kubeconfig of the hosting cluster into which the virtual garden shall be installed.
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`
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
}

// ETCDBackup contains configuration for the backup of the main etcd for the virtual garden.
type ETCDBackup struct {
	// InfrastructureProvider is the provider type of the underlying infrastructure for the blob storage bucket.
	InfrastructureProvider InfrastructureProviderType `json:"infrastructureProvider" yaml:"infrastructureProvider"`
	// Region is the name of the region in which the blob storage bucket shall be created.
	Region string `json:"region" yaml:"region"`
	// BucketName is the name of the blob storage bucket.
	BucketName string `json:"bucketName" yaml:"bucketName"`
	// CredentialsRef is the name of a key in the credentials that shall be used for the creation of the blob storage
	// bucket.
	CredentialsRef string `json:"credentialsRef" yaml:"credentialsRef"`
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

	HVPAEnabled bool        `json:"hvpaEnabled,omitempty" yaml:"hvpaEnabled,omitempty"`
	HVPA        *HvpaConfig `json:"hvpa,omitempty" yaml:"hvpa,omitempty"`

	EventTTL      *string `json:"eventTTL,omitempty" yaml:"eventTTL,omitempty"`
	OidcIssuerURL *string `json:"oidcIssuerURL,omitempty" yaml:"oidcIssuerURL,omitempty"`

	AdditionalVolumeMounts []corev1.VolumeMount `json:"additionalVolumeMounts,omitempty" yaml:"additionalVolumeMounts,omitempty"`
	AdditionalVolumes      []corev1.Volume      `json:"additionalVolumes,omitempty" yaml:"additionalVolumes,omitempty"`

	HorizontalPodAutoscaler *HorizontalPodAutoscaler `json:"horizontalPodAutoscaler,omitempty" yaml:"horizontalPodAutoscaler,omitempty"`
}

type HorizontalPodAutoscaler struct {
	DownscaleStabilization  string
	ReadinessDelay          string
	CpuInitializationPeriod string
	SyncPeriod              string
	Tolerance               string
}

type GardenerControlplane struct {
	ValidatingWebhookEnabled bool `json:"validatingWebhookEnabled,omitempty" yaml:"validatingWebhookEnabled,omitempty"`
	MutatingWebhookEnabled   bool `json:"mutatingWebhookEnabled,omitempty" yaml:"mutatingWebhookEnabled,omitempty"`
}

type AuditWebhookConfig struct {
	Config string `json:"config,omitempty" yaml:"config,omitempty"`
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
	// Type is the credentials type.
	Type InfrastructureProviderType `json:"type" yaml:"type"`
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
)
