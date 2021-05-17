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

	// Git Repoistory Type (Public, Private)
	Type string `json:"type,omitempty"`
	// Terraform CLI Version
	Version string `json:"version,omitempty"`
	// Git URL (HCL Code)
	URL string `json:"url,omitempty"`
	// Git Branch
	Branch string `json:"branch,omitempty"`
	// Secret Name for Git Credential
	Secret string `json:"secret,omitempty"`
	// Whether to perform "terraform destory"
	Destroy bool `json:"destroy,omitempty"`
	// Terraform Variable
	Variable string `json:"variable,omitempty"`
}

// TFApplyClaimStatus defines the observed state of TFApplyClaim
type TFApplyClaimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Action defines the action that administrator apply
	Action string `json:"action,omitempty"`
	// Phase defines the current step for Terraform Claim
	Phase string `json:"phase,omitempty"`
	// Plans defines the information about "terraform plan"
	Plans []Plan `json:"plans,omitempty"`
	// Apply defines the information about "terraform apply"
	Apply string `json:"apply,omitempty"`
	// Destroy defines the information about "terraform destroy"
	Destroy string `json:"destroy,omitempty"`
	// State defines the contents for Terraform State File
	State string `json:"state,omitempty"`
	// Commit defines the latest commit id when apply or destroy
	Commit string `json:"commit,omitempty"`
	// Resource defines the count about added, updated, or deleted resources in Cloud Platform
	Resource Resource `json:"resource,omitempty"`
	// State defines the contents of stdout-stderr log for terraform cmd
	Log string `json:"log,omitempty"`
}

type Plan struct {
	// Last time that "terraform plan" performed.
	LastExectionTime string `json:"lastexectiontime,omitempty"`
	// The latest Commid ID that "terraform plan" peformed in
	Commit string `json:"commit,omitempty"`
	// Stdout-StdErr Log about Plan Cmd
	Log string `json:"log,omitempty"`
}

type Resource struct {
	Added   int `json:"added,omitempty"`
	Updated int `json:"updated,omitempty"`
	Deleted int `json:"deleted,omitempty"`
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
