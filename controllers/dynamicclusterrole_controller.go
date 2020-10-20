/*


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
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbacv1alpha1 "github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	helpers "github.com/redhat-cop/dynamic-rbac-operator/helpers"
)

// DynamicClusterRoleReconciler reconciles a DynamicClusterRole object
type DynamicClusterRoleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=dynamicclusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=dynamicclusterroles/status,verbs=get;update;patch

func (r *DynamicClusterRoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("dynamicclusterrole", req.NamespacedName)

	instance := &rbacv1alpha1.DynamicClusterRole{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return ReconcileDynamicClusterRole(instance, r.Client, r.Scheme, r.Log)
}

func ReconcileDynamicClusterRole(dynamicClusterRole *rbacv1alpha1.DynamicClusterRole, client client.Client, scheme *runtime.Scheme, logger logr.Logger) (ctrl.Result, error) {
	rules, err := helpers.BuildPolicyRules(client, helpers.ClusterRole, "", dynamicClusterRole.Spec.Inherit, dynamicClusterRole.Spec.Allow, dynamicClusterRole.Spec.Deny)
	if err != nil {
		return reconcile.Result{}, err
	}

	outputRole := &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: dynamicClusterRole.Name,
			Annotations: map[string]string{
				"managed-by": "dynamic-rbac-operator",
			},
		},
		Rules: *rules,
	}

	if err := controllerutil.SetControllerReference(dynamicClusterRole, outputRole, scheme); err != nil {
		return reconcile.Result{}, err
	}

	logger.Info(fmt.Sprintf("Computed role with %d rules.", len(outputRole.Rules)))
	logger.Info("Creating or Updating Role")
	err = helpers.CreateOrUpdateClusterRole(outputRole, client)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *DynamicClusterRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1alpha1.DynamicClusterRole{}).
		Complete(r)
}
