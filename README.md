# htcondor-operator

> Let's make one for HTCondor too!

This will be an operator that attempts to use [htcondor](https://github.com/htcondor/htcondor) to create a cluster to run tasks. Note that I'm going off of the list [here](https://github.com/dask/dask-jobqueue/tree/main/ci) (plus Flux). We will use the design and referenced docker-compose [from here](https://github.com/dask/dask-jobqueue/blob/main/ci/htcondor/docker-compose.yml),
and the base images [are here](https://github.com/htcondor/htcondor/tree/main/build/docker/services).

## Development

### Creation

```bash
mkdir htcondor-operator
cd htcondor-operator/
operator-sdk init --domain flux-framework.org --repo github.com/converged-computing/htcondor-operator
operator-sdk create api --version v1alpha1 --kind HTCondor --resource --controller
```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

Create a cluster with kind:

```bash
$ kind create cluster
```

You'll need to install the jobset API, which eventually will be added to Kubernetes proper (but is not yet!)

```bash
VERSION=v0.1.3
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```
or development (main) version:

```bash
$ kubectl apply --server-side -k github.com/kubernetes-sigs/jobset/config/default?ref=main
```

But if you need a previous version (e.g., the current main is v1alpha2, and I want to use v1alpha1 until there is an official release) you can install from a commit:

```
# This is right before upgrade to v1alpha2, or June 2nd when I was testing!
# This is also a strategy for deploying a test version
git clone https://github.com/kubernetes-sigs/jobset /tmp/jobset
cd /tmp/jobset
git checkout 93bd85c76fc8afa79b4b5c6d1df9075c99c9f22d
IMAGE_TAG=vanessa/jobset:test make image-build
IMAGE_TAG=vanessa/jobset:test make image-push
IMAGE_TAG=vanessa/jobset:test make deploy
```

Then (back in the operator directory here) you can generate the custom resource definition

```bash
# Build and push the image, and generate the examples/dist/htcondor-operator-dev.yaml
$ make test-deploy DEVIMG=<some-registry>/htcondor-operator:tag

# As an example
$ make test-deploy DEVIMG=vanessa/htcondor-operator:test
```

Make our namespace:

```bash
$ kubectl create namespace htcondor-operator
```

Apply the new config!

```bash
$ kubectl apply -f examples/dist/htcondor-operator-dev.yaml
```

See logs for the operator

```bash
$ kubectl logs -n htcondor-operator-system htcondor-operator-controller-manager-6f6945579-9pknp 
```

Create a "hello-world" interactive cluster:

```bash
$ kubectl apply -f examples/tests/hello-world/htcondor.yaml 
```

Ensure pods are running (it will take about a minute to pull the containers):

```bash
$ make list
kubectl get -n htcondor-operator pods
NAME                                READY   STATUS    RESTARTS   AGE
htcondor-sample-execute-0-0-wwzxk   1/1     Running   0          57s
htcondor-sample-execute-0-1-tszkn   1/1     Running   0          57s
htcondor-sample-manager-0-0-4x6dj   1/1     Running   0          57s
htcondor-sample-submit-0-0-jfbh6    1/1     Running   0          57s
```

The cluster will have a central manager, a submit node, and two execution nodes.
You can look at their logs to see the cluster running:

```console
++ condor_config_val -evaluate MAX_FILE_DESCRIPTORS
+ config_max=
+ [[ '' =~ ^[0-9]+$ ]]
+ [[ -s /etc/condor/config.d/01-fdfix.conf ]]
+ [[ -f /root/config/pre-exec.sh ]]
+ bash -x /root/config/pre-exec.sh
+ exec /usr/bin/supervisord -c /etc/supervisord.conf
2023-06-18 21:23:07,831 INFO Set uid to user 0 succeeded
2023-06-18 21:23:07,851 INFO RPC interface 'supervisor' initialized
2023-06-18 21:23:07,851 CRIT Server 'unix_http_server' running without any HTTP authentication checking
2023-06-18 21:23:07,851 INFO supervisord started with pid 1
2023-06-18 21:23:08,855 INFO spawned: 'condor_master' with pid 39
2023-06-18 21:23:09,857 INFO success: condor_master entered RUNNING state, process has stayed up for > than 1 seconds (startsecs)
```

I found that if I shell into the submit node right away, it will tell me it can't reach the host.
But I've also seen it go up more quickly, so I'm not sure. I need to use it more to reflect.
For now let's shell into submit and see if we can interact! This is only the second time I'll use htcondor so with me luck :)

```bash
$ kubectl exec -it -n htcondor-operator htcondor-sample-submit-0-0-jfbh6 bash
```

The working queue should look like this:

```bash
$ condor_q
```
```console
-- Schedd: htcondor-sample-submit-0-0.htc-service.htcondor-operator.svc.cluster.local : <10.244.0.8:40519?... @ 06/19/23 00:53:49
OWNER BATCH_NAME      SUBMITTED   DONE   RUN    IDLE   HOLD  TOTAL JOB_IDS

Total for query: 0 jobs; 0 completed, 0 removed, 0 idle, 0 running, 0 held, 0 suspended 
Total for all users: 0 jobs; 0 completed, 0 removed, 0 idle, 0 running, 0 held, 0 suspended
```

I think the instructions we want are for [condor_submit](https://htcondor.readthedocs.io/en/latest/users-manual/submitting-a-job.html)
and it also looks like we need a submit file for that, let's write one, and saving logs to `/tmp` for now.

```bash
tee -a submit.sh <<EOF

# Example 1
# Simple HTCondor submit description file
# Everything with a leading # is a comment

executable   = /usr/bin/echo
arguments    = hello world

output       = /tmp/hello-world.out
error        = /tmp/hello-world.err
log          = /tmp/hello-world.log

request_cpus   = 1
queue
EOF
```

Note that if you don't write to `/tmp`, you can get a permission denied error, and you can look in `cat /var/log/condor/ShadowLog`
to see the error. In a real world setup you would be running as a condor user, and not root, but we are just playing
around for now. Here is how to submit the job:

```bash
$ condor_submit ./submit.sh
```

If you quickly look at the queue you'd see it, but it seems to run so fast that I missed it! But I can see the output:

```bash
$ cat /tmp/hello-world.out
hello world
```

Important for debugging - you can find some logs here:

```bash
$ cat /var/log/condor/              
.master_address      MasterLog            ProcLog              SchedLog             ScheddRestartReport
ShadowLog            SharedPortLog        XferStatsLog         transfer_history  
```

I found the ShadowLog helpful for debugging submits, and the Master/SchedLog helpful too!
And that's it for now - I am next going to try to use these images to prepare something that can run LAMMPS.
We could also try a container - I will likely do both.

## TODO

- A non-interactive submit needs to write a submission file (might not be needed)
- Tweak the configs to get envars that have the cluster size (right now just hard coded to 2)
- Test out [containers](https://chtc.cs.wisc.edu/uw-research-computing/docker-jobs)
- LAMMPS example working


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
