terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_hosts" "all" {}

output "hosts" {
  value = data.mythicbeasts_vps_hosts.all
}


