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
	"github.com/terraform-providers/terraform-provider-tls/tls"
	"github.com/tmax-cloud/terraform-operator/terranova"
)

var code string

const stateFilename = "simple.tfstate"

type TerraVars struct {
	// Provider
	ProviderName string
	Cloud        string
	AccessKey    string
	SecretKey    string
	Region       string

	// Network
	NetworkName string
	VPCCIDR     string
	SubnetCIDR  string
	RouteCIDR   string

	// Instance
	InstanceName string
	InstanceType string
	AMI          string
	KeyName      string
}

func ExecuteTerraform(input TerraVars, resourceType string, destroy bool) error {
	var platform *terranova.Platform
	var err error

	// Define template code corrensponding to resource type
	//if resourceType == "AWS_PROVIDER" {
	//	code = AWS_PROVIDER_TEMPLATE
	//} else
	if resourceType == "AWS_NETWORK" {
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
			PersistStateToFile(input.ProviderName + ".tfstate")

		if err != nil {
			return err
		}

	} else if resourceType == "AWS_INSTANCE" {

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
			Var("ami", input.AMI).
			PersistStateToFile(input.ProviderName + ".tfstate")

		if err != nil {
			return err
		}
		/*
			code = `
			variable "c"    { default = 2 }
			variable "key_name" {}
			provider "aws" {
			region        = "us-west-2"
			}
			resource "aws_instance" "server" {
			instance_type = "t2.micro"
			ami           = "ami-6e1a0117"
			count         = "${var.c}"
			key_name      = "${var.key_name}"
			}
			`
			count := 1
			keyName := "cloud"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("c", count).
				Var("key_name", keyName).
				PersistStateToFile(stateFilename)

			if err != nil {
				return err
			}
		*/
	} else {
		err = errors.New("Not Found Error: Resource Type")
		return err
	}

	/*
		platform, err := terranova.NewPlatform(code).
			AddProvider("aws", aws.Provider()).
			Var("c", count).
			Var("key_name", keyName).
			Var("access_key", access_key).
			Var("secret_key", secret_key).
			PersistStateToFile(stateFilename)

		if err != nil {
			return err
		}
	*/
	//terminate := (count == 0)

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
