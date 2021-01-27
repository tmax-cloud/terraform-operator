package util

const (
	HCL_DIR           = "/terraform"
	TERRAFORM_VERSION = "0.11.13"

	AWS_PROVIDER_TEMPLATE = `
	variable "access_key" {}
	variable "secret_key" {}
	variable "region" {}

	provider "aws" {
		version    = "~> 2.31"
		access_key = "${var.access_key}"
		secret_key = "${var.secret_key}"
		region     = "${var.region}"
	  }
	`

	AWS_NETWORK_TEMPLATE = `
	variable "vpc_cidr" {}
	variable "subnet_cidr" {}
	variable "route_cidr" {}

	# Configure the VPC-Subnet
	resource "aws_vpc" "{{NET_NAME}}-vpc" {
		cidr_block = "${var.vpc_cidr}"
		tags = {
			Name = "{{NET_NAME}}-vpc"
		}
	}

	resource "aws_subnet" "{{NET_NAME}}-subnet-c" {
		vpc_id = "${aws_vpc.{{NET_NAME}}-vpc.id}"
		cidr_block = "${var.subnet_cidr}"
		availability_zone = "${var.region}a"
	}

	# Configure the Gateway
	resource "aws_internet_gateway" "{{NET_NAME}}-gateway" {
		vpc_id = "${aws_vpc.{{NET_NAME}}-vpc.id}"
		tags = {
			Name = "{{NET_NAME}}-gateway"
		}
	}

	# Configure the Routes
	resource "aws_route_table" "{{NET_NAME}}-route-table" {
		vpc_id = "${aws_vpc.{{NET_NAME}}-vpc.id}"
		route {
			cidr_block = "${var.route_cidr}"
			gateway_id = "${aws_internet_gateway.{{NET_NAME}}-gateway.id}"
		}
		tags = {
			Name = "{{NET_NAME}}-route-table"
		}
	}

	resource "aws_route_table_association" "{{NET_NAME}}-subnet-association" {
		subnet_id      = "${aws_subnet.{{NET_NAME}}-subnet-c.id}"
		route_table_id = "${aws_route_table.{{NET_NAME}}-route-table.id}"
	}

	# Configure the Security Group
	resource "aws_security_group" "{{NET_NAME}}-sg" {
		vpc_id      = "${aws_vpc.{{NET_NAME}}-vpc.id}"
		name        = "{{NET_NAME}}-sg"
		description = "This security group is for kubernetes"
		tags = { Name = "{{NET_NAME}}-sg" }
	}

	# Configure the Security Rules
	resource "aws_security_group_rule" "kube-cluster-traffic" {
		type              = "ingress"
		from_port         = 0
		to_port           = 0
		protocol = "-1"
		cidr_blocks       = ["10.0.0.0/16"]
		security_group_id = "${aws_security_group.{{NET_NAME}}-sg.id}"
		lifecycle { create_before_destroy = true }
	}
	resource "aws_security_group_rule" "instance-ssh" {
		type              = "ingress"
		from_port         = 22
		to_port           = 22
		protocol = "TCP"
		cidr_blocks       = ["0.0.0.0/0"]
		security_group_id = "${aws_security_group.{{NET_NAME}}-sg.id}"
		lifecycle { create_before_destroy = true }
	  }
	  
	resource "aws_security_group_rule" "outbound-traffic" {
		type              = "egress"
		from_port         = 0
		to_port           = 0
		protocol          = "-1"
		cidr_blocks       = ["0.0.0.0/0"]
		security_group_id = "${aws_security_group.{{NET_NAME}}-sg.id}"
		lifecycle { create_before_destroy = true }
	}
	  
	# Configure the Output
	output "{{NET_NAME}}-subnet-c-id" {
		value = "${aws_subnet.{{NET_NAME}}-subnet-c.id}"
	}
	  
	output "{{NET_NAME}}-sg-id" {
		value = "${aws_security_group.{{NET_NAME}}-sg.id}"
	}
	`

	AWS_INSTANCE_TEMPLATE = `
	variable "key_pair" {default = "aws-key"}
	variable "instance_type" {}
	variable "ami" {}
	  
	resource "aws_instance" "{{INS_NAME}}" {
		ami = "${var.ami}"
		instance_type = "${var.instance_type}"
		subnet_id = "${aws_subnet.{{NET_NAME}}-subnet-c.id}"
		vpc_security_group_ids = [
			"${aws_security_group.{{NET_NAME}}-sg.id}"
		]
		key_name = "${var.key_pair}"
		count = 1
		tags = {
			Name = "{{INS_NAME}}"
		}
		associate_public_ip_address = true
	} 
	`

	AWS_KEY_TEMPLATE = `
	resource "tls_private_key" "example" {
		algorithm = "RSA"
		rsa_bits  = 4096
	  }
	  
	  resource "aws_key_pair" "terraform-key" {
		key_name = "aws-key"
		public_key = "${tls_private_key.example.public_key_openssh}"
	  }	  
	`
	/*
			AWS_INSTANCE_TEMPLATE = `
			variable "c"    { default = 2 }
			variable "key_name" {}
			variable "instance_name" {}
			variable "instance_type" {}
			variable "ami" {}

			resource "aws_instance" "${var.instance_name}" {
			  instance_type = "${var.instance_type}"
			  ami           = "${var.ami}"
			  count         = "${var.c}"
			  key_name      = "${var.key_name}"
			}
		    `
	*/
	AZURE_PROVIDER_TEMPLATE = ""
	AZURE_NETWORK_TEMPLATE  = ""
	AZURE_INSTANCE_TEMPLATE = ""
)
