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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"

	//"os/exec"
	"fmt"
	"reflect"

	"github.com/tmax-cloud/terraform-operator/util"
)

// ResourceReconciler reconciles a Resource object
type ResourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=resources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=resources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=resources/finalizers,verbs=update

func (r *ResourceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("resource", req.NamespacedName)

	// Fetch the Resource instance
	resource := &terraformv1alpha1.Resource{}
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

			// Recover "Network" Data using ConfigMap
			input := r.configmapToVars(cm)

			// Destroy the Provisioned Resources for Deleted Object (Network)
			//err = util.ExecuteTerraform_CLI(util.HCL_DIR, isDestroy)
			destroy := true
			err = util.ExecuteTerraform(input, input.Type, destroy)
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
		log.Error(err, "Failed to get Network")
		return ctrl.Result{}, err
	}

	// your logic here

	input := util.TerraVars{}

	input.Namespace = resource.Namespace
	input.Name = resource.Name
	input.Type = resource.Spec.Type
	input.VPCCIDR = resource.Spec.AWS_VPC.CIDR
	input.SubnetCIDR = resource.Spec.AWS_SUBNET.CIDR
	input.RouteCIDR = resource.Spec.AWS_ROUTE.CIDR

	fmt.Println("Name:" + input.Name)
	fmt.Println("VPCCIDR:" + input.VPCCIDR)
	fmt.Println("SubnetCIDR:" + input.SubnetCIDR)
	fmt.Println("RouteCIDR:" + input.RouteCIDR)
	fmt.Println("Provider:" + resource.Spec.Provider)

	// Fetch the "Provider" instance related to "Network" (Network -> Provider)
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
	ctrl.SetControllerReference(provider, resource, r.Scheme)
	if err = r.Update(ctx, resource); err != nil {
		log.Error(err, "Failed to update Resource field - ownerReferences")
		return ctrl.Result{}, err
	}

	// Check if the configmap already exists, if not create a new one
	cmList := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: resource.Name, Namespace: resource.Namespace}, cmList)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		cm := r.configmapForResource(resource, input)
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
		err = util.ExecuteTerraform(input, input.Type, false)
	}

	// Set 'Phase' Status depending on the result of 'ExecuteTerraform'
	if err != nil {
		resource.Status.Phase = "error"
		tErr := r.Status().Update(ctx, resource)
		if tErr != nil {
			log.Error(err, "Failed to update Resource Status")
			return ctrl.Result{}, tErr
		}
		if err != nil {
			log.Error(err, "Terraform Apply Error")
			return ctrl.Result{}, err
		}
	} else {
		resource.Status.Phase = "provisioned"
		tErr := r.Status().Update(ctx, resource)
		if tErr != nil {
			log.Error(tErr, "Failed to update Resource Status")
			return ctrl.Result{}, tErr
		}
	}

	return ctrl.Result{}, nil
}

// configmapForResource returns a Resource ConfigMap object
func (r *ResourceReconciler) configmapForResource(m *terraformv1alpha1.Resource, input util.TerraVars) *corev1.ConfigMap {
	configMapData := make(map[string]string, 0)

	e := reflect.ValueOf(&input).Elem()

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		//varType := e.Type().Field(i).Type
		varValue := fmt.Sprintf("%v", e.Field(i).Interface())

		configMapData[varName] = varValue
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Data: configMapData,
	}
	return cm
}

// configmapToVars returns a Terraform Variable Struct
func (r *ResourceReconciler) configmapToVars(cm *corev1.ConfigMap) util.TerraVars {

	configMapData := cm.Data

	output := util.TerraVars{
		Namespace: configMapData["Namespace"],

		ProviderName:   configMapData["ProviderName"],
		Cloud:          configMapData["Cloud"],
		AccessKey:      configMapData["AccessKey"],
		SecretKey:      configMapData["SecretKey"],
		Region:         configMapData["Region"],
		SubscriptionID: configMapData["SubscriptionID"],
		ClientID:       configMapData["ClientID"],
		ClientSecret:   configMapData["ClientSecret"],
		TenantID:       configMapData["TenantID"],

		NetworkName: configMapData["NetworkName"],
		VPCCIDR:     configMapData["VPCCIDR"],
		SubnetCIDR:  configMapData["SubnetCIDR"],
		RouteCIDR:   configMapData["RouteCIDR"],
	}

	return output
}

func (r *ResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.Resource{}).
		Complete(r)
}
