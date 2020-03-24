# Continuous Deployment Guidelines for Terraform Infrastructure


| Revision | Date         | Author         | Remarks                                |
| -------: | ------------ | -------------- | -------------------------------------- |
|      0.1 | Mar-16, 2020 | Nathaniel Rose | Initial Draft                          |



## 1. Overview

It is highly desireable to have a fully-automated CI/CD system in place to assist with testing and validation of source code changes made to the application services and supporting infrastructure. Without such a system, it is difficult to have any degree of confidence that a new deployment will succeed without introducing new regressions.

This document outlines a solution that seeks to validate queued terraform modifications through a gitops pipeline with robust integration tests, monitored staging environments and rollback mechanisms to ensure the successful deployment of terraform infrastructure.

Components of this design are based on the learnings from:
- [Microsoft Bedrock Kubernetes Workflow Project](github.com/microsoft/bedrock)
- [Microsoft Cobalt Infrastructure-As-Code project](github.com/microsoft/cobalt)
- [SPK Infrastructure Generation Pipeline](https://github.com/CatalystCode/spk/blob/master/guides/infra/spk-infra-generation-pipeline.md) 
- [Terraform Recommended Practices](https://www.terraform.io/docs/cloud/guides/recommended-practices/) 

## 2. Scope

This design shall only target the following:
- Triggered validation Tests for incoming infrastructure changes
- Management of staging environments using Azure DevOps
- Automated Rollback levers with human approvals

## 3. Design Details

### 3.1 Validation Tests with Terraform Infrastructure in Azure DevOps

![](/images/Validation_Cycle.png)

In the current implementation on SPK infra practices, the tool is embedded into a generation pipeline that provisions terraform files respective to the modified infra HLD. In addition, this pipeline YAML is expected to run linting, static unit tests, negative testing, initialize terraform state and output a terraform plan file.

Within the Infrastructure Deployment Pipeline, the incoming pull request made by the Generation Pipeline is adjusted by mandated variable secret policies. Next a build is validated in a multi-stage environment during an infrastructure composition workflow, and reviewed by an administrative operator upon successful test pass.  

> Resolving immutable infrastructure variables, global parameters and secrets can be handled through a protected key vault or variable group that populates environment variables based on respective staged environments. Alternatively, entire files can be encrypted within the repo upon check in using a git encryption tooling CLI.

> Upon completion of the variable resolve, the Deployment Pipeline loads a configuration matrix for environment staging that is maintained in AzDO.

### 3.2 Management of staging environments using Azure DevOps

The Configuration Matrix for Environment Staging is built within this pipeline to generate separate deployment resources for end-to-end integration testing based on the incoming pull request. Depending on the rights of the user and the approval policy of the organization, certain environments need approval before deployment.

![](/images/buildSample.png)

Each Deployment Environment is template from a Terraform Workspace which allows the operator to associate a repository with an Azure Devops Terraform environment to assure a git history for each environment is persisted. During each stage environment, a respective build integration test is executed to assure quality.

> At a Glance: Nat has three environments mapped to their configuration matrix. In the first environment, pull requests deploy whitelisted terraform resources to assure cloud provisioning. In the second environment, all incoming modifications are applied on an existing terraform deployment to assure validation. Upon approval, the build is promoted to the final stage to run stateful integration tests.

### 3.3 Automated rollback and alerting using Azure DevOps

Rollback can be triggered in a separate AzDO Pipeline upon failure to build an Environment stage. Using automation and the release history of a pipeline, an environment can be reverted to a previous state pending operator approval.


## 4. Dependencies
- AZDO

## 5. Risks & Mitigations
Severe complications can arise if there are incompatible database changes during failed apply.

## 6. Documentation
Documentation should be done in a `md` file and a new YAML template should be produced for the Infra Deploymnent Validation.