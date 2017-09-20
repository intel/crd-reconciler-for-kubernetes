package main

import (
	"context"
	"flag"
	streampredictionclient "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/client"
	streampredictioncontroller "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/controller"
	"github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/util"
	apiv1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")
	namespace := flag.String("namespace", apiv1.NamespaceAll, "Namespace to monitor (Default all)")
	threads := flag.Int("threads", 1, "Number of threads monitoring the state")
	flag.Set("logtostderr", "true")
	flag.Parse()

	config, err := util.BuildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	apiExtensionsClientSet, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	_, err = streampredictionclient.CreateCustomResourceDefinition(apiExtensionsClientSet)
	if err != nil {
		panic(err)
	}

	client, scheme, err := streampredictionclient.NewClient(config)
	if err != nil {
		panic(err)
	}

	controller, err := streampredictioncontroller.NewStreamPredictionController(client, scheme, *namespace)
	if err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go controller.Run(ctx, *threads)

	<-ctx.Done()
}
