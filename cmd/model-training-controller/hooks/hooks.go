package hooks

import (
	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/model-training-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
)

// ModelTrainingHooks implements controller.Hooks interface
type ModelTrainingHooks struct {
	resourceClients []resource.Client
	crdClient       crd.Client
}

// NewModelTrainingHooks creates and returns a new instance of the ModelTrainingHooks
func NewModelTrainingHooks(crdClient crd.Client, resourceClients []resource.Client) *ModelTrainingHooks {
	return &ModelTrainingHooks{
		resourceClients: resourceClients,
		crdClient:       crdClient,
	}
}

// Add handles the addition of a new model training object
func (h *ModelTrainingHooks) Add(obj interface{}) {
	modelTrainingCrd, ok := obj.(*crv1.ModelTraining)
	if !ok {
		glog.Errorf("object received is not of type ModelTraining %v", obj)
		return
	}
	glog.V(4).Infof("Model Training add hook - got: %v", modelTrainingCrd)
}

// Update handles the update of a model training object
func (h *ModelTrainingHooks) Update(oldObj, newObj interface{}) {
	newModelTraining, ok := newObj.(*crv1.ModelTraining)
	if !ok {
		glog.Errorf("object received is not of type ModelTraining %v", newObj)
		return
	}

	oldModelTraining, ok := oldObj.(*crv1.ModelTraining)
	if !ok {
		glog.Errorf("object received is not of type ModelTraining %v", oldObj)
		return
	}

	glog.V(4).Infof("Model Training update hook - got old: %v new: %v", oldModelTraining, newModelTraining)
}

// Delete handles the deletion of a model training object
func (h *ModelTrainingHooks) Delete(obj interface{}) {
	modelTrain, ok := obj.(*crv1.ModelTraining)
	if !ok {
		glog.Errorf("object received is not of type ModelTraining %v", obj)
		return
	}
	glog.V(4).Infof("Model Training add hook - got: %v", modelTrain)

	//Delete the resources using name for now.
	h.deleteResources(modelTrain)
}

func (h *ModelTrainingHooks) addResources(modelTrain *crv1.ModelTraining) error {
	// Add controller reference.
	// See https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
	// for more details on owner references.
	ownerRef := metav1.NewControllerRef(modelTrain, crv1.GVK)

	for _, resourceClient := range h.resourceClients {
		err := resourceClient.Create(modelTrain.Namespace(), struct {
			*crv1.ModelTraining
			metav1.OwnerReference
		}{
			modelTrain,
			*ownerRef,
		})
		if err != nil {
			glog.Errorf("received err: %v while creating object", err)
			return err
		}
	}
	glog.Infof("resource creation complete for model training \"%s\"", modelTrain.Name())
	return nil
}

func (h *ModelTrainingHooks) deleteResources(modelTrain *crv1.ModelTraining) {
	for _, resourceClient := range h.resourceClients {
		if err := resourceClient.Delete(modelTrain.Namespace(), modelTrain.Name()); err != nil {
			glog.Errorf("resource deletion failed for model training \"%s\": %v", modelTrain.Name(), err)
		}
	}
	glog.Info("resource deletion complete for model training \"%s\"", modelTrain.Name())
}
