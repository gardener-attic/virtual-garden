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
	//"embed"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//go:embed resources/hvpa.yaml
var hvpaCrd []byte

// DeployKubeAPIServer deploys a kubernetes api server.
func (o *operation) deployHVPACrd(ctx context.Context) error {
	crd := &v1beta1.CustomResourceDefinition{}
	return nil
}
