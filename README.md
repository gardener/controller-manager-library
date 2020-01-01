# Controller Manager Library

This basic library is intended to easily implement kubernetes controllers
and webhooks.

It contains some parts which support writing controllers, webhooks and a
generic controller manager definition. The controller manager can be used
to aggregate any number of controllers. Those controllers might also be
developed in different projects.

The library provides a cluster and resource abstraction, allowing
to work with several logical clusters that are finally mapped to
effective clusters when instantiating controller for a controller manager.

_For example:_
  There might be a controller watching resources in a source cluster and
  creating other dependent resources in a target cluster, wheras
  source and target cluster might also be identical
  
There are two basic usage modes for the controller manager:

- explicitly configuring controller manager definitions
- configuring and running a default controller manager basied on
  controller and/or webhook registrations provided by `init` functions.

The controller manger itself is just a frame for embedded
extension types. Tee implementatio of the extension does the
actual work. So far, two extension types are available:
- `controller` manages kubernetes controllers
- `webhook` manages kubernetes admission webhooks.
There might be more extension in the future.

It is possible to use the controller manager to run only controller or
webhooks, or a combination of both. Depending on the anonymous
imports only those extension types (`controller` or `webhook`) will be
incorporarted, that are really used.

## Quick and Easy

Building a controller manager consists of 4(7) steps:

- Define and register a controller and/or webhook
- Implement at least one reconciler (which does the real work) for a controller
- _Optional:_ Define the physical clusters supported by the controller manager
- Just import the package of the controller/webhook definitions that should
  be aggragated into the controller manager.
- _Optional:_ Provide the cluster mapping for the various controllers 
- Implement a simple main function.
- _Optional:_ Register additional non-standard API Groups

### Defining a Controller

The definition for a _reconciler_ watching _configmaps_ could look like this:

```go
import (
    "time"
    
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	controller.Configure("config-maps").
		Reconciler(Create).
		RequireLease().
		DefaultWorkerPool(10, 0*time.Second).
		Commands("poll").
		StringOption("test", "Controller argument").
		MainResource("core", "ConfigMap").
		MustRegister()
}
```

It also requests a command line argument (`test`) and a non-resource 
event (`poll`). Such events are called `command`.

#### The reconciler interface

A _reconciler_ is defined by a creation function (`Create` in the example above)
that is called to create a reconciler instance, when a controller is instantiated.

```go
func Create(controller controller.Interface) (reconcile.Interface, error) {

	val, err := controller.GetStringOption("test")
	if err == nil {
		controller.Infof("found option test: %s", val)
	}

	return &reconciler{controller: controller}, nil
}
```

Therefore it can access the values for the requested command line arguments.
Typically the reconciler struct should contain a field holding the actual
controller instance, because this one can be used to call several useful
methods, for example it can trigger further (subsequent) events.

```go
type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
}
```

The task of a reconciler is to handle events: commands and resource events.
Therefore it has to implement the [`reconcile.Interface`](pkg/controllermanager/controller/reconcile/interface.go)
interface. To concentrate on the function required for the actual scenario
the implementing struct can use the `reconcile.DefaultReconciler` as anonymous
member to provide a default implementation for unrequired methods.

The (optional) method `Start` is called when the controller finally is started.
Here it just issues an initial `poll` command.

```go
func (h *reconciler) Start() {
	h.controller.EnqueueCommand("poll")
}
```

The (optional) method `Commands` handles command events.

```go
func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger).RescheduleAfter(60*time.Second)
}
```

For resource reconciliation the methods `Reconcile`, `Delete` and `Deleted`
can be implemented.

```go
func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	switch o := obj.Data().(type) {
	case *corev1.ConfigMap:
		return h.reconcileConfigMap(logger, o)
	}

	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	//logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("should delete")
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	//logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("is deleted")
	return reconcile.Succeeded(logger)
}
```

`Delete` is only called if finalizers are set for the object, so it
should only be used if the reconciler uses an own finalizer and has to
remove it again. In this case the `Delete` method can be omitted.

`Deleted` is called after the object has been finally deleted. It is called
only once and cannot be retriggered. In this sense it is just a notification.
If cleanup tasks are required a finalizer should be used.

