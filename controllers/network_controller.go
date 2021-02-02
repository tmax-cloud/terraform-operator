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

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=networks/finalizers,verbs=update

func (r *NetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("network", req.NamespacedName)

	// Fetch the Network instance
	network := &terraformv1alpha1.Network{}
	err := r.Get(ctx, req.NamespacedName, network)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Network resource not found. Ignoring since object must be deleted")

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

			input := r.configmapToVars(cm)

			// Destroy the Provisioned Resources for Deleted Object (Network)
			destroy := true
			//err = util.ExecuteTerraform_CLI(util.HCL_DIR, isDestroy)
			err = util.ExecuteTerraform(input, "AWS_NETWORK", destroy)
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

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: network.Name, Namespace: network.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForNetwork(network)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := network.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// your logic here
	input := util.TerraVars{}

	input.NetworkName = network.Name
	input.VPCCIDR = network.Spec.VPCCIDR
	input.SubnetCIDR = network.Spec.SubnetCIDR
	input.RouteCIDR = network.Spec.RouteCIDR

	fmt.Println("NetworkName:" + input.NetworkName)
	fmt.Println("VPCCIDR:" + input.VPCCIDR)
	fmt.Println("SubnetCIDR:" + input.SubnetCIDR)
	fmt.Println("RouteCIDR:" + input.RouteCIDR)
	fmt.Println("NetworkProvider:" + network.Spec.Provider)

	provider := &terraformv1alpha1.Provider{}
	err = r.Get(ctx, types.NamespacedName{Name: network.Spec.Provider, Namespace: network.Namespace}, provider)
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

	input.ProviderName = provider.Name
	input.Cloud = provider.Spec.Cloud
	// AWS
	input.AccessKey = provider.Spec.AccessKey
	input.SecretKey = provider.Spec.SecretKey
	input.Region = provider.Spec.Region
	// Azure
	input.SubscriptionID = provider.Spec.SubscriptionID
	input.ClientID = provider.Spec.ClientID
	input.ClientSecret = provider.Spec.ClientSecret
	input.TenantID = provider.Spec.TenantID

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

	//fileName := strings.ToLower(providerCloud) + "-network.tf"

	//terraDir := util.HCL_DIR + "/" + networkProvider

	// Set Provider instance as the owner and controller
	ctrl.SetControllerReference(provider, network, r.Scheme)
	err = r.Update(ctx, network)
	if err != nil {
		log.Error(err, "Failed to update Network field - ownerReferences")
		return ctrl.Result{}, err
	}
	/*
		// Replace HCL template into HCL file
		input, err := ioutil.ReadFile(util.HCL_DIR + "/" + fileName + "_template")
		if err != nil {
			log.Error(err, "Failed to read HCL template")
			return ctrl.Result{}, err
		}

		lines := strings.Split(string(input), "\n")

		for i, line := range lines {
			if strings.Contains(line, "{{NAME}}") {
				lines[i] = strings.Replace(lines[i], "{{NAME}}", networkName, -1)
			}
			if strings.Contains(line, "{{VPC_CIDR}}") {
				lines[i] = strings.Replace(lines[i], "{{VPC_CIDR}}", networkVPC, -1)
			}
			if strings.Contains(line, "{{SUBNET_CIDR}}") {
				lines[i] = strings.Replace(lines[i], "{{SUBNET_CIDR}}", networkSubnet, -1)
			}
			if strings.Contains(line, "{{ROUTE_CIDR}}") {
				lines[i] = strings.Replace(lines[i], "{{ROUTE_CIDR}}", networkRoute, -1)
			}
			if strings.Contains(line, "{{REGION}}") {
				lines[i] = strings.Replace(lines[i], "{{REGION}}", providerRegion, -1)
			}
		}
		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(terraDir+"/"+fileName, []byte(output), 0644)
		if err != nil {
			log.Error(err, "Failed to write HCL file")
			return ctrl.Result{}, err
		}

		// Provision the resource provisioning by Terraform CLI
		isDestroy := false
		err = util.ExecuteTerraform_CLI(terraDir, isDestroy)

		if err != nil {
			provider.Status.Phase = "error"
			tErr := r.Status().Update(ctx, network)
			if tErr != nil {
				log.Error(err, "Failed to update Network Status")
				return ctrl.Result{}, tErr
			}
			if err != nil {
				log.Error(err, "Terraform Apply Error")
				return ctrl.Result{}, err
			}
		} else {
			provider.Status.Phase = "provisioned"
			tErr := r.Status().Update(ctx, network)
			if tErr != nil {
				log.Error(tErr, "Failed to update Network Status")
				return ctrl.Result{}, tErr
			}
		}
	*/
	// Check if the configmap already exists, if not create a new one
	cmList := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: network.Name, Namespace: network.Namespace}, cmList)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		cm := r.configmapForNetwork(network, input)
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

	// Provision the Network Resource by Terraform
	if network.Status.Phase != "provisioned" {
		err = util.ExecuteTerraform(input, "AWS_NETWORK", false)
	}

	if err != nil {
		network.Status.Phase = "error"
		tErr := r.Status().Update(ctx, network)
		if tErr != nil {
			log.Error(err, "Failed to update Network Status")
			return ctrl.Result{}, tErr
		}
		if err != nil {
			log.Error(err, "Terraform Apply Error")
			return ctrl.Result{}, err
		}
	} else {
		network.Status.Phase = "provisioned"
		tErr := r.Status().Update(ctx, network)
		if tErr != nil {
			log.Error(tErr, "Failed to update Network Status")
			return ctrl.Result{}, tErr
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForNetwork returns a network Deployment object
func (r *NetworkReconciler) deploymentForNetwork(m *terraformv1alpha1.Network) *appsv1.Deployment {
	ls := labelsForNetwork(m.Name)
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

// configmapForNetwork returns a network ConfigMap object
func (r *NetworkReconciler) configmapForNetwork(m *terraformv1alpha1.Network, input util.TerraVars) *corev1.ConfigMap {
	configMapData := make(map[string]string, 0)

	e := reflect.ValueOf(&input).Elem()

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		//varType := e.Type().Field(i).Type
		varValue := fmt.Sprintf("%v", e.Field(i).Interface())

		configMapData[varName] = varValue
	}
	/*
		configMapData["ProviderName"] = input.ProviderName
		configMapData["Cloud"] = input.Cloud
		configMapData["AccessKey"] = input.AccessKey
		configMapData["SecretKey"] = input.SecretKey
		configMapData["Region"] = input.Region
		configMapData["NetworkName"] = input.NetworkName
		configMapData["VPCCIDR"] = input.VPCCIDR
		configMapData["SubnetCIDR"] = input.SubnetCIDR
		configMapData["RouteCIDR"] = input.RouteCIDR
		configMapData["InstanceName"] = input.InstanceName
		configMapData["InstanceType"] = input.InstanceType
		configMapData["AMI"] = input.AMI
		configMapData["KeyName"] = input.KeyName
	*/
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
func (r *NetworkReconciler) configmapToVars(cm *corev1.ConfigMap) util.TerraVars {

	configMapData := cm.Data

	output := util.TerraVars{
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
		//InstanceName string
		//InstanceType string
		//AMI          string
		//KeyName      string
	}

	return output
}

// labelsForNetwork returns the labels for selecting the resources
// belonging to the given Network CR name.
func labelsForNetwork(name string) map[string]string {
	return map[string]string{"app": "network", "network_cr": name}
}

func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.Network{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
