package controller

import (
	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// StreamPredictionController is the struct which stores info about the controller.
type StreamPredictionController struct {
	client    *rest.RESTClient
	scheme    *runtime.Scheme
	informer  cache.Controller
	queue     workqueue.RateLimitingInterface
	namespace string
	indexer   crv1.StreamPredictionLister
}
