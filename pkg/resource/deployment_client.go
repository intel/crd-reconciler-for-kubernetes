//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: EPL-2.0
//

package resource

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/intel/crd-reconciler-for-kubernetes/pkg/resource/reify"
	"github.com/intel/crd-reconciler-for-kubernetes/pkg/states"
)

type deploymentClient struct {
	globalTemplateValues GlobalTemplateValues
	k8sClientset         *kubernetes.Clientset
	restClient           rest.Interface
	resourcePluralForm   string
	templateFileName     string
}

// NewDeploymentClient returns a new generic resource client.
func NewDeploymentClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client {
	return &deploymentClient{
		globalTemplateValues: globalTemplateValues,
		k8sClientset:         clientSet,
		restClient:           clientSet.ExtensionsV1beta1().RESTClient(),
		resourcePluralForm:   "deployments",
		templateFileName:     templateFileName,
	}
}

func (c *deploymentClient) Reify(templateValues interface{}) ([]byte, error) {
	result, err := reify.Reify(c.templateFileName, templateValues, c.globalTemplateValues)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *deploymentClient) Create(namespace string, templateValues interface{}) error {
	resourceBody, err := c.Reify(templateValues)
	if err != nil {
		return err
	}

	request := c.restClient.Post().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Body(resourceBody)

	glog.Infof("[DEBUG] create resource URL: %s", request.URL())

	var statusCode int
	err = request.Do().StatusCode(&statusCode).Error()

	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code (%d)", statusCode)
	}
	return nil
}

func (c *deploymentClient) Delete(namespace, name string) error {
	// For deployments the propagation policy in delete options must be set
	// to Foreground to delete the pods along with the replica sets.
	// See https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#additional-note-on-deployments.
	deletePolicy := metav1.DeletePropagationForeground

	request := c.restClient.Delete().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name).
		Body(&metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})

	glog.Infof("[DEBUG] delete resource URL: %s", request.URL())

	return request.Do().Error()
}

func (c *deploymentClient) Update(namespace string, name string, templateValues interface{}) error {
	resourceBody, err := c.Reify(templateValues)
	if err != nil {
		return err
	}

	request := c.restClient.Put().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name).
		Body(resourceBody)

	glog.Infof("[DEBUG] update resource URL: %s", request.URL())

	var statusCode int
	err = request.Do().StatusCode(&statusCode).Error()

	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code (%d)", statusCode)
	}
	return nil
}
func (c *deploymentClient) Patch(namespace string, name string, data []byte) error {

	request := c.restClient.Patch(types.JSONPatchType).
		Resource(c.resourcePluralForm).
		Namespace(namespace).
		Name(name).
		Body(data)

	glog.Infof("[DEBUG] patch resource URL: %s", request.URL())

	return request.Do().Error()
}

func (c *deploymentClient) Get(namespace, name string) (result runtime.Object, err error) {
	result = &v1beta1.Deployment{}
	opts := metav1.GetOptions{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)

	return result, err
}

func (c *deploymentClient) List(namespace string, labels map[string]string) (result []metav1.Object, err error) {
	list := &v1beta1.DeploymentList{}
	opts := metav1.ListOptions{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(list)

	if err != nil {
		return result, err
	}

	for _, item := range list.Items {
		// We need a copy of the item here because item has function scope whereas the copy below has a local scope.
		// Ex: When we iterate through items, the result list will only contain multiple copies of the last item in the list.
		depCopy := item
		result = append(result, &depCopy)
	}

	return
}

func (c *deploymentClient) IsEphemeral() bool {
	return false
}

func (c *deploymentClient) Plural() string {
	return c.resourcePluralForm
}

func (c *deploymentClient) IsFailed(namespace string, name string) bool {
	obj, err := c.Get(namespace, name)
	if err != nil {
		return false
	}
	return c.isFailed(obj)
}

func (c *deploymentClient) isFailed(obj runtime.Object) bool {
	dep, ok := obj.(*v1beta1.Deployment)
	if !ok {
		panic("object was not a *v1beta1.Deployment")
	}
	conditions := dep.Status.Conditions
	if len(conditions) == 0 {
		return false
	}
	latestCondition := conditions[0]
	for i := range conditions {
		time1 := &latestCondition.LastUpdateTime
		time2 := &conditions[i].LastUpdateTime
		if time1.Before(time2) {
			latestCondition = conditions[i]
		}
	}

	if latestCondition.Type == v1beta1.DeploymentReplicaFailure {
		return true
	}

	// If the deployment is not in a failed state we inspect whether the
	// containers controlled by the deployment are healthy.
	// This is required because the definition of pod failure in kubernetes is
	// strict. The pod is considered failed iff all containers in the pod have
	// terminated, and at least one container has terminated in a failure (exited
	// with a non-zero exit code or was stopped by the system). If the pod gets to
	// a failed state the controlling object (e.g., deployment),
	// enters a failed state (i.e., DeploymentReplicaFailure state in case of
	// deployment) as well.
	podClient := NewPodClient(GlobalTemplateValues{}, c.k8sClientset, "")

	// List all the pods with the same labels as the deployment and check if
	// they have failed.
	podList, err := podClient.List(dep.ObjectMeta.Namespace, dep.ObjectMeta.Labels)
	if err != nil {
		return false
	}

	for _, pod := range podList {
		if podClient.IsFailed(pod.GetNamespace(), pod.GetName()) {
			return true
		}
	}

	return false
}

func (c *deploymentClient) GetStatusState(obj runtime.Object) states.State {
	if c.isFailed(obj) {
		return states.Failed
	}
	// TODO(CD): Detect Pending state. Completed doesn't make sense for this type.
	return states.Running
}
