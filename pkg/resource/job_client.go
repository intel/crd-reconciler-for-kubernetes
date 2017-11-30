package resource

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/NervanaSystems/kube-controllers-go/pkg/resource/reify"
)

type jobClient struct {
	globalTemplateValues GlobalTemplateValues
	restClient           rest.Interface
	resourcePluralForm   string
	templateFileName     string
}

// NewJobClient returns a new generic resource client.
func NewJobClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client {
	return &jobClient{
		globalTemplateValues: globalTemplateValues,
		restClient:           clientSet.BatchV1().RESTClient(),
		resourcePluralForm:   "jobs",
		templateFileName:     templateFileName,
	}
}

func (c *jobClient) Reify(templateValues interface{}) ([]byte, error) {
	result, err := reify.Reify(c.templateFileName, templateValues, c.globalTemplateValues)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *jobClient) Create(namespace string, templateValues interface{}) error {
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

func (c *jobClient) Delete(namespace, name string) error {
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

func (c *jobClient) Get(namespace, name string) (result runtime.Object, err error) {
	result = &batchv1.Job{}
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

func (c *jobClient) List(namespace string, labels map[string]string) (result []metav1.Object, err error) {
	list := &batchv1.JobList{}
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

func (c *jobClient) Plural() string {
	return c.resourcePluralForm
}

func (c *jobClient) IsFailed(namespace string, name string) bool {
	return false
}

func (c *jobClient) IsEphemeral() bool {
	return false
}
