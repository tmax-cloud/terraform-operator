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
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	claimv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"
	"github.com/tmax-cloud/terraform-operator/util"
)

// TFApplyClaimReconciler reconciles a TFApplyClaim object
type TFApplyClaimReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=claim.tmax.io,resources=tfapplyclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=claim.tmax.io,resources=tfapplyclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=claim.tmax.io,resources=tfapplyclaims/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create

func (r *TFApplyClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tfapplyclaim", req.NamespacedName)

	// Fetch the "TFApplyClaim" instance
	apply := &claimv1alpha1.TFApplyClaim{}
	err := r.Get(ctx, req.NamespacedName, apply)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("TFApplyClaim resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get TFApplyClaim")
		return ctrl.Result{}, err
	}

	// your logic here
	repoType := apply.Spec.Type
	version := apply.Spec.Version

	if version == "" {
		version = "0.11.13"
	}

	url := apply.Spec.URL
	email := apply.Spec.Email
	id := apply.Spec.ID
	pw := apply.Spec.PW
	branch := apply.Spec.Branch
	dest := "working_dir"
	//opt_terraform := "-chdir=/" + dest // only terrform 0.14+

	fmt.Println(repoType)
	fmt.Println(url)
	fmt.Println(email)
	fmt.Println(id)
	fmt.Println(pw)
	fmt.Println(branch)

	helper, _ := patch.NewHelper(apply, r.Client)

	defer func() {
		if err := helper.Patch(ctx, apply); err != nil {
			log.Error(err, "apply patch error")
		}
	}()

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: apply.Name, Namespace: apply.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForApply(apply)
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
	size := apply.Spec.Size
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

	// Update the Provider status with the pod names
	// List the pods for this provider's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(apply.Namespace),
		client.MatchingLabels(labelsForApply(apply.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "TFApplyClaim.Namespace", apply.Namespace, "TFApplyClaim.Name", apply.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	if len(podNames) < 1 {
		log.Info("Not yet create Terraform Pod...")
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	fmt.Println("10 seconds delay....")
	time.Sleep(time.Second * 10)

	fmt.Println(podNames)
	fmt.Println("podNames[0]:" + podNames[0])

	//var stdin os.Stdin
	//var stdout os.Stdout
	//var stderr os.Stderr

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "Failed to create in-cluster config")
		return ctrl.Result{}, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create clientset")
		return ctrl.Result{}, err
	}

	//err = util.ExecCmdExample(clientset, config, podNames[0], "ls", os.Stdin, os.Stdout, os.Stderr)
	//err = util.ExecCmdExample(clientset, config, podNames[0], cmd, os.Stdin, os.Stdout, os.Stderr)

	// Go Client - POD EXEC
	// 1. Git Clone Repository
	if apply.Status.Phase == "" {
		if repoType == "private" {
			var protocol string

			if strings.Contains(url, "http://") {
				protocol = "http://"
			} else if strings.Contains(url, "https://") {
				protocol = "https://"
			}

			url = strings.TrimLeft(url, protocol)
			url = protocol + id + ":" + pw + "@" + url
		}

		stdout.Reset()
		stderr.Reset()

		cmd := "git clone " + url + " " + dest
		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		if err != nil {
			log.Error(err, "Failed to Clone Git Repository")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Status.Phase = "cloned"
		}
	}

	// 2. Terraform Initialization
	if apply.Status.Phase == "cloned" {
		stdout.Reset()
		stderr.Reset()

		//cmd := "terraform init" + " " + opt_terraform
		name := "terraform"
		releases := "https://releases.hashicorp.com/terraform"

		cmd := "cd /tmp;" +
			fmt.Sprintf("wget %s/%s/%s_%s_linux_amd64.zip;", releases, version, name, version) +
			fmt.Sprintf("wget %s/%s/%s_%s_SHA256SUMS;", releases, version, name, version) +
			fmt.Sprintf("unzip -d /bin %s_%s_linux_amd64.zip;", name, version) +
			"rm -rf /tmp/build;"

		fmt.Println("CMD:" + cmd)

		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		if err != nil {
			log.Error(err, "Failed to Initialize Terraform")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Status.Phase = "awating"
		}

		stdout.Reset()
		stderr.Reset()

		//cmd := "terraform init" + " " + opt_terraform
		cmd = "cd " + dest + ";" + "terraform init"
		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		if err != nil {
			log.Error(err, "Failed to Initialize Terraform")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Status.Phase = "awating"
		}
	}

	// 3. Terraform Plan
	if (apply.Status.Phase == "awating" || apply.Status.Phase == "approved" || apply.Status.Phase == "applied") && apply.Spec.Plan == true {
		stdout.Reset()
		stderr.Reset()

		//cmd := "terraform init" + " " + opt_terraform
		cmd := "cd " + dest + ";" + "terraform plan"
		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		stdoutStderr := stdout.String() + "\n" + stderr.String()

		if err != nil {
			log.Error(err, "Failed to Plan Terraform")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Spec.Plan = false
			apply.Status.Plan = stdoutStderr
		}
	}

	// 4. Terraform Apply
	if apply.Status.Phase == "awating" && apply.Spec.Apply == true {
		apply.Status.Phase = "approved"
		apply.Spec.Apply = false
		return ctrl.Result{Requeue: true}, nil
	}
	if apply.Status.Phase == "approved" {
		stdout.Reset()
		stderr.Reset()

		//cmd := "terraform init" + " " + opt_terraform
		cmd := "cd " + dest + ";" + "terraform apply -auto-approve"
		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		stdoutStderr := stdout.String() + "\n" + stderr.String()

		if err != nil {
			log.Error(err, "Failed to Apply Terraform")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Status.Phase = "applied"
			apply.Status.Apply = stdoutStderr
		}

		var matched string
		var added, changed, destroyed int

		lines := strings.Split(string(stdoutStderr), "\n")

		for i, line := range lines {
			if strings.Contains(line, "Apply complete!") {
				matched = lines[i]
				s := strings.Split(string(matched), " ")

				added, _ = strconv.Atoi(s[3])
				changed, _ = strconv.Atoi(s[5])
				destroyed, _ = strconv.Atoi(s[7])
			}
		}

		if added > 0 || changed > 0 || destroyed > 0 { // if Terrform State changed
			stdout.Reset()
			stderr.Reset()

			//cmd := "terraform init" + " " + opt_terraform
			cmd = "cd " + dest + ";" +
				"git config --global user.email " + email + ";" +
				"git config --global user.name " + id + ";" +
				"git config --global user.password " + pw + ";" +
				"git add terraform.tfstate;" +
				"git commit -m \"Commited by TFApplyClaim Opeator\";" +
				"git push"

			err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

			fmt.Println(stdout.String())
			fmt.Println(stderr.String())

			if err != nil {
				log.Error(err, "Failed to Push tfstate file")
				apply.Status.Phase = "error"
				return ctrl.Result{}, err
			}
		}
	}

	// 5. Terraform Destroy (if required)
	if apply.Status.Phase == "applied" && apply.Spec.Destroy == true {
		stdout.Reset()
		stderr.Reset()

		//cmd := "terraform init" + " " + opt_terraform
		cmd := "cd " + dest + ";" + "terraform destroy -auto-approve"
		err = util.ExecCmdExample(clientset, config, podNames[0], cmd, nil, &stdout, &stderr)

		fmt.Println(stdout.String())
		fmt.Println(stderr.String())

		stdoutStderr := stdout.String() + "\n" + stderr.String()

		if err != nil {
			log.Error(err, "Failed to Destroy Terraform")
			apply.Status.Phase = "error"
			return ctrl.Result{}, err
		} else {
			apply.Status.Phase = "awating"
			apply.Status.Destroy = stdoutStderr
		}
	}

	//if err != nil {
	//	log.Error(err, "Failed to pull Repository")
	//	apply.Status.Phase = "error"
	//} else {
	//	apply.Status.Phase = "success"
	//}

	// Create Terraform Working Directory
	//terraDir := util.HCL_DIR + "/" + providerName
	fmt.Println("**END OF RECONCILE LOOP***")
	return ctrl.Result{RequeueAfter: time.Second * 60}, nil // Reconcile loop rescheduled after 60 seconds

}

// deploymentForProvider returns a provider Deployment object
func (r *TFApplyClaimReconciler) deploymentForApply(m *claimv1alpha1.TFApplyClaim) *appsv1.Deployment {
	ls := labelsForApply(m.Name)
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
						Image:           "192.168.6.197:5000/ubuntu:0.2",
						Name:            "ubuntu",
						Command:         []string{"/bin/sleep", "3650d"},
						ImagePullPolicy: "Always",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "ubuntu",
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

// labelsForProvider returns the labels for selecting the resources
// belonging to the given Provider CR name.
func labelsForApply(name string) map[string]string {
	return map[string]string{"app": "tfapplyclaim", "tfapplyclaim_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *TFApplyClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claimv1alpha1.TFApplyClaim{}).
		Complete(r)
}
