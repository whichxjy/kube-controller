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

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	myappv1 "github.com/whichxjy/kube-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		hello.Status.Phase = myappv1.InitPhase
	}

	logger.Info("Check phase", "Phase", hello.Status.Phase)

	switch hello.Status.Phase {
	case myappv1.InitPhase:
		// Create a pod to run commands.
		pod := newHelloPod(hello)
		if err := ctrl.SetControllerReference(hello, pod, r.Scheme); err != nil {
			logger.Error(err, "Fail to set controller reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, pod); err != nil {
			logger.Error(err, "Fail to set create pod")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case myappv1.RunningPhase:
	case myappv1.CompletedPhase:
	default:
		logger.Error(
			nil,
			"Invalid phase",
			"Phase",
			hello.Status.Phase,
		)
		return ctrl.Result{}, errors.New("Invalid phase")
	}

	return ctrl.Result{}, nil
}

func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.Hello{}).
		Complete(r)
}

func newHelloPod(hello *myappv1.Hello) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hello.Namespace,
			Name:      hello.Name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"echo hello"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}
