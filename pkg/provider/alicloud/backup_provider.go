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

package alicloud

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	corev1 "k8s.io/api/core/v1"
)

const (
	envVarAccessKeyID     = "ALICLOUD_ACCESS_KEY_ID"
	envVarAccessKeySecret = "ALICLOUD_ACCESS_KEY_SECRET"
	envVarStorageEndpoint = "ALICLOUD_ENDPOINT"

	DataKeyAccessKeyID     = "accessKeyID"
	DataKeyAccessKeySecret = "accessKeySecret"
	DataKeyStorageEndpoint = "storageEndpoint"

	storageProviderNameOSS = "OSS"

	ossErrorCodeNoSuchBucket = "NoSuchBucket"
)

type backupProvider struct {
	bucketName      string
	accessKeyID     string
	secretAccessKey string
	storageEndpoint string
	log             *logrus.Logger
}

// NewBackupProvider creates a new oss backup provider implementation.
func NewBackupProvider(
	credentialsData map[string]string,
	bucketName, storageEndpoint string,
	log *logrus.Logger,
) (*backupProvider, error) {
	accessKeyID, ok := credentialsData[DataKeyAccessKeyID]
	if !ok {
		return nil, fmt.Errorf("data map doesn't have an access key id")
	}

	secretAccessKey, ok := credentialsData[DataKeyAccessKeySecret]
	if !ok {
		return nil, fmt.Errorf("data map doesn't have an access key secret")
	}

	return &backupProvider{
		bucketName:      bucketName,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		storageEndpoint: storageEndpoint,
		log:             log,
	}, nil
}

func (b *backupProvider) CreateBucket(ctx context.Context) error {
	svc, err := b.getClient()
	if err != nil {
		return err
	}

	b.log.Infof("Ensuring that alicloud backup bucket %q exists", b.bucketName)
	err = svc.CreateBucket(b.bucketName, oss.ACL(oss.ACLPrivate))
	if err != nil {
		return fmt.Errorf("failed to create alicloud backup bucket %q: %w", b.bucketName, err)
	}

	return nil
}

func (b *backupProvider) DeleteBucket(ctx context.Context) error {
	svc, err := b.getClient()
	if err != nil {
		return err
	}

	bucket, err := svc.Bucket(b.bucketName)
	if err != nil {
		return fmt.Errorf("failed to get alicloud backup bucket %q: %w", b.bucketName, err)
	}

	b.log.Infof("Deleting objects of alicloud backup bucket %q", b.bucketName)
	marker := ""
	for {
		b.log.Infof("Deleting objects of bucket %q starting with marker %q", b.bucketName, marker)
		listResult, err := bucket.ListObjects(oss.Marker(marker))
		if err != nil {
			if ossErr, ok := err.(oss.ServiceError); ok {
				switch ossErr.Code {
				case ossErrorCodeNoSuchBucket:
					return nil
				default:
					return fmt.Errorf("failed to list objects of alicloud backup bucket %q: %w", b.bucketName, err)
				}
			}

			return fmt.Errorf("failed to list objects of alicloud backup bucket %q: %w", b.bucketName, err)
		}

		for _, object := range listResult.Objects {
			err = bucket.DeleteObject(object.Key)
			if err != nil {
				return fmt.Errorf("failed to delete object %q of alicloud backup bucket %q: %w", object.Key, b.bucketName, err)
			}
		}

		if listResult.IsTruncated {
			marker = listResult.NextMarker
		} else {
			break
		}
	}

	b.log.Infof("Deleting alicloud backup bucket %q", b.bucketName)
	err = svc.DeleteBucket(b.bucketName)
	if err != nil {
		if ossErr, ok := err.(oss.ServiceError); ok {
			switch ossErr.Code {
			case ossErrorCodeNoSuchBucket:
				return nil
			default:
				return fmt.Errorf("failed to delete alicloud backup bucket %q: %w", b.bucketName, err)
			}
		}

		return fmt.Errorf("failed to delete alicloud backup bucket %q: %w", b.bucketName, err)
	}

	return nil
}

func (b *backupProvider) BucketExists(_ context.Context) (bool, error) {
	svc, err := b.getClient()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of alicloud backup bucket %q: %w", b.bucketName, err)
	}

	return svc.IsBucketExist(b.bucketName)
}

func (b *backupProvider) ComputeETCDBackupConfiguration(_, etcdSecretNameBackup string) (
	storageProviderName string,
	secretData map[string][]byte,
	environment []corev1.EnvVar,
) {
	storageProviderName = storageProviderNameOSS

	secretData = map[string][]byte{
		DataKeyAccessKeyID:     []byte(b.accessKeyID),
		DataKeyAccessKeySecret: []byte(b.secretAccessKey),
		DataKeyStorageEndpoint: []byte(b.storageEndpoint),
	}

	environment = []corev1.EnvVar{
		b.envVar(envVarAccessKeyID, etcdSecretNameBackup, DataKeyAccessKeyID),
		b.envVar(envVarAccessKeySecret, etcdSecretNameBackup, DataKeyAccessKeySecret),
		b.envVar(envVarStorageEndpoint, etcdSecretNameBackup, DataKeyStorageEndpoint),
	}

	return
}

func (b *backupProvider) envVar(envVarName, etcdSecretNameBackup, dataKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envVarName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: etcdSecretNameBackup,
				},
				Key: dataKey,
			},
		},
	}
}

func (b *backupProvider) getClient() (*oss.Client, error) {
	client, err := oss.New(b.storageEndpoint, b.accessKeyID, b.secretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get alicloud client to manage backup bucket: %w", err)
	}

	return client, nil
}
