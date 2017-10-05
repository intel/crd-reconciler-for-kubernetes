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

// StreamPredictionSpec is the spec for the crd.
type StreamPredictionSpec struct {
	NeonRepoSpec    NeonRepoSpec
	SecuritySpec    SecuritySpec
	StreamDataSpec  StreamDataSpec
	KryptonRepoSpec KryptonRepoSpec
}

type KryptonRepoSpec struct {
	RepoURL             string `json:"repoURL"`
	Commit              string `json:"commit"`
	KryptonImage        string `json:"kryptonImage"`
	KryptonSidecarImage string `json:"kryptonSidecarImage"`
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

// StreamPredictionStatus is the status for the crd.
type StreamPredictionStatus struct {
	State   StreamPredictionState `json:"state,omitempty"`
	Message string                `json:"message,omitempty"`
}

// StreamPredictionState is the current state
type StreamPredictionState string

const (
	// StreamPredictionCreated is set when the the resource is created
	StreamPredictionCreated StreamPredictionState = "Created"
	// StreamPredictionProcessed is set when the the resource is processed by the controller
	StreamPredictionProcessed StreamPredictionState = "Processed"
	// StreamPredictionError is set when there was an error in the deployment
	StreamPredictionError StreamPredictionState = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamPredictionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StreamPrediction `json:"items"`
}
