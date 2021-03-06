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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HelloPhase string

const (
	HelloPending   HelloPhase = "Pending"
	HelloRunning   HelloPhase = "Running"
	HelloSucceeded HelloPhase = "Succeeded"
	HelloFailed    HelloPhase = "Failed"
)

// HelloSpec defines the desired state of Hello
type HelloSpec struct {
	HelloTimes uint `json:"helloTimes,omitempty"`
}

// HelloStatus defines the observed state of Hello
type HelloStatus struct {
	Phase HelloPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hello is the Schema for the hellos API
type Hello struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelloSpec   `json:"spec,omitempty"`
	Status HelloStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelloList contains a list of Hello
type HelloList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Hello `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hello{}, &HelloList{})
}
