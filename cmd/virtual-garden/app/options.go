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

package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// Options has all the context and parameters needed to run the virtual garden deployer.
type Options struct {
	// OperationType is the operation to be executed.
	OperationType OperationType
	// ImportsPath is the path to the imports file.
	ImportsPath string
	// ExportsPath is the path to the exports file. The parent directory exists; the export file itself must be created.
	// The format of the exports file must be json or yaml.
	ExportsPath string
	// ComponentDescriptorPath is the path to the component descriptor file.
	ComponentDescriptorPath string

	// HandleETCDPersistentVolumes defines whether the PV(C)s that are getting automatically created by the etcd
	// statefulset shall be handled or not (false by default). If true then they will be deleted when the virtual
	// garden is deleted. Otherwise, they will remain in the system for manual cleanup (to prevent data loss).
	HandleETCDPersistentVolumes bool
}

// NewOptions returns a new options structure.
func NewOptions() *Options {
	return &Options{OperationType: OperationTypeReconcile}
}

// AddFlags adds flags for a specific Scheduler to the specified FlagSet.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.HandleETCDPersistentVolumes, "handle-etcd-persistent-volumes", false, "delete PV(C)s for the etcds automatically when virtual garden is deleted")
}

// InitializeFromEnvironment initializes the options from the found environment variables.
func (o *Options) InitializeFromEnvironment() {
	if op := os.Getenv("OPERATION"); len(op) > 0 {
		o.OperationType = OperationType(op)
	}
	o.ImportsPath = os.Getenv("IMPORTS_PATH")
	o.ExportsPath = os.Getenv("EXPORTS_PATH")
	o.ComponentDescriptorPath = os.Getenv("COMPONENT_DESCRIPTOR_PATH")
}

// validate validates all the required options.
func (o *Options) validate(args []string) error {
	if o.OperationType != OperationTypeReconcile && o.OperationType != OperationTypeDelete {
		return fmt.Errorf("operation must be %q or %q", OperationTypeReconcile, OperationTypeDelete)
	}

	if len(o.ImportsPath) == 0 {
		return fmt.Errorf("missing path for imports file")
	}

	if len(o.ExportsPath) == 0 {
		return fmt.Errorf("missing path for exports file")
	}

	if len(o.ComponentDescriptorPath) == 0 {
		return fmt.Errorf("missing path for component descriptor file")
	}

	if len(args) != 0 {
		return errors.New("arguments are not supported")
	}

	return nil
}
