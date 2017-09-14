# kube-controllers-go

- Custom resource definitions (CRDs) for Nervana Cloud.

- Controllers to interpret CRDs to Kubernetes-native constructs and
  report health of those sub-resources.

- Library and docs for writing controllers that reconcile against CRDs.

## Build

- Requires `docker`

```
$ make controllers
```

## Test

### End-to-end tests

- Requires `docker-compose`

```
$ make test-e2e
```
