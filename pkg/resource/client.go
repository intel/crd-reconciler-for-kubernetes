package resource

import (
	"fmt"
	"net/http"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/resource/reify"
	"github.com/golang/glog"
)

// Client manipulates Kubernetes API resources backed by template files.
type Client interface {
	// Create creates a new object using the supplied data object for
	// template expansion.
	Create(namespace string, templateData interface{}) error
	// Delete deletes the object.
	Delete(namespace string, name string) error
	// Get retrieves the object.
	Get(namespace, name string) (runtime.Object, error)
	// List lists objects based on group, version and kind.
	List(namespace string) (runtime.Object, error)
	// Plural returns the plural form of the resource.
	Plural() string
}

type client struct {
	restClient rest.Interface
	// TODO(CD): Try to get this automatically from the template contents.
	resourcePluralForm string
	templateFileName   string
}

// NewClient returns a new resource client.
func NewClient(restClient rest.Interface, resourcePluralForm string,
	templateFileName string) Client {

	return &client{
		restClient:         restClient,
		resourcePluralForm: resourcePluralForm,
		templateFileName:   templateFileName,
	}
}

func (c *client) Create(namespace string, templateData interface{}) error {
	resourceBody, err := reify.Reify(c.templateFileName, templateData)
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

func (c *client) Delete(namespace, name string) error {
	request := c.restClient.Delete().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name)

	glog.Infof("[DEBUG] delete resource URL: %s", request.URL())

	return request.Do().Error()
}

func (c *client) Get(namespace, name string) (result runtime.Object, err error) {
	switch c.resourcePluralForm {
	case "deployments":
		result = &v1beta1.Deployment{}
	case "services":
		result = &corev1.Service{}
	case "ingresses":
		result = &v1beta1.Ingress{}
	case "horizontalpodautoscalers":
		result = &autoscalingv1.HorizontalPodAutoscaler{}
	default:
		errMsg := fmt.Sprintf("unexpected resource client type (plural: %v)", c.resourcePluralForm)
		glog.Errorf(errMsg)
		return result, fmt.Errorf(errMsg)
	}

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

func (c *client) List(namespace string) (result runtime.Object, err error) {
	switch c.resourcePluralForm {
	case "deployments":
		result = &v1beta1.DeploymentList{}
	case "services":
		result = &corev1.ServiceList{}
	case "ingresses":
		result = &v1beta1.IngressList{}
	case "horizontalpodautoscalers":
		result = &autoscalingv1.HorizontalPodAutoscalerList{}
	default:
		errMsg := fmt.Sprintf("unexpected resource client list type (plural: %v)", c.resourcePluralForm)
		glog.Errorf(errMsg)
		return result, fmt.Errorf(errMsg)
	}

	opts := metav1.ListOptions{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)

	return result, err
}

func (c *client) Plural() string {
	return c.resourcePluralForm
}
