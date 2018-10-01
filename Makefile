.PHONY: docker test

VERSION := $(shell git describe --tags --always --dirty)

GOOGLE_PROJECT_ID=
GOOGLE_AUTH=
IMAGE_NAME=crd-reconciler-for-kubernetes
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
	/go/bin/deepcopy-gen --output-base=/go/src --input-dirs=github.com/intel/crd-reconciler-for-kubernetes/pkg/crd/fake/... --output-package=pkg/crd/fake
	/go/bin/deepcopy-gen --output-base=/go/src --input-dirs=github.com/intel/crd-reconciler-for-kubernetes/pkg/resource/fake/... --output-package=pkg/resource/fake

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
	docker-compose exec --privileged $(DEBUG_TARGET) env GODEBUGGER=$(GODEBUGGER) /go/src/github.com/intel/crd-reconciler-for-kubernetes/scripts/godebug attach $(DEBUG_TARGET)

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
