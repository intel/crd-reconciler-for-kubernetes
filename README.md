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

## Dependency management

This project uses [`dep`](https://github.com/golang/dep).

Cheatsheet:
- `dep ensure` restores source dependencies
- `dep ensure --add github.com/<foo>/<bar>` adds a new source dependency
