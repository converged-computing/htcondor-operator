/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	api "github.com/converged-computing/htcondor-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	entrypointSuffix = "-entrypoint"
)

// GetVolumeMounts returns read only volume for entrypoint scripts, etc.
func getVolumeMounts(cluster *api.HTCondor) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      cluster.Name + entrypointSuffix,
			MountPath: "/htcondor_operator/",
			ReadOnly:  true,
		},
	}
	return mounts
}

// getVolumes for the Indexed Jobs
func getVolumes(cluster *api.HTCondor) []corev1.Volume {

	// Runner start scripts
	makeExecutable := int32(0777)

	// Each of the server and nodes are given the entrypoint scripts
	// Although they won't both use them, this makes debugging easier
	runnerScripts := []corev1.KeyToPath{
		{
			Key:  "start-manager",
			Path: "start-manager.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  "start-execute",
			Path: "start-execute.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  "start-submit",
			Path: "start-submit.sh",
			Mode: &makeExecutable,
		},
		//		{
		//			Key:  tokenKey,
		//			Path: "token",
		//		},
	}

	volumes := []corev1.Volume{
		{
			Name: cluster.Name + entrypointSuffix,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{

					// Namespace based on the cluster
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.Name + entrypointSuffix,
					},
					// /htcondor_operator/start-manager.sh
					// /htcondor_operator/start-execute.sh
					// /htcondor_operator/start-submit.sh
					// /htcondor_operator/token -> /root/secrets/token
					Items: runnerScripts,
				},
			},
		},
	}
	return volumes
}
