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

package gcp

import (
	"encoding/json"
	"fmt"
)

// DataKeyServiceAccountJSON is a constant for a data key whose value is a GCP service account in JSON format.
const DataKeyServiceAccountJSON = "serviceaccount.json"

// ReadServiceAccount reads the ServiceAccount from the given data map.
func ReadServiceAccount(data map[string]string) (string, error) {
	serviceAccount, ok := data[DataKeyServiceAccountJSON]
	if !ok {
		return "", fmt.Errorf("data map doesn't have a service account json")
	}
	return serviceAccount, nil
}

// ExtractServiceAccountProjectID extracts the project id from the given service account JSON.
func ExtractServiceAccountProjectID(serviceAccountJSON string) (string, error) {
	var serviceAccount struct {
		ProjectID string `json:"project_id"`
	}

	if err := json.Unmarshal([]byte(serviceAccountJSON), &serviceAccount); err != nil {
		return "", err
	}
	if serviceAccount.ProjectID == "" {
		return "", fmt.Errorf("no service account specified")
	}

	return serviceAccount.ProjectID, nil
}
