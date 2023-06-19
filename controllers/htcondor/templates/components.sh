#!/bin/sh

# Shared components for the broker and worker template
{{define "init"}}

# Give networking a few seconds...
echo "Sleeping for networking..."
sleep 3

# Copy over our token, and remove newline at top
# mkdir -p /root/secrets
# tr --delete '\n' < /htcondor_operator/token > /root/secrets/token

# Initialization commands
{{ .Node.Commands.Init}} > /dev/null 2>&1

# The working directory should be set by the CRD or the container
workdir=${PWD}

# And if we are using fusefs / object storage, ensure we can see contents
mkdir -p ${workdir}

# End init logic
{{end}}

{{define "config"}}

# This won't trigger to the right set of conditions in 
# /etc/condor/config.d/01-security.conf if we set to yes
# with yes, the basic cluster works, but parallel universe does not
export USE_POOL_PASSWORD=yes

# Shared logic to write a config across nodes
echo "NEGOTIATOR_INTERVAL=10" >> /etc/condor/condor_config.local

# note that this container setup seems to already be using a USE_POOL_PASSWORD with host auth
# These don't seem to work for this setup
# echo "use security:host_based" >> /etc/condor/config.d/00-insecure.config

# The top one is shown for docker-compose, the second in the config example
mkdir -p /root/secrets
condor_store_cred -p {{.Spec.Config.Password}} -f /root/secrets/pool_password 
# condor_store_cred -p {{.Spec.Config.Password}} -f /var/lib/condor/pool_password

# Try updating to allow TOKEN and PASSWORD
echo "SEC_DEFAULT_AUTHENTICATION_METHODS = FS, PASSWORD, TOKEN, IDTOKENS" >> /etc/condor/config.d/01-security.conf 
echo "ALLOW_WRITE = *" >> /etc/condor/config.d/01-security.conf 

# Austin DANGER POWERS!
echo 'ALLOW_ADVERTISE_STARTD = condor_pool@*/* $(ALLOW_ADVERTISE_STARTD)' >> /etc/condor/config.d/01-security.conf 
echo 'ALLOW_ADVERTISE_SCHEDD = condor_pool@*/* $(ALLOW_ADVERTISE_SCHEDD)' >> /etc/condor/config.d/01-security.conf 
echo 'ALLOW_ADVERTISE_MASTER = condor_pool@*/* $(ALLOW_ADVERTISE_MASTER)' >> /etc/condor/config.d/01-security.conf 

# ALLOW_WRITE = *

# TODO this should be actual cpus, not nodes
export NUM_CPUS={{.Spec.Size}}
{{end}}

{{define "exit"}}
{{ if .Spec.Interactive }}sleep infinity{{ end }}
{{ end }}

{{define "condor-host"}}
export CONDOR_HOST={{ .ClusterName }}-manager-0-0.{{ .Spec.ServiceName }}.{{ .Namespace }}.svc.cluster.local
# export CONDOR_SERVICE_HOST=${CONDOR_HOST}
{{ end }}

{{define "approve-tokens"}}
# This snippet will always approve token requests, if needed
# Install jq for now - a hack to approve token requests
yum install -y jq

# Ideally we can provide this via a config or the condor_token_request_auto_approve that takes a hostname
while true
do
    for requestid in $(condor_token_request_list -json | jq -r .[].RequestId); do
        echo "yes" | condor_token_request_approve -reqid ${requestid}
    done
    sleep 15
done
{{end}}

{{define "security-config"}}
tee -a /etc/condor/config.d/01-security.conf <<EOF
# Require authentication and integrity checking by default.
use SECURITY : With_Authentication

# Host-based security is fine in a container environment, especially if
# we're also using a pool password or a token.
use SECURITY : Host_Based
# We also want root to be able to do reconfigs, restarts, etc.
ALLOW_ADMINISTRATOR = root@$(FULL_HOSTNAME) condor@$(FULL_HOSTNAME) $(ALLOW_ADMINISTRATOR)

# SEC_DEFAULT_AUTHENTICATION_METHODS = FS, PASSWORD, TOKEN
# ALLOW_ADVERTISE_STARTD = condor_pool@*/* $(ALLOW_ADVERTISE_STARTD)
# ALLOW_ADVERTISE_SCHEDD = condor_pool@*/* $(ALLOW_ADVERTISE_SCHEDD)
# ALLOW_ADVERTISE_MASTER = condor_pool@*/* $(ALLOW_ADVERTISE_MASTER)

# Allow public reads and writes
ALLOW_READ = *
ALLOW_WRITE = *
SEC_READ_AUTHENTICATION = OPTIONAL

ALLOW_ADVERTISE_MASTER = \
    $(ALLOW_ADVERTISE_MASTER) \
    $(ALLOW_WRITE_COLLECTOR) \
    dockerworker@example.net

ALLOW_ADVERTISE_STARTD = \
    $(ALLOW_ADVERTISE_STARTD) \
    $(ALLOW_WRITE_COLLECTOR) \
    dockerworker@example.net

ALLOW_ADVERTISE_SCHEDD = \
    $(ALLOW_ADVERTISE_STARTD) \
    $(ALLOW_WRITE_COLLECTOR) \
    dockersubmit@example.net
EOF
{{end}}
# Ohno, not DNS again!
# 06/18/23 22:49:02 WARNING: Saw slow DNS query, which may impact entire system: getaddrinfo(htcondor-sample-manager-0-0.htc-service.htcondor-operator.svc.cluster.local) took 3.918414 seconds.