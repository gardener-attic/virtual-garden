package api

import (
	"github.com/gardener/hvpa-controller/api/v1alpha1"
	"k8s.io/utils/pointer"
)

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

func (h *HvpaConfig) GetMaxReplicas(def int32) int32 {
	if h == nil || h.MaxReplicas == nil {
		return def
	}
	return *h.MaxReplicas
}

func (h *HvpaConfig) GetMinReplicas() *int32 {
	var def int32 = 1
	if h == nil || h.MinReplicas == nil {
		return pointer.Int32Ptr(def)
	}
	return h.MinReplicas
}

func (h *HvpaConfig) GetTargetAverageUtilizationMemory(def int32) *int32 {
	if h == nil || h.TargetAverageUtilizationMemory == nil {
		return pointer.Int32Ptr(def)
	}
	return h.TargetAverageUtilizationMemory
}

func (h *HvpaConfig) GetTargetAverageUtilizationCpu(def int32) *int32 {
	if h == nil || h.TargetAverageUtilizationCpu == nil {
		return pointer.Int32Ptr(def)
	}
	return h.TargetAverageUtilizationCpu
}

func (h *HvpaConfig) GetVpaScaleUpMode(def string) *string {
	if h == nil || h.VpaScaleUpMode == nil {
		return pointer.StringPtr(def)
	}
	return h.VpaScaleUpMode
}

func (h *HvpaConfig) GetVpaScaleUpStabilisationDuration(def string) *string {
	if h == nil || h.VpaScaleUpStabilization == nil {
		return pointer.StringPtr(def)
	}
	return h.VpaScaleUpStabilization.StabilizationDuration
}

func (h *HvpaConfig) GetVpaScaleUpMinChange(def v1alpha1.ScaleParams) v1alpha1.ScaleParams {
	if h == nil || h.VpaScaleUpStabilization == nil || h.VpaScaleUpStabilization.MinChange == nil {
		return def
	}
	return *h.VpaScaleUpStabilization.MinChange
}

func (h *HvpaConfig) GetVpaScaleDownMode(def string) *string {
	if h == nil || h.VpaScaleDownMode == nil {
		return pointer.StringPtr(def)
	}
	return h.VpaScaleDownMode
}

func (h *HvpaConfig) GetVpaScaleDownStabilisationDuration(def string) *string {
	if h == nil || h.VpaScaleDownStabilization == nil {
		return pointer.StringPtr(def)
	}
	return h.VpaScaleDownStabilization.StabilizationDuration
}

func (h *HvpaConfig) GetVpaScaleDownMinChange(def v1alpha1.ScaleParams) v1alpha1.ScaleParams {
	if h == nil || h.VpaScaleDownStabilization == nil || h.VpaScaleDownStabilization.MinChange == nil {
		return def
	}
	return *h.VpaScaleDownStabilization.MinChange
}

func (h *HvpaConfig) GetLimitsRequestsGapScaleParams(def v1alpha1.ScaleParams) v1alpha1.ScaleParams {
	if h == nil || h.LimitsRequestsGapScaleParams == nil {
		return def
	}
	return *h.LimitsRequestsGapScaleParams
}
