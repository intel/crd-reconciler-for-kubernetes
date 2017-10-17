package crd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
)

// CustomResource is the base type of custom resource objects.
// This allows them to be manipulated generically by the CRD client.
type CustomResource interface {
	Name() string
	Namespace() string
	JSON() (string, error)
	GetStatusState() states.State
	SetStatusStateWithMessage(states.State, string)
	GetErrorState() states.State
	DeepCopyObject() runtime.Object
	GetObjectKind() schema.ObjectKind
}
