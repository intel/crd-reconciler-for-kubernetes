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

package main

import (
	"fmt"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/example-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
)

// exampleHooks implements controller.Hooks.
type exampleHooks struct {
	crdClient crd.Client
}

func (c *exampleHooks) Add(obj interface{}) {
	example := obj.(*crv1.Example)
	fmt.Printf("[CONTROLLER] OnAdd %s\n", example.ObjectMeta.SelfLink)
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify
	// this copy or create a copy manually for better performance.
	exampleCopy := example.DeepCopy()
	exampleCopy.Status = crv1.ExampleStatus{
		State:   states.Completed,
		Message: "Successfully processed by controller",
	}

	_, err := c.crdClient.Update(exampleCopy)
	if err != nil {
		fmt.Printf("ERROR updating status: %v\n", err)
	} else {
		fmt.Printf("UPDATED status: %#v\n", exampleCopy)
	}
}

func (c *exampleHooks) Update(oldObj, newObj interface{}) {
	oldExample := oldObj.(*crv1.Example)
	newExample := newObj.(*crv1.Example)
	fmt.Printf("[CONTROLLER] OnUpdate oldObj: %s\n", oldExample.ObjectMeta.SelfLink)
	fmt.Printf("[CONTROLLER] OnUpdate newObj: %s\n", newExample.ObjectMeta.SelfLink)
}

func (c *exampleHooks) Delete(obj interface{}) {
	example := obj.(*crv1.Example)
	fmt.Printf("[CONTROLLER] OnDelete %s\n", example.ObjectMeta.SelfLink)
}
