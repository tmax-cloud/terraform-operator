module github.com/tmax-cloud/terraform-operator

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v36.2.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/hashicorp/terraform v0.12.20
	github.com/jen20/awspolicyequivalence v1.1.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/prometheus/common v0.4.1
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20191003145700-f8707a46c6ec
	github.com/terraform-providers/terraform-provider-azurerm v1.34.0
	github.com/terraform-providers/terraform-provider-null v1.0.0
	github.com/terraform-providers/terraform-provider-tls v2.1.0+incompatible
	github.com/zclconf/go-cty v1.7.1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v10.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
)

replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v32.5.0+incompatible
	github.com/terraform-providers/terraform-provider-tls => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281
	k8s.io/client-go => k8s.io/client-go v0.18.6
)
