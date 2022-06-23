module github.com/gardener/virtual-garden

go 1.16

require (
	cloud.google.com/go/storage v1.10.0
	github.com/aliyun/aliyun-oss-go-sdk v2.1.10+incompatible
	github.com/aws/aws-sdk-go v1.40.47
	github.com/gardener/component-cli v0.36.0
	github.com/gardener/component-spec/bindings-go v0.0.53
	github.com/gardener/gardener v1.26.0
	github.com/gardener/landscaper/apis v0.20.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.6.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/tools v0.1.3 // indirect
	google.golang.org/api v0.44.0
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/component-base v0.21.2
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/gardener/landscaper/apis => github.com/gardener/landscaper/apis v0.20.0
	k8s.io/api => k8s.io/api v0.21.2
	k8s.io/client-go => k8s.io/client-go v0.21.2
)
