// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localenvtest

import (
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1beta1"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// Environment is a internal landscaper test environment
type Environment struct {
	Env    *envtest.Environment
	Client client.Client
	Logger *logrus.Logger
}

// New creates a new test environment with the landscaper known crds.
func New(projectRoot string) (*Environment, error) {
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}
	testBinPath := filepath.Join(projectRoot, "tmp", "test", "bin")
	// if the default Landscaper test bin does not exist we default to the kubebuilder testenv default
	// that uses the KUBEBUILDER_ASSETS env var.
	if _, err := os.Stat(testBinPath); err == nil {
		if err := os.Setenv("TEST_ASSET_KUBE_APISERVER", filepath.Join(testBinPath, "kube-apiserver")); err != nil {
			return nil, err
		}
		if err := os.Setenv("TEST_ASSET_ETCD", filepath.Join(testBinPath, "etcd")); err != nil {
			return nil, err
		}
		if err := os.Setenv("TEST_ASSET_KUBECTL", filepath.Join(testBinPath, "kubectl")); err != nil {
			return nil, err
		}
	}

	log := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{
			DisableColors: true,
		},
	}

	return &Environment{
		Env:    &envtest.Environment{},
		Logger: log,
	}, nil
}

// Start starts the fake environment and creates a client for the started kubernetes cluster.
func (e *Environment) Start() (client.Client, error) {
	restConfig, err := e.Env.Start()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(apiextensions.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(policyv1.AddToScheme(scheme))

	fakeClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	e.Client = fakeClient
	return fakeClient, nil
}

// Stop stops the running dev environment
func (e *Environment) Stop() error {
	return e.Env.Stop()
}
