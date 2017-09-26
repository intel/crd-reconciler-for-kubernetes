all: docker

version=v0.1.0

docker:
	(cd ../../ && make docker)
	docker build --no-cache -t stream-prediction-controller:$(version) .

lint:
	gometalinter --config=lint.json .
	# Disabling golint for apis since it conflicts with the deepcopy-gen annotations.
	gometalinter --config=lint.json --disable=golint ./apis/...
	gometalinter --config=lint.json ./test/...

test: lint
	go test -v .
	go test -v ./apis/...

test-e2e:
	go test -v ./test/e2e/...
