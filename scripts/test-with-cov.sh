#!/bin/bash

set -e

function __test-cov {
  local pkg=$1

  # confirm that a package was given. If not, exit with "invalid argument".
  if [ -z "$pkg" ]; then
    echo "error: no package given"
    exit 22
  fi

  # Retrieve the threshold. If one is not given, set it to 80%
  local thresh=${2:-80}

  # Run tests and store the output.
  local result=$(go test --cover $pkg)

  # Parse the output for the coverage.
  # the expected input here is:
  # ok <pkg-name> <run-duration>  coverage: <coverage-percentage> of statements
  local cov=$(echo $result | sed -rn 's/.*coverage: ([[:digit:]]*)\.[[:digit:]]% of statements/\1/p')

  # If we're unable to parse the coverage, say so, and exit, printing the
  # result.
  # NB: The most likely case of a parsing error is the absence of test files.
  if [ -z "$cov" ]; then
    echo "unable to parse coverage output for $pkg"
    echo $result
    exit 1
  fi

  # Fail if coverage is less than threshold.
  if [ "$cov" -lt "$thresh" ]; then
    echo "coverage for $pkg was below required threshold of $thresh"
    exit 1
  fi

  # Finally, print our result.
  echo $result
}

__test-cov $1 $2
