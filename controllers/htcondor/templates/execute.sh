#!/bin/sh

echo "Hello, I am an execute note with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

{{template "config" .}}

# Environment variables specific to submit
{{template "condor-host" . }}

# https://github.com/htcondor/htcondor/blob/main/build/docker/services/base/start.sh
exec bash -x /start.sh

{{template "exit" .}}