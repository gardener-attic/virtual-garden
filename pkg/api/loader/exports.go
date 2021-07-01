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
	"io/ioutil"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	"github.com/gardener/virtual-garden/pkg/api"
)

// ExportsToFile writes export data to a file.
func ExportsToFile(exports *api.Exports, path string) error {
	b, err := yaml.Marshal(exports)
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

// ExportsFromFile reads export data from a file.
func ExportsFromFile(path string) (*api.Exports, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	exports := &api.Exports{}
	if err := yaml.Unmarshal(data, exports); err != nil {
		return nil, err
	}

	return exports, nil
}
