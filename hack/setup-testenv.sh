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

K8S_VERSION="1.21.x"

echo "> Setup Test Environment for K8s Version ${K8S_VERSION}"

CURRENT_DIR=$(dirname $0)
PROJECT_ROOT="${CURRENT_DIR}"/..

ARCH_ARG=""
if [[ $(go env GOOS) == "darwin" && $(go env GOARCH) == "arm64" ]]; then
  ARCH_ARG="--arch amd64"
fi

export KUBEBUILDER_ASSETS=$(setup-envtest use -p path ${K8S_VERSION} ${ARCH_ARG})

mkdir -p ${PROJECT_ROOT}/tmp/test
rm -f ${PROJECT_ROOT}/tmp/test/bin
ln -s "${KUBEBUILDER_ASSETS}" ${PROJECT_ROOT}/tmp/test/bin
