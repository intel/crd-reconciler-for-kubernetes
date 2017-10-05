package hooks

import (
	"fmt"
	"testing"

	crv1 "github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller/apis/cr/v1"
	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	"github.com/NervanaSystems/kube-controllers-go/pkg/util"
	"github.com/stretchr/testify/assert"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testResourceClient struct {
	createCalled   bool
	createWillFail bool

	deleteCalled bool
}

func (trc *testResourceClient) Create(namespace string, templateData interface{}) error {
	trc.createCalled = true

	if trc.createWillFail {
		return fmt.Errorf("Resource client creation failed on purpose")
	}

	return nil
}

func (trc *testResourceClient) Delete(namespace string, name string) error {
	trc.deleteCalled = true
	return nil
}

func TestStreampredictionHooks(t *testing.T) {
	config, err := util.BuildConfig("/go/src/github.com/NervanaSystems/kube-controllers-go/resources/config")
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
	if err != nil {
		panic(err)
	}

	sp := &crv1.StreamPrediction{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stream-foobar",
		},
		Status: crv1.StreamPredictionStatus{
			State:   crv1.StreamPredictionCreated,
			Message: "Created, not processed",
		},
	}

	//
	// First test, make sure the success case pass.
	// Both resources should be created and delete should not be called.
	//
	foo := &testResourceClient{createWillFail: false}
	bar := &testResourceClient{createWillFail: false}

	hooks := NewStreamPredictionHooks(crdClient, []resource.Client{foo, bar})

	hooks.Add(sp)

	assert.True(t, foo.createCalled)
	assert.True(t, bar.createCalled)
	assert.False(t, foo.deleteCalled)
	assert.False(t, bar.deleteCalled)

	//
	// Second test, make sure that if the first resource fails, no other resources
	// wil be created. Both resources should be attempted to be deleted.
	//
	foo = &testResourceClient{createWillFail: true}
	bar = &testResourceClient{createWillFail: false}

	hooks = NewStreamPredictionHooks(crdClient, []resource.Client{foo, bar})

	hooks.Add(sp)

	//
	// Third test, if the second resource fails. The first one should have been
	// attempted to be created. Both resources should be deleted.
	//
	assert.True(t, foo.createCalled)
	assert.False(t, bar.createCalled)
	assert.True(t, foo.deleteCalled)
	assert.True(t, bar.deleteCalled)

	foo = &testResourceClient{createWillFail: false}
	bar = &testResourceClient{createWillFail: true}

	hooks = NewStreamPredictionHooks(crdClient, []resource.Client{foo, bar})

	hooks.Add(sp)

	assert.True(t, foo.createCalled)
	assert.True(t, bar.createCalled)
	assert.True(t, foo.deleteCalled)
	assert.True(t, bar.deleteCalled)
}
