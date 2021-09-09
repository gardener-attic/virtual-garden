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

package validation

import (
	gardenerutils "github.com/gardener/gardener/pkg/utils"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/virtual-garden/pkg/api"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateImports validates an Imports object.
func ValidateImports(obj *api.Imports) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateCluster(&obj.Cluster, field.NewPath("cluster"))...)
	allErrs = append(allErrs, ValidateHostingCluster(&obj.HostingCluster, field.NewPath("hostingCluster"))...)
	allErrs = append(allErrs, ValidateVirtualGarden(&obj.VirtualGarden, field.NewPath("virtualGarden"))...)

	return allErrs
}

// ValidInfrastructureProviderTypes is a set of valid infrastructure provider types.
var ValidInfrastructureProviderTypes = sets.NewString(
	string(api.InfrastructureProviderAWS),
	string(api.InfrastructureProviderGCP),
	string(api.InfrastructureProviderAlicloud),
	string(api.InfrastructureProviderFake),
)

// ValidateCluster validates the cluster.
func ValidateCluster(obj *lsv1alpha1.Target, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj == nil {
		allErrs = append(allErrs, field.Required(fldPath, "target is required"))
	} else if len(obj.Spec.Configuration.RawMessage) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, "kubeconfig is required"))
	}

	return allErrs
}

// ValidateHostingCluster validates a HostingCluster object.
func ValidateHostingCluster(obj *api.HostingCluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(obj.Namespace) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), "namespace for deployment in hosting cluster is required"))
	}
	if !ValidInfrastructureProviderTypes.Has(string(obj.InfrastructureProvider)) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("infrastructureProvider"), obj.InfrastructureProvider, ValidInfrastructureProviderTypes.UnsortedList()))
	}

	return allErrs
}

// ValidateVirtualGarden validates a VirtualGarden object.
func ValidateVirtualGarden(obj *api.VirtualGarden, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.ETCD != nil {
		if obj.ETCD.StorageClassName != nil && len(*obj.ETCD.StorageClassName) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("etcd", "storageClassName"), "storage class name cannot be empty if key is provided"))
		}
		if obj.ETCD.Backup != nil {
			allErrs = append(allErrs, ValidateETCDBackup(obj.ETCD.Backup, fldPath.Child("etcd", "backup"))...)
		}
	}

	if obj.KubeAPIServer != nil {
		if obj.KubeAPIServer.SNI != nil {
			allErrs = append(allErrs, ValidateSNI(obj.KubeAPIServer.SNI, fldPath.Child("exposure", "sni"))...)
		}

		if obj.KubeAPIServer.ServiceAccountKeyPem != nil && len(*obj.KubeAPIServer.ServiceAccountKeyPem) > 0 {
			allErrs = append(allErrs, ValidateRSAPrivateKey(obj.KubeAPIServer.ServiceAccountKeyPem, fldPath.Child("serviceAccountKeyPem"))...)
		}
	}

	return allErrs
}

// ValidateETCDBackup validates an ETCDBackup object.
func ValidateETCDBackup(obj *api.ETCDBackup, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if !ValidInfrastructureProviderTypes.Has(string(obj.InfrastructureProvider)) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("infrastructureProvider"), obj.InfrastructureProvider, ValidInfrastructureProviderTypes.UnsortedList()))
	}
	if len(obj.Region) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), "region must be given"))
	}
	if len(obj.BucketName) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("bucketName"), "bucketName must be given"))
	}
	if obj.Credentials == nil || len(obj.Credentials.Data) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("credentials"), "etcd backup credentials are emtpy"))
	}

	return allErrs
}

// ValidateSNI validates a SNI object.
func ValidateSNI(obj *api.SNI, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(obj.Hostname) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("hostnames"), "at least one hostname is required"))
	}
	if obj.TTL != nil && (*obj.TTL < 60 || *obj.TTL > 600) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ttl"), *obj.TTL, "ttl must be between 60 and 600"))
	}

	return allErrs
}

// ValidateRSAPrivateKey validates an RSA private key in PEM format
func ValidateRSAPrivateKey(obj *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	_, err := gardenerutils.DecodePrivateKey([]byte(*obj))
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "(value not logged)", err.Error()))
	}

	return allErrs
}
