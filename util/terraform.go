package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/terraform-providers/terraform-provider-aws/aws"
	"github.com/terraform-providers/terraform-provider-tls/tls"
	terraformv1alpha1 "github.com/tmax-cloud/terraform-operator/api/v1alpha1"
	"github.com/tmax-cloud/terraform-operator/terranova"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Terraform HCL Input Parameters Structure
type TerraVars struct {
	/* Common */
	Name      string
	Namespace string
	Type      string
	/*
		AWS_VPC     AWS_VPC
		AWS_SUBNET  AWS_SUBNET
		AWS_GATEWAY AWS_GATEWAY
		AWS_ROUTE   AWS_ROUTE
		AWS_SG      AWS_SG
		AWS_SG_RULE AWS_SG_RULE
	*/
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

	/* AWSVPC */
	VPCName string
	VPCCIDR string

	/* AWSSubnet */
	SubnetName string
	SubnetCIDR string
	Zone       string

	/* AWSRoute */
	RouteName string
	RouteCIDR string

	/* AWSGateway */
	GatewayName string

	/* AWSSecurityGroup */
	SGName string

	/* AWSSecurityGroupRule */
	SGRuleName string
	SGType     string
	FromPort   string
	ToPort     string
	Protocol   string
	SGCIDR     string

	/* AWSKey */
	KeyName string

	/* AWSInstance */
	InstanceName string
	InstanceType string
	ImageID      string

	/* Network */
	NetworkName string
	//VPCCIDR     string
	//SubnetCIDR  string
	//RouteCIDR   string

	/* Instance */
	//InstanceName string
	//InstanceType string
	//ImageID      string
	//KeyName      string
}

type AWS_VPC struct {
	Name string `json:"name,omitempty"`
	CIDR string `json:"cidr,omitempty"`
}

type AWS_SUBNET struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
	CIDR string `json:"cidr,omitempty"`
	Zone string `json:"zone,omitempty"`
}

type AWS_GATEWAY struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
}

type AWS_ROUTE struct {
	VPC     string `json:"vpc,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Name    string `json:"name,omitempty"`
	CIDR    string `json:"cidr,omitempty"`
}

type AWS_SG struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
}

type AWS_SG_RULE struct {
	SG       string `json:"sg,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	FromPort string `json:"fromport,omitempty"`
	ToPort   string `json:"toport,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	CIDR     string `json:"cidr,omitempty"`
}

// ConfigmapToVars returns a Terraform Variable Struct
func ConfigmapToVars(cm *corev1.ConfigMap) TerraVars {

	configMapData := cm.Data

	output := TerraVars{
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

type Params struct {
	AWSVPC *terraformv1alpha1.AWSVPC
}

// ConfigmapForResource returns a ConfigMap object
func ConfigmapForResource(input TerraVars) *corev1.ConfigMap {

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
			Name:      input.Name,
			Namespace: input.Namespace,
		},
		Data: configMapData,
	}
	return cm
}

// Execute Terraform (Go Package)
// Provison or Destroy the remote resource
func ExecuteTerraform(input TerraVars, destroy bool) error {
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
		if input.Type == "AWSVPC" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_VPC_TEMPLATE
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSSubnet" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SUBNET_TEMPLATE
			code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("zone", input.Zone).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSGatewy" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_GATEWAY_TEMPLATE
			code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSRoute" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_ROUTE_TEMPLATE
			code = strings.Replace(code, "{{ROUTE_NAME}}", input.RouteName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)
			code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("route_cidr", input.RouteCIDR).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSSecurityGroup" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_TEMPLATE
			code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSSecurityGroupRule" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_RULE_TEMPLATE
			code = strings.Replace(code, "{{SG_RULE_NAME}}", input.SGRuleName, -1)
			code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSKey" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_KEY_TEMPLATE
			code = strings.Replace(code, "{{KEY_NAME}}", input.KeyName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("key_pair", input.KeyName).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else if input.Type == "AWSInstance" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_INSTANCE_TEMPLATE
			code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)
			code = strings.Replace(code, "{{INS_NAME}}", input.InstanceName, -1)

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("instance_type", input.InstanceType).
				Var("image_id", input.ImageID).
				Var("key_pair", input.KeyName).
				PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

			if err != nil {
				return err
			}
		} else {
			err = errors.New("Not Found Error: Resource Type")
			return err
		}
	} else if input.Cloud == "Azure" { // Platform : Azure

	} else if input.Cloud == "GCP" { // Platform : Google Cloud Platform

	} else if input.Cloud == "OpenStack" { // Platform : OpenStack

	} else if input.Cloud == "VSphere" { // Platform : VSphere

	} else {
		err = errors.New("Not Found Error: Cloud Platform")
		return err
	}

	// terminate := (count == 0)
	// Apply brings the platform to the desired state. (Provision / Destroy)
	if err := platform.Apply(destroy); err != nil {
		return err
	}

	/*
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
					PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

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
					PersistStateToFile(input.Namespace + "-" + input.ProviderName + ".tfstate")

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

		} else if input.Cloud == "OpenStack" { // Platform : OpenStack

		} else if input.Cloud == "VSphere" { // Platform : VSphere

		} else {
			err = errors.New("Not Found Error: Cloud Platform")
			return err
		}

		// terminate := (count == 0)
		// Apply brings the platform to the desired state. (Provision / Destroy)
		if err := platform.Apply(destroy); err != nil {
			return err
		}
	*/
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
