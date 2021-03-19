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

package virtualgarden_test

import (
	. "github.com/gardener/virtual-garden/pkg/virtualgarden"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EtcdStatefulset", func() {
	role := "foobar"

	Describe("#ETCDStatefulSetName", func() {
		It("should return the correct stateful set name", func() {
			Expect(ETCDStatefulSetName(role)).To(Equal("virtual-garden-etcd-" + role))
		})
	})

	Describe("#ETCDPersistentVolumeClaimName", func() {
		It("should return the correct pvc name for the main role", func() {
			Expect(ETCDPersistentVolumeClaimName(ETCDRoleMain)).To(Equal("main-virtual-garden-etcd-virtual-garden-etcd-main-0"))
		})

		It("should return the correct pvc name for the events role", func() {
			Expect(ETCDPersistentVolumeClaimName(ETCDRoleEvents)).To(Equal("virtual-garden-etcd-events-virtual-garden-etcd-events-0"))
		})

		It("should return the correct pvc name for any other role", func() {
			Expect(ETCDPersistentVolumeClaimName(role)).To(Equal("virtual-garden-etcd-" + role + "-virtual-garden-etcd-" + role + "-0"))
		})
	})

	Describe("#ETCDDataVolumeName", func() {
		It("should return the correct data volume name for the main role", func() {
			Expect(ETCDDataVolumeName(ETCDRoleMain)).To(Equal("main-virtual-garden-etcd"))
		})

		It("should return the correct data volume name for the events role", func() {
			Expect(ETCDDataVolumeName(ETCDRoleEvents)).To(Equal("virtual-garden-etcd-events"))
		})

		It("should return the correct data volume name for any other role", func() {
			Expect(ETCDDataVolumeName(role)).To(Equal("virtual-garden-etcd-" + role))
		})
	})
})
