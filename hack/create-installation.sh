#!/bin/bash
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

SOURCE_PATH="$(dirname $0)/.."
INSTALLATION_PATH=${SOURCE_PATH}"/tmp/installation.yaml"

> ${INSTALLATION_PATH}

cat << EOF >> ${INSTALLATION_PATH}
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

