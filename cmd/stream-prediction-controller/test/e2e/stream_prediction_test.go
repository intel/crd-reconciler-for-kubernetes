package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/util"
)

func TestStreamPrediction(t *testing.T) {

	namespace := "default"
	config, err := util.BuildConfig("/root/.kube/config")
	assert.Nil(t, err)

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

	crdClient, err := crd.NewClient(*config, crdHandle)
	assert.Nil(t, err)

	streamName := "examplestreampredict"
	foo := "test"
	bar := true

	streamPredict := &crv1.StreamPrediction{
		ObjectMeta: metav1.ObjectMeta{
			Name: streamName,
		},
		Spec: crv1.StreamPredictionSpec{
			Foo: foo,
			Bar: bar,
		},
		Status: crv1.StreamPredictionStatus{
			State:   crv1.StreamPredictionCreated,
			Message: "Created, not processed",
		},
	}

	var result crv1.StreamPrediction
	err = crdClient.Post().
		Resource(crv1.StreamPredictionResourcePlural).
		Namespace(namespace).
		Body(streamPredict).
		Do().
		Into(&result)

	if err == nil {
		t.Logf("Created stream prediction: %#v\n", result)
	} else if apierrors.IsAlreadyExists(err) {
		t.Errorf("Stream prediction already exists: %#v\n", result)
	} else {
		t.Fatal(err)
	}

	// Check if the crd got created
	var streamPrediction crv1.StreamPrediction
	err = crdClient.Get().
		Resource(crv1.StreamPredictionResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name(streamName).
		Do().Into(&streamPrediction)
	assert.Nil(t, err)

	testSpec(streamPrediction, foo, t, bar)

	// Wait for the stream predict crd to get created and being processed
	err = waitForStreamPredictionInstanceProcessed(crdClient, streamName)
	assert.Nil(t, err)

	t.Logf("Processed crd: %s", streamName)
	streamPredictList := crv1.StreamPredictionList{}
	err = crdClient.Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	t.Logf("List: %v\n", streamPredictList)
	assert.Equal(t, 1, len(streamPredictList.Items), "List does not contain correct number of stream predictions.")

	testSpec(streamPredictList.Items[0], foo, t, bar)

	err = crdClient.Delete().Resource(crv1.StreamPredictionResourcePlural).Namespace(namespace).Name(streamName).Do().Error()
	assert.Nil(t, err)

	streamPredictList = crv1.StreamPredictionList{}
	err = crdClient.Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	t.Logf("List: %v\n", streamPredictList)
	assert.Equal(t, 0, len(streamPredictList.Items), "List does not contain correct number of stream predictions.")
}

func testSpec(streamPrediction crv1.StreamPrediction, foo string, t *testing.T, bar bool) {
	// Check if all the fields are right
	assert.Equal(t, foo, streamPrediction.Spec.Foo, "foo not %s", foo)
	assert.Equal(t, bar, streamPrediction.Spec.Bar, "bar not %s", bar)
}

// waitForStreamPredictionInstanceProcessed waits for the stream prediction to be processed.
func waitForStreamPredictionInstanceProcessed(streamPredictionClient *rest.RESTClient, name string) error {
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
