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

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
)

const GroupName = "aipg.intel.com"

const Version = "v1"

// The kind of the crd
const ModelTrainingResourceKind = "ModelTraining"

// The singular form of the crd
const ModelTrainingResourceSingular = "modeltraining"

// The plural form of the crd
const ModelTrainingResourcePlural = "modeltrainings"

var (
	// GVK unambiguously identifies the model training kind.
	GVK = schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    ModelTrainingResourceKind,
	}
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ModelTraining struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ModelTrainingSpec   `json:"spec"`
	Status            ModelTrainingStatus `json:"status,omitempty"`
}

func (s *ModelTraining) Name() string {
	return s.ObjectMeta.Name
}

func (s *ModelTraining) Namespace() string {
	return s.ObjectMeta.Namespace
}

func (s *ModelTraining) JSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *ModelTraining) GetStatusState() states.State {
	return s.Status.State
}

func (s *ModelTraining) SetStatusStateWithMessage(state states.State, msg string) {
	s.Status.State = state
	s.Status.Message = msg
}

func (s *ModelTraining) GetErrorState() states.State {
	return Failed
}

var terminalStates = map[states.State]struct{}{
	Failed:    {},
	Completed: {},
}

func (s *ModelTraining) IsTerminal() bool {
	_, isElement := terminalStates[s.Status.State]
	return isElement
}

// ModelTrainingState is the current job state.
type ModelTrainingState string

// ModelTrainingSpec is the spec for the crd.
type ModelTrainingSpec struct {
	ResourceSpec ResourceSpec `json:"resourceSpec"`
	State        states.State `json:"state"`
}

type ResourceSpec struct {
	Requests map[string]resource.Quantity `json:"requests"`
}

// ModelTrainingStatus is the status for the crd.
type ModelTrainingStatus struct {
	State   states.State `json:"state,omitempty"`
	Message string       `json:"message,omitempty"`
}

const (
	// Pending: In this state, a job has been created, but its sub-resources are pending.
	Pending states.State = "Pending"

	// Running: This is the _ready_ state for a model training job.
	// In this state, it is running as expected.
	Running states.State = "Running"

	// Completed: A `Completed` job has been undeployed. `Completed` is a terminal state.
	Completed states.State = "Completed"

	// Failed: A job is in an `Failed` state if an error has caused it to no longer be running as expected.
	Failed states.State = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ModelTrainingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ModelTraining `json:"items"`
}
