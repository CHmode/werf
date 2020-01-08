#!/bin/bash -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
project_dir=$script_dir/../..

find_dir=${1:-.}
test_binaries_output_dirname=${2:-$project_dir/precompiled_test_binaries}
go_test_extra_tags="${*:3}"


if [[ "$OSTYPE" == "darwin"* ]]; then
  brew install findutils
  package_paths=$(gfind "$find_dir" -type f -name '*_test.go' -printf '%h\n' | sort -u)
else
  package_paths=$(find "$find_dir" -type f -name '*_test.go' -printf '%h\n' | sort -u)
fi

for package_path in $package_paths; do
	test_binary_filename=$(echo "$package_path" | tr / _).test
	test_binary_filename=${test_binary_filename:2}
	go test --tags "dfrunmount dfssh $go_test_extra_tags" "$package_path" -coverpkg=./... -c -o "$test_binaries_output_dirname"/"$test_binary_filename"
done
