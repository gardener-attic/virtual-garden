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

const kubeAPIServerContainerName = "kube-apiserver"
