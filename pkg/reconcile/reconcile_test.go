package reconcile

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"fmt"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	rf "github.com/NervanaSystems/kube-controllers-go/pkg/resource/fake"
	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

// covert an array to a map for comparison purposes
// TODO have Reconciler return a deterministic slice array
func sliceToMap(subresources []*subresource) map[runtime.Object]*subresource {
	subMap := make(map[runtime.Object]*subresource, len(subresources))
	for _, subresource := range subresources {
		obj := subresource.object
		subMap[obj] = subresource
	}
	return subMap
}

// compare actual with expected
func compareSubresourceMaps(expected subresourceMap, actual subresourceMap) func() bool {
	compare := func() bool {
		if len(expected) == 0 && len(actual) == 0 {
			return true
		}
		// this checks the inverse subset relation between expected and actual
		if len(actual) != len(expected) {
			return false
		}
		for controllerName, expectedSubresources := range expected {
			actualSubresources := actual[controllerName]
			actualMap := sliceToMap(actualSubresources)
			expectedMap := sliceToMap(expectedSubresources)
			for key, expectedSubresource := range expectedMap {
				actualSubresource := actualMap[key]
				assert.ObjectsAreEqualValues(expectedSubresource, actualSubresource)
			}

		}
		return true
	}
	return compare
}

func TestGroupSubresourcesByCustomResource(t *testing.T) {
	controllerRef := true
	tests := map[string]struct {
		namespace                   string
		gvk                         schema.GroupVersionKind
		resourceClients             []resource.Client
		expectedEmptysubResourceMap bool
	}{
		"no subresources under this CR": {
			namespace: "namespace1",
			gvk: schema.GroupVersionKind{
				Group:   "aipg.intel.com",
				Version: "v1",
				Kind:    "InteractiveSession",
			},
			resourceClients: []resource.Client{
				&rf.SubresourceClient{
					Error:       "no subresources",
					PluralValue: "interactivesessions",
				},
			},
			expectedEmptysubResourceMap: true,
		},
		"no controller for a subresource": {
			namespace: "namespace2",
			gvk: schema.GroupVersionKind{
				Group:   "aipg.intel.com",
				Version: "v1",
				Kind:    "StreamPrediction",
			},
			resourceClients: []resource.Client{
				&rf.SubresourceClient{
					Subresource: &rf.Subresource{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "service1",
							Namespace: "namespace2",
						},
						SpecState:   states.Running,
						StatusState: states.Running,
						Ephemeral:   true,
						Plural:      "Services",
						Lifecycle:   fmt.Sprintf("%v", exists),
					},
				},
			},
			expectedEmptysubResourceMap: true,
		},
		"wrong controller for a subresource": {
			namespace: "namespace3",
			gvk: schema.GroupVersionKind{
				Group:   "aipg.intel.com",
				Version: "v1",
				Kind:    "ModelTraining",
			},
			resourceClients: []resource.Client{
				&rf.SubresourceClient{
					Subresource: &rf.Subresource{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "job1",
							Namespace: "namespace3",
							OwnerReferences: []metav1.OwnerReference{
								{
									APIVersion: "aipg.intel.com/v1",
									Kind:       "StreamPrediction",
									Name:       "stream1",
									UID:        "8888",
									Controller: &controllerRef,
								}},
						},
						SpecState:   states.Running,
						StatusState: states.Running,
						Ephemeral:   true,
						Plural:      "Deployments",
						Lifecycle:   fmt.Sprintf("%v", exists),
					},
				},
			},
			expectedEmptysubResourceMap: true,
		},
		"valid controller ref, but not a valid runtime.Object": {
			namespace: "namespace4",
			gvk: schema.GroupVersionKind{
				Group:   "aipg.intel.com",
				Version: "v1",
				Kind:    "StreamPrediction",
			},
			resourceClients: []resource.Client{
				&rf.SubresourceClient{
					Subresource: &metav1.ObjectMeta{
						Name:      "job1",
						Namespace: "namespace4",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "aipg.intel.com/v1",
								Kind:       "StreamPrediction",
								Name:       "stream1",
								UID:        "8888",
								Controller: &controllerRef,
							}},
					},
				},
			},
			expectedEmptysubResourceMap: true,
		},
		"valid controller ref": {
			namespace: "namespace5",
			gvk: schema.GroupVersionKind{
				Group:   "aipg.intel.com",
				Version: "v1",
				Kind:    "StreamPrediction",
			},
			resourceClients: []resource.Client{
				&rf.SubresourceClient{
					Subresource: &rf.Subresource{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Ingress",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ingress1",
							Namespace: "namespace5",
							OwnerReferences: []metav1.OwnerReference{
								{
									APIVersion:         "aipg.intel.com/v1",
									Kind:               "StreamPrediction",
									Name:               "stream2",
									UID:                "3982",
									Controller:         &controllerRef,
									BlockOwnerDeletion: nil,
								}},
						},
						SpecState:   states.Running,
						StatusState: states.Running,
						Ephemeral:   true,
						Plural:      "Ingresses",
						Lifecycle:   fmt.Sprintf("%v", exists),
					},
				},
			},
			expectedEmptysubResourceMap: false,
		},
	}

	for _, tc := range tests {
		reconciler := &Reconciler{
			namespace:       tc.namespace,
			gvk:             tc.gvk,
			crdHandle:       nil,
			crdClient:       nil,
			resourceClients: tc.resourceClients,
		}
		actual := reconciler.groupSubresourcesByCustomResource()
		if tc.expectedEmptysubResourceMap {
			assert.Empty(t, actual)
		} else {
			for controllerName := range actual {
				expected := subresourceMap{}
				// TODO ACCEN-207 reconciler should work with no subresources
				if tc.resourceClients[0] != nil {
					subResourceClient := tc.resourceClients[0].(*rf.SubresourceClient)
					subResource := subResourceClient.Subresource.(*rf.Subresource)
					sub := &subresource{
						client:    subResourceClient,
						object:    subResourceClient.Subresource.(runtime.Object),
						lifecycle: lifecycle(subResource.Lifecycle),
					}
					expected[controllerName] = subresources{sub}
				}
				assert.Condition(t, compareSubresourceMaps(expected, actual))
			}
		}
	}

}
