#!/bin/sh

# Shared components for the broker and worker template
{{define "init"}}

# Give networking a few seconds...
echo "Sleeping for networking..."
sleep 3

# Initialization commands
{{ .Node.Commands.Init}} > /dev/null 2>&1

# The working directory should be set by the CRD or the container
workdir=${PWD}

# And if we are using fusefs / object storage, ensure we can see contents
mkdir -p ${workdir}

# End init logic
{{end}}

{{define "config"}}
# Shared logic to write a config across nodes
echo "NEGOTIATOR_INTERVAL=10" >> /etc/condor/condor_config.local
# echo "SEC_PASSWORD_FILE = $(LOCAL_DIR)/lib/condor/pool_password" >> /etc/condor/condor_config.local

# Generate password
mkdir -p /root/secrets
# chmod 0700 /root/secrets
# The top one is shown for docker-compose, the second in the config example
condor_store_cred -p {{.Spec.Config.Password}} -f /root/secrets/pool_password 
# condor_store_cred -p {{.Spec.Config.Password}} -f /var/lib/condor/pool_password

# TODO this should be actual cpus, not nodes
export NUM_CPUS={{.Spec.Size}}
{{end}}

{{define "exit"}}
{{ if .Spec.Interactive }}sleep infinity{{ end }}
{{ end }}

{{define "condor-host"}}

export USE_POOL_PASSWORD=yes
export CONDOR_HOST={{ .ClusterName }}-manager-0-0.{{ .Spec.ServiceName }}.{{ .Namespace }}.svc.cluster.local
# export CONDOR_SERVICE_HOST=${CONDOR_HOST}
{{ end }}

# Ohno, not DNS again!
# 06/18/23 22:49:02 WARNING: Saw slow DNS query, which may impact entire system: getaddrinfo(htcondor-sample-manager-0-0.htc-service.htcondor-operator.svc.cluster.local) took 3.918414 seconds.