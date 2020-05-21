# `github.com/microsoft/terraform-templates`
~~Azure Terraform Catalog | ATC~~

~~Agnostic Terraform Catalog~~

~~Agnostic Templates for Cloud~~

~~Arbitrary Terraform Catalog~~

~~A Terraform Catalog~~


Templates are the implementation of Advocated Patterns. The scope of a template typically covers most of if not all of the infrastructure required to host an application and may provision resources in multiple cloud provider. Templates compose modules to create an advocated pattern. They are implemented as Terraform Modules so that they can be composed if needed, though it is more commonly the case that they need not be composed.


```
ATC
├── aws
├── azure
│   ├── modules
│   │   ├── acr
│   │   ├── ad-application
│   │   ├── aks
│   │   ├── aks-gitops
│   │   ├── api-mgmt
│   │   ├── app-gateway
│   │   ├── app-insights
│   │   ├── app-monitoring
│   │   ├── app-service
│   │   ├── backend-state
│   │   ├── cosmos-mongo-db-simple
│   │   ├── function-app
│   │   ├── keyvault
│   │   ├── keyvault_flexvol
│   │   ├── keyvault_flexvol_rol
│   │   ├── keyvault_policy
│   │   ├── keyvault_secret
│   │   ├── provider
│   │   ├── README-maintenance.md
│   │   ├── README.md
│   │   ├── redis-cache
│   │   ├── service-bus
│   │   ├── service-plan
│   │   ├── service-principal
│   │   ├── storage-account
│   │   ├── subnet
│   │   ├── tm-endpoint-i
│   │   ├── tm-profile
│   │   ├── vnet
│   │   └── waf
│   └── templates
│       ├── az-aks
│       │   ├── az-aks-common
│       │   ├── az-aks-kv-single
│       │   ├── az-aks-msi
│       │   └── az-aks-simple
│       ├── az-app-service-simple
│       │   └── tests
│       │      └── integration
│       │           └── az_hw_test.go
│       ├── az-service-single-region
│       └── minikube
│           └── deploy_minikube.sh
├── common
│   ├── flux
│   ├── kubediff
│   ├── provider
│   └── velero
├── gcp
└── README.md

```

### 1 Embed new Infrastructure DevOps Model Flow - Continuous Integration

Bedrock infrastructure integration tests have problematic gaps that do not
account for terraform unit testing, state validation to live environments and
staged release management for Bedrock versioning. Bedrock test harness does not
contain module targeted fail fast resource definition validation outside the
scope of an environment `terraform plan`. In addition, integration tests are
validated through new deployments that require extensive time to provision.
Furthermore, releases of features contain no issue reporting benchmark,
automated deployment validation, or guidance process for merging into master. In
this design we wish to provide a single template leveraging MSI that verifies a
new Infrastructure Testing Workflow that improves on the current Bedrock test
harness.

This design is intended to address expected core testing functionality
including:

- Support deployment of application-hosting infrastructure that will eventually
  house the actual application service components capture basic metrics and
  telemetry from the deployment process for monitoring of ongoing pipeline
  performance and diagnosis of any deployment failures
- Support deployment into multiple staging environments
- Execute automated unit-level and integration-level tests against the
  resources, prior to deployment into any long-living environments
- Provide a manual approval process to gate deployment into long-living
  environments
- Provide detection, abort, and reporting of deployment status when a failure
  occurs.

![](infratestflow.png)

The proposed new Infrastructure Devops Flow for Terraform Testing can be
separated by 4 key steps:

1. Test Suite Initialization - Provisioning global artifacts, secrets and
   dependencies needed for targeted whitelisted test matrix.
2. Static Validation - Environment initialization, code validation, inspection,
   terraform security compliance, and terraform module unit tests.
3. Dynamic Validation - Targeted environment interoperability, integration
   tests, cloud deployment, de-provisioning of resources, error reporting.
4. QA- Peer approval, release management, feature staging, acceptance test
   within live cluster.

> The diagram above contains green check marks that indicate preexisting Bedrock
> testing components that are already implemented through the current test
> harness.

### 2 Creation of terraform templates enable AKS Gitops Environments

A new AKS Bedrock template with terraform templates enabled, (`azure-MI`), will be
added to the collection of environment templates. This template will be an
upgraded derivative of the `azure-simple` template, with a new dependency on
`azure-common-infra` and will contain the following:

