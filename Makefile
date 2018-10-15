#
# Copyright (c) 2018 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: EPL-2.0
#

.PHONY: docker test

VERSION := $(shell git describe --tags --always --dirty)

GOOGLE_PROJECT_ID=
GOOGLE_AUTH=
IMAGE_NAME=kube-controllers-go
TARGET ?= test
DEBUG_TARGET ?= example-controller
GODEBUGGER ?= gdb

all: controllers

test: lint
	go test -cover -v ./pkg/...

dep:
	docker build \
		-t $(IMAGE_NAME)-dep:$(VERSION) \
		-t $(IMAGE_NAME)-dep:latest \
		-f Dockerfile.dep .

docker:
	docker build \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest .

controllers: example

code-generation:
	/go/bin/deepcopy-gen --output-base=/go/src --input-dirs=github.com/NervanaSystems/kube-controllers-go/pkg/crd/fake/... --output-package=pkg/crd/fake
	/go/bin/deepcopy-gen --output-base=/go/src --input-dirs=github.com/NervanaSystems/kube-controllers-go/pkg/resource/fake/... --output-package=pkg/resource/fake

example:
	(cd cmd/example-controller && make)

env-up: env-down
	docker-compose up -d
	docker-compose ps

env-down:
	docker-compose down
	# resources is mounted as ~/.kube in the test container. This removes the
	# artifacts created during testing.
	rm -rf resources/cache

dev:
	docker-compose exec --privileged $(TARGET) /bin/bash

debug:
	docker-compose exec --privileged $(DEBUG_TARGET) env GODEBUGGER=$(GODEBUGGER) /go/src/github.com/NervanaSystems/kube-controllers-go/scripts/godebug attach $(DEBUG_TARGET)

install-linter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	gometalinter --config=lint.json --disable=golint ./pkg/...

push-images:
	@ (cd cmd/example-controller && \
		make push-image \
		  GOOGLE_AUTH=$(GOOGLE_AUTH) \
		  GOOGLE_PROJECT_ID=$(GOOGLE_PROJECT_ID))
