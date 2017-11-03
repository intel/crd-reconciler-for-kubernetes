package resource

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Client manipulates Kubernetes API resources backed by template files.
type Client interface {
	// Reify returns the raw request body given the supplied template values.
	Reify(templateValues interface{}) ([]byte, error)
	// Create creates a new object using the supplied data object for
	// template expansion.
	Create(namespace string, templateValues interface{}) error
	// Delete deletes the object.
	Delete(namespace string, name string) error
	// Get retrieves the object.
	Get(namespace, name string) (runtime.Object, error)
	// List lists objects based on group, version and kind.
	List(namespace string) ([]runtime.Object, error)
	// IsFailed returns true if this resource is in a broken state.
	IsFailed(namespace string, name string) bool
	// Plural returns the plural form of the resource.
	IsEphemeral() bool
	// Plural returns the plural form of the resource.
	Plural() string
}

// GlobalTemplateValues encodes values which will be available to all template specializations.
type GlobalTemplateValues map[string]string
