package crd

// CustomResource is the base type of custom resource objects.
// This allows them to be manipulated generically by the CRD client.
type CustomResource interface {
	Name() string
	Namespace() string
	JSON() (string, error)
}
