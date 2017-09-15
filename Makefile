.PHONY: docker test

all: controllers

version=v0.1.0

test:
	go test -v ./pkg/...

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

dev:
	docker-compose exec --privileged test /bin/bash

test-e2e: env-up
	docker-compose exec test go test -v ./test/e2e/...
