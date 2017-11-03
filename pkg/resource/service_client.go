package resource

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/resource/reify"
)

type serviceClient struct {
	globalTemplateValues GlobalTemplateValues
	restClient           rest.Interface
	resourcePluralForm   string
	templateFileName     string
}

// NewServiceClient returns a new service client.
func NewServiceClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client {
	return &serviceClient{
		globalTemplateValues: globalTemplateValues,
		restClient:           clientSet.CoreV1().RESTClient(),
		resourcePluralForm:   "services",
		templateFileName:     templateFileName,
	}
}

func (c *serviceClient) Reify(templateValues interface{}) ([]byte, error) {
	result, err := reify.Reify(c.templateFileName, templateValues, c.globalTemplateValues)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *serviceClient) Create(namespace string, templateValues interface{}) error {
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

func (c *serviceClient) Delete(namespace, name string) error {
	request := c.restClient.Delete().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name)

	glog.Infof("[DEBUG] delete resource URL: %s", request.URL())

	return request.Do().Error()
}

func (c *serviceClient) Get(namespace, name string) (result runtime.Object, err error) {
	result = &corev1.Service{}
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

func (c *serviceClient) List(namespace string) (result runtime.Object, err error) {
	result = &corev1.ServiceList{}
	opts := metav1.ListOptions{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)

	return result, err
}

func (c *serviceClient) IsEphemeral() bool {
	return true
}

func (c *serviceClient) Plural() string {
	return c.resourcePluralForm
}
