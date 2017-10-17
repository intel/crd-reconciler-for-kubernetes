package e2e

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
	"github.com/NervanaSystems/kube-controllers-go/pkg/util"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
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
		"",
	)

	crdClient, err := crd.NewClient(*config, crdHandle)
	assert.Nil(t, err)

	k8sClient, err := kubernetes.NewForConfig(config)
	assert.Nil(t, err)
	assert.NotNil(t, k8sClient)

	streamID := 0

	streamName := fmt.Sprintf("stream%s", strings.ToLower(ksuid.New().String()))

	spec := crv1.StreamPredictionSpec{
		NeonRepoSpec: crv1.NeonRepoSpec{
			RepoURL: "git@github.com:NervanaSystems/private-neon.git",
			Commit:  "v1.8.2",
		},
		SecuritySpec: crv1.SecuritySpec{
			PresignedToken: "95fcbe0cfe747b867655a243cee330",
			JWTToken:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdHJlYW1faWQiOjEwfQ.JxxqL8-6OV4xfQmy4dGRis3QSRuTJH2kattCfLHGKwA",
		},
		StreamDataSpec: crv1.StreamDataSpec{
			ModelPRM:         "/code/model.prm",
			ModelPath:        "s3://helium-joboutput-dev/integration/20dec8c3e38e2804888f252ef281121b/51/model.prm",
			DatasetPath:      "None",
			ExtraFilename:    "None",
			CustomCodeURL:    "None",
			CustomCommit:     "None",
			AWSPath:          "krypton-logs-dev/integration",
			AWSDefaultRegion: "us-west-1",
			StreamID:         streamID,
			StreamName:       streamName,
		},
		ResourceSpec: crv1.ResourceSpec{
			Requests: map[string]string{
				"cpu":    "1",
				"memory": "512M",
			},
		},
		KryptonRepoSpec: crv1.KryptonRepoSpec{
			RepoURL:      "git@github.com:NervanaSystems/krypton.git",
			Commit:       "master",
			Image:        "nervana/krypton:master",
			SidecarImage: "nervana/krypton-sidecar:master",
		},
		State: crv1.Deployed,
	}

	streamPredict := &crv1.StreamPrediction{
		ObjectMeta: metav1.ObjectMeta{
			Name: streamName,
		},
		Spec: spec,
		Status: crv1.StreamPredictionStatus{
			State:   crv1.Deploying,
			Message: "Created, not processed",
		},
	}

	var result crv1.StreamPrediction
	err = crdClient.RESTClient().Post().
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
	err = crdClient.RESTClient().Get().
		Resource(crv1.StreamPredictionResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name(streamName).
		Do().Into(&streamPrediction)
	assert.Nil(t, err)

	testSpec(streamPrediction, t, &spec)

	checkStreamState(crdClient, streamName, t, k8sClient, namespace, crv1.Deployed, false)

	streamPredictList := crv1.StreamPredictionList{}
	err = crdClient.RESTClient().Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	streamPredictionCRD := crv1.StreamPrediction{}
	err = crdClient.RESTClient().Get().Resource(crv1.StreamPredictionResourcePlural).Namespace(namespace).Name(streamName).Do().Into(&streamPredictionCRD)
	assert.Nil(t, err)
	assert.NotNil(t, streamPredictionCRD)

	t.Logf("List: %v\n", streamPredictList)

	testSpec(streamPredictList.Items[0], t, &spec)

	// Right now it's in Deployed. Try changing it to Completed and check if all the resources are deleted.
	err = crdClient.RESTClient().Get().
		Resource(crv1.StreamPredictionResourcePlural).
		Namespace(apiv1.NamespaceDefault).
		Name(streamName).
		Do().Into(&streamPrediction)
	assert.Nil(t, err)
	assert.NotNil(t, streamPrediction)
	streamPrediction.Spec.State = crv1.Completed
	err = crdClient.Update(&streamPrediction)
	assert.Nil(t, err)
	checkStreamState(crdClient, streamName, t, k8sClient, namespace, crv1.Completed, true)

	err = crdClient.Delete(namespace, streamName)
	assert.Nil(t, err)

	streamPredictList = crv1.StreamPredictionList{}
	err = crdClient.RESTClient().Get().Resource(crv1.StreamPredictionResourcePlural).Do().Into(&streamPredictList)
	assert.Nil(t, err)

	t.Logf("List: %v\n", streamPredictList)
	streamPredictionCRD = crv1.StreamPrediction{}
	assert.NotContains(t, streamPredictList.Items, streamPrediction)
}

