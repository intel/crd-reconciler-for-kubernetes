package resource

import (
	"fmt"
	"net/http"

	v1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// Delete deletes the object
	Delete(namespace string, name string) error
	// List lists objects based on group, version and kind.
	List(namespace string, gvk schema.GroupVersionKind) (result *v1beta1.DeploymentList, err error)
}

type client struct {
	restClient rest.Interface
	// TODO(CD): Try to get this automatically from the template contents.
	resourcePluralForm string
	templateFileName   string
}

// NewClient returns a new resource client.
func NewClient(restClient rest.Interface, resourcePluralForm string, templateFileName string) Client {
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

func (c *client) Delete(namespace string, name string) error {
	request := c.restClient.Delete().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		Name(name)

	glog.Infof("[DEBUG] delete resource URL: %s", request.URL())

	return request.Do().Error()
}

func (c *client) List(namespace string, gvk schema.GroupVersionKind) (result *v1beta1.DeploymentList, err error) {
	selector := fields.Set{
		"metadata.ownerReferences[0].apiVersion": gvk.GroupVersion().String(),
		"metadata.ownerReferences[0].kind":       gvk.Kind,
	}.AsSelector().String()
	opts := metav1.ListOptions{FieldSelector: selector}
	result = &v1beta1.DeploymentList{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource(c.resourcePluralForm).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)

	return
}
