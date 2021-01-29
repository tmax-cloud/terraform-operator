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
	AZURE_PROVIDER_TEMPLATE = `
# Configure the Microsoft Azure Provider
provider "azurerm" {
	# The "feature" block is required for AzureRM provider 2.x.
	# If you're using version 1.x, the "features" block is not allowed.
	version = "~>2.0"

	subscription_id = "{{SUBSCRIPTION_ID}}"
	client_id       = "{{CLIENT_ID}}"
	client_secret   = "{{CLIENT_SECRET}}"
	tenant_id       = "{{TENENT_ID}}"

	features {}
}

# Create a resource group if it doesn't exist
resource "azurerm_resource_group" "{{NET_NAME}}-group" {
	name     = "terrformtestgroup"
	location = "{{REGION}}"

	tags = {
		environment = "{{NET_NAME}}-group"
	}
}
`
	AZURE_NETWORK_TEMPLATE = `
# Create virtual network
resource "azurerm_virtual_network" "{{NET_NAME}}-vpc" {
    name                = "{{NET_NAME}}"
    address_space       = ["{{VPC_CIDR}}"]
    location            = "{{REGION}}"
    resource_group_name = azurerm_resource_group.myterraformgroup.name

    tags = {
        environment = "{{NET_NAME}}-vpc"
    }
}

# Create subnet
resource "azurerm_subnet" "{{NET_NAME}}-subnet" {
    name                 = "{{NET_NAME}}"
    resource_group_name  = azurerm_resource_group.myterraformgroup.name
    virtual_network_name = azurerm_virtual_network.{{NET_NAME}}-vpc.name
    address_prefixes       = ["{{SUBNET_CIDR}}"]
}

# Create public IPs
resource "azurerm_public_ip" "{{NET_NAME}}-publicip" {
    name                         = "{{NET_NAME}}-publicip"
    location                     = "{{REGION}}"
    resource_group_name          = azurerm_resource_group.myterraformgroup.name
    allocation_method            = "Dynamic"

    tags = {
        environment = "{{NET_NAME}}-publicip"
    }
}

# Create Network Security Group and rule
resource "azurerm_network_security_group" "{{NET_NAME}}-sg" {
    name                = "{{NET_NAME}}-sg"
    location            = "{{REGION}}"
    resource_group_name = azurerm_resource_group.myterraformgroup.name

    security_rule {
        name                       = "SSH"
        priority                   = 1001
        direction                  = "Inbound"
        access                     = "Allow"
        protocol                   = "Tcp"
        source_port_range          = "*"
        destination_port_range     = "22"
        source_address_prefix      = "*"
        destination_address_prefix = "*"
    }

    tags = {
        environment = "{{NET_NAME}}-sg"
    }
}
# Create network interface
resource "azurerm_network_interface" "{{NET_NAME}}-nic" {
    name                      = "{{NET_NAME}}-nic"
    location                  = "{{REGION}}"
    resource_group_name       = azurerm_resource_group.myterraformgroup.name

    ip_configuration {
        name                          = "{{NET_NAME}}-nicconfiguration"
        subnet_id                     = azurerm_subnet.{{NET_NAME}}-subnet.id
        private_ip_address_allocation = "Dynamic"
        public_ip_address_id          = azurerm_public_ip.{{NET_NAME}}-publicip.id
    }

    tags = {
        environment = "{{NET_NAME}}-nic"
    }
}

# Connect the security group to the network interface
resource "azurerm_network_interface_security_group_association" "{{NET_NAME}}" {
    network_interface_id      = azurerm_network_interface.{{NET_NAME}}-nic.id
    network_security_group_id = azurerm_network_security_group.{{NET_NAME}}-sg.id
}
`
	AZURE_INSTANCE_TEMPLATE = `
# Generate random text for a unique storage account name
resource "random_id" "randomId" {
    keepers = {
        # Generate a new ID only when a new resource group is defined
        resource_group = azurerm_resource_group.myterraformgroup.name
    }

    byte_length = 8
}

# Create storage account for boot diagnostics
resource "azurerm_storage_account" "mystorageaccount" {
    name                        = "diag${random_id.randomId.hex}"
    resource_group_name         = azurerm_resource_group.myterraformgroup.name
    location                    = "{{REGION}}"
    account_tier                = "Standard"
    account_replication_type    = "LRS"

    tags = {
        environment = "Terraform Demo"
    }
}

# Create virtual machine
resource "azurerm_linux_virtual_machine" "myterraformvm" {
    name                  = "{{NAME}}"
    location              = "{{REGION}}"
    resource_group_name   = azurerm_resource_group.myterraformgroup.name
    network_interface_ids = [azurerm_network_interface.{{NET_NAME}}-nic.id]
    size                  = "Standard_DS1_v2"

    os_disk {
        name              = "myOsDisk"
        caching           = "ReadWrite"
        storage_account_type = "Premium_LRS"
    }

    source_image_reference {
        publisher = "Canonical"
        offer     = "UbuntuServer"
        sku       = "18.04-LTS"
        version   = "latest"
    }

    computer_name  = "{{NAME}}"
    admin_username = "azureuser"
    disable_password_authentication = true

    admin_ssh_key {
        username       = "azureuser"
        public_key     = tls_private_key.example_ssh.public_key_openssh
    }

    boot_diagnostics {
        storage_account_uri = azurerm_storage_account.mystorageaccount.primary_blob_endpoint
    }

    tags = {
        environment = "{{NAME}}"
    }
}
	`
	AZURE_KEY_TEMPLATE = `
# Create (and display) an SSH key
resource "tls_private_key" "example_ssh" {
  algorithm = "RSA"
  rsa_bits = 4096
}
output "tls_private_key" { value = tls_private_key.example_ssh.private_key_pem }
	`
)
