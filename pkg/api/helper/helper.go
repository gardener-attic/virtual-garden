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

package helper

import (
	"github.com/gardener/virtual-garden/pkg/api"
)

// KubeAPIServerSNIEnabled returns true if SNI is enabled, otherwise false.
func KubeAPIServerSNIEnabled(kubeAPIServer *api.KubeAPIServer) bool {
	return kubeAPIServer != nil && kubeAPIServer.Exposure != nil && kubeAPIServer.Exposure.SNI != nil
}

// ETCDBackupEnabled returns true if the etcd backup settings are configured.
func ETCDBackupEnabled(etcd *api.ETCD) bool {
	return etcd != nil && etcd.Backup != nil
}

// ETCDHVPAEnabled returns true if the etcd HVPA settings are configured.
func ETCDHVPAEnabled(etcd *api.ETCD) bool {
	return etcd != nil && etcd.HVPAEnabled
}
