#!/usr/bin/env bash
set -ex

host=ubuntu@starlight.i10rint.com
tmpdir=$(mktemp -d)

GOOS=linux GOARCH=amd64 go build -o $tmpdir/starlightd github.com/interstellar/starlight/cmd/starlightd
ssh $host 'sudo systemctl stop starlight'
scp $tmpdir/starlightd $host:~/starlightd
ssh $host 'sudo systemctl start starlight'
