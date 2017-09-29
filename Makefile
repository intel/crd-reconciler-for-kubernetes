.PHONY: docker test

COV_THRESHOLD=80

all: controllers

version=v0.1.0

test: lint validate_schemas
	# TODO(danielscottt): Once there are tests on these packages, enable the
	# coverage checking.
	# ./scripts/test-with-cov.sh ./pkg/crd $(COV_THRESHOLD)
	go test ./pkg/crd
	# ./scripts/test-with-cov.sh ./pkg/controller $(COV_THRESHOLD)
	go test ./pkg/controller
	# ./scripts/test-with-cov.sh ./pkg/util $(COV_THRESHOLD)
	go test ./pkg/util

dep:
	docker build -t kube-controllers-go-dep:$(version) -f Dockerfile.dep .

docker:
	docker build -t kube-controllers-go:$(version) .

controllers: stream-prediction example

stream-prediction:
	(cd cmd/stream-prediction-controller && make)

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
	docker-compose exec --privileged test /bin/bash

test-e2e: env-up
	docker-compose exec test ./resources/wait-port kubernetes 8080
	# Run the stream-prediction controller tests in a new container with
	# the same configuration as the service, inside the docker-compose
	# environment.
	docker-compose run stream-prediction-controller make test-e2e

install-linter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	gometalinter --config=lint.json ./pkg/...
	gometalinter --config=lint.json ./test/...

validate_schemas:
	(cd api/crd && make)
