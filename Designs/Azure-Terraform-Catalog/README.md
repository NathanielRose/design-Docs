# `github.com/microsoft/terraform-templates`
Project Azure Terraform Catalog | ATC

~~Agnostic Terraform Catalog~~

~~Agnostic Templates for Cloud~~

~~Arbitrary Terraform Catalog~~

~~A Terraform Catalog~~

## Overview
This repo will serve as a catalog of maintained deployable terraform templates with Azure advocated patterns. The scope of a template typically covers most of if not all of the infrastructure required to host an application and may provision resources in multiple cloud provider. Templates should be loosely coupled and maintained in isolation. Modules should be tightly paired with updates to the latest terraform version, azureRM provider and golang version.

## Merging Bedrock & Cobalt

Bedrock contains patterns for implementation and automation of productionized Kubernetes clusters with GitOps Workflow. Cobalt similarly provides patterns for implementing Application Services in Azure with dependent resources. Below is a tree capture of the two project's terraform templates merged into a shared repository.
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
│   |   │   └── tests
│   |   │      └── unit tests
│   |   │           └── az_hw_test.go
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
│       │   ├── az-aks-kv-single
│       │   ├── az-aks-msi
│       │   └── az-aks-simple
│       ├── az-app-service-simple
│       │   └── tests
│       │      └── integration
│       │           └── az_hw_test.go
│       ├── az-common
│       ├── az-service-single-region
│       └── minikube
│           └── deploy_minikube.sh
├── common
│   ├── flux
│   ├── kubediff
│   ├── provider
│   └── velero
├── gcp
├── devops
│   ├── azure-pipeline.yml
│   ├── scheduled-pipeline.yml
│   ├── updates.yml
│   ├── lint.yml
│   ├── tests-unit.yml
│   ├── tests-int.yml
│   ├── tf-plan.yml
│   ├── tf-apply.yml
│   └── tf-destroy.yml
└── README.md

```

### Templates
Bedrock will migrate 4 Kubernetes templates:
- Azure Single Keyvault
- Azure Managed Identity
- Azure Simple

Cobalt wil migrate 2 Templates
- Azure App Service Single Region (ACR, VNet, KeyVault migrated to AZ Common )
- Azure App Service Simple

A modified Azure Common infra template will be created for use by both Bedrock and Cobalt templates. This template will deploy a keyvault, vnet, and ACR.

Each Template will contain an integration test folder that validates the templates application state upon deployment. 

```
├── az-app-service-simple
│   └── tests
│      └── integration
│           └── az_hw_test.go
```

Questions & Limitations:

- With the deployment of the azure-common template for Key Vault, will that also need to be modified for Manage Identity to whitelist AKS to access keyvault?
- With integration tests folder embedded into the template directory, will bedrock-cli be able to support this?


### Modules

Modules will be tightly paired with the latest AzureRM provider through releases.

Each Module will contain a respective unit test mapped with a configuration for rapid failure in the testing pipleine.

```
│   │   ├── app-service
│   |   │   └── tests
│   |   │      └── unit tests
│   |   │           └── az_hw_test.go
```


## Infrastructure DevOps Model Flow - Template Testing


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

### Terratest Abstraction (Testing Fixtures)

Tests written with [`terratest-abstraction`](https://github.com/microsoft/terratest-abstraction) will invoke components to your testing pipeline like any other Golang test. We separate unit and integration tests so that they can be easily targeted at build and deploy time within an automated CICD pipeline. 

It is easy to invoke tests written with terratest-abstraction using the following commands:

Unit Tests

#### before executing `terraform plan`
```
go test -v $(go list ./... | grep unit)
Integration Tests
```

#### after executing `terraform apply`
```
go test -v $(go list ./... | grep integration)
```

### Unit Tests

Unit tests are implemented quality assurance tests that validates E2E functional assertions against your infrastructure resources. Each template comes pre-packaged with a module that asserts a configuration when a `terraform plan` is invoked in the pipeline. The plan output is checked against a go script local to the module directory.

**Sample service bus `unit-test.go`**

``` go
package unit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/microsoft/cobalt/infra/modules/providers/azure/service-bus/tests"
	"github.com/microsoft/cobalt/test-harness/infratests"
)

var workspace = "osdu-services-" + strings.ToLower(random.UniqueId())

// helper function to parse blocks of JSON into a generic Go map
func asMap(t *testing.T, jsonString string) map[string]interface{} {
	var theMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &theMap); err != nil {
		t.Fatal(err)
	}
	return theMap
}

func TestTemplate(t *testing.T) {

	expectedSBNamespace := map[string]interface{}{
		"capacity": 0.0,
		"name":     tests.NamespaceName,
		"sku":      "Standard",
		"tags": map[string]interface{}{
			"source": "terraform",
		},
	}

	expectedNamespaceAuth := map[string]interface{}{
		"name":   "policy",
		"listen": true,
		"send":   true,
		"manage": false,
	}

	expectedSubscription := map[string]interface{}{
		"name":                                 "sub_test",
		"max_delivery_count":                   1.0,
		"lock_duration":                        "PT5M",
		"forward_to":                           "",
		"dead_lettering_on_message_expiration": true,
	}

	expectedTopic := map[string]interface{}{
		"name":                         "topic_test",
		"default_message_ttl":          "PT30M",
		"enable_partitioning":          true,
		"support_ordering":             true,
		"requires_duplicate_detection": true,
	}

	expectedTopicAuth := map[string]interface{}{
		"name":   "policy",
		"listen": true,
		"send":   true,
		"manage": false,
	}

	expectedSubRules := map[string]interface{}{
		"name":        "sub_test",
		"filter_type": "SqlFilter",
		"sql_filter":  "color = 'red'",
		"action":      "",
	}

	testFixture := infratests.UnitTestFixture{
		GoTest:                t,
		TfOptions:             tests.ServicebusTFOptions,
		Workspace:             workspace,
		PlanAssertions:        nil,
		ExpectedResourceCount: 6,
		ExpectedResourceAttributeValues: infratests.ResourceDescription{
			"azurerm_servicebus_namespace.servicebus":                            expectedSBNamespace,
			"azurerm_servicebus_namespace_authorization_rule.sbnamespaceauth[0]": expectedNamespaceAuth,
			"azurerm_servicebus_topic.sptopic[0]":                                expectedTopic,
			"azurerm_servicebus_subscription.subscription[0]":                    expectedSubscription,
			"azurerm_servicebus_topic_authorization_rule.topicaauth[0]":          expectedTopicAuth,
			"azurerm_servicebus_subscription_rule.subrules[0]":                   expectedSubRules,
		},
	}

	infratests.RunUnitTests(&testFixture)
}
```

### Integration Tests

Integration tests will validate resource interoperability upon deployment.
Pending a successful `terraform apply`, using a go script and terratest go
library, this design will create an integration test for the respective
environment template that verifies application functioniality
 and resource status.


### Acceptance Test (Optional)

Acceptance tests are defined in this design as a system affirmation that the
incoming PR has a successful build in a live staging environment once applied.
Maintain a live QA environment that successful builds from an incoming PR are
applied to the state file. This would sit in the template folder.

## Scheduling

Test are ran on a cyclical CRON Job configuration for the pipeline. The jobs yaml conducts 3 things:

- Triggered Whitelisting tests during PR for validation
- Update check against AzureRM Provider and terraform version to check compatibility
- Scheduled checks for outstanding PRs weekly

## Reporting

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