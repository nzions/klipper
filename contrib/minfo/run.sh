#!/bin/bash

echo "copy >"
scp main.go pi@fluiddpi.local:
scp go.mod pi@fluiddpi.local:
scp -r klippyclient pi@fluiddpi.local:
echo "run>"
ssh pi@fluiddpi.local go run main.go