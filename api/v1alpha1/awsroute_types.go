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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSRouteSpec defines the desired state of AWSRoute
type AWSRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AWSRoute. Edit AWSRoute_types.go to remove/update
	Provider string `json:"provider,omitempty"`
	VPC      string `json:"vpc,omitempty"`
	Subnet   string `json:"subnet,omitempty"`
	Gateway  string `json:"gateway,omitempty"`
	ID       string `json:"id,omitempty"`
	CIDR     string `json:"cidr,omitempty"`
}

// AWSRouteStatus defines the observed state of AWSRoute
type AWSRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
	Phase string   `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AWSRoute is the Schema for the awsroutes API
type AWSRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSRouteSpec   `json:"spec,omitempty"`
	Status AWSRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSRouteList contains a list of AWSRoute
type AWSRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSRoute{}, &AWSRouteList{})
}
