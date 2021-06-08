package virtualgarden

const (
	LabelKeyApp       = "app"
	LabelKeyComponent = "component"
	LabelKeyRole      = "role"

	LabelValueAllowed = "allowed"
)

const (
	ChecksumKeyKubeAPIServerAuditPolicyConfig  = "checksum/configmap-kube-apiserver-audit-policy-config"   // configmap-kube-apiserver-audit-policy-config.yaml
	ChecksumKeyKubeAPIServerEncryptionConfig   = "checksum/secret-kube-apiserver-encryption-config"        // secret-kube-apiserver-encryption-config.yaml
	ChecksumKeyKubeAggregatorCA                = "checksum/secret-kube-aggregator-ca"                      // secret-kube-aggregator-ca.yaml
	ChecksumKeyKubeAggregatorClient            = "checksum/secret-kube-aggregator-client"                  // secret-kube-aggregator-client-tls.yaml
	ChecksumKeyKubeAPIServerCA                 = "checksum/secret-kube-apiserver-ca"                       // secret-kube-apiserver-ca.yaml
	ChecksumKeyKubeAPIServerServer             = "checksum/secret-kube-apiserver-server"                   // secret-kube-apiserver-server-tls.yaml
	ChecksumKeyKubeAPIServerAuditWebhookConfig = "checksum/secret-kube-apiserver-audit-webhook-config"     // secret-kube-apiserver-audit-webhook-config.yaml
	ChecksumKeyKubeAPIServerBasicAuth          = "checksum/secret-kube-apiserver-basic-auth"               // secret-kube-apiserver-basic-auth.yaml
	ChecksumKeyKubeAPIServerAdmissionConfig    = "checksum/virtual-garden-kube-apiserver-admission-config" // configmap-kube-apiserver-admission-config.yaml CONDITION
	ChecksumKeyKubeControllerManagerClient     = "checksum/secret-kube-controller-manager-client"          // secret-kube-controller-manager-tls.yaml
	ChecksumKeyServiceAccountKey               = "checksum/secret-service-account-key"                     // secret-service-account-key.yaml
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
