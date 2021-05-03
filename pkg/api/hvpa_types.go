package api

import (
	"github.com/gardener/hvpa-controller/api/v1alpha1"
)

type HvpaConfig struct {
	MaxReplicas                    *int32                          `json:"maxReplicas,omitempty" yaml:"maxReplicas,omitempty"`
	MinReplicas                    *int32                          `json:"minReplicas,omitempty" yaml:"minReplicas,omitempty"`
	TargetAverageUtilizationCpu    *int32                          `json:"targetAverageUtilizationCpu,omitempty" yaml:"targetAverageUtilizationCpu,omitempty"`
	TargetAverageUtilizationMemory *int32                          `json:"targetAverageUtilizationMemory,omitempty" yaml:"targetAverageUtilizationMemory,omitempty"`
	VpaScaleUpMode                 *string                         `json:"vpaScaleUpMode,omitempty" yaml:"vpaScaleUpMode,omitempty"`
	VpaScaleDownMode               *string                         `json:"vpaScaleDownMode,omitempty" yaml:"vpaScaleDownMode,omitempty"`
	VpaScaleUpStabilization        *ScaleType                      `json:"vpaScaleUpStabilization,omitempty" yaml:"vpaScaleUpStabilization,omitempty"`
	VpaScaleDownStabilization      *ScaleType                      `json:"vpaScaleDownStabilization,omitempty" yaml:"vpaScaleDownStabilization,omitempty"`
	LimitsRequestsGapScaleParams   *v1alpha1.ScaleParams           `json:"limitsRequestsGapScaleParams,omitempty" yaml:"limitsRequestsGapScaleParams,omitempty"`
	MaintenanceWindow              *v1alpha1.MaintenanceTimeWindow `json:"maintenanceWindow,omitempty" yaml:"maintenanceWindow,omitempty"`
}

type ScaleType struct {
	StabilizationDuration *string               `json:"stabilizationDuration,omitempty" yaml:"stabilizationDuration,omitempty"`
	MinChange             *v1alpha1.ScaleParams `json:"minChange,omitempty" yaml:"minChange,omitempty"`
}
