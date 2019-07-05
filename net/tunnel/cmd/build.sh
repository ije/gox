#!/bin/bash

goos="linux"
read -p "please enter the deploy operating system(default is 'linux'): " sys
if [ "$sys" != "" ]; then
	goos="$sys"
fi
export GOOS=$goos

goarch="amd64"
read -p "please enter the deploy OS Architecture(default is 'amd64'): " arch
if [ "$arch" != "" ]; then
	goarch="$arch"
fi
export GOARCH=$goarch

echo "--- compiling the tunnel.$1 (${goos}_$goarch)..."
go build $1/$1.go
