#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

{{template "config" .}}

export USE_POOL_PASSWORD=yes
condor_store_cred -p password -f /htcondor_operator/pool_password

# See https://github.com/htcondor/htcondor/blob/main/build/docker/services/base/start.sh
exec bash -x /start.sh

# TODO how to run a job one off?
{{template "exit" .}}