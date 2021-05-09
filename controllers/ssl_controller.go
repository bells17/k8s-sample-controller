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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	limitv1beta1 "github.com/bells17/k8s-sample-controller/api/v1beta1"
)

// SSLReconciler reconciles a SSL object
type SSLReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=limit.bells17.io,resources=ssls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=limit.bells17.io,resources=ssls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=limit.bells17.io,resources=ssls/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SSL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *SSLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("type", "reconcile", "ssl", req.NamespacedName)
	log.Info("start reconcile")

	var ssl limitv1beta1.SSL
	if err := r.Get(ctx, req.NamespacedName, &ssl); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "failed to get ssl")
			return ctrl.Result{}, err
		}
		log.Info("start reconcile because the target resource is not found")
		return ctrl.Result{}, nil
	}

	log.Info("start reconciling")
	t := time.Now().Add(time.Duration(ssl.Spec.LimitBefore) * 24 * time.Hour) // TODO
	err := r.check(ctx, log, ssl, t)
	if err != nil {
		log.Error(err, "failed to check resource")
		return ctrl.Result{}, err
	}

	log.Info("complete reconciling")
	return ctrl.Result{}, nil
}

func (r *SSLReconciler) check(ctx context.Context, log logr.Logger, ssl limitv1beta1.SSL, targetTime time.Time) error {
	var secret corev1.Secret
	err := r.Get(ctx, types.NamespacedName{
		Name:      ssl.Spec.SecretName,
		Namespace: ssl.GetNamespace(),
	}, &secret)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return errors.Wrap(err, "failed to get secret")
		}

		meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
			Type:    limitv1beta1.ConditionReady,
			Status:  metav1.ConditionUnknown,
			Reason:  limitv1beta1.ConditionReasonSecretIsNotFound,
			Message: fmt.Sprintf("%s/%s secret is not found", ssl.GetNamespace(), ssl.Spec.SecretName),
		})
		if err := r.Status().Update(ctx, &ssl); err != nil {
			return errors.Wrap(err, "failed to update status")
		}
		return nil
	}

	if secret.Type != corev1.SecretTypeTLS {
		meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
			Type:    limitv1beta1.ConditionReady,
			Status:  metav1.ConditionUnknown,
			Reason:  limitv1beta1.ConditionReasonSecretTypeIsInvalid,
			Message: fmt.Sprintf("%s/%s secret type is %q", ssl.GetNamespace(), ssl.Spec.SecretName, secret.Type),
		})
		if err := r.Status().Update(ctx, &ssl); err != nil {
			return errors.Wrap(err, "failed to update status")
		}
		return nil
	}

	if _, ok := secret.StringData[corev1.TLSPrivateKeyKey]; !ok {
		meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
			Type:    limitv1beta1.ConditionReady,
			Status:  metav1.ConditionUnknown,
			Reason:  limitv1beta1.ConditionReasonTLSKeyNotFound,
			Message: fmt.Sprintf("%s/%s secret doesn't have tls key", ssl.GetNamespace(), ssl.Spec.SecretName),
		})
		if err := r.Status().Update(ctx, &ssl); err != nil {
			return errors.Wrap(err, "failed to update status")
		}
		return nil
	}

	block, _ := pem.Decode([]byte(secret.StringData[corev1.TLSPrivateKeyKey]))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
			Type:    limitv1beta1.ConditionReady,
			Status:  metav1.ConditionUnknown,
			Reason:  limitv1beta1.ConditionReasonTLSKeyCanNotParse,
			Message: fmt.Sprintf("%s/%s secret tls key can not parse: %s", ssl.GetNamespace(), ssl.Spec.SecretName, err.Error()),
		})
		if err := r.Status().Update(ctx, &ssl); err != nil {
			return errors.Wrap(err, "failed to update status")
		}
		return nil
	}

	if cert.NotBefore.Before(targetTime) || cert.NotAfter.After(targetTime) {
		meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
			Type:   limitv1beta1.ConditionReady,
			Status: metav1.ConditionFalse,
		})
		if err := r.Status().Update(ctx, &ssl); err != nil {
			return errors.Wrap(err, "failed to update status")
		}
		return nil
	}

	meta.SetStatusCondition(&ssl.Status.Conditions, metav1.Condition{
		Type:   limitv1beta1.ConditionReady,
		Status: metav1.ConditionTrue,
	})
	if err := r.Status().Update(ctx, &ssl); err != nil {
		return errors.Wrap(err, "failed to update status")
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SSLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := r.Log.WithValues("type", "secret-handler")
	secretMapFunc := func(obj client.Object) []reconcile.Request {
		if obj == nil {
			return []reconcile.Request{}
		}

		ctx := context.Background()
		var secret corev1.Secret
		err := r.Get(ctx, types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}, &secret)

		if err != nil {
			if !apierrs.IsNotFound(err) {
				log.Error(err, "failed to get secret")
			}
			return []reconcile.Request{}
		}

		var sslList limitv1beta1.SSLList
		o := client.MatchingFields(map[string]string{"spec.secretName": secret.GetName()})
		if err := r.List(ctx, &sslList, o); err != nil {
			if !apierrs.IsNotFound(err) {
				log.Error(err, "failed to get ssl list")
			}
			return []reconcile.Request{}
		}

		var result []reconcile.Request
		for _, item := range sslList.Items {
			result = append(result, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: obj.GetNamespace(),
				},
			})
		}
		return result
	}

	secretPred := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		UpdateFunc:  func(event.UpdateEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	watcher := newPollingSSLWatcher(24 * time.Hour) // TODO
	err := mgr.Add(watcher)
	if err != nil {
		return err
	}
	src := source.Channel{
		Source: watcher.channel,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&limitv1beta1.SSL{}).
		Watches(&src, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(secretMapFunc), builder.WithPredicates(secretPred)).
		Complete(r)
}

func newPollingSSLWatcher(d time.Duration) *pollingSSLWatcher {
	ch := make(chan event.GenericEvent)
	return &pollingSSLWatcher{
		channel:  ch,
		duration: d,
	}
}

type pollingSSLWatcher struct {
	channel  chan event.GenericEvent
	client   client.Client
	duration time.Duration
}

func (r *pollingSSLWatcher) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

func (r pollingSSLWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.duration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.run(ctx)
		}
	}
}

func (r pollingSSLWatcher) run(ctx context.Context) {
	var namespaceList corev1.NamespaceList
	if err := r.client.List(ctx, &namespaceList); err != nil {
		return
	}

	for _, ns := range namespaceList.Items {
		var sslList limitv1beta1.SSLList
		o := client.InNamespace(ns.GetName())
		if err := r.client.List(ctx, &sslList, o); err != nil {
			return
		}

		for _, ssl := range sslList.Items {
			r.channel <- event.GenericEvent{
				Object: &ssl,
			}
		}
	}
}
