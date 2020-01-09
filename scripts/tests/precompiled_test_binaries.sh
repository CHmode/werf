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

if ! [ -x "$(command -v upx)" ]; then
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    sudo apt-get install upx
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    brew install upx
  fi
fi

i=0 # FIXME
for package_path in $package_paths; do
  test_binary_filename=$(basename -- "$package_path").test
	test_binary_path="$test_binaries_output_dirname"/"$package_path"/"$test_binary_filename"
	go test -ldflags="-s -w" --tags "dfrunmount dfssh $go_test_extra_tags" "$package_path" -coverpkg=./... -c -o "$test_binary_path"

  if [[ ! -f $test_binary_path ]]; then
     continue
  fi

	if [[ "$OSTYPE" == "linux-gnu" ]] || [[ "$OSTYPE" == "darwin"* ]]; then
	  upx "$test_binary_path"
  fi

  i=$((i+1))
  if [[ $i -eq 2 ]]; then
    break
  fi
done
