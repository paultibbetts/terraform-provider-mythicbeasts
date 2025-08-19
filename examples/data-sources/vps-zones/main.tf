terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_zones" "all" {}

output "vps_zones" {
  value = data.mythicbeasts_vps_zones.all
}

