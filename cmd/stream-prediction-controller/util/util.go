package util

import (
	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// BuildConfig gets the client config.
func BuildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// WaitForStreamPredictionInstanceProcessed waits for the stream prediction to be processed.
func WaitForStreamPredictionInstanceProcessed(streamPredictionClient *rest.RESTClient, name string) error {
	return wait.Poll(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		var streamPrediction crv1.StreamPrediction
		err := streamPredictionClient.Get().
			Resource(crv1.StreamPredictionResourcePlural).
			Namespace(apiv1.NamespaceDefault).
			Name(name).
			Do().Into(&streamPrediction)

		if err == nil && streamPrediction.Status.State == crv1.StreamPredictionProcessed {
			return true, nil
		}

		return false, err
	})
}
