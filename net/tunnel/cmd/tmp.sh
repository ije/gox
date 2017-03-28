#!/bin/bash

export GOOS=linux
export GOARCH=amd64

echo "--- compiling the x.tunnel.client (linux_amd64)..."
go build x.tunnel.client.go
