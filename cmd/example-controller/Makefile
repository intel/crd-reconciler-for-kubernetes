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

all: docker

VERSION := $(shell git describe --tags --always --dirty)

GOOGLE_PROJECT_ID=
GOOGLE_AUTH=
IMAGE_NAME=example-controller

docker:
	(cd ../../ && make docker)
	docker build \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		.

lint:
	gometalinter --config=lint.json .
	gometalinter --config=lint.json ./apis/...

test: lint
	go test --cover .
	go test --cover ./apis/cr/v1

code-generation:
	/go/bin/deepcopy-gen --output-base=/go/src --input-dirs=github.com/NervanaSystems/kube-controllers-go/cmd/example-controller/apis/cr/v1/...

push-image-preflight:
ifeq ($(GOOGLE_PROJECT_ID),)
	$(error GOOGLE_PROJECT_ID must be set)
endif
ifeq ($(GOOGLE_AUTH),)
	$(error GOOGLE_AUTH must be set)
endif
	# Setting auth and project config for gcloud SDK.
	# Running this locally may override your CLI settings.
	# TODO(CD): Stash the active auth and project and restore them
	# after pushing.
	@echo "$(GOOGLE_AUTH)" | base64 --decode > /tmp/gcp-key.json
	gcloud auth activate-service-account --key-file /tmp/gcp-key.json
	gcloud config set project "$(GOOGLE_PROJECT_ID)"

push-image: push-image-preflight docker
	@ echo "tagging container"
	docker tag $(IMAGE_NAME):$(VERSION) gcr.io/$(GOOGLE_PROJECT_ID)/$(IMAGE_NAME):$(VERSION)
	@ echo "pushing container to gcr.io"
	gcloud docker -- push gcr.io/$(GOOGLE_PROJECT_ID)/$(IMAGE_NAME):$(VERSION)
