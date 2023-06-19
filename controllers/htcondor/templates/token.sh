#!/bin/sh

# Start the server, and keep trying to generate token until it works

export USE_POOL_PASSWORD=yes
condor_store_cred -p {{.Spec.Config.Password}} -f /root/secrets/pool_password 
exec bash -x /start.sh &

# Create a token for the execute nodes. 
# And this will bind to /htcondor_operator and be copied to /root/secrets/token
retval=1
while [ ${retval} -ne 0 ]
  do
    echo "Waiting for cluster to become ready to generate token...";
    condor_token_create -authz ADVERTISE_MASTER -authz ADVERTISE_STARTD -authz READ -identity dockerworker@example.net -token dockerworker_token
    retval=$?
    sleep 2
  done
echo "HTCondor properly configured"
echo "CUT HERE"
cat /etc/condor/tokens.d/dockerworker_token