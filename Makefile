.PHONY: docker test

all: controllers

version=v0.1.0

test: lint
	go test -v ./pkg/...

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

dev:
	docker-compose exec --privileged test /bin/bash

test-e2e: env-up
	docker-compose exec test ./resources/wait-port kubernetes 8080
	docker-compose exec stream-prediction-controller go test -v ./test/...
	docker-compose exec test go test -v ./test/e2e/...

install_linter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install

lint: install_linter
	gometalinter --config=lint.json ./pkg/...
	gometalinter --config=lint.json ./test/...
