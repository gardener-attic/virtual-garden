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
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed resources/hvpa.yaml
var hvpaCrd []byte

// deployHVPACRD deploys the HVPA CRD.
func (o *operation) deployHVPACRD(ctx context.Context) error {
	o.log.Infof("Deploying hvpa crd")

	newCrd := &v1beta1.CustomResourceDefinition{}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(hvpaCrd), 32)
	err := decoder.Decode(newCrd)
	if err != nil {
		return fmt.Errorf("failed to decode HVPA CRD: %w", err)
	}

	crd := emptyHVPACRD()
	_, err = controllerutil.CreateOrUpdate(ctx, o.client, crd, func() error {
		crd.Spec = newCrd.Spec
		return nil
	})

	return err
}

// deleteHPVACRD deletes the HPVA CRD.
func (o *operation) deleteHPVACRD(ctx context.Context) error {
	o.log.Infof("Deleting hvpa crd")
	return client.IgnoreNotFound(o.client.Delete(ctx, emptyHVPACRD()))
}

func emptyHVPACRD() *v1beta1.CustomResourceDefinition {
	return &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hvpas.autoscaling.k8s.io",
		},
	}
}
