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
	"strings"

	"github.com/gardener/gardener/pkg/utils/flow"
	"github.com/gardener/virtual-garden/pkg/api"
	"github.com/gardener/virtual-garden/pkg/api/helper"
	"github.com/gardener/virtual-garden/pkg/provider"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface is an interface for the operation.
type Interface interface {
	// Reconcile performs a reconcile operation.
	Reconcile(context.Context) (*api.Exports, error)
	// Delete performs a delete operation.
	Delete(context.Context) error
}

// Prefix is the prefix for resource names related to the virtual garden.
const Prefix = "virtual-garden"

// operation contains the configuration for a operation.
type operation struct {
	// client is the Kubernetes client for the hosting cluster.
	client client.Client

	// log is a logger.
	log logrus.FieldLogger

	// infrastructureProvider is a specific implementation for infrastructure providers.
	infrastructureProvider provider.InfrastructureProvider

	// backupProvider is a specific implementation for backup providers.
	backupProvider provider.BackupProvider

	// namespace is the namespace in the hosting cluster into which the virtual garden shall be deployed.
	namespace string

	// imports contains the imports configuration.
	imports *api.Imports

	exports api.Exports

	// imageRefs contains the image references from the component descriptor that are needed for the Deployments and
	// StatefulSet.
	imageRefs api.ImageRefs
}

// NewOperation returns a new operation structure that implements Interface.
func NewOperation(
	c client.Client,
	log *logrus.Logger,
	namespace string,
	imports *api.Imports,
	imageRefs *api.ImageRefs,
) (Interface, error) {
	op := &operation{
		client: c,
		log:    log,

		namespace: namespace,
		imports:   imports,
		imageRefs: *imageRefs,
	}

	infrastructureProvider, err := provider.NewInfrastructureProvider(imports.HostingCluster.InfrastructureProvider)
	if err != nil {
		return nil, err
	}
	op.infrastructureProvider = infrastructureProvider

	if helper.ETCDBackupEnabled(imports.VirtualGarden.ETCD) {
		backupProvider, err := provider.NewBackupProvider(imports.VirtualGarden.ETCD.Backup.InfrastructureProvider,
			imports.VirtualGarden.ETCD.Backup.Credentials, imports.VirtualGarden.ETCD.Backup.BucketName,
			imports.VirtualGarden.ETCD.Backup.Region, log)
		if err != nil {
			return nil, err
		}
		op.backupProvider = backupProvider
	}

	return op, nil
}

func (o *operation) progressReporter(_ context.Context, stats *flow.Stats) {
	if stats.ProgressPercent() == 0 && stats.Running.Len() == 0 {
		return
	}

	executionNow := ""
	if stats.Running.Len() > 0 {
		executionNow = fmt.Sprintf(" - Executing now: %q", strings.Join(stats.Running.StringList(), ", "))
	}
	o.log.Infof("%d%% of all tasks completed (%d/%d)%s", stats.ProgressPercent(), stats.Failed.Len()+stats.Succeeded.Len(), stats.All.Len(), executionNow)
}
