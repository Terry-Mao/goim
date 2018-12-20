#!/usr/bin/env bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v cmd | grep -v docs | grep -v srcipts | grep -v benchmarks | grep -v examples); do
    echo "testing for $d ..."
    go test -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
