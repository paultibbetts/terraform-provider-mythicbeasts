terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_zones" "all" {}

output "vps_zones" {
  value = data.mythicbeasts_vps_zones.all
}

