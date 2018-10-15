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

package fake

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
)

// ClientImpl is a fake implementation of crd.Client
type ClientImpl struct {
	CustomResourceImpl     *CustomResourceImpl
	CustomResourceListImpl *CustomResourceListImpl
	Error                  string
}

// returns a fake RESTClient.
// TODO Not used in unit tests, returns nil
func (c *ClientImpl) RESTClient() rest.Interface {
	return nil
}

// Create creates the supplied CRD.
func (c *ClientImpl) Create(cr crd.CustomResource) (e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	return
}

// Get retrieves the CRD from the Kubernetes API server.
func (c *ClientImpl) Get(namespace string, name string) (result runtime.Object, e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	result = c.CustomResourceImpl
	return
}

// Update updates the CRD on the Kubernetes API server.
func (c *ClientImpl) Update(cr crd.CustomResource) (result runtime.Object, e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	result = c.CustomResourceImpl
	return
}

// Delete deletes the CRD from the Kubernetes API server.
func (c *ClientImpl) Delete(namespace string, name string) (e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	return
}

// Validate validates a custom resource against a json schema.
// Returns nil if object adheres to the schema.
func (c *ClientImpl) Validate(cr crd.CustomResource) (e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	return
}

func (c *ClientImpl) List(namespace string, labels map[string]string) (result runtime.Object, e error) {
	if c.Error != "" {
		e = fmt.Errorf(c.Error)
	}
	result = c.CustomResourceListImpl
	return
}