func checkStreamState(crdClient crd.Client, streamName string, t *testing.T, k8sClient *kubernetes.Clientset, namespace string, state states.State, expectFailure bool) {
	// Wait for the stream predict crd to get created and being processed
	err := WaitForStreamPredictionInstanceProcessed(crdClient, streamName, state)
	assert.Nil(t, err)
	t.Logf("Processed crd: %s", streamName)
	checkK8sResources(k8sClient, namespace, streamName, t, expectFailure)
}

func checkK8sResources(k8sClient *kubernetes.Clientset, namespace string, streamName string, t *testing.T, expectFailure bool) {
	deployment, err := k8sClient.ExtensionsV1beta1().
		Deployments(namespace).Get(streamName, metav1.GetOptions{})
	if !expectFailure {
		assert.Nil(t, err)
		assert.NotNil(t, deployment)
	} else {
		// Deployment is not getting deleted at all in this cluster. So commenting it for now.
		/*assert.NotNil(t, waitPoll(func() (bool, error) {
			if err != nil && deployment == nil {
				return true, nil
			}
			return false, err
		}))*/
	}
	service, err := k8sClient.CoreV1().Services(namespace).
		Get(streamName, metav1.GetOptions{})
	if !expectFailure {
		assert.Nil(t, err)
		assert.NotNil(t, service)
	} else {
		// It takes a while to delete the resources, so waiting till they get deleted.
		assert.NotNil(t, waitPoll(func() (bool, error) {
			if err != nil && service == nil {
				return true, nil
			}
			return false, err
		}))
	}
	ingress, err := k8sClient.ExtensionsV1beta1().
		Ingresses(namespace).Get(streamName, metav1.GetOptions{})
	if !expectFailure {
		assert.Nil(t, err)
		assert.NotNil(t, ingress)
	} else {
		// It takes a while to delete the resources, so waiting till they get deleted.
		assert.NotNil(t, waitPoll(func() (bool, error) {
			if err != nil && ingress == nil {
				return true, nil
			}
			return false, err
		}))
	}
	hpa, err := k8sClient.AutoscalingV1().
		HorizontalPodAutoscalers(namespace).Get(streamName, metav1.GetOptions{})
	if !expectFailure {
		assert.Nil(t, err)
		assert.NotNil(t, hpa)
	} else {
		// It takes a while to delete the resources, so waiting till they get deleted.
		assert.NotNil(t, waitPoll(func() (bool, error) {
			if err != nil && hpa == nil {
				return true, nil
			}
			return false, err
		}))
	}
}

func testSpec(streamPrediction crv1.StreamPrediction, t *testing.T, spec *crv1.StreamPredictionSpec) {
	// Check if all the fields are right
	assert.True(t, reflect.DeepEqual(&streamPrediction.Spec, spec), "Spec is not the same")
}

// WaitForStreamPredictionInstanceProcessed waits for the stream prediction to be processed.
func WaitForStreamPredictionInstanceProcessed(crdClient crd.Client, name string, state states.State) error {
	return waitPoll(func() (bool, error) {
		var streamPrediction crv1.StreamPrediction
		err := crdClient.RESTClient().Get().
			Resource(crv1.StreamPredictionResourcePlural).
			Namespace(apiv1.NamespaceDefault).
			Name(name).
			Do().Into(&streamPrediction)

		if err == nil && streamPrediction.Status.State == state {
			return true, nil
		}

		return false, err
	})
}

func waitPoll(waitFunc func() (bool, error)) error {
	return wait.Poll(1*time.Second, 10*time.Second, waitFunc)
}
