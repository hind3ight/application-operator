package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "github.com/hind3ight/application-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var GenericRequeueDuration = 10 * time.Second

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, app *appsv1.Application) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	dp := &v1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)
	if err == nil {
		log.Info("the Deployment has already exist")
		if reflect.DeepEqual(dp.Status, app.Status.Workflow) {
			return ctrl.Result{}, nil
		}
		app.Status.Workflow = dp.Status
		if err = r.Status().Update(ctx, app); err != nil {
			log.Error(err, "fail to update Application status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("the Application status has been updated")
		return ctrl.Result{}, nil
	}
	if errors.IsNotFound(err) {
		log.Error(err, "fail to get Deployment, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	newDp := &v1.Deployment{}
	newDp.SetName(app.Name)
	newDp.SetNamespace(app.Namespace)
	newDp.SetLabels(app.Labels)
	newDp.Spec = app.Spec.Deployment.DeploymentSpec
	newDp.Spec.Template.SetLabels(app.Labels)

	if err = ctrl.SetControllerReference(app, newDp, r.Scheme); err != nil {
		log.Error(err, "fail to SetControllerReference, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	if err = r.Create(ctx, newDp); err != nil {
		log.Error(err, "fail to create Deployment, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	log.Info("the Deployment has been created")
	return ctrl.Result{}, nil
}
