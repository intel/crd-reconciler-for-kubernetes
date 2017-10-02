package main

import (
	"fmt"

	"k8s.io/client-go/rest"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/example-controller/apis/cr/v1"
)

// exampleHooks implements controller.Hooks.
type exampleHooks struct {
	crdClient *rest.RESTClient
}

func (c *exampleHooks) Add(obj interface{}) {
	example := obj.(*crv1.Example)
	fmt.Printf("[CONTROLLER] OnAdd %s\n", example.ObjectMeta.SelfLink)
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify
	// this copy or create a copy manually for better performance.
	exampleCopy := example.DeepCopy()
	exampleCopy.Status = crv1.ExampleStatus{
		State:   crv1.ExampleStateProcessed,
		Message: "Successfully processed by controller",
	}

	err := c.crdClient.Put().
		Name(example.ObjectMeta.Name).
		Namespace(example.ObjectMeta.Namespace).
		Resource(crv1.ExampleResourcePlural).
		Body(exampleCopy).
		Do().
		Error()
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
