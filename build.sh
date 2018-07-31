#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"

echo "Getting build dependencies"
go get -t github.com/docker-machine-driver-spotinst

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        arch="$GOOS-$GOARCH"
        binary="bin/docker-machine-driver-spotinst.$arch"
        echo "Building $binary"
        GOOS=$GOOS GOARCH=$GOARCH go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -o $binary github.com/docker-machine-driver-spotinst
    done
done

echo "Install adapter in local conputer"

cp bin/docker-machine-driver-spotinst.darwin-amd64 /usr/local/bin/docker-machine-driver-spotinst
