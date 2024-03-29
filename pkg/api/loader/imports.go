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

	"sigs.k8s.io/yaml"

	"github.com/gardener/virtual-garden/pkg/api"
)

// ImportsFromFile will read the file from the given path and try to unmarshal it into an api.Imports structure.
func ImportsFromFile(path string) (*api.Imports, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	imports := &api.Imports{}
	if err := yaml.Unmarshal(data, imports); err != nil {
		return nil, err
	}

	return imports, nil
}
