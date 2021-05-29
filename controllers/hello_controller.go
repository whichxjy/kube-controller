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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	myappv1 "github.com/whichxjy/kube-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelloReconciler reconciles a Hello object
type HelloReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=myapp.whichxjy.com,resources=hellos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myapp.whichxjy.com,resources=hellos/status,verbs=get;update;patch

func (r *HelloReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("hello", req.NamespacedName)
	logger.Info(
		"Received request",
		"Namespace",
		req.Namespace,
		"Name",
		req.Name,
	)

	ctx := context.Background()

	hello := new(myappv1.Hello)
	if err := r.Get(ctx, req.NamespacedName, hello); err != nil {
		if kerrors.IsNotFound(err) {
			err = nil
		}
		return ctrl.Result{}, err
	}

	if hello.Status.Phase == "" {
		hello.Status.Phase = myappv1.HelloPending
	}

	logger.Info("Check phase", "Phase", hello.Status.Phase)
	requeue := false

	switch hello.Status.Phase {
	case myappv1.HelloPending:
		// Create a pod to run commands.
		pod := getHelloPod(hello)
		if err := ctrl.SetControllerReference(hello, pod, r.Scheme); err != nil {
			logger.Error(err, "Fail to set controller reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, pod); err != nil {
			logger.Error(err, "Fail to create pod")
			return ctrl.Result{}, err
		}
		hello.Status.Phase = myappv1.HelloRunning
	case myappv1.HelloRunning:
		pod := &corev1.Pod{}
		if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
			logger.Error(err, "Fail to get pod")
			return ctrl.Result{}, err
		}

		if pod.Status.Phase == corev1.PodSucceeded {
			hello.Status.Phase = myappv1.HelloSucceeded
		} else if pod.Status.Phase == corev1.PodFailed {
			hello.Status.Phase = myappv1.HelloFailed
		} else {
			requeue = true
		}
	case myappv1.HelloSucceeded:
		logger.Info("Done")
		return ctrl.Result{}, nil
	case myappv1.HelloFailed:
		pod := getHelloPod(hello)
		if err := r.Delete(ctx, pod); err != nil {
			logger.Error(err, "Fail to delete pod")
			return ctrl.Result{}, err
		}
		hello.Status.Phase = myappv1.HelloPending
	default:
		logger.Error(
			nil,
			"Invalid phase",
			"Phase",
			hello.Status.Phase,
		)
		return ctrl.Result{}, errors.New("Invalid phase")
	}

	// Update hello status.
	if err := r.Status().Update(ctx, hello); err != nil {
		logger.Error(err, "Fail to update hello status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: requeue}, nil
}

func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.Hello{}).
		Complete(r)
}

func getHelloPod(hello *myappv1.Hello) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hello.Namespace,
			Name:      hello.Name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "ubuntu",
					Image: "ubuntu",
					Command: []string{
						"/bin/sh",
						"-c",
						fmt.Sprintf(
							"seq %d | xargs -I{} echo \"Hello\"",
							hello.Spec.HelloTimes,
						),
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}
