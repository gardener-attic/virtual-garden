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
	"fmt"

	"github.com/gardener/virtual-garden/pkg/api"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateImports validates an Imports object.
func ValidateImports(obj *api.Imports) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateHostingCluster(&obj.HostingCluster, field.NewPath("hostingCluster"))...)
	allErrs = append(allErrs, ValidateVirtualGarden(&obj.VirtualGarden, obj.Credentials, field.NewPath("virtualGarden"))...)
	for name, credentials := range obj.Credentials {
		allErrs = append(allErrs, ValidateCredentials(credentials, field.NewPath("credentials", name))...)
	}

	return allErrs
}

// ValidInfrastructureProviderTypes is a set of valid infrastructure provider types.
var ValidInfrastructureProviderTypes = sets.NewString(string(api.InfrastructureProviderAWS), string(api.InfrastructureProviderGCP))

// ValidateHostingCluster validates a HostingCluster object.
func ValidateHostingCluster(obj *api.HostingCluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(obj.Kubeconfig) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubeconfig"), "kubeconfig of hosting cluster is required"))
	}
	if len(obj.Namespace) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), "namespace for deployment in hosting cluster is required"))
	}
	if !ValidInfrastructureProviderTypes.Has(string(obj.InfrastructureProvider)) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("infrastructureProvider"), obj.InfrastructureProvider, ValidInfrastructureProviderTypes.UnsortedList()))
	}

	return allErrs
}

// ValidateVirtualGarden validates a VirtualGarden object.
func ValidateVirtualGarden(obj *api.VirtualGarden, credentials map[string]api.Credentials, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.ETCD != nil {
		if obj.ETCD.StorageClassName != nil && len(*obj.ETCD.StorageClassName) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("etcd", "storageClassName"), "storage class name cannot be empty if key is provided"))
		}
		if obj.ETCD.Backup != nil {
			allErrs = append(allErrs, ValidateETCDBackup(obj.ETCD.Backup, credentials, fldPath.Child("etcd", "backup"))...)
		}
	}

	if obj.KubeAPIServer != nil {
		if obj.KubeAPIServer.Exposure != nil && obj.KubeAPIServer.Exposure.SNI != nil {
			allErrs = append(allErrs, ValidateSNI(obj.KubeAPIServer.Exposure.SNI, fldPath.Child("exposure", "sni"))...)
		}
	}

	return allErrs
}

// ValidateCredentials validates a Credentials object.
func ValidateCredentials(obj api.Credentials, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(obj.Type) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("type"), "type must be given"))
	}
	if len(obj.Data) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("data"), "at least one key-value pair must be given"))
	}

	return allErrs
}

// ValidateETCDBackup validates an ETCDBackup object.
func ValidateETCDBackup(obj *api.ETCDBackup, credentials map[string]api.Credentials, fldPath *field.Path) field.ErrorList {
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
	if len(obj.CredentialsRef) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("credentialsRef"), "credentialsRef must be given"))
	} else if credentials, ok := credentials[obj.CredentialsRef]; !ok {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("credentialsRef"), obj.CredentialsRef, fmt.Sprintf("%q was not found in .credentials", obj.CredentialsRef)))
	} else if credentials.Type != obj.InfrastructureProvider {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("credentialsRef"), obj.CredentialsRef, fmt.Sprintf("referenced credentials are not of type %q but %q", obj.InfrastructureProvider, credentials.Type)))
	}

	return allErrs
}

// ValidateSNI validates a SNI object.
func ValidateSNI(obj *api.SNI, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(obj.Hostnames) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("hostnames"), "at least one hostname is required"))
	}
	if obj.TTL != nil && (*obj.TTL < 60 || *obj.TTL > 600) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ttl"), *obj.TTL, "ttl must be between 60 and 600"))
	}

	return allErrs
}
