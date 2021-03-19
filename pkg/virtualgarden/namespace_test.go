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
	"fmt"

	mockclient "github.com/gardener/virtual-garden/pkg/mock/controller-runtime/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Namespace", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient

		ctx       = context.TODO()
		name      = "foo"
		namespace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
		fakeErr   = fmt.Errorf("fail")

		op *operation
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)

		op = &operation{client: c, namespace: name}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#CreateNamespace", func() {
		It("should succeed creating the namespace", func() {
			c.EXPECT().Create(ctx, namespace)
			Expect(op.CreateNamespace(ctx)).To(Succeed())
		})

		It("should do nothing if the namespace already exists", func() {
			c.EXPECT().Create(ctx, namespace).Return(apierrors.NewAlreadyExists(schema.GroupResource{}, name))
			Expect(op.CreateNamespace(ctx)).To(Succeed())
		})

		It("should fail creating the namespace if an unexpected error occurs", func() {
			c.EXPECT().Create(ctx, namespace).Return(fakeErr)
			Expect(op.CreateNamespace(ctx)).To(MatchError(fakeErr))
		})
	})

	Describe("#DeleteNamespace", func() {
		It("should succeed deleting the namespace", func() {
			c.EXPECT().Delete(ctx, namespace)
			Expect(op.DeleteNamespace(ctx)).To(Succeed())
		})

		It("should succeed deleting the namespace if it does not exist", func() {
			c.EXPECT().Delete(ctx, namespace).Return(apierrors.NewNotFound(schema.GroupResource{}, name))
			Expect(op.DeleteNamespace(ctx)).To(Succeed())
		})

		It("should fail deleting the namespace if an unexpected error occurs", func() {
			c.EXPECT().Delete(ctx, namespace).Return(fakeErr)
			Expect(op.DeleteNamespace(ctx)).To(MatchError(fakeErr))
		})
	})
})
