package crd

import (
	"errors"
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/xeipuuv/gojsonschema"
)

const apiRoot = "/apis"

// Client is used to handle CRD operations.
type Client interface {
	Create(crd CustomResource) error
	Get(namespace string, name string) (runtime.Object, error)
	Update(crd CustomResource) (runtime.Object, error)
	Delete(namespace string, name string) error
	Validate(crd CustomResource) error
	RESTClient() rest.Interface
}

type client struct {
	restClient rest.Interface
	handle     *Handle
}

// NewClient returns a new REST client wrapper for the supplied CRD handle.
func NewClient(config rest.Config, h *Handle) (Client, error) {
	// TODO(balajismaninam): move scheme building to register.go in crv1.
	// We can enable metav1.GetOptions and metav1.ListOptions after that.
	scheme := runtime.NewScheme()

	scheme.AddKnownTypes(h.SchemaGroupVersion, h.ResourceType, h.ResourceListType)
	metav1.AddToGroupVersion(scheme, h.SchemaGroupVersion)

	config.GroupVersion = &h.SchemaGroupVersion
	config.APIPath = apiRoot
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	restClient, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &client{restClient, h}, nil
}

func (c *client) RESTClient() rest.Interface {
	return c.restClient
}

// Create creates the supplied CRD.
func (c *client) Create(crd CustomResource) error {
	if c.handle.SchemaURL != "" {
		if err := c.Validate(crd); err != nil {
			return err
		}
	}

	return c.restClient.Post().
		Namespace(crd.Namespace()).
		Resource(c.handle.Plural).
		Name(crd.Name()).
		Body(crd).
		Do().
		Error()
}

// Get retrieves the CRD from the Kubernetes API server.
func (c *client) Get(namespace string, name string) (runtime.Object, error) {
	// TODO(balajismaniam): Move scheme building to register.go in crv1 and
	// enable the usage of metav1.GetOptions{}.
	result := c.handle.ResourceType.DeepCopyObject()
	err := c.restClient.Get().
		Namespace(namespace).
		Resource(c.handle.Plural).
		Name(name).
		Do().
		Into(result)

	return result, err
}

// Update updates the CRD on the Kubernetes API server.
func (c *client) Update(crd CustomResource) (runtime.Object, error) {
	if c.handle.SchemaURL != "" {
		if err := c.Validate(crd); err != nil {
			return nil, err
		}
	}

	resp := c.restClient.Put().
		Namespace(crd.Namespace()).
		Resource(c.handle.Plural).
		Name(crd.Name()).
		Body(crd).
		Do()

	obj, err := resp.Get()
	if err != nil {
		return nil, err
	}

	return obj, resp.Error()
}

// Delete deletes the CRD from the Kubernetes API server.
func (c *client) Delete(namespace string, name string) error {
	return c.restClient.Delete().
		Namespace(namespace).
		Resource(c.handle.Plural).
		Name(name).
		Do().
		Error()
}

// Validate validates a custom resource against a json schema.
// Returns nil if object adheres to the schema.
func (c *client) Validate(cr CustomResource) error {
	if c.handle.SchemaURL == "" {
		return fmt.Errorf("Validate called without schema URL set")
	}

	schemaLoader := gojsonschema.NewReferenceLoader(c.handle.SchemaURL)

	json, err := cr.JSON()
	if err != nil {
		return err
	}

	documentLoader := gojsonschema.NewStringLoader(json)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errorOutput := "Invalid JSON: '" + json + "': "

		for _, desc := range result.Errors() {
			errorOutput = errorOutput + " - " + desc.String() + "\n"
		}

		return errors.New(errorOutput)
	}

	return nil
}

// WriteDefinition writes the supplied CRD to the Kubernetes API server
// using the supplied client set.
func WriteDefinition(clientset apiextensionsclient.Interface, h *Handle) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(h.Definition)
	if err != nil {
		return err
	}

	var crd *apiextensionsv1beta1.CustomResourceDefinition
	// Wait for CRD to be established.
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(h.resourceName(), metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := DeleteDefinition(clientset, h)
		if deleteErr != nil {
			return k8serrors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}

	// Update the definition in the supplied handle.
	h.Definition = crd

	return nil
}

// DeleteDefinition removes the supplied CRD to the Kubernetes API server
// using the supplied client set.
func DeleteDefinition(clientset apiextensionsclient.Interface, h *Handle) error {
	return clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(h.Definition.Name, nil)
}
