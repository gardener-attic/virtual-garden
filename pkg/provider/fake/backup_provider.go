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

package fake

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

var FakeEnv = []corev1.EnvVar{
	{
		Name:  "testname",
		Value: "testval",
	},
}

const FakeProviderName = "fakeprovider"

type fakeBackupProvider struct {
	backupSecretData map[string][]byte
}

// NewBackupProvider returns a fake backup provider for testing.
func NewBackupProvider(backupSecretData map[string][]byte) *fakeBackupProvider {
	return &fakeBackupProvider{backupSecretData}
}

func (f *fakeBackupProvider) CreateBucket(_ context.Context) error {
	return nil
}

func (f *fakeBackupProvider) DeleteBucket(_ context.Context) error {
	return nil
}

func (f *fakeBackupProvider) BucketExists(_ context.Context) (bool, error) {
	return false, nil
}

func (f *fakeBackupProvider) ComputeETCDBackupConfiguration(_, _ string) (string, map[string][]byte, []corev1.EnvVar) {
	return FakeProviderName, f.backupSecretData, FakeEnv
}
