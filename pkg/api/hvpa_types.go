package api

import "github.com/gardener/hvpa-controller/api/v1alpha1"

// hvpa:
//  enabled: false
//  maxReplicas: 6
//  minReplicas: 1
//  targetAverageUtilizationCpu: 80
//  targetAverageUtilizationMemory: 80
//  vpaScaleUpMode: "Auto"
//  vpaScaleDownMode: "Auto"

//  vpaScaleUpStabilization:
//    stabilizationDuration: "3m"
//    minChange:
//      cpu:
//        value: 300m
//        percentage: 80
//      memory:
//        value: 200M
//        percentage: 80
//  vpaScaleDownStabilization:
//    stabilizationDuration: "15m"
//    minChange:
//      cpu:
//        value: 600m
//        percentage: 80
//      memory:
//        value: 600M
//        percentage: 80
//  limitsRequestsGapScaleParams:
//    cpu:
//      value: "1"
//      percentage: 40
//    memory:
//      value: "1G"
//      percentage: 40

type HvpaConfig struct {
	MaxReplicas                    *int32       `json:"maxReplicas,omitempty" yaml:"maxReplicas,omitempty"`
	MinReplicas                    *int32       `json:"minReplicas,omitempty" yaml:"minReplicas,omitempty"`
	TargetAverageUtilizationCpu    *int32       `json:"targetAverageUtilizationCpu,omitempty" yaml:"targetAverageUtilizationCpu,omitempty"`
	TargetAverageUtilizationMemory *int32       `json:"targetAverageUtilizationMemory,omitempty" yaml:"targetAverageUtilizationMemory,omitempty"`
	VpaScaleUpMode                 *string      `json:"vpaScaleUpMode,omitempty" yaml:"vpaScaleUpMode,omitempty"`
	VpaScaleDownMode               *string      `json:"vpaScaleDownMode,omitempty" yaml:"vpaScaleDownMode,omitempty"`
	VpaScaleUpStabilization        *ScaleType   `json:"vpaScaleUpStabilization,omitempty" yaml:"vpaScaleUpStabilization,omitempty"`
	VpaScaleDownStabilization      *ScaleType   `json:"vpaScaleDownStabilization,omitempty" yaml:"vpaScaleDownStabilization,omitempty"`
	LimitsRequestsGapScaleParams   *ScaleParams `json:"limitsRequestsGapScaleParams,omitempty" yaml:"limitsRequestsGapScaleParams,omitempty"`
}

type ScaleType struct {
	StabilizationDuration *string      `json:"stabilizationDuration,omitempty" yaml:"stabilizationDuration,omitempty"`
	MinChange             *ScaleParams `json:"minChange,omitempty" yaml:"minChange,omitempty"`
}

type ScaleParams struct {
	// Scale parameters for CPU
	CPU v1alpha1.ChangeParams `json:"cpu,omitempty"`
	// Scale parameters for memory
	Memory v1alpha1.ChangeParams `json:"memory,omitempty"`
}
