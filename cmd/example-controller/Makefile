all: docker

COV_THRESHOLD=80

version=v0.1.0

docker:
	(cd ../../ && make docker)
	docker build --no-cache -t example-controller:$(version) .

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
