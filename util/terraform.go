package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/terraform-providers/terraform-provider-aws/aws"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm"
	"github.com/terraform-providers/terraform-provider-tls/tls"
	"github.com/tmax-cloud/terraform-operator/terranova"
)

// Terraform HCL Input Parameters Structure
type TerraVars struct {
	/* Common */
	Namespace string
	Type      string

	/* Provider */
	ProviderName string
	Cloud        string
	Region       string
	// AWS Field
	AccessKey string
	SecretKey string
	// Azure Field
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string

	/* Network */
	NetworkName string
	VPCCIDR     string
	SubnetCIDR  string
	RouteCIDR   string

	/* Instance */
	InstanceName string
	InstanceType string
	ImageID      string
	KeyName      string
}

// Execute Terraform (Go Package)
// Provison or Destroy the remote resource
func ExecuteTerraform(input TerraVars, resourceType string, destroy bool) error {
	var platform *terranova.Platform // Platform is the platform to be managed by Terraform
	var code string                  // HCL (Hashicorp Configuration Language)
	var err error

	/*
		platform, err = terranova.NewPlatform(code). 		// HCL 코드 기반으로 Platform 초기화 (Default Variable)
		AddProvider("aws", aws.Provider()).					// Provider 추가 (e.g. AWS, Azure, TLS 등)
		Var("access_key", input.AccessKey).					// HCL 코드 내 변수 설정
		PersistStateToFile(input.NetworkName + ".tfstate") 	// Terraform State File 설정
		...
		platform.Apply(destroy) 							// 설정된 Context 내용 기반으로 클라우드 리소스 생성/삭제 수행
	*/

	// Define the platform corrensponding to Cloud - Resource type
	if input.Cloud == "AWS" { // Platform : AWS
		if resourceType == "NETWORK" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_NETWORK_TEMPLATE
			code = strings.Replace(code, "{{NET_NAME}}", input.NetworkName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("route_cidr", input.RouteCIDR).
				PersistStateToFile(input.Namespace + "-" + input.NetworkName + ".tfstate")

			if err != nil {
				return err
			}
		} else if resourceType == "INSTANCE" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_NETWORK_TEMPLATE + "\n" + AWS_INSTANCE_TEMPLATE + "\n" + AWS_KEY_TEMPLATE
			code = strings.Replace(code, "{{NET_NAME}}", input.NetworkName, -1)
			code = strings.Replace(code, "{{INS_NAME}}", input.InstanceName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("route_cidr", input.RouteCIDR).
				Var("instance_type", input.InstanceType).
				Var("image_id", input.ImageID).
				Var("key_pair", input.KeyName).
				PersistStateToFile(input.Namespace + "-" + input.InstanceName + ".tfstate")

			if err != nil {
				return err
			}
		} else {
			err = errors.New("Not Found Error: Resource Type")
			return err
		}
	} else if input.Cloud == "Azure" { // Platform : Azure
		if resourceType == "NETWORK" {
			code = AZURE_PROVIDER_TEMPLATE + "\n" + AZURE_NETWORK_TEMPLATE
			code = strings.Replace(code, "{{NET_NAME}}", input.NetworkName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("azurerm", azurerm.Provider()).
				Var("subscription_id", input.SubscriptionID).
				Var("client_id", input.ClientID).
				Var("client_secret", input.ClientSecret).
				Var("tenant_id", input.TenantID).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("route_cidr", input.RouteCIDR).
				PersistStateToFile(input.Namespace + "-" + input.NetworkName + ".tfstate")

			if err != nil {
				return err
			}
		} else if resourceType == "INSTANCE" {
			code = AZURE_PROVIDER_TEMPLATE + "\n" + AZURE_NETWORK_TEMPLATE + "\n" + AZURE_INSTANCE_TEMPLATE + "\n" + AZURE_KEY_TEMPLATE
			code = strings.Replace(code, "{{NET_NAME}}", input.NetworkName, -1)
			code = strings.Replace(code, "{{INS_NAME}}", input.InstanceName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("azurerm", azurerm.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("subscription_id", input.SubscriptionID).
				Var("client_id", input.ClientID).
				Var("client_secret", input.ClientSecret).
				Var("tenant_id", input.TenantID).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("route_cidr", input.RouteCIDR).
				Var("instance_type", input.InstanceType).
				Var("image_id", input.ImageID).
				Var("key_pair", input.KeyName).
				PersistStateToFile(input.Namespace + "-" + input.InstanceName + ".tfstate")

		} else {
			err = errors.New("Not Found Error: Resource Type")
			return err
		}
	} else if input.Cloud == "GCP" { // Platform : Google Cloud Platform
		/* To Do */
	} else if input.Cloud == "OpenStack" { // Platform : OpenStack
		/* To Do */
	} else if input.Cloud == "VSphere" { // Platform : VSphere
		/* To Do */
	} else {
		err = errors.New("Not Found Error: Cloud Platform")
		return err
	}

	// terminate := (count == 0)
	// Apply brings the platform to the desired state. (Provision / Destroy)
	if err := platform.Apply(destroy); err != nil {
		return err
	}
	return nil
}

// Initialize Terraform Working Directory
func InitTerraform_CLI(targetDir string, cloudType string) error {
	// Download Terraform Plugin (e.g. AWS, Azure, GCP)
	//cmd1 := exec.Command("bash", "-c", "terraform init")
	//cmd1.Dir = targetDir
	//stdoutStderr1, err1 := cmd1.CombinedOutput()
	//fmt.Printf("%s\n", stdoutStderr1)

	// Select the Terraform Plugin (cloudType: AWS, Azure, GCP)
	orgDir := HCL_DIR + "/" + ".terraform" + cloudType
	dstDir := targetDir + "/" + ".terraform"

	// Make the Destination Directory for plugin
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		err = os.Mkdir(dstDir, 0755)
		if err != nil {
			return err
		}
		// Copy the Terraform Plugin (e.g. AWS, Azure, GCP) at Woring Directory
		err = copy(orgDir, dstDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute Terraform (Apply / Destroy)
func ExecuteTerraform_CLI(targetDir string, isDestroy bool) error {

	// Provision the Resources by Terraform
	cmd := exec.Command("bash", "-c", "terraform apply -auto-approve")

	// Swith the command from "apply" to "destroy"
	if isDestroy {
		// Destroy the Reosource by Terraform
		cmd = exec.Command("bash", "-c", "terraform destroy -auto-approve")
	}

	cmd.Dir = targetDir
	stdoutStderr, err := cmd.CombinedOutput()

	fmt.Printf("%s\n", stdoutStderr)

	return err
}

// Copy a Dierectory (preserve directory structure)
func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}
