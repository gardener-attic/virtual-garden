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
	"context"
	"fmt"
	"os"

	"github.com/gardener/virtual-garden/pkg/api/helper"
	"github.com/gardener/virtual-garden/pkg/api/loader"
	"github.com/gardener/virtual-garden/pkg/api/validation"
	"github.com/gardener/virtual-garden/pkg/virtualgarden"

	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	kubernetesscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OperationType is a string alias.
type OperationType string

const (
	// OperationTypeReconcile is a  constant for the RECONCILE operation type.
	OperationTypeReconcile OperationType = "RECONCILE"
	// OperationTypeDelete is a constant for the DELETE operation type.
	OperationTypeDelete OperationType = "DELETE"
)

// NewCommandVirtualGarden creates a *cobra.Command object with default parameters.
func NewCommandVirtualGarden() *cobra.Command {
	opts := NewOptions()

	cmd := &cobra.Command{
		Use:   "virtual-garden",
		Short: "Launch the virtual garden deployer",
		Long:  `The virtual garden deployer deploys a virtual garden cluster into a hosting cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			opts.InitializeFromEnvironment()
			utilruntime.Must(opts.validate(args))

			log := &logrus.Logger{
				Out:   os.Stderr,
				Level: logrus.InfoLevel,
				Formatter: &logrus.TextFormatter{
					DisableColors: true,
				},
			}

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Infof("FLAG: --%s=%s", flag.Name, flag.Value)
			})

			if err := run(cmd.Context(), log, opts); err != nil {
				panic(err)
			}

			log.Infof("Execution finished successfully.")
		},
	}

	verflag.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())
	return cmd
}

// run runs the virtual garden deployer.
func run(ctx context.Context, log *logrus.Logger, opts *Options) error {
	log.Infof("Reading imports file from IMPORTS_PATH(%s)", opts.ImportsPath)
	imports, err := loader.FromFile(opts.ImportsPath)
	if err != nil {
		return err
	}

	imports.HostingCluster.Kubeconfig = imports.Cluster

	etcdImage, err := helper.GetImageFromCompDescr(ctx, "etcd")
	if err != nil{
		return err
	}

	etcdBackupRestoreImage, err := helper.GetImageFromCompDescr(ctx, "etcdBackupRestore")
	if err != nil{
		return err
	}

	log.Infof("Validating imports file")
	if errList := validation.ValidateImports(imports); len(errList) > 0 {
		return errList.ToAggregate()
	}

	log.Infof("Creating REST config and Kubernetes client based on given kubeconfig")
	client, err := NewClientFromKubeconfig([]byte(imports.HostingCluster.Kubeconfig))
	if err != nil {
		return err
	}

	operation, err := virtualgarden.NewOperation(client, log, imports.HostingCluster.Namespace, opts.HandleNamespace,
		opts.HandleETCDPersistentVolumes, imports, etcdImage, etcdBackupRestoreImage)
	if err != nil {
		return err
	}
	log.Infof("Initialization %q operation complete.", opts.OperationType)

	if opts.OperationType == OperationTypeReconcile {
		exports, err := operation.Reconcile(ctx)
		if err != nil {
			return err
		}

		log.Infof("Writing exports file to EXPORTS_PATH(%s)", opts.ExportsPath)
		err = loader.ToFile(exports, opts.ExportsPath)
		if err != nil {
			return err
		}

		return nil
	} else if opts.OperationType == OperationTypeDelete {
		return operation.Delete(ctx)
	}
	return fmt.Errorf("unknown operation type: %q", opts.OperationType)
}

// NewClientFromKubeconfig creates a new Kubernetes client for the given kubeconfig.
func NewClientFromKubeconfig(kubeconfig []byte) (client.Client, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(kubernetesscheme.AddToScheme(scheme))
	utilruntime.Must(hvpav1alpha1.AddToScheme(scheme))
	utilruntime.Must(autoscalingv1beta2.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1beta1.AddToScheme(scheme))

	return client.New(restConfig, client.Options{Scheme: scheme})
}
