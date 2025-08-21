terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.2.0"
    }
  }
}

data "mythicbeasts_vps_zones" "all" {}

output "vps_zones" {
  value = data.mythicbeasts_vps_zones.all
}

