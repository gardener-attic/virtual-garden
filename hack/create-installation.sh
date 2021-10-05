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
TMP_DIR="$(mktemp -d)"
INSTALLATION_PATH="${TMP_DIR}/installation.yaml"

cat << EOF > ${INSTALLATION_PATH}
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
metadata:
  name: virtual-garden
spec:
  componentDescriptor:
    ref:
      repositoryContext:
        type: ociRegistry
        baseUrl: eu.gcr.io/gardener-project/development
      componentName: github.com/gardener/virtual-garden
      version: ${EFFECTIVE_VERSION}

  blueprint:
    ref:
      resourceName: blueprint

  imports:
    targets:
    - name: cluster
      target: "#cluster"

  importDataMappings:
    hostingCluster:
      namespace: garden
      infrastructureProvider: gcp

    virtualGarden:
      deleteNamespace: true
      etcd:
        storageClassName: my-virtual-garden-storage-class
        handleETCDPersistentVolumes: true
      kubeAPIServer:
        replicas: 1
        dnsAccessDomain: ""
        gardenerControlplane:
          validatingWebhookEnabled: true
          mutatingWebhookEnabled: true

  exports:
    data:
    - name: kubeApiserverCaPem
      dataRef: "kubeapiservercapem"
    - name: etcdCaPem
      dataRef: "etcdcapem"
    - name: etcdClientTlsPem
      dataRef: "etcdclienttlspem"
    - name: etcdClientTlsKeyPem
      dataRef: "etcdclienttlskeypem"
    - name: virtualGardenEndpoint
      dataRef: "virtualgardenendpoint"

    targets:
    - name: virtualGardenKubeconfig
      target: "virtualgardenkubeconfig"
EOF

echo "Installation stored at ${INSTALLATION_PATH}"
