#!/usr/bin/env bash

find_dir=${1:-.}
binaries_dir=${2:-precompiled_test_binaries}
go_test_extra_args="${@:3}"

packages=$(find $find_dir -type f -name '*_test.go' -printf '%h\n' | sort -u)
for package in $packages; do
	filename=$(echo $package | tr / _).test;
	filename=${filename:2}
	go test $go_test_extra_args $package -coverpkg=./... -c -o $binaries_dir/$filename
done
