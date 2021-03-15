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

	"github.com/hashicorp/terraform/plans"
	"github.com/prometheus/common/log"
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
	VPCID   string
	VPCName string
	VPCCIDR string

	/* AWSSubnet */
	SubnetID   string
	SubnetName string
	SubnetCIDR string
	Zone       string

	/* AWSRoute */
	RouteID   string
	RouteName string
	RouteCIDR string

	/* AWSGateway */
	GatewayID   string
	GatewayName string

	/* AWSSecurityGroup */
	SGID   string
	SGName string

	/* AWSSecurityGroupRule */
	SGRuleID   string
	SGRuleName string
	SGType     string
	FromPort   string
	ToPort     string
	Protocol   string
	SGCIDR     string

	/* AWSKey */
	KeyID   string
	KeyName string

	/* AWSInstance */
	InstanceID   string
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
		Name:      configMapData["Name"],
		Namespace: configMapData["Namespace"],
		Type:      configMapData["Type"],

		ProviderName:   configMapData["ProviderName"],
		Cloud:          configMapData["Cloud"],
		Region:         configMapData["Region"],
		AccessKey:      configMapData["AccessKey"],
		SecretKey:      configMapData["SecretKey"],
		SubscriptionID: configMapData["SubscriptionID"],
		ClientID:       configMapData["ClientID"],
		ClientSecret:   configMapData["ClientSecret"],
		TenantID:       configMapData["TenantID"],

		VPCID:   configMapData["VPCID"],
		VPCName: configMapData["VPCName"],
		VPCCIDR: configMapData["VPCCIDR"],

		SubnetID:   configMapData["SubnetID"],
		SubnetName: configMapData["SubnetName"],
		SubnetCIDR: configMapData["SubnetCIDR"],
		Zone:       configMapData["Zone"],

		RouteID:   configMapData["RouteID"],
		RouteName: configMapData["RouteName"],
		RouteCIDR: configMapData["RouteCIDR"],

		GatewayID:   configMapData["GatewayID"],
		GatewayName: configMapData["GatewayName"],

		SGID:   configMapData["SGID"],
		SGName: configMapData["SGName"],

		SGRuleID:   configMapData["SGRuleID"],
		SGRuleName: configMapData["SGRuleName"],
		SGType:     configMapData["SGType"],
		FromPort:   configMapData["FromPort"],
		ToPort:     configMapData["ToPort"],
		Protocol:   configMapData["Protocol"],
		SGCIDR:     configMapData["SGCIDR"],

		KeyID:   configMapData["KeyID"],
		KeyName: configMapData["KeyName"],

		InstanceID:   configMapData["InstanceID"],
		InstanceName: configMapData["InstanceName"],
		InstanceType: configMapData["InstanceType"],
		ImageID:      configMapData["ImageID"],
	}

	/*
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
	*/
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

// SearchResourceID returns a ConfigMap object
/*
func SearchResourceID(input TerraVars) TerraVars {

	if input.VPCName != nil && input.VPCID == nil {

	}


	e := reflect.ValueOf(&input).Elem()

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		//varType := e.Type().Field(i).Type
		varValue := fmt.Sprintf("%v", e.Field(i).Interface())

		configMapData[varName] = varValue
	}


	return output
}
*/

// ReadIDFromFile returns a Cloud Resource ID from Terraform State File
func ReadIDFromFile(filename string) (string, error) {
	var matched string // line with id
	var id string

	input, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error(err, "Failed to read Terraform State File")
		return "", err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "\"id\"") {
			matched = lines[i]
		}
	}

	t := strings.Split(matched, "\"")

	if len(t) >= 4 {
		id = t[3]
	} else {
		err = errors.New("Index out of range")
		return "", err
	}

	return id, nil
}

