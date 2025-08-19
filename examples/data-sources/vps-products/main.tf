terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_products" "all" {}

output "vps_products" {
  value = data.mythicbeasts_vps_products.all
}

