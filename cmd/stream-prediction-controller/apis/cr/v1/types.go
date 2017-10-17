/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	"encoding/json"

	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GroupName = "aipg.intel.com"

const Version = "v1"

// The kind of the crd
const StreamPredictionResourceKind = "StreamPrediction"

// The singular form of the crd
const StreamPredictionResourceSingular = "streamprediction"

// The plural form of the crd
const StreamPredictionResourcePlural = "streampredictions"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamPrediction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              StreamPredictionSpec   `json:"spec"`
	Status            StreamPredictionStatus `json:"status,omitempty"`
}

func (s *StreamPrediction) Name() string {
	return s.ObjectMeta.Name
}

func (s *StreamPrediction) Namespace() string {
	return s.ObjectMeta.Namespace
}

func (s *StreamPrediction) JSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// StreamPredictionState is the current job state.
type StreamPredictionState string

// StreamPredictionSpec is the spec for the crd.
type StreamPredictionSpec struct {
	NeonRepoSpec    NeonRepoSpec    `json:"neonRepoSpec"`
	SecuritySpec    SecuritySpec    `json:"securitySpec"`
	StreamDataSpec  StreamDataSpec  `json:"streamDataSpec"`
	KryptonRepoSpec KryptonRepoSpec `json:"kryptonRepoSpec"`
	ResourceSpec    ResourceSpec    `json:"resourceSpec"`
	State           states.State    `json:"state"`
}

type KryptonRepoSpec struct {
	RepoURL      string `json:"repoURL"`
	Commit       string `json:"commit"`
	Image        string `json:"image"`
	SidecarImage string `json:"sidecarImage"`
}

type NeonRepoSpec struct {
	RepoURL string `json:"repoURL"`
	Commit  string `json:"commit"`
}

type StreamDataSpec struct {
	ModelPRM         string `json:"modelPRM"`
	ModelPath        string `json:"modelPath"`
	DatasetPath      string `json:"datasetPath"`
	ExtraFilename    string `json:"extraFilename"`
	CustomCodeURL    string `json:"customCodeURL"`
	CustomCommit     string `json:"customCommit"`
	AWSPath          string `json:"awsPath"`
	AWSDefaultRegion string `json:"awsDefaultRegion"`
	StreamID         int    `json:"streamID"`
	StreamName       string `json:"streamName"`
}

type SecuritySpec struct {
	PresignedToken string `json:"presignedToken"`
	JWTToken       string `json:"jwtToken"`
}

type ResourceSpec struct {
	Requests map[string]string `json:"requests"`
}

// StreamPredictionStatus is the status for the crd.
type StreamPredictionStatus struct {
	State   states.State `json:"state,omitempty"`
	Message string       `json:"message,omitempty"`
}

const (
	// Deploying In this state, a job has been created, but its sub-resources are pending.
	Deploying states.State = "Deploying"

	// Deployed This is the _ready_ state for a stream prediction job.
	// In this state, it is ready to respond to queries.
	Deployed states.State = "Deployed"

	// Completed A `Completed` job has been undeployed.  `Completed` is a terminal state.
	Completed states.State = "Completed"

	// Error A job is in an `Error` state if an error has caused it to no longer be available to respond to queries.
	Error states.State = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamPredictionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StreamPrediction `json:"items"`
}

func getStreamPredictionFSM() *states.FSM {
	fsm := states.NewFSM(
		Deploying, Deployed,
		Completed, Error,
	)
	fsm.SetAdj(Deploying, Error)
	fsm.SetAdj(Deploying, Deployed)
	fsm.SetAdj(Deploying, Completed)
	fsm.SetAdj(Deployed, Error)
	fsm.SetAdj(Deployed, Completed)
	return fsm
}

// StreamPredictionFSM represents the FSM for a Stream Prediction job
var StreamPredictionFSM *states.FSM = getStreamPredictionFSM()
