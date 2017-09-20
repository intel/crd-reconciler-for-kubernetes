package v1

import "k8s.io/apimachinery/pkg/labels"

import (
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// StreamPredictionListerExpansion allows custom methods to be added to
// StreamPredictionLister.
type StreamPredictionListerExpansion interface{}

// StreamPredictionLister helps list StreamPredictSion.
type StreamPredictionLister interface {
	// List lists all CustomResourceDefinitions in the indexer.
	List(selector labels.Selector) (ret []*StreamPrediction, err error)
	// Get retrieves the CustomResourceDefinition from the index for a given name.
	Get(name string) (*StreamPrediction, error)
	StreamPredictionListerExpansion
}

// streamPredictionLister implements the StreamPredictionLister interface.
type streamPredictionLister struct {
	indexer cache.Indexer
}

// NewStreamPredictionLister returns a new StreamPredictionLister.
func NewStreamPredictionLister(indexer cache.Indexer) StreamPredictionLister {
	return &streamPredictionLister{indexer: indexer}
}

// List lists all StreamPredictions in the indexer.
func (s *streamPredictionLister) List(selector labels.Selector) (ret []*StreamPrediction, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*StreamPrediction))
	})
	return ret, err
}

// Get retrieves the StreamPrediction from the index for a given name.
func (s *streamPredictionLister) Get(name string) (*StreamPrediction, error) {
	key := &StreamPrediction{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	obj, exists, err := s.indexer.Get(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apierrors.NewNotFound(v1beta1.Resource("customresourcedefinition"), name)
	}
	return obj.(*StreamPrediction), nil
}
