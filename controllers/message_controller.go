/*
Copyright 2021.

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
	"fmt"

	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hellov1beta1 "github.com/bells17/k8s-sample-controller/api/v1beta1"
)

// MessageReconciler reconciles a Message object
type MessageReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const messageTemplate string = "Hello %s"

//+kubebuilder:rbac:groups=hello.bells17.io,resources=messages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hello.bells17.io,resources=messages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hello.bells17.io,resources=messages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Message object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MessageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("type", "reconcile", "message", req.NamespacedName)
	log.Info("start reconcile")

	var msg hellov1beta1.Message
	if err := r.Get(ctx, req.NamespacedName, &msg); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "failed to get message resource")
			return ctrl.Result{}, err
		}
		log.Info("skip reconcile because the target resource is not found")
		return ctrl.Result{}, nil
	}

	log.Info("msg debug", "msg", msg)
	msg.Status.Message = fmt.Sprintf(messageTemplate, msg.Spec.Message)
	if err := r.Status().Update(ctx, &msg); err != nil {
		log.Error(err, "failed to update message status")
		return ctrl.Result{}, err
	}
	log.Info("updating message status is succeeded")

	log.Info("complete reconciling")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MessageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1beta1.Message{}).
		Complete(r)
}
