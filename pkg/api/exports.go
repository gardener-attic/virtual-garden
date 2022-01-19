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

package api

// Exports defines the structure for the exported data which might be consumed by other components.
type Exports struct {
	VirtualGardenApiserverCaPem string `json:"virtualGardenApiserverCaPem,omitempty" yaml:"virtualGardenApiserverCaPem,omitempty"`
	ServiceAccountKeyPem        string `json:"serviceAccountKeyPem,omitempty" yaml:"serviceAccountKeyPem,omitempty"`
	EtcdCaPem             		string `json:"etcdCaPem,omitempty" yaml:"etcdCaPem,omitempty"`
	EtcdClientTlsPem      		string `json:"etcdClientTlsPem,omitempty" yaml:"etcdClientTlsPem,omitempty"`
	EtcdClientTlsKeyPem   		string `json:"etcdClientTlsKeyPem,omitempty" yaml:"etcdClientTlsKeyPem,omitempty"`
	EtcdUrl   					string `json:"etcdUrl,omitempty" yaml:"etcdUrl,omitempty"`
	KubeconfigYaml        		string `json:"kubeconfigYaml,omitempty" yaml:"kubeconfigYaml,omitempty"`
	VirtualGardenEndpoint 		string `json:"virtualGardenEndpoint,omitempty" yaml:"virtualGardenEndpoint,omitempty"`
}
