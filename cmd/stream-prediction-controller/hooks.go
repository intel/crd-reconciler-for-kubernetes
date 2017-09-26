package main

import (
	"github.com/golang/glog"
	"k8s.io/client-go/rest"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
)

// Implements controller.Hooks interface.
type streamPredictionHooks struct {
	crdClient *rest.RESTClient
}

func (h *streamPredictionHooks) Add(obj interface{}) {
	sp, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.V(4).Infof("Add, Got CRD: %v", sp)
	glog.Infof("onAdd, Got crd: %s", sp.SelfLink)
	h.process(sp)
}

func (h *streamPredictionHooks) Update(oldObj, newObj interface{}) {
	glog.Infof("Got update")
	sp, ok := newObj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", newObj)
		return
	}
	glog.Infof("Update, Got crd: %s", sp.ObjectMeta.SelfLink)
	h.process(sp)
}

func (h *streamPredictionHooks) Delete(obj interface{}) {
	sp, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.Infof("Delete, Got crd: %s", sp.SelfLink)
	h.process(sp)
}

func (h *streamPredictionHooks) process(sp *crv1.StreamPrediction) error {
	glog.Infof("Processing crd: %s", sp.Name)

	glog.Infof("Deepcopy'ing crd so that we don't change the version in cache")
	sp = sp.DeepCopy()
	sp.Status = crv1.StreamPredictionStatus{
		State:   crv1.StreamPredictionProcessed,
		Message: "Updating stream prediction",
	}

	err := h.crdClient.Put().
		Name(sp.Name).
		Namespace(sp.Namespace).
		Resource(crv1.StreamPredictionResourcePlural).
		Body(sp).
		Do().
		Error()

	if err != nil {
		glog.Infof("Error updating status: %v\n", err)
	} else {
		glog.Infof("Updated status: %v\n", sp)
	}
	return err
}
