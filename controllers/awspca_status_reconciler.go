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
	api "github.com/awspca-issuer/api/v1alpha2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AWSPCAStatusReconciler struct {
	*AWSPCAIssuerReconciler
	issuer *api.AWSPCAIssuer
	logger logr.Logger
}

func newAWSPCAStatusReconciler(r *AWSPCAIssuerReconciler,
	iss *api.AWSPCAIssuer,
	log logr.Logger) *AWSPCAStatusReconciler {
	return &AWSPCAStatusReconciler{
		AWSPCAIssuerReconciler: r,
		issuer:               iss,
		logger:               log,
	}
}

func (r *AWSPCAStatusReconciler) Update(ctx context.Context,
	status api.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	r.setCondition(status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == api.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(r.issuer, eventType, reason, completeMessage)

	return r.Client.Status().Update(ctx, r.issuer)
}

func (r *AWSPCAStatusReconciler) UpdateNoError(ctx context.Context, status api.ConditionStatus, reason, message string, args ...interface{}) {
	if err := r.Update(ctx, status, reason, message, args...); err != nil {
		r.logger.Error(err, "failed to update", "status", status, "reason", reason)
	}
}

// setCondition will set a 'condition' on the given api.AWSPCAIssuer resource.
//
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
func (r *AWSPCAStatusReconciler) setCondition(status api.ConditionStatus, reason, message string) {
	now := meta.NewTime(r.Clock.Now())
	c := api.AWSPCAIssuerCondition{
		Type:               api.ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	}

	// Search through existing conditions
	for idx, cond := range r.issuer.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != api.ConditionReady {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			c.LastTransitionTime = cond.LastTransitionTime
		} else {
			r.logger.Info("found status change for AWSPCAIssuer condition; setting lastTransitionTime", "condition", cond.Type, "old_status", cond.Status, "new_status", status, "time", now.Time)
		}

		// Overwrite the existing condition
		r.issuer.Status.Conditions[idx] = c
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	r.issuer.Status.Conditions = append(r.issuer.Status.Conditions, c)
	r.logger.Info("setting lastTransitionTime for AWSPCAIssuer condition", "condition", api.ConditionReady, "time", now.Time)
}
