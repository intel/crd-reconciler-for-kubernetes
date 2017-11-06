package resource

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/resource/reify"
)

type deploymentClient struct {
	globalTemplateValues GlobalTemplateValues
	restClient           rest.Interface
	resourcePluralForm   string
	templateFileName     string
}

// NewDeploymentClient returns a new generic resource client.
func NewDeploymentClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client {
	return &deploymentClient{
		globalTemplateValues: globalTemplateValues,
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

func (c *deploymentClient) List(namespace string) (result []metav1.Object, err error) {
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
		result = append(result, &item)
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
	d, err := c.Get(namespace, name)
	if err != nil {
		return false
	}
	dep, ok := d.(*v1beta1.Deployment)
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

	return latestCondition.Type == v1beta1.DeploymentReplicaFailure
}
