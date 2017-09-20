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

// The plural form of the crd
const StreamPredictionResourcePlural = "streampredictions"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamPrediction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              StreamPredictionSpec   `json:"spec"`
	Status            StreamPredictionStatus `json:"status,omitempty"`
}

// StreamPredictionSpec is the spec for the crd.
type StreamPredictionSpec struct {
	Foo string `json:"foo"`
	Bar bool   `json:"bar"`
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
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StreamPredictionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StreamPrediction `json:"items"`
}
