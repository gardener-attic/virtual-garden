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

package api

import (
	"encoding/json"
	"fmt"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

type ImageRefs struct {
	ETCDImage                  string
	ETCDBackupRestoreImage     string
	KubeControllerManagerImage string
	KubeAPIServerImage         string
}

func NewImageRefsFromComponentDescriptor(cd *cdv2.ComponentDescriptor) (*ImageRefs, error) {
	const (
		resourceNameETCD                  = "etcd"
		resourceNameETCDBackupRestore     = "etcd-backup-restore"
		resourceNameKubeAPIServer         = "kube-apiserver"
		resourceNameKubeControllerManager = "kube-controller-manager"
	)

	imageRefs := ImageRefs{}
	var err error

	imageRefs.ETCDImage, err = getImageRef(resourceNameETCD, cd)
	if err != nil {
		return nil, err
	}

	imageRefs.ETCDBackupRestoreImage, err = getImageRef(resourceNameETCDBackupRestore, cd)
	if err != nil {
		return nil, err
	}

	imageRefs.KubeAPIServerImage, err = getImageRef(resourceNameKubeAPIServer, cd)
	if err != nil {
		return nil, err
	}

	imageRefs.KubeControllerManagerImage, err = getImageRef(resourceNameKubeControllerManager, cd)
	if err != nil {
		return nil, err
	}

	return &imageRefs, nil
}

func getImageRef(resourceName string, cd *cdv2.ComponentDescriptor) (string, error) {
	for i := range cd.Resources {
		resource := &cd.Resources[i]

		if resource.Name == resourceName {
			access := cdv2.OCIRegistryAccess{}
			if err := json.Unmarshal(resource.Access.Raw, &access); err != nil {
				return "", err
			}

			return access.ImageReference, nil
		}
	}

	return "", fmt.Errorf("No resource with name %s found in component descriptor", resourceName)
}
