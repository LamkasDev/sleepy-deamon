#!/bin/bash
# Build
cd ../cmd
go build -o ../bin/sleepy-daemon .
cd ../..

# Run
killall sleepy-daemon
./dev/bin/sleepy-daemon -config="dev.json"