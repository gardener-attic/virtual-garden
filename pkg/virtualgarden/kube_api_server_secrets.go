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
	"fmt"

	"github.com/gardener/gardener/pkg/utils"
	"github.com/gardener/virtual-garden/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	KubeApiServerSecretNameAdmissionKubeconfig = Prefix + "-kube-apiserver-admission-kubeconfig"
	KubeApiServerSecretNameAuditWebhookConfig  = "kube-apiserver-audit-webhook-config"
	KubeApiServerSecretNameBasicAuth           = Prefix + "-kube-apiserver-basic-auth"
)

//go:embed resources/validating-webhook-kubeconfig.yaml
var validatingWebhookKubeconfig []byte

//go:embed resources/mutating-webhook-kubeconfig.yaml
var mutatingWebhookKubeconfig []byte

func (o *operation) deployKubeAPIServerSecrets(ctx context.Context) error {
	if err := o.deployKubeApiServerSecretAdmissionKubeconfig(ctx); err != nil {
		return err
	}

	if err := o.deployKubeApiServerSecretAuditWebhookConfig(ctx); err != nil {
		return err
	}

	if err := o.deployKubeApiServerSecretBasicAuth(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) deleteKubeAPIServerSecrets(ctx context.Context) error {
	for _, name := range []string{
		KubeApiServerSecretNameAdmissionKubeconfig,
		KubeApiServerSecretNameAuditWebhookConfig,
		KubeApiServerSecretNameBasicAuth,
	} {
		secret := o.emptySecret(name)
		if err := o.client.Delete(ctx, secret); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (o *operation) deployKubeApiServerSecretAdmissionKubeconfig(ctx context.Context) error {
	controlplane := o.imports.VirtualGarden.KubeAPIServer.GardenerControlplane
	if !controlplane.ValidatingWebhookEnabled && !controlplane.MutatingWebhookEnabled {
		return nil
	}

	secret := o.emptySecret(KubeApiServerSecretNameAdmissionKubeconfig)

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

func (o *operation) deployKubeApiServerSecretAuditWebhookConfig(ctx context.Context) error {
	config := o.imports.VirtualGarden.KubeAPIServer.AuditWebhookConfig.Config
	if len(config) == 0 {
		return nil
	}

	secret := o.emptySecret(KubeApiServerSecretNameAuditWebhookConfig)

	_, err := controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data["audit-webhook-config.yaml"] = []byte(config)
		return nil
	})

	return err
}

func (o *operation) deployKubeApiServerSecretBasicAuth(ctx context.Context) error {
	const basicAuthKey = "basic_auth.csv"

	var basicAuthValue []byte

	secret := o.emptySecret(KubeApiServerSecretNameBasicAuth)
	err := o.client.Get(ctx, util.GetKey(secret), secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}

		// secret does not exist: generate password
		pw, err2 := utils.GenerateRandomString(32)
		if err2 != nil {
			return err2
		}

		basicAuthValue = []byte(fmt.Sprintf("%s,admin,admin,system:masters", pw))
	} else {
		// secret exists: use existing value
		basicAuthValue = secret.Data[basicAuthKey]
	}

	_, err = controllerutil.CreateOrUpdate(ctx, o.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[basicAuthKey] = basicAuthValue
		return nil
	})

	return err
}

func (o *operation) emptySecret(name string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: o.namespace}}
}
