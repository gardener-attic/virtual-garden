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

	"github.com/gardener/virtual-garden/pkg/api"
	mockprovider "github.com/gardener/virtual-garden/pkg/provider/mock"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EtcdBackup", func() {
	var (
		ctrl           *gomock.Controller
		backupProvider *mockprovider.MockBackupProvider

		ctx        = context.TODO()
		fakeErr    = fmt.Errorf("fail")
		bucketName = "foobucket"
		region     = "europe"

		op *operation
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		backupProvider = mockprovider.NewMockBackupProvider(ctrl)

		op = &operation{
			backupProvider: backupProvider,
			imports: &api.Imports{
				VirtualGarden: api.VirtualGarden{
					ETCD: &api.ETCD{
						Backup: &api.ETCDBackup{
							BucketName: bucketName,
							Region:     region,
						},
					},
				},
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DeployBackupBucket", func() {
		It("should create a bucket", func() {
			backupProvider.EXPECT().CreateBucket(ctx)
			Expect(op.DeployBackupBucket(ctx)).To(Succeed())
		})

		It("should return the error occurred during creation", func() {
			backupProvider.EXPECT().CreateBucket(ctx).Return(fakeErr)
			Expect(op.DeployBackupBucket(ctx)).To(MatchError(fakeErr))
		})
	})

	Describe("#DeleteBackupBucket", func() {
		It("should delete a bucket", func() {
			backupProvider.EXPECT().DeleteBucket(ctx)
			Expect(op.DeleteBackupBucket(ctx)).To(Succeed())
		})

		It("should return the error occurred during deletion", func() {
			backupProvider.EXPECT().DeleteBucket(ctx).Return(fakeErr)
			Expect(op.DeleteBackupBucket(ctx)).To(MatchError(fakeErr))
		})
	})
})