// plan Terraform (Go Package)
func PlanTerraform(input TerraVars) (string, error) {
	var platform *terranova.Platform // Platform is the platform to be managed by Terraform
	var code string                  // HCL (Hashicorp Configuration Language)
	var filename string
	var err error

	var plan *plans.Plan
	var stats *terranova.Stats
	var status string

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

			filename = input.Namespace + "-" + input.Type + "-" + input.VPCName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSubnet" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SUBNET_TEMPLATE
			code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SubnetName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("zone", input.Zone).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSGatewy" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_GATEWAY_TEMPLATE
			code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.GatewayName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSRoute" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_ROUTE_TEMPLATE
			code = strings.Replace(code, "{{ROUTE_NAME}}", input.RouteName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			code = strings.Replace(code, "{{GATEWAY_ID}}", input.GatewayID, -1)
			code = strings.Replace(code, "{{SUBNET_ID}}", input.SubnetID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)
			//code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			//code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.RouteName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("route_cidr", input.RouteCIDR).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSecurityGroup" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_TEMPLATE
			code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SGName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSecurityGroupRule" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_RULE_TEMPLATE
			code = strings.Replace(code, "{{SG_RULE_NAME}}", input.SGRuleName, -1)
			code = strings.Replace(code, "{{SG_NAME}}", input.SGID, -1)
			//code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SGRuleName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSKey" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_KEY_TEMPLATE
			code = strings.Replace(code, "{{KEY_NAME}}", input.KeyName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.KeyName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("key_pair", input.KeyName).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSInstance" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_INSTANCE_TEMPLATE
			code = strings.Replace(code, "{{INS_NAME}}", input.InstanceName, -1)
			code = strings.Replace(code, "{{SUBNET_ID}}", input.SubnetID, -1)
			code = strings.Replace(code, "{{SG_ID}}", input.SGID, -1)
			//code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			//code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.InstanceName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("instance_type", input.InstanceType).
				Var("image_id", input.ImageID).
				Var("key_pair", input.KeyName).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else {
			err = errors.New("Not Found Error: Resource Type")
			return "", err
		}
	} else if input.Cloud == "Azure" { // Platform : Azure

	} else if input.Cloud == "GCP" { // Platform : Google Cloud Platform

	} else if input.Cloud == "OpenStack" { // Platform : OpenStack

	} else if input.Cloud == "VSphere" { // Platform : VSphere

	} else {
		err = errors.New("Not Found Error: Cloud Platform")
		return "", err
	}

	// terminate := (count == 0)
	// Apply brings the platform to the desired state. (Provision / Destroy)
	if plan, err = platform.Plan(false); err != nil {
		return "", err
	}

	stats = terranova.NewStats().FromPlan(plan)

	status = "provisioned"
	if stats.Change >= 1 || stats.Destroy >= 1 {
		status = "chanaged"
	}
	if stats.Add >= 1 {
		status = "destroyed"
	}

	return status, nil
}

// Execute Terraform (Go Package)
// Provison or Destroy the remote resource
func ExecuteTerraform(input TerraVars, destroy bool) (string, error) {
	var platform *terranova.Platform // Platform is the platform to be managed by Terraform
	var code string                  // HCL (Hashicorp Configuration Language)
	var filename string
	var err error
	var id string

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

			filename = input.Namespace + "-" + input.Type + "-" + input.VPCName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("vpc_cidr", input.VPCCIDR).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSubnet" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SUBNET_TEMPLATE
			code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SubnetName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("subnet_cidr", input.SubnetCIDR).
				Var("zone", input.Zone).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSGatewy" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_GATEWAY_TEMPLATE
			code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.GatewayName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSRoute" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_ROUTE_TEMPLATE
			code = strings.Replace(code, "{{ROUTE_NAME}}", input.RouteName, -1)
			code = strings.Replace(code, "{{VPC_ID}}", input.VPCID, -1)
			code = strings.Replace(code, "{{GATEWAY_ID}}", input.GatewayID, -1)
			code = strings.Replace(code, "{{SUBNET_ID}}", input.SubnetID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)
			//code = strings.Replace(code, "{{GATEWAY_NAME}}", input.GatewayName, -1)
			//code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.RouteName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("route_cidr", input.RouteCIDR).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSecurityGroup" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_TEMPLATE
			code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)
			code = strings.Replace(code, "{{VPC_NAME}}", input.VPCID, -1)
			//code = strings.Replace(code, "{{VPC_NAME}}", input.VPCName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SGName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSSecurityGroupRule" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_SECURITY_GROUP_RULE_TEMPLATE
			code = strings.Replace(code, "{{SG_RULE_NAME}}", input.SGRuleName, -1)
			code = strings.Replace(code, "{{SG_NAME}}", input.SGID, -1)
			//code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.SGRuleName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSKey" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_KEY_TEMPLATE
			code = strings.Replace(code, "{{KEY_NAME}}", input.KeyName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.KeyName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("key_pair", input.KeyName).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else if input.Type == "AWSInstance" {
			code = AWS_PROVIDER_TEMPLATE + "\n" + AWS_INSTANCE_TEMPLATE
			code = strings.Replace(code, "{{INS_NAME}}", input.InstanceName, -1)
			code = strings.Replace(code, "{{SUBNET_ID}}", input.SubnetID, -1)
			code = strings.Replace(code, "{{SG_ID}}", input.SGID, -1)
			//code = strings.Replace(code, "{{SUBNET_NAME}}", input.SubnetName, -1)
			//code = strings.Replace(code, "{{SG_NAME}}", input.SGName, -1)

			filename = input.Namespace + "-" + input.Type + "-" + input.InstanceName + ".tfstate"

			platform, err = terranova.NewPlatform(code).
				AddProvider("aws", aws.Provider()).
				AddProvider("tls", tls.Provider()).
				Var("access_key", input.AccessKey).
				Var("secret_key", input.SecretKey).
				Var("region", input.Region).
				Var("instance_type", input.InstanceType).
				Var("image_id", input.ImageID).
				Var("key_pair", input.KeyName).
				PersistStateToFile(filename)

			if err != nil {
				return "", err
			}
		} else {
			err = errors.New("Not Found Error: Resource Type")
			return "", err
		}
	} else if input.Cloud == "Azure" { // Platform : Azure

	} else if input.Cloud == "GCP" { // Platform : Google Cloud Platform

	} else if input.Cloud == "OpenStack" { // Platform : OpenStack

	} else if input.Cloud == "VSphere" { // Platform : VSphere

	} else {
		err = errors.New("Not Found Error: Cloud Platform")
		return "", err
	}

	// terminate := (count == 0)
	// Apply brings the platform to the desired state. (Provision / Destroy)
	if err := platform.Apply(destroy); err != nil {
		return "", err
	}

	if !destroy {
		if id, err = ReadIDFromFile(filename); err != nil {
			return "", err
		}
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
	return id, nil
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