- terraform templates System Level for AKS
- Pod Identity Security Policy
- Backend State

**Sample `Main.tf`**

```
resource "azurerm_resource_group" "aks_rg" {
  name     = local.aks_rg_name
  location = local.region
}

module "aks-gitops" {
  source = "github.com/microsoft/bedrock?ref=aks_msi_integration//cluster/azure/aks-gitops"

  acr_enabled              = true
  agent_vm_count           = var.aks_agent_vm_count
  agent_vm_size            = var.aks_agent_vm_size
  cluster_name             = local.aks_cluster_name
  dns_prefix               = local.aks_dns_prefix
  flux_recreate            = var.flux_recreate
  gc_enabled               = true
  msi_enabled              = true
  gitops_ssh_url           = var.gitops_ssh_url
  gitops_ssh_key           = var.gitops_ssh_key_file
  gitops_path              = var.gitops_path
  gitops_poll_interval     = var.gitops_poll_interval
  gitops_label             = var.gitops_label
  gitops_url_branch        = var.gitops_url_branch
  kubernetes_version       = var.kubernetes_version
  resource_group_name      = azurerm_resource_group.aks_rg.name
  service_principal_id     = module.app_management_service_principal.service_principal_application_id
  service_principal_secret = module.app_management_service_principal.service_principal_password
  ssh_public_key           = file(var.ssh_public_key_file)
  vnet_subnet_id           = module.vnet.vnet_subnet_ids[0]
  network_plugin           = var.network_plugin
  network_policy           = var.network_policy
  oms_agent_enabled        = var.oms_agent_enabled
}
```

Questions & Limitations:

- With the deployment of the `azure-common-infra` template for Key Vault, will
  that also need to be modified for Manage Identity to whitelist AKS to access
  keyvault?

### 3 Testing for terraform templates enable AKS Gitops Environments

The testing for the terraform templates enabled AKS gitops environment will
incorporate the aforementioned new Infrastructure DevOps Model Flow for
Terraform to assess pod identity access for a Voting App service deployed using
terraform and a flux manifest repository.

#### Unit Tests

Terratest Abstraction Test Fixtures includes a library that simplifies writing
unit terraform tests against templates. It extracts out pieces of this process
and provides a static validation for a json sample output per module. For this,
we require Unit Tests for the following modules:

- AKS
- Key Vault
- VNet
- Subnet
- Gitops

#### Integration Tests

Integration tests will validate resource interoperability upon deployment.
Pending a successful `terraform apply`, using a go script and terratest go
library, this design will create an integration test for the respective
environment template that verifies

- Access to cluster through MI
- Flux namespace
- Access to voting app using Pod Identity
- Access to key using flex-volume
  ([Unable to use Env Vars](https://github.com/Azure/kubernetes-keyvault-flexvol/issues/28))
- 200 response on Voting App

#### Acceptance Test

Acceptance tests are defined in this design as a system affirmation that the
incoming PR has a successful build in a live staging environment once applied.
Maintain a live QA environment that successful builds from an incoming PR are
applied to the state file.

Questions & Limitations:

- With an incoming change to an azure provider module, how will this be applied
  to an existing terraform deployment. If fail, should we redeploy a new
  `azure-MI` environment for QA?

#### Reporting

Output a test failure report using out-of-box terratest JUnit compiler to
capture errors thrown during build.

The whitelisted integration test for `azure-MI` will include:

> `go test -v -run TestIT_Bedrock_AzureMI_Test -timeout 99999s | tee TestIT_Bedrock_AzureMI_Test.log`

> `terratest_log_parser -testlog TestIT_Bedrock_AzureSimple_Test.log -outputdir single_test_output`

The pipeline will publish the XML report as an artifact that is uniquely named
to AzDO.

```
 task: PublishPipelineArtifact@1
        inputs:
          path: $(modulePath)/test/single_test_output
          artifact: simple_test_logs
        condition: always()
      - task: PublishTestResults@2
        inputs:
          testResultsFormat: 'JUnit'
          testResultsFiles: '**/report.xml'
          searchFolder: $(modulePath)/test
        condition: and(eq(variables['Agent.JobStatus'], 'Succeeded'), endsWith(variables['Agent.JobName'], 'Bedrock_Build_Azure_MI'))
```