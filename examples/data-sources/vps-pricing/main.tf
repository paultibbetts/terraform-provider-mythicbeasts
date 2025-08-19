terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_pricing" "all" {}

output "pricing" {
  value = data.mythicbeasts_vps_pricing.all
}


