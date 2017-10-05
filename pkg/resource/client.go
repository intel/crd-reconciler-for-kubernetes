package resource

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/resource/reify"
)

// Client manipulates Kubernetes API resources backed by template files.
type Client interface {
	// Create creates a new object using the supplied data object for
	// template expansion.
	Create(namespace string, templateData interface{}) error
	// Delete deletes the object
	Delete(namespace string, name string) error
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
