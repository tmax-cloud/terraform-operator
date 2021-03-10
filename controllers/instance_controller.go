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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"

	"fmt"

	"github.com/tmax-cloud/terraform-operator/util"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=instances/finalizers,verbs=update

func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	/*
		ctx := context.Background()
		log := r.Log.WithValues("instance", req.NamespacedName)


			// Fetch the "Instacne" instance
			instance := &terraformv1alpha1.Instance{}
			err := r.Get(ctx, req.NamespacedName, instance)

			if err != nil {
				if errors.IsNotFound(err) {
					// Request object not found, could have been deleted after reconcile request.
					// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
					// Return and don't requeue
					log.Info("Instance resource not found. Ignoring since object must be deleted")

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

					// Recover "Instance" Data using ConfigMap
					input := r.configmapToVars(cm)

					// Destroy the Provisioned Resources for Deleted Object (Network)
					destroy := true
					//err = util.ExecuteTerraform_CLI(util.HCL_DIR, isDestroy)
					err = util.ExecuteTerraform(input, destroy)
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
				log.Error(err, "Failed to get Instance")
				return ctrl.Result{}, err
			}

			// your logic here
			input := util.TerraVars{}

			input.Namespace = instance.Namespace
			input.InstanceName = instance.Name
			input.InstanceType = instance.Spec.Type
			input.ImageID = instance.Spec.Image
			input.KeyName = instance.Spec.Key

			instanceNetwork := instance.Spec.Network

			fmt.Println("InstanceName:" + input.InstanceName)
			fmt.Println("InstanceType:" + input.InstanceType)
			fmt.Println("ImageID:" + input.ImageID)
			fmt.Println("KeyName:" + input.KeyName)

			// Fetch the "Network" instance related to "Instance" (Instance -> Network)
			network := &terraformv1alpha1.Network{}
			err = r.Get(ctx, types.NamespacedName{Name: instanceNetwork, Namespace: instance.Namespace}, network)
			if err != nil {
				// Error reading the object - requeue the request.
				log.Error(err, "Failed to get Network")
				return ctrl.Result{}, err
			}

			// Fetch the "Provider" instance related to "Network" (Network -> Provider)
			provider := &terraformv1alpha1.Provider{}
			err = r.Get(ctx, types.NamespacedName{Name: network.Spec.Provider, Namespace: network.Namespace}, provider)
			if err != nil {
				// Error reading the object - requeue the request.
				log.Error(err, "Failed to get Provider")
				return ctrl.Result{}, err
			}

			input.NetworkName = network.Name
			input.VPCCIDR = network.Spec.VPCCIDR
			input.SubnetCIDR = network.Spec.SubnetCIDR
			input.RouteCIDR = network.Spec.RouteCIDR

			fmt.Println("NetworkName:" + input.NetworkName)
			fmt.Println("VPCCIDR:" + input.VPCCIDR)
			fmt.Println("SubnetCIDR:" + input.SubnetCIDR)
			fmt.Println("RouteCIDR:" + input.RouteCIDR)
			fmt.Println("NetworkProvider:" + network.Spec.Provider)

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

			//fileName := strings.ToLower(providerCloud) + "-instance.tf"
			//terraDir := util.HCL_DIR + "/" + providerName

			// Set Network as the owner and controller in Instance CR
			ctrl.SetControllerReference(network, instance, r.Scheme)
			if err = r.Update(ctx, instance); err != nil {
				log.Error(err, "Failed to update Instance field - ownerReferences")
				return ctrl.Result{}, err
			}

			// Check if the configmap already exists, if not create a new one
			cmList := &corev1.ConfigMap{}
			err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cmList)
			if err != nil && errors.IsNotFound(err) {
				// Define a new ConfigMap
				cm := r.configmapForInstance(instance, input)
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

			// Provision the Network Resource by Terraform. It'll skip
			// when `Phase` is `provisioned`.
			if instance.Status.Phase != "provisioned" {
				err = util.ExecuteTerraform(input, false)
			}

			// Set 'Phase' Status depending on the result of 'ExecuteTerraform'
			if err != nil {
				instance.Status.Phase = "error"
				tErr := r.Status().Update(ctx, instance)
				if tErr != nil {
					log.Error(err, "Failed to update Instance Status")
					return ctrl.Result{}, tErr
				}
				if err != nil {
					log.Error(err, "Terraform Apply Error")
					return ctrl.Result{}, err
				}
			} else {
				instance.Status.Phase = "provisioned"
				tErr := r.Status().Update(ctx, instance)
				if tErr != nil {
					log.Error(tErr, "Failed to update Instance Status")
					return ctrl.Result{}, tErr
				}
			}
	*/
	return ctrl.Result{}, nil
}

// deploymentForInstance returns a network Deployment object
func (r *InstanceReconciler) deploymentForInstance(m *terraformv1alpha1.Instance) *appsv1.Deployment {
	ls := labelsForInstance(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "memcached",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}
	// Set Provider instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// configmapForInstance returns a instance ConfigMap object
func (r *InstanceReconciler) configmapForInstance(m *terraformv1alpha1.Instance, input util.TerraVars) *corev1.ConfigMap {
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
func (r *InstanceReconciler) configmapToVars(cm *corev1.ConfigMap) util.TerraVars {

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

		InstanceName: configMapData["InstanceName"],
		InstanceType: configMapData["InstanceType"],
		ImageID:      configMapData["ImageID"],
		KeyName:      configMapData["KeyName"],
	}

	return output
}

// labelsForNetwork returns the labels for selecting the resources
// belonging to the given Network CR name.
func labelsForInstance(name string) map[string]string {
	return map[string]string{"app": "instance", "instance_cr": name}
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.Instance{}).
		Complete(r)
}
