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
	api "github.com/awspca-issuer/api/v1alpha2"
	"github.com/awspca-issuer/provisioners"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AWSPCAIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=certmanager.awspca,resources=awspcaissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certmanager.awspca,resources=awspcaissuers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile will read and validate the AWSPCAIssuer resources, it will set the
// status condition ready to true if everything is right.
func (r *AWSPCAIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("awspcaissuer", req.NamespacedName)

	iss := new(api.AWSPCAIssuer)
	if err := r.Client.Get(ctx, req.NamespacedName, iss); err != nil {
		log.Error(err, "failed to retrieve AWSPCAIssuer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	statusReconciler := newAWSPCAStatusReconciler(r, iss, log)
	if err := validateAWSPCAIssuerSpec(iss.Spec); err != nil {
		log.Error(err, "failed to validate AWSPCAIssuer resource")
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "Validation", "Failed to validate resource: %v", err)
		return ctrl.Result{}, err
	}

	// Initialize and store the provisioner
	p := provisioners.NewProvisioner(iss.Spec.Provisioner.AccessKey,
		iss.Spec.Provisioner.SecretKey, iss.Spec.Provisioner.Arn, iss.Spec.Provisioner.Region)

	issNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	provisioners.Store(issNamespaceName, p)

	return ctrl.Result{}, statusReconciler.Update(ctx, api.ConditionTrue, "Verified", "AWSPCAIssuer verified and ready to sign certificates")
}

// SetupWithManager initializes the AWSPCAIssuer controller into the controller
// runtime.
func (r *AWSPCAIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AWSPCAIssuer{}).
		Complete(r)
}

func validateAWSPCAIssuerSpec(s api.AWSPCAIssuerSpec) error {
	switch {
	case s.Provisioner.AccessKey == "":
		return fmt.Errorf("spec.provisioner.accesskey cannot be empty")
	case s.Provisioner.SecretKey == "":
		return fmt.Errorf("spec.provisioner.secretkey cannot be empty")
	case s.Provisioner.Region == "":
		return fmt.Errorf("spec.provisioner.region cannot be empty")
	case s.Provisioner.Arn == "":
		return fmt.Errorf("spec.provisioner.arn cannot be empty")
	default:
		return nil
	}
}
