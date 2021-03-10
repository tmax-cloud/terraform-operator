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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"

	//"os/exec"
	"fmt"

	"github.com/tmax-cloud/terraform-operator/util"
)

// AWSVPCReconciler reconciles a AWSVPC object
type AWSVPCReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=awsvpcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=awsvpcs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=awsvpcs/finalizers,verbs=update

func (r *AWSVPCReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("awsvpc", req.NamespacedName)
	var id string

	// Fetch the AWS-VPC instance
	resource := &terraformv1alpha1.AWSVPC{}
	err := r.Get(ctx, req.NamespacedName, resource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("`Resource` resource not found. Ignoring since object must be deleted")

			cm := &corev1.ConfigMap{}
			err = r.Get(ctx, req.NamespacedName, cm)
			if err != nil {
				if errors.IsNotFound(err) {
					// Request object not found, could have been deleted after reconcile request.
					// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
					// Return and don't requeue
					log.Info("ConfigMap resource not found. Ignoring since object must be deleted")
					return ctrl.Result{}, nil
				}
				// Error reading the object - requeue the request.
				log.Error(err, "Failed to get ConfigMap")
				return ctrl.Result{}, err
			}

			// Recover "Resource" Data using ConfigMap
			input := util.ConfigmapToVars(cm)

			// Destroy the Provisioned Resources for Deleted Object (Resource)
			//err = util.ExecuteTerraform_CLI(util.HCL_DIR, isDestroy)
			destroy := true
			id, err = util.ExecuteTerraform(input, destroy)
			if err != nil {
				log.Error(err, "Terraform Destroy Error")
				return ctrl.Result{}, err
			}

			err = r.Delete(ctx, cm)
			if err != nil {
				log.Error(err, "Failed to delete new Confgimap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Resource")
		return ctrl.Result{}, err
	}

	// your logic here

	defer func() {
		if err = r.Update(ctx, resource); err != nil {
			log.Error(err, "Failed to update Resource")
		}
		if err = r.Status().Update(ctx, resource); err != nil {
			log.Error(err, "Failed to update Resource Status")
		}
	}()

	input := util.TerraVars{}

	input.Name = resource.Name
	input.Namespace = resource.Namespace
	input.VPCName = resource.Name
	//input.VPCID = resource.Spec.ID
	input.Type = resource.Kind
	input.VPCCIDR = resource.Spec.CIDR

	// Fetch the "Provider" instance related to "Resource" (Resource -> Provider)
	provider := &terraformv1alpha1.Provider{}
	err = r.Get(ctx, types.NamespacedName{Name: resource.Spec.Provider, Namespace: resource.Namespace}, provider)
	if err != nil {
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Provider")
		return ctrl.Result{}, err
	}

	input.ProviderName = provider.Name
	input.Cloud = provider.Spec.Cloud
	input.Region = provider.Spec.Region

	// AWS
	input.AccessKey = provider.Spec.AWS.AccessKey
	input.SecretKey = provider.Spec.AWS.SecretKey
	// Azure
	input.SubscriptionID = provider.Spec.Azure.SubscriptionID
	input.ClientID = provider.Spec.Azure.ClientID
	input.ClientSecret = provider.Spec.Azure.ClientSecret
	input.TenantID = provider.Spec.Azure.TenantID

	fmt.Println("ProviderName:" + input.ProviderName)
	fmt.Println("Cloud:" + input.Cloud)
	// AWS
	fmt.Println("AccessKey:" + input.AccessKey)
	fmt.Println("SecretKey:" + input.SecretKey)
	fmt.Println("Region:" + input.Region)
	// Azure
	fmt.Println("AccessKey:" + input.SubscriptionID)
	fmt.Println("ClientID:" + input.ClientID)
	fmt.Println("ClietnSecret:" + input.ClientSecret)
	fmt.Println("TenantID:" + input.TenantID)

	// Set Provider as the owner and controller in Resource CR
	if err = ctrl.SetControllerReference(provider, resource, r.Scheme); err != nil {
		log.Error(err, "Failed to set ownerReferences")
		return ctrl.Result{}, err
	}

	// Check if the configmap already exists, if not create a new one
	cmList := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: resource.Name, Namespace: resource.Namespace}, cmList)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap

		cm := util.ConfigmapForResource(input)
		log.Info("Creating a new Configmap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.Create(ctx, cm)
		if err != nil {
			log.Error(err, "Failed to create new Confgimap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
			return ctrl.Result{}, err
		}
		// Configmap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Configmap")
		return ctrl.Result{}, err
	}

	// Provision the Resource Resource by Terraform. It'll skip
	// when `Phase` is `provisioned`.
	if resource.Status.Phase != "provisioned" {
		id, err = util.ExecuteTerraform(input, false)

		if err != nil {
			resource.Status.Phase = "error"
		} else {
			resource.Status.Phase = "provisioned"
			resource.Spec.ID = id
		}
	}
	return ctrl.Result{}, nil
}

func (r *AWSVPCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.AWSVPC{}).
		Complete(r)
}
