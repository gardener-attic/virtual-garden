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

package helper_test

import (
	"github.com/gardener/virtual-garden/pkg/api"
	. "github.com/gardener/virtual-garden/pkg/api/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Helper", func() {
	DescribeTable("#KubeAPIServerSNIEnabled",
		func(kubeAPIServer *api.KubeAPIServer, matcher types.GomegaMatcher) {
			Expect(KubeAPIServerSNIEnabled(kubeAPIServer)).To(matcher)
		},

		Entry("kubeAPIServer is nil", nil, BeFalse()),
		Entry("exposure is nil", &api.KubeAPIServer{}, BeFalse()),
		Entry("SNI is not nil", &api.KubeAPIServer{SNI: &api.SNI{}}, BeTrue()),
	)

	DescribeTable("#ETCDBackupEnabled",
		func(etcd *api.ETCD, matcher types.GomegaMatcher) {
			Expect(ETCDBackupEnabled(etcd)).To(matcher)
		},

		Entry("etcd is nil", nil, BeFalse()),
		Entry("backup is nil", &api.ETCD{}, BeFalse()),
		Entry("backup is not nil", &api.ETCD{Backup: &api.ETCDBackup{}}, BeTrue()),
	)
})
