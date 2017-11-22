package reconcile

import (
	"context"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/NervanaSystems/kube-controllers-go/pkg/crd"
	"github.com/NervanaSystems/kube-controllers-go/pkg/resource"
	"github.com/NervanaSystems/kube-controllers-go/pkg/states"
	"github.com/golang/glog"
)

// Reconciler periodically checks the status of subresources and takes
// various self-healing and convergence actions. These include updating
// the top-level custom resource status, re-creating missing subresources,
// deleting orphaned subresources, et cetera.
//
// See the docs/reconciliation.md file for a detailed description of the
// reconciliation policy.
type Reconciler struct {
	namespace       string
	gvk             schema.GroupVersionKind
	crdHandle       *crd.Handle
	crdClient       crd.Client
	resourceClients []resource.Client
}

// New returns a new Reconciler.
func New(namespace string, gvk schema.GroupVersionKind, crdHandle *crd.Handle, crdClient crd.Client, resourceClients []resource.Client) *Reconciler {
	return &Reconciler{
		namespace:       namespace,
		gvk:             gvk,
		crdHandle:       crdHandle,
		crdClient:       crdClient,
		resourceClients: resourceClients,
	}
}

// Run starts the reconciliation loop and blocks until the context is done, or
// there is an unrecoverable error. Reconciliation actions are done at the
// supplied interval.
func (r *Reconciler) Run(ctx context.Context, interval time.Duration) error {
	glog.V(4).Infof("Starting reconciler for %v.%v.%v", r.gvk.Group, r.gvk.Version, r.gvk.Kind)
	go wait.Until(r.run, interval, ctx.Done())
	<-ctx.Done()
	return ctx.Err()
}

type subresource struct {
	client resource.Client
	object metav1.Object
}

type action struct {
	newCRState           states.State
	newCRReason          string
	subresourcesToCreate []*subresource
	subresourcesToDelete []*subresource
}

func (a action) String() string {
	var sCreateNames []string
	for _, s := range a.subresourcesToCreate {
		sCreateNames = append(sCreateNames, s.client.Plural())
	}
	var sDeleteNames []string
	for _, s := range a.subresourcesToDelete {
		sDeleteNames = append(sDeleteNames, s.client.Plural())
	}
	return fmt.Sprintf(
		`{
  newCRState: "%s",
  newCRReason: "%s",
  subresourcesToCreate: "%s",
  subresourcesToDelete: "%s"
}`,
		a.newCRState,
		a.newCRReason,
		strings.Join(sCreateNames, ", "),
		strings.Join(sDeleteNames, ", "))
}

// Contains subresources grouped by their controlling resource.
type subresourceMap map[string][]*subresource

func (r *Reconciler) run() {
	subresourcesByCR := r.groupSubresourcesByCustomResource()
	for crName, subs := range subresourcesByCR {
		a, cr, err := r.planAction(crName, subs)
		if err != nil {
			glog.Errorf(`failed to plan action for custom resource: [%s] subresources: [%v] error: [%s]`, crName, subresourcesByCR, err.Error())
			continue
		}
		glog.Infof("planned action: %s", a.String())
		errs := r.executeAction(crName, cr, a)
		if len(errs) > 0 {
			glog.Errorf(`failed to execute action for custom resource: [%s] subresources: %v errors: %v`, crName, subresourcesByCR, errs)
		}
	}
}

