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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"

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

			input := r.configmapToVars(cm)

			// Destroy the Provisioned Resources for Deleted Object (Network)
			destroy := true
			//err = util.ExecuteTerraform_CLI(util.HCL_DIR, isDestroy)
			err = util.ExecuteTerraform(input, "AWS_INSTANCE", destroy)
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

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForInstance(instance)
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
	size := instance.Spec.Size
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

	input.InstanceName = instance.Name
	input.InstanceType = instance.Spec.Type
	input.AMI = instance.Spec.Image
	input.KeyName = "aws-key"

	instanceNetwork := instance.Spec.Network

	fmt.Println("InstanceName:" + input.InstanceName)
	fmt.Println("InstanceType:" + input.InstanceType)
	fmt.Println("AMI:" + input.AMI)

	network := &terraformv1alpha1.Network{}
	err = r.Get(ctx, types.NamespacedName{Name: instanceNetwork, Namespace: instance.Namespace}, network)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Network resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Network")
		return ctrl.Result{}, err
	}

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
	input.AccessKey = provider.Spec.AccessKey
	input.SecretKey = provider.Spec.SecretKey
	input.Region = provider.Spec.Region

	fmt.Println("ProviderName:" + input.ProviderName)
	fmt.Println("Cloud:" + input.Cloud)
	// AWS
	fmt.Println("AccessKey:" + input.AccessKey)
	fmt.Println("SecretKey:" + input.SecretKey)
	fmt.Println("Region:" + input.Region)

	//fileName := strings.ToLower(providerCloud) + "-instance.tf"
	//terraDir := util.HCL_DIR + "/" + providerName

	// Set Network instance as the owner and controller
	ctrl.SetControllerReference(network, instance, r.Scheme)
	err = r.Update(ctx, instance)
	if err != nil {
		log.Error(err, "Failed to update Instance field - ownerReferences")
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
			if strings.Contains(line, "{{IMAGE}}") {
				lines[i] = strings.Replace(lines[i], "{{IMAGE}}", instanceImage, -1)
			}
			if strings.Contains(line, "{{TYPE}}") {
				lines[i] = strings.Replace(lines[i], "{{TYPE}}", instanceType, -1)
			}
			if strings.Contains(line, "{{NAME}}") {
				lines[i] = strings.Replace(lines[i], "{{NAME}}", instanceName, -1)
			}
			if strings.Contains(line, "{{NET_NAME}}") {
				lines[i] = strings.Replace(lines[i], "{{NET_NAME}}", networkName, -1)
			}
		}

		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(terraDir+"/"+fileName, []byte(output), 0644)
		if err != nil {
			log.Error(err, "Failed to write HCL file")
			return ctrl.Result{}, err
		}

		// Generate Key file
		fileName = strings.ToLower(providerCloud) + "-key.tf"

		src, err := os.Open(util.HCL_DIR + "/" + fileName + "_template")
		if err != nil {
			log.Error(err, "Failed to open HCL template")
			return ctrl.Result{}, err
		}
		defer src.Close()

		dst, err := os.Create(terraDir + "/" + fileName)
		if err != nil {
			log.Error(err, "Failed to Create HCL file")
			return ctrl.Result{}, err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			log.Error(err, "Failed to copy HCL file")
			return ctrl.Result{}, err
		}

		// Provision the resource provisioning by Terraform CLI
		isDestroy := false
		err = util.ExecuteTerraform_CLI(terraDir, isDestroy)

		if err != nil {
			provider.Status.Phase = "error"
			tErr := r.Status().Update(ctx, provider)
			if tErr != nil {
				log.Error(err, "Failed to update Instance Status")
				return ctrl.Result{}, tErr
			}
			if err != nil {
				log.Error(err, "Terraform Apply Error")
				return ctrl.Result{}, err
			}
		} else {
			provider.Status.Phase = "applyed"
			tErr := r.Status().Update(ctx, provider)
			if tErr != nil {
				log.Error(tErr, "Failed to update Instance Status")
				return ctrl.Result{}, tErr
			}
		}
	*/

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

	// Provision the Network Resource by Terraform
	err = util.ExecuteTerraform(input, "AWS_INSTANCE", false)
	if err != nil {
		provider.Status.Phase = "error"
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
		provider.Status.Phase = "provisioned"
		tErr := r.Status().Update(ctx, instance)
		if tErr != nil {
			log.Error(tErr, "Failed to update Instance Status")
			return ctrl.Result{}, tErr
		}
	}
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

func (r *InstanceReconciler) configmapToVars(cm *corev1.ConfigMap) util.TerraVars {

	configMapData := cm.Data

	output := util.TerraVars{
		ProviderName: configMapData["ProviderName"],
		Cloud:        configMapData["Cloud"],
		AccessKey:    configMapData["AccessKey"],
		SecretKey:    configMapData["SecretKey"],
		Region:       configMapData["Region"],

		NetworkName: configMapData["NetworkName"],
		VPCCIDR:     configMapData["VPCCIDR"],
		SubnetCIDR:  configMapData["SubnetCIDR"],
		RouteCIDR:   configMapData["RouteCIDR"],

		InstanceName: configMapData["InstanceName"],
		InstanceType: configMapData["InstanceType"],
		AMI:          configMapData["AMI"],
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
