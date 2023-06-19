# LAMMPS example

Follow the getting started instructions in the README.md of the repository to create a cluster
with JobSet installed and the htcondor-operator. Then apply the yaml:

```bash
$ kubectl apply -f htcondor.yaml
```

Shell in...

```bash
$ kubectl exec -it -n htcondor-operator htcondor-sample-submit-0-0-6fmjz bash
```

Note that we are already in the correct working directory - it doesn't have our LAMMPS files,
but it has the same path for the execute nodes!

```bash
$ echo $PWD
/opt/lammps/examples/reaxff/HNS
```

I was looking for a working directory [here](https://www.cl.cam.ac.uk/manuals/condor-V6_8_3-Manual/condor_submit.html) but only
found "initialdir" and the implication was that it needed to exist on the submit node too, so for now let's just make them the same.
Make sure your queue is ready (this can be delayed sometimes, and I don't know why)! It should look like this:

```bash
$ condor_q
```
```console
-- Schedd: htcondor-sample-submit-0-0.htc-service.htcondor-operator.svc.cluster.local : <10.244.0.8:40519?... @ 06/19/23 00:53:49
OWNER BATCH_NAME      SUBMITTED   DONE   RUN    IDLE   HOLD  TOTAL JOB_IDS

Total for query: 0 jobs; 0 completed, 0 removed, 0 idle, 0 running, 0 held, 0 suspended 
Total for all users: 0 jobs; 0 completed, 0 removed, 0 idle, 0 running, 0 held, 0 suspended
```

For a parallel universe job to work, we need dedicated nodes. You should see:

```bash
$ condor_status -const '!isUndefined(DedicatedScheduler)' \
      -format "%s\t" Machine -format "%s\n" DedicatedScheduler
```
```console
htcondor-sample-execute-0-0.htc-service.htcondor-operator.svls c.cluster.local     DedicatedScheduler@htcondor-sample-manager-0-0.htc-service.htcondor-operator.svc.cluster.local
htcondor-sample-execute-0-1.htc-service.htcondor-operator.svc.cluster.local     DedicatedScheduler@htcondor-sample-manager-0-0.htc-service.htcondor-operator.svc.cluster.local
```

We can first test a simple "parallel universe" job:

```
tee -a submit.sh <<EOF
universe = parallel
executable = /bin/sleep
arguments = 5
machine_count = 2
output = /tmp/sleep.out
error = /tmp/sleep.err
log = /tmp/sleep.log
request_cpus   = 1

queue
EOF
```

That should work:


Write your job file:

```bash
tee -a submit.sh <<EOF

# LAMMPS
# Simple HTCondor submit description file
# Everything with a leading # is a comment

executable   = /usr/lib64/mpich/bin/mpirun
arguments    = -np 2 --map-by socket /opt/lammps/build/lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite

output       = /tmp/lammps.out
error        = /tmp/lammps.err
log          = /tmp/lammps.log

universe = parallel
machine_count = 2
request_cpus  = 1
queue
EOF
```

Note that if you don't write to `/tmp`, you can get a permission denied error, and you can look in `cat /var/log/condor/ShadowLog`
to see the error. Here is how to submit:

```bash
$ condor_submit ./submit.sh
```

I'm currently debugging - sometimes condor_q doesn't work properly, and other times I don't see a .out or a .err.
Also important for debugging, you can find some logs here:

```bash
$ ls /var/log/condor/              
```