func (r *Reconciler) groupSubresourcesByCustomResource() subresourceMap {
	result := subresourceMap{}
	for _, resourceClient := range r.resourceClients {
		objects, err := resourceClient.List(r.namespace)
		if err != nil {
			glog.Warningf(`[reconcile] failed to list "%s" subresources`, resourceClient.Plural())
			continue
		}

		for _, obj := range objects {
			controllerRef := metav1.GetControllerOf(obj)
			if controllerRef == nil {
				glog.V(4).Infof("[reconcile] ignoring sub-resource %v, %v as it doesn not have a controller reference", obj.GetName(), r.namespace)
				continue
			}
			// Only manipulate controller-created subresources.
			if controllerRef.APIVersion != r.gvk.GroupVersion().String() || controllerRef.Kind != r.gvk.Kind {
				glog.V(4).Infof("[reconcile] ignoring sub-resource %v, %v as controlling custom resource is from a different group, version and kind", obj.GetName(), r.namespace)
				continue
			}
			controllerName := controllerRef.Name
			objList := result[controllerName]
			result[controllerName] = append(objList, &subresource{resourceClient, obj})
		}
	}
	return result
}

func (r *Reconciler) planAction(controllerName string, subs []*subresource) (*action, crd.CustomResource, error) {
	glog.V(4).Infof("planning action for controller: [%s]", controllerName)

	// If the controller name is empty, these are not our subresources;
	// do nothing.
	if controllerName == "" {
		return &action{}, nil, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Does not exist                | *                                          | Delete sub-resource.                    |
	crObj, err := r.crdClient.Get(r.namespace, controllerName)
	if err != nil && apierrors.IsNotFound(err) {
		return &action{subresourcesToDelete: subs}, nil, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Terminal                      | *                                          | Delete sub-resource.                    |
	cr, ok := crObj.(crd.CustomResource)
	if !ok {
		return &action{}, nil, fmt.Errorf("object retrieved from CRD client not an instance of crd.CustomResource: [%v]", crObj)
	}
	// Check whether the spec (desired state) or status (current state) is terminal.
	if states.IsTerminal(cr.GetSpecState()) || states.IsTerminal(cr.GetStatusState()) {
		subsToDelete := []*subresource{}
		for _, sub := range subs {
			subMeta, err := meta.Accessor(sub.object)
			if err != nil {
				glog.Warningf("[reconcile] error getting meta accessor for subresource: %v", err)
				continue
			}
			if subMeta.GetDeletionTimestamp() == nil {
				subsToDelete = append(subsToDelete, sub)
			}
		}
		return &action{subresourcesToDelete: subsToDelete}, cr, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Deleted                       | *                                          | Delete sub-resource.                    |
	crMeta, err := meta.Accessor(crObj)
	if err != nil {
		glog.Warningf("[reconcile] error getting meta accessor for controlling custom resource: %v", err)
	} else if crMeta.GetDeletionTimestamp() != nil {
		return &action{subresourcesToDelete: subs}, cr, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Does not exist, Non-ephemeral              | Set custom resource state to failed.    |

	// TODO(CD): need to be careful here, there is a race between the controller
	//           hooks creating the subresources in the first place and this
	//           reconcile loop.

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Deleted, Non-ephemeral                     | Set custom resource state to failed.    |
	// | Non-terminal                  | Terminal, Non-ephemeral                    | Set custom resource state to failed.    |
	for _, sub := range subs {
		subMeta, err := meta.Accessor(sub.object)
		if err != nil {
			glog.Warningf("[reconcile] error getting meta accessor for subresource: %v", err)
			continue
		}
		if !sub.client.IsEphemeral() {
			if subMeta.GetDeletionTimestamp() != nil {
				return &action{
					newCRState:  states.Failed,
					newCRReason: fmt.Sprintf(`non-ephemeral subresource "%s" for "%s" is deleted`, sub.client.Plural(), cr.Name()),
				}, cr, nil
			}
			// TODO(CD): Widen subresource terminal state detection to include
			//           all terminal states.
			if sub.client.IsFailed(r.namespace, cr.Name()) {
				return &action{
					newCRState:  states.Failed,
					newCRReason: fmt.Sprintf(`non-ephemeral subresource "%s" for "%s" is in a terminal state`, sub.client.Plural(), cr.Name()),
				}, cr, nil
			}
		}
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Non-terminal, Non-ephemeral, Spec mismatch | Set custom resource state to failed.    |

	// TODO

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Pending, Spec matches                      | Set custom resource state to pending.   |

	// TODO

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Non-terminal, Ephemeral, Spec mismatch     | Update sub-resource.                    |

	// TODO

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Terminal, Ephemeral                        | Recreate the sub-resource.              |

	// Delete terminal ephemeral subresources. They will be recreated in a
	// subsequent iteration when they are found not to exist.
	subsToDelete := []*subresource{}
	for _, sub := range subs {
		subMeta, err := meta.Accessor(sub.object)
		if err != nil {
			glog.Warningf("[reconcile] error getting meta accessor for subresource: %v", err)
			continue
		}
		if sub.client.IsEphemeral() && subMeta.GetDeletionTimestamp() == nil {
			// TODO(CD): Widen subresource terminal state detection to include
			//           all terminal states.
			if sub.client.IsFailed(r.namespace, cr.Name()) {
				subsToDelete = append(subsToDelete, sub)
			}
		}
	}
	if len(subsToDelete) > 0 {
		return &action{subresourcesToDelete: subsToDelete}, cr, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Does not exist, Ephemeral                  | Recreate the sub-resource.              |

	// ASSUMPTION: There is at most one subresource of each kind per
	//             custom resource. We use the plural form as a key.
	existingSubs := map[string]struct{}{}
	for _, sub := range subs {
		existingSubs[sub.client.Plural()] = struct{}{}
	}
	subsToCreate := []*subresource{}
	for _, subClient := range r.resourceClients {
		// TODO(CD): handle non-ephemeral subresources that do not exist.
		_, exists := existingSubs[subClient.Plural()]
		if !exists && subClient.IsEphemeral() {
			subsToCreate = append(subsToCreate, &subresource{client: subClient})
		}
	}
	if len(subsToCreate) > 0 {
		return &action{subresourcesToCreate: subsToCreate}, cr, nil
	}

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Deleted, Ephemeral                         | Do nothing.                             |

	// Nothing to do.

	// | Custom resource desired state | Sub-resource current state                 | Action                                  |
	// |:------------------------------|:-------------------------------------------|:----------------------------------------|
	// | Non-terminal                  | Running, Spec matches                      | Set custom resource state to running.   |

	// TODO

	// Fall-through case; do nothing.
	return &action{}, cr, nil

}

func (r *Reconciler) executeAction(controllerName string, cr crd.CustomResource, a *action) []error {
	errors := []error{}

	glog.V(4).Infof(`executing reconcile action for "%s" resource "%s" in namespace "%s"`, r.crdHandle.Plural, controllerName, r.namespace)
	if a.newCRState != "" {
		glog.Infof(`updating "%s" custom resource for controller "%s" in namespace "%s"`, r.crdHandle.Plural, controllerName, r.namespace)
		cr.SetStatusStateWithMessage(states.Failed, a.newCRReason)
		_, err := r.crdClient.Update(cr)
		if err != nil {
			glog.Errorf(`error updating custom resource state for "%s" in namespace "%s"`, controllerName, r.namespace)
			errors = append(errors, err)
		}
	}

	for _, s := range a.subresourcesToCreate {
		glog.Infof(`creating "%s" subresource for controller "%s" in namespace "%s"`, s.client.Plural(), controllerName, r.namespace)
		err := s.client.Create(r.namespace, cr)
		if err != nil {
			glog.Errorf(`error creating "%s" subresource for controller "%s" in namespace "%s"`, s.client.Plural(), controllerName, r.namespace)
			errors = append(errors, err)
		}
	}

	for _, s := range a.subresourcesToDelete {
		glog.Infof(`deleting "%s" subresource for controller "%s" in namespace "%s"`, s.client.Plural(), controllerName, r.namespace)
		err := s.client.Delete(r.namespace, controllerName)
		if err != nil {
			glog.Errorf(`error deleting "%s" subresource for controller "%s" in namespace "%s"`, s.client.Plural(), controllerName, r.namespace)
			errors = append(errors, err)
		}
	}

	return errors
}
