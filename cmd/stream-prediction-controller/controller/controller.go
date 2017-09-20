package controller

import (
	"context"
	"fmt"
	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

// NewStreamPredictionController creates a new stream prediction CRD controller.
func NewStreamPredictionController(client *rest.RESTClient, scheme *runtime.Scheme, namespace string) (*StreamPredictionController, error) {

	streamPredictionController := &StreamPredictionController{
		client:    client,
		scheme:    scheme,
		queue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		namespace: namespace,
	}

	listwatch := cache.NewListWatchFromClient(
		streamPredictionController.client,
		crv1.StreamPredictionResourcePlural,
		namespace,
		fields.Everything())

	indexer, informer := cache.NewIndexerInformer(
		listwatch,
		&crv1.StreamPrediction{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    streamPredictionController.onAdd,
			UpdateFunc: streamPredictionController.onUpdate,
			DeleteFunc: streamPredictionController.onDelete,
		}, cache.Indexers{},
	)

	streamPredictionController.informer = informer
	streamPredictionController.indexer = crv1.NewStreamPredictionLister(indexer)
	return streamPredictionController, nil
}

func (stream_prediction_controller *StreamPredictionController) enqueue(crd *crv1.StreamPrediction) {
	if key, err := cache.MetaNamespaceKeyFunc(crd); err == nil {
		glog.Infof("Adding key: %s", key)
		stream_prediction_controller.queue.Add(key)
	} else {
		glog.Infof("Failed to add CRD to queue %s", crd.Name)
	}
}

func (stream_prediction_controller *StreamPredictionController) onAdd(obj interface{}) {

	streamCrd, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.V(4).Infof("onAdd, Got CRD: %v", streamCrd)
	glog.Infof("onAdd, Got crd: %s", streamCrd.ObjectMeta.SelfLink)
	stream_prediction_controller.enqueue(streamCrd)
}

func (stream_prediction_controller *StreamPredictionController) onUpdate(oldObj, newObj interface{}) {
	glog.Infof("Got update")
	streamCrd, ok := newObj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", newObj)
		return
	}
	glog.Infof("onUpdate, Got crd: %s", streamCrd.ObjectMeta.SelfLink)
	stream_prediction_controller.enqueue(streamCrd)
}

func (stream_prediction_controller *StreamPredictionController) onDelete(obj interface{}) {
	streamCrd, ok := obj.(*crv1.StreamPrediction)
	if !ok {
		glog.Errorf("Object received is not of type StreamPrediction %v", obj)
		return
	}
	glog.Infof("onDelete, Got crd: %s", streamCrd.ObjectMeta.SelfLink)
	stream_prediction_controller.enqueue(streamCrd)
}

// Run starts the controller.
func (stream_prediction_controller *StreamPredictionController) Run(ctx context.Context, threads int) {
	defer stream_prediction_controller.queue.ShutDown()
	glog.Info("Starting stream prediction CRD controller")
	go stream_prediction_controller.informer.Run(ctx.Done())

	// Wait for cache to sync
	if !cache.WaitForCacheSync(ctx.Done(), stream_prediction_controller.informer.HasSynced) {
		glog.Errorf("Cache sync timed out")
		return
	}
	for i := 0; i < threads; i++ {
		go wait.Until(stream_prediction_controller.runWorker, time.Second, ctx.Done())
	}

	<-ctx.Done()
	glog.Infof("Stopping controller")
}

func (stream_prediction_controller *StreamPredictionController) runWorker() {
	for stream_prediction_controller.processQueue() {
	}
}

func (stream_prediction_controller *StreamPredictionController) processQueue() bool {
	key, quit := stream_prediction_controller.queue.Get()
	if quit {
		return false
	}

	defer stream_prediction_controller.queue.Done(key)

	err := stream_prediction_controller.proccessItem(key.(string))
	stream_prediction_controller.handleError(err, key)
	return true
}

func (stream_prediction_controller *StreamPredictionController) handleError(err error, key interface{}) {
	if err == nil {
		glog.V(4).Info("No error encountered, removing from queue")
		stream_prediction_controller.queue.Forget(key)
		return
	}

	if stream_prediction_controller.queue.NumRequeues(key) < 5 {
		glog.V(4).Infof("Error processing crd %s: %s\n Queuing again", key, err)
		stream_prediction_controller.queue.AddRateLimited(key)
		return
	}

	glog.V(4).Info("Max attempts reached. Forgetting %s from queue.", key)
	stream_prediction_controller.queue.Forget(key)
}

func (stream_prediction_controller *StreamPredictionController) proccessItem(key string) error {
	crd, err := stream_prediction_controller.indexer.Get(key)
	if err != nil {
		return fmt.Errorf("Could not get crd: %s from indexer", key)
	}

	glog.Infof("Processing crd: %s", crd.Name)

	glog.Infof("Deepcopy'ing crd so that we don't change the version in cache")
	streamPredict := crd.DeepCopy()
	streamPredict.Status = crv1.StreamPredictionStatus{
		State:   crv1.StreamPredictionProcessed,
		Message: "Updating stream prediction",
	}

	err = stream_prediction_controller.client.Put().
		Name(streamPredict.ObjectMeta.Name).
		Namespace(streamPredict.ObjectMeta.Namespace).
		Resource(crv1.StreamPredictionResourcePlural).
		Body(streamPredict).
		Do().
		Error()

	if err != nil {
		glog.Infof("Error updating status: %v\n", err)
	} else {
		glog.Infof("Updated status: %v\n", streamPredict)
	}
	return err
}
