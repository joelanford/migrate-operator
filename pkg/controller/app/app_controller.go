package app

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	migratedv1 "github.com/joelanford/migrate-operator/pkg/apis/migrated/v1"
	originalv1alpha1 "github.com/joelanford/migrate-operator/pkg/apis/original/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_app")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new App Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	originalApp := originalv1alpha1.App{}
	originalApp.SetGroupVersionKind(schema.GroupVersionKind{Group: "original.com", Version: "v1alpha1", Kind: "App"})
	if err := add(mgr, &originalApp, newReconciler(mgr, originalApp.GroupVersionKind())); err != nil {
		return err
	}

	migratedApp := migratedv1.App{}
	migratedApp.SetGroupVersionKind(schema.GroupVersionKind{Group: "migrated.com", Version: "v1", Kind: "App"})
	if err := add(mgr, &migratedApp, newReconciler(mgr, migratedApp.GroupVersionKind())); err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, gvk schema.GroupVersionKind) *ReconcileApp {
	return &ReconcileApp{client: mgr.GetClient(), scheme: mgr.GetScheme(), gvk: gvk}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, app runtime.Object, r *ReconcileApp) error {
	// Create a new controller
	controllerName := fmt.Sprintf("%s-%s-app-controller", r.gvk.Group, r.gvk.Version)
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource App
	err = c.Watch(&source.Kind{Type: app}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner App
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    app,
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileApp{}

// ReconcileApp reconciles a App object
type ReconcileApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	gvk    schema.GroupVersionKind
}

// Reconcile reads that state of the cluster for a App object and makes changes based on the state read
// and what is in the App.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the App instance
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.gvk)
	err := r.client.Get(context.TODO(), request.NamespacedName, u)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("App not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	instance := &migratedv1.App{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, instance); err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Reconciling App", "GVK", instance.GroupVersionKind())

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set App instance as the owner and controller
	//
	// controllerutil.SetControllerReference function uses the scheme to determine the GVK, so we can't use
	// `instance` here, because its type (`&migratedv1.App{}`) is registered as `migrated.com/v1, Kind=App`.
	// Instead, we can use the unstructured object to always use the right owner type.
	if err := controllerutil.SetControllerReference(u, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *migratedv1.App) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
