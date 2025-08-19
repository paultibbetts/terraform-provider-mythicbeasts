terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_hosts" "all" {}

output "hosts" {
  value = data.mythicbeasts_vps_hosts.all
}


