Testing
=======

## Coverage

For a build to pass, any given package must meet a threshold of coverage.

To test a package, and fail in the case of insufficient coverage, a script is provided which wraps `go test`. `./scripts/test-with-cov.sh` takes two arguments: the package path to test, and the threshold.  If a threshold is not given, the script will default to 80%.

```shell
$ ./scripts/test-with-cov.sh . 80
```

Is the same as

```shell
$ ./scripts/test-with-cov.sh .
```
