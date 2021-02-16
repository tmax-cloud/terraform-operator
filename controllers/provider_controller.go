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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"

	"fmt"
)

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=providers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=providers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *ProviderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("provider", req.NamespacedName)

	// Fetch the Provider instance
	provider := &terraformv1alpha1.Provider{}
	err := r.Get(ctx, req.NamespacedName, provider)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Provider resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Provider")
		return ctrl.Result{}, err
	}

	// your logic here
	providerName := provider.Name
	providerCloud := provider.Spec.Cloud
	providerRegion := provider.Spec.Region

	// AWS
	providerAK := provider.Spec.AWS.AccessKey
	providerSK := provider.Spec.AWS.SecretKey

	// Azure
	providerSubID := provider.Spec.Azure.SubscriptionID
	providerClientID := provider.Spec.Azure.ClientID
	providerClientSecret := provider.Spec.Azure.ClientSecret
	providerTenantID := provider.Spec.Azure.TenantID

	//fileName := strings.ToLower(providerCloud) + "-provider.tf"

	fmt.Println("providerName:" + providerName)
	fmt.Println("providerCloud:" + providerCloud)
	// AWS
	fmt.Println("providerAK:" + providerAK)
	fmt.Println("providerSK:" + providerSK)
	fmt.Println("providerRegion:" + providerRegion)
	// Azure
	fmt.Println("providerSubID:" + providerSubID)
	fmt.Println("providerClientID:" + providerClientID)
	fmt.Println("providerClientSecret:" + providerClientSecret)
	fmt.Println("providerTenantID:" + providerTenantID)

	// Create Terraform Working Directory
	//terraDir := util.HCL_DIR + "/" + providerName

	return ctrl.Result{}, nil
}

// labelsForProvider returns the labels for selecting the resources
// belonging to the given Provider CR name.
func labelsForProvider(name string) map[string]string {
	return map[string]string{"app": "provider", "provider_cr": name}
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.Provider{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
