#!/bin/bash
#
# Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

SOURCE_PATH="$(dirname $0)/.."
REPO_CTX="eu.gcr.io/gardener-project/development"
CA_PATH="$(mktemp -d)"
BASE_DEFINITION_PATH="${CA_PATH}/component-descriptor.yaml"

if ! which component-cli 1>/dev/null; then
  echo -n "component-cli is required to generate the component descriptors"
  echo -n "Trying to installing it..."
  go get github.com/gardener/component-cli/cmd/component-cli

  if ! which component-cli 1>/dev/null; then
    echo -n "component-cli was successfully installed but the binary cannot be found"
    echo -n "Try adding the \$GOPATH/bin to your \$PATH..."
    exit 1
  fi
fi
if ! which jq 1>/dev/null; then
  echo -n "jq canot be found"
  exit 1
fi

echo "> Generate Component Descriptor ${EFFECTIVE_VERSION}"
echo "> Creating base definition"
component-cli ca create "${CA_PATH}" \
    --component-name=github.com/gardener/virtual-garden \
    --component-version=${EFFECTIVE_VERSION} \
    --repo-ctx=${REPO_CTX}

echo "> Extending resources.yaml: adding image of virtual-garden deployer"
RESOURCES_BASE_PATH="$(mktemp -d)"
cp -R ".landscaper/" "${RESOURCES_BASE_PATH}"

RESOURCES_FILE_PATH="${RESOURCES_BASE_PATH}/resources.yaml"
cat << EOF >> ${RESOURCES_FILE_PATH}
---
type: ociImage
name: virtual-garden-container-deployer
relation: local
access:
  type: ociRegistry
  imageReference: eu.gcr.io/gardener-project/development/images/virtual-garden:${EFFECTIVE_VERSION}
...
EOF

echo "> Creating ctf"
CTF_PATH=./gen/ctf.tar
mkdir -p ./gen
[ -e $CTF_PATH ] && rm ${CTF_PATH}
CTF_PATH=${CTF_PATH} BASE_DEFINITION_PATH=${BASE_DEFINITION_PATH} CURRENT_COMPONENT_REPOSITORY=${REPO_CTX} RESOURCES_FILE_PATH=${RESOURCES_FILE_PATH} bash $SOURCE_PATH/.ci/component_descriptor

component-cli ctf push --repo-ctx=${REPO_CTX} "${CTF_PATH}"
