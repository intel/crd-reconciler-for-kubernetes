package test

import (
	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	streampredictionclient "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/client"
	util "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/util"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestStreamPrediction(t *testing.T) {

	namespace := "default"
	config, err := util.BuildConfig("/root/.kube/config")
	assert.Nil(t, err)

	client, _, err := streampredictionclient.NewClient(config)
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
	err = client.Post().
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
	err = client.Get().
		Resource(crv1.StreamPredictionResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name(streamName).
		Do().Into(&streamPrediction)
	assert.Nil(t, err)

	testSpec(streamPrediction, foo, t, bar)

	// Wait for the stream predict crd to get created and being processed
	err = util.WaitForStreamPredictionInstanceProcessed(client, streamName)
	assert.Nil(t, err)

	t.Logf("Processed crd: %s", streamName)
	streamPredictList := crv1.StreamPredictionList{}
	err = client.Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	t.Logf("List: %v\n", streamPredictList)
	assert.Equal(t, 1, len(streamPredictList.Items), "List does not contain correct number of stream predictions.")

	testSpec(streamPredictList.Items[0], foo, t, bar)

	err = client.Delete().Resource(crv1.StreamPredictionResourcePlural).Namespace(namespace).Name(streamName).Do().Error()
	assert.Nil(t, err)

	streamPredictList = crv1.StreamPredictionList{}
	err = client.Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	t.Logf("List: %v\n", streamPredictList)
	assert.Equal(t, 0, len(streamPredictList.Items), "List does not contain correct number of stream predictions.")
}

func testSpec(streamPrediction crv1.StreamPrediction, foo string, t *testing.T, bar bool) {
	// Check if all the fields are right
	assert.Equal(t, foo, streamPrediction.Spec.Foo, "foo not %s", foo)
	assert.Equal(t, bar, streamPrediction.Spec.Bar, "bar not %s", bar)
}
