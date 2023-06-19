#!/bin/sh

echo "Hello, I am an execute note with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

{{template "config" .}}

# Environment variables specific to submit
{{template "condor-host" . }}

# Target nodes for dedicated scheduler (allows for parallel universe for MPI)
# The variable CONDOR_HOST is set in the snippet above
# https://htcondor.readthedocs.io/en/latest/admin-manual/setting-up-special-environments.html#configuration-examples-for-dedicated-resources
echo "DedicatedScheduler = \"DedicatedScheduler@${CONDOR_HOST}\"" >> /etc/condor/condor_config.local
echo "STARTD_ATTRS = $(STARTD_ATTRS), DedicatedScheduler" >> /etc/condor/condor_config.local

# https://github.com/htcondor/htcondor/blob/main/build/docker/services/base/start.sh
exec bash -x /start.sh

{{template "exit" .}}