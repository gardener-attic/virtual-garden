# SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

#### BUILDER ####
FROM golang:1.17.11 AS builder

WORKDIR /go/src/github.com/gardener/virtual-garden
COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

#### BASE ####
FROM gcr.io/distroless/static-debian11:nonroot AS base

#### Landscaper Controller ####
FROM base as virtual-garden

COPY --from=builder /go/bin/virtual-garden /virtual-garden

WORKDIR /

ENTRYPOINT ["/virtual-garden"]