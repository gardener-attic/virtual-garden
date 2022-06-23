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
	_ "embed"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/gardener/virtual-garden/pkg/api/helper"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// checkHVPACRD checks if the HVPA CRD is deployed if HVPA is enabled in the etcd or api server
func (o *operation) checkHVPACRD(ctx context.Context) error {
	o.log.Infof("Check if hvpa crd exists if hvpa is enabled")

	if helper.ETCDHVPAEnabled(o.imports.VirtualGarden.ETCD) || o.imports.VirtualGarden.KubeAPIServer.HVPAEnabled {
		_, err := o.getHVPACRD(ctx)
		return err
	}

	return nil
}

// getHVPACRD return the HVPA CRD.
func (o *operation) getHVPACRD(ctx context.Context) (*v1.CustomResourceDefinition, error) {
	hvpaCrd := emptyHVPACRD()
	err := o.client.Get(ctx, client.ObjectKeyFromObject(hvpaCrd), hvpaCrd)
	if err != nil {
		return nil, err
	}

	return hvpaCrd, nil
}

func emptyHVPACRD() *v1.CustomResourceDefinition {
	return &v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hvpas.autoscaling.k8s.io",
		},
	}
}
