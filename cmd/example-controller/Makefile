all: docker

version=v0.1.0

docker:
	(cd ../../ && make docker)
	docker build --no-cache -t example-controller:$(version) .

test:
	go test -v ./...
