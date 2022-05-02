# Terraform-Operator
* Terraform-Operator manages terraform resource with Kubernetes API. Based on [Terranova Project](https://github.com/johandry/terranova), custom resources of Terraform-Operator are mapped to terraform resources.
* Terraform-Operator is based on [Terraform v0.12.20](https://releases.hashicorp.com/terraform/0.12.20)

# Supported Providers
Because Terraform-Operator is based on TerraNova Project, the support of cloud providers is as limited as TerraNova.
The primary goal of the project is to control the resources of all cloud providers supported by TerraNova.

* AWS
  * Provider (Credential)
  * Instance (EC2)
  * VPC, Subnet
  * Key
  * SecurityGroup / SecurityGroup Rule
  * Route, Gateway
* Azure (To be supported)
* GCP (To be supported)
* vSphere (To be supported)
