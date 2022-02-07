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
	"reflect"

	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/virtual-garden/pkg/api"
	"github.com/gardener/virtual-garden/pkg/provider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	maxReplicas      = int32(6)
	minReplicas      = int32(3)
	vpaScaleUpMode   = hvpav1alpha1.UpdateModeOff
	vpaScaleDownMode = hvpav1alpha1.UpdateModeOff
)

var vpaScaleUpStabilization = api.ScaleType{
	MinChange: &hvpav1alpha1.ScaleParams{
		CPU: hvpav1alpha1.ChangeParams{
			Value:      pointer.StringPtr("350m"),
			Percentage: pointer.Int32Ptr(40),
		},
		Memory: hvpav1alpha1.ChangeParams{
			Value:      pointer.StringPtr("250M"),
			Percentage: pointer.Int32Ptr(87),
		},
	},
	StabilizationDuration: pointer.StringPtr("4m"),
}

var vpaScaleDownStabilization = api.ScaleType{
	MinChange: &hvpav1alpha1.ScaleParams{
		CPU: hvpav1alpha1.ChangeParams{
			Value:      pointer.StringPtr("360m"),
			Percentage: pointer.Int32Ptr(42),
		},
		Memory: hvpav1alpha1.ChangeParams{
			Value:      pointer.StringPtr("260M"),
			Percentage: pointer.Int32Ptr(89),
		},
	},
	StabilizationDuration: pointer.StringPtr("9m"),
}

var limitsRequestsGapScaleParams = hvpav1alpha1.ScaleParams{
	CPU: hvpav1alpha1.ChangeParams{
		Value:      pointer.StringPtr("320m"),
		Percentage: pointer.Int32Ptr(45),
	},
	Memory: hvpav1alpha1.ChangeParams{
		Value:      pointer.StringPtr("220M"),
		Percentage: pointer.Int32Ptr(81),
	},
}

var _ = Describe("Api Server pod auto scaling test", func() {

	It("Should create api auto scaling settings", func() {
		namespaceName := "apiserverhvpa"

		ctx := context.Background()
		defer ctx.Done()

		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		err := testenv.Client.Create(ctx, &namespace)
		Expect(err).To(BeNil())

		imports := getImportsForHvpa()

		infrastructureProvider, err := provider.NewInfrastructureProvider(api.InfrastructureProviderGCP)
		Expect(err).To(BeNil())

		operation := &operation{
			client:                 testenv.Client,
			log:                    testenv.Logger,
			infrastructureProvider: infrastructureProvider,
			backupProvider:         nil,
			namespace:              namespaceName,
			imports:                &imports,
			exports:                api.Exports{},
			imageRefs:              api.ImageRefs{},
		}

		err = DeployHVPACRD(ctx, testenv.Client)
		Expect(err).To(BeNil())

		err = DeployVPACRD(ctx, testenv.Client)
		Expect(err).To(BeNil())

		err = operation.deployKubeAPIServerPodAutoscaling(ctx)
		Expect(err).To(BeNil())

		hvpa := operation.emptyKubeAPIServerHvpa()

		err = testenv.Client.Get(ctx, client.ObjectKeyFromObject(hvpa), hvpa)
		Expect(err).To(BeNil())

		Expect(*hvpa.Spec.Replicas == 1).To(BeTrue())
		Expect(hvpa.Spec.Hpa.Template.Spec.MaxReplicas == maxReplicas).To(BeTrue())
		Expect(*hvpa.Spec.Hpa.Template.Spec.MinReplicas == minReplicas).To(BeTrue())
		Expect(*hvpa.Spec.Vpa.ScaleUp.UpdatePolicy.UpdateMode == vpaScaleUpMode).To(BeTrue())
		Expect(*hvpa.Spec.Vpa.ScaleDown.UpdatePolicy.UpdateMode == vpaScaleDownMode).To(BeTrue())
		Expect(reflect.DeepEqual(hvpa.Spec.Vpa.ScaleUp.MinChange, *vpaScaleUpStabilization.MinChange)).To(BeTrue())
		Expect(reflect.DeepEqual(*hvpa.Spec.Vpa.ScaleUp.StabilizationDuration, *vpaScaleUpStabilization.StabilizationDuration)).To(BeTrue())
		Expect(reflect.DeepEqual(hvpa.Spec.Vpa.ScaleDown.MinChange, *vpaScaleDownStabilization.MinChange)).To(BeTrue())
		Expect(reflect.DeepEqual(*hvpa.Spec.Vpa.ScaleDown.StabilizationDuration, *vpaScaleDownStabilization.StabilizationDuration)).To(BeTrue())
		Expect(reflect.DeepEqual(hvpa.Spec.Vpa.LimitsRequestsGapScaleParams, limitsRequestsGapScaleParams)).To(BeTrue())

		Expect(operation.deleteKubeAPIServerPodAutoscaling(ctx)).To(Succeed())
	})
})

func getHVPASettings() *api.HvpaConfig {
	return &api.HvpaConfig{
		MaxReplicas:                  pointer.Int32Ptr(maxReplicas),
		MinReplicas:                  pointer.Int32Ptr(minReplicas),
		VpaScaleUpMode:               pointer.StringPtr(vpaScaleUpMode),
		VpaScaleDownMode:             pointer.StringPtr(vpaScaleDownMode),
		VpaScaleUpStabilization:      &vpaScaleUpStabilization,
		VpaScaleDownStabilization:    &vpaScaleDownStabilization,
		LimitsRequestsGapScaleParams: &limitsRequestsGapScaleParams,
		MaintenanceWindow:            nil,
	}
}

func getImportsForHvpa() api.Imports {
	return api.Imports{
		RuntimeCluster:         lsv1alpha1.Target{},
		RuntimeClusterSettings: api.ClusterSettings{},
		VirtualGarden: api.VirtualGarden{
			ETCD: nil,
			KubeAPIServer: &api.KubeAPIServer{
				Replicas:        1,
				SNI:             nil,
				DnsAccessDomain: "com.our.test",
				GardenerControlplane: api.GardenerControlplane{
					ValidatingWebhook: api.AdmissionWebhookConfig{
						Token: api.AdmissionWebhookTokenConfig{Enabled: true},
					},
					MutatingWebhook: api.AdmissionWebhookConfig{
						Token: api.AdmissionWebhookTokenConfig{Enabled: true},
					},
				},
				ServiceAccountKeyPem:     pointer.String("test-service-account-key"),
				AuditWebhookConfig:       api.AuditWebhookConfig{Config: "testconfig"},
				AuditWebhookBatchMaxSize: "",
				SeedAuthorizer: api.SeedAuthorizer{
					Enabled:                  true,
					CertificateAuthorityData: "test-ca-data",
				},
				HVPAEnabled:             true,
				HVPA:                    getHVPASettings(),
				EventTTL:                nil,
				OidcIssuerURL:           nil,
				AdditionalVolumeMounts:  nil,
				AdditionalVolumes:       nil,
				HorizontalPodAutoscaler: nil,
			},
			DeleteNamespace:   false,
			PriorityClassName: "",
		},
	}
}
