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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	KubeApiServerSecretNameAdmissionKubeconfig = Prefix + "-kube-apiserver-admission-kubeconfig"
)

//go:embed resources/validating-webhook-kubeconfig.yaml
var validatingWebhookKubeconfig []byte

//go:embed resources/mutating-webhook-kubeconfig.yaml
var mutatingWebhookKubeconfig []byte

func (o *operation) deploySecrets(ctx context.Context) error {
	if err := o.deployKubeApiServerSecretAdmissionKubeconfig(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deployKubeApiServerSecretAdmissionKubeconfig(ctx context.Context) error {
	controlplane := o.imports.VirtualGarden.KubeAPIServer.GardenerControlplane
	if !controlplane.ValidatingWebhookEnabled && !controlplane.MutatingWebhookEnabled {
		return nil
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: KubeApiServerSecretNameAdmissionKubeconfig, Namespace: o.namespace}}

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data["validating-webhook"] = validatingWebhookKubeconfig
		secret.Data["mutating-webhook"] = mutatingWebhookKubeconfig
		return nil
	})

	return err
}
