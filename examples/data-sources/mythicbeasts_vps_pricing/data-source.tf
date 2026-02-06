terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_pricing" "all" {}

output "pricing" {
  value = data.mythicbeasts_vps_pricing.all
}


