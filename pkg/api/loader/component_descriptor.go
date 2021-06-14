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

package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

func ReadComponentDescriptor(componentDescriptorPath string) (*cdv2.ComponentDescriptor, error) {
	data, err := ioutil.ReadFile(componentDescriptorPath)
	if err != nil {
		return nil, err
	}

	cdList := &cdv2.ComponentDescriptorList{}
	if err := yaml.Unmarshal(data, cdList); err != nil {
		return nil, err
	}

	if len(cdList.Components) != 1 {
		return nil, fmt.Errorf("Component descriptor list does not contain a unique entry")
	}

	return &cdList.Components[0], nil
}

func ComponentDescriptorToFile(componentDescriptor *cdv2.ComponentDescriptorList, path string) error {
	b, err := yaml.Marshal(componentDescriptor)
	if err != nil {
		return err
	}

	folderPath := filepath.Dir(path)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, 0700); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path, b, os.ModePerm)
}

func CreateComponentDescriptorFromResourcesFile(resourcesFilePath, componentDescriptorPath string) error {
	data, err := ioutil.ReadFile(resourcesFilePath)
	if err != nil {
		return err
	}

	var resources interface{}
	if err := yaml.Unmarshal(data, &resources); err != nil {
		return err
	}

	fmt.Println("Resources read")

	return nil
}
