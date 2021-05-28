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

	"github.com/gardener/gardener/pkg/utils/flow"
)

// Delete runs the delete operation.
func (o *operation) Delete(ctx context.Context) error {
	var (
		graph = flow.NewGraph("Virtual Garden Deletion")

		deleteKubeAPIServer = graph.Add(flow.Task{
			Name: "Deleting kube-apiserver",
			Fn:   o.DeleteKubeAPIServer,
		})

		deleteETCD = graph.Add(flow.Task{
			Name:         "Deleting the main and events etcds",
			Fn:           o.DeleteETCD,
			Dependencies: flow.NewTaskIDs(deleteKubeAPIServer),
		})

		deleteBackupBucket = graph.Add(flow.Task{
			Name:         "Deleting the backup bucket for the main etcd",
			Fn:           flow.TaskFn(o.DeleteBackupBucket).DoIf(helper.ETCDBackupEnabled(o.imports.VirtualGarden.ETCD)),
			Dependencies: flow.NewTaskIDs(deleteETCD),
		})

		deleteKubeAPIServerService = graph.Add(flow.Task{
			Name:         "Deleting the service for exposing the virtual garden kube-apiserver",
			Fn:           o.DeleteKubeAPIServerService,
			Dependencies: flow.NewTaskIDs(deleteBackupBucket),
		})

		_ = graph.Add(flow.Task{
			Name:         "Deleting namespace for virtual-garden deployment in hosting cluster",
			Fn:           flow.TaskFn(o.DeleteNamespace).SkipIf(!o.imports.VirtualGarden.DeleteNamespace),
			Dependencies: flow.NewTaskIDs(deleteKubeAPIServerService),
		})
	)

	return graph.Compile().Run(flow.Opts{
		Context:          ctx,
		Logger:           o.log,
		ProgressReporter: flow.NewImmediateProgressReporter(o.progressReporter),
	})
}
