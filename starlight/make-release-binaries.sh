#!/usr/bin/env bash
set -ex 

dir=releases
mkdir $dir
cd $dir

GOOS=linux GOARCH=amd64 go build -o starlightd i10r.io/cmd/starlightd
tar -czf starlightd-linux-amd64.tar.gz starlightd
rm starlightd
GOOS=darwin GOARCH=amd64 go build -o starlightd i10r.io/cmd/starlightd
tar -czf starlightd-darwin-amd64.tar.gz starlightd
rm starlightd
GOOS=windows GOARCH=amd64 go build -o starlightd.exe i10r.io/cmd/starlightd
zip starlightd-windows-amd64.zip starlightd.exe
rm starlightd.exe
