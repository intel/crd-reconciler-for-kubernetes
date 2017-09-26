package main

import (
	"context"
	"flag"

	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	// Uncomment the following line to load the gcp plugin (only required to
	// authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/controller"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/util"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")
	namespace := flag.String("namespace", apiv1.NamespaceAll, "Namespace to monitor (Default all)")
	flag.Set("logtostderr", "true")
	flag.Parse()

	config, err := util.BuildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := extclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create new CRD handle for the stream prediction resource type.
	crdHandle := crd.New(
		&crv1.StreamPrediction{},
		&crv1.StreamPredictionList{},
		crv1.GroupName,
		crv1.Version,
		crv1.StreamPredictionResourceKind,
		crv1.StreamPredictionResourceSingular,
		crv1.StreamPredictionResourcePlural,
		extv1beta1.NamespaceScoped,
	)

	err = crd.WriteDefinition(clientset, crdHandle)
	if err != nil {
		panic(err)
	}

	crdClient, err := crd.NewClient(*config, crdHandle)
	if err != nil {
		panic(err)
	}

	// Start a controller for instances of our custom resource.
	controller := controller.New(crdHandle, &streamPredictionHooks{crdClient}, crdClient)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go controller.Run(ctx, *namespace)

	<-ctx.Done()
}
