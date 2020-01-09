#!/bin/bash -e

dir_with_test_binaries=${1:-precompilied_test_binaries}

test_binaries=$(find "$dir_with_test_binaries" -type f -name '*.test' )
for test_binary in $test_binaries; do
  filename=$(echo "$test_binary" | tr / _)_coverage.out;
  $test_binary -test.coverprofile=$WERF_TEST_COVERAGE_DIR/$filename;
done
