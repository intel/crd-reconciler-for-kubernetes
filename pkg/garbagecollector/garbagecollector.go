package garbagecollector

import (
	"context"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
)

var (
	errGCNotInitialized = errors.New("gc is not initialized")

	gc *garbageCollector
)

// GarbageCollector does its job.
type garbageCollector struct {
	namespace       string
	gvk             metav1.GroupVersionKind
	crdClient       crd.Client
	resourceClients []resource.Client
}

// Init returns a new GC.
func Init(namespace string, gvk metav1.GroupVersionKind,
	crdClient crd.Client, resourceClients []resource.Client) error {

	gc = &garbageCollector{
		namespace:       namespace,
		gvk:             gvk,
		crdClient:       crdClient,
		resourceClients: resourceClients,
	}
	return nil
}

// Run starts the GC loop via run() after checking if GC is initialized.
func Run(ctx context.Context) error {
	if gc == nil {
		return errGCNotInitialized
	}
	return gc.run(ctx)
}

// run executes the GC loop.
func (gc *garbageCollector) run(ctx context.Context) error {
	fmt.Printf("Starting GC for %v.%v.%v", gc.gvk.Group, gc.gvk.Version, gc.gvk.Kind)
	go wait.Until(gc.runGCLoop, 30*time.Second, ctx.Done())

	<-ctx.Done()
	return ctx.Err()
}

func (gc *garbageCollector) runGCLoop() {
	for _, resourceClient := range gc.resourceClients {
		resourceList := resourceClient.List(gc.namespace, gc.gvk)
	}
}
