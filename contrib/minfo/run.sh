#!/bin/bash

echo "> buid"
GOOS=linux GOARCH=arm go build -o minfo main.go || exit 1

echo "> copy"
rsync -r ./ pi@fluiddpi.local:minfo || exit 1

echo "> run"
ssh -t pi@fluiddpi.local "minfo/minfo"