/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/converged-computing/htcondor-operator/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	jobset "sigs.k8s.io/jobset/api/v1alpha1"
)

// newJobSet creates the jobset for the HTCondor
func (r *HTCondorReconciler) newJobSet(
	cluster *api.HTCondor,
) (*jobset.JobSet, error) {

	// When we have a success policy
	// serverName := cluster.Name + "-server"

	// When suspend is true we have a hard time debugging jobs, so keep false
	suspend := false
	jobs := jobset.JobSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: jobset.JobSetSpec{

			// The job is successful when the broker job finishes with completed (0)
			//SuccessPolicy: &jobset.SuccessPolicy{
			//	Operator:             jobset.OperatorAny,
			//	TargetReplicatedJobs: []string{serverName},
			//},
			FailurePolicy: &jobset.FailurePolicy{
				MaxRestarts: 0,
			},

			// This might be the control for child jobs (worker)
			// But I don't think we need this anymore.
			Suspend: &suspend,
		},
	}

	// Get manager job, the parent in the JobSet
	managerJob, err := r.getJob(cluster, cluster.Spec.Manager, 1, "manager", true)
	if err != nil {
		r.Log.Error(err, "There was an error getting the manager ReplicatedJob")
		return &jobs, err
	}

	submitJob, err := r.getJob(cluster, cluster.Spec.Submit, 1, "submit", true)
	if err != nil {
		r.Log.Error(err, "There was an error getting the submit ReplicatedJob")
		return &jobs, err
	}

	// Create a cluster (JobSet) with or without workers
	if cluster.Spec.Size > 0 {
		executeJob, err := r.getJob(cluster, cluster.Spec.Execute, cluster.Spec.Size, "execute", true)
		if err != nil {
			r.Log.Error(err, "There was an error getting the worker ReplicatedJob")
			return &jobs, err
		}
		jobs.Spec.ReplicatedJobs = []jobset.ReplicatedJob{managerJob, submitJob, executeJob}

	} else {
		jobs.Spec.ReplicatedJobs = []jobset.ReplicatedJob{managerJob, submitJob}
	}
	ctrl.SetControllerReference(cluster, &jobs, r.Scheme)
	return &jobs, nil
}

// getJob creates a job for a main leader (broker) or worker (followers)
func (r *HTCondorReconciler) getJob(
	cluster *api.HTCondor,
	node api.Node,
	size int32,
	entrypoint string,
	indexed bool,
) (jobset.ReplicatedJob, error) {

	backoffLimit := int32(100)
	podLabels := r.getPodLabels(cluster)
	enableDNSHostnames := false
	completionMode := batchv1.NonIndexedCompletion

	// Is this an indexed job?
	if indexed {
		completionMode = batchv1.IndexedCompletion
	}

	job := jobset.ReplicatedJob{
		Name: entrypoint,

		// This would allow pods to be reached by their hostnames!
		// It doesn't work at the moment, but could if we can specify the service name
		// <jobSet.name>-<spec.replicatedJob.name>-<job-index>-<pod-index>.<jobSet.name>-<spec.replicatedJob.name>
		Network: &jobset.Network{
			EnableDNSHostnames: &enableDNSHostnames,
		},

		Template: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			},
		},
		// This is the default, but let's be explicit
		Replicas: 1,
	}

	// Create the JobSpec for the job -> Template -> Spec
	jobspec := batchv1.JobSpec{
		BackoffLimit:          &backoffLimit,
		Completions:           &size,
		Parallelism:           &size,
		CompletionMode:        &completionMode,
		ActiveDeadlineSeconds: &cluster.Spec.DeadlineSeconds,

		// Note there is parameter to limit runtime
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    podLabels,
			},
			Spec: corev1.PodSpec{
				// matches the service
				Subdomain:     cluster.Spec.ServiceName,
				Volumes:       getVolumes(cluster),
				RestartPolicy: corev1.RestartPolicyOnFailure,
			},
		},
	}

	// Do we have a pull secret for the image?
	if node.PullSecret != "" {
		jobspec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: node.PullSecret},
		}
	}

	// Get resources for the node (server or worker)
	resources, err := r.getNodeResources(cluster, node)
	r.Log.Info("👑️ HTCondor", "Pod.Resources", resources)
	if err != nil {
		r.Log.Error(err, "❌ HTCondor", "Pod.Resources", resources)
		return job, err
	}
	jobspec.Template.Spec.Overhead = resources

	// Get volume mounts, add on container specific ones
	mounts := getVolumeMounts(cluster)
	containers, err := r.getContainers(
		cluster,
		node,
		mounts,
		entrypoint,
	)
	// Error creating containers
	if err != nil {
		r.Log.Error(err, "❌ HTCondor", "Pod.Resources", resources)
		return job, err
	}
	jobspec.Template.Spec.Containers = containers
	job.Template.Spec = jobspec
	return job, err
}
