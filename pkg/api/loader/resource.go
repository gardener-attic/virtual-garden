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
	"io"
	"os"

	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

// ResourcesFromFile reads resources from a file.
func ResourcesFromFile(resourcesFilePath string) ([]cdresources.ResourceOptions, error) {
	file, err := os.Open(resourcesFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resources, err := readResources(file)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func readResources(reader *os.File) ([]cdresources.ResourceOptions, error) {
	resources := make([]cdresources.ResourceOptions, 0)
	yamldecoder := yamlutil.NewYAMLOrJSONDecoder(reader, 1024)
	for {
		resource := cdresources.ResourceOptions{}
		if err := yamldecoder.Decode(&resource); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to decode resource: %w", err)
		}

		if resource.Input != nil && resource.Access != nil {
			return nil, fmt.Errorf("the resources %q input and access is defind. Only one option is allowed", resource.Name)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}
