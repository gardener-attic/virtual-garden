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

package gcp

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
)

type backupProvider struct {
	storageClient      *storage.Client
	serviceAccountJSON []byte
	projectID          string
	bucketName         string
	region             string
}

// NewBackupProvider creates a new GCP backup provider implementation from the given service account JSON.
func NewBackupProvider(credentialsData map[string]string, bucketName, region string) (*backupProvider, error) {
	serviceAccountJSON, err := ReadServiceAccount(credentialsData)
	if err != nil {
		return nil, err
	}

	projectID, err := ExtractServiceAccountProjectID(serviceAccountJSON)
	if err != nil {
		return nil, err
	}

	return &backupProvider{
		serviceAccountJSON: []byte(serviceAccountJSON),
		projectID:          projectID,
		bucketName:         bucketName,
		region:             region,
	}, nil
}

const (
	errCodeBucketNotFound          = 404
	errCodeBucketAlreadyOwnedByYou = 409
)

func (b *backupProvider) initializeStorageClient(ctx context.Context) error {
	if b.storageClient != nil {
		return nil
	}

	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(b.serviceAccountJSON), option.WithScopes(storage.ScopeFullControl))
	if err != nil {
		return err
	}
	b.storageClient = storageClient

	return nil
}

func (b *backupProvider) CreateBucket(ctx context.Context) error {
	if b.storageClient == nil {
		if err := b.initializeStorageClient(ctx); err != nil {
			return err
		}
	}

	if err := b.storageClient.Bucket(b.bucketName).Create(ctx, b.projectID, &storage.BucketAttrs{
		Name:     b.bucketName,
		Location: b.region,
	}); err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == errCodeBucketAlreadyOwnedByYou {
			return nil
		}
		return err
	}
	return nil
}

func (b *backupProvider) DeleteBucket(ctx context.Context) error {
	if b.storageClient == nil {
		if err := b.initializeStorageClient(ctx); err != nil {
			return err
		}
	}

	if err := deleteAllObjects(ctx, b.storageClient, b.bucketName); err != nil && err != storage.ErrBucketNotExist {
		return err
	}

	if err := b.storageClient.Bucket(b.bucketName).Delete(ctx); err != nil {
		gerr, ok := err.(*googleapi.Error)
		if ok && gerr.Code == errCodeBucketNotFound {
			return nil
		}

		return err
	}

	return nil
}

func (b *backupProvider) BucketExists(ctx context.Context) (bool, error) {
	if b.storageClient == nil {
		if err := b.initializeStorageClient(ctx); err != nil {
			return false, err
		}
	}

	if _, err := b.storageClient.Bucket(b.bucketName).Attrs(ctx); err != nil {
		if err != storage.ErrBucketNotExist {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func deleteAllObjects(ctx context.Context, storageClient *storage.Client, bucketName string) error {
	var (
		bucketHandle = storageClient.Bucket(bucketName)
		itr          = bucketHandle.Objects(ctx, &storage.Query{})
	)

	for {
		attr, err := itr.Next()
		if err != nil {
			if err == iterator.Done {
				return nil
			}
			return err
		}

		if err := bucketHandle.Object(attr.Name).Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
			return err
		}
	}
}

func (b *backupProvider) ComputeETCDBackupConfiguration(etcdBackupSecretVolumeMountPath, _ string) (storageProviderName string, secretData map[string][]byte, environment []corev1.EnvVar) {
	storageProviderName = "GCS"
	secretData = map[string][]byte{DataKeyServiceAccountJSON: b.serviceAccountJSON}
	environment = []corev1.EnvVar{{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: etcdBackupSecretVolumeMountPath + "/" + DataKeyServiceAccountJSON,
	}}
	return
}
