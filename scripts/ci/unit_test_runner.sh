#!/bin/bash -e

test_binaries=$(find . -type f -name '*.test')
for test_binary in $test_binaries; do
  coverage_file_name="$(date +%s.%N | sha256sum | cut -c 1-10)-$(date +%s)_coverage.out"
  $test_binary -test.coverprofile="$WERF_TEST_COVERAGE_DIR"/"$coverage_file_name";
done
