all: docker

COV_THRESHOLD=80

VERSION := $(shell git describe --tags --always --dirty)
IMAGE_NAME="example-controller"

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
	# TODO(danielscottt): Once there are tests on these packages, enable the
	# coverage checking.
	# ../../scripts/test-with-cov.sh . $(COV_THRESHOLD)
	go test .
	# ../../scripts/test-with-cov.sh ./apis/cr/v1 $(COV_THRESHOLD)
	go test ./apis/cr/v1
