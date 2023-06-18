#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

{{template "config" .}}

# Environment variables specific to submit
{{template "condor-host" . }}

# TODO what happens here?
# https://github.com/htcondor/htcondor/blob/main/build/docker/services/base/start.sh
exec bash -x /start.sh &
while [ `condor_status -af activity|grep Idle|wc -l` -ne 2 ]
  do
    echo "Waiting for cluster to become ready";
    sleep 2
  done
echo "HTCondor properly configured"

# Keep the submit node running
sleep infinity

{{template "exit" .}}