_Note:_ A finalizer can be set or removed with methods of the controller interface.

A complete example can be found [here](pkg/controllermanager/examples/controller/test/controller.go)
with the [command main package](cmds/test-controller/main.go).

### Defining a Webhook


The definition for a _webhook_ working on _resource quotas_ could look like this:

```go
import (
    "time"
    
	"github.com/gardener/controller-manager-library/pkg/controllermanager/webhook"
	corev1 "k8s.io/api/core/v1"
)


func init() {
	webhook.Configure("test.gardener.cloud").
		Cluster(cluster.DEFAULT).
		Resource("core", "ResourceQuota").
		Handler(MyHandlerType).
		MustRegister()
}
```

#### The handler interface

A _handler_ is defined by a creation function (`MyHandlerType` in the example above)
that is called to create a handler instance, when a controller is instantiated.

```go

func MyHandlerType(ext webhook.Interface) (admission.Interface, error) {
	return &MyHandler{ext: ext}, nil
}
```

Therefore it can access the values for the requested command line arguments.
Typically the handler struct should contain a field holding the actual
webhook instance, because this one can be used to call several useful
methods, for example it can be used marshal or unmarshal objects

```go

type MyHandler struct {
	admission.DefaultHandler
	ext webhook.Interface
}

func (this *MyHandler) Handle(context.Context, admission.Request) admission.Response {
	return admission.Allowed("yepp")
}

```

The task of a reconciler is to admission requests.
Therefore it has to implement the [`admission.Interface`](pkg/controllermanager/webhook/admission/interface.go)
interface. To concentrate on the function required for the actual scenario
the implementing struct can use the `admission.DefaultHandler` as anonymous
member to provide a default implementation for unrequired methods.

So far, there is no `Start`function as for the `controller`. This will change in 
later releases. To be save for such changes the `DefaultHandler` should always
be added.

A complete example can be found [here](pkg/controllermanager/examples/webhook/test/webhook.go)
with the [command main package](cmds/test-webhook/main.go).


### The Main of the Controller Manager

The main module of a controller should import the definition packages
of the desired controller with anonymous (`_`) imports

```go
import (
	_ "github.com/gardener/controller-manager-library/pkg/controller/test"
)
```

The main function then just needs to call the default controller
manager:

```go
import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager"
)

func main() {
	controllermanager.Start("test-controller", "Launch the Test Controller", "A test controller using the controller-manager-library")
}
```

Please refer to a complete [example](cmd/test-controller/main.go)


### Using Own API Groups

The used resource abstraction requires information about the object
implementations for the resources of the used API Groups.
Some standard API Groups are registered by default:

```go
import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	
	"github.com/gardener/controller-manager-library/pkg/resources"
)

func init() {
	resources.Register(corev1.SchemeBuilder)
	resources.Register(extensions.SchemeBuilder)
	resources.Register(apps.SchemeBuilder)
}
```

If other API groups are used by the controllers, they must explicity be
registered according the example above. If this libraray is redistributed
together with a new API group, this registration can directy be done in
the package defining its _SchemeBuilder_. Otherwise it could be done together
with the controller registration.


### Command Line Interface

The settings for all the configured controllers will be gathers and finally
lead to a set of command line options.

For the example controlelr above, this would look like this:

```
$ ./test-manager --help
A test controller using the controller-manager-library

Usage:
  test-controller [flags]

Flags:
      --cm.default.pool.size int   worker pool size for pool default of controller cm
      --cm.test string             Controller argument
      --controllers string         comma separated list of controllers to start (<name>,source,target,all) (default "all")
  -h, --help                       help for test-controller
      --kubeconfig string          default cluster access
      --kubeconfig.id string       id for cluster default
  -D, --log-level string           logrus log level
      --plugin-dir string          directory containing go plugins
      --pool.size int              default for all controller "pool.size" options
      --server-port-http int       directory containing go plugins
      --test string                default for all controller "test" options
time="2019-01-17T17:56:37+01:00" level=info msg="waiting for everything to shutdown (max. 120 seconds)"

```
## The complete Story

TBD
