#!/bin/sh

echo "Hello, I am a config manager node with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

{{template "config" .}}

# See https://github.com/htcondor/htcondor/blob/main/build/docker/services/base/start.sh
exec bash -x /start.sh

# TODO how to run a job one off?
{{template "exit" .}}