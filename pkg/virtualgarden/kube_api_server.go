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
	"time"

	"github.com/gardener/virtual-garden/pkg/util"
)

// DeployKubeAPIServer deploys a kubernetes api server.
func (o *operation) DeployKubeAPIServer(ctx context.Context) error {
	o.log.Infof("Deploying the HVPA CRD")
	if err := o.deployHVPACRD(ctx); err != nil {
		return err
	}

	loadbalancer, err := o.computeKubeApiserverLoadbalancer(ctx)
	if err != nil {
		return err
	}

	aggregatorCACertificate, aggregatorCACertChecksum, err := o.deployKubeApiServerAggregatorCACertificate(ctx)
	if err != nil {
		return err
	}

	_, aggregatorClientCertChecksum, err := o.deployKubeApiServerAggregatorClientCertificate(ctx, aggregatorCACertificate)
	if err != nil {
		return err
	}

	apiServerCACertificate, apiServerCACertChecksum, err := o.deployKubeApiServerApiServerCACertificate(ctx)
	if err != nil {
		return err
	}

	_, apiServerServerCertChecksum, err := o.deployKubeApiServerApiServerServerCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return err
	}

	_, kubeControllerManagerClientCertChecksum, err := o.deployKubeApiServerKubeControllerManagerClientCertificate(ctx, apiServerCACertificate)
	if err != nil {
		return err
	}

	_, clientAdminCertChecksum, err := o.deployKubeApiServerClientAdminCertificate(ctx, apiServerCACertificate, loadbalancer)
	if err != nil {
		return err
	}

	// temporarily
	fmt.Println("Checksums: ",
		aggregatorCACertChecksum,
		aggregatorClientCertChecksum,
		apiServerCACertChecksum,
		apiServerServerCertChecksum,
		kubeControllerManagerClientCertChecksum,
		clientAdminCertChecksum,
	)

	return nil
}

// DeleteKubeAPIServer deletes the kube-apiserver and all related resources.
func (o *operation) DeleteKubeAPIServer(ctx context.Context) error {
	o.log.Infof("Deleting the HPVA CRD")
	if err := o.deleteHPVACRD(ctx); err != nil {
		return err
	}

	return nil
}

func (o *operation) computeKubeApiserverLoadbalancer(ctx context.Context) (string, error) {
	var err error
	var loadbalancer string

	util.Repeat(func() bool {
		loadbalancer, err := o.computeKubeApiserverLoadbalancerOnce(ctx)
		return (err != nil || loadbalancer != "")
	}, 10, time.Second)

	return loadbalancer, err
}

func (o *operation) computeKubeApiserverLoadbalancerOnce(ctx context.Context) (string, error) {
	service := emptyKubeAPIServerService(o.namespace)

	err := o.client.Get(ctx, util.GetKey(service), service)
	if err != nil {
		return "", err
	}

	return o.infrastructureProvider.GetLoadBalancer(service), nil
}
