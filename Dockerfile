# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

#### BUILDER ####
FROM golang:1.16 AS builder

WORKDIR /go/src/github.com/gardener/virtual-garden
COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

#### BASE ####
FROM eu.gcr.io/gardenlinux/gardenlinux:184.0 AS base

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get --yes -o Dpkg::Options::="--force-confnew" install ca-certificates \
    && rm -rf /var/lib/apt /var/cache/apt

#### Landscaper Controller ####
FROM base as virtual-garden

COPY --from=builder /go/bin/virtual-garden /virtual-garden

WORKDIR /

ENTRYPOINT ["/virtual-garden"]