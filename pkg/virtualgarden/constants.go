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

const (
	LabelKeyApp       = "app"
	LabelKeyComponent = "component"
	LabelKeyRole      = "role"

	LabelValueAllowed = "allowed"
)

// Keys of annotations for checksums
const (
	ChecksumKeyKubeAPIServerAuditPolicyConfig  = "checksum/configmap-kube-apiserver-audit-policy-config"
	ChecksumKeyKubeAPIServerEncryptionConfig   = "checksum/secret-kube-apiserver-encryption-config"
	ChecksumKeyKubeAggregatorCA                = "checksum/secret-kube-aggregator-ca"
	ChecksumKeyKubeAggregatorClient            = "checksum/secret-kube-aggregator-client"
	ChecksumKeyKubeAPIServerCA                 = "checksum/secret-kube-apiserver-ca"
	ChecksumKeyKubeAPIServerServer             = "checksum/secret-kube-apiserver-server"
	ChecksumKeyKubeAPIServerAuditWebhookConfig = "checksum/secret-kube-apiserver-audit-webhook-config"
	ChecksumKeyKubeAPIServerAuthWebhookConfig  = "checksum/secret-kube-apiserver-auth-webhook-config"
	ChecksumKeyKubeAPIServerStaticToken        = "checksum/secret-kube-apiserver-static-token"
	ChecksumKeyKubeAPIServerAdmissionConfig    = "checksum/virtual-garden-kube-apiserver-admission-config"
	ChecksumKeyKubeControllerManagerClient     = "checksum/secret-kube-controller-manager-client"
	ChecksumKeyServiceAccountKey               = "checksum/secret-service-account-key"
)

// Names of volumes and corresponding volume mounts
const (
	volumeNameKubeAggregator                   = "kube-aggregator"
	volumeNameKubeAPIServer                    = "kube-apiserver"
	volumeNameKubeAPIServerCA                  = "ca-kube-apiserver"
	volumeNameKubeAPIServerStaticToken         = "kube-apiserver-static-token"
	volumeNameKubeAPIServerAdmissionConfig     = "kube-apiserver-admission-config"
	volumeNameKubeAPIServerAdmissionKubeconfig = "kube-apiserver-admission-kubeconfig"
	volumeNameKubeAPIServerAdmissionTokens     = "kube-apiserver-admission-tokens"
	volumeNameKubeAPIServerEncryptionConfig    = "kube-apiserver-encryption-config"
	volumeNameKubeAPIServerAuthWebhookConfig   = "kube-apiserver-auth-webhook-config"
	volumeNameKubeAPIServerAuditPolicyConfig   = "kube-apiserver-audit-policy-config"
	volumeNameKubeAPIServerAuditWebhookConfig  = "kube-apiserver-audit-webhook-config"
	volumeNameKubeControllerManager            = "kube-controller-manager"
	volumeNameServiceAccountKey                = "service-account-key"
	volumeNameCAETCD                           = "ca-etcd"
	volumeNameCAFrontProxy                     = "ca-front-proxy"
	volumeNameETCDClientTLS                    = "etcd-client-tls"
	volumeNameSNITLS                           = "sni-tls"
	volumeNameFedora                           = "fedora-rhel6-openelec-cabundle"
	volumeNameCentos                           = "centos-rhel7-cabundle"
	volumeNameETCSSL                           = "etc-ssl"
)

// Keys of secrets and configmaps
const (
	ValidatingWebhookKey  = "validating-webhook"
	MutatingWebhookKey    = "mutating-webhook"
	AuditWebhookConfigKey = "audit-webhook-config.yaml"
	ConfigYamlKey         = "config.yaml"
	StaticTokenKey        = "static_tokens.csv"
	EncryptionConfigKey   = "encryption-config.yaml"
	ServiceAccountKey     = "service_account.key"
	ConfigurationYamlKey  = "configuration.yaml"
	AuditPolicyYamlKey    = "audit-policy.yaml"
)

const SecretKeyKubeconfig = "kubeconfig"

const kubeAPIServerContainerName = "kube-apiserver"

const kubeControllerManager = "kube-controller-manager"
