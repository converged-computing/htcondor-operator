/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HTCondorSpec defines the desired state of HTCondor
type HTCondorSpec struct {
	// Config Manager is the main server to run HTCondor
	//+optional
	Manager Node `json:"manager"`

	// Submission node
	//+optional
	Submit Node `json:"submit"`

	// Name for the cluster service
	//+optional
	ServiceName string `json:"serviceName"`

	// Configuration values
	//+optional
	Config Config `json:"config"`

	// Execute is for an execution worker node
	//+optional
	Execute Node `json:"execute"`

	// Size of the HTCondor (1 server + (N-1) nodes)
	Size int32 `json:"size"`

	// Interactive mode keeps the cluster running
	// +optional
	Interactive bool `json:"interactive"`

	// Time limit for the job
	// Approximately one year. This cannot be zero or job won't start
	// +kubebuilder:default=31500000
	// +default=31500000
	// +optional
	DeadlineSeconds int64 `json:"deadlineSeconds,omitempty"`

	// Resources include limits and requests
	// +optional
	Resources Resource `json:"resources"`

	// Security Context
	// These are applied to all nodes
	// https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	// +optional
	SecurityContext SecurityContext `json:"securityContext"`
}

type SecurityContext struct {

	// Privileged container
	// +optional
	Privileged bool `json:"privileged,omitempty"`
}

type Config struct {

	// Password for htcondor setup
	// +optional
	Password string `json:"password"`
}

// Node corresponds to a pod (server or worker)
type Node struct {

	// Image to use for HTCondor
	// +optional
	Image string `json:"image"`

	// Resources include limits and requests
	// +optional
	Resources Resources `json:"resources"`

	// PullSecret for the node, if needed
	// +optional
	PullSecret string `json:"pullSecret"`

	// Command will be honored by a server node
	// +optional
	Command string `json:"command,omitempty"`

	// Commands to run around different parts of the hyperqueu setup
	// +optional
	Commands Commands `json:"commands,omitempty"`

	// Working directory
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// PullAlways will always pull the container
	// +optional
	PullAlways bool `json:"pullAlways"`

	// Ports to be exposed to other containers in the cluster
	// We take a single list of integers and map to the same
	// +optional
	// +listType=atomic
	Ports []int32 `json:"ports"`

	// Key/value pairs for the environment
	// +optional
	Environment map[string]string `json:"environment"`
}

// ContainerResources include limits and requests
type Commands struct {

	// Init runs before anything in both scripts
	// +optional
	Init string `json:"init,omitempty"`
}

// ContainerResources include limits and requests
type Resources struct {

	// +optional
	Limits Resource `json:"limits"`

	// +optional
	Requests Resource `json:"requests"`
}

type Resource map[string]intstr.IntOrString

// Validate the HTCondor
func (hq *HTCondor) Validate() bool {
	if hq.Spec.Manager.Image == "" {
		hq.Spec.Manager.Image = "htcondor/cm:el7"
	}
	// It seems like the docker-compose setup used execute as the final submit image
	if hq.Spec.Submit.Image == "" {
		hq.Spec.Submit.Image = "htcondor/submit:el7"
	}
	if hq.Spec.Execute.Image == "" {
		hq.Spec.Execute.Image = "htcondor/execute:el7"
	}
	if hq.Spec.ServiceName == "" {
		hq.Spec.ServiceName = "htc-service"
	}
	// This obviously isn't great
	if hq.Spec.Config.Password == "" {
		hq.Spec.Config.Password = "password"
	}
	return true
}

// WorkerNodes returns the number of worker nodes
// At this point we've already validated size is >= 1
func (hq *HTCondor) WorkerNodes() int32 {
	return hq.Spec.Size - 1
}

// HTCondorStatus defines the observed state of HTCondor
type HTCondorStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HTCondor is the Schema for the htcondors API
type HTCondor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTCondorSpec   `json:"spec,omitempty"`
	Status HTCondorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HTCondorList contains a list of HTCondor
type HTCondorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTCondor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HTCondor{}, &HTCondorList{})
}
