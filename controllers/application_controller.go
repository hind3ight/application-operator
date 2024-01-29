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
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/hind3ight/application-operator/api/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.have2bfun.com,resources=applications/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var CounterReconcileApplication int32

	<-time.NewTicker(100 * time.Millisecond).C
	log := logr.FromContext(ctx)

	CounterReconcileApplication += 1
	log.Info("Starting a reconcile", "number", CounterReconcileApplication)

	app := &v1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Fail to get the Application, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: genericRequeueDuration}, err
	}

	var result ctrl.Result
	var err error
	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "Fail to reconcile Deployment.")
		return result, err
	}
	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "Fail to reconcile Service.")
		return result, err
	}
	log.Info("All resources has been reconciled.")
	return ctrl.Result{}, nil
}

var genericRequeueDuration = 10 * time.Second

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Application{}).
		Complete(r)
}
