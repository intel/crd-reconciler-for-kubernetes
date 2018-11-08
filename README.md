# crd-reconciler-for-kubernetes

- Controllers to interpret CRDs to Kubernetes-native constructs and
  report health of those sub-resources.

- Library and docs for writing controllers that reconcile against CRDs.

## Build

The build depends on:

* `make`
* [`docker`](https://docs.docker.com/engine/installation)
* [`docker-compose`](https://docs.docker.com/compose/install)
  (end-to-end tests only)
* [`gcloud`](https://cloud.google.com/sdk/docs/#linux)
  (we use the Google Cloud SDK to push images to GCR)

### Quick-start

```shell
# Build a docker image containing only source dependencies:
$ make dep

# Run library tests, controller tests and build the controller images:
$ make
```

### Intermediate container images

All builds and tests happen inside of a container. Each controller program
(packages in `./cmd/...`) builds its own result container that can be
deployed in the local integration environment or a target cluster.

There are two intermediate container images:

- `crd-reconciler-for-kubernetes-dep` -- contains only source dependencies, separate
  to speed up local dev iterations

- `crd-reconciler-for-kubernetes` -- base image for all controller images, built on
  top of crd-reconciler-for-kubernetes-dep.

### Most useful Make targets

- **`make dep`**: Build a docker image containing only source dependencies
  and tag it as `crd-reconciler-for-kubernetes-dep:$(version)`. This step is a
  prerequisite to run the other targets, and must be run explicitly.

- **`make docker`**: Build a docker image containing the source repo, running
  the `./pkg/...` tests in the process.

- **`make controllers`**: Build all controller images (also runs tests.)

- **`make <controller-name>`**: Build the <controller-name> controller image
  (also runs tests.)

- **`make env-up`** and **`make env-down`**: Bring up/down the integration
  environment using `docker-compose`. List service status using e.g.
  `docker-compose ps`.

- **`make dev TARGET=[test|example-controller...]`**: Drop into bash inside the source repo container in the
  integration environment. Run after `make env-up`. TARGET defaults to 'test' and can also be set in your SHELL profile

- **`make test-e2e`**: Run the end-to-end integration tests.

- **`make debug TARGET=[test|example-controller...]`**: Attach to the process running in the TARGET container. See [tutorial](docs/debug.md). TARGET defaults to 'test' and can also be set in your SHELL profile

## Dependency management

This project uses [`dep`](https://github.com/golang/dep).

Cheatsheet:
- `dep ensure` restores source dependencies
- `dep ensure --add github.com/<foo>/<bar>` adds a new source dependency

After running `dep ensure --add` or manually editing `Gopkg.toml`,
you must manually re-run `make dep` to update your local
`crd-reconciler-for-kubernetes-dep` image. Otherwise, the other make targets will
be based off of an outdated set of source dependencies.

