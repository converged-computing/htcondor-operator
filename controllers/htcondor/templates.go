/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"bytes"
	"fmt"
	"text/template"

	api "github.com/converged-computing/htcondor-operator/api/v1alpha1"

	_ "embed"
)

//go:embed templates/manager.sh
var startManagerTemplate string

//go:embed templates/execute.sh
var startExecuteTemplate string

//go:embed templates/submit.sh
var startSubmitTemplate string

//go:embed templates/components.sh
var startComponents string

//go:embed templates/token.sh
var tokenTemplate string

// NodeTemplate populates a node entrypoint
type NodeTemplate struct {
	Node        api.Node
	Spec        api.HTCondorSpec
	ClusterName string
	Namespace   string
}

// combineTemplates into one "start"
func combineTemplates(listing ...string) (t *template.Template, err error) {
	t = template.New("start")

	for i, templ := range listing {
		_, err = t.New(fmt.Sprint("_", i)).Parse(templ)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

// generateWorkerScript generates the main script to start everything up!
func generateScript(cluster *api.HTCondor, node api.Node, startTemplate string) (string, error) {
	nt := NodeTemplate{
		Node:        node,
		Spec:        cluster.Spec,
		ClusterName: cluster.Name,
		Namespace:   cluster.Namespace,
	}

	// Wrap the named template to identify it later
	startTemplate = `{{define "start"}}` + startTemplate + "{{end}}"

	// We assemble different strings (including the components) into one!
	t, err := combineTemplates(startComponents, startTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := t.ExecuteTemplate(&output, "start", nt); err != nil {
		return "", err
	}
	return output.String(), nil
}
