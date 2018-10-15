//
// Copyright (c) 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: EPL-2.0
//

package crd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
)

// CustomResource is the base type of custom resource objects.
// This allows them to be manipulated generically by the CRD client.
type CustomResource interface {
	Name() string
	Namespace() string
	JSON() (string, error)
	GetSpecState() states.State
	GetStatusState() states.State
	SetStatusStateWithMessage(states.State, string)
	DeepCopyObject() runtime.Object
	GetObjectKind() schema.ObjectKind
}

type CustomResourceList interface {
	GetItems() []runtime.Object
	DeepCopyObject() runtime.Object
	GetObjectKind() schema.ObjectKind
}
