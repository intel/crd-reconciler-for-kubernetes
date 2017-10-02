package main

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"k8s.io/client-go/rest"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/xeipuuv/gojsonschema"
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

func validate(sp *crv1.StreamPrediction) error {
	schemaLoader := gojsonschema.NewReferenceLoader(
		"file:///go/src/github.com/NervanaSystems/kube-controllers-go/api/crd/stream-prediction-job-spec.json")

	data, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	documentLoader := gojsonschema.NewStringLoader(string(data))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errorOutput := ""

		for _, desc := range result.Errors() {
			errorOutput = errorOutput + " - " + desc.String() + "\n"
		}

		return fmt.Errorf(errorOutput)
	}

	return nil
}

func (h *streamPredictionHooks) process(sp *crv1.StreamPrediction) error {
	glog.Infof("Processing crd: %s", sp.Name)

	glog.Infof("Deepcopy'ing crd so that we don't change the version in cache")
	sp = sp.DeepCopy()
	sp.Status = crv1.StreamPredictionStatus{
		State:   crv1.StreamPredictionProcessed,
		Message: "Updating stream prediction",
	}

	if err := validate(sp); err != nil {
		glog.Infof("Validation error: %v\n", err)
		// TODO(niklas): When example stream prediction controller pass the spec, enable here.
		// return err
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
