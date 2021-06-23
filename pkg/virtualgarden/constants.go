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
	ChecksumKeyKubeAPIServerBasicAuth          = "checksum/secret-kube-apiserver-basic-auth"
	ChecksumKeyKubeAPIServerAdmissionConfig    = "checksum/virtual-garden-kube-apiserver-admission-config"
	ChecksumKeyKubeControllerManagerClient     = "checksum/secret-kube-controller-manager-client"
	ChecksumKeyServiceAccountKey               = "checksum/secret-service-account-key"
)

// Names of volumes and corresponding volume mounts
const (
	volumeNameKubeAggregator                   = "kube-aggregator"
	volumeNameKubeAPIServer                    = "kube-apiserver"
	volumeNameKubeAPIServerCA                  = "ca-kube-apiserver"
	volumeNameKubeAPIServerBasicAuth           = "kube-apiserver-basic-auth"
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

const SecretKeyKubeconfig = "kubeconfig"

const kubeAPIServerContainerName = "kube-apiserver"

const (
	kubeControllerManager = "kube-controller-manager"
)
