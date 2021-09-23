# Generated by: tyk-ci/wf-gen
# Generated on: Thu Sep 23 14:04:37 UTC 2021

# Generation commands:
# ./pr.zsh -repos tyk-sync -title releng: latest releng -branch releng/updates
# m4 -E -DxREPO=tyk-sync


data "terraform_remote_state" "integration" {
  backend = "remote"

  config = {
    organization = "Tyk"
    workspaces = {
      name = "base-prod"
    }
  }
}

output "tyk-sync" {
  value = data.terraform_remote_state.integration.outputs.tyk-sync
  description = "ECR creds for tyk-sync repo"
}

output "region" {
  value = data.terraform_remote_state.integration.outputs.region
  description = "Region in which the env is running"
}
