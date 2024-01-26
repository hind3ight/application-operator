package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "github.com/hind3ight/application-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ApplicationReconciler) reconcileService(ctx context.Context, app *appsv1.Application) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	dp := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)
	if err == nil {
		log.Info("the Service has already exist")
		if reflect.DeepEqual(dp.Status, app.Status.Network) {
			return ctrl.Result{}, nil
		}
		app.Status.Network = dp.Status
		if err = r.Status().Update(ctx, app); err != nil {
			log.Error(err, "fail to update Application status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("the Application status has been updated")
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "fail to get Service, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	newSv := &corev1.Service{}
	newSv.SetName(app.Name)
	newSv.SetNamespace(app.Namespace)
	newSv.SetLabels(app.Labels)
	newSv.Spec = app.Spec.Service.ServiceSpec
	newSv.Spec.Selector = app.Labels

	if err = ctrl.SetControllerReference(app, newSv, r.Scheme); err != nil {
		log.Error(err, "fail to SetControllerReference, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	if err = r.Create(ctx, newSv); err != nil {
		log.Error(err, "fail to create Service, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	log.Info("the Service has been created")
	return ctrl.Result{}, nil
}
