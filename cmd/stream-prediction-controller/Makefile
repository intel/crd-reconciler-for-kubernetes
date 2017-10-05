all: docker

COV_THRESHOLD=80

version=v0.1.0

docker:
	(cd ../../ && make docker)
	docker build --no-cache -t stream-prediction-controller:$(version) .

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
