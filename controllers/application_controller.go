/*
Copyright 2024 hind3ight.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "github.com/hind3ight/application-operator/api/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var counterReconcileApplication int64

//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("application", req.NamespacedName)
	log := logr.FromContext(ctx)
	<-time.NewTicker(100 * time.Millisecond).C
	counterReconcileApplication += 1
	log.Info("starting reconciling", "number: ", counterReconcileApplication)

	app := &appsv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "fail to get application")
			return ctrl.Result{}, err
		}
	}

	var result ctrl.Result
	var err error
	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "fail to reconcile deployment")
		return result, err
	}
	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "fail to reconcile service")
		return result, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Application{}, builder.WithPredicates(
			predicate.Funcs{
				CreateFunc: func(event event.CreateEvent) bool {
					return true
				},
				DeleteFunc: func(event event.DeleteEvent) bool {
					setupLog.Info("the Application has been deleted.", "name", event.Object.GetName())
					return false
				},
				UpdateFunc: func(event event.UpdateEvent) bool {
					if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
						return false
					}
					if reflect.DeepEqual(event.ObjectNew.(*appsv1.Application).Spec, event.ObjectOld.(*appsv1.Application).Spec) {
						return true
					}
					return true
				},
				GenericFunc: nil,
			})).
		Owns(&v1.Deployment{}, builder.WithPredicates(
			predicate.Funcs{
				CreateFunc: func(event event.CreateEvent) bool {
					return true
				},
				DeleteFunc: func(event event.DeleteEvent) bool {
					setupLog.Info("the Deployment has been deleted.", "name", event.Object.GetName())
					return false
				},
				UpdateFunc: func(event event.UpdateEvent) bool {
					if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
						return false
					}
					if reflect.DeepEqual(event.ObjectNew.(*v1.Deployment).Spec, event.ObjectOld.(*v1.Deployment).Spec) {
						return true
					}
					return true
				},
				GenericFunc: nil,
			})).
		Owns(&corev1.Service{}, builder.WithPredicates(
			predicate.Funcs{
				CreateFunc: func(event event.CreateEvent) bool {
					return true
				},
				DeleteFunc: func(event event.DeleteEvent) bool {
					setupLog.Info("the Deployment has been deleted.", "name", event.Object.GetName())
					return false
				},
				UpdateFunc: func(event event.UpdateEvent) bool {
					if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
						return false
					}
					if reflect.DeepEqual(event.ObjectNew.(*corev1.Service).Spec, event.ObjectOld.(*corev1.Service).Spec) {
						return true
					}
					return true
				},
				GenericFunc: nil,
			})).
		Complete(r)
}
