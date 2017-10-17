package hooks

import (
	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
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
	fsm             *states.FSM
}

// NewStreamPredictionHooks creates and returns a new instance of the StreamPredictionHooks
func NewStreamPredictionHooks(crdClient crd.Client, resourceClients []resource.Client, fsm *states.FSM) *StreamPredictionHooks {
	return &StreamPredictionHooks{
		resourceClients: resourceClients,
		crdClient:       crdClient,
		fsm:             fsm,
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

	if streamPredict.Spec.State != crv1.Deployed || streamPredict.Status.State != crv1.Deploying {
		glog.Info("New stream spec is not in a deployed state and status is not in deploying state")
		streamPredict.Status = crv1.StreamPredictionStatus{
			State:   crv1.Error,
			Message: "Failed to deploy StreamPrediction",
		}
		if err := h.crdClient.Update(streamPredict); err != nil {
			glog.Infof("error updating status: %v\n", err)
		}
		return
	}

	if err := h.addResources(streamPredict); err != nil {
		// Delete all of the sub-resources.
		h.deleteResources(streamPredict)

		streamPredict.Status = crv1.StreamPredictionStatus{
			State:   crv1.Error,
			Message: "Failed to deploy StreamPrediction",
		}
		if err = h.crdClient.Update(streamPredict); err != nil {
			glog.Infof("error updating status: %v\n", err)
		}
		return
	}

	streamPredict.Status = crv1.StreamPredictionStatus{
		State:   crv1.Deployed,
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
	newStreamPredict, ok := newObj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("object received is not of type StreamPrediction %v", newObj)
		return
	}

	oldStreamPredict, ok := oldObj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("object received is not of type StreamPrediction %v", oldObj)
		return
	}

	if newStreamPredict.Spec.State == oldStreamPredict.Spec.State {
		glog.Infof("Received an update of the same state. Old crd: %s, New crd: %s", oldStreamPredict, newStreamPredict)
		return
	}

	if newStreamPredict.Spec.State == newStreamPredict.Status.State {
		glog.Infof("Received an update in the same state: Old state %s, New state: %s", newStreamPredict.Status.State, newStreamPredict.Spec.State)
		return
	}

	if !h.fsm.PathExists(newStreamPredict.Status.State, newStreamPredict.Spec.State) {
		glog.Infof("Got an update to an invalid state. Current state: %v, requested state %v", oldStreamPredict.Status.State, newStreamPredict.Spec.State)
		return
	}

	switch newStreamPredict.Spec.State {

	case crv1.Completed:
		glog.Infof("Got an update for completing the stream predict %v", newStreamPredict)
		// Delete the subresources and update the status
		h.deleteResources(newStreamPredict)
		newStreamPredict.Status = crv1.StreamPredictionStatus{
			State:   crv1.Completed,
			Message: "Stream Prediction completed",
		}
		if err := h.crdClient.Update(newStreamPredict); err != nil {
			glog.Infof("error updating status: %v\n", err)
			return
		}
		glog.Info("Successfully deleted subresources")
	}
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
