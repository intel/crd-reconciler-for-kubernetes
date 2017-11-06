package garbagecollector

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	"github.com/golang/glog"
)

// GarbageCollector does the following:
// - deletes any orphaned and dangling sub-resources
// - rolls-up error state of the sub-resource to the custom resource
// - re-creates missing sub-resources
type GarbageCollector struct {
	namespace       string
	gvk             schema.GroupVersionKind
	crdHandle       *crd.Handle
	crdClient       crd.Client
	resourceClients []resource.Client
}

// New returns a new GC.
func New(namespace string, gvk schema.GroupVersionKind, crdHandle *crd.Handle,
	crdClient crd.Client, resourceClients []resource.Client) *GarbageCollector {

	return &GarbageCollector{
		namespace:       namespace,
		gvk:             gvk,
		crdHandle:       crdHandle,
		crdClient:       crdClient,
		resourceClients: resourceClients,
	}
}

// Run executes the GC loop.
func (gc *GarbageCollector) Run(ctx context.Context) error {
	glog.V(4).Infof("Starting GC for %v.%v.%v", gc.gvk.Group, gc.gvk.Version, gc.gvk.Kind)
	// TODO(balajismaniam): Make the loop interval configurable.
	go wait.Until(gc.runGCLoop, 30*time.Second, ctx.Done())

	<-ctx.Done()
	return ctx.Err()
}

func (gc *GarbageCollector) runGCLoop() {
	gc.processResourceList()
}

func (gc *GarbageCollector) processResourceList() {
	for _, resourceClient := range gc.resourceClients {
		resources, err := resourceClient.List(gc.namespace)
		if err != nil {
			glog.Errorf("[crd-gc] error listing sub-resource: %v", err)
			continue
		}

		for _, resource := range resources {
			gc.processResource(resourceClient, resource)
		}
	}
}

func (gc *GarbageCollector) processResource(resourceClient resource.Client, resource metav1.Object) {
	// Get a meta.Interface object for the resource.
	rObj, err := meta.Accessor(resource)
	if err != nil {
		glog.Errorf("[crd-gc] error getting meta accessor for sub-resource: %v", err)
		return
	}

	// Check if the deletion timestamp is set on the sub-resource.
	// If it is set, there is nothing to do. Kubernetes GC will delete
	// this sub-resource.
	if rObj.GetDeletionTimestamp() != nil {
		glog.V(4).Infof("[crd-gc] ignoring sub-resource %v, %v since deletion timestamp is set",
			rObj.GetName(), rObj.GetNamespace())
		return
	}

	// Get the controller reference for the sub-resource.
	controllerRef := metav1.GetControllerOf(rObj)
	// If there is no controller reference, there is nothing to do.
	if controllerRef == nil {
		return
	}

	// If the sub-resouce is not controlled by a custom resource we
	// care about, there is nothing to do.
	if controllerRef.APIVersion != gc.gvk.GroupVersion().String() || controllerRef.Kind != gc.gvk.Kind {
		glog.V(4).Infof("[crd-gc] ignoring sub-resource %v, %v as controlling custom resource from a different group, version and kind",
			rObj.GetName(), rObj.GetNamespace())
		return
	}

	// Get the controlling custom resource.
	crObj, err := gc.crdClient.Get(rObj.GetNamespace(), controllerRef.Name)
	if err != nil {
		// If the controlling custom resource doesn't exist, then this is a
		// dangling sub-resource, we can safely delete it.
		if apierrors.IsNotFound(err) {
			err := resourceClient.Delete(rObj.GetNamespace(), rObj.GetName())
			if err != nil {
				glog.Errorf("[crd-gc] error deleting dangling sub-resource [%v, %v]: %v",
					rObj.GetName(), rObj.GetNamespace(), err)
			}
			return
		}

		glog.Errorf("[crd-gc] error getting custom resource [%v, %v]: %v",
			controllerRef.Name, rObj.GetNamespace(), err)
		return
	}

	// Get a meta.Interface object for the controlling custom resource.
	crMetaObj, err := meta.Accessor(crObj)
	if err != nil {
		glog.Errorf("[crd-gc] error getting meta accessor for controlling custom resource: %v", err)
		return
	}

	// If the deletion timestamp is set on the custom resource, then
	// there is nothing to do.
	if crMetaObj.GetDeletionTimestamp() != nil {
		glog.V(4).Infof("[crd-gc] ignoring sub-resource %v, %v since deletion timestamp is set on the controlling custom resource",
			rObj.GetName(), rObj.GetNamespace())
		return
	}

	// Assert and check if the custom resource object is a
	// crd.CustomResource.
	cr, ok := crObj.(crd.CustomResource)
	if !ok {
		glog.Errorf("[crd-gc] assertion error. expected CustomResource but got %T",
			crObj)
		return
	}

	gc.handleErrors(resourceClient, cr, rObj)
}

func (gc *GarbageCollector) handleErrors(resourceClient resource.Client, cr crd.CustomResource, rObj metav1.Object) {
	// If the custom resource is in a terminal state, or the desired state is
	// terminal, delete the sub-resource.
	if cr.IsSpecTerminal() || cr.IsStatusTerminal() {
		err := resourceClient.Delete(rObj.GetNamespace(), rObj.GetName())
		if err != nil {
			glog.Errorf("[crd-gc] error deleting failed sub-resource: %v", err)
			return
		}
	}

	// If the deletion timestamp is set, there is nothing to do.
	if rObj.GetDeletionTimestamp() != nil {
		return
	}

	if resourceClient.IsFailed(rObj.GetNamespace(), rObj.GetName()) {
		// Check whether this sub-resource is re-creatable
		if !resourceClient.IsEphemeral() {
			// Set the custom resource state to error with a message
			// and update the custom resource.
			msg := fmt.Sprintf("sub-resoure [%v, %v] is in a failed state", rObj.GetName(), rObj.GetNamespace())
			cr.SetStatusStateWithMessage(cr.GetErrorState(), msg)
			if _, err := gc.crdClient.Update(cr); err != nil {
				glog.Errorf("error updating cr [%v, %v] status after sub-resource failure [msg: %v]: %v", cr.Name(), cr.Namespace(), msg, err)
			}
			return
		}
		// Otherwise, delete the sub-resource.
		// TODO(CD): Re-create missing sub-resources.
		if err := gc.crdClient.Delete(rObj.GetNamespace(), rObj.GetName()); err != nil {
			glog.Errorf("error deleting failed sub-resource [%v, %v]: %v", cr.Name(), cr.Namespace(), err)
		}
	}
}
