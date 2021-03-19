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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespace creates the namespace in case it does not exist yet. Otherwise, it does nothing.
func (o *operation) CreateNamespace(ctx context.Context) error {
	if err := o.client.Create(ctx, emptyNamespace(o.namespace)); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// DeleteNamespace deletes the namespace.
func (o *operation) DeleteNamespace(ctx context.Context) error {
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyNamespace(o.namespace)))
}

func emptyNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}
