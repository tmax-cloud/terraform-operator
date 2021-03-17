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
	"io/ioutil"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"

	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// RepositoryReconciler reconciles a Repository object
type RepositoryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=terraform.tmax.io,resources=repositories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=repositories/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraform.tmax.io,resources=repositories/finalizers,verbs=update

func (r *RepositoryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("repository", req.NamespacedName)

	// Fetch the Repository instance
	repository := &terraformv1alpha1.Repository{}
	err := r.Get(ctx, req.NamespacedName, repository)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Repository resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Repository")
		return ctrl.Result{}, err
	}

	// your logic here
	helper, _ := patch.NewHelper(repository, r.Client)

	defer func() {
		if err := helper.Patch(ctx, repository); err != nil {
			log.Error(err, "repository patch error")
		}
	}()

	if repository.Spec.Type == "" {
		repository.Spec.Type = "Public"
	}

	repositoryName := repository.Name
	repositoryType := repository.Spec.Type
	repositoryURL := repository.Spec.URL
	repositoryBranch := repository.Spec.Branch
	repositoryID := repository.Spec.ID
	repositoryPW := repository.Spec.PW

	repositoryAuth := &http.BasicAuth{
		Username: repositoryID,
		Password: repositoryPW,
	}

	fmt.Println("repositoryName:" + repositoryName)
	fmt.Println("repositoryType:" + repositoryType)
	fmt.Println("repositoryURL:" + repositoryURL)
	fmt.Println("repositoryBranch:" + repositoryBranch)
	fmt.Println("repositoryID:" + repositoryID)
	fmt.Println("repositoryPW:" + repositoryPW)

	// Clone Repository
	if repositoryType == "Public" {
		_, err = git.PlainClone(repositoryName, false, &git.CloneOptions{
			URL:      repositoryURL,
			Progress: os.Stdout,
		})
	} else {
		_, err = git.PlainClone(repositoryName, false, &git.CloneOptions{
			Auth:     repositoryAuth,
			URL:      repositoryURL,
			Progress: os.Stdout,
		})
	}
	if err != nil {
		log.Error(err, "Failed to clone Repository")
	}

	// Open Repository
	gitrepo, err := git.PlainOpen(repositoryName)
	if err != nil {
		log.Error(err, "Failed to Open Repository")
	}

	// Create Work Tree
	worktree, err := gitrepo.Worktree()
	if err != nil {
		log.Error(err, "Failed to Create WorkTree")
	}

	// Get All Remote Branches
	err = gitrepo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	if err != nil {
		log.Error(err, "Failed to get all remote branches")
	}
	//fmt.Sprintf("refs/heads/%s", repositoryBranch),

	// Checkout the selected branch
	if repositoryBranch != "" {
		branch := "refs/heads/" + repositoryBranch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(branch),
			Force:  true,
		})
		if err != nil {
			log.Error(err, "Failed to checkout the branch")
		}
	}

	// Pull the Git Repository
	if repositoryType == "Public" {
		err = worktree.Pull(&git.PullOptions{RemoteName: "origin"})
	} else {
		err = worktree.Pull(&git.PullOptions{
			Auth:       repositoryAuth,
			RemoteName: "origin",
		})
	}
	if err != nil {
		log.Error(err, "Failed to pull Repository")
		repository.Status.Phase = "error"
	} else {
		repository.Status.Phase = "success"
	}

	//targetDir := repositoryName
	files, err := ioutil.ReadDir(repositoryName)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		// 파일명
		fmt.Println(file.Name())
		// 파일의 절대경로
		fmt.Println(fmt.Sprintf("%v/%v", repositoryName, file.Name()))
		// 파일 내용
		//fmt.Printf("%s", file)
		content, _ := ioutil.ReadFile(fmt.Sprintf("%v/%v", repositoryName, file.Name()))
		fmt.Printf("%s", content)
	}

	// Create Terraform Working Directory
	//terraDir := util.HCL_DIR + "/" + providerName

	return ctrl.Result{RequeueAfter: time.Second * 60}, nil // Reconcile loop rescheduled after 60 seconds
	//return ctrl.Result{}, nil
}

func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terraformv1alpha1.Repository{}).
		Complete(r)
}
