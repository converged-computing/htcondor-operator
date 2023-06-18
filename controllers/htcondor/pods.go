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
)

// Get labels for any pod in the cluster
func (r *HTCondorReconciler) getPodLabels(cluster *api.HTCondor) map[string]string {
	podLabels := map[string]string{}
	podLabels["cluster-name"] = cluster.Name
	podLabels["namespace"] = cluster.Namespace
	podLabels["app.kubernetes.io/name"] = cluster.Name
	return podLabels
}
