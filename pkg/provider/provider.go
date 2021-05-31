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

package provider

import (
	"context"
	"fmt"

	"github.com/gardener/virtual-garden/pkg/api"
	"github.com/gardener/virtual-garden/pkg/provider/alicloud"
	"github.com/gardener/virtual-garden/pkg/provider/aws"
	"github.com/gardener/virtual-garden/pkg/provider/fake"
	"github.com/gardener/virtual-garden/pkg/provider/gcp"

	corev1 "k8s.io/api/core/v1"
)

// InfrastructureProvider is an interface for infrastructure providers.
type InfrastructureProvider interface {
	// ComputeStorageClassConfiguration shall return the Kubernetes volume provisioner name as well as optional
	// parameters for the storage class that can be used by etcd.
	ComputeStorageClassConfiguration() (provisioner string, parameters map[string]string)
	GetLoadBalancer(service *corev1.Service) string
	GetKubeAPIServerURL(kubeAPIServer *api.KubeAPIServer, loadBalancer string) string
}

// NewInfrastructureProvider returns a new InfrastructureProvider interface for the given provider type.
func NewInfrastructureProvider(providerType api.InfrastructureProviderType) (InfrastructureProvider, error) {
	switch providerType {
	case api.InfrastructureProviderAlicloud:
		return alicloud.NewInfrastructureProvider(), nil
	case api.InfrastructureProviderAWS:
		return aws.NewInfrastructureProvider(), nil
	case api.InfrastructureProviderGCP:
		return gcp.NewInfrastructureProvider(), nil
	}

	return nil, fmt.Errorf("unsupported infrastructure provider type: %q", providerType)
}

// BackupProvider is an interface for backup providers.
type BackupProvider interface {
	// CreateBucket shall create a blob storage bucket with the given name in the given region.
	CreateBucket(ctx context.Context, bucketName, region string) error
	// DeleteBucket shall delete a blob storage bucket and all its contents with the given name.
	DeleteBucket(ctx context.Context, bucketName string) error
	// BucketExists shall return whether the blob storage bucket exists.
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// ComputeETCDBackupConfiguration shall compute the configuration for the etcd-backup-restore sidecar container that
	// runs in the etcd statefulset. It takes the volume name of the etcd backup secret and should return the name of
	// the blob storage provider, the secret data for the etcd backup secret, and optional environment variables that
	// will be injected into the sidecar container.
	ComputeETCDBackupConfiguration(etcdBackupSecretVolumeName string) (storageProviderName string, secretData map[string][]byte, environment []corev1.EnvVar)
}

// NewBackupProvider returns a new InfrastructureProvider interface for the given provider type.
func NewBackupProvider(providerType api.InfrastructureProviderType, credentials map[string]api.Credentials,
	credentialsRef string) (BackupProvider, error) {
	creds, ok := credentials[credentialsRef]
	if !ok {
		return nil, fmt.Errorf("could not find referenced credentials with name %q", credentialsRef)
	}
	if creds.Type != providerType {
		return nil, fmt.Errorf("referenced credentials type %q does not match provider type %q", creds.Type, providerType)
	}

	switch providerType {
	case api.InfrastructureProviderGCP:
		return gcp.NewBackupProvider(creds.Data)
	case api.InfrastructureProviderFake:
		backupSecretData := map[string][]byte{}

		for k, v := range creds.Data {
			backupSecretData[k] = []byte(v)
		}
		return fake.NewBackupProvider(backupSecretData), nil
	}

	return nil, fmt.Errorf("unsupported backup provider type: %q", providerType)
}
