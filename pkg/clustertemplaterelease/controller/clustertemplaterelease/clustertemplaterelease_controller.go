package clustertemplaterelease

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	appv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/clustertemplaterelease/v1alpha1"
	kubesynchronizer "open-cluster-management.io/multicloud-operators-subscription/pkg/synchronizer/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	finalizer = "uninstall-helm-release"

	defaultMaxConcurrent = 10
)

func Add(mgr manager.Manager) error {
	synchronizer := kubesynchronizer.GetDefaultSynchronizer()

	if synchronizer == nil {
		err := fmt.Errorf("failed to get default synchronizer for ClusterTemplateRelease controller")
		klog.Error(err)

		return err
	}

	return add(mgr, newReconciler(mgr, synchronizer))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, synchronizer *kubesynchronizer.KubeSynchronizer) reconcile.Reconciler {

	return &ReconcileClusterTemplateRelease{mgr, synchronizer}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	klog.Info("The MaxConcurrentReconciles is set to: ", defaultMaxConcurrent)

	// Create a new controller
	c, err := controller.New("helmrelease-controller", mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: defaultMaxConcurrent})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HelmRelease
	if err := c.Watch(&source.Kind{Type: &appv1alpha1.ClusterTemplateRelease{}}, &handler.EnqueueRequestForObject{},
		predicate.GenerationChangedPredicate{}); err != nil {
		return err
	}

	return nil
}

type ReconcileClusterTemplateRelease struct {
	manager.Manager
	synchronizer *kubesynchronizer.KubeSynchronizer
}

func (r *ReconcileClusterTemplateRelease) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
}
