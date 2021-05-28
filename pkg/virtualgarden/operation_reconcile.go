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

	"github.com/gardener/virtual-garden/pkg/api/helper"

	"github.com/gardener/virtual-garden/pkg/api"

	"github.com/gardener/gardener/pkg/utils/flow"
)

// Reconcile runs the reconcile operation.
func (o *operation) Reconcile(ctx context.Context) (*api.Exports, error) {
	var (
		graph = flow.NewGraph("Virtual Garden Reconciliation")

		createNamespace = graph.Add(flow.Task{
			Name: "Creating namespace for virtual-garden deployment in hosting cluster",
			Fn:   o.CreateNamespace,
		})

		createKubeAPIServerService = graph.Add(flow.Task{
			Name:         "Deploying the service for exposing the virtual garden kube-apiserver",
			Fn:           o.DeployKubeAPIServerService,
			Dependencies: flow.NewTaskIDs(createNamespace),
		})

		deployBackupBucket = graph.Add(flow.Task{
			Name:         "Deploying the backup bucket for the main etcd",
			Fn:           flow.TaskFn(o.DeployBackupBucket).DoIf(helper.ETCDBackupEnabled(o.imports.VirtualGarden.ETCD)),
			Dependencies: flow.NewTaskIDs(createKubeAPIServerService),
		})

		createETCD = graph.Add(flow.Task{
			Name:         "Deploying the main and events etcds",
			Fn:           o.DeployETCD,
			Dependencies: flow.NewTaskIDs(deployBackupBucket),
		})

		_ = graph.Add(flow.Task{
			Name:         "Deploying kube-apiserver",
			Fn:           o.DeployKubeAPIServer,
			Dependencies: flow.NewTaskIDs(createETCD),
		})
	)

	err := graph.Compile().Run(flow.Opts{
		Context:          ctx,
		Logger:           o.log,
		ProgressReporter: flow.NewImmediateProgressReporter(o.progressReporter),
	})
	if err != nil {
		return nil, err
	}

	return &o.exports, nil
}
