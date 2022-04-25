package clustertemplaterelease

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	appv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/clustertemplaterelease/v1alpha1"
	kubesynchronizer "open-cluster-management.io/multicloud-operators-subscription/pkg/synchronizer/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	helmreleasev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/helmrelease/v1"
	appsClient "open-cluster-management.io/multicloud-operators-subscription/pkg/client/clientset/versioned"
	helmInformer "open-cluster-management.io/multicloud-operators-subscription/pkg/client/informers/externalversions"
)

const (
	finalizer = "uninstall-helm-release"

	defaultMaxConcurrent = 10
)

func Add(mgr manager.Manager, helmReleaseInformer helmInformer.SharedInformerFactory, appsClient appsClient.Clientset) error {
	synchronizer := kubesynchronizer.GetDefaultSynchronizer()

	if synchronizer == nil {
		err := fmt.Errorf("failed to get default synchronizer for ClusterTemplateRelease controller")
		klog.Error(err)

		return err
	}

	return add(mgr, newReconciler(mgr, synchronizer, helmReleaseInformer.Lister(), appsClient))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, synchronizer *kubesynchronizer.KubeSynchronizer, helmReleaseLister helmLister.HelmReleaseLister, appsClient appsClient.Clientset) reconcile.Reconciler {

	return &ReconcileClusterTemplateRelease{mgr, synchronizer, helmReleaseLister, appsClient}
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
	synchronizer      *kubesynchronizer.KubeSynchronizer
	helmReleaseLister helmLister.HelmReleaseLister
	appsClient        appsClient.Clientset
}

func (r *ReconcileClusterTemplateRelease) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	klog.V(1).Info("Reconciling HelmRelease: ", request.Namespace, "/", request.Name)

	// Fetch the HelmRelease instance
	instance := &appv1alpha1.ClusterTemplateRelease{}
	err := r.GetClient().Get(context.TODO(), request.NamespacedName, instance)

	if err != nil {
		klog.Error("Failed to lookup resource ", request.Namespace, "/", request.Name, " ", err)
		return reconcile.Result{}, err
	}

	helmReleases, err := r.helmReleaseLister.List(labels.Everything())

	if err != nil {
		klog.Error("Failed to lookup helm releases")
		return reconcile.Result{}, err
	}

	helmReleaseFound := false

	for _, v := range helmReleases {
		if v.ObjectMeta.Name == instance.ObjectMeta.Name {
			helmReleaseFound = true
		}
	}

	if !helmReleaseFound {
		r.appsClient.AppsV1alpha1().Create(ctx, &helmreleasev1.HelmRelease{
			TypeMeta: metav1.TypeMeta{
				Kind:       "HelmRelease",
				APIVersion: "apps.open-cluster-management.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.ObjectMeta.Name,
				Namespace: instance.ObjectMeta.Namespace,
			},
			Repo: helmreleasev1.HelmReleaseRepo{
				Source: &helmreleasev1.Source{
					SourceType: helmreleasev1.GitHubSourceType,
					GitHub: &helmreleasev1.GitHub{
						Urls:      []string{"https://foo.git"},
						ChartPath: "testhr/github/subscription-release-test-3",
						Branch:    "main",
					},
				},
				ChartName: "subscription-release-test-1",
			},
		}, metav1.CreateOptions{})
	}

	return reconcile.Result{}, err

}
