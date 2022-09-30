#!/bin/bash
# Settings
SLEEPY_VERSION=$(cat current_version.txt)

# Run
killall sleepy-daemon
./$SLEEPY_VERSION/bin/sleepy-daemon -config="current.json"