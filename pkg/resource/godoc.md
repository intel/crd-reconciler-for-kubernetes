# resource
--
    import "github.com/NervanaSystems/kube-controllers-go/pkg/resource"


## Usage

#### type Client

```go
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
	List(namespace string, labels map[string]string) ([]metav1.Object, error)
	// IsFailed returns true if this resource is in a broken state.
	IsFailed(namespace string, name string) bool
	// Plural returns the plural form of the resource.
	IsEphemeral() bool
	// Plural returns the plural form of the resource.
	Plural() string
	// GetStatusState returns the current status of the resource.
	GetStatusState(runtime.Object) states.State
}
```

Client manipulates Kubernetes API resources backed by template files.

#### func  NewDeploymentClient

```go
func NewDeploymentClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewDeploymentClient returns a new generic resource client.

#### func  NewHPAClient

```go
func NewHPAClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewHPAClient returns a new horizontal pod autoscaler client.

#### func  NewIngressClient

```go
func NewIngressClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewIngressClient returns a new ingress client.

#### func  NewJobClient

```go
func NewJobClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewJobClient returns a new generic resource client.

#### func  NewPodClient

```go
func NewPodClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewPodClient returns a new pod client.

#### func  NewServiceClient

```go
func NewServiceClient(globalTemplateValues GlobalTemplateValues, clientSet *kubernetes.Clientset, templateFileName string) Client
```
NewServiceClient returns a new service client.

#### type GlobalTemplateValues

```go
type GlobalTemplateValues map[string]string
```

GlobalTemplateValues encodes values which will be available to all template
specializations.
