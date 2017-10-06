package main

import (
	"context"
	"flag"
	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/hooks"
	"github.com/NervanaSystems/kube-controllers-go/pkg/controller"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	"github.com/NervanaSystems/kube-controllers-go/pkg/util"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")
	namespace := flag.String("namespace", apiv1.NamespaceAll, "Namespace to monitor (Default all)")
	deploymentTemplateFile := flag.String("deploymentFile", "/etc/streampredictions/deployment.tmpl", "Path to a deployment file")
	serviceTemplateFile := flag.String("serviceFile", "/etc/streampredictions/service.tmpl", "Path to a service file")
	ingressTemplateFile := flag.String("ingressFile", "/etc/streampredictions/ingress.tmpl", "Path to an ingress file")
	hpaTemplateFile := flag.String("hpaFile", "/etc/streampredictions/hpa.tmpl", "Path to a hpa file")
	schemaFile := flag.String("schema", "", "Path to a custom resource schema file")
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

	k8sclientset, err := kubernetes.NewForConfig(config)
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
		*schemaFile,
	)

	err = crd.WriteDefinition(clientset, crdHandle)
	if err != nil {
		panic(err)
	}

	crdClient, err := crd.NewClient(*config, crdHandle)
	if err != nil {
		panic(err)

	}

	//Create hooks
	hooks := hooks.NewStreamPredictionHooks(
		crdClient,
		// TODO: Get appropriate client interfaces and plural forms from API
		//       discovery instead.
		[]resource.Client{
			resource.NewClient(k8sclientset.ExtensionsV1beta1().RESTClient(), "deployments", *deploymentTemplateFile),
			resource.NewClient(k8sclientset.CoreV1().RESTClient(), "services", *serviceTemplateFile),
			resource.NewClient(k8sclientset.ExtensionsV1beta1().RESTClient(), "ingresses", *ingressTemplateFile),
			resource.NewClient(k8sclientset.AutoscalingV1().RESTClient(), "horizontalpodautoscalers", *hpaTemplateFile),
		})

	// Start a controller for instances of our custom resource.
	controller := controller.New(crdHandle, hooks, crdClient.RESTClient())

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go controller.Run(ctx, *namespace)

	<-ctx.Done()
}
