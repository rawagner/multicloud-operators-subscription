// Copyright 2021 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package subscription

import (
	"context"
	"log"
	stdlog "log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"open-cluster-management.io/multicloud-operators-subscription/pkg/apis"
	appv1alpha1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/v1"
)

var cfg *rest.Config

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "deploy", "crds"),
			filepath.Join("..", "..", "..", "hack", "test"),
		},
	}

	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	var c client.Client

	if c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme}); err != nil {
		log.Fatal(err)
	}

	err = c.Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "ns-sub"},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-chn-namespace"},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-sub-namespace"},
	})
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	t.Stop()
	os.Exit(code)
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		res, err := inner.Reconcile(ctx, req)
		requests <- req
		return res, err
	})

	return fn, requests
}

type Reconciliation struct {
	request reconcile.Request
	result  reconcile.Result
	err     error
}

func ReconcilerSpy(inner reconcile.Reconciler) (reconcile.Reconciler, chan Reconciliation) {
	requests := make(chan Reconciliation)
	fn := reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(ctx, req)
		requests <- Reconciliation{
			request: req,
			result:  result,
			err:     err,
		}

		return result, err
	})

	return fn, requests
}

func spyReconciler(mgr manager.Manager,
	hubclient client.Client,
	subscribers map[string]appv1alpha1.Subscriber,
	clk clock,
	standalone bool) reconcile.Reconciler {
	rec := &ReconcileSubscription{
		Client:      mgr.GetClient(),
		scheme:      mgr.GetScheme(),
		hubclient:   hubclient,
		subscribers: subscribers,
		clk:         clk,
		standalone:  standalone,
	}

	return rec
}

// StartTestManager adds recFn
func StartTestManager(ctx context.Context, mgr manager.Manager, g *gomega.GomegaWithT) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		wg.Done()
		mgr.Start(ctx)
	}()

	return wg
}
