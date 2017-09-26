all: docker

version=v0.1.0

docker:
	(cd ../../ && make docker)
	docker build --no-cache -t example-controller:$(version) .

lint:
	gometalinter --config=lint.json .
	gometalinter --config=lint.json ./apis/...

test: lint
	go test -v .
	go test -v ./apis/...
