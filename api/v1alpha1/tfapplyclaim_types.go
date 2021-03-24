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

// TFApplyClaimSpec defines the desired state of TFApplyClaim
type TFApplyClaimSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TFApplyClaim. Edit TFApplyClaim_types.go to remove/update
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Email   string `json:"email,omitempty"`
	ID      string `json:"id,omitempty"`
	PW      string `json:"pw,omitempty"`
	Size    int32  `json:"size,omitempty"`
	Plan    bool   `json:"plan,omitempty"`
	Apply   bool   `json:"apply,omitempty"`
	Destroy bool   `json:"destroy,omitempty"`
}

// TFApplyClaimStatus defines the observed state of TFApplyClaim
type TFApplyClaimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes   []string `json:"nodes,omitempty"`
	Phase   string   `json:"phase,omitempty"`
	Plan    string   `json:"plan,omitempty"`
	Apply   string   `json:"apply,omitempty"`
	Destroy string   `json:"destroy,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TFApplyClaim is the Schema for the tfapplyclaims API
type TFApplyClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TFApplyClaimSpec   `json:"spec,omitempty"`
	Status TFApplyClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TFApplyClaimList contains a list of TFApplyClaim
type TFApplyClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TFApplyClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TFApplyClaim{}, &TFApplyClaimList{})
}
