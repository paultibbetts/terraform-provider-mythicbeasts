terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.2.0"
    }
  }
}

data "mythicbeasts_vps_pricing" "all" {}

output "pricing" {
  value = data.mythicbeasts_vps_pricing.all
}


