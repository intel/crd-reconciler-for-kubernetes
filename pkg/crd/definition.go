package crd

import (
	"fmt"

	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Handle aggregates a CRD definition with additional data for
// client side (de)serialization.
type Handle struct {
	SchemaGroupVersion schema.GroupVersion
	Definition         *extv1beta1.CustomResourceDefinition
	ResourceType       runtime.Object
	ResourceListType   runtime.Object
}

// New returns a new CRD Handle.
func New(
	resourceType runtime.Object,
	resourceListType runtime.Object,
	group string,
	version string,
	kind string,
	singular string,
	plural string,
	scope extv1beta1.ResourceScope,
) *Handle {
	definition := &extv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", plural, group),
		},
		Spec: extv1beta1.CustomResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Scope:   scope,
			Names: extv1beta1.CustomResourceDefinitionNames{
				Kind:     kind,
				Singular: singular,
				Plural:   plural,
			},
		},
	}

	return &Handle{
		SchemaGroupVersion: schema.GroupVersion{Group: group, Version: version},
		Definition:         definition,
		ResourceType:       resourceType,
		ResourceListType:   resourceListType,
	}
}

func (h *Handle) resourceName() string {
	return fmt.Sprintf("%s.%s", h.Definition.Spec.Names.Plural, h.Definition.Spec.Group)
}
