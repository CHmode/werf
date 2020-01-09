#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..
project_bin_tests_dir=$project_dir/bin/tests

mkdir -p "$project_bin_tests_dir"
cd "$project_dir"

unameOut="$(uname -s)"
case "${unameOut}" in
    CYGWIN*|MINGW*|MSYS*) binary_name=werf_with_coverage.exe;;
    *)                    binary_name=werf_with_coverage
esac

go test -ldflags="-s -w" -tags "dfrunmount dfssh integration_coverage" -coverpkg=./... -c cmd/werf/main.go cmd/werf/main_test.go -o "$project_bin_tests_dir"/$binary_name
if [[ "$OSTYPE" == "linux-gnu" ]] || [[ "$OSTYPE" == "darwin"* ]]; then
  if ! [ -x "$(command -v upx)" ]; then
    if [[ "$OSTYPE" == "linux-gnu" ]]; then
      sudo apt-get install upx
    elif [[ "$OSTYPE" == "darwin"* ]]; then
      brew install upx
    fi
  fi

  upx "$project_bin_tests_dir"/$binary_name
fi

chmod +x "$project_bin_tests_dir"/$binary_name
