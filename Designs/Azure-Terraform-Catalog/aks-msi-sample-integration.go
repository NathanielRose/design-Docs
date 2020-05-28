package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	aksGitOpsIntegTests "github.com/microsoft/cobalt/infra/modules/providers/azure/aks-gitops/tests/integration"
	appGatewayIntegTests "github.com/microsoft/cobalt/infra/modules/providers/azure/app-gateway/tests/integration"
	"github.com/microsoft/cobalt/test-harness/infratests"
)

var subscription = os.Getenv("ARM_SUBSCRIPTION_ID")
var kubeConfig = "../../output/bedrock_kube_config"
var tfOptions = &terraform.Options{
	TerraformDir: "../../",
	BackendConfig: map[string]interface{}{
		"storage_account_name": os.Getenv("TF_VAR_remote_state_account"),
		"container_name":       os.Getenv("TF_VAR_remote_state_container"),
	},
}

func validateAADIdentityControllers(kubeConfig string, namespace string) func(t *testing.T, output infratests.TerraformOutput) {

	return func(t *testing.T, output infratests.TerraformOutput) {
		aksGitOpsIntegTests.ValidateAADIdentityControllers(t, kubeConfig, namespace)
	}
}

func validateFluxNamespace(kubeConfig string) func(t *testing.T, output infratests.TerraformOutput) {

	return func(t *testing.T, output infratests.TerraformOutput) {
		aksGitOpsIntegTests.ValidateFluxNamespace(t, kubeConfig)
	}
}

func TestContainerClusterEnvironment(t *testing.T) {
	testFixture := infratests.IntegrationTestFixture{
		GoTest:                t,
		TfOptions:             tfOptions,
		ExpectedTfOutputCount: 17,
		TfOutputAssertions: []infratests.TerraformOutputValidation{
			validateAADIdentityControllers(kubeConfig, output["aks_pod_identity_namespace"].(string)),
			validateFluxNamespace(kubeConfig),
			verifyServicePrincipalRoleAssignments,
			verifyAppGWMSIRoleAssignments,
			appGatewayIntegTests.InspectAppGateway("resource_group", "app_gw_name", "keyvault_secret_id"),
		},
	}
	infratests.RunIntegrationTests(&testFixture)
}
