all: docker

VERSION := $(shell git describe --tags --always --dirty)

GOOGLE_PROJECT_ID=
GOOGLE_AUTH=
IMAGE_NAME=stream-prediction-controller
COV_THRESHOLD=80

docker:
	(cd ../../ && make docker)
	docker build \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		.

lint:
	gometalinter --config=lint.json .
	# Disabling golint for apis since it conflicts with the deepcopy-gen
	# annotations.
	gometalinter --config=lint.json --disable=golint ./apis/...
	gometalinter --config=lint.json ./hooks/...
	gometalinter --config=lint.json ./test/...

test: lint
	# TODO(danielscottt): Once there are tests on these packages, enable the
	# coverage checking.
	# ../../scripts/test-with-cov.sh . $(COV_THRESHOLD)
	go test .
	# ../../scripts/test-with-cov.sh ./apis/cr/v1 $(COV_THRESHOLD)
	go test ./apis/cr/v1
	# ../../scripts/test-with-cov.sh ./apis/... $(COV_THRESHOLD)
	go test ./apis/...
	#  ../../scripts/test-with-cov.sh ./hooks/... $(COV_THRESHOLD)
	go test ./hooks/...

test-e2e:
	go test -v ./test/e2e/...

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
	echo "$(GOOGLE_AUTH)" | base64 --decode > /tmp/gcp-key.json
	gcloud auth activate-service-account --key-file /tmp/gcp-key.json
	gcloud config set project "$(GOOGLE_PROJECT_ID)"

push-image: push-image-preflight docker
	@ echo "tagging container"
	docker tag $(IMAGE_NAME):$(VERSION) gcr.io/$(GOOGLE_PROJECT_ID)/$(IMAGE_NAME):$(VERSION)
	@ echo "pushing container to gcr.io"
	gcloud docker -- push gcr.io/$(GOOGLE_PROJECT_ID)/$(IMAGE_NAME):$(VERSION)
