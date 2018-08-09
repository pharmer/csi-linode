#!/usr/bin/env bash

pushd $GOPATH/src/github.com/pharmer/csi-linode/hack/gendocs
go run main.go
popd
