package hooks

import (
	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
)

var (
	// gvk unambiguously identifies the stream predicition kind.
	gvk = schema.GroupVersionKind{
		Group:   crv1.GroupName,
		Version: crv1.Version,
		Kind:    crv1.StreamPredictionResourceKind,
	}
)

// StreamPredictionHooks implements controller.Hooks interface
type StreamPredictionHooks struct {
	resourceClients []resource.Client
	crdClient       crd.Client
}

// NewStreamPredictionHooks creates and returns a new instance of the StreamPredictionHooks
func NewStreamPredictionHooks(crdClient crd.Client, resourceClients []resource.Client) *StreamPredictionHooks {
	return &StreamPredictionHooks{
		resourceClients: resourceClients,
		crdClient:       crdClient,
	}
}

// Add handles the addition of a new stream prediction object
func (h *StreamPredictionHooks) Add(obj interface{}) {
	streamCrd, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.V(4).Infof("add, got CRD: %v", streamCrd)

	streamPredict := streamCrd.DeepCopy()

	if err := h.addResources(streamPredict); err != nil {
		// Delete all of the sub-resources.
		h.deleteResources(streamPredict)

		streamPredict.Status = crv1.StreamPredictionStatus{
			State:   crv1.StreamPredictionError,
			Message: "Failed to deploy StreamPrediction",
		}
		if h.crdClient.Update(streamPredict); err != nil {
			glog.Infof("error updating status: %v\n", err)
		}
		return
	}

	streamPredict.Status = crv1.StreamPredictionStatus{
		State:   crv1.StreamPredictionDeployed,
		Message: "Deployed Sub-Resources",
	}
	if err := h.crdClient.Update(streamPredict); err != nil {
		glog.Infof("error updating status: %v\n", err)
		return
	}
	glog.Infof("updated status: %v\n", streamPredict)
}

// Update handles the update of a stream prediction object
func (h *StreamPredictionHooks) Update(oldObj, newObj interface{}) {
	streamPredict, ok := newObj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("object received is not of type StreamPrediction %v", newObj)
		return
	}
	glog.Infof("update, Got crd: %s", streamPredict.ObjectMeta.SelfLink)
}

// Delete handles the deletion of a stream prediction object
func (h *StreamPredictionHooks) Delete(obj interface{}) {
	streamPredict, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.Infof("delete, got crd: %s", streamPredict.SelfLink)

	//Delete the resources using name for now.
	h.deleteResources(streamPredict)
}

func (h *StreamPredictionHooks) addResources(streamPredict *crv1.StreamPrediction) error {
	// Add controller reference.
	// See https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
	// for more details on owner references.
	ownerRef := metav1.NewControllerRef(streamPredict, gvk)

	for _, resourceClient := range h.resourceClients {
		err := resourceClient.Create(streamPredict.Namespace(), struct {
			*crv1.StreamPrediction
			metav1.OwnerReference
		}{
			streamPredict,
			*ownerRef,
		})
		if err != nil {
			glog.Errorf("received err: %v while creating object", err)
			return err
		}
	}
	glog.Info("resource creation complete for stream prediction \"%s\"", streamPredict.Name())
	return nil
}

func (h *StreamPredictionHooks) deleteResources(streamPredict *crv1.StreamPrediction) {
	for _, resourceClient := range h.resourceClients {
		if err := resourceClient.Delete(streamPredict.Namespace(), streamPredict.Name()); err != nil {
			glog.Errorf("resource deletion failed for stream prediction \"%s\": %v", streamPredict.Name(), err)
		}
	}
	glog.Info("resource deletion complete for stream prediction \"%s\"", streamPredict.Name())
}